package formatter

import (
	"strings"
	"testing"
)

func TestFormatXML_BasicIndentation(t *testing.T) {
	input := []byte(
		`<?xml version="1.0"?><package><metadata><title>Test</title></metadata></package>`,
	)
	result, err := FormatXML(input, "  ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should contain indentation
	if !strings.Contains(result, "  ") {
		t.Error("expected indented output")
	}

	// Should end with newline
	if !strings.HasSuffix(result, "\n") {
		t.Error("expected trailing newline")
	}
}

func TestFormatXML_PreservesDeclaration(t *testing.T) {
	input := []byte(`<?xml version="1.0" encoding="UTF-8"?><root><child/></root>`)
	result, err := FormatXML(input, "  ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.HasPrefix(result, `<?xml version="1.0" encoding="UTF-8"?>`) {
		t.Error("expected XML declaration to be preserved")
	}
}

func TestFormatXML_TabIndent(t *testing.T) {
	input := []byte(`<root><child>text</child></root>`)
	result, err := FormatXML(input, "\t")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "\t") {
		t.Error("expected tab indentation")
	}
}

func TestFormatXML_InvalidXML(t *testing.T) {
	input := []byte(`<root><unclosed>`)
	_, err := FormatXML(input, "  ")
	if err == nil {
		t.Error("expected error for invalid XML")
	}
}

func TestFormatXML_EmptyDocument(t *testing.T) {
	input := []byte(`<root/>`)
	result, err := FormatXML(input, "  ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "root") {
		t.Error("expected root element in output")
	}
}
