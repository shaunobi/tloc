package tokenizer

import (
	"fmt"
	"math"
	"slices"
	"strings"
)

type estimator struct {
	base      Counter
	factor    float64
	overrides []CalibrationOverride
}

func newEstimator(base Counter, factor float64, overrides []CalibrationOverride) (*estimator, error) {
	if base == nil {
		return nil, fmt.Errorf("estimator base counter is nil")
	}
	if !validCalibrationFactor(factor) {
		return nil, fmt.Errorf("invalid calibration factor %v", factor)
	}

	ownedOverrides := slices.Clone(overrides)
	for _, override := range ownedOverrides {
		trimmedLanguage := strings.TrimSpace(override.Language)
		if trimmedLanguage == "" {
			return nil, fmt.Errorf("calibration override language is empty")
		}
		if trimmedLanguage != override.Language {
			return nil, fmt.Errorf("calibration override language %q has surrounding whitespace", override.Language)
		}
		if !validCalibrationFactor(override.Factor) {
			return nil, fmt.Errorf("invalid calibration factor %v for language %q", override.Factor, override.Language)
		}
	}
	slices.SortFunc(ownedOverrides, func(a, b CalibrationOverride) int {
		return strings.Compare(a.Language, b.Language)
	})
	for index := 1; index < len(ownedOverrides); index++ {
		if ownedOverrides[index-1].Language == ownedOverrides[index].Language {
			return nil, fmt.Errorf("duplicate calibration override for language %q", ownedOverrides[index].Language)
		}
	}

	return &estimator{base: base, factor: factor, overrides: ownedOverrides}, nil
}

func (e *estimator) Count(content []byte) (int64, error) {
	return e.countWithFactor(content, e.factor)
}

func (e *estimator) CountForLanguage(content []byte, language string) (int64, error) {
	factor := e.factor
	if index, found := slices.BinarySearchFunc(e.overrides, language, func(override CalibrationOverride, language string) int {
		return strings.Compare(override.Language, language)
	}); found {
		factor = e.overrides[index].Factor
	}
	return e.countWithFactor(content, factor)
}

func (e *estimator) calibrationOverridesCopy() []CalibrationOverride {
	return slices.Clone(e.overrides)
}

func (e *estimator) countWithFactor(content []byte, factor float64) (int64, error) {
	baseCount, err := e.base.Count(content)
	if err != nil {
		return 0, err
	}
	if baseCount < 0 {
		return 0, fmt.Errorf("base tokenizer returned negative token count %d", baseCount)
	}

	scaled := float64(baseCount) * factor
	if scaled >= float64(math.MaxInt64) {
		return 0, fmt.Errorf("estimated token count overflows int64")
	}
	return int64(math.Round(scaled)), nil
}

func validCalibrationFactor(factor float64) bool {
	return factor > 0 && !math.IsNaN(factor) && !math.IsInf(factor, 0)
}
