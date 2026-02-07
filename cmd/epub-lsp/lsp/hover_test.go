package lsp

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/toba/epub-lsp/internal/epub"
	"github.com/toba/epub-lsp/internal/epub/validator"
)

func TestHandleHover_MetaProperty(t *testing.T) {
	ws := newMockWorkspace()
	opfContent := []byte(`<?xml version="1.0"?>
<package xmlns="http://www.idpf.org/2007/opf">
  <metadata>
    <meta property="schema:accessMode">textual</meta>
  </metadata>
</package>`)
	ws.files["file:///book/content.opf"] = opfContent
	ws.fileTypes["file:///book/content.opf"] = epub.FileTypeOPF

	// Position cursor on the <meta> element
	offset := findSubstring(opfContent, `<meta property="schema:accessMode"`)
	pos := epub.ByteOffsetToPosition(opfContent, offset+1) // on 'm' of meta
	lspPosition := lspPos(pos)

	data := makeRequest(t, 1, MethodHover, HoverParams{
		TextDocument: TextDocumentIdentifier{Uri: "file:///book/content.opf"},
		Position:     lspPosition,
	})

	resp := HandleHover(data, ws)

	// Should get hover info about schema:accessMode
	var result ResponseMessage[*Hover]
	if err := unmarshalJSON(resp, &result); err != nil {
		t.Fatal(err)
	}

	if result.Result != nil && result.Result.Contents.Value != "" {
		if !strings.Contains(result.Result.Contents.Value, "accessMode") {
			t.Errorf(
				"expected hover to mention accessMode, got %q",
				result.Result.Contents.Value,
			)
		}
	}
}

func TestHandleHover_ItemrefIdref(t *testing.T) {
	ws := newMockWorkspace()
	opfContent := []byte(`<?xml version="1.0"?>
<package xmlns="http://www.idpf.org/2007/opf">
  <manifest>
    <item id="ch1" href="chapter1.xhtml" media-type="application/xhtml+xml"/>
  </manifest>
  <spine>
    <itemref idref="ch1"/>
  </spine>
</package>`)
	ws.files["file:///book/content.opf"] = opfContent
	ws.fileTypes["file:///book/content.opf"] = epub.FileTypeOPF
	ws.manifest = &validator.ManifestInfo{
		Items: []validator.ManifestItem{
			{ID: "ch1", Href: "chapter1.xhtml", MediaType: "application/xhtml+xml"},
		},
	}

	// Position on "ch1" in idref="ch1"
	offset := findSubstring(opfContent, `idref="ch1"`)
	pos := epub.ByteOffsetToPosition(opfContent, offset+8) // inside "ch1"
	lspPosition := lspPos(pos)

	data := makeRequest(t, 1, MethodHover, HoverParams{
		TextDocument: TextDocumentIdentifier{Uri: "file:///book/content.opf"},
		Position:     lspPosition,
	})

	resp := HandleHover(data, ws)

	var result ResponseMessage[*Hover]
	if err := unmarshalJSON(resp, &result); err != nil {
		t.Fatal(err)
	}

	if result.Result != nil {
		if !strings.Contains(result.Result.Contents.Value, "chapter1.xhtml") {
			t.Errorf(
				"expected hover to mention chapter1.xhtml, got %q",
				result.Result.Contents.Value,
			)
		}
	}
}

func TestHandleHover_NoContent(t *testing.T) {
	ws := newMockWorkspace()
	data := makeRequest(t, 1, MethodHover, HoverParams{
		TextDocument: TextDocumentIdentifier{Uri: "file:///nonexistent.opf"},
		Position:     Position{Line: 0, Character: 0},
	})

	resp := HandleHover(data, ws)

	var result ResponseMessage[*Hover]
	if err := unmarshalJSON(resp, &result); err != nil {
		t.Fatal(err)
	}

	if result.Result != nil {
		t.Fatal("expected nil result for nonexistent file")
	}
}

func unmarshalJSON(data []byte, v any) error {
	return json.Unmarshal(data, v)
}
