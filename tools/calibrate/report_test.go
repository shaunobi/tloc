package main

import (
	"encoding/json"
	"math"
	"strings"
	"testing"
	"time"
)

func TestSummarizeUsesRatioOfTotals(t *testing.T) {
	measurements := []measurement{
		{O200KTokens: 100, ClaudeContentTokens: 120, ClaudePerO200KRatio: 1.2},
		{O200KTokens: 50, ClaudeContentTokens: 55, ClaudePerO200KRatio: 1.1},
	}
	summary := summarize(measurements)
	if summary.Samples != 2 || summary.O200KTokens != 150 || summary.ClaudeContentTokens != 175 {
		t.Errorf("summary totals = %+v", summary)
	}
	assertNear(t, summary.Factor, 175.0/150.0, 1e-12)
	assertNear(t, summary.MeanPerFileRatio, 1.15, 1e-12)
	assertNear(t, summary.MedianPerFileRatio, 1.15, 1e-12)
	// Predictions are round(100*factor)=117 and round(50*factor)=58.
	assertNear(t, summary.MeanAbsolutePercentError, (3.0/120.0*100+3.0/55.0*100)/2, 1e-12)
}

func TestEvaluateFactorUsesRoundedPerFilePredictions(t *testing.T) {
	measurements := []measurement{
		{O200KTokens: 10, ClaudeContentTokens: 14},
		{O200KTokens: 5, ClaudeContentTokens: 5},
	}

	errors := evaluateFactor(measurements, 1.2)
	// Per-file predictions are 12 and 6, so the signed aggregate error is
	// (18-19)/19 and the per-file absolute errors are 2/14 and 1/5.
	assertNear(t, errors.AggregatePercentError, -1.0/19.0*100, 1e-12)
	assertNear(t, errors.MeanAbsolutePercentError, (2.0/14.0*100+1.0/5.0*100)/2, 1e-12)
}

func TestLeaveOneOutMeanAbsolutePercentError(t *testing.T) {
	measurements := []measurement{
		{O200KTokens: 100, ClaudeContentTokens: 120},
		{O200KTokens: 50, ClaudeContentTokens: 55},
		{O200KTokens: 20, ClaudeContentTokens: 24},
	}

	got, ok := leaveOneOutMeanAbsolutePercentError(measurements)
	if !ok {
		t.Fatal("leave-one-out MAPE unavailable, want a value")
	}
	// Predictions are round(100*79/70)=113, round(50*144/120)=60,
	// and round(20*175/150)=23.
	want := (7.0/120.0*100 + 5.0/55.0*100 + 1.0/24.0*100) / 3
	assertNear(t, got, want, 1e-12)

	if got, ok := leaveOneOutMeanAbsolutePercentError(measurements[:1]); ok {
		t.Fatalf("single-sample leave-one-out MAPE = %f, want unavailable", got)
	}
}

func TestSummarizeLanguageEvaluatesGlobalFactorAndLeaveOneOut(t *testing.T) {
	measurements := []measurement{
		{O200KTokens: 10, ClaudeContentTokens: 14, ClaudePerO200KRatio: 1.4},
		{O200KTokens: 5, ClaudeContentTokens: 5, ClaudePerO200KRatio: 1},
	}

	summary := summarizeLanguage(measurements, 1.2)
	assertNear(t, summary.GlobalFactorSignedAggregatePercentError, -1.0/19.0*100, 1e-12)
	assertNear(t, summary.GlobalFactorMeanAbsolutePercentError, (2.0/14.0*100+1.0/5.0*100)/2, 1e-12)
	if summary.LeaveOneOutMeanAbsolutePercentError == nil {
		t.Fatal("leave-one-out MAPE is nil, want a value")
	}
	// Holding out the first sample predicts 10 tokens; holding out the second
	// predicts 7 tokens.
	assertNear(t, *summary.LeaveOneOutMeanAbsolutePercentError, (4.0/14.0*100+2.0/5.0*100)/2, 1e-12)
}

func TestFinalizeModelReportAndMarkdown(t *testing.T) {
	report := modelReport{
		Label:           "claude",
		Model:           "model-id",
		FramingBaseline: 7,
		Samples: []measurement{
			{Path: "a.go", Language: "Go", O200KTokens: 10, ClaudeContentTokens: 12, ClaudePerO200KRatio: 1.2},
			{Path: "b.py", Language: "Python", O200KTokens: 20, ClaudeContentTokens: 22, ClaudePerO200KRatio: 1.1},
		},
	}
	finalizeModelReport(&report)
	if len(report.PerLanguage) != 2 || report.PerLanguage["Go"].Samples != 1 {
		t.Errorf("per-language summaries = %+v", report.PerLanguage)
	}
	assertNear(t, report.PerLanguage["Go"].GlobalFactorSignedAggregatePercentError, -1.0/12.0*100, 1e-12)
	assertNear(t, report.PerLanguage["Go"].GlobalFactorMeanAbsolutePercentError, 1.0/12.0*100, 1e-12)
	if report.PerLanguage["Go"].LeaveOneOutMeanAbsolutePercentError != nil {
		t.Errorf("single-sample LOO MAPE = %v, want nil", *report.PerLanguage["Go"].LeaveOneOutMeanAbsolutePercentError)
	}

	markdown := renderMarkdown(calibrationReport{
		GeneratedAt: time.Date(2026, 7, 22, 12, 0, 0, 0, time.UTC),
		Method:      "test method",
		Models:      []modelReport{report},
	})
	for _, expected := range []string{
		"model-id", "Framing baseline", "Python", "a.go",
		"Global-factor signed aggregate error", "Global-factor MAPE", "LOO language MAPE",
		"-8.33%", "N/A (<2 samples)",
	} {
		if !strings.Contains(markdown, expected) {
			t.Errorf("markdown missing %q", expected)
		}
	}

	jsonData, err := json.Marshal(report)
	if err != nil {
		t.Fatalf("marshal report: %v", err)
	}
	var encoded map[string]any
	if err := json.Unmarshal(jsonData, &encoded); err != nil {
		t.Fatalf("unmarshal report JSON: %v", err)
	}
	perLanguage := encoded["per_language"].(map[string]any)
	goSummary := perLanguage["Go"].(map[string]any)
	assertNear(t, goSummary["global_factor_signed_aggregate_percent_error"].(float64), -1.0/12.0*100, 1e-12)
	assertNear(t, goSummary["global_factor_mean_absolute_percent_error"].(float64), 1.0/12.0*100, 1e-12)
	if goSummary["leave_one_out_mean_absolute_percent_error"] != nil {
		t.Errorf("JSON LOO MAPE = %v, want null", goSummary["leave_one_out_mean_absolute_percent_error"])
	}
	for _, legacyField := range []string{"samples", "o200k_tokens", "claude_content_tokens", "factor", "mean_per_file_ratio", "median_per_file_ratio", "mean_absolute_percent_error"} {
		if _, ok := goSummary[legacyField]; !ok {
			t.Errorf("JSON missing existing field %q", legacyField)
		}
	}
}

func TestFinalizeHeldOutEvaluationUsesFinalizedTrainingFactor(t *testing.T) {
	report := modelReport{
		Label:  "spot-model",
		Global: ratioSummary{Factor: 1.2},
	}
	heldOut := []measurement{
		{Path: "holdout/a.go", Language: "Go", O200KTokens: 10, ClaudeContentTokens: 15, ClaudePerO200KRatio: 1.5},
		{Path: "holdout/b.py", Language: "Python", O200KTokens: 20, ClaudeContentTokens: 20, ClaudePerO200KRatio: 1},
	}
	if err := finalizeHeldOutEvaluation(&report, heldOut); err != nil {
		t.Fatal(err)
	}
	if report.HeldOut == nil {
		t.Fatal("held-out evaluation is nil")
	}
	if report.HeldOut.FactorSource != "training global factor (no production mapping)" || report.HeldOut.GlobalFactor != 1.2 {
		t.Fatalf("held-out factor plan = %+v", report.HeldOut)
	}
	if report.HeldOut.Summary.Samples != 2 || report.HeldOut.Summary.PredictedTokens != 36 {
		t.Fatalf("held-out summary = %+v", report.HeldOut.Summary)
	}
	assertNear(t, report.HeldOut.Summary.SignedAggregateError, 1.0/35.0*100, 1e-12)
	assertNear(t, report.HeldOut.Summary.MeanAbsolutePercentError, 20, 1e-12)
	if report.HeldOut.PerLanguage["Go"].PredictedTokens != 12 || report.HeldOut.PerLanguage["Python"].PredictedTokens != 24 {
		t.Fatalf("held-out language summaries = %+v", report.HeldOut.PerLanguage)
	}

	markdown := renderMarkdown(calibrationReport{Models: []modelReport{report}})
	for _, expected := range []string{
		"Held-out evaluation", "not used to fit or select", "training global factor",
		"holdout/a.go", "**Overall**", "+2.86%", "20.00%",
	} {
		if !strings.Contains(markdown, expected) {
			t.Errorf("held-out Markdown missing %q", expected)
		}
	}
}

func TestFinalizeHeldOutEvaluationUsesProductionOverrides(t *testing.T) {
	report := modelReport{Label: "claude", Global: ratioSummary{Factor: 99}}
	heldOut := []measurement{
		{Language: "C#", O200KTokens: 10, ClaudeContentTokens: 19},
		{Language: "Go", O200KTokens: 10, ClaudeContentTokens: 17},
	}
	if err := finalizeHeldOutEvaluation(&report, heldOut); err != nil {
		t.Fatal(err)
	}
	if report.HeldOut.FactorSource != "production calibration factors" {
		t.Fatalf("factor source = %q", report.HeldOut.FactorSource)
	}
	if report.HeldOut.GlobalFactor != 1.654377 {
		t.Fatalf("global factor = %.6f", report.HeldOut.GlobalFactor)
	}
	if report.HeldOut.PerLanguage["C#"].PredictedTokens != 19 {
		t.Errorf("C# override prediction = %d, want 19", report.HeldOut.PerLanguage["C#"].PredictedTokens)
	}
	if report.HeldOut.PerLanguage["Go"].PredictedTokens != 17 {
		t.Errorf("Go fallback prediction = %d, want 17", report.HeldOut.PerLanguage["Go"].PredictedTokens)
	}
}

func TestFramingBaselineFromProbe(t *testing.T) {
	for _, test := range []struct {
		name string
		raw  int64
		want int64
	}{
		{name: "typical framing", raw: 12, want: 11},
		{name: "zero framing", raw: 1, want: 0},
	} {
		t.Run(test.name, func(t *testing.T) {
			baseline, err := framingBaselineFromProbe(test.raw)
			if err != nil {
				t.Fatalf("framingBaselineFromProbe(%d): %v", test.raw, err)
			}
			if baseline != test.want {
				t.Fatalf("framingBaselineFromProbe(%d) = %d, want %d", test.raw, baseline, test.want)
			}
		})
	}

	for _, raw := range []int64{0, -1} {
		if _, err := framingBaselineFromProbe(raw); err == nil {
			t.Errorf("framingBaselineFromProbe(%d) succeeded, want an error", raw)
		}
	}
}

func assertNear(t *testing.T, got, want, tolerance float64) {
	t.Helper()
	if math.Abs(got-want) > tolerance {
		t.Errorf("got %.12f, want %.12f (tolerance %g)", got, want, tolerance)
	}
}
