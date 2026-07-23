// Package model contains the reporting domain shared by the analyzer,
// aggregators, and output renderers.
package model

// Metrics contains additive measurements for a file or a collection of files.
// All counters use int64 so large repositories can be accumulated safely.
type Metrics struct {
	Files      int64
	Lines      int64
	Code       int64
	Comments   int64
	Blanks     int64
	Complexity int64
	Bytes      int64
	Tokens     int64
}

// Add adds every counter in other to m.
func (m *Metrics) Add(other Metrics) {
	m.Files += other.Files
	m.Lines += other.Lines
	m.Code += other.Code
	m.Comments += other.Comments
	m.Blanks += other.Blanks
	m.Complexity += other.Complexity
	m.Bytes += other.Bytes
	m.Tokens += other.Tokens
}

// TokensPerCodeLine returns token density using code lines as the denominator.
// A record with no code lines has a density of zero.
func (m Metrics) TokensPerCodeLine() float64 {
	if m.Code == 0 {
		return 0
	}
	return float64(m.Tokens) / float64(m.Code)
}

// InputKind describes whether a command-line input is a directory or file.
type InputKind string

const (
	InputDirectory InputKind = "directory"
	InputFile      InputKind = "file"
)

// InputRoot identifies one command-line input. IDs, rather than resolved paths,
// keep overlapping or repeated inputs as separate folder subtrees.
type InputRoot struct {
	ID    int
	Given string
	Abs   string
	Kind  InputKind
}

// FileRecord is the complete result for one counted file.
// RelPath is relative to the InputRoot and is used for folder aggregation.
type FileRecord struct {
	InputID  int
	Path     string
	RelPath  string
	Language string
	Metrics  Metrics
}

// LanguageRow is a cumulative per-language record.
type LanguageRow struct {
	Language string
	Metrics  Metrics
}

// FolderRow is a cumulative folder record in pre-order traversal order.
// Name and Depth drive the indented tabular view; Path is the flat path used by
// JSON and CSV. Synthetic marks the special (root files) row.
type FolderRow struct {
	InputID   int
	Path      string
	Name      string
	Depth     int
	Synthetic bool
	Metrics   Metrics
}

// CalibrationOverride is a language-specific factor that replaces the
// tokenizer's default calibration factor for that language.
type CalibrationOverride struct {
	Language string
	Factor   float64
}

// SkippedEntry describes one filesystem entry omitted after a recoverable
// scan error. Reports containing skipped entries are incomplete.
type SkippedEntry struct {
	Stage string
	Path  string
	Error string
}

// Metadata describes the tool and tokenizer responsible for a report.
type Metadata struct {
	Version              string
	Tokenizer            string
	CalibrationFactor    float64
	CalibrationOverrides []CalibrationOverride
	Estimated            bool
	Complete             bool
	Skipped              []SkippedEntry
}

// View selects the primary records shown by a renderer.
type View uint8

const (
	ViewLanguage View = iota
	ViewFile
	ViewFolder
)

// Valid reports whether v is a supported report view.
func (v View) Valid() bool {
	return v <= ViewFolder
}

// String returns the command-facing name of a report view.
func (v View) String() string {
	switch v {
	case ViewLanguage:
		return "language"
	case ViewFile:
		return "file"
	case ViewFolder:
		return "folder"
	default:
		return "unknown"
	}
}

// SortKey selects the primary ordering for report rows.
type SortKey string

const (
	SortTokens SortKey = "tokens"
	SortCode   SortKey = "code"
	SortLines  SortKey = "lines"
	SortFiles  SortKey = "files"
	SortName   SortKey = "name"
)

// Valid reports whether s is a supported sort key.
func (s SortKey) Valid() bool {
	switch s {
	case SortTokens, SortCode, SortLines, SortFiles, SortName:
		return true
	default:
		return false
	}
}

// Report is the renderer-ready result. Languages is always populated. Files
// and Folders are populated only for their selected views.
type Report struct {
	Languages []LanguageRow
	Files     []FileRecord
	Folders   []FolderRow
	Totals    Metrics
	Metadata  Metadata
}
