package accessibility

import (
	"github.com/toba/epub-lsp/internal/epub"
	"github.com/toba/epub-lsp/internal/epub/validator"
	"github.com/toba/epub-lsp/internal/epub/validator/opf"
)

// OPFAccessibilityValidator checks OPF-level accessibility requirements
// that overlap with Ace rules (epub-title, epub-lang).
type OPFAccessibilityValidator struct{}

func (v *OPFAccessibilityValidator) FileTypes() []epub.FileType {
	return []epub.FileType{epub.FileTypeOPF}
}

func (v *OPFAccessibilityValidator) Validate(
	_ string,
	content []byte,
	ctx *validator.WorkspaceContext,
) []epub.Diagnostic {
	if ctx != nil && ctx.AccessibilitySeverity == 0 {
		return nil
	}

	_, metadata := opf.ParseOPFMetadata(content)
	if metadata == nil {
		return nil
	}

	var diags []epub.Diagnostic
	pos := epub.ByteOffsetToPosition(content, int(metadata.Offset))
	rng := epub.Range{Start: pos, End: pos}

	// epub-title: must have dc:title
	titles := metadata.FindAllNS(epub.NSDC, "title")
	if len(titles) == 0 {
		diags = append(diags, epub.Diagnostic{
			Code:     "epub-title",
			Severity: epub.SeverityError,
			Message:  "missing publication title (dc:title)",
			Source:   source,
			Range:    rng,
		})
	}

	// epub-lang: must have dc:language
	languages := metadata.FindAllNS(epub.NSDC, "language")
	if len(languages) == 0 {
		diags = append(diags, epub.Diagnostic{
			Code:     "epub-lang",
			Severity: epub.SeverityError,
			Message:  "missing language in OPF (dc:language)",
			Source:   source,
			Range:    rng,
		})
	}

	if ctx != nil && ctx.AccessibilitySeverity != 0 {
		for i := range diags {
			diags[i].Severity = ctx.AccessibilitySeverity
		}
	}

	return diags
}
