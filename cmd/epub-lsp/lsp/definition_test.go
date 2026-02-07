package lsp

import (
	"testing"

	"github.com/toba/epub-lsp/internal/epub"
)

func TestHandleDefinition_ItemrefToItem(t *testing.T) {
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

	// Position the cursor on "ch1" in idref="ch1"
	// Find the offset of the value "ch1" in the idref attribute
	content := opfContent
	// <itemref idref="ch1"/>
	// We need to find the line/character for "ch1" inside idref value
	offset := findSubstring(content, `idref="ch1"`)
	// Position inside the value
	pos := epub.ByteOffsetToPosition(content, offset+8) // offset of 'c' in ch1
	lspPosition := lspPos(pos)

	data := makeRequest(t, 1, MethodDefinition, DefinitionParams{
		TextDocument: TextDocumentIdentifier{Uri: "file:///book/content.opf"},
		Position:     lspPosition,
	})

	resp := HandleDefinition(data, ws)
	locations := unmarshalResult[[]Location](t, resp)

	if len(locations) == 0 {
		t.Fatal("expected at least 1 location for itemref→item definition")
	}

	if locations[0].URI != "file:///book/content.opf" {
		t.Errorf("expected same URI, got %q", locations[0].URI)
	}
}

func TestHandleDefinition_HrefInXHTML(t *testing.T) {
	ws := newMockWorkspace()
	xhtmlContent := []byte(`<?xml version="1.0"?>
<html xmlns="http://www.w3.org/1999/xhtml">
<body>
  <a href="chapter2.xhtml#section1">Link</a>
</body>
</html>`)
	chapter2 := []byte(`<?xml version="1.0"?>
<html xmlns="http://www.w3.org/1999/xhtml">
<body>
  <div id="section1">Content</div>
</body>
</html>`)
	ws.files["file:///book/chapter1.xhtml"] = xhtmlContent
	ws.files["file:///book/chapter2.xhtml"] = chapter2
	ws.fileTypes["file:///book/chapter1.xhtml"] = epub.FileTypeXHTML
	ws.fileTypes["file:///book/chapter2.xhtml"] = epub.FileTypeXHTML

	// Position cursor on the href value
	offset := findSubstring(xhtmlContent, `href="chapter2.xhtml#section1"`)
	pos := epub.ByteOffsetToPosition(xhtmlContent, offset+6) // inside the value
	lspPosition := lspPos(pos)

	data := makeRequest(t, 1, MethodDefinition, DefinitionParams{
		TextDocument: TextDocumentIdentifier{Uri: "file:///book/chapter1.xhtml"},
		Position:     lspPosition,
	})

	resp := HandleDefinition(data, ws)
	locations := unmarshalResult[[]Location](t, resp)

	if len(locations) == 0 {
		t.Fatal("expected at least 1 location for href definition")
	}

	if locations[0].URI != "file:///book/chapter2.xhtml" {
		t.Errorf("expected chapter2 URI, got %q", locations[0].URI)
	}
}

func TestHandleDefinition_NoContent(t *testing.T) {
	ws := newMockWorkspace()
	data := makeRequest(t, 1, MethodDefinition, DefinitionParams{
		TextDocument: TextDocumentIdentifier{Uri: "file:///nonexistent.opf"},
		Position:     Position{Line: 0, Character: 0},
	})

	resp := HandleDefinition(data, ws)
	locations := unmarshalResult[[]Location](t, resp)

	if len(locations) != 0 {
		t.Fatalf("expected 0 locations, got %d", len(locations))
	}
}

// findSubstring returns the byte offset of the first occurrence of substr in content.
func findSubstring(content []byte, substr string) int {
	s := string(content)
	for i := range len(s) - len(substr) + 1 {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
