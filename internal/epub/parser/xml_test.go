package parser

import (
	"testing"
)

func TestParse_ValidXML(t *testing.T) {
	content := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<root xmlns="http://example.com">
  <child attr="value">text</child>
  <empty/>
</root>`)

	root, diags := Parse(content)
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got %d: %v", len(diags), diags)
	}

	if len(root.Children) != 1 {
		t.Fatalf("expected 1 root child, got %d", len(root.Children))
	}

	rootElem := root.Children[0]
	if rootElem.Local != "root" {
		t.Errorf("expected root element 'root', got %q", rootElem.Local)
	}

	if len(rootElem.Children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(rootElem.Children))
	}

	child := rootElem.Children[0]
	if child.Local != "child" {
		t.Errorf("expected child element 'child', got %q", child.Local)
	}
	if child.Attr("attr") != "value" {
		t.Errorf("expected attr='value', got %q", child.Attr("attr"))
	}
	if child.CharData != "text" {
		t.Errorf("expected chardata 'text', got %q", child.CharData)
	}
}

func TestParse_MalformedXML(t *testing.T) {
	content := []byte(`<root><unclosed>`)

	_, diags := Parse(content)
	if len(diags) == 0 {
		t.Error("expected diagnostics for malformed XML, got none")
	}
}

func TestParse_MismatchedTags(t *testing.T) {
	content := []byte(`<root><a></b></root>`)

	_, diags := Parse(content)
	if len(diags) == 0 {
		t.Error("expected diagnostics for mismatched tags, got none")
	}
}

func TestXMLNode_FindAll(t *testing.T) {
	content := []byte(`<root>
  <item id="a"/>
  <group>
    <item id="b"/>
  </group>
  <item id="c"/>
</root>`)

	root, diags := Parse(content)
	if len(diags) != 0 {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	items := root.FindAll("item")
	if len(items) != 3 {
		t.Errorf("expected 3 items, got %d", len(items))
	}
}

func TestXMLNode_FindFirst(t *testing.T) {
	content := []byte(`<root><a><b/></a></root>`)

	root, diags := Parse(content)
	if len(diags) != 0 {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	b := root.FindFirst("b")
	if b == nil {
		t.Error("expected to find element 'b'")
	}

	c := root.FindFirst("c")
	if c != nil {
		t.Error("expected nil for missing element 'c'")
	}
}

func TestXMLNode_FindAllNS(t *testing.T) {
	content := []byte(`<root xmlns:dc="http://purl.org/dc/elements/1.1/">
  <dc:title>Test</dc:title>
  <dc:creator>Author</dc:creator>
</root>`)

	root, diags := Parse(content)
	if len(diags) != 0 {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	dcNS := "http://purl.org/dc/elements/1.1/"
	titles := root.FindAllNS(dcNS, "title")
	if len(titles) != 1 {
		t.Errorf("expected 1 dc:title, got %d", len(titles))
	}
}

func TestXMLNode_HasAttr(t *testing.T) {
	content := []byte(`<root attr="val"/>`)

	root, diags := Parse(content)
	if len(diags) != 0 {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	elem := root.Children[0]
	if !elem.HasAttr("attr") {
		t.Error("expected HasAttr('attr') to be true")
	}
	if elem.HasAttr("missing") {
		t.Error("expected HasAttr('missing') to be false")
	}
}
