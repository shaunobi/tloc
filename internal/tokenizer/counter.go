// Package tokenizer counts source text using the tokenizers exposed by tloc.
package tokenizer

import (
	"fmt"
	"strings"
)

const (
	// NameO200K is the exact OpenAI o200k_base tokenizer.
	NameO200K = "o200k"
	// NameCodex is a compatibility alias for NameO200K.
	NameCodex = "codex"
	// NameClaude is the estimator for the current Claude tokenizer generation.
	NameClaude = "claude"
	// NameClaudeLegacy is the estimator for pre-Opus-4.7 Claude models.
	NameClaudeLegacy = "claude-legacy"
)

// Counter counts tokens in content. Implementations must be safe for concurrent
// use by multiple goroutines.
type Counter interface {
	Count(content []byte) (int64, error)
}

// LanguageCounter optionally provides language-specific token estimates. The
// language must be the exact canonical label reported by scc.
type LanguageCounter interface {
	Counter
	CountForLanguage(content []byte, language string) (int64, error)
}

// CountForLanguage uses a counter's language-specific implementation when it
// has one, and otherwise falls back to Count.
func CountForLanguage(counter Counter, content []byte, language string) (int64, error) {
	if languageCounter, ok := counter.(LanguageCounter); ok {
		return languageCounter.CountForLanguage(content, language)
	}
	return counter.Count(content)
}

// CalibrationOverride applies a language-specific factor to an estimated
// tokenizer. Language is an exact canonical scc label.
type CalibrationOverride struct {
	Language string  `json:"language"`
	Factor   float64 `json:"factor"`
}

// Metadata describes the tokenizer behind a Counter.
type Metadata struct {
	Name                 string                `json:"name"`
	Encoding             string                `json:"encoding"`
	Estimated            bool                  `json:"estimated"`
	CalibrationFactor    float64               `json:"calibration_factor"`
	CalibrationOverrides []CalibrationOverride `json:"calibration_overrides,omitempty"`
}

// Supported returns the accepted tokenizer names in CLI display order.
func Supported() []string {
	return []string{NameClaude, NameClaudeLegacy, NameO200K, NameCodex}
}

// New constructs a counter and its reporting metadata. The codex alias is
// canonicalized to o200k in the returned metadata.
func New(name string) (Counter, Metadata, error) {
	canonical := strings.ToLower(strings.TrimSpace(name))
	if canonical != NameO200K && canonical != NameCodex && canonical != NameClaude && canonical != NameClaudeLegacy {
		return nil, Metadata{}, fmt.Errorf("unknown tokenizer %q (supported: %s)", name, strings.Join(Supported(), ", "))
	}

	base, err := newO200KCounter()
	if err != nil {
		return nil, Metadata{}, fmt.Errorf("initialize o200k tokenizer: %w", err)
	}

	switch canonical {
	case NameO200K, NameCodex:
		return base, Metadata{
			Name:              NameO200K,
			Encoding:          "o200k_base",
			Estimated:         false,
			CalibrationFactor: 1,
		}, nil
	case NameClaude:
		if !ClaudeCurrentCalibrationReady {
			return nil, Metadata{}, fmt.Errorf("%s tokenizer is unavailable: calibration has not been completed", NameClaude)
		}
		counter, err := newEstimator(base, ClaudeCurrentCalibrationFactor, claudeCurrentCalibrationOverrides[:])
		if err != nil {
			return nil, Metadata{}, err
		}
		return counter, Metadata{
			Name:                 NameClaude,
			Encoding:             "o200k_base",
			Estimated:            true,
			CalibrationFactor:    ClaudeCurrentCalibrationFactor,
			CalibrationOverrides: counter.calibrationOverridesCopy(),
		}, nil
	case NameClaudeLegacy:
		if !ClaudeLegacyCalibrationReady {
			return nil, Metadata{}, fmt.Errorf("%s tokenizer is unavailable: calibration has not been completed", NameClaudeLegacy)
		}
		counter, err := newEstimator(base, ClaudeLegacyCalibrationFactor, claudeLegacyCalibrationOverrides[:])
		if err != nil {
			return nil, Metadata{}, err
		}
		return counter, Metadata{
			Name:                 NameClaudeLegacy,
			Encoding:             "o200k_base",
			Estimated:            true,
			CalibrationFactor:    ClaudeLegacyCalibrationFactor,
			CalibrationOverrides: counter.calibrationOverridesCopy(),
		}, nil
	}
	panic("unreachable tokenizer name")
}
