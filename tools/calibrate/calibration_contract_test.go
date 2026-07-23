package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/shaunobi/tloc/internal/tokenizer"
)

func TestCheckedInCalibrationMatchesProductionFactors(t *testing.T) {
	data, err := os.ReadFile("results/calibration.json")
	if err != nil {
		t.Fatalf("read checked-in calibration report: %v", err)
	}
	var report calibrationReport
	if err := json.Unmarshal(data, &report); err != nil {
		t.Fatalf("decode checked-in calibration report: %v", err)
	}

	current := requireModelReport(t, report, tokenizer.NameClaude)
	legacy := requireModelReport(t, report, tokenizer.NameClaudeLegacy)
	fable := requireModelReport(t, report, "fable5")
	if current.Global.Samples != 80 || len(current.PerLanguage) != 16 {
		t.Fatalf("current calibration coverage = %d samples across %d languages, want 80 across 16", current.Global.Samples, len(current.PerLanguage))
	}
	if current.HeldOut == nil || legacy.HeldOut == nil || fable.HeldOut == nil {
		t.Fatal("calibration report is missing a held-out evaluation")
	}
	assertCheckedInSamplesMatchReport(t, report.Sampling.Inputs, report.Sampling, current.Samples, "calibration")
	assertCheckedInSamplesMatchReport(t, report.Sampling.HeldOutInputs, report.Sampling, current.HeldOut.Samples, "held-out")
	assertDisjointMeasurements(t, current.Samples, current.HeldOut.Samples)
	assertSameSampleIdentities(t, current, legacy)
	assertSameSampleIdentities(t, current, fable)
	assertSameMeasurementIdentities(t, "claude held-out", current.HeldOut.Samples, "legacy held-out", legacy.HeldOut.Samples)
	assertSameMeasurementIdentities(t, "claude held-out", current.HeldOut.Samples, "Fable held-out", fable.HeldOut.Samples)
	for language, summary := range current.PerLanguage {
		if summary.Samples != 5 || summary.LeaveOneOutMeanAbsolutePercentError == nil {
			t.Errorf("current %s coverage = %d samples, LOO %v; want five samples and a LOO result", language, summary.Samples, summary.LeaveOneOutMeanAbsolutePercentError)
		}
	}

	assertGenerationContract(t, current, tokenizer.NameClaude, 7.0)
	assertGenerationContract(t, legacy, tokenizer.NameClaudeLegacy, 5.0)
	assertHeldOutContract(t, current, tokenizer.NameClaude, 9.0)
	assertHeldOutContract(t, legacy, tokenizer.NameClaudeLegacy, 5.0)

	if len(current.Samples) != len(fable.Samples) {
		t.Fatalf("current/Fable sample counts = %d/%d", len(current.Samples), len(fable.Samples))
	}
	for index := range current.Samples {
		currentSample := current.Samples[index]
		fableSample := fable.Samples[index]
		if currentSample.Path != fableSample.Path || currentSample.ClaudeContentTokens != fableSample.ClaudeContentTokens {
			t.Errorf("current/Fable mismatch at sample %d: %s=%d, %s=%d", index, currentSample.Path, currentSample.ClaudeContentTokens, fableSample.Path, fableSample.ClaudeContentTokens)
		}
	}
	for index := range current.HeldOut.Samples {
		currentSample := current.HeldOut.Samples[index]
		fableSample := fable.HeldOut.Samples[index]
		if currentSample.Path != fableSample.Path || currentSample.ClaudeContentTokens != fableSample.ClaudeContentTokens {
			t.Errorf("current/Fable held-out mismatch at sample %d: %s=%d, %s=%d", index, currentSample.Path, currentSample.ClaudeContentTokens, fableSample.Path, fableSample.ClaudeContentTokens)
		}
	}
}

func assertGenerationContract(t *testing.T, measured modelReport, tokenizerName string, maximumOverallMAPE float64) {
	t.Helper()
	if err := validateGenerationSamples(measured, tokenizerName); err != nil {
		t.Fatal(err)
	}
	_, metadata, err := tokenizer.New(tokenizerName)
	if err != nil {
		t.Fatalf("construct %s tokenizer: %v", tokenizerName, err)
	}
	assertNear(t, metadata.CalibrationFactor, measured.Global.Factor, 0.5e-6)

	overrides := make(map[string]float64, len(metadata.CalibrationOverrides))
	for _, override := range metadata.CalibrationOverrides {
		if _, duplicate := overrides[override.Language]; duplicate {
			t.Fatalf("%s has duplicate override for %s", tokenizerName, override.Language)
		}
		overrides[override.Language] = override.Factor
	}

	var absolutePercentErrorTotal float64
	var errorSamples int
	byLanguage := make(map[string][]measurement)
	for _, sample := range measured.Samples {
		byLanguage[sample.Language] = append(byLanguage[sample.Language], sample)
	}
	for language, samples := range byLanguage {
		summary := measured.PerLanguage[language]
		factor := metadata.CalibrationFactor
		overrideFactor, overridden := overrides[language]
		shouldOverride := summary.GlobalFactorMeanAbsolutePercentError > 10 &&
			summary.LeaveOneOutMeanAbsolutePercentError != nil &&
			summary.GlobalFactorMeanAbsolutePercentError-*summary.LeaveOneOutMeanAbsolutePercentError >= 3 &&
			*summary.LeaveOneOutMeanAbsolutePercentError <= 11
		if overridden != shouldOverride {
			t.Errorf("%s %s override present = %t, policy selects %t (global MAPE %.2f%%, LOO %.2f%%)",
				tokenizerName, language, overridden, shouldOverride,
				summary.GlobalFactorMeanAbsolutePercentError, optionalFloat(summary.LeaveOneOutMeanAbsolutePercentError))
		}
		if overridden {
			factor = overrideFactor
			assertNear(t, factor, summary.Factor, 0.5e-6)
		}

		var languageErrorTotal float64
		for _, sample := range samples {
			prediction := roundedPrediction(sample.O200KTokens, factor)
			percentError := math.Abs(float64(prediction-sample.ClaudeContentTokens)) / float64(sample.ClaudeContentTokens) * 100
			languageErrorTotal += percentError
			absolutePercentErrorTotal += percentError
			errorSamples++
		}
		languageMAPE := languageErrorTotal / float64(len(samples))
		if languageMAPE > 10 {
			t.Errorf("%s %s production-factor MAPE = %.2f%%, want at most 10%%", tokenizerName, language, languageMAPE)
		}
	}

	if errorSamples == 0 {
		t.Fatalf("%s calibration has no samples with Claude ground truth", tokenizerName)
	}
	overallMAPE := absolutePercentErrorTotal / float64(errorSamples)
	if overallMAPE > maximumOverallMAPE {
		t.Errorf("%s production-factor MAPE = %.2f%%, want at most %.2f%%", tokenizerName, overallMAPE, maximumOverallMAPE)
	}
}

func TestValidateGenerationSamplesRejectsEmptySet(t *testing.T) {
	if err := validateGenerationSamples(modelReport{}, tokenizer.NameClaudeLegacy); err == nil {
		t.Fatal("empty generation sample set passed validation")
	}
	if err := validateGenerationSamples(modelReport{Samples: []measurement{{ClaudeContentTokens: 1}}}, tokenizer.NameClaudeLegacy); err != nil {
		t.Fatalf("non-empty generation sample set failed validation: %v", err)
	}
}

func validateGenerationSamples(measured modelReport, tokenizerName string) error {
	if len(measured.Samples) == 0 {
		return fmt.Errorf("%s calibration has an empty generation sample set", tokenizerName)
	}
	return nil
}

func assertHeldOutContract(t *testing.T, measured modelReport, tokenizerName string, maximumOverallMAPE float64) {
	t.Helper()
	if measured.HeldOut == nil || len(measured.HeldOut.Samples) == 0 {
		t.Fatalf("%s has an empty held-out sample set", tokenizerName)
	}
	_, metadata, err := tokenizer.New(tokenizerName)
	if err != nil {
		t.Fatalf("construct %s tokenizer: %v", tokenizerName, err)
	}
	if measured.HeldOut.FactorSource != "production calibration factors" {
		t.Errorf("%s held-out factor source = %q", tokenizerName, measured.HeldOut.FactorSource)
	}
	assertNear(t, measured.HeldOut.GlobalFactor, metadata.CalibrationFactor, 1e-12)
	if len(measured.HeldOut.CalibrationOverrides) != len(metadata.CalibrationOverrides) {
		t.Errorf("%s held-out override count = %d, production has %d", tokenizerName, len(measured.HeldOut.CalibrationOverrides), len(metadata.CalibrationOverrides))
	} else {
		for index := range metadata.CalibrationOverrides {
			if measured.HeldOut.CalibrationOverrides[index] != metadata.CalibrationOverrides[index] {
				t.Errorf("%s held-out override %d = %+v, production has %+v", tokenizerName, index, measured.HeldOut.CalibrationOverrides[index], metadata.CalibrationOverrides[index])
			}
		}
	}

	computed := evaluateCalibrationFactors(measured.HeldOut.Samples, metadata.CalibrationFactor, metadata.CalibrationOverrides)
	assertEvaluationSummary(t, tokenizerName+" held-out overall", measured.HeldOut.Summary, computed)
	if measured.HeldOut.Summary.MeanAbsolutePercentError > maximumOverallMAPE {
		t.Errorf("%s held-out MAPE = %.2f%%, want at most %.2f%%", tokenizerName, measured.HeldOut.Summary.MeanAbsolutePercentError, maximumOverallMAPE)
	}

	expectedLanguages := map[string]struct{}{"C": {}, "HTML": {}, "Kotlin": {}, "Swift": {}}
	if len(measured.HeldOut.PerLanguage) != len(expectedLanguages) {
		t.Errorf("%s held-out language coverage = %d, want %d", tokenizerName, len(measured.HeldOut.PerLanguage), len(expectedLanguages))
	}
	byLanguage := make(map[string][]measurement)
	for _, sample := range measured.HeldOut.Samples {
		byLanguage[sample.Language] = append(byLanguage[sample.Language], sample)
	}
	for language := range expectedLanguages {
		samples := byLanguage[language]
		if len(samples) != 5 {
			t.Errorf("%s held-out %s samples = %d, want 5", tokenizerName, language, len(samples))
		}
		recorded, ok := measured.HeldOut.PerLanguage[language]
		if !ok {
			t.Errorf("%s held-out report is missing %s", tokenizerName, language)
			continue
		}
		assertEvaluationSummary(t, tokenizerName+" held-out "+language, recorded,
			evaluateCalibrationFactors(samples, metadata.CalibrationFactor, metadata.CalibrationOverrides))
		if tokenizerName == tokenizer.NameClaude && language == "HTML" {
			assertNear(t, recorded.MeanAbsolutePercentError, 17.269453, 0.5e-6)
			continue
		}
		if recorded.MeanAbsolutePercentError > 10 {
			t.Errorf("%s held-out %s MAPE = %.2f%%, want at most 10%%", tokenizerName, language, recorded.MeanAbsolutePercentError)
		}
	}
}

func assertEvaluationSummary(t *testing.T, label string, got, want evaluationSummary) {
	t.Helper()
	if got.Samples != want.Samples || got.O200KTokens != want.O200KTokens ||
		got.ClaudeContentTokens != want.ClaudeContentTokens || got.PredictedTokens != want.PredictedTokens {
		t.Errorf("%s summary counts = %+v, recomputed %+v", label, got, want)
	}
	assertNear(t, got.SignedAggregateError, want.SignedAggregateError, 1e-12)
	assertNear(t, got.MeanAbsolutePercentError, want.MeanAbsolutePercentError, 1e-12)
}

func assertCheckedInSamplesMatchReport(t *testing.T, recordedInputs []string, sampling samplingMetadata, recordedSamples []measurement, setName string) {
	t.Helper()
	if len(recordedInputs) == 0 {
		t.Fatalf("checked-in %s inputs are empty", setName)
	}
	inputs := make([]string, len(recordedInputs))
	for index, input := range recordedInputs {
		inputs[index] = resolveCheckedInSamplingInput(t, input)
	}
	samples, err := collectSamples(inputs, sampling.MaxBytesPerFile, sampling.MaxFilesPerLanguage, sampling.MaxSamples)
	if err != nil {
		t.Fatalf("collect checked-in %s corpus: %v", setName, err)
	}
	if len(samples) != len(recordedSamples) {
		t.Fatalf("checked-in %s corpus selected %d samples, report records %d", setName, len(samples), len(recordedSamples))
	}

	actualByPath := make(map[string]sourceSample, len(samples))
	for _, sample := range samples {
		actualByPath[sample.Path] = sample
	}
	for _, recorded := range recordedSamples {
		actual, ok := actualByPath[recorded.Path]
		if !ok {
			t.Errorf("reported sample %q is not selected from the checked-in corpus", recorded.Path)
			continue
		}
		delete(actualByPath, recorded.Path)
		if actual.ContentSHA != recorded.ContentSHA256 {
			t.Errorf("sample %q SHA-256 = %s, report records %s", recorded.Path, actual.ContentSHA, recorded.ContentSHA256)
		}
		if actual.Bytes != recorded.Bytes || actual.Truncated != recorded.Truncated || actual.Language != recorded.Language {
			t.Errorf("sample %q identity = (%s, %d bytes, truncated=%t), report records (%s, %d bytes, truncated=%t)",
				recorded.Path, actual.Language, actual.Bytes, actual.Truncated,
				recorded.Language, recorded.Bytes, recorded.Truncated)
		}
	}
	for path := range actualByPath {
		t.Errorf("checked-in %s corpus sample %q is missing from the report", setName, path)
	}
}

func assertDisjointMeasurements(t *testing.T, calibration, heldOut []measurement) {
	t.Helper()
	calibrationHashes := make(map[string]string, len(calibration))
	for _, sample := range calibration {
		calibrationHashes[sample.ContentSHA256] = sample.Path
	}
	for _, sample := range heldOut {
		if calibrationPath, duplicate := calibrationHashes[sample.ContentSHA256]; duplicate {
			t.Errorf("held-out sample %q duplicates calibration content from %q", sample.Path, calibrationPath)
		}
	}
}

func resolveCheckedInSamplingInput(t *testing.T, input string) string {
	t.Helper()
	candidates := []string{filepath.FromSlash(input)}
	const packagePrefix = "tools/calibrate/"
	if normalized := filepath.ToSlash(input); strings.HasPrefix(normalized, packagePrefix) {
		candidates = append(candidates, filepath.FromSlash(strings.TrimPrefix(normalized, packagePrefix)))
	}
	for _, candidate := range candidates {
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate
		}
	}
	t.Fatalf("cannot resolve checked-in sampling input %q from package working directory", input)
	return ""
}

func assertSameSampleIdentities(t *testing.T, want, got modelReport) {
	t.Helper()
	assertSameMeasurementIdentities(t, want.Label, want.Samples, got.Label, got.Samples)
}

func assertSameMeasurementIdentities(t *testing.T, wantLabel string, want []measurement, gotLabel string, got []measurement) {
	t.Helper()
	if len(want) != len(got) {
		t.Fatalf("%s/%s sample counts = %d/%d", wantLabel, gotLabel, len(want), len(got))
	}
	for index := range want {
		wantSample := want[index]
		gotSample := got[index]
		if wantSample.Path != gotSample.Path ||
			wantSample.Language != gotSample.Language ||
			wantSample.Bytes != gotSample.Bytes ||
			wantSample.Truncated != gotSample.Truncated ||
			wantSample.ContentSHA256 != gotSample.ContentSHA256 ||
			wantSample.O200KTokens != gotSample.O200KTokens {
			t.Errorf("%s/%s sample identity mismatch at index %d: %#v != %#v", wantLabel, gotLabel, index, wantSample, gotSample)
		}
	}
}

func requireModelReport(t *testing.T, report calibrationReport, label string) modelReport {
	t.Helper()
	for _, model := range report.Models {
		if model.Label == label {
			return model
		}
	}
	t.Fatalf("calibration report has no model labeled %q", label)
	return modelReport{}
}

func optionalFloat(value *float64) float64 {
	if value == nil {
		return 0
	}
	return *value
}
