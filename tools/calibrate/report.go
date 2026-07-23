package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/shaunobi/tloc/internal/tokenizer"
)

type calibrationReport struct {
	GeneratedAt      time.Time        `json:"generated_at"`
	Endpoint         string           `json:"endpoint"`
	AnthropicVersion string           `json:"anthropic_version"`
	Method           string           `json:"method"`
	Models           []modelReport    `json:"models"`
	Sampling         samplingMetadata `json:"sampling"`
}

type samplingMetadata struct {
	Inputs              []string `json:"inputs"`
	HeldOutInputs       []string `json:"held_out_inputs,omitempty"`
	MaxBytesPerFile     int      `json:"max_bytes_per_file"`
	MaxFilesPerLanguage int      `json:"max_files_per_language"`
	MaxSamples          int      `json:"max_samples"`
}

type modelReport struct {
	Label           string                          `json:"label"`
	Model           string                          `json:"model"`
	FramingBaseline int64                           `json:"framing_baseline"`
	Global          ratioSummary                    `json:"global"`
	PerLanguage     map[string]languageRatioSummary `json:"per_language"`
	Samples         []measurement                   `json:"samples"`
	HeldOut         *heldOutEvaluation              `json:"held_out,omitempty"`
}

type heldOutEvaluation struct {
	FactorSource         string                          `json:"factor_source"`
	GlobalFactor         float64                         `json:"global_factor"`
	CalibrationOverrides []tokenizer.CalibrationOverride `json:"calibration_overrides,omitempty"`
	Summary              evaluationSummary               `json:"summary"`
	PerLanguage          map[string]evaluationSummary    `json:"per_language"`
	Samples              []measurement                   `json:"samples"`
}

type evaluationSummary struct {
	Samples                  int     `json:"samples"`
	O200KTokens              int64   `json:"o200k_tokens"`
	ClaudeContentTokens      int64   `json:"claude_content_tokens"`
	PredictedTokens          int64   `json:"predicted_tokens"`
	SignedAggregateError     float64 `json:"signed_aggregate_percent_error"`
	MeanAbsolutePercentError float64 `json:"mean_absolute_percent_error"`
}

type measurement struct {
	Path                 string  `json:"path"`
	Language             string  `json:"language"`
	Bytes                int     `json:"bytes"`
	Truncated            bool    `json:"truncated"`
	ContentSHA256        string  `json:"content_sha256"`
	O200KTokens          int64   `json:"o200k_tokens"`
	ClaudeRawInputTokens int64   `json:"claude_raw_input_tokens"`
	ClaudeContentTokens  int64   `json:"claude_content_tokens"`
	ClaudePerO200KRatio  float64 `json:"claude_per_o200k_ratio"`
}

type ratioSummary struct {
	Samples                  int     `json:"samples"`
	O200KTokens              int64   `json:"o200k_tokens"`
	ClaudeContentTokens      int64   `json:"claude_content_tokens"`
	Factor                   float64 `json:"factor"`
	MeanPerFileRatio         float64 `json:"mean_per_file_ratio"`
	MedianPerFileRatio       float64 `json:"median_per_file_ratio"`
	MeanAbsolutePercentError float64 `json:"mean_absolute_percent_error"`
}

type languageRatioSummary struct {
	ratioSummary
	GlobalFactorSignedAggregatePercentError float64  `json:"global_factor_signed_aggregate_percent_error"`
	GlobalFactorMeanAbsolutePercentError    float64  `json:"global_factor_mean_absolute_percent_error"`
	LeaveOneOutMeanAbsolutePercentError     *float64 `json:"leave_one_out_mean_absolute_percent_error"`
}

type factorErrorSummary struct {
	AggregatePercentError    float64
	MeanAbsolutePercentError float64
}

func summarize(measurements []measurement) ratioSummary {
	if len(measurements) == 0 {
		return ratioSummary{}
	}

	summary := ratioSummary{Samples: len(measurements)}
	ratios := make([]float64, 0, len(measurements))
	for _, sample := range measurements {
		summary.O200KTokens += sample.O200KTokens
		summary.ClaudeContentTokens += sample.ClaudeContentTokens
		if sample.O200KTokens > 0 {
			ratios = append(ratios, sample.ClaudePerO200KRatio)
		}
	}
	if summary.O200KTokens > 0 {
		summary.Factor = float64(summary.ClaudeContentTokens) / float64(summary.O200KTokens)
	}
	if len(ratios) > 0 {
		for _, ratio := range ratios {
			summary.MeanPerFileRatio += ratio
		}
		summary.MeanPerFileRatio /= float64(len(ratios))
		sort.Float64s(ratios)
		middle := len(ratios) / 2
		if len(ratios)%2 == 0 {
			summary.MedianPerFileRatio = (ratios[middle-1] + ratios[middle]) / 2
		} else {
			summary.MedianPerFileRatio = ratios[middle]
		}
	}

	errors := evaluateFactor(measurements, summary.Factor)
	summary.MeanAbsolutePercentError = errors.MeanAbsolutePercentError
	return summary
}

func evaluateFactor(measurements []measurement, factor float64) factorErrorSummary {
	var summary factorErrorSummary
	var predictedTotal int64
	var actualTotal int64
	var absolutePercentErrorTotal float64
	var percentErrorSamples int
	for _, sample := range measurements {
		prediction := roundedPrediction(sample.O200KTokens, factor)
		predictedTotal += prediction
		actualTotal += sample.ClaudeContentTokens
		if sample.ClaudeContentTokens == 0 {
			continue
		}
		absolutePercentErrorTotal += math.Abs(float64(prediction-sample.ClaudeContentTokens)) / float64(sample.ClaudeContentTokens) * 100
		percentErrorSamples++
	}
	if actualTotal != 0 {
		summary.AggregatePercentError = float64(predictedTotal-actualTotal) / float64(actualTotal) * 100
	}
	if percentErrorSamples > 0 {
		summary.MeanAbsolutePercentError = absolutePercentErrorTotal / float64(percentErrorSamples)
	}
	return summary
}

func roundedPrediction(o200kTokens int64, factor float64) int64 {
	return int64(math.Round(float64(o200kTokens) * factor))
}

func leaveOneOutMeanAbsolutePercentError(measurements []measurement) (float64, bool) {
	if len(measurements) < 2 {
		return 0, false
	}

	var totalO200KTokens int64
	var totalClaudeContentTokens int64
	for _, sample := range measurements {
		totalO200KTokens += sample.O200KTokens
		totalClaudeContentTokens += sample.ClaudeContentTokens
	}

	var errorTotal float64
	var errorSamples int
	for _, sample := range measurements {
		if sample.ClaudeContentTokens == 0 {
			continue
		}
		otherO200KTokens := totalO200KTokens - sample.O200KTokens
		if otherO200KTokens <= 0 {
			return 0, false
		}
		otherClaudeContentTokens := totalClaudeContentTokens - sample.ClaudeContentTokens
		factor := float64(otherClaudeContentTokens) / float64(otherO200KTokens)
		prediction := roundedPrediction(sample.O200KTokens, factor)
		errorTotal += math.Abs(float64(prediction-sample.ClaudeContentTokens)) / float64(sample.ClaudeContentTokens) * 100
		errorSamples++
	}
	if errorSamples == 0 {
		return 0, false
	}
	return errorTotal / float64(errorSamples), true
}

func summarizeLanguage(measurements []measurement, globalFactor float64) languageRatioSummary {
	summary := languageRatioSummary{ratioSummary: summarize(measurements)}
	globalErrors := evaluateFactor(measurements, globalFactor)
	summary.GlobalFactorSignedAggregatePercentError = globalErrors.AggregatePercentError
	summary.GlobalFactorMeanAbsolutePercentError = globalErrors.MeanAbsolutePercentError
	if leaveOneOutMAPE, ok := leaveOneOutMeanAbsolutePercentError(measurements); ok {
		summary.LeaveOneOutMeanAbsolutePercentError = &leaveOneOutMAPE
	}
	return summary
}

func finalizeModelReport(report *modelReport) {
	report.Global = summarize(report.Samples)
	byLanguage := make(map[string][]measurement)
	for _, sample := range report.Samples {
		byLanguage[sample.Language] = append(byLanguage[sample.Language], sample)
	}
	report.PerLanguage = make(map[string]languageRatioSummary, len(byLanguage))
	for language, samples := range byLanguage {
		report.PerLanguage[language] = summarizeLanguage(samples, report.Global.Factor)
	}
}

func finalizeHeldOutEvaluation(report *modelReport, samples []measurement) error {
	if len(samples) == 0 {
		report.HeldOut = nil
		return nil
	}

	factorSource := "training global factor (no production mapping)"
	globalFactor := report.Global.Factor
	var overrides []tokenizer.CalibrationOverride
	if report.Label == tokenizer.NameClaude || report.Label == tokenizer.NameClaudeLegacy {
		_, metadata, err := tokenizer.New(report.Label)
		if err != nil {
			return fmt.Errorf("load production factors for held-out %s evaluation: %w", report.Label, err)
		}
		factorSource = "production calibration factors"
		globalFactor = metadata.CalibrationFactor
		overrides = append([]tokenizer.CalibrationOverride(nil), metadata.CalibrationOverrides...)
	}

	byLanguage := make(map[string][]measurement)
	for _, sample := range samples {
		byLanguage[sample.Language] = append(byLanguage[sample.Language], sample)
	}
	perLanguage := make(map[string]evaluationSummary, len(byLanguage))
	for language, languageSamples := range byLanguage {
		perLanguage[language] = evaluateCalibrationFactors(languageSamples, globalFactor, overrides)
	}
	report.HeldOut = &heldOutEvaluation{
		FactorSource:         factorSource,
		GlobalFactor:         globalFactor,
		CalibrationOverrides: overrides,
		Summary:              evaluateCalibrationFactors(samples, globalFactor, overrides),
		PerLanguage:          perLanguage,
		Samples:              append([]measurement(nil), samples...),
	}
	return nil
}

func evaluateCalibrationFactors(measurements []measurement, globalFactor float64, overrides []tokenizer.CalibrationOverride) evaluationSummary {
	factors := make(map[string]float64, len(overrides))
	for _, override := range overrides {
		factors[override.Language] = override.Factor
	}

	summary := evaluationSummary{Samples: len(measurements)}
	var absolutePercentErrorTotal float64
	var errorSamples int
	for _, sample := range measurements {
		factor := globalFactor
		if override, ok := factors[sample.Language]; ok {
			factor = override
		}
		prediction := roundedPrediction(sample.O200KTokens, factor)
		summary.O200KTokens += sample.O200KTokens
		summary.ClaudeContentTokens += sample.ClaudeContentTokens
		summary.PredictedTokens += prediction
		if sample.ClaudeContentTokens > 0 {
			absolutePercentErrorTotal += math.Abs(float64(prediction-sample.ClaudeContentTokens)) / float64(sample.ClaudeContentTokens) * 100
			errorSamples++
		}
	}
	if summary.ClaudeContentTokens > 0 {
		summary.SignedAggregateError = float64(summary.PredictedTokens-summary.ClaudeContentTokens) / float64(summary.ClaudeContentTokens) * 100
	}
	if errorSamples > 0 {
		summary.MeanAbsolutePercentError = absolutePercentErrorTotal / float64(errorSamples)
	}
	return summary
}

func calibrationFactorForLanguage(globalFactor float64, overrides []tokenizer.CalibrationOverride, language string) (float64, string) {
	for _, override := range overrides {
		if override.Language == language {
			return override.Factor, "language override"
		}
	}
	return globalFactor, "global fallback"
}

func writeReports(outputDirectory string, report calibrationReport) (string, string, error) {
	if err := os.MkdirAll(outputDirectory, 0o755); err != nil {
		return "", "", fmt.Errorf("create output directory: %w", err)
	}

	jsonData, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return "", "", fmt.Errorf("marshal calibration JSON: %w", err)
	}
	jsonData = append(jsonData, '\n')
	jsonPath := filepath.Join(outputDirectory, "calibration.json")
	if err := os.WriteFile(jsonPath, jsonData, 0o644); err != nil {
		return "", "", fmt.Errorf("write calibration JSON: %w", err)
	}

	markdownPath := filepath.Join(outputDirectory, "calibration.md")
	if err := os.WriteFile(markdownPath, []byte(renderMarkdown(report)), 0o644); err != nil {
		return "", "", fmt.Errorf("write calibration Markdown: %w", err)
	}
	return jsonPath, markdownPath, nil
}

func renderMarkdown(report calibrationReport) string {
	var output strings.Builder
	fmt.Fprintf(&output, "# Claude tokenizer calibration\n\n")
	fmt.Fprintf(&output, "Generated: %s\n\n", report.GeneratedAt.Format(time.RFC3339))
	fmt.Fprintf(&output, "Method: %s\n\n", report.Method)
	fmt.Fprintf(&output, "| Generation | Model | Framing baseline | Samples | o200k tokens | Claude content tokens | Factor | MAPE |\n")
	fmt.Fprintf(&output, "|---|---|---:|---:|---:|---:|---:|---:|\n")
	for _, model := range report.Models {
		fmt.Fprintf(&output, "| %s | `%s` | %d | %d | %d | %d | %.6f | %.2f%% |\n",
			model.Label, model.Model, model.FramingBaseline, model.Global.Samples,
			model.Global.O200KTokens, model.Global.ClaudeContentTokens,
			model.Global.Factor, model.Global.MeanAbsolutePercentError)
	}

	for _, model := range report.Models {
		fmt.Fprintf(&output, "\n## %s (`%s`)\n\n", model.Label, model.Model)
		fmt.Fprintf(&output, "| Language | Samples | o200k tokens | Claude content tokens | Fitted factor | Mean ratio | Median ratio | Fitted-factor MAPE | Global-factor signed aggregate error | Global-factor MAPE | LOO language MAPE |\n")
		fmt.Fprintf(&output, "|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|\n")
		languages := make([]string, 0, len(model.PerLanguage))
		for language := range model.PerLanguage {
			languages = append(languages, language)
		}
		sort.Strings(languages)
		for _, language := range languages {
			summary := model.PerLanguage[language]
			fmt.Fprintf(&output, "| %s | %d | %d | %d | %.6f | %.6f | %.6f | %.2f%% | %+.2f%% | %.2f%% | %s |\n",
				language, summary.Samples, summary.O200KTokens, summary.ClaudeContentTokens,
				summary.Factor, summary.MeanPerFileRatio, summary.MedianPerFileRatio,
				summary.MeanAbsolutePercentError, summary.GlobalFactorSignedAggregatePercentError,
				summary.GlobalFactorMeanAbsolutePercentError,
				formatLeaveOneOutMAPE(summary.LeaveOneOutMeanAbsolutePercentError, summary.Samples))
		}

		if model.HeldOut != nil {
			heldOut := model.HeldOut
			fmt.Fprintf(&output, "\n### Held-out evaluation\n\n")
			fmt.Fprintf(&output, "These samples were not used to fit or select calibration factors. Factor source: %s.\n\n", heldOut.FactorSource)
			fmt.Fprintf(&output, "| Language | Samples | Factor used | Factor basis | o200k tokens | Claude content tokens | Predicted tokens | Signed aggregate error | MAPE |\n")
			fmt.Fprintf(&output, "|---|---:|---:|---|---:|---:|---:|---:|---:|\n")
			heldOutLanguages := make([]string, 0, len(heldOut.PerLanguage))
			for language := range heldOut.PerLanguage {
				heldOutLanguages = append(heldOutLanguages, language)
			}
			sort.Strings(heldOutLanguages)
			for _, language := range heldOutLanguages {
				summary := heldOut.PerLanguage[language]
				factor, basis := calibrationFactorForLanguage(heldOut.GlobalFactor, heldOut.CalibrationOverrides, language)
				fmt.Fprintf(&output, "| %s | %d | %.6f | %s | %d | %d | %d | %+.2f%% | %.2f%% |\n",
					language, summary.Samples, factor, basis, summary.O200KTokens,
					summary.ClaudeContentTokens, summary.PredictedTokens,
					summary.SignedAggregateError, summary.MeanAbsolutePercentError)
			}
			fmt.Fprintf(&output, "| **Overall** | **%d** |  |  | **%d** | **%d** | **%d** | **%+.2f%%** | **%.2f%%** |\n",
				heldOut.Summary.Samples, heldOut.Summary.O200KTokens,
				heldOut.Summary.ClaudeContentTokens, heldOut.Summary.PredictedTokens,
				heldOut.Summary.SignedAggregateError, heldOut.Summary.MeanAbsolutePercentError)

			fmt.Fprintf(&output, "\n<details>\n<summary>Held-out per-file measurements</summary>\n\n")
			fmt.Fprintf(&output, "| File | Language | Bytes | Truncated | o200k | Claude content | Ratio |\n")
			fmt.Fprintf(&output, "|---|---|---:|:---:|---:|---:|---:|\n")
			for _, sample := range heldOut.Samples {
				fmt.Fprintf(&output, "| `%s` | %s | %d | %t | %d | %d | %.6f |\n",
					escapeMarkdown(sample.Path), sample.Language, sample.Bytes, sample.Truncated,
					sample.O200KTokens, sample.ClaudeContentTokens, sample.ClaudePerO200KRatio)
			}
			fmt.Fprintf(&output, "\n</details>\n")
		}

		fmt.Fprintf(&output, "\n<details>\n<summary>Per-file measurements</summary>\n\n")
		fmt.Fprintf(&output, "| File | Language | Bytes | Truncated | o200k | Claude content | Ratio |\n")
		fmt.Fprintf(&output, "|---|---|---:|:---:|---:|---:|---:|\n")
		for _, sample := range model.Samples {
			fmt.Fprintf(&output, "| `%s` | %s | %d | %t | %d | %d | %.6f |\n",
				escapeMarkdown(sample.Path), sample.Language, sample.Bytes, sample.Truncated,
				sample.O200KTokens, sample.ClaudeContentTokens, sample.ClaudePerO200KRatio)
		}
		fmt.Fprintf(&output, "\n</details>\n")
	}
	return output.String()
}

func formatLeaveOneOutMAPE(value *float64, samples int) string {
	if value != nil {
		return fmt.Sprintf("%.2f%%", *value)
	}
	if samples < 2 {
		return "N/A (<2 samples)"
	}
	return "N/A"
}

func escapeMarkdown(value string) string {
	return strings.ReplaceAll(value, "|", "\\|")
}
