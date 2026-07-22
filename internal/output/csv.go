package output

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"

	"github.com/shaunobi/tloc/internal/model"
)

var metricHeadings = []string{"files", "lines", "code", "comments", "blanks", "complexity", "bytes", "tokens"}

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
		if err := write(append([]string{"language", "path"}, metricHeadings...)); err != nil {
			return err
		}
		for _, file := range report.Files {
			if err := write(append([]string{file.Language, file.Path}, csvMetrics(file.Metrics)...)); err != nil {
				return err
			}
		}
	case model.ViewFolder:
		if err := write(append([]string{"folder", "input_id", "depth", "synthetic"}, metricHeadings...)); err != nil {
			return err
		}
		for _, folder := range report.Folders {
			identity := []string{
				folder.Path,
				strconv.Itoa(folder.InputID),
				strconv.Itoa(folder.Depth),
				strconv.FormatBool(folder.Synthetic),
			}
			if err := write(append(identity, csvMetrics(folder.Metrics)...)); err != nil {
				return err
			}
		}
	default:
		if err := write(append([]string{"language"}, metricHeadings...)); err != nil {
			return err
		}
		for _, language := range report.Languages {
			if err := write(append([]string{language.Language}, csvMetrics(language.Metrics)...)); err != nil {
				return err
			}
		}
	}
	writer.Flush()
	return writer.Error()
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
