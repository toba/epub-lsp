package opf

import (
	"github.com/toba/epub-lsp/internal/epub"
	"github.com/toba/epub-lsp/internal/epub/parser"
)

func validateSpine(content []byte, pkg *parser.XMLNode) []epub.Diagnostic {
	var diags []epub.Diagnostic

	spine := pkg.FindFirst("spine")
	if spine == nil {
		pos := epub.ByteOffsetToPosition(content, int(pkg.Offset))
		diags = append(diags, epub.Diagnostic{
			Code:     "OPF_019",
			Severity: epub.SeverityError,
			Message:  "missing required <spine> element",
			Source:   source,
			Range:    epub.Range{Start: pos, End: pos},
		})
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
			pos := epub.ByteOffsetToPosition(content, int(itemref.Offset))
			diags = append(diags, epub.Diagnostic{
				Code:     "OPF_003",
				Severity: epub.SeverityError,
				Message:  "spine itemref references nonexistent manifest id: \"" + idref + "\"",
				Source:   source,
				Range:    epub.Range{Start: pos, End: pos},
			})
		}
	}

	return diags
}
