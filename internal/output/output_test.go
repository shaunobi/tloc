package output

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/mattn/go-runewidth"

	"github.com/shaunobi/tloc/internal/model"
)

func TestGoldenAllFormatsAndViews(t *testing.T) {
	report := goldenReport()
	tests := []struct {
		name   string
		view   model.View
		format Format
		file   string
	}{
		{"summary tabular", model.ViewLanguage, FormatTabular, "summary.tabular"},
		{"files tabular", model.ViewFile, FormatTabular, "files.tabular"},
		{"folders tabular", model.ViewFolder, FormatTabular, "folders.tabular"},
		{"summary json", model.ViewLanguage, FormatJSON, "summary.json"},
		{"files json", model.ViewFile, FormatJSON, "files.json"},
		{"folders json", model.ViewFolder, FormatJSON, "folders.json"},
		{"summary csv", model.ViewLanguage, FormatCSV, "summary.csv"},
		{"files csv", model.ViewFile, FormatCSV, "files.csv"},
		{"folders csv", model.ViewFolder, FormatCSV, "folders.csv"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var actual bytes.Buffer
			if err := Write(&actual, report, test.view, test.format); err != nil {
				t.Fatal(err)
			}
			goldenPath := filepath.Join("..", "..", "testdata", "golden", test.file)
			expected, err := os.ReadFile(goldenPath)
			if err != nil {
				t.Fatalf("read golden %s: %v\nactual output:\n%s", goldenPath, err, actual.String())
			}
			if !bytes.Equal(actual.Bytes(), expected) {
				t.Fatalf("output differs from %s\n--- got ---\n%s\n--- want ---\n%s", goldenPath, actual.String(), expected)
			}
		})
	}
}

func TestJSONSelectedArraysAndMachineFields(t *testing.T) {
	report := goldenReport()
	for _, test := range []struct {
		view        model.View
		wantFiles   bool
		wantFolders bool
	}{
		{model.ViewLanguage, false, false},
		{model.ViewFile, true, false},
		{model.ViewFolder, false, true},
	} {
		var buffer bytes.Buffer
		if err := WriteJSON(&buffer, report, test.view); err != nil {
			t.Fatal(err)
		}
		var document map[string]json.RawMessage
		if err := json.Unmarshal(buffer.Bytes(), &document); err != nil {
			t.Fatal(err)
		}
		_, hasFiles := document["files"]
		_, hasFolders := document["folders"]
		if hasFiles != test.wantFiles || hasFolders != test.wantFolders {
			t.Fatalf("view %s fields: files=%v folders=%v", test.view, hasFiles, hasFolders)
		}
		if _, ok := document["languages"]; !ok {
			t.Fatalf("view %s omitted languages", test.view)
		}
		if _, ok := document["totals"]; !ok {
			t.Fatalf("view %s omitted totals", test.view)
		}
		if _, ok := document["metadata"]; !ok {
			t.Fatalf("view %s omitted metadata", test.view)
		}
	}

	empty := report
	empty.Files = nil
	var buffer bytes.Buffer
	if err := WriteJSON(&buffer, empty, model.ViewFile); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buffer.String(), `"files": []`) {
		t.Fatalf("selected empty array was omitted or null:\n%s", buffer.String())
	}
}

func TestJSONCalibrationMetadataIsExplicitAndDeterministic(t *testing.T) {
	exact := goldenReport()
	var exactBuffer bytes.Buffer
	if err := WriteJSON(&exactBuffer, exact, model.ViewLanguage); err != nil {
		t.Fatal(err)
	}
	var exactDocument struct {
		Metadata map[string]json.RawMessage `json:"metadata"`
	}
	if err := json.Unmarshal(exactBuffer.Bytes(), &exactDocument); err != nil {
		t.Fatal(err)
	}
	if _, ok := exactDocument.Metadata["estimated"]; ok {
		t.Fatalf("exact tokenizer metadata includes estimated:\n%s", exactBuffer.String())
	}
	if _, ok := exactDocument.Metadata["calibration_overrides"]; ok {
		t.Fatalf("exact tokenizer metadata includes empty overrides:\n%s", exactBuffer.String())
	}
	if got := string(exactDocument.Metadata["calibration_factor"]); got != "1" {
		t.Fatalf("calibration_factor = %s, want global fallback 1", got)
	}

	estimated := goldenReport()
	estimated.Metadata = model.Metadata{
		Version:           "test-version",
		Tokenizer:         "claude",
		CalibrationFactor: 1.25,
		Estimated:         true,
		CalibrationOverrides: []model.CalibrationOverride{
			{Language: "Python", Factor: 1.4},
			{Language: "Go", Factor: 1.1},
			{Language: "Go", Factor: 1.05},
		},
	}
	var estimatedBuffer bytes.Buffer
	if err := WriteJSON(&estimatedBuffer, estimated, model.ViewLanguage); err != nil {
		t.Fatal(err)
	}
	var estimatedDocument struct {
		Metadata struct {
			CalibrationFactor    float64 `json:"calibration_factor"`
			Estimated            bool    `json:"estimated"`
			CalibrationOverrides []struct {
				Language string  `json:"language"`
				Factor   float64 `json:"factor"`
			} `json:"calibration_overrides"`
		} `json:"metadata"`
	}
	if err := json.Unmarshal(estimatedBuffer.Bytes(), &estimatedDocument); err != nil {
		t.Fatal(err)
	}
	metadata := estimatedDocument.Metadata
	if !metadata.Estimated || metadata.CalibrationFactor != 1.25 {
		t.Fatalf("metadata = %+v", metadata)
	}
	wantOrder := []struct {
		language string
		factor   float64
	}{{"Go", 1.05}, {"Go", 1.1}, {"Python", 1.4}}
	if len(metadata.CalibrationOverrides) != len(wantOrder) {
		t.Fatalf("calibration overrides = %+v", metadata.CalibrationOverrides)
	}
	for index, want := range wantOrder {
		got := metadata.CalibrationOverrides[index]
		if got.Language != want.language || got.Factor != want.factor {
			t.Fatalf("calibration override %d = %+v, want %s=%g", index, got, want.language, want.factor)
		}
	}
	if got := estimated.Metadata.CalibrationOverrides[0].Language; got != "Python" {
		t.Fatalf("renderer mutated report override order: first language = %q", got)
	}
}

func TestFolderMachineOutputPreservesSubtreeIdentity(t *testing.T) {
	report := model.Report{Folders: []model.FolderRow{
		{InputID: 0, Path: "src/pkg", Name: "pkg", Depth: 1},
		{InputID: 1, Path: "src/pkg", Name: "src/pkg", Depth: 0},
		{InputID: 1, Path: "src/pkg/(root files)", Name: "(root files)", Depth: 1, Synthetic: true},
	}}

	var jsonBuffer bytes.Buffer
	if err := WriteJSON(&jsonBuffer, report, model.ViewFolder); err != nil {
		t.Fatal(err)
	}
	var document struct {
		Folders []struct {
			Folder    string `json:"folder"`
			InputID   int    `json:"input_id"`
			Depth     int    `json:"depth"`
			Synthetic bool   `json:"synthetic"`
		} `json:"folders"`
	}
	if err := json.Unmarshal(jsonBuffer.Bytes(), &document); err != nil {
		t.Fatal(err)
	}
	if len(document.Folders) != 3 || document.Folders[0].InputID == document.Folders[1].InputID || !document.Folders[2].Synthetic {
		t.Fatalf("JSON folder identity = %#v", document.Folders)
	}

	var csvBuffer bytes.Buffer
	if err := WriteCSV(&csvBuffer, report, model.ViewFolder); err != nil {
		t.Fatal(err)
	}
	rows, err := csv.NewReader(&csvBuffer).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 4 || rows[1][1] != "0" || rows[2][1] != "1" || rows[3][3] != "true" {
		t.Fatalf("CSV folder identity = %#v", rows)
	}
}

func TestCSVQuotesAndHasNoTotalsRow(t *testing.T) {
	var buffer bytes.Buffer
	if err := WriteCSV(&buffer, goldenReport(), model.ViewFile); err != nil {
		t.Fatal(err)
	}
	rows, err := csv.NewReader(strings.NewReader(buffer.String())).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 4 {
		t.Fatalf("got %d rows, want header + 3 files: %v", len(rows), rows)
	}
	if got := rows[3][1]; got != "README,notes.md" {
		t.Fatalf("quoted path decoded as %q", got)
	}
	for _, row := range rows[1:] {
		if row[0] == "Total" || row[1] == "Total" {
			t.Fatalf("unexpected totals row: %v", row)
		}
	}
}

func TestTabularEstimatedFooterAndZeroCodeDensity(t *testing.T) {
	report := goldenReport()
	report.Metadata = model.Metadata{Tokenizer: "claude", CalibrationFactor: 1.3, Estimated: true}
	var buffer bytes.Buffer
	if err := WriteTabular(&buffer, report, model.ViewLanguage); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buffer.String(), "Markdown") || !strings.Contains(buffer.String(), "0.0") {
		t.Fatalf("zero-code record not rendered correctly:\n%s", buffer.String())
	}
	if !strings.HasSuffix(buffer.String(), "Tokenizer: claude (estimated, calibration factor 1.3)\n") {
		t.Fatalf("estimate footer missing:\n%s", buffer.String())
	}
}

func TestTabularEstimatedFooterSortsCalibrationOverrides(t *testing.T) {
	report := goldenReport()
	report.Metadata = model.Metadata{
		Tokenizer:         "claude",
		CalibrationFactor: 1.25,
		Estimated:         true,
		CalibrationOverrides: []model.CalibrationOverride{
			{Language: "Python", Factor: 1.4},
			{Language: "Go", Factor: 1.1},
		},
	}
	var buffer bytes.Buffer
	if err := WriteTabular(&buffer, report, model.ViewLanguage); err != nil {
		t.Fatal(err)
	}
	want := "Tokenizer: claude (estimated, default calibration factor 1.25; language overrides: Go=1.1, Python=1.4)\n"
	if !strings.HasSuffix(buffer.String(), want) {
		t.Fatalf("estimate footer not deterministic:\n%s\nwant suffix:\n%s", buffer.String(), want)
	}
}

func TestCSVDoesNotExposeCalibrationMetadata(t *testing.T) {
	baseline := goldenReport()
	withOverrides := goldenReport()
	withOverrides.Metadata = model.Metadata{
		Tokenizer:         "claude",
		CalibrationFactor: 1.25,
		Estimated:         true,
		CalibrationOverrides: []model.CalibrationOverride{
			{Language: "Go", Factor: 1.1},
		},
	}

	var baselineBuffer, overrideBuffer bytes.Buffer
	if err := WriteCSV(&baselineBuffer, baseline, model.ViewLanguage); err != nil {
		t.Fatal(err)
	}
	if err := WriteCSV(&overrideBuffer, withOverrides, model.ViewLanguage); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(overrideBuffer.Bytes(), baselineBuffer.Bytes()) {
		t.Fatalf("calibration metadata changed CSV output:\n%s", overrideBuffer.String())
	}
}

func TestTruncateIsRuneSafe(t *testing.T) {
	if got, want := truncate("αβγδε", 4), "αβγ…"; got != want {
		t.Fatalf("truncate() = %q, want %q", got, want)
	}
	if got := truncate("unchanged", 20); got != "unchanged" {
		t.Fatalf("unexpected truncation: %q", got)
	}
	if got := truncate(strings.Repeat("界", 10), 8); runewidth.StringWidth(got) > 8 || !strings.HasSuffix(got, "…") {
		t.Fatalf("wide-rune truncation = %q (width %d)", got, runewidth.StringWidth(got))
	}
}

func TestWriteValidationAndWriterErrors(t *testing.T) {
	report := goldenReport()
	if err := Write(io.Discard, report, model.View(99), FormatJSON); err == nil {
		t.Fatal("invalid view unexpectedly succeeded")
	}
	if err := Write(io.Discard, report, model.ViewLanguage, Format("yaml")); err == nil {
		t.Fatal("invalid format unexpectedly succeeded")
	}
	for _, format := range []Format{FormatTabular, FormatJSON, FormatCSV} {
		err := Write(failingWriter{}, report, model.ViewLanguage, format)
		if err == nil {
			t.Fatalf("%s did not propagate writer error", format)
		}
	}
}

func TestCSVHeaders(t *testing.T) {
	tests := []struct {
		view model.View
		want []string
	}{
		{model.ViewLanguage, []string{"language", "files", "lines", "code", "comments", "blanks", "complexity", "bytes", "tokens"}},
		{model.ViewFile, []string{"language", "path", "files", "lines", "code", "comments", "blanks", "complexity", "bytes", "tokens"}},
		{model.ViewFolder, []string{"folder", "input_id", "depth", "synthetic", "files", "lines", "code", "comments", "blanks", "complexity", "bytes", "tokens"}},
	}
	for _, test := range tests {
		var buffer bytes.Buffer
		if err := WriteCSV(&buffer, goldenReport(), test.view); err != nil {
			t.Fatal(err)
		}
		got, err := csv.NewReader(&buffer).Read()
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, test.want) {
			t.Fatalf("view %s header = %v, want %v", test.view, got, test.want)
		}
	}
}

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) {
	return 0, errors.New("write failed")
}

func goldenReport() model.Report {
	goMetrics := model.Metrics{Files: 2, Lines: 15, Code: 12, Comments: 1, Blanks: 2, Complexity: 3, Bytes: 150, Tokens: 28}
	markdownMetrics := model.Metrics{Files: 1, Lines: 3, Comments: 2, Blanks: 1, Bytes: 30, Tokens: 5}
	totals := model.Metrics{Files: 3, Lines: 18, Code: 12, Comments: 3, Blanks: 3, Complexity: 3, Bytes: 180, Tokens: 33}
	mainMetrics := model.Metrics{Files: 1, Lines: 10, Code: 8, Comments: 1, Blanks: 1, Complexity: 2, Bytes: 100, Tokens: 20}
	packageMetrics := model.Metrics{Files: 1, Lines: 5, Code: 4, Blanks: 1, Complexity: 1, Bytes: 50, Tokens: 8}
	rootFilesMetrics := model.Metrics{Files: 2, Lines: 13, Code: 8, Comments: 3, Blanks: 2, Complexity: 2, Bytes: 130, Tokens: 25}
	return model.Report{
		Languages: []model.LanguageRow{
			{Language: "Go", Metrics: goMetrics},
			{Language: "Markdown", Metrics: markdownMetrics},
		},
		Files: []model.FileRecord{
			{InputID: 0, Path: "main.go", RelPath: "main.go", Language: "Go", Metrics: mainMetrics},
			{InputID: 0, Path: "pkg/b.go", RelPath: "pkg/b.go", Language: "Go", Metrics: packageMetrics},
			{InputID: 0, Path: "README,notes.md", RelPath: "README,notes.md", Language: "Markdown", Metrics: markdownMetrics},
		},
		Folders: []model.FolderRow{
			{InputID: 0, Path: ".", Name: ".", Depth: 0, Metrics: totals},
			{InputID: 0, Path: "(root files)", Name: "(root files)", Depth: 1, Synthetic: true, Metrics: rootFilesMetrics},
			{InputID: 0, Path: "pkg", Name: "pkg", Depth: 1, Metrics: packageMetrics},
		},
		Totals: totals,
		Metadata: model.Metadata{
			Version:           "test-version",
			Tokenizer:         "o200k",
			CalibrationFactor: 1,
		},
	}
}
