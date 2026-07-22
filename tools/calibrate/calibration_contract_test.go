package main

import (
	"encoding/json"
	"math"
	"os"
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
	for language, summary := range current.PerLanguage {
		if summary.Samples != 5 || summary.LeaveOneOutMeanAbsolutePercentError == nil {
			t.Errorf("current %s coverage = %d samples, LOO %v; want five samples and a LOO result", language, summary.Samples, summary.LeaveOneOutMeanAbsolutePercentError)
		}
	}

	assertGenerationContract(t, current, tokenizer.NameClaude, 7.0)
	assertGenerationContract(t, legacy, tokenizer.NameClaudeLegacy, 5.0)

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
}

func assertGenerationContract(t *testing.T, measured modelReport, tokenizerName string, maximumOverallMAPE float64) {
	t.Helper()
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

	overallMAPE := absolutePercentErrorTotal / float64(errorSamples)
	if overallMAPE > maximumOverallMAPE {
		t.Errorf("%s production-factor MAPE = %.2f%%, want at most %.2f%%", tokenizerName, overallMAPE, maximumOverallMAPE)
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
