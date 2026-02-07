package accessibility

import (
	"github.com/toba/epub-lsp/internal/epub"
	"github.com/toba/epub-lsp/internal/epub/parser"
	"github.com/toba/epub-lsp/internal/epub/validator"
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
	_ *validator.WorkspaceContext,
) []epub.Diagnostic {
	root, xmlDiags := parser.Parse(content)
	if len(xmlDiags) > 0 {
		return nil
	}

	pkg := root.FindFirst("package")
	if pkg == nil {
		return nil
	}

	metadata := pkg.FindFirst("metadata")
	if metadata == nil {
		return nil
	}

	var diags []epub.Diagnostic
	pos := epub.ByteOffsetToPosition(content, int(metadata.Offset))
	rng := epub.Range{Start: pos, End: pos}

	dcNS := "http://purl.org/dc/elements/1.1/"

	// epub-title: must have dc:title
	titles := metadata.FindAllNS(dcNS, "title")
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
	languages := metadata.FindAllNS(dcNS, "language")
	if len(languages) == 0 {
		diags = append(diags, epub.Diagnostic{
			Code:     "epub-lang",
			Severity: epub.SeverityError,
			Message:  "missing language in OPF (dc:language)",
			Source:   source,
			Range:    rng,
		})
	}

	return diags
}
