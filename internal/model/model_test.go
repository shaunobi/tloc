package model

import "testing"

func TestMetricsAddAndDensity(t *testing.T) {
	metrics := Metrics{Files: 1, Lines: 10, Code: 4, Comments: 3, Blanks: 3, Complexity: 2, Bytes: 100, Tokens: 10}
	metrics.Add(Metrics{Files: 2, Lines: 5, Code: 2, Comments: 1, Blanks: 2, Complexity: 3, Bytes: 50, Tokens: 5})
	want := Metrics{Files: 3, Lines: 15, Code: 6, Comments: 4, Blanks: 5, Complexity: 5, Bytes: 150, Tokens: 15}
	if metrics != want {
		t.Fatalf("Add() = %+v, want %+v", metrics, want)
	}
	if got := metrics.TokensPerCodeLine(); got != 2.5 {
		t.Fatalf("TokensPerCodeLine() = %v, want 2.5", got)
	}
	if got := (Metrics{Tokens: 9}).TokensPerCodeLine(); got != 0 {
		t.Fatalf("zero-code density = %v, want 0", got)
	}
}

func TestViewAndSortKeyValidation(t *testing.T) {
	for _, view := range []View{ViewLanguage, ViewFile, ViewFolder} {
		if !view.Valid() {
			t.Errorf("%v should be valid", view)
		}
	}
	if View(99).Valid() {
		t.Fatal("unexpected valid unknown view")
	}
	for _, key := range []SortKey{SortTokens, SortCode, SortLines, SortFiles, SortName} {
		if !key.Valid() {
			t.Errorf("%q should be valid", key)
		}
	}
	if SortKey("bytes").Valid() {
		t.Fatal("unexpected valid unsupported sort key")
	}
}
