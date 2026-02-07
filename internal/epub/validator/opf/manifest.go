package opf

import (
	"github.com/toba/epub-lsp/internal/epub"
	"github.com/toba/epub-lsp/internal/epub/parser"
)

func validateManifest(content []byte, pkg *parser.XMLNode) []epub.Diagnostic {
	var diags []epub.Diagnostic

	manifest := pkg.FindFirst("manifest")
	if manifest == nil {
		return diags
	}

	seenIDs := make(map[string]*parser.XMLNode)

	for _, item := range manifest.Children {
		if item.Local != "item" {
			continue
		}

		pos := epub.ByteOffsetToPosition(content, int(item.Offset))

		// Check for missing media-type
		mediaType := item.Attr("media-type")
		if mediaType == "" {
			diags = append(diags, epub.Diagnostic{
				Code:     "OPF_025",
				Severity: epub.SeverityWarning,
				Message:  "manifest item missing media-type attribute",
				Source:   source,
				Range:    epub.Range{Start: pos, End: pos},
			})
		}

		// Check for empty href
		href := item.Attr("href")
		if href == "" {
			diags = append(diags, epub.Diagnostic{
				Severity: epub.SeverityWarning,
				Message:  "manifest item href is empty",
				Source:   source,
				Range:    epub.Range{Start: pos, End: pos},
			})
		}

		// Check for duplicate IDs
		id := item.Attr("id")
		if id != "" {
			if _, exists := seenIDs[id]; exists {
				diags = append(diags, epub.Diagnostic{
					Severity: epub.SeverityWarning,
					Message:  "duplicate manifest item id: \"" + id + "\"",
					Source:   source,
					Range:    epub.Range{Start: pos, End: pos},
				})
			}
			seenIDs[id] = item
		}
	}

	return diags
}
