package lsp

import (
	"testing"

	"github.com/toba/epub-lsp/internal/epub"
)

func TestHandleFormatting_XML(t *testing.T) {
	ws := newMockWorkspace()
	opfContent := []byte(
		`<?xml version="1.0"?><package><metadata><dc:title>Test</dc:title></metadata></package>`,
	)
	ws.files["file:///book/content.opf"] = opfContent
	ws.fileTypes["file:///book/content.opf"] = epub.FileTypeOPF

	data := makeRequest(t, 1, MethodFormatting, DocumentFormattingParams{
		TextDocument: TextDocumentIdentifier{Uri: "file:///book/content.opf"},
		Options: FormattingOptions{
			TabSize:      2,
			InsertSpaces: true,
		},
	})

	resp := HandleFormatting(data, ws)
	edits := unmarshalResult[[]TextEdit](t, resp)

	if len(edits) != 1 {
		t.Fatalf("expected 1 edit (whole-document replace), got %d", len(edits))
	}

	if edits[0].NewText == string(opfContent) {
		t.Error("formatted text should differ from original")
	}
}

func TestHandleFormatting_CSS(t *testing.T) {
	ws := newMockWorkspace()
	cssContent := []byte(`body{color:red;font-size:12px;}`)
	ws.files["file:///book/style.css"] = cssContent
	ws.fileTypes["file:///book/style.css"] = epub.FileTypeCSS

	data := makeRequest(t, 1, MethodFormatting, DocumentFormattingParams{
		TextDocument: TextDocumentIdentifier{Uri: "file:///book/style.css"},
		Options: FormattingOptions{
			TabSize:      2,
			InsertSpaces: true,
		},
	})

	resp := HandleFormatting(data, ws)
	edits := unmarshalResult[[]TextEdit](t, resp)

	if len(edits) != 1 {
		t.Fatalf("expected 1 edit, got %d", len(edits))
	}
}

func TestHandleFormatting_NoContent(t *testing.T) {
	ws := newMockWorkspace()
	data := makeRequest(t, 1, MethodFormatting, DocumentFormattingParams{
		TextDocument: TextDocumentIdentifier{Uri: "file:///nonexistent.css"},
		Options: FormattingOptions{
			TabSize:      2,
			InsertSpaces: true,
		},
	})

	resp := HandleFormatting(data, ws)
	edits := unmarshalResult[[]TextEdit](t, resp)

	if len(edits) != 0 {
		t.Fatalf("expected 0 edits, got %d", len(edits))
	}
}

func TestHandleFormatting_TabIndent(t *testing.T) {
	ws := newMockWorkspace()
	opfContent := []byte(`<?xml version="1.0"?><package><metadata></metadata></package>`)
	ws.files["file:///book/content.opf"] = opfContent
	ws.fileTypes["file:///book/content.opf"] = epub.FileTypeOPF

	data := makeRequest(t, 1, MethodFormatting, DocumentFormattingParams{
		TextDocument: TextDocumentIdentifier{Uri: "file:///book/content.opf"},
		Options: FormattingOptions{
			InsertSpaces: false,
		},
	})

	resp := HandleFormatting(data, ws)
	edits := unmarshalResult[[]TextEdit](t, resp)

	if len(edits) != 1 {
		t.Fatalf("expected 1 edit, got %d", len(edits))
	}
}
