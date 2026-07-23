package app

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/shaunobi/tloc/internal/analyze"
	"github.com/shaunobi/tloc/internal/tokenizer"
)

func TestPreflightOutputRefusesExistingFileWithoutTruncating(t *testing.T) {
	path := filepath.Join(t.TempDir(), "report.json")
	original := []byte("existing report")
	if err := os.WriteFile(path, original, 0o600); err != nil {
		t.Fatal(err)
	}

	if _, err := preflightOutput(path, false); err == nil || !strings.Contains(err.Error(), "--force") {
		t.Fatalf("preflight error = %v, want no-clobber guidance", err)
	}
	assertFileContent(t, path, original)

	plan, err := preflightOutput(path, true)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = plan.close() }()
	if plan.path != filepath.Clean(path) {
		t.Fatalf("resolved path = %q, want %q", plan.path, filepath.Clean(path))
	}
	assertFileContent(t, path, original)
}

func TestMainPreflightsExistingOutputBeforeScanning(t *testing.T) {
	path := filepath.Join(t.TempDir(), "report.json")
	original := []byte("do not truncate")
	if err := os.WriteFile(path, original, 0o600); err != nil {
		t.Fatal(err)
	}
	called := false
	analyzer := func([]string, tokenizer.Counter, analyze.Options) ([]analyze.InputRoot, []analyze.FileRecord, []analyze.ScanWarning, error) {
		called = true
		return nil, nil, nil, nil
	}

	var stdout, stderr bytes.Buffer
	if code := mainWithAnalyzer([]string{"--output", path, "."}, &stdout, &stderr, analyzer); code != 1 {
		t.Fatalf("code=%d stderr=%q", code, stderr.String())
	}
	if called {
		t.Fatal("analyzer ran before the output no-clobber check")
	}
	assertFileContent(t, path, original)
}

func TestWriteOutputRefusesFileCreatedAfterPreflight(t *testing.T) {
	for _, test := range []struct {
		name      string
		force     bool
		wantError string
	}{
		{name: "default", wantError: "pass --force"},
		{name: "force requested while absent", force: true, wantError: "preflighted as an existing destination"},
	} {
		t.Run(test.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "report.json")
			plan, err := preflightOutput(path, test.force)
			if err != nil {
				t.Fatal(err)
			}
			racingContent := []byte("created by another process")
			if err := os.WriteFile(path, racingContent, 0o600); err != nil {
				t.Fatal(err)
			}
			err = writeOutput(&plan, []byte("new report"))
			if err == nil || !strings.Contains(err.Error(), test.wantError) {
				t.Fatalf("race-created output error = %v, want %q", err, test.wantError)
			}
			assertFileContent(t, path, racingContent)
		})
	}
}

func TestWriteOutputRefusesIdentityMismatch(t *testing.T) {
	dir := t.TempDir()
	originalPath := filepath.Join(dir, "original-report.txt")
	replacementPath := filepath.Join(dir, "replacement-source.go")
	originalContent := []byte("old report")
	replacementContent := []byte("package replacement\n")
	if err := os.WriteFile(originalPath, originalContent, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(replacementPath, replacementContent, 0o600); err != nil {
		t.Fatal(err)
	}
	original, err := os.OpenFile(originalPath, os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	plan := outputPlan{path: replacementPath, existing: original}
	defer func() { _ = plan.close() }()

	if err := writeOutput(&plan, []byte("new report")); err == nil || !strings.Contains(err.Error(), "changed after preflight") {
		t.Fatalf("identity-mismatch error = %v", err)
	}
	assertFileContent(t, originalPath, originalContent)
	assertFileContent(t, replacementPath, replacementContent)
}

func TestMainForceRefusesOutputRetargetedDuringScan(t *testing.T) {
	dir := t.TempDir()
	sourcePath := filepath.Join(dir, "source.go")
	outputPath := filepath.Join(dir, "report.txt")
	sourceContent := []byte("package source\n")
	if err := os.WriteFile(sourcePath, sourceContent, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(outputPath, []byte("old report"), 0o600); err != nil {
		t.Fatal(err)
	}

	analyzer := func([]string, tokenizer.Counter, analyze.Options) ([]analyze.InputRoot, []analyze.FileRecord, []analyze.ScanWarning, error) {
		if err := os.Remove(outputPath); err != nil {
			t.Skipf("platform would not allow swapping an open output path: %v", err)
		}
		if err := os.Link(sourcePath, outputPath); err != nil {
			t.Skipf("hard links unavailable for output-swap regression: %v", err)
		}
		return []analyze.InputRoot{{
				ID:    0,
				Given: sourcePath,
				Abs:   sourcePath,
				Kind:  analyze.InputFile,
			}}, []analyze.FileRecord{{
				InputID:  0,
				Path:     sourcePath,
				RelPath:  filepath.Base(sourcePath),
				Language: "Go",
				Metrics:  analyze.Metrics{Files: 1, Lines: 1, Code: 1},
			}}, nil, nil
	}

	var stdout, stderr bytes.Buffer
	code := mainWithAnalyzer([]string{"--force", "--output", outputPath, sourcePath}, &stdout, &stderr, analyzer)
	if code != 1 {
		t.Fatalf("code=%d, want 1; stderr=%q", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "changed after preflight") {
		t.Fatalf("stderr=%q", stderr.String())
	}
	assertFileContent(t, sourcePath, sourceContent)
	assertFileContent(t, outputPath, sourceContent)
}

func TestPreflightOutputRejectsWindowsCaseVariant(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows paths are case-insensitive")
	}
	path := filepath.Join(t.TempDir(), "Report.JSON")
	original := []byte("source-like content")
	if err := os.WriteFile(path, original, 0o600); err != nil {
		t.Fatal(err)
	}
	variant := filepath.Join(filepath.Dir(path), "report.json")
	if _, err := preflightOutput(variant, false); err == nil {
		t.Fatalf("case-variant output %q was not recognized as existing", variant)
	}
	assertFileContent(t, path, original)
}

func assertFileContent(t *testing.T, path string, want []byte) {
	t.Helper()
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("%s content = %q, want %q", path, got, want)
	}
}
