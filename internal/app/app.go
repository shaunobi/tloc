package app

import (
	"bytes"
	"fmt"
	"io"

	"github.com/shaunobi/tloc/internal/aggregate"
	"github.com/shaunobi/tloc/internal/analyze"
	"github.com/shaunobi/tloc/internal/buildinfo"
	"github.com/shaunobi/tloc/internal/model"
	"github.com/shaunobi/tloc/internal/output"
	"github.com/shaunobi/tloc/internal/tokenizer"
)

// Main runs the CLI and returns a process exit code.
func Main(args []string, stdout, stderr io.Writer) int {
	return mainWithAnalyzer(args, stdout, stderr, analyze.Run)
}

type analyzerFunc func([]string, tokenizer.Counter, analyze.Options) ([]analyze.InputRoot, []analyze.FileRecord, []analyze.ScanWarning, error)

func mainWithAnalyzer(args []string, stdout, stderr io.Writer, runAnalyzer analyzerFunc) int {
	cfg, err := parseConfig(args, stdout)
	if err != nil {
		fmt.Fprintf(stderr, "tloc: %v\n", err)
		return 2
	}
	if cfg.help {
		return 0
	}
	if cfg.version {
		fmt.Fprintf(stdout, "tloc %s\n", buildinfo.Version())
		return 0
	}

	outputPlan, err := preflightOutput(cfg.output, cfg.force)
	if err != nil {
		fmt.Fprintf(stderr, "tloc: %v\n", err)
		return 1
	}
	defer func() { _ = outputPlan.close() }()

	counter, tokenizerMetadata, err := tokenizer.New(cfg.tokenizer)
	if err != nil {
		fmt.Fprintf(stderr, "tloc: %v\n", err)
		return 1
	}
	var excludedFiles []string
	if outputPlan.path != "" {
		excludedFiles = []string{outputPlan.path}
	}
	inputs, files, warnings, err := runAnalyzer(cfg.paths, counter, analyze.Options{
		IncludeExt:   cfg.includeExt,
		ExcludeExt:   cfg.excludeExt,
		ExcludeDir:   cfg.excludeDir,
		ExcludeFiles: excludedFiles,
		MaxFileBytes: cfg.maxFileBytes,
		NoIgnore:     cfg.noIgnore,
		NoGitignore:  cfg.noGitignore,
	})
	if err != nil {
		fmt.Fprintf(stderr, "tloc: %v\n", err)
		return 1
	}
	for _, warning := range warnings {
		fmt.Fprintf(stderr, "tloc: warning: incomplete scan: %v\n", warning)
	}

	view := model.ViewLanguage
	if cfg.byFile {
		view = model.ViewFile
	} else if cfg.byFolder {
		view = model.ViewFolder
	}
	reportMetadata := toModelMetadata(tokenizerMetadata)
	reportMetadata.Version = buildinfo.Version()
	reportMetadata.Complete = len(warnings) == 0
	reportMetadata.Skipped = toModelSkippedEntries(warnings)
	report, err := aggregate.Build(toModelInputs(inputs), toModelFiles(files), view, model.SortKey(cfg.sort), reportMetadata)
	if err != nil {
		fmt.Fprintf(stderr, "tloc: build report: %v\n", err)
		return 1
	}

	var rendered bytes.Buffer
	if err := output.Write(&rendered, report, view, output.Format(cfg.format)); err != nil {
		fmt.Fprintf(stderr, "tloc: render report: %v\n", err)
		return 1
	}
	if cfg.output == "" {
		if _, err := stdout.Write(rendered.Bytes()); err != nil {
			fmt.Fprintf(stderr, "tloc: write stdout: %v\n", err)
			return 1
		}
		if len(warnings) > 0 {
			return 1
		}
		return 0
	}
	if err := writeOutput(&outputPlan, rendered.Bytes()); err != nil {
		fmt.Fprintf(stderr, "tloc: write %q: %v\n", cfg.output, err)
		return 1
	}
	if len(warnings) > 0 {
		return 1
	}
	return 0
}

func toModelMetadata(metadata tokenizer.Metadata) model.Metadata {
	overrides := make([]model.CalibrationOverride, len(metadata.CalibrationOverrides))
	for index, override := range metadata.CalibrationOverrides {
		overrides[index] = model.CalibrationOverride{
			Language: override.Language,
			Factor:   override.Factor,
		}
	}
	return model.Metadata{
		Tokenizer:            metadata.Name,
		CalibrationFactor:    metadata.CalibrationFactor,
		CalibrationOverrides: overrides,
		Estimated:            metadata.Estimated,
	}
}

func toModelSkippedEntries(warnings []analyze.ScanWarning) []model.SkippedEntry {
	result := make([]model.SkippedEntry, 0, len(warnings))
	for _, warning := range warnings {
		result = append(result, model.SkippedEntry{
			Stage: warning.Stage,
			Path:  warning.Path,
			Error: warning.Message,
		})
	}
	return result
}

func toModelInputs(inputs []analyze.InputRoot) []model.InputRoot {
	result := make([]model.InputRoot, 0, len(inputs))
	for _, input := range inputs {
		kind := model.InputFile
		if input.Kind == analyze.InputDirectory {
			kind = model.InputDirectory
		}
		result = append(result, model.InputRoot{
			ID:    input.ID,
			Given: input.Given,
			Abs:   input.Abs,
			Kind:  kind,
		})
	}
	return result
}

func toModelFiles(files []analyze.FileRecord) []model.FileRecord {
	result := make([]model.FileRecord, 0, len(files))
	for _, file := range files {
		result = append(result, model.FileRecord{
			InputID:  file.InputID,
			Path:     file.Path,
			RelPath:  file.RelPath,
			Language: file.Language,
			Metrics: model.Metrics{
				Files:      file.Metrics.Files,
				Lines:      file.Metrics.Lines,
				Code:       file.Metrics.Code,
				Comments:   file.Metrics.Comments,
				Blanks:     file.Metrics.Blanks,
				Complexity: file.Metrics.Complexity,
				Bytes:      file.Metrics.Bytes,
				Tokens:     file.Metrics.Tokens,
			},
		})
	}
	return result
}
