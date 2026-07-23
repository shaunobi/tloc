package output

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"

	"github.com/shaunobi/tloc/internal/model"
)

var metricHeadings = []string{"files", "lines", "code", "comments", "blanks", "complexity", "bytes", "tokens"}
var statusHeadings = []string{"record_type", "complete", "skipped_stage", "skipped_path", "skipped_error"}

// WriteCSV writes the selected view's records. CSV intentionally has no
// synthetic totals row; folder records are cumulative and must not be summed.
func WriteCSV(w io.Writer, report model.Report, view model.View) error {
	if !view.Valid() {
		return fmt.Errorf("unsupported view %d", view)
	}

	writer := csv.NewWriter(w)
	write := func(record []string) error {
		if err := writer.Write(record); err != nil {
			return err
		}
		return nil
	}

	switch view {
	case model.ViewFile:
		identityWidth := 2
		if err := write(csvHeader([]string{"language", "path"})); err != nil {
			return err
		}
		for _, file := range report.Files {
			if err := write(csvDataRow([]string{file.Language, file.Path}, file.Metrics, report.Metadata.Complete)); err != nil {
				return err
			}
		}
		if err := writeSkippedRows(write, identityWidth, report.Metadata.Skipped); err != nil {
			return err
		}
	case model.ViewFolder:
		identityWidth := 4
		if err := write(csvHeader([]string{"folder", "input_id", "depth", "synthetic"})); err != nil {
			return err
		}
		for _, folder := range report.Folders {
			identity := []string{
				folder.Path,
				strconv.Itoa(folder.InputID),
				strconv.Itoa(folder.Depth),
				strconv.FormatBool(folder.Synthetic),
			}
			if err := write(csvDataRow(identity, folder.Metrics, report.Metadata.Complete)); err != nil {
				return err
			}
		}
		if err := writeSkippedRows(write, identityWidth, report.Metadata.Skipped); err != nil {
			return err
		}
	default:
		identityWidth := 1
		if err := write(csvHeader([]string{"language"})); err != nil {
			return err
		}
		for _, language := range report.Languages {
			if err := write(csvDataRow([]string{language.Language}, language.Metrics, report.Metadata.Complete)); err != nil {
				return err
			}
		}
		if err := writeSkippedRows(write, identityWidth, report.Metadata.Skipped); err != nil {
			return err
		}
	}
	writer.Flush()
	return writer.Error()
}

func csvHeader(identity []string) []string {
	header := append(slicesClone(identity), metricHeadings...)
	return append(header, statusHeadings...)
}

func csvDataRow(identity []string, metrics model.Metrics, complete bool) []string {
	row := append(slicesClone(identity), csvMetrics(metrics)...)
	return append(row, "data", strconv.FormatBool(complete), "", "", "")
}

func writeSkippedRows(write func([]string) error, identityWidth int, skipped []model.SkippedEntry) error {
	blankWidth := identityWidth + len(metricHeadings)
	for _, entry := range skipped {
		row := make([]string, blankWidth)
		row = append(row, "skipped", "false", entry.Stage, entry.Path, entry.Error)
		if err := write(row); err != nil {
			return err
		}
	}
	return nil
}

func slicesClone(values []string) []string {
	return append([]string(nil), values...)
}

func csvMetrics(metrics model.Metrics) []string {
	return []string{
		strconv.FormatInt(metrics.Files, 10),
		strconv.FormatInt(metrics.Lines, 10),
		strconv.FormatInt(metrics.Code, 10),
		strconv.FormatInt(metrics.Comments, 10),
		strconv.FormatInt(metrics.Blanks, 10),
		strconv.FormatInt(metrics.Complexity, 10),
		strconv.FormatInt(metrics.Bytes, 10),
		strconv.FormatInt(metrics.Tokens, 10),
	}
}
