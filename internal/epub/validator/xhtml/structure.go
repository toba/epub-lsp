package xhtml

import (
	"github.com/toba/epub-lsp/internal/epub"
	"github.com/toba/epub-lsp/internal/epub/parser"
)

func validateStructure(content []byte, root *parser.XMLNode) []epub.Diagnostic {
	var diags []epub.Diagnostic

	// Check img elements for alt attribute
	imgs := root.FindAll("img")
	for _, img := range imgs {
		if !img.HasAttr("alt") {
			pos := epub.ByteOffsetToPosition(content, int(img.Offset))
			diags = append(diags, epub.Diagnostic{
				Code:     "HTM_008",
				Severity: epub.SeverityWarning,
				Message:  "<img> element missing alt attribute",
				Source:   source,
				Range:    epub.Range{Start: pos, End: pos},
			})
		}
	}

	return diags
}
