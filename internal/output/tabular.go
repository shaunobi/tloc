package output

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/mattn/go-runewidth"

	"github.com/shaunobi/tloc/internal/model"
)

const (
	maxLanguageWidth = 24
	maxFileWidth     = 48
	maxFolderWidth   = 48
)

// WriteTabular writes a compact human-readable table followed by a tokenizer
// footer. Only the selected view's rows are included.
func WriteTabular(w io.Writer, report model.Report, view model.View) error {
	if !view.Valid() {
		return fmt.Errorf("unsupported view %d", view)
	}

	table := tabularTable(report, view)
	if err := writeTable(w, table); err != nil {
		return err
	}
	overrides := sortedCalibrationOverrides(report.Metadata.CalibrationOverrides)
	if len(overrides) > 0 {
		pairs := make([]string, 0, len(overrides))
		for _, override := range overrides {
			pairs = append(pairs, fmt.Sprintf("%s=%g", override.Language, override.Factor))
		}
		_, err := fmt.Fprintf(
			w,
			"Tokenizer: %s (estimated, default calibration factor %g; language overrides: %s)\n",
			report.Metadata.Tokenizer,
			report.Metadata.CalibrationFactor,
			strings.Join(pairs, ", "),
		)
		return err
	}
	if report.Metadata.Estimated {
		_, err := fmt.Fprintf(w, "Tokenizer: %s (estimated, calibration factor %g)\n", report.Metadata.Tokenizer, report.Metadata.CalibrationFactor)
		return err
	}
	_, err := fmt.Fprintf(w, "Tokenizer: %s\n", report.Metadata.Tokenizer)
	return err
}

type tableData struct {
	headings   []string
	rows       [][]string
	total      []string
	rightAlign []bool
}

func tabularTable(report model.Report, view model.View) tableData {
	switch view {
	case model.ViewFile:
		rows := make([][]string, 0, len(report.Files))
		for _, file := range report.Files {
			rows = append(rows, append([]string{
				truncateLeft(file.Path, maxFileWidth),
				truncate(file.Language, maxLanguageWidth),
			}, tabularMetrics(file.Metrics)...))
		}
		return tableData{
			headings:   []string{"File", "Language", "Files", "Lines", "Code", "Tokens", "Tok/Line"},
			rows:       rows,
			total:      append([]string{"Total", ""}, tabularMetrics(report.Totals)...),
			rightAlign: []bool{false, false, true, true, true, true, true},
		}
	case model.ViewFolder:
		rows := make([][]string, 0, len(report.Folders))
		for _, folder := range report.Folders {
			label := strings.Repeat("  ", max(0, folder.Depth)) + truncate(folder.Name, maxFolderWidth)
			rows = append(rows, append([]string{label}, tabularMetrics(folder.Metrics)...))
		}
		return tableData{
			headings:   []string{"Folder", "Files", "Lines", "Code", "Tokens", "Tok/Line"},
			rows:       rows,
			total:      append([]string{"Total"}, tabularMetrics(report.Totals)...),
			rightAlign: []bool{false, true, true, true, true, true},
		}
	default:
		rows := make([][]string, 0, len(report.Languages))
		for _, language := range report.Languages {
			rows = append(rows, append([]string{truncate(language.Language, maxLanguageWidth)}, tabularMetrics(language.Metrics)...))
		}
		return tableData{
			headings:   []string{"Language", "Files", "Lines", "Code", "Tokens", "Tok/Line"},
			rows:       rows,
			total:      append([]string{"Total"}, tabularMetrics(report.Totals)...),
			rightAlign: []bool{false, true, true, true, true, true},
		}
	}
}

func tabularMetrics(metrics model.Metrics) []string {
	return []string{
		strconv.FormatInt(metrics.Files, 10),
		strconv.FormatInt(metrics.Lines, 10),
		strconv.FormatInt(metrics.Code, 10),
		strconv.FormatInt(metrics.Tokens, 10),
		strconv.FormatFloat(metrics.TokensPerCodeLine(), 'f', 1, 64),
	}
}

func writeTable(w io.Writer, table tableData) error {
	widths := make([]int, len(table.headings))
	for index, heading := range table.headings {
		widths[index] = runewidth.StringWidth(heading)
	}
	measure := func(row []string) {
		for index, cell := range row {
			if width := runewidth.StringWidth(cell); width > widths[index] {
				widths[index] = width
			}
		}
	}
	for _, row := range table.rows {
		measure(row)
	}
	measure(table.total)

	if err := writeTableRow(w, table.headings, widths, table.rightAlign); err != nil {
		return err
	}
	separator := make([]string, len(widths))
	for index, width := range widths {
		separator[index] = strings.Repeat("-", width)
	}
	if err := writeTableRow(w, separator, widths, table.rightAlign); err != nil {
		return err
	}
	for _, row := range table.rows {
		if err := writeTableRow(w, row, widths, table.rightAlign); err != nil {
			return err
		}
	}
	if err := writeTableRow(w, separator, widths, table.rightAlign); err != nil {
		return err
	}
	return writeTableRow(w, table.total, widths, table.rightAlign)
}

func writeTableRow(w io.Writer, row []string, widths []int, rightAlign []bool) error {
	for index, cell := range row {
		if index > 0 {
			if _, err := io.WriteString(w, "  "); err != nil {
				return err
			}
		}
		padding := strings.Repeat(" ", max(0, widths[index]-runewidth.StringWidth(cell)))
		if rightAlign[index] {
			if _, err := io.WriteString(w, padding); err != nil {
				return err
			}
		}
		if _, err := io.WriteString(w, cell); err != nil {
			return err
		}
		if !rightAlign[index] {
			if _, err := io.WriteString(w, padding); err != nil {
				return err
			}
		}
	}
	_, err := io.WriteString(w, "\n")
	return err
}

func truncate(value string, maxWidth int) string {
	if maxWidth <= 0 || runewidth.StringWidth(value) <= maxWidth {
		return value
	}
	if maxWidth == 1 {
		return "\u2026"
	}
	return runewidth.Truncate(value, maxWidth, "\u2026")
}

// truncateLeft preserves the distinguishing tail of a path while constraining
// the result to maxWidth terminal cells. TruncateLeft operates on grapheme
// clusters and pads when a wide cluster straddles the cut, so it cannot split a
// displayed character or exceed the requested width.
func truncateLeft(value string, maxWidth int) string {
	width := runewidth.StringWidth(value)
	if maxWidth <= 0 || width <= maxWidth {
		return value
	}
	if maxWidth == 1 {
		return "\u2026"
	}

	keptWidth := maxWidth - runewidth.StringWidth("\u2026")
	return runewidth.TruncateLeft(value, width-keptWidth, "\u2026")
}
