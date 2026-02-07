package lsp

import (
	"testing"

	"github.com/toba/epub-lsp/internal/epub"
)

func TestHandleDocumentLink_OPF(t *testing.T) {
	ws := newMockWorkspace()
	opfContent := []byte(`<?xml version="1.0"?>
<package xmlns="http://www.idpf.org/2007/opf">
  <manifest>
    <item id="ch1" href="chapter1.xhtml" media-type="application/xhtml+xml"/>
    <item id="css" href="style.css" media-type="text/css"/>
  </manifest>
</package>`)
	ws.files["file:///book/content.opf"] = opfContent
	ws.fileTypes["file:///book/content.opf"] = epub.FileTypeOPF

	data := makeRequest(t, 1, MethodDocumentLink, DocumentLinkParams{
		TextDocument: TextDocumentIdentifier{Uri: "file:///book/content.opf"},
	})

	resp := HandleDocumentLink(data, ws)
	links := unmarshalResult[[]DocumentLink](t, resp)

	if len(links) < 2 {
		t.Fatalf("expected at least 2 links, got %d", len(links))
	}

	// Verify links have targets
	for _, link := range links {
		if link.Target == "" {
			t.Error("link target should not be empty")
		}
	}
}

func TestHandleDocumentLink_XHTML(t *testing.T) {
	ws := newMockWorkspace()
	xhtmlContent := []byte(`<?xml version="1.0"?>
<html xmlns="http://www.w3.org/1999/xhtml">
<head><link href="style.css" rel="stylesheet"/></head>
<body>
  <a href="chapter2.xhtml">Next</a>
  <img src="image.png" alt="test"/>
</body>
</html>`)
	ws.files["file:///book/chapter1.xhtml"] = xhtmlContent
	ws.fileTypes["file:///book/chapter1.xhtml"] = epub.FileTypeXHTML

	data := makeRequest(t, 1, MethodDocumentLink, DocumentLinkParams{
		TextDocument: TextDocumentIdentifier{Uri: "file:///book/chapter1.xhtml"},
	})

	resp := HandleDocumentLink(data, ws)
	links := unmarshalResult[[]DocumentLink](t, resp)

	if len(links) != 3 {
		t.Fatalf("expected 3 links (link, a, img), got %d", len(links))
	}
}

func TestHandleDocumentLink_NoContent(t *testing.T) {
	ws := newMockWorkspace()

	data := makeRequest(t, 1, MethodDocumentLink, DocumentLinkParams{
		TextDocument: TextDocumentIdentifier{Uri: "file:///nonexistent.opf"},
	})

	resp := HandleDocumentLink(data, ws)
	links := unmarshalResult[[]DocumentLink](t, resp)

	if len(links) != 0 {
		t.Fatalf("expected 0 links for nonexistent file, got %d", len(links))
	}
}

func TestHandleDocumentLink_SkipsRemoteURLs(t *testing.T) {
	ws := newMockWorkspace()
	xhtmlContent := []byte(`<?xml version="1.0"?>
<html xmlns="http://www.w3.org/1999/xhtml">
<body>
  <a href="https://example.com">External</a>
  <a href="local.xhtml">Local</a>
</body>
</html>`)
	ws.files["file:///book/chapter1.xhtml"] = xhtmlContent
	ws.fileTypes["file:///book/chapter1.xhtml"] = epub.FileTypeXHTML

	data := makeRequest(t, 1, MethodDocumentLink, DocumentLinkParams{
		TextDocument: TextDocumentIdentifier{Uri: "file:///book/chapter1.xhtml"},
	})

	resp := HandleDocumentLink(data, ws)
	links := unmarshalResult[[]DocumentLink](t, resp)

	if len(links) != 1 {
		t.Fatalf("expected 1 link (only local), got %d", len(links))
	}
}
