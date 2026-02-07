package lsp

import (
	"testing"

	"github.com/toba/epub-lsp/internal/epub"
)

func TestHandleReferences_ManifestItem(t *testing.T) {
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

	// Position cursor on <item id="ch1"...>
	offset := findSubstring(opfContent, `<item id="ch1"`)
	pos := epub.ByteOffsetToPosition(opfContent, offset+1) // on the 'i' of item
	lspPosition := lspPos(pos)

	data := makeRequest(t, 1, MethodReferences, ReferenceParams{
		TextDocument: TextDocumentIdentifier{Uri: "file:///book/content.opf"},
		Position:     lspPosition,
	})

	resp := HandleReferences(data, ws)
	locations := unmarshalResult[[]Location](t, resp)

	// Should find the <itemref idref="ch1"/> reference
	if len(locations) == 0 {
		t.Fatal("expected at least 1 reference for manifest item")
	}
}

func TestHandleReferences_IDInXHTML(t *testing.T) {
	ws := newMockWorkspace()
	ch1 := []byte(`<?xml version="1.0"?>
<html xmlns="http://www.w3.org/1999/xhtml">
<body>
  <div id="section1">Content</div>
</body>
</html>`)
	ch2 := []byte(`<?xml version="1.0"?>
<html xmlns="http://www.w3.org/1999/xhtml">
<body>
  <a href="chapter1.xhtml#section1">Link</a>
</body>
</html>`)
	ws.files["file:///book/chapter1.xhtml"] = ch1
	ws.files["file:///book/chapter2.xhtml"] = ch2
	ws.fileTypes["file:///book/chapter1.xhtml"] = epub.FileTypeXHTML
	ws.fileTypes["file:///book/chapter2.xhtml"] = epub.FileTypeXHTML

	// Position cursor on id="section1"
	offset := findSubstring(ch1, `id="section1"`)
	pos := epub.ByteOffsetToPosition(ch1, offset+4) // inside the value
	lspPosition := lspPos(pos)

	data := makeRequest(t, 1, MethodReferences, ReferenceParams{
		TextDocument: TextDocumentIdentifier{Uri: "file:///book/chapter1.xhtml"},
		Position:     lspPosition,
	})

	resp := HandleReferences(data, ws)
	locations := unmarshalResult[[]Location](t, resp)

	if len(locations) == 0 {
		t.Fatal("expected at least 1 reference to section1 id")
	}
}

func TestHandleReferences_NoContent(t *testing.T) {
	ws := newMockWorkspace()
	data := makeRequest(t, 1, MethodReferences, ReferenceParams{
		TextDocument: TextDocumentIdentifier{Uri: "file:///nonexistent.opf"},
		Position:     Position{Line: 0, Character: 0},
	})

	resp := HandleReferences(data, ws)
	locations := unmarshalResult[[]Location](t, resp)

	if len(locations) != 0 {
		t.Fatalf("expected 0 locations, got %d", len(locations))
	}
}
