package lsp

import (
	"testing"

	"github.com/toba/epub-lsp/internal/epub"
	"github.com/toba/epub-lsp/internal/epub/validator"
)

func TestHandleCompletion_MetaProperty(t *testing.T) {
	ws := newMockWorkspace()
	opfContent := []byte(`<?xml version="1.0"?>
<package xmlns="http://www.idpf.org/2007/opf">
  <metadata>
    <meta property="schema:"></meta>
  </metadata>
</package>`)
	ws.files["file:///book/content.opf"] = opfContent
	ws.fileTypes["file:///book/content.opf"] = epub.FileTypeOPF

	// Position cursor inside property=""
	offset := findSubstring(opfContent, `property="schema:"`)
	pos := epub.ByteOffsetToPosition(opfContent, offset+11) // inside the value
	lspPosition := lspPos(pos)

	data := makeRequest(t, 1, MethodCompletion, CompletionParams{
		TextDocument: TextDocumentIdentifier{Uri: "file:///book/content.opf"},
		Position:     lspPosition,
	})

	resp := HandleCompletion(data, ws)
	result := unmarshalResult[CompletionList](t, resp)

	if len(result.Items) < 5 {
		t.Fatalf(
			"expected at least 5 schema property completions, got %d",
			len(result.Items),
		)
	}
}

func TestHandleCompletion_ItemrefIdref(t *testing.T) {
	ws := newMockWorkspace()
	opfContent := []byte(`<?xml version="1.0"?>
<package xmlns="http://www.idpf.org/2007/opf">
  <manifest>
    <item id="ch1" href="chapter1.xhtml" media-type="application/xhtml+xml"/>
    <item id="ch2" href="chapter2.xhtml" media-type="application/xhtml+xml"/>
  </manifest>
  <spine>
    <itemref idref=""/>
  </spine>
</package>`)
	ws.files["file:///book/content.opf"] = opfContent
	ws.fileTypes["file:///book/content.opf"] = epub.FileTypeOPF
	ws.manifest = &validator.ManifestInfo{
		Items: []validator.ManifestItem{
			{ID: "ch1", Href: "chapter1.xhtml", MediaType: "application/xhtml+xml"},
			{ID: "ch2", Href: "chapter2.xhtml", MediaType: "application/xhtml+xml"},
		},
	}

	// Position cursor inside idref=""
	offset := findSubstring(opfContent, `idref=""`)
	pos := epub.ByteOffsetToPosition(opfContent, offset+7) // between the quotes
	lspPosition := lspPos(pos)

	data := makeRequest(t, 1, MethodCompletion, CompletionParams{
		TextDocument: TextDocumentIdentifier{Uri: "file:///book/content.opf"},
		Position:     lspPosition,
	})

	resp := HandleCompletion(data, ws)
	result := unmarshalResult[CompletionList](t, resp)

	if len(result.Items) != 2 {
		t.Fatalf("expected 2 idref completions, got %d", len(result.Items))
	}
}

func TestHandleCompletion_MediaType(t *testing.T) {
	ws := newMockWorkspace()
	opfContent := []byte(`<?xml version="1.0"?>
<package xmlns="http://www.idpf.org/2007/opf">
  <manifest>
    <item id="ch1" href="chapter1.xhtml" media-type=""/>
  </manifest>
</package>`)
	ws.files["file:///book/content.opf"] = opfContent
	ws.fileTypes["file:///book/content.opf"] = epub.FileTypeOPF

	// Position cursor inside media-type=""
	offset := findSubstring(opfContent, `media-type=""`)
	pos := epub.ByteOffsetToPosition(opfContent, offset+12) // between the quotes
	lspPosition := lspPos(pos)

	data := makeRequest(t, 1, MethodCompletion, CompletionParams{
		TextDocument: TextDocumentIdentifier{Uri: "file:///book/content.opf"},
		Position:     lspPosition,
	})

	resp := HandleCompletion(data, ws)
	result := unmarshalResult[CompletionList](t, resp)

	if len(result.Items) < 5 {
		t.Fatalf("expected at least 5 media type completions, got %d", len(result.Items))
	}
}

func TestHandleCompletion_NoContent(t *testing.T) {
	ws := newMockWorkspace()
	data := makeRequest(t, 1, MethodCompletion, CompletionParams{
		TextDocument: TextDocumentIdentifier{Uri: "file:///nonexistent.opf"},
		Position:     Position{Line: 0, Character: 0},
	})

	resp := HandleCompletion(data, ws)
	result := unmarshalResult[CompletionList](t, resp)

	if len(result.Items) != 0 {
		t.Fatalf("expected 0 completions, got %d", len(result.Items))
	}
}
