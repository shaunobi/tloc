package app

import (
	"bytes"
	"strings"
	"testing"

	"github.com/shaunobi/tloc/internal/tokenizer"
)

func TestParseConfigDefaults(t *testing.T) {
	cfg, err := parseConfig(nil, &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.paths) != 1 || cfg.paths[0] != "." {
		t.Fatalf("paths = %#v", cfg.paths)
	}
	if cfg.tokenizer != "o200k" || cfg.format != "tabular" || cfg.sort != "tokens" {
		t.Fatalf("unexpected defaults: %#v", cfg)
	}
	if cfg.maxFileBytes != defaultMaxFileBytes {
		t.Fatalf("maxFileBytes = %d", cfg.maxFileBytes)
	}
}

func TestParseConfigInterspersedAndLists(t *testing.T) {
	cfg, err := parseConfig([]string{"src", "--include-ext", ".Go,TS", "--include-ext", "go", "--exclude-dir", "vendor,generated", "--format", "JSON"}, &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.paths) != 1 || cfg.paths[0] != "src" {
		t.Fatalf("paths = %#v", cfg.paths)
	}
	if got := strings.Join(cfg.includeExt, ","); got != "go,ts" {
		t.Fatalf("includeExt = %q", got)
	}
	if got := strings.Join(cfg.excludeDir, ","); got != "vendor,generated" {
		t.Fatalf("excludeDir = %q", got)
	}
	if cfg.format != "json" {
		t.Fatalf("format = %q", cfg.format)
	}
}

func TestParseConfigValidation(t *testing.T) {
	tests := [][]string{
		{"--by-file", "--by-folder"},
		{"--tokenizer", "gpt2"},
		{"--format", "xml"},
		{"--sort", "bytes"},
		{"--max-file-bytes", "0"},
		{"--force"},
	}
	for _, args := range tests {
		if _, err := parseConfig(args, &bytes.Buffer{}); err == nil {
			t.Fatalf("parseConfig(%v) succeeded", args)
		}
	}
}

func TestUsageCoversAccuracyAndFlags(t *testing.T) {
	var output bytes.Buffer
	cfg, err := parseConfig([]string{"--help"}, &output)
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.help {
		t.Fatal("help was not set")
	}
	for _, want := range []string{
		"--tokenizer", "--by-file", "--by-folder", "--format", "--output", "--force", "--sort",
		"--include-ext", "--exclude-ext", "--exclude-dir", "--max-file-bytes",
		"--no-ignore", "--no-gitignore", "--version", "--help",
	} {
		if !strings.Contains(output.String(), want) {
			t.Fatalf("usage missing %q:\n%s", want, output.String())
		}
	}
	calibrationText := "Both embedded Claude calibrations are enabled"
	if !tokenizer.ClaudeCurrentCalibrationReady || !tokenizer.ClaudeLegacyCalibrationReady {
		calibrationText = "Unavailable pending calibration"
	}
	if !strings.Contains(output.String(), calibrationText) {
		t.Fatalf("usage missing %q:\n%s", calibrationText, output.String())
	}
	if !strings.Contains(output.String(), "10% error on represented languages") {
		t.Fatalf("usage missing roughly 10%% accuracy target:\n%s", output.String())
	}
	if !strings.Contains(output.String(), "without direct validation") {
		t.Fatalf("usage missing unrepresented-language caveat:\n%s", output.String())
	}
}
