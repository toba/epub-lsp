package opf

import (
	"github.com/toba/epub-lsp/internal/epub"
	"github.com/toba/epub-lsp/internal/epub/parser"
)

const dcNS = "http://purl.org/dc/elements/1.1/"

func validateMetadata(content []byte, pkg *parser.XMLNode) []epub.Diagnostic {
	var diags []epub.Diagnostic

	metadata := pkg.FindFirst("metadata")
	if metadata == nil {
		pos := epub.ByteOffsetToPosition(content, int(pkg.Offset))
		diags = append(diags, epub.Diagnostic{
			Code:     "OPF_030",
			Severity: epub.SeverityError,
			Message:  "missing required <metadata> element",
			Source:   source,
			Range:    epub.Range{Start: pos, End: pos},
		})
		return diags
	}

	// Check dc:identifier
	identifiers := metadata.FindAllNS(dcNS, "identifier")
	if len(identifiers) == 0 {
		pos := epub.ByteOffsetToPosition(content, int(metadata.Offset))
		diags = append(diags, epub.Diagnostic{
			Code:     "OPF_030",
			Severity: epub.SeverityError,
			Message:  "missing required <dc:identifier> in metadata",
			Source:   source,
			Range:    epub.Range{Start: pos, End: pos},
		})
	}

	// Check unique-identifier references a valid dc:identifier
	uniqueID := pkg.Attr("unique-identifier")
	if uniqueID != "" && len(identifiers) > 0 {
		found := false
		for _, id := range identifiers {
			if id.Attr("id") == uniqueID {
				found = true
				break
			}
		}
		if !found {
			pos := epub.ByteOffsetToPosition(content, int(pkg.Offset))
			diags = append(diags, epub.Diagnostic{
				Code:     "OPF_031",
				Severity: epub.SeverityError,
				Message:  "unique-identifier \"" + uniqueID + "\" does not match any dc:identifier/@id",
				Source:   source,
				Range:    epub.Range{Start: pos, End: pos},
			})
		}
	}

	// Check dc:title
	titles := metadata.FindAllNS(dcNS, "title")
	if len(titles) == 0 {
		pos := epub.ByteOffsetToPosition(content, int(metadata.Offset))
		diags = append(diags, epub.Diagnostic{
			Code:     "OPF_032",
			Severity: epub.SeverityError,
			Message:  "missing required <dc:title> in metadata",
			Source:   source,
			Range:    epub.Range{Start: pos, End: pos},
		})
	}

	// Check dc:language
	languages := metadata.FindAllNS(dcNS, "language")
	if len(languages) == 0 {
		pos := epub.ByteOffsetToPosition(content, int(metadata.Offset))
		diags = append(diags, epub.Diagnostic{
			Code:     "OPF_034",
			Severity: epub.SeverityError,
			Message:  "missing required <dc:language> in metadata",
			Source:   source,
			Range:    epub.Range{Start: pos, End: pos},
		})
	}

	return diags
}
