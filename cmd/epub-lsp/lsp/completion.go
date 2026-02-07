package lsp

import (
	"encoding/json"
	"log/slog"

	"github.com/toba/epub-lsp/internal/epub"
	"github.com/toba/epub-lsp/internal/epub/parser"
)

// HandleCompletion processes textDocument/completion requests.
func HandleCompletion(data []byte, ws WorkspaceReader) []byte {
	var req RequestMessage[CompletionParams]
	if err := json.Unmarshal(data, &req); err != nil {
		slog.Error("error unmarshalling completion: " + err.Error())
		return marshalResponse(req.Id, CompletionList{})
	}

	uri := req.Params.TextDocument.Uri
	content := ws.GetContent(uri)
	if content == nil {
		return marshalResponse(req.Id, CompletionList{})
	}

	pos := posToEpub(req.Params.Position)
	offset := epub.PositionToByteOffset(content, pos)
	if offset < 0 {
		return marshalResponse(req.Id, CompletionList{})
	}

	fileType := ws.GetFileType(uri)
	var items []CompletionItem

	root, xmlDiags := parser.Parse(content)
	if len(xmlDiags) > 0 {
		return marshalResponse(req.Id, CompletionList{})
	}

	result := parser.LocateAtPosition(root, content, offset)
	if result == nil {
		return marshalResponse(req.Id, CompletionList{})
	}

	switch fileType {
	case epub.FileTypeOPF:
		items = completionOPF(result, ws)
	case epub.FileTypeXHTML, epub.FileTypeNav:
		items = completionXHTML(result)
	}

	return marshalResponse(req.Id, CompletionList{Items: items})
}

func completionOPF(result *parser.LocateResult, ws WorkspaceReader) []CompletionItem {
	if result.Attr == nil || !result.InValue {
		return nil
	}

	node := result.Node
	attr := result.Attr

	// <meta property="..."> → suggest schema: property names
	if node.Local == "meta" && attr.Local == "property" {
		return schemaPropertyCompletions()
	}

	// <itemref idref="..."> → suggest manifest item IDs
	if node.Local == "itemref" && attr.Local == "idref" {
		return manifestIDCompletions(ws)
	}

	// <item media-type="..."> → suggest media types
	if node.Local == "item" && attr.Local == "media-type" {
		return mediaTypeCompletions()
	}

	return nil
}

func completionXHTML(result *parser.LocateResult) []CompletionItem {
	if result.Attr == nil || !result.InValue {
		return nil
	}

	attr := result.Attr

	// epub:type="..." → suggest valid epub:type values
	if attr.Local == "type" && attr.Space == epub.NSEpub {
		return epubTypeCompletions()
	}

	return nil
}

func schemaPropertyCompletions() []CompletionItem {
	props := []struct {
		name, detail string
	}{
		{"schema:accessMode", "A human sensory perceptual system needed for the content"},
		{
			"schema:accessModeSufficient",
			"Access modes sufficient to understand the content",
		},
		{"schema:accessibilityFeature", "Accessibility features of the resource"},
		{"schema:accessibilityHazard", "Physiologically dangerous characteristics"},
		{"schema:accessibilitySummary", "Human-readable accessibility summary"},
	}

	items := make([]CompletionItem, len(props))
	for i, p := range props {
		items[i] = CompletionItem{
			Label:  p.name,
			Kind:   CompletionKindProperty,
			Detail: p.detail,
		}
	}
	return items
}

func manifestIDCompletions(ws WorkspaceReader) []CompletionItem {
	manifest := ws.GetManifest()
	if manifest == nil {
		return nil
	}

	items := make([]CompletionItem, 0, len(manifest.Items))
	for _, item := range manifest.Items {
		items = append(items, CompletionItem{
			Label:  item.ID,
			Kind:   CompletionKindValue,
			Detail: item.Href + " (" + item.MediaType + ")",
		})
	}
	return items
}

func mediaTypeCompletions() []CompletionItem {
	types := []string{
		"application/xhtml+xml",
		"application/x-dtbncx+xml",
		"text/css",
		"image/jpeg",
		"image/png",
		"image/gif",
		"image/svg+xml",
		"image/webp",
		"application/javascript",
		"audio/mpeg",
		"audio/mp4",
		"video/mp4",
		"application/font-woff",
		"application/font-sfnt",
		"font/otf",
		"font/ttf",
		"font/woff",
		"font/woff2",
	}

	items := make([]CompletionItem, len(types))
	for i, t := range types {
		items[i] = CompletionItem{
			Label: t,
			Kind:  CompletionKindEnum,
		}
	}
	return items
}

func epubTypeCompletions() []CompletionItem {
	types := []struct {
		name, detail string
	}{
		{"toc", "Table of Contents"},
		{"landmarks", "Landmarks navigation"},
		{"page-list", "Page list navigation"},
		{"cover", "Cover image"},
		{"titlepage", "Title page"},
		{"frontmatter", "Front matter"},
		{"bodymatter", "Body matter"},
		{"backmatter", "Back matter"},
		{"chapter", "Chapter"},
		{"part", "Part"},
		{"footnote", "Footnote"},
		{"endnote", "Endnote"},
		{"noteref", "Note reference"},
		{"bibliography", "Bibliography"},
		{"glossary", "Glossary"},
		{"index", "Index"},
		{"preface", "Preface"},
		{"foreword", "Foreword"},
		{"appendix", "Appendix"},
		{"dedication", "Dedication"},
		{"epigraph", "Epigraph"},
		{"abstract", "Abstract"},
		{"colophon", "Colophon"},
		{"pagebreak", "Page break marker"},
	}

	items := make([]CompletionItem, len(types))
	for i, t := range types {
		items[i] = CompletionItem{
			Label:  t.name,
			Kind:   CompletionKindEnum,
			Detail: t.detail,
		}
	}
	return items
}
