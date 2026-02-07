package opf

import (
	"github.com/toba/epub-lsp/internal/epub"
	"github.com/toba/epub-lsp/internal/epub/parser"
)

func validateSpine(content []byte, pkg *parser.XMLNode) []epub.Diagnostic {
	var diags []epub.Diagnostic

	spine := pkg.FindFirst("spine")
	if spine == nil {
		diags = append(diags, epub.NewDiag(content, int(pkg.Offset), source).
			Code("OPF_019").Error("missing required <spine> element").Build())
		return diags
	}

	// Build a set of manifest item IDs
	manifestIDs := make(map[string]bool)
	manifest := pkg.FindFirst("manifest")
	if manifest != nil {
		for _, item := range manifest.Children {
			if item.Local == "item" {
				if id := item.Attr("id"); id != "" {
					manifestIDs[id] = true
				}
			}
		}
	}

	// Check spine itemrefs reference valid manifest items
	for _, itemref := range spine.Children {
		if itemref.Local != "itemref" {
			continue
		}

		idref := itemref.Attr("idref")
		if idref == "" {
			continue
		}

		if !manifestIDs[idref] {
			diags = append(diags, epub.NewDiag(content, int(itemref.Offset), source).
				Code("OPF_003").
				Error("spine itemref references nonexistent manifest id: \""+idref+"\"").
				Build())
		}
	}

	return diags
}
