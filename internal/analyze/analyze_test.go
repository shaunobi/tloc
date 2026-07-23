package analyze

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"
)

type byteCounter struct{}

func (byteCounter) Count(content []byte) (int64, error) { return int64(len(content)), nil }

type languageAwareCounter struct {
	languageCalls int
	plainCalls    int
	language      string
	content       string
	tokens        int64
	err           error
}

func (c *languageAwareCounter) Count([]byte) (int64, error) {
	c.plainCalls++
	return 0, errors.New("plain Count fallback called")
}

func (c *languageAwareCounter) CountForLanguage(content []byte, language string) (int64, error) {
	c.languageCalls++
	c.language = language
	c.content = string(content)
	return c.tokens, c.err
}

type recordingPlainCounter struct {
	calls  int
	tokens int64
}

func (c *recordingPlainCounter) Count([]byte) (int64, error) {
	c.calls++
	return c.tokens, nil
}

func TestRunPassesCanonicalLanguageToLanguageAwareCounter(t *testing.T) {
	dir := t.TempDir()
	content := "package main\n"
	writeTestFile(t, dir, "main.go", content)
	counter := &languageAwareCounter{tokens: 37}

	_, records, _, err := Run([]string{dir}, counter, Options{Workers: 1})
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 1 {
		t.Fatalf("records = %#v", records)
	}
	if counter.languageCalls != 1 || counter.plainCalls != 0 {
		t.Fatalf("language calls=%d plain calls=%d", counter.languageCalls, counter.plainCalls)
	}
	if counter.language != "Go" || counter.language != records[0].Language {
		t.Fatalf("counter language=%q record language=%q", counter.language, records[0].Language)
	}
	if counter.content != content || records[0].Metrics.Tokens != 37 {
		t.Fatalf("counter content=%q record=%#v", counter.content, records[0])
	}
}

func TestRunFallsBackToPlainCounter(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "main.go", "package main\n")
	counter := &recordingPlainCounter{tokens: 23}

	_, records, _, err := Run([]string{dir}, counter, Options{Workers: 1})
	if err != nil {
		t.Fatal(err)
	}
	if counter.calls != 1 || len(records) != 1 || records[0].Metrics.Tokens != 23 {
		t.Fatalf("calls=%d records=%#v", counter.calls, records)
	}
}

func TestRunReadsCountedFileOnce(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "main.go")
	content := []byte("package main\n\n// hello\nfunc main() {}\n")
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatal(err)
	}

	var mu sync.Mutex
	reads := map[string]int{}
	reader := func(path string, maxFileBytes int64) ([]byte, error) {
		mu.Lock()
		reads[path]++
		mu.Unlock()
		return readFileLimited(path, maxFileBytes)
	}
	inputs, records, _, err := runWithReader([]string{dir}, byteCounter{}, Options{Workers: 2}, reader)
	if err != nil {
		t.Fatal(err)
	}
	if len(inputs) != 1 || len(records) != 1 {
		t.Fatalf("inputs=%d records=%d", len(inputs), len(records))
	}
	if reads[path] != 1 {
		t.Fatalf("reads = %d", reads[path])
	}
	if records[0].Language != "Go" || records[0].Metrics.Files != 1 || records[0].Metrics.Tokens != int64(len(content)) {
		t.Fatalf("record = %#v", records[0])
	}
	if records[0].Metrics.Code != 2 || records[0].Metrics.Comments != 1 || records[0].Metrics.Blanks != 1 {
		t.Fatalf("metrics = %#v", records[0].Metrics)
	}
}

func TestRunKeepsReadableRecordsAndReturnsDeterministicReadWarnings(t *testing.T) {
	dir := t.TempDir()
	goodPath := filepath.Join(dir, "good.go")
	firstBadPath := filepath.Join(dir, "a_bad.go")
	secondBadPath := filepath.Join(dir, "z_bad.go")
	writeTestFile(t, dir, "good.go", "package good\n")
	writeTestFile(t, dir, "a_bad.go", "package bad\n")
	writeTestFile(t, dir, "z_bad.go", "package bad\n")

	reader := func(path string, maxFileBytes int64) ([]byte, error) {
		if path == firstBadPath || path == secondBadPath {
			return nil, &os.PathError{Op: "open", Path: path, Err: errors.New("locked for test")}
		}
		return readFileLimited(path, maxFileBytes)
	}
	inputs, records, warnings, err := runWithReader([]string{dir}, byteCounter{}, Options{Workers: 3}, reader)
	if err != nil {
		t.Fatalf("recoverable read failures became fatal: %v", err)
	}
	if len(inputs) != 1 || len(records) != 1 || records[0].Path != filepath.ToSlash(goodPath) {
		t.Fatalf("inputs=%#v records=%#v", inputs, records)
	}
	if len(warnings) != 2 {
		t.Fatalf("warnings = %#v", warnings)
	}
	if warnings[0].Stage != "read" || warnings[0].Path != filepath.ToSlash(firstBadPath) || warnings[0].Message != "locked for test" {
		t.Fatalf("first warning = %#v", warnings[0])
	}
	if warnings[1].Stage != "read" || warnings[1].Path != filepath.ToSlash(secondBadPath) || warnings[1].Message != "locked for test" {
		t.Fatalf("second warning = %#v", warnings[1])
	}
}

func TestRunBoundsReadsAndAcceptsAnExactCap(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "main.go")
	content := []byte("package p\n")
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatal(err)
	}

	_, records, _, err := Run([]string{dir}, byteCounter{}, Options{
		MaxFileBytes: int64(len(content)),
		Workers:      1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 1 {
		t.Fatalf("exact-cap records = %#v", records)
	}

	readerCalls := 0
	growingReader := func(string, int64) ([]byte, error) {
		readerCalls++
		return append(content, 'x'), nil
	}
	_, records, _, err = runWithReader([]string{dir}, byteCounter{}, Options{
		MaxFileBytes: int64(len(content)),
		Workers:      1,
	}, growingReader)
	if err != nil {
		t.Fatal(err)
	}
	if readerCalls != 1 || len(records) != 0 {
		t.Fatalf("growth calls=%d records=%#v", readerCalls, records)
	}

	larger := filepath.Join(dir, "larger.go")
	if err := os.WriteFile(larger, []byte(strings.Repeat("x", 100)), 0o600); err != nil {
		t.Fatal(err)
	}
	bounded, err := readFileLimited(larger, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(bounded) != 11 {
		t.Fatalf("bounded read length = %d, want 11", len(bounded))
	}
}

func TestRunSkipsBinaryLargeUnknownAndFilteredFiles(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "keep.go", "package keep\n")
	writeTestFile(t, dir, "skip.ts", "export const n = 1\n")
	writeTestFile(t, dir, "large.go", strings.Repeat("x", 50))
	writeTestFile(t, dir, "binary.go", "package x\x00more")
	writeTestFile(t, dir, "unknown.zzz", "unknown")

	_, records, _, err := Run([]string{dir}, byteCounter{}, Options{
		IncludeExt:   []string{"go", "zzz"},
		MaxFileBytes: 20,
		Workers:      3,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 1 || !strings.HasSuffix(records[0].Path, "keep.go") {
		t.Fatalf("records = %#v", records)
	}
}

func TestRunHonorsIgnoreFilesAndDisableFlags(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "keep.go", "package keep\n")
	writeTestFile(t, dir, "gitignored.go", "package ignored\n")
	writeTestFile(t, dir, "ignored.go", "package ignored\n")
	writeTestFile(t, dir, "sccignored.go", "package ignored\n")
	writeTestFile(t, dir, ".gitignore", "gitignored.go\n")
	writeTestFile(t, dir, ".ignore", "ignored.go\n")
	writeTestFile(t, dir, ".sccignore", "sccignored.go\n")

	_, records, _, err := Run([]string{dir}, byteCounter{}, Options{Workers: 2})
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 1 || !strings.HasSuffix(records[0].Path, "keep.go") {
		t.Fatalf("default records = %#v", records)
	}

	_, records, _, err = Run([]string{dir}, byteCounter{}, Options{Workers: 2, NoIgnore: true, NoGitignore: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 4 {
		t.Fatalf("disabled ignores records = %#v", records)
	}

	_, records, _, err = Run([]string{dir}, byteCounter{}, Options{Workers: 2, NoIgnore: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 3 {
		t.Fatalf("no-ignore records = %#v", records)
	}

	_, records, _, err = Run([]string{dir}, byteCounter{}, Options{Workers: 2, NoGitignore: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 2 {
		t.Fatalf("no-gitignore records = %#v", records)
	}
}

func TestRunHonorsNestedIgnoreNegation(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "nested/drop.go", "package drop\n")
	writeTestFile(t, dir, "nested/keep.go", "package keep\n")
	writeTestFile(t, dir, ".gitignore", "nested/*.go\n!nested/keep.go\n")

	_, records, _, err := Run([]string{dir}, byteCounter{}, Options{Workers: 2})
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 1 || records[0].RelPath != "nested/keep.go" {
		t.Fatalf("negation records = %#v", records)
	}
}

func TestRunExtensionFiltersDoNotOverrideIgnoreRules(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "keep.go", "package keep\n")
	writeTestFile(t, dir, "other.py", "print('other')\n")
	writeTestFile(t, dir, "gitignored.go", "package ignored\n")
	writeTestFile(t, dir, "ignored.go", "package ignored\n")
	writeTestFile(t, dir, "sccignored.go", "package ignored\n")
	writeTestFile(t, dir, ".gitignore", "gitignored.go\n")
	writeTestFile(t, dir, ".ignore", "ignored.go\n")
	writeTestFile(t, dir, ".sccignore", "sccignored.go\n")

	_, records, _, err := Run([]string{dir}, byteCounter{}, Options{
		IncludeExt: []string{"go"},
		ExcludeExt: []string{"ts"},
		Workers:    2,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 1 || !strings.HasSuffix(records[0].Path, "keep.go") {
		t.Fatalf("records = %#v", records)
	}
}

func TestRunExtensionFiltersMatchSCCCompoundAndDotfileExtensions(t *testing.T) {
	dir := t.TempDir()
	blade := filepath.Join(dir, "view.blade.php")
	definition := filepath.Join(dir, "types.d.ts")
	bashrc := filepath.Join(dir, ".bashrc")
	writeTestFile(t, dir, "view.blade.php", "<?php echo 'hello';\n")
	writeTestFile(t, dir, "types.d.ts", "export interface Item { id: string }\n")
	writeTestFile(t, dir, ".bashrc", "#!/usr/bin/env bash\necho hello\n")
	writeTestFile(t, dir, "xmake.lua", "target('demo')\n")

	_, records, _, err := Run([]string{blade}, byteCounter{}, Options{
		IncludeExt: []string{"blade.php"},
		Workers:    1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 1 || !strings.HasSuffix(records[0].Path, "view.blade.php") {
		t.Fatalf("compound direct-input records = %#v", records)
	}

	_, records, _, err = Run([]string{dir}, byteCounter{}, Options{
		ExcludeExt: []string{"d.ts"},
		Workers:    1,
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, record := range records {
		if strings.HasSuffix(record.Path, "types.d.ts") {
			t.Fatalf("excluded compound extension was counted: %#v", records)
		}
	}

	_, records, _, err = Run([]string{bashrc}, byteCounter{}, Options{Workers: 1})
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 1 || !strings.HasSuffix(records[0].Path, ".bashrc") {
		t.Fatalf("dotfile records = %#v", records)
	}

	_, records, _, err = Run([]string{definition}, byteCounter{}, Options{
		IncludeExt: []string{"ts"},
		Workers:    1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 0 {
		t.Fatalf("scc compound extension unexpectedly matched final suffix: %#v", records)
	}

	_, records, _, err = Run([]string{dir}, byteCounter{}, Options{
		IncludeExt: []string{"lua"},
		Workers:    1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 1 || !strings.HasSuffix(records[0].Path, "xmake.lua") || records[0].Language != "Lua" {
		t.Fatalf("special filename allow-list records = %#v", records)
	}
}

func TestRunSkipsSymlinkFilesLikeSCC(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "target.go")
	link := filepath.Join(dir, "link.go")
	writeTestFile(t, dir, "target.go", "package target\n")
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}

	_, records, _, err := Run([]string{link}, byteCounter{}, Options{Workers: 1})
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 0 {
		t.Fatalf("symlink records = %#v", records)
	}
}

func TestRunSkipsExplicitIgnoreControlFiles(t *testing.T) {
	dir := t.TempDir()
	paths := make([]string, 0, 3)
	for _, name := range []string{".gitignore", ".ignore", ".sccignore"} {
		writeTestFile(t, dir, name, "ignored.go\n")
		paths = append(paths, filepath.Join(dir, name))
	}

	_, records, _, err := Run(paths, byteCounter{}, Options{Workers: 1})
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 0 {
		t.Fatalf("ignore control records = %#v", records)
	}
}

func TestRunSupportsShebangAndMultipleInputs(t *testing.T) {
	dir := t.TempDir()
	script := filepath.Join(dir, "hello")
	writeTestFile(t, dir, "hello", "#!/usr/bin/env python3\nprint('hello')\n")
	other := filepath.Join(dir, "other.go")
	writeTestFile(t, dir, "other.go", "package other\n")

	inputs, records, _, err := Run([]string{script, other}, byteCounter{}, Options{Workers: 2})
	if err != nil {
		t.Fatal(err)
	}
	if len(inputs) != 2 || len(records) != 2 {
		t.Fatalf("inputs=%#v records=%#v", inputs, records)
	}
	if records[0].InputID == records[1].InputID {
		t.Fatalf("input IDs were not preserved: %#v", records)
	}
}

func TestRunIsDeterministicAcrossWorkerCompletionOrder(t *testing.T) {
	dir := t.TempDir()
	for index, delay := range []int{5, 1, 4, 2, 3} {
		writeTestFile(t, dir, filepath.ToSlash(filepath.Join("pkg", string(rune('a'+index))+".go")), "package p\n// delay:"+string(rune('0'+delay))+"\n")
	}

	_, sequential, _, err := Run([]string{dir}, delayedCounter{}, Options{Workers: 1})
	if err != nil {
		t.Fatal(err)
	}
	_, concurrent, _, err := Run([]string{dir}, delayedCounter{}, Options{Workers: 5})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(concurrent, sequential) {
		t.Fatalf("completion order changed records\nconcurrent: %#v\nsequential: %#v", concurrent, sequential)
	}
}

type delayedCounter struct{}

func (delayedCounter) Count(content []byte) (int64, error) {
	marker := string(content[len(content)-2])
	delay := int(marker[0] - '0')
	time.Sleep(time.Duration(delay) * time.Millisecond)
	return int64(len(content)), nil
}

func TestRunReturnsTokenizerErrors(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "main.go", "package main\n")
	counter := errorCounter{err: errors.New("boom")}
	_, _, _, err := Run([]string{dir}, counter, Options{Workers: 1})
	expectedPath := filepath.ToSlash(filepath.Join(dir, "main.go"))
	if err == nil || !strings.Contains(err.Error(), "tokenize") || !strings.Contains(err.Error(), expectedPath) || !strings.Contains(err.Error(), "boom") {
		t.Fatalf("err = %v", err)
	}
}

type errorCounter struct{ err error }

func (c errorCounter) Count([]byte) (int64, error) { return 0, c.err }

func writeTestFile(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}
