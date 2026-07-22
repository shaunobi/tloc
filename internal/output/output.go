// Package output renders reports as tabular text, JSON, or CSV.
package output

import (
	"fmt"
	"io"

	"github.com/shaunobi/tloc/internal/model"
)

// Format is a supported output serialization.
type Format string

const (
	FormatTabular Format = "tabular"
	FormatJSON    Format = "json"
	FormatCSV     Format = "csv"
)

// Valid reports whether f is supported.
func (f Format) Valid() bool {
	switch f {
	case FormatTabular, FormatJSON, FormatCSV:
		return true
	default:
		return false
	}
}

// Write renders report in format using view as the primary row set.
func Write(w io.Writer, report model.Report, view model.View, format Format) error {
	if !view.Valid() {
		return fmt.Errorf("unsupported view %d", view)
	}
	switch format {
	case FormatTabular:
		return WriteTabular(w, report, view)
	case FormatJSON:
		return WriteJSON(w, report, view)
	case FormatCSV:
		return WriteCSV(w, report, view)
	default:
		return fmt.Errorf("unsupported format %q", format)
	}
}
