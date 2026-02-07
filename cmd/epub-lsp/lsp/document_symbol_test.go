package lsp

import (
	"testing"

	"github.com/toba/epub-lsp/internal/epub"
)

func TestHandleDocumentSymbol_OPF(t *testing.T) {
	ws := newMockWorkspace()
	opfContent := []byte(`<?xml version="1.0"?>
<package xmlns="http://www.idpf.org/2007/opf">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
    <dc:title>Test Book</dc:title>
    <dc:language>en</dc:language>
  </metadata>
  <manifest>
    <item id="ch1" href="chapter1.xhtml" media-type="application/xhtml+xml"/>
  </manifest>
  <spine>
    <itemref idref="ch1"/>
  </spine>
</package>`)
	ws.files["file:///book/content.opf"] = opfContent
	ws.fileTypes["file:///book/content.opf"] = epub.FileTypeOPF

	data := makeRequest(t, 1, MethodDocumentSymbol, DocumentSymbolParams{
		TextDocument: TextDocumentIdentifier{Uri: "file:///book/content.opf"},
	})

	resp := HandleDocumentSymbol(data, ws)
	symbols := unmarshalResult[[]DocumentSymbol](t, resp)

	// Expect 3 top-level symbols: metadata, manifest, spine
	if len(symbols) != 3 {
		t.Fatalf("expected 3 symbols, got %d", len(symbols))
	}

	if symbols[0].Name != "metadata" {
		t.Errorf("expected first symbol to be metadata, got %q", symbols[0].Name)
	}
	if symbols[1].Name != "manifest" {
		t.Errorf("expected second symbol to be manifest, got %q", symbols[1].Name)
	}
	if symbols[2].Name != "spine" {
		t.Errorf("expected third symbol to be spine, got %q", symbols[2].Name)
	}

	// Metadata should have children
	if len(symbols[0].Children) < 2 {
		t.Errorf(
			"expected metadata to have at least 2 children, got %d",
			len(symbols[0].Children),
		)
	}

	// Manifest should have 1 child
	if len(symbols[1].Children) != 1 {
		t.Errorf("expected manifest to have 1 child, got %d", len(symbols[1].Children))
	}
}

func TestHandleDocumentSymbol_XHTML(t *testing.T) {
	ws := newMockWorkspace()
	xhtmlContent := []byte(`<?xml version="1.0"?>
<html xmlns="http://www.w3.org/1999/xhtml">
<body>
  <h1>Chapter One</h1>
  <h2>Section A</h2>
  <h2>Section B</h2>
</body>
</html>`)
	ws.files["file:///book/chapter1.xhtml"] = xhtmlContent
	ws.fileTypes["file:///book/chapter1.xhtml"] = epub.FileTypeXHTML

	data := makeRequest(t, 1, MethodDocumentSymbol, DocumentSymbolParams{
		TextDocument: TextDocumentIdentifier{Uri: "file:///book/chapter1.xhtml"},
	})

	resp := HandleDocumentSymbol(data, ws)
	symbols := unmarshalResult[[]DocumentSymbol](t, resp)

	if len(symbols) != 3 {
		t.Fatalf("expected 3 heading symbols, got %d", len(symbols))
	}

	if symbols[0].Name != "Chapter One" {
		t.Errorf("expected 'Chapter One', got %q", symbols[0].Name)
	}
}

func TestHandleDocumentSymbol_CSS(t *testing.T) {
	ws := newMockWorkspace()
	cssContent := []byte(`@charset "utf-8";
@font-face {
  font-family: "MyFont";
}
body { color: red; }`)
	ws.files["file:///book/style.css"] = cssContent
	ws.fileTypes["file:///book/style.css"] = epub.FileTypeCSS

	data := makeRequest(t, 1, MethodDocumentSymbol, DocumentSymbolParams{
		TextDocument: TextDocumentIdentifier{Uri: "file:///book/style.css"},
	})

	resp := HandleDocumentSymbol(data, ws)
	symbols := unmarshalResult[[]DocumentSymbol](t, resp)

	if len(symbols) < 2 {
		t.Fatalf(
			"expected at least 2 CSS symbols (@charset, @font-face), got %d",
			len(symbols),
		)
	}
}

func TestHandleDocumentSymbol_NoContent(t *testing.T) {
	ws := newMockWorkspace()
	data := makeRequest(t, 1, MethodDocumentSymbol, DocumentSymbolParams{
		TextDocument: TextDocumentIdentifier{Uri: "file:///nonexistent.opf"},
	})

	resp := HandleDocumentSymbol(data, ws)
	symbols := unmarshalResult[[]DocumentSymbol](t, resp)

	if len(symbols) != 0 {
		t.Fatalf("expected 0 symbols, got %d", len(symbols))
	}
}
