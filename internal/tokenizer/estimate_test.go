package tokenizer

import (
	"errors"
	"math"
	"reflect"
	"sync"
	"testing"
)

type fixedCounter struct {
	count int64
	err   error
}

func (c fixedCounter) Count([]byte) (int64, error) {
	return c.count, c.err
}

func TestEstimatorRounding(t *testing.T) {
	tests := []struct {
		name   string
		base   int64
		factor float64
		want   int64
	}{
		{name: "zero", base: 0, factor: 1.3, want: 0},
		{name: "below half", base: 1, factor: 1.49, want: 1},
		{name: "half rounds up", base: 1, factor: 1.5, want: 2},
		{name: "larger", base: 11, factor: 1.25, want: 14},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			counter, err := newEstimator(fixedCounter{count: test.base}, test.factor, nil)
			if err != nil {
				t.Fatalf("newEstimator: %v", err)
			}
			got, err := counter.Count(nil)
			if err != nil {
				t.Fatalf("Count: %v", err)
			}
			if got != test.want {
				t.Errorf("Count = %d, want %d", got, test.want)
			}
		})
	}
}

func TestEstimatorErrors(t *testing.T) {
	for _, factor := range []float64{0, -1, math.NaN(), math.Inf(1)} {
		if _, err := newEstimator(fixedCounter{}, factor, nil); err == nil {
			t.Errorf("newEstimator factor %v succeeded, want error", factor)
		}
	}
	if _, err := newEstimator(nil, 1, nil); err == nil {
		t.Error("newEstimator nil base succeeded, want error")
	}

	wantErr := errors.New("tokenizer failed")
	counter, err := newEstimator(fixedCounter{err: wantErr}, 1.2, nil)
	if err != nil {
		t.Fatalf("newEstimator: %v", err)
	}
	if _, err := counter.Count(nil); !errors.Is(err, wantErr) {
		t.Errorf("Count error = %v, want %v", err, wantErr)
	}

	counter, err = newEstimator(fixedCounter{count: -1}, 1.2, nil)
	if err != nil {
		t.Fatalf("newEstimator: %v", err)
	}
	if _, err := counter.Count(nil); err == nil {
		t.Error("negative base count succeeded, want error")
	}

	counter, err = newEstimator(fixedCounter{count: math.MaxInt64}, 2, nil)
	if err != nil {
		t.Fatalf("newEstimator: %v", err)
	}
	if _, err := counter.Count(nil); err == nil {
		t.Error("overflowing estimate succeeded, want error")
	}
}

func TestEstimatorLanguageOverridesAndFallback(t *testing.T) {
	overrides := []CalibrationOverride{
		{Language: "Go", Factor: 2},
		{Language: "C#", Factor: 1.1},
	}
	counter, err := newEstimator(fixedCounter{count: 10}, 1.5, overrides)
	if err != nil {
		t.Fatal(err)
	}
	var _ LanguageCounter = counter

	tests := []struct {
		name     string
		language string
		want     int64
	}{
		{name: "first sorted override", language: "C#", want: 11},
		{name: "second sorted override", language: "Go", want: 20},
		{name: "case differs", language: "go", want: 15},
		{name: "unknown", language: "Rust", want: 15},
		{name: "empty", language: "", want: 15},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, countErr := counter.CountForLanguage(nil, test.language)
			if countErr != nil {
				t.Fatal(countErr)
			}
			if got != test.want {
				t.Errorf("CountForLanguage(%q) = %d, want %d", test.language, got, test.want)
			}
		})
	}

	got, err := counter.Count(nil)
	if err != nil {
		t.Fatal(err)
	}
	if got != 15 {
		t.Errorf("Count = %d, want global fallback 15", got)
	}
}

type languageFixedCounter struct {
	fixedCounter
	languageCount int64
}

func (c languageFixedCounter) CountForLanguage([]byte, string) (int64, error) {
	return c.languageCount, c.err
}

func TestCountForLanguageHelper(t *testing.T) {
	got, err := CountForLanguage(fixedCounter{count: 7}, []byte("content"), "Go")
	if err != nil || got != 7 {
		t.Fatalf("plain counter result = (%d, %v), want (7, nil)", got, err)
	}

	got, err = CountForLanguage(languageFixedCounter{
		fixedCounter:  fixedCounter{count: 7},
		languageCount: 11,
	}, []byte("content"), "Go")
	if err != nil || got != 11 {
		t.Fatalf("language counter result = (%d, %v), want (11, nil)", got, err)
	}

	wantErr := errors.New("language count failed")
	_, err = CountForLanguage(languageFixedCounter{fixedCounter: fixedCounter{err: wantErr}}, nil, "Go")
	if !errors.Is(err, wantErr) {
		t.Fatalf("language counter error = %v, want %v", err, wantErr)
	}
}

func TestEstimatorValidatesOverrides(t *testing.T) {
	tests := []struct {
		name      string
		overrides []CalibrationOverride
	}{
		{name: "empty language", overrides: []CalibrationOverride{{Language: "", Factor: 1}}},
		{name: "whitespace language", overrides: []CalibrationOverride{{Language: "  ", Factor: 1}}},
		{name: "surrounding whitespace", overrides: []CalibrationOverride{{Language: " Go", Factor: 1}}},
		{name: "zero factor", overrides: []CalibrationOverride{{Language: "Go", Factor: 0}}},
		{name: "negative factor", overrides: []CalibrationOverride{{Language: "Go", Factor: -1}}},
		{name: "nan factor", overrides: []CalibrationOverride{{Language: "Go", Factor: math.NaN()}}},
		{name: "positive infinity factor", overrides: []CalibrationOverride{{Language: "Go", Factor: math.Inf(1)}}},
		{name: "negative infinity factor", overrides: []CalibrationOverride{{Language: "Go", Factor: math.Inf(-1)}}},
		{name: "duplicate language", overrides: []CalibrationOverride{
			{Language: "Go", Factor: 1.1},
			{Language: "C#", Factor: 1.2},
			{Language: "Go", Factor: 1.3},
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := newEstimator(fixedCounter{}, 1, test.overrides); err == nil {
				t.Fatal("newEstimator succeeded, want error")
			}
		})
	}
}

func TestEstimatorOwnsSortedOverrides(t *testing.T) {
	input := []CalibrationOverride{
		{Language: "YAML", Factor: 1.4},
		{Language: "C#", Factor: 1.9},
	}
	wantInput := append([]CalibrationOverride(nil), input...)
	counter, err := newEstimator(fixedCounter{count: 10}, 1.5, input)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(input, wantInput) {
		t.Fatalf("constructor mutated input: got %#v, want %#v", input, wantInput)
	}

	input[0] = CalibrationOverride{Language: "C#", Factor: 100}
	got, err := counter.CountForLanguage(nil, "YAML")
	if err != nil || got != 14 {
		t.Fatalf("count after input mutation = (%d, %v), want (14, nil)", got, err)
	}

	firstCopy := counter.calibrationOverridesCopy()
	wantSorted := []CalibrationOverride{
		{Language: "C#", Factor: 1.9},
		{Language: "YAML", Factor: 1.4},
	}
	if !reflect.DeepEqual(firstCopy, wantSorted) {
		t.Fatalf("stored overrides = %#v, want %#v", firstCopy, wantSorted)
	}
	firstCopy[0].Factor = 100
	if got := counter.calibrationOverridesCopy(); !reflect.DeepEqual(got, wantSorted) {
		t.Fatalf("copy mutation changed estimator: got %#v, want %#v", got, wantSorted)
	}
}

func TestEstimatorLanguageCountRetainsErrorsAndOverflowChecks(t *testing.T) {
	wantErr := errors.New("tokenizer failed")
	counter, err := newEstimator(fixedCounter{err: wantErr}, 1, []CalibrationOverride{{Language: "Go", Factor: 2}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := counter.CountForLanguage(nil, "Go"); !errors.Is(err, wantErr) {
		t.Fatalf("CountForLanguage error = %v, want %v", err, wantErr)
	}

	counter, err = newEstimator(fixedCounter{count: -1}, 1, []CalibrationOverride{{Language: "Go", Factor: 2}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := counter.CountForLanguage(nil, "Go"); err == nil {
		t.Fatal("negative base count succeeded, want error")
	}

	counter, err = newEstimator(fixedCounter{count: math.MaxInt64}, 1, []CalibrationOverride{{Language: "Go", Factor: 2}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := counter.CountForLanguage(nil, "Go"); err == nil {
		t.Fatal("overflowing language estimate succeeded, want error")
	}
}

func TestEstimatorConcurrentLanguageCounts(t *testing.T) {
	counter, err := newEstimator(fixedCounter{count: 10}, 1.5, []CalibrationOverride{{Language: "Go", Factor: 2}})
	if err != nil {
		t.Fatal(err)
	}

	var workers sync.WaitGroup
	errCh := make(chan error, 100)
	for index := range 100 {
		workers.Add(1)
		go func() {
			defer workers.Done()
			language := "Rust"
			want := int64(15)
			if index%2 == 0 {
				language = "Go"
				want = 20
			}
			got, countErr := counter.CountForLanguage(nil, language)
			if countErr != nil {
				errCh <- countErr
			} else if got != want {
				errCh <- errors.New("unexpected concurrent count")
			}
		}()
	}
	workers.Wait()
	close(errCh)
	for err := range errCh {
		t.Error(err)
	}
}
