package tokenizer

import (
	"crypto/sha256"
	"fmt"
	"sync"
	"testing"
)

func TestEmbeddedO200KAsset(t *testing.T) {
	const wantSHA256 = "446a9538cb6c348e3516120d7c08b09f57c36495e2acfffe59a5bf8b0cfb1a2d"
	if got := fmt.Sprintf("%x", sha256.Sum256(o200kRanks)); got != wantSHA256 {
		t.Fatalf("embedded o200k_base SHA-256 = %s, want %s", got, wantSHA256)
	}

	ranks, err := (embeddedBPELoader{}).LoadTiktokenBpe("https://openaipublic.blob.core.windows.net/encodings/o200k_base.tiktoken")
	if err != nil {
		t.Fatalf("LoadTiktokenBpe: %v", err)
	}
	if got := len(ranks); got != o200kExpectedRanks {
		t.Fatalf("rank count = %d, want %d", got, o200kExpectedRanks)
	}
	if got := ranks["!"]; got != 0 {
		t.Errorf("rank for ! = %d, want 0", got)
	}
	if got := ranks["\""]; got != 1 {
		t.Errorf("rank for quote = %d, want 1", got)
	}
}

func TestEmbeddedBPELoaderRejectsOtherEncodings(t *testing.T) {
	if _, err := (embeddedBPELoader{}).LoadTiktokenBpe("cl100k_base.tiktoken"); err == nil {
		t.Fatal("LoadTiktokenBpe(cl100k_base) succeeded, want error")
	}
}

func TestO200KKnownVectors(t *testing.T) {
	counter, _, err := New(NameO200K)
	if err != nil {
		t.Fatalf("New(o200k): %v", err)
	}

	tests := []struct {
		name    string
		content string
		want    int64
	}{
		{name: "empty", content: "", want: 0},
		{name: "single word", content: "hello", want: 1},
		{name: "latin sentence", content: "hallo world!", want: 4},
		{name: "Chinese", content: "你好世界！", want: 3},
		{name: "Japanese", content: "こんにちは世界！", want: 3},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := counter.Count([]byte(test.content))
			if err != nil {
				t.Fatalf("Count: %v", err)
			}
			if got != test.want {
				t.Errorf("Count(%q) = %d, want %d", test.content, got, test.want)
			}
		})
	}
}

func TestO200KCountsSpecialTokenTextOrdinarily(t *testing.T) {
	counter, _, err := New(NameO200K)
	if err != nil {
		t.Fatalf("New(o200k): %v", err)
	}
	// This string is a special token only when explicitly allowed. Source files
	// must count its literal characters and must never trigger a panic.
	got, err := counter.Count([]byte("<|endoftext|>"))
	if err != nil {
		t.Fatalf("Count: %v", err)
	}
	if got != 7 {
		t.Errorf("ordinary special-token text count = %d, want 7", got)
	}
}

func TestO200KConcurrentUse(t *testing.T) {
	counter, _, err := New(NameO200K)
	if err != nil {
		t.Fatalf("New(o200k): %v", err)
	}
	content := []byte("package main\n\nfunc main() { println(\"hello\") }\n")
	want, err := counter.Count(content)
	if err != nil {
		t.Fatalf("Count: %v", err)
	}

	var wg sync.WaitGroup
	errors := make(chan error, 32)
	for range 32 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 20 {
				got, err := counter.Count(content)
				if err != nil {
					errors <- err
					return
				}
				if got != want {
					errors <- fmt.Errorf("count = %d, want %d", got, want)
					return
				}
			}
		}()
	}
	wg.Wait()
	close(errors)
	for err := range errors {
		t.Error(err)
	}
}
