package app

import (
	"bufio"
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/shaunobi/tloc/internal/aggregate"
	"github.com/shaunobi/tloc/internal/analyze"
	"github.com/shaunobi/tloc/internal/model"
	"github.com/shaunobi/tloc/internal/tokenizer"
)

func TestIntegrationFixtureMatchesSCCOracle(t *testing.T) {
	fixture := filepath.Join("..", "..", "testdata", "integration", "project")
	counter, _, err := tokenizer.New(tokenizer.NameO200K)
	if err != nil {
		t.Fatal(err)
	}
	inputs, files, err := analyze.Run([]string{fixture}, counter, analyze.Options{Workers: 4})
	if err != nil {
		t.Fatal(err)
	}

	gotPaths := make([]string, 0, len(files))
	for _, file := range files {
		gotPaths = append(gotPaths, file.RelPath)
	}
	slices.Sort(gotPaths)
	wantPaths := readOracle(t, filepath.Join("..", "..", "testdata", "integration", "scc-files.txt"))
	if !slices.Equal(gotPaths, wantPaths) {
		t.Fatalf("selected paths differ from scc oracle\n got: %v\nwant: %v", gotPaths, wantPaths)
	}

	report, err := aggregate.Build(toModelInputs(inputs), toModelFiles(files), model.ViewFolder, model.SortName, model.Metadata{Tokenizer: "o200k", CalibrationFactor: 1})
	if err != nil {
		t.Fatal(err)
	}
	wantLanguages := map[string]int64{"Go": 2, "Markdown": 1, "Python": 1, "SQL": 1, "TypeScript": 1}
	if len(report.Languages) != len(wantLanguages) {
		t.Fatalf("languages = %#v", report.Languages)
	}
	for _, language := range report.Languages {
		if wantLanguages[language.Language] != language.Metrics.Files {
			t.Fatalf("language row = %#v", language)
		}
	}
	if report.Totals.Files != int64(len(wantPaths)) || report.Totals.Tokens == 0 || report.Totals.Code == 0 {
		t.Fatalf("totals = %#v", report.Totals)
	}
	wantFolders := map[string]int64{"(root files)": 2, "db": 1, "src": 2, "web": 1}
	for _, folder := range report.Folders {
		if want, ok := wantFolders[folder.Name]; ok && folder.Metrics.Files != want {
			t.Fatalf("folder row = %#v, want files=%d", folder, want)
		}
		delete(wantFolders, folder.Name)
	}
	if len(wantFolders) != 0 {
		t.Fatalf("missing folder rows: %v; report=%#v", wantFolders, report.Folders)
	}
}

func TestIntegrationDisableIgnoreHandling(t *testing.T) {
	fixture := filepath.Join("..", "..", "testdata", "integration", "project")
	counter, _, err := tokenizer.New(tokenizer.NameO200K)
	if err != nil {
		t.Fatal(err)
	}
	_, files, err := analyze.Run([]string{fixture}, counter, analyze.Options{Workers: 4, NoIgnore: true, NoGitignore: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 10 {
		paths := make([]string, 0, len(files))
		for _, file := range files {
			paths = append(paths, file.RelPath)
		}
		t.Fatalf("counted %d files, want 10: %v", len(files), paths)
	}
}

func TestIntegrationCLIJSONAndOutputFile(t *testing.T) {
	fixture := filepath.Join("..", "..", "testdata", "integration", "project")
	outputPath := filepath.Join(t.TempDir(), "report.json")
	var stdout, stderr bytes.Buffer
	code := Main([]string{"--format", "json", "--by-file", "--output", outputPath, fixture}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("code=%d stderr=%q", code, stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout=%q", stdout.String())
	}
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`"languages"`, `"files"`, `"totals"`, `"metadata"`, `"tokenizer": "o200k"`} {
		if !bytes.Contains(content, []byte(want)) {
			t.Fatalf("JSON missing %s:\n%s", want, content)
		}
	}
}

func TestIntegrationOutputInsideScanIsStableAndCannotOverwriteInput(t *testing.T) {
	dir := t.TempDir()
	sourcePath := filepath.Join(dir, "main.go")
	outputPath := filepath.Join(dir, "report.json")
	content := []byte("package main\n")
	if err := os.WriteFile(sourcePath, content, 0o600); err != nil {
		t.Fatal(err)
	}

	for run := 1; run <= 2; run++ {
		var stdout, stderr bytes.Buffer
		code := Main([]string{"--format", "json", "--by-file", "--output", outputPath, dir}, &stdout, &stderr)
		if code != 0 {
			t.Fatalf("run %d code=%d stderr=%q", run, code, stderr.String())
		}
		reportBytes, err := os.ReadFile(outputPath)
		if err != nil {
			t.Fatal(err)
		}
		var report struct {
			Totals struct {
				Files int64 `json:"files"`
			} `json:"totals"`
		}
		if err := json.Unmarshal(reportBytes, &report); err != nil {
			t.Fatal(err)
		}
		if report.Totals.Files != 1 {
			t.Fatalf("run %d files=%d, want 1", run, report.Totals.Files)
		}
	}

	var stdout, stderr bytes.Buffer
	if code := Main([]string{"--output", sourcePath, sourcePath}, &stdout, &stderr); code == 0 {
		t.Fatalf("input/output collision succeeded")
	}
	got, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, content) {
		t.Fatalf("input was overwritten: %q", got)
	}

	aliasPath := filepath.Join(dir, "source-alias.go")
	if err := os.Link(sourcePath, aliasPath); err == nil {
		stdout.Reset()
		stderr.Reset()
		if code := Main([]string{"--output", aliasPath, sourcePath}, &stdout, &stderr); code == 0 {
			t.Fatalf("hard-linked input/output collision succeeded")
		}
		got, err = os.ReadFile(sourcePath)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(got, content) {
			t.Fatalf("hard-linked input was overwritten: %q", got)
		}

		stdout.Reset()
		stderr.Reset()
		if code := Main([]string{"--output", aliasPath, dir}, &stdout, &stderr); code == 0 {
			t.Fatalf("directory scan with hard-linked output collision succeeded")
		}
		got, err = os.ReadFile(sourcePath)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(got, content) {
			t.Fatalf("directory-scan alias overwrote source: %q", got)
		}
	}
}

func readOracle(t *testing.T, path string) []string {
	t.Helper()
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if line := strings.TrimSpace(scanner.Text()); line != "" {
			lines = append(lines, filepath.ToSlash(line))
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	slices.Sort(lines)
	return lines
}
