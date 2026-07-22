package tokenizer

import (
	"reflect"
	"strings"
	"testing"
)

func TestFactoryMetadataAndAliases(t *testing.T) {
	tests := []struct {
		input     string
		wantName  string
		estimated bool
		factor    float64
	}{
		{input: NameO200K, wantName: NameO200K, factor: 1},
		{input: " CODEX ", wantName: NameO200K, factor: 1},
	}
	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			counter, metadata, err := New(test.input)
			if err != nil {
				t.Fatalf("New: %v", err)
			}
			if counter == nil {
				t.Fatal("New returned nil counter")
			}
			if metadata.Name != test.wantName || metadata.Estimated != test.estimated || metadata.CalibrationFactor != test.factor {
				t.Errorf("metadata = %+v", metadata)
			}
			if metadata.Encoding != "o200k_base" {
				t.Errorf("encoding = %q, want o200k_base", metadata.Encoding)
			}
			if len(metadata.CalibrationOverrides) != 0 {
				t.Errorf("calibration overrides = %#v, want none", metadata.CalibrationOverrides)
			}
		})
	}
}

func TestFactoryClaudeAvailabilityMatchesCalibrationState(t *testing.T) {
	tests := []struct {
		name   string
		ready  bool
		factor float64
	}{
		{NameClaude, ClaudeCurrentCalibrationReady, ClaudeCurrentCalibrationFactor},
		{NameClaudeLegacy, ClaudeLegacyCalibrationReady, ClaudeLegacyCalibrationFactor},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			counter, metadata, err := New(test.name)
			if !test.ready {
				if err == nil || !strings.Contains(err.Error(), "calibration has not been completed") {
					t.Fatalf("New(%q) error = %v", test.name, err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if counter == nil || metadata.Name != test.name || !metadata.Estimated || metadata.CalibrationFactor != test.factor {
				t.Fatalf("counter=%v metadata=%+v", counter, metadata)
			}
		})
	}
}

func TestFactoryClaudeCalibrationOverrides(t *testing.T) {
	tests := []struct {
		name string
		want []CalibrationOverride
	}{
		{
			name: NameClaude,
			want: []CalibrationOverride{
				{Language: "C#", Factor: 1.907865},
				{Language: "JSON", Factor: 1.481303},
				{Language: "SQL", Factor: 1.997848},
				{Language: "YAML", Factor: 1.459644},
			},
		},
		{
			name: NameClaudeLegacy,
			want: []CalibrationOverride{
				{Language: "C#", Factor: 1.404494},
				{Language: "Markdown", Factor: 1.119048},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			firstCounter, firstMetadata, err := New(test.name)
			if err != nil {
				t.Fatal(err)
			}
			secondCounter, secondMetadata, err := New(test.name)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(firstMetadata.CalibrationOverrides, test.want) {
				t.Fatalf("first overrides = %#v, want %#v", firstMetadata.CalibrationOverrides, test.want)
			}
			if !reflect.DeepEqual(secondMetadata.CalibrationOverrides, test.want) {
				t.Fatalf("second overrides = %#v, want %#v", secondMetadata.CalibrationOverrides, test.want)
			}

			firstMetadata.CalibrationOverrides[0].Factor = 999
			if !reflect.DeepEqual(secondMetadata.CalibrationOverrides, test.want) {
				t.Fatalf("metadata instances share override storage: %#v", secondMetadata.CalibrationOverrides)
			}
			firstEstimator, ok := firstCounter.(*estimator)
			if !ok {
				t.Fatalf("first counter type = %T, want *estimator", firstCounter)
			}
			secondEstimator, ok := secondCounter.(*estimator)
			if !ok {
				t.Fatalf("second counter type = %T, want *estimator", secondCounter)
			}
			if got := firstEstimator.calibrationOverridesCopy(); !reflect.DeepEqual(got, test.want) {
				t.Fatalf("metadata mutation changed first counter: %#v", got)
			}
			if got := secondEstimator.calibrationOverridesCopy(); !reflect.DeepEqual(got, test.want) {
				t.Fatalf("first factory result changed second counter: %#v", got)
			}
		})
	}
}

func TestFactoryRejectsUnknown(t *testing.T) {
	if _, _, err := New("cl100k"); err == nil {
		t.Fatal("New(cl100k) succeeded, want error")
	} else if !strings.Contains(err.Error(), "supported") {
		t.Errorf("error %q does not list supported values", err)
	}
}

func TestSupportedReturnsIndependentSlice(t *testing.T) {
	first := Supported()
	first[0] = "changed"
	if Supported()[0] == "changed" {
		t.Fatal("Supported returned shared mutable storage")
	}
}
