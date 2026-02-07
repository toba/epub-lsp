package epub

// DiagBuilder provides a fluent API for constructing diagnostics.
type DiagBuilder struct {
	diag Diagnostic
}

// NewDiag creates a DiagBuilder with position computed from content and offset.
func NewDiag(content []byte, offset int, source string) *DiagBuilder {
	pos := ByteOffsetToPosition(content, offset)
	return &DiagBuilder{
		diag: Diagnostic{
			Source: source,
			Range:  Range{Start: pos, End: pos},
		},
	}
}

// Code sets the diagnostic code.
func (b *DiagBuilder) Code(code string) *DiagBuilder {
	b.diag.Code = code
	return b
}

// Error sets the message and severity to Error.
func (b *DiagBuilder) Error(msg string) *DiagBuilder {
	b.diag.Message = msg
	b.diag.Severity = SeverityError
	return b
}

// Warning sets the message and severity to Warning.
func (b *DiagBuilder) Warning(msg string) *DiagBuilder {
	b.diag.Message = msg
	b.diag.Severity = SeverityWarning
	return b
}

// Info sets the message and severity to Info.
func (b *DiagBuilder) Info(msg string) *DiagBuilder {
	b.diag.Message = msg
	b.diag.Severity = SeverityInfo
	return b
}

// Build returns the constructed Diagnostic.
func (b *DiagBuilder) Build() Diagnostic {
	return b.diag
}
