package app

import (
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/spf13/pflag"

	"github.com/shaunobi/tloc/internal/tokenizer"
)

const defaultMaxFileBytes int64 = 1_000_000

type config struct {
	paths        []string
	tokenizer    string
	byFile       bool
	byFolder     bool
	format       string
	output       string
	sort         string
	includeExt   []string
	excludeExt   []string
	excludeDir   []string
	maxFileBytes int64
	noIgnore     bool
	noGitignore  bool
	version      bool
	help         bool
}

func parseConfig(args []string, helpWriter io.Writer) (config, error) {
	cfg := config{}
	flags := pflag.NewFlagSet("tloc", pflag.ContinueOnError)
	flags.SetOutput(helpWriter)
	flags.SetInterspersed(true)
	flags.SortFlags = false
	flags.StringVar(&cfg.tokenizer, "tokenizer", "o200k", "tokenizer: claude, claude-legacy, o200k, or codex")
	flags.BoolVar(&cfg.byFile, "by-file", false, "report one row per file")
	flags.BoolVar(&cfg.byFolder, "by-folder", false, "report a cumulative folder tree")
	flags.StringVarP(&cfg.format, "format", "f", "tabular", "output format: tabular, json, or csv")
	flags.StringVarP(&cfg.output, "output", "o", "", "write output to file instead of stdout")
	flags.StringVar(&cfg.sort, "sort", "tokens", "sort by: tokens, code, lines, files, or name")
	flags.StringSliceVar(&cfg.includeExt, "include-ext", nil, "only include extensions (comma-separated)")
	flags.StringSliceVar(&cfg.excludeExt, "exclude-ext", nil, "exclude extensions (comma-separated; takes precedence)")
	flags.StringSliceVar(&cfg.excludeDir, "exclude-dir", nil, "exclude directory names (comma-separated)")
	flags.Int64Var(&cfg.maxFileBytes, "max-file-bytes", defaultMaxFileBytes, "skip files larger than this many bytes")
	flags.BoolVar(&cfg.noIgnore, "no-ignore", false, "disable .ignore and .sccignore handling")
	flags.BoolVar(&cfg.noGitignore, "no-gitignore", false, "disable .gitignore handling")
	flags.BoolVar(&cfg.version, "version", false, "print version and exit")
	flags.BoolVarP(&cfg.help, "help", "h", false, "show help")
	flags.Usage = func() { writeUsage(helpWriter, flags) }

	if err := flags.Parse(args); err != nil {
		return config{}, err
	}
	cfg.paths = flags.Args()
	if len(cfg.paths) == 0 {
		cfg.paths = []string{"."}
	}

	cfg.tokenizer = strings.ToLower(strings.TrimSpace(cfg.tokenizer))
	cfg.format = strings.ToLower(strings.TrimSpace(cfg.format))
	cfg.sort = strings.ToLower(strings.TrimSpace(cfg.sort))
	cfg.includeExt = normalizeList(cfg.includeExt, true)
	cfg.excludeExt = normalizeList(cfg.excludeExt, true)
	cfg.excludeDir = normalizeList(cfg.excludeDir, false)

	if cfg.help {
		flags.Usage()
		return cfg, nil
	}
	if cfg.version {
		return cfg, nil
	}
	if cfg.byFile && cfg.byFolder {
		return config{}, errors.New("--by-file and --by-folder are mutually exclusive")
	}
	if !slices.Contains([]string{"o200k", "codex", "claude", "claude-legacy"}, cfg.tokenizer) {
		return config{}, fmt.Errorf("invalid --tokenizer %q (want claude, claude-legacy, o200k, or codex)", cfg.tokenizer)
	}
	if !slices.Contains([]string{"tabular", "json", "csv"}, cfg.format) {
		return config{}, fmt.Errorf("invalid --format %q (want tabular, json, or csv)", cfg.format)
	}
	if !slices.Contains([]string{"tokens", "code", "lines", "files", "name"}, cfg.sort) {
		return config{}, fmt.Errorf("invalid --sort %q (want tokens, code, lines, files, or name)", cfg.sort)
	}
	if cfg.maxFileBytes <= 0 {
		return config{}, errors.New("--max-file-bytes must be greater than zero")
	}
	return cfg, nil
}

func normalizeList(values []string, trimDot bool) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		for part := range strings.SplitSeq(value, ",") {
			part = strings.TrimSpace(part)
			if trimDot {
				part = strings.TrimPrefix(part, ".")
				part = strings.ToLower(part)
			}
			if part == "" {
				continue
			}
			if _, ok := seen[part]; ok {
				continue
			}
			seen[part] = struct{}{}
			result = append(result, part)
		}
	}
	return result
}

func writeUsage(w io.Writer, flags *pflag.FlagSet) {
	fmt.Fprintln(w, "tloc counts source lines and LLM tokens in one pass.")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  tloc [flags] [paths...]")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Flags:")
	flags.PrintDefaults()
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Tokenizer accuracy:")
	fmt.Fprintln(w, "  o200k and codex count the o200k encoding exactly.")
	fmt.Fprintln(w, "  claude and claude-legacy are estimates and may differ from Anthropic's")
	fmt.Fprintln(w, "  count_tokens API.")
	pending := make([]string, 0, 2)
	if !tokenizer.ClaudeCurrentCalibrationReady {
		pending = append(pending, tokenizer.NameClaude)
	}
	if !tokenizer.ClaudeLegacyCalibrationReady {
		pending = append(pending, tokenizer.NameClaudeLegacy)
	}
	if len(pending) > 0 {
		fmt.Fprintf(w, "  Unavailable pending calibration in this build: %s.\n", strings.Join(pending, ", "))
	} else {
		fmt.Fprintln(w, "  Both embedded Claude calibrations are enabled in this build.")
	}
}
