package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectSampleLanguage(t *testing.T) {
	tests := []struct {
		path string
		want string
		ok   bool
	}{
		{path: "main.go", want: "Go", ok: true},
		{path: "component.TSX", want: "TypeScript", ok: true},
		{path: "types.d.ts", want: "TypeScript", ok: true},
		{path: "config.json", want: "JSON", ok: true},
		{path: "Cargo.toml", want: "TOML", ok: true},
		{path: "Dockerfile", want: "Dockerfile", ok: true},
		{path: "Makefile", want: "Makefile", ok: true},
		{path: "image.png", ok: false},
	}
	for _, test := range tests {
		got, ok := detectSampleLanguage(test.path)
		if got != test.want || ok != test.ok {
			t.Errorf("detectSampleLanguage(%q) = (%q, %t), want (%q, %t)", test.path, got, ok, test.want, test.ok)
		}
	}
}

func TestCollectSamplesBalancesLanguagesAndSkipsNoise(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "a.go"), "package a\nfunc A() {}\n")
	writeTestFile(t, filepath.Join(root, "b.go"), "package b\nfunc B() {}\n")
	writeTestFile(t, filepath.Join(root, "a.py"), "def a():\n    return 1\n")
	writeTestFile(t, filepath.Join(root, "b.py"), "def b():\n    return 2\n")
	writeTestFile(t, filepath.Join(root, "package-lock.json"), "{}")
	writeTestFile(t, filepath.Join(root, ".cache", "ignored.ts"), "export const cached = true\n")
	writeTestFile(t, filepath.Join(root, "node_modules", "ignored.js"), "export default 1\n")
	writeTestFile(t, filepath.Join(root, "results", "calibration.json"), "{}")
	if err := os.WriteFile(filepath.Join(root, "binary.go"), []byte{'p', 0, 'q'}, 0o644); err != nil {
		t.Fatal(err)
	}

	samples, err := collectSamples([]string{root}, 1024, 2, 2)
	if err != nil {
		t.Fatalf("collectSamples: %v", err)
	}
	if len(samples) != 2 {
		t.Fatalf("sample count = %d, want 2", len(samples))
	}
	gotLanguages := map[string]int{}
	for _, sample := range samples {
		gotLanguages[sample.Language]++
		if sample.ContentSHA == "" {
			t.Errorf("sample %q missing content hash", sample.Path)
		}
	}
	if gotLanguages["Go"] != 1 || gotLanguages["Python"] != 1 {
		t.Errorf("language counts = %v, want one Go and one Python", gotLanguages)
	}
}

func TestLoadSampleTruncatesOnUTF8Boundary(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "unicode.py")
	writeTestFile(t, path, "# éééé\nprint('ok')\n")

	sample, ok, err := loadSample(path, "unicode.py", 6)
	if err != nil {
		t.Fatalf("loadSample: %v", err)
	}
	if !ok {
		t.Fatal("loadSample skipped valid UTF-8")
	}
	if !sample.Truncated {
		t.Error("Truncated = false, want true")
	}
	if string(sample.Content) != "# éé" {
		t.Errorf("content = %q, want valid UTF-8 prefix", sample.Content)
	}
}

func writeTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
