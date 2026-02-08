package lsp

import (
	"testing"

	"github.com/toba/epub-lsp/internal/epub"
)

func TestHandleCodeAction_MissingAccessMode(t *testing.T) {
	ws := newMockWorkspace()
	opfContent := []byte(`<?xml version="1.0"?>
<package xmlns="http://www.idpf.org/2007/opf">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
    <dc:title>Test</dc:title>
  </metadata>
</package>`)
	ws.files["file:///book/content.opf"] = opfContent
	ws.fileTypes["file:///book/content.opf"] = epub.FileTypeOPF

	data := makeRequest(t, 1, MethodCodeAction, CodeActionParams{
		TextDocument: TextDocumentIdentifier{Uri: "file:///book/content.opf"},
		Range:        Range{},
		Context: CodeActionContext{
			Diagnostics: []Diagnostic{
				{
					Code:    "metadata-accessmode",
					Message: "missing schema:accessMode metadata",
					Range:   Range{},
				},
			},
		},
	})

	resp := HandleCodeAction(data, ws)
	actions := unmarshalResult[[]CodeAction](t, resp)

	if len(actions) != 1 {
		t.Fatalf("expected 1 code action, got %d", len(actions))
	}

	if actions[0].Kind != "quickfix" {
		t.Errorf("expected quickfix kind, got %q", actions[0].Kind)
	}

	if actions[0].Edit == nil {
		t.Fatal("expected edit to be non-nil")
	}
}

func TestHandleCodeAction_NoDiagnostics(t *testing.T) {
	ws := newMockWorkspace()
	opfContent := []byte(`<?xml version="1.0"?><package><metadata></metadata></package>`)
	ws.files["file:///book/content.opf"] = opfContent
	ws.fileTypes["file:///book/content.opf"] = epub.FileTypeOPF

	data := makeRequest(t, 1, MethodCodeAction, CodeActionParams{
		TextDocument: TextDocumentIdentifier{Uri: "file:///book/content.opf"},
		Range:        Range{},
		Context: CodeActionContext{
			Diagnostics: []Diagnostic{},
		},
	})

	resp := HandleCodeAction(data, ws)
	actions := unmarshalResult[[]CodeAction](t, resp)

	if len(actions) != 0 {
		t.Fatalf("expected 0 code actions, got %d", len(actions))
	}
}

func TestHandleCodeAction_UnknownDiagCode(t *testing.T) {
	ws := newMockWorkspace()
	opfContent := []byte(`<?xml version="1.0"?><package><metadata></metadata></package>`)
	ws.files["file:///book/content.opf"] = opfContent
	ws.fileTypes["file:///book/content.opf"] = epub.FileTypeOPF

	data := makeRequest(t, 1, MethodCodeAction, CodeActionParams{
		TextDocument: TextDocumentIdentifier{Uri: "file:///book/content.opf"},
		Range:        Range{},
		Context: CodeActionContext{
			Diagnostics: []Diagnostic{
				{Code: "UNKNOWN_123", Message: "Unknown diagnostic"},
			},
		},
	})

	resp := HandleCodeAction(data, ws)
	actions := unmarshalResult[[]CodeAction](t, resp)

	if len(actions) != 0 {
		t.Fatalf("expected 0 code actions for unknown diag code, got %d", len(actions))
	}
}

func TestHandleCodeAction_FixAll(t *testing.T) {
	ws := newMockWorkspace()
	opfContent := []byte(`<?xml version="1.0"?>
<package xmlns="http://www.idpf.org/2007/opf">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
    <dc:title>Test</dc:title>
  </metadata>
</package>`)
	uri := "file:///book/content.opf"
	ws.files[uri] = opfContent
	ws.fileTypes[uri] = epub.FileTypeOPF
	ws.diagnostics[uri] = []epub.Diagnostic{
		{
			Code:     "metadata-accessmode",
			Severity: epub.SeverityWarning,
			Message:  "missing schema:accessMode metadata",
		},
		{
			Code:     "metadata-accessibilityhazard",
			Severity: epub.SeverityWarning,
			Message:  "missing schema:accessibilityHazard metadata",
		},
	}

	data := makeRequest(t, 1, MethodCodeAction, CodeActionParams{
		TextDocument: TextDocumentIdentifier{Uri: uri},
		Range:        Range{},
		Context: CodeActionContext{
			Only: []string{"source.fixAll"},
		},
	})

	resp := HandleCodeAction(data, ws)
	actions := unmarshalResult[[]CodeAction](t, resp)

	if len(actions) != 1 {
		t.Fatalf("expected 1 source.fixAll action, got %d", len(actions))
	}

	if actions[0].Kind != "source.fixAll" {
		t.Errorf("expected source.fixAll kind, got %q", actions[0].Kind)
	}

	if actions[0].Edit == nil {
		t.Fatal("expected edit to be non-nil")
	}

	edits := actions[0].Edit.Changes[uri]
	if len(edits) != 2 {
		t.Fatalf("expected 2 text edits, got %d", len(edits))
	}

	if len(actions[0].Diagnostics) != 2 {
		t.Fatalf("expected 2 diagnostics in action, got %d", len(actions[0].Diagnostics))
	}
}

func TestHandleCodeAction_FixAllNoDiagnostics(t *testing.T) {
	ws := newMockWorkspace()
	opfContent := []byte(`<?xml version="1.0"?><package><metadata></metadata></package>`)
	uri := "file:///book/content.opf"
	ws.files[uri] = opfContent
	ws.fileTypes[uri] = epub.FileTypeOPF

	data := makeRequest(t, 1, MethodCodeAction, CodeActionParams{
		TextDocument: TextDocumentIdentifier{Uri: uri},
		Range:        Range{},
		Context: CodeActionContext{
			Only: []string{"source.fixAll"},
		},
	})

	resp := HandleCodeAction(data, ws)
	actions := unmarshalResult[[]CodeAction](t, resp)

	if actions != nil {
		t.Fatalf(
			"expected nil actions for fixAll with no diagnostics, got %d",
			len(actions),
		)
	}
}
