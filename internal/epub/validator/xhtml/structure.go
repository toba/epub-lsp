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
			diags = append(diags, epub.NewDiag(content, int(img.Offset), source).
				Code("HTM_008").Warning("<img> element missing alt attribute").Build())
		}
	}

	return diags
}
