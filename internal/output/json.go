package output

import (
	"cmp"
	"encoding/json"
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/shaunobi/tloc/internal/model"
)

// WriteJSON writes a single pretty-printed JSON report. Languages and totals
// are always present; the selected detail view is present even when empty.
func WriteJSON(w io.Writer, report model.Report, view model.View) error {
	if !view.Valid() {
		return fmt.Errorf("unsupported view %d", view)
	}

	overrides := sortedCalibrationOverrides(report.Metadata.CalibrationOverrides)
	document := jsonDocument{
		Languages: make([]jsonLanguage, 0, len(report.Languages)),
		Totals:    newJSONMetrics(report.Totals),
		Metadata: jsonMetadata{
			Version:              report.Metadata.Version,
			Tokenizer:            report.Metadata.Tokenizer,
			CalibrationFactor:    report.Metadata.CalibrationFactor,
			Estimated:            report.Metadata.Estimated,
			CalibrationOverrides: make([]jsonCalibrationOverride, 0, len(overrides)),
		},
	}
	for _, override := range overrides {
		document.Metadata.CalibrationOverrides = append(document.Metadata.CalibrationOverrides, jsonCalibrationOverride{
			Language: override.Language,
			Factor:   override.Factor,
		})
	}
	for _, language := range report.Languages {
		document.Languages = append(document.Languages, jsonLanguage{
			Language:    language.Language,
			jsonMetrics: newJSONMetrics(language.Metrics),
		})
	}
	if view == model.ViewFile {
		files := make([]jsonFile, 0, len(report.Files))
		for _, file := range report.Files {
			files = append(files, jsonFile{
				Language:    file.Language,
				Path:        file.Path,
				jsonMetrics: newJSONMetrics(file.Metrics),
			})
		}
		document.Files = &files
	}
	if view == model.ViewFolder {
		folders := make([]jsonFolder, 0, len(report.Folders))
		for _, folder := range report.Folders {
			folders = append(folders, jsonFolder{
				Folder:      folder.Path,
				InputID:     folder.InputID,
				Depth:       folder.Depth,
				Synthetic:   folder.Synthetic,
				jsonMetrics: newJSONMetrics(folder.Metrics),
			})
		}
		document.Folders = &folders
	}

	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	return encoder.Encode(document)
}

type jsonDocument struct {
	Languages []jsonLanguage `json:"languages"`
	Files     *[]jsonFile    `json:"files,omitempty"`
	Folders   *[]jsonFolder  `json:"folders,omitempty"`
	Totals    jsonMetrics    `json:"totals"`
	Metadata  jsonMetadata   `json:"metadata"`
}

type jsonMetrics struct {
	Files      int64 `json:"files"`
	Lines      int64 `json:"lines"`
	Code       int64 `json:"code"`
	Comments   int64 `json:"comments"`
	Blanks     int64 `json:"blanks"`
	Complexity int64 `json:"complexity"`
	Bytes      int64 `json:"bytes"`
	Tokens     int64 `json:"tokens"`
}

type jsonLanguage struct {
	Language string `json:"language"`
	jsonMetrics
}

type jsonFile struct {
	Language string `json:"language"`
	Path     string `json:"path"`
	jsonMetrics
}

type jsonFolder struct {
	Folder    string `json:"folder"`
	InputID   int    `json:"input_id"`
	Depth     int    `json:"depth"`
	Synthetic bool   `json:"synthetic"`
	jsonMetrics
}

type jsonMetadata struct {
	Version              string                    `json:"version"`
	Tokenizer            string                    `json:"tokenizer"`
	CalibrationFactor    float64                   `json:"calibration_factor"`
	Estimated            bool                      `json:"estimated,omitempty"`
	CalibrationOverrides []jsonCalibrationOverride `json:"calibration_overrides,omitempty"`
}

type jsonCalibrationOverride struct {
	Language string  `json:"language"`
	Factor   float64 `json:"factor"`
}

func newJSONMetrics(metrics model.Metrics) jsonMetrics {
	return jsonMetrics{
		Files:      metrics.Files,
		Lines:      metrics.Lines,
		Code:       metrics.Code,
		Comments:   metrics.Comments,
		Blanks:     metrics.Blanks,
		Complexity: metrics.Complexity,
		Bytes:      metrics.Bytes,
		Tokens:     metrics.Tokens,
	}
}

func sortedCalibrationOverrides(overrides []model.CalibrationOverride) []model.CalibrationOverride {
	result := slices.Clone(overrides)
	slices.SortFunc(result, func(a, b model.CalibrationOverride) int {
		return cmp.Or(strings.Compare(a.Language, b.Language), cmp.Compare(a.Factor, b.Factor))
	})
	return result
}
