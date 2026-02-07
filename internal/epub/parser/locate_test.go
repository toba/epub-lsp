package parser

import "testing"

func TestLocateAtPosition_OnElement(t *testing.T) {
	content := []byte(`<root><child attr="value">text</child></root>`)
	root, _ := Parse(content)

	// Offset pointing to <child...>
	childOffset := 6
	result := LocateAtPosition(root, content, childOffset)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Node.Local != "child" {
		t.Errorf("expected node 'child', got %q", result.Node.Local)
	}
}

func TestLocateAtPosition_OnAttribute(t *testing.T) {
	content := []byte(`<root><child attr="value">text</child></root>`)
	root, _ := Parse(content)

	// Find the offset of 'attr' within the tag
	// <root><child attr="value">
	// The attr name starts at offset 13
	result := LocateAtPosition(root, content, 13)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Node.Local != "child" {
		t.Errorf("expected node 'child', got %q", result.Node.Local)
	}
	if result.Attr == nil {
		t.Fatal("expected attribute to be located")
	}
	if result.Attr.Local != "attr" {
		t.Errorf("expected attr 'attr', got %q", result.Attr.Local)
	}
}

func TestLocateAtPosition_InAttributeValue(t *testing.T) {
	content := []byte(`<root><child attr="value">text</child></root>`)
	root, _ := Parse(content)

	// Offset inside the attribute value "value"
	// <root><child attr="value">
	// 0123456789012345678901234
	// The value starts at offset 19 (after the opening quote)
	result := LocateAtPosition(root, content, 20)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Attr == nil {
		t.Fatal("expected attribute to be located")
	}
	if !result.InValue {
		t.Error("expected InValue to be true")
	}
}

func TestLocateAtPosition_OutOfRange(t *testing.T) {
	content := []byte(`<root/>`)
	root, _ := Parse(content)

	if result := LocateAtPosition(root, content, -1); result != nil {
		t.Error("expected nil for negative offset")
	}
	if result := LocateAtPosition(root, content, 100); result != nil {
		t.Error("expected nil for offset beyond content")
	}
}

func TestFindStartTagEnd(t *testing.T) {
	content := []byte(`<item id="x" href="test.html">`)
	end := findStartTagEnd(content, 0)
	if end != 29 {
		t.Errorf("expected 29, got %d", end)
	}
}

func TestFindStartTagEnd_SelfClosing(t *testing.T) {
	content := []byte(`<item id="x"/>`)
	end := findStartTagEnd(content, 0)
	if end != 13 {
		t.Errorf("expected 13, got %d", end)
	}
}

func TestNamespacePrefixes(t *testing.T) {
	tagText := `<html xmlns:epub="http://www.idpf.org/2007/ops" xmlns:dc="http://purl.org/dc/elements/1.1/">`
	prefixes := namespacePrefixes(tagText)
	if prefixes["http://www.idpf.org/2007/ops"] != "epub" {
		t.Errorf("expected epub prefix, got %q", prefixes["http://www.idpf.org/2007/ops"])
	}
	if prefixes["http://purl.org/dc/elements/1.1/"] != "dc" {
		t.Errorf(
			"expected dc prefix, got %q",
			prefixes["http://purl.org/dc/elements/1.1/"],
		)
	}
}
