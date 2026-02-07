package lsp

import (
	"encoding/json"
	"log/slog"
	"strings"

	"github.com/toba/epub-lsp/internal/epub"
	"github.com/toba/epub-lsp/internal/epub/parser"
)

// HandleDocumentSymbol processes textDocument/documentSymbol requests.
func HandleDocumentSymbol(data []byte, ws WorkspaceReader) []byte {
	var req RequestMessage[DocumentSymbolParams]
	if err := json.Unmarshal(data, &req); err != nil {
		slog.Error("error unmarshalling documentSymbol: " + err.Error())
		return marshalResponse(req.Id, []DocumentSymbol{})
	}

	uri := req.Params.TextDocument.Uri
	content := ws.GetContent(uri)
	if content == nil {
		return marshalResponse(req.Id, []DocumentSymbol{})
	}

	fileType := ws.GetFileType(uri)
	var symbols []DocumentSymbol

	switch fileType {
	case epub.FileTypeOPF:
		symbols = opfSymbols(content)
	case epub.FileTypeXHTML, epub.FileTypeNav:
		symbols = xhtmlSymbols(content)
	case epub.FileTypeCSS:
		symbols = cssSymbols(content)
	}

	return marshalResponse(req.Id, symbols)
}

func opfSymbols(content []byte) []DocumentSymbol {
	root, diags := parser.Parse(content)
	if len(diags) > 0 {
		return nil
	}

	pkg := root.FindFirst("package")
	if pkg == nil {
		return nil
	}

	var symbols []DocumentSymbol

	// Metadata section
	if metadata := pkg.FindFirst("metadata"); metadata != nil {
		metaSym := nodeSymbol(metadata, "metadata", SymbolKindNamespace, content)
		for _, child := range metadata.Children {
			name := child.Local
			detail := strings.TrimSpace(child.CharData)
			if prop := child.Attr("property"); prop != "" {
				name = prop
			}
			if detail == "" {
				detail = child.Attr("content")
			}
			childSym := nodeSymbol(child, name, SymbolKindProperty, content)
			childSym.Detail = detail
			metaSym.Children = append(metaSym.Children, childSym)
		}
		symbols = append(symbols, metaSym)
	}

	// Manifest section
	if manifest := pkg.FindFirst("manifest"); manifest != nil {
		manSym := nodeSymbol(manifest, "manifest", SymbolKindNamespace, content)
		for _, item := range manifest.Children {
			if item.Local != "item" {
				continue
			}
			id := item.Attr("id")
			href := item.Attr("href")
			name := id
			if name == "" {
				name = href
			}
			childSym := nodeSymbol(item, name, SymbolKindFile, content)
			childSym.Detail = item.Attr("media-type")
			manSym.Children = append(manSym.Children, childSym)
		}
		symbols = append(symbols, manSym)
	}

	// Spine section
	if spine := pkg.FindFirst("spine"); spine != nil {
		spineSym := nodeSymbol(spine, "spine", SymbolKindNamespace, content)
		for _, itemref := range spine.Children {
			if itemref.Local != "itemref" {
				continue
			}
			idref := itemref.Attr("idref")
			childSym := nodeSymbol(itemref, idref, SymbolKindKey, content)
			spineSym.Children = append(spineSym.Children, childSym)
		}
		symbols = append(symbols, spineSym)
	}

	return symbols
}

func xhtmlSymbols(content []byte) []DocumentSymbol {
	root, diags := parser.Parse(content)
	if len(diags) > 0 {
		return nil
	}

	var symbols []DocumentSymbol

	// Extract heading hierarchy
	headings := []string{"h1", "h2", "h3", "h4", "h5", "h6"}
	for _, h := range headings {
		for _, node := range root.FindAll(h) {
			text := strings.TrimSpace(node.CharData)
			if text == "" {
				text = "<" + h + ">"
			}
			sym := nodeSymbol(node, text, SymbolKindString, content)
			sym.Detail = h
			symbols = append(symbols, sym)
		}
	}

	// Nav elements
	for _, nav := range root.FindAll("nav") {
		epubType := nav.AttrNS(epub.NSEpub, "type")
		if epubType == "" {
			epubType = "nav"
		}
		sym := nodeSymbol(nav, epubType, SymbolKindNamespace, content)
		symbols = append(symbols, sym)
	}

	return symbols
}

func cssSymbols(content []byte) []DocumentSymbol {
	props, atRules, _ := parser.ScanCSS(content)
	symbols := make([]DocumentSymbol, 0, len(atRules))

	for _, at := range atRules {
		pos := epub.ByteOffsetToPosition(content, at.Offset)
		lp := lspPos(pos)
		symbols = append(symbols, DocumentSymbol{
			Name:           at.Name,
			Kind:           SymbolKindNamespace,
			Range:          Range{Start: lp, End: lp},
			SelectionRange: Range{Start: lp, End: lp},
		})
	}

	_ = props // Properties are children of selectors; we just show at-rules and top-level
	return symbols
}

func nodeSymbol(
	node *parser.XMLNode,
	name string,
	kind SymbolKind,
	content []byte,
) DocumentSymbol {
	pos := epub.ByteOffsetToPosition(content, int(node.Offset))
	lp := lspPos(pos)
	return DocumentSymbol{
		Name:           name,
		Kind:           kind,
		Range:          Range{Start: lp, End: lp},
		SelectionRange: Range{Start: lp, End: lp},
	}
}
