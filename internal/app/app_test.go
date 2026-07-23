package app

import (
	"bytes"
	"strings"
	"testing"

	"github.com/shaunobi/tloc/internal/analyze"
	"github.com/shaunobi/tloc/internal/tokenizer"
)

func TestMainHelpAndVersionAvoidScanning(t *testing.T) {
	for _, test := range []struct {
		args []string
		want string
	}{
		{args: []string{"--help"}, want: "Tokenizer accuracy:"},
		{args: []string{"--version"}, want: "tloc "},
	} {
		var stdout, stderr bytes.Buffer
		if code := Main(test.args, &stdout, &stderr); code != 0 {
			t.Fatalf("Main(%v) code=%d stderr=%q", test.args, code, stderr.String())
		}
		if got := stdout.String(); !strings.Contains(got, test.want) {
			t.Fatalf("Main(%v) stdout %q missing %q", test.args, got, test.want)
		}
		if stderr.Len() != 0 {
			t.Fatalf("Main(%v) stderr = %q, want empty", test.args, stderr.String())
		}
	}
}

func TestToModelMetadataCopiesCalibrationOverrides(t *testing.T) {
	source := tokenizer.Metadata{
		Name:              tokenizer.NameClaude,
		CalibrationFactor: 1.2,
		Estimated:         true,
		CalibrationOverrides: []tokenizer.CalibrationOverride{
			{Language: "Go", Factor: 1.1},
		},
	}

	got := toModelMetadata(source)
	if got.Tokenizer != source.Name || got.CalibrationFactor != source.CalibrationFactor || !got.Estimated {
		t.Fatalf("metadata = %+v", got)
	}
	if len(got.CalibrationOverrides) != 1 || got.CalibrationOverrides[0].Language != "Go" || got.CalibrationOverrides[0].Factor != 1.1 {
		t.Fatalf("calibration overrides = %+v", got.CalibrationOverrides)
	}

	source.CalibrationOverrides[0].Language = "changed source"
	if got.CalibrationOverrides[0].Language != "Go" {
		t.Fatalf("model metadata aliases tokenizer metadata: %+v", got.CalibrationOverrides)
	}
	got.CalibrationOverrides[0].Factor = 9
	if source.CalibrationOverrides[0].Factor != 1.1 {
		t.Fatalf("tokenizer metadata aliases model metadata: %+v", source.CalibrationOverrides)
	}
}

func TestMainRejectsInvalidFlags(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := Main([]string{"--by-file", "--by-folder"}, &stdout, &stderr); code != 2 {
		t.Fatalf("code=%d stderr=%q", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "mutually exclusive") {
		t.Fatalf("stderr=%q", stderr.String())
	}
}

func TestMainRendersPartialReportWarnsAndReturnsFailure(t *testing.T) {
	analyzer := func([]string, tokenizer.Counter, analyze.Options) ([]analyze.InputRoot, []analyze.FileRecord, []analyze.ScanWarning, error) {
		return []analyze.InputRoot{{
				ID:    0,
				Given: ".",
				Abs:   "test-root",
				Kind:  analyze.InputDirectory,
			}}, []analyze.FileRecord{{
				InputID:  0,
				Path:     "good.go",
				RelPath:  "good.go",
				Language: "Go",
				Metrics: analyze.Metrics{
					Files:  1,
					Lines:  1,
					Code:   1,
					Bytes:  13,
					Tokens: 3,
				},
			}}, []analyze.ScanWarning{{
				Stage:   "read",
				Path:    "locked.go",
				Message: "sharing violation",
			}}, nil
	}

	var stdout, stderr bytes.Buffer
	code := mainWithAnalyzer([]string{"--format", "json", "--by-file", "."}, &stdout, &stderr, analyzer)
	if code != 1 {
		t.Fatalf("code=%d, want 1; stderr=%q stdout=%q", code, stderr.String(), stdout.String())
	}
	for _, want := range []string{`"path": "good.go"`, `"complete": false`, `"stage": "read"`, `"path": "locked.go"`, `"error": "sharing violation"`} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("partial JSON missing %q:\n%s", want, stdout.String())
		}
	}
	if !strings.Contains(stderr.String(), "warning: incomplete scan") ||
		!strings.Contains(stderr.String(), "locked.go") ||
		!strings.Contains(stderr.String(), "sharing violation") {
		t.Fatalf("stderr=%q", stderr.String())
	}
}
