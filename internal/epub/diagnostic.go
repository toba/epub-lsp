// Package epub provides core types for EPUB validation.
package epub

// Severity constants match the LSP DiagnosticSeverity values.
const (
	SeverityError   = 1
	SeverityWarning = 2
	SeverityInfo    = 3
	SeverityHint    = 4
)

// Diagnostic represents a validation issue found in an EPUB file.
type Diagnostic struct {
	Code     string `json:"code"`
	Severity int    `json:"severity"`
	Message  string `json:"message"`
	Range    Range  `json:"range"`
	Source   string `json:"source"`
}
