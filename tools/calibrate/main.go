// Command calibrate derives Claude-estimator factors from Anthropic's
// count_tokens endpoint and a representative source-code corpus.
package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/shaunobi/tloc/internal/tokenizer"
)

const (
	defaultEndpoint           = "https://api.anthropic.com/v1/messages/count_tokens"
	framingProbe              = "x"
	framingProbeContentTokens = int64(1)
)

type stringList []string

func (values *stringList) String() string {
	return strings.Join(*values, ",")
}

func (values *stringList) Set(value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("path cannot be empty")
	}
	*values = append(*values, value)
	return nil
}

type namedModel struct {
	label string
	model string
}

type namedModelList []namedModel

func (models *namedModelList) String() string {
	values := make([]string, 0, len(*models))
	for _, model := range *models {
		values = append(values, model.label+"="+model.model)
	}
	return strings.Join(values, ",")
}

func (models *namedModelList) Set(value string) error {
	parts := strings.SplitN(value, "=", 2)
	if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" || strings.TrimSpace(parts[1]) == "" {
		return fmt.Errorf("spot model must use label=model-id")
	}
	*models = append(*models, namedModel{
		label: strings.TrimSpace(parts[0]),
		model: strings.TrimSpace(parts[1]),
	})
	return nil
}

type configuration struct {
	inputs          stringList
	heldOutInputs   stringList
	currentModel    string
	legacyModel     string
	spotModels      namedModelList
	endpoint        string
	apiVersion      string
	outputDirectory string
	maxBytes        int
	maxPerLanguage  int
	maxSamples      int
	maxRetries      int
	requestTimeout  time.Duration
	requestDelay    time.Duration
	dryRun          bool
	confirmSpend    bool
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "calibrate:", err)
		os.Exit(1)
	}
}

func run() error {
	config := parseFlags()
	config.inputs = append(config.inputs, flag.Args()...)

	samples, err := collectSamples(config.inputs, config.maxBytes, config.maxPerLanguage, config.maxSamples)
	if err != nil {
		return err
	}
	if len(samples) == 0 {
		return fmt.Errorf("no supported, non-empty UTF-8 source files found")
	}
	var heldOutSamples []sourceSample
	if len(config.heldOutInputs) > 0 {
		heldOutSamples, err = collectSamples(config.heldOutInputs, config.maxBytes, config.maxPerLanguage, config.maxSamples)
		if err != nil {
			return fmt.Errorf("collect held-out samples: %w", err)
		}
		if len(heldOutSamples) == 0 {
			return fmt.Errorf("no supported, non-empty UTF-8 held-out source files found")
		}
		if err := validateDisjointSamples(samples, heldOutSamples); err != nil {
			return err
		}
	}

	o200k, metadata, err := tokenizer.New(tokenizer.NameO200K)
	if err != nil {
		return err
	}
	for index := range samples {
		samples[index].O200K, err = o200k.Count(samples[index].Content)
		if err != nil {
			return fmt.Errorf("count o200k tokens in %q: %w", samples[index].Path, err)
		}
	}
	for index := range heldOutSamples {
		heldOutSamples[index].O200K, err = o200k.Count(heldOutSamples[index].Content)
		if err != nil {
			return fmt.Errorf("count o200k tokens in held-out sample %q: %w", heldOutSamples[index].Path, err)
		}
	}
	printSamplePlan("Calibration", samples, metadata)
	if len(heldOutSamples) > 0 {
		printSamplePlan("Held out", heldOutSamples, metadata)
	}
	if config.dryRun {
		fmt.Println("Dry run: no Anthropic API requests were made.")
		return nil
	}

	if !config.confirmSpend {
		return fmt.Errorf("refusing to call Anthropic without --confirm-spend (use --dry-run to inspect the sample plan)")
	}
	if strings.TrimSpace(config.currentModel) == "" || strings.TrimSpace(config.legacyModel) == "" {
		return fmt.Errorf("--current-model and --legacy-model are both required")
	}
	apiKey := strings.TrimSpace(os.Getenv("ANTHROPIC_API_KEY"))
	if apiKey == "" {
		return fmt.Errorf("ANTHROPIC_API_KEY is not set")
	}
	if config.requestTimeout <= 0 || config.requestDelay < 0 || config.maxRetries < 0 {
		return fmt.Errorf("request timeout/retry settings are invalid")
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	client := &anthropicClient{
		apiKey:     apiKey,
		apiVersion: config.apiVersion,
		endpoint:   config.endpoint,
		maxRetries: config.maxRetries,
		httpClient: &http.Client{Timeout: config.requestTimeout},
	}

	models := []namedModel{
		{label: tokenizer.NameClaude, model: config.currentModel},
		{label: tokenizer.NameClaudeLegacy, model: config.legacyModel},
	}
	seenLabels := map[string]struct{}{
		tokenizer.NameClaude:       {},
		tokenizer.NameClaudeLegacy: {},
	}
	for _, spot := range config.spotModels {
		if _, duplicate := seenLabels[spot.label]; duplicate {
			return fmt.Errorf("duplicate model label %q", spot.label)
		}
		seenLabels[spot.label] = struct{}{}
		models = append(models, spot)
	}

	report := calibrationReport{
		GeneratedAt:      time.Now().UTC(),
		Endpoint:         config.endpoint,
		AnthropicVersion: config.apiVersion,
		Method:           `Claude framing tokens = count_tokens("x") - 1 known probe token; Claude content tokens = count_tokens(message text) - framing tokens; factor = sum(Claude content tokens) / sum(o200k_base tokens); held-out samples are excluded from fitting and evaluated only after factors are finalized`,
		Sampling: samplingMetadata{
			Inputs:              append([]string(nil), config.inputs...),
			HeldOutInputs:       append([]string(nil), config.heldOutInputs...),
			MaxBytesPerFile:     config.maxBytes,
			MaxFilesPerLanguage: config.maxPerLanguage,
			MaxSamples:          config.maxSamples,
		},
	}

	totalRequests := len(models) * (len(samples) + len(heldOutSamples) + 1)
	fmt.Printf("Calling count_tokens %d times across %d models...\n", totalRequests, len(models))
	requestNumber := 0
	for _, selectedModel := range models {
		if err := delayRequest(ctx, &requestNumber, config.requestDelay); err != nil {
			return err
		}
		probeCount, err := client.countTokens(ctx, selectedModel.model, framingProbe)
		if err != nil {
			return fmt.Errorf("measure %s framing baseline: %w", selectedModel.label, err)
		}
		baseline, err := framingBaselineFromProbe(probeCount)
		if err != nil {
			return fmt.Errorf("measure %s framing baseline: %w", selectedModel.label, err)
		}
		fmt.Printf("%s (%s): probe count = %d tokens; framing baseline = %d tokens\n", selectedModel.label, selectedModel.model, probeCount, baseline)

		modelResult := modelReport{
			Label:           selectedModel.label,
			Model:           selectedModel.model,
			FramingBaseline: baseline,
			Samples:         make([]measurement, 0, len(samples)),
		}
		for index, sample := range samples {
			if err := delayRequest(ctx, &requestNumber, config.requestDelay); err != nil {
				return err
			}
			rawCount, err := client.countTokens(ctx, selectedModel.model, string(sample.Content))
			if err != nil {
				return fmt.Errorf("measure %s sample %q: %w", selectedModel.label, sample.Path, err)
			}
			if rawCount < baseline {
				return fmt.Errorf("%s sample %q raw count %d is below framing baseline %d", selectedModel.label, sample.Path, rawCount, baseline)
			}
			contentCount := rawCount - baseline
			ratio := 0.0
			if sample.O200K > 0 {
				ratio = float64(contentCount) / float64(sample.O200K)
			}
			modelResult.Samples = append(modelResult.Samples, measurement{
				Path:                 sample.Path,
				Language:             sample.Language,
				Bytes:                sample.Bytes,
				Truncated:            sample.Truncated,
				ContentSHA256:        sample.ContentSHA,
				O200KTokens:          sample.O200K,
				ClaudeRawInputTokens: rawCount,
				ClaudeContentTokens:  contentCount,
				ClaudePerO200KRatio:  ratio,
			})
			fmt.Printf("[%d/%d] %s %-14s %s: o200k=%d claude=%d ratio=%.4f\n",
				index+1, len(samples), selectedModel.label, sample.Language, sample.Path,
				sample.O200K, contentCount, ratio)
		}
		finalizeModelReport(&modelResult)

		heldOutMeasurements := make([]measurement, 0, len(heldOutSamples))
		for index, sample := range heldOutSamples {
			if err := delayRequest(ctx, &requestNumber, config.requestDelay); err != nil {
				return err
			}
			rawCount, err := client.countTokens(ctx, selectedModel.model, string(sample.Content))
			if err != nil {
				return fmt.Errorf("measure %s held-out sample %q: %w", selectedModel.label, sample.Path, err)
			}
			if rawCount < baseline {
				return fmt.Errorf("%s held-out sample %q raw count %d is below framing baseline %d", selectedModel.label, sample.Path, rawCount, baseline)
			}
			contentCount := rawCount - baseline
			ratio := 0.0
			if sample.O200K > 0 {
				ratio = float64(contentCount) / float64(sample.O200K)
			}
			heldOutMeasurements = append(heldOutMeasurements, measurement{
				Path:                 sample.Path,
				Language:             sample.Language,
				Bytes:                sample.Bytes,
				Truncated:            sample.Truncated,
				ContentSHA256:        sample.ContentSHA,
				O200KTokens:          sample.O200K,
				ClaudeRawInputTokens: rawCount,
				ClaudeContentTokens:  contentCount,
				ClaudePerO200KRatio:  ratio,
			})
			fmt.Printf("[held out %d/%d] %s %-14s %s: o200k=%d claude=%d ratio=%.4f\n",
				index+1, len(heldOutSamples), selectedModel.label, sample.Language, sample.Path,
				sample.O200K, contentCount, ratio)
		}
		if err := finalizeHeldOutEvaluation(&modelResult, heldOutMeasurements); err != nil {
			return err
		}
		report.Models = append(report.Models, modelResult)
	}

	jsonPath, markdownPath, err := writeReports(config.outputDirectory, report)
	if err != nil {
		return err
	}
	for _, model := range report.Models {
		fmt.Printf("%s global factor %.6f; MAPE %.2f%%\n", model.Label, model.Global.Factor, model.Global.MeanAbsolutePercentError)
		if model.HeldOut != nil {
			fmt.Printf("%s held-out production-factor MAPE %.2f%% across %d samples\n",
				model.Label, model.HeldOut.Summary.MeanAbsolutePercentError, model.HeldOut.Summary.Samples)
		}
	}
	fmt.Printf("Wrote %s and %s\n", jsonPath, markdownPath)
	return nil
}

func framingBaselineFromProbe(rawCount int64) (int64, error) {
	if rawCount < framingProbeContentTokens {
		return 0, fmt.Errorf("probe %q returned %d tokens, want at least %d", framingProbe, rawCount, framingProbeContentTokens)
	}
	return rawCount - framingProbeContentTokens, nil
}

func parseFlags() configuration {
	var config configuration
	flag.Var(&config.inputs, "samples", "sample file or directory (repeatable; positional paths are also accepted)")
	flag.Var(&config.heldOutInputs, "holdout", "held-out sample file or directory, excluded from factor fitting (repeatable)")
	flag.StringVar(&config.currentModel, "current-model", "", "Anthropic model ID for the current tokenizer generation")
	flag.StringVar(&config.legacyModel, "legacy-model", "", "Anthropic model ID for the legacy tokenizer generation")
	flag.Var(&config.spotModels, "spot-model", "additional label=model-id to measure without assigning a compile-time factor (repeatable)")
	flag.StringVar(&config.endpoint, "endpoint", defaultEndpoint, "Anthropic count_tokens endpoint")
	flag.StringVar(&config.apiVersion, "anthropic-version", "2023-06-01", "Anthropic API version header")
	flag.StringVar(&config.outputDirectory, "out", "tools/calibrate/results", "directory for calibration.json and calibration.md")
	flag.IntVar(&config.maxBytes, "sample-bytes", 32768, "maximum bytes sent from each source file")
	flag.IntVar(&config.maxPerLanguage, "max-per-language", 5, "maximum sampled files per language")
	flag.IntVar(&config.maxSamples, "max-samples", 50, "maximum sampled files across all languages")
	flag.IntVar(&config.maxRetries, "max-retries", 4, "retries for rate limits and transient failures")
	flag.DurationVar(&config.requestTimeout, "request-timeout", 60*time.Second, "timeout for each API request")
	flag.DurationVar(&config.requestDelay, "request-delay", 100*time.Millisecond, "minimum delay between API requests")
	flag.BoolVar(&config.dryRun, "dry-run", false, "select samples and count o200k tokens without calling Anthropic")
	flag.BoolVar(&config.confirmSpend, "confirm-spend", false, "confirm that API requests may incur usage or cost")
	flag.Parse()
	return config
}

func delayRequest(ctx context.Context, requestNumber *int, delay time.Duration) error {
	if *requestNumber > 0 && delay > 0 {
		timer := time.NewTimer(delay)
		defer timer.Stop()
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
		}
	}
	(*requestNumber)++
	return nil
}

func printSamplePlan(label string, samples []sourceSample, metadata tokenizer.Metadata) {
	byLanguage := make(map[string]struct {
		files  int
		tokens int64
	})
	var totalTokens int64
	var totalBytes int
	for _, sample := range samples {
		entry := byLanguage[sample.Language]
		entry.files++
		entry.tokens += sample.O200K
		byLanguage[sample.Language] = entry
		totalTokens += sample.O200K
		totalBytes += sample.Bytes
	}
	languages := make([]string, 0, len(byLanguage))
	for language := range byLanguage {
		languages = append(languages, language)
	}
	sort.Strings(languages)
	fmt.Printf("%s: selected %d files (%d bytes, %d %s tokens) across %d languages:\n",
		label, len(samples), totalBytes, totalTokens, metadata.Encoding, len(languages))
	for _, language := range languages {
		entry := byLanguage[language]
		fmt.Printf("  %-14s %3d files %8d tokens\n", language, entry.files, entry.tokens)
	}
}
