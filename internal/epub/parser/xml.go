// Package parser provides XML and CSS parsing helpers for EPUB validation.
package parser

import (
	"bytes"
	"encoding/xml"
	"errors"
	"io"

	"github.com/toba/epub-lsp/internal/epub"
)

// XMLAttr represents an XML attribute with namespace.
type XMLAttr struct {
	Space string
	Local string
	Value string
}

// XMLNode represents a parsed XML element.
type XMLNode struct {
	Space    string
	Local    string
	Attrs    []XMLAttr
	Children []*XMLNode
	CharData string
	Offset   int64
	Line     int
	Col      int
}

// Attr returns the value of the named attribute, or empty string if not found.
func (n *XMLNode) Attr(local string) string {
	for _, a := range n.Attrs {
		if a.Local == local && a.Space == "" {
			return a.Value
		}
	}
	return ""
}

// AttrNS returns the value of the namespaced attribute, or empty string if not found.
func (n *XMLNode) AttrNS(space, local string) string {
	for _, a := range n.Attrs {
		if a.Local == local && a.Space == space {
			return a.Value
		}
	}
	return ""
}

// HasAttr returns true if the element has the named attribute.
func (n *XMLNode) HasAttr(local string) bool {
	for _, a := range n.Attrs {
		if a.Local == local && a.Space == "" {
			return true
		}
	}
	return false
}

// FindAll returns all descendant elements matching the given local name.
func (n *XMLNode) FindAll(local string) []*XMLNode {
	var results []*XMLNode
	for _, child := range n.Children {
		if child.Local == local {
			results = append(results, child)
		}
		results = append(results, child.FindAll(local)...)
	}
	return results
}

// FindAllNS returns all descendant elements matching the given namespace and local name.
func (n *XMLNode) FindAllNS(space, local string) []*XMLNode {
	var results []*XMLNode
	for _, child := range n.Children {
		if child.Local == local && child.Space == space {
			results = append(results, child)
		}
		results = append(results, child.FindAllNS(space, local)...)
	}
	return results
}

// FindFirst returns the first descendant element matching the local name, or nil.
func (n *XMLNode) FindFirst(local string) *XMLNode {
	for _, child := range n.Children {
		if child.Local == local {
			return child
		}
		if found := child.FindFirst(local); found != nil {
			return found
		}
	}
	return nil
}

// FindFirstNS returns the first descendant matching namespace and local name, or nil.
func (n *XMLNode) FindFirstNS(space, local string) *XMLNode {
	for _, child := range n.Children {
		if child.Local == local && child.Space == space {
			return child
		}
		if found := child.FindFirstNS(space, local); found != nil {
			return found
		}
	}
	return nil
}

// Parse parses XML content into a tree of XMLNodes and returns
// any well-formedness errors as diagnostics.
func Parse(content []byte) (*XMLNode, []epub.Diagnostic) {
	decoder := xml.NewDecoder(bytes.NewReader(content))

	root := &XMLNode{Local: "#document"}
	var stack []*XMLNode
	stack = append(stack, root)

	var diags []epub.Diagnostic

	for {
		offset := decoder.InputOffset()
		tok, err := decoder.Token()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			diags = append(diags, epub.NewDiag(content, int(offset), "epub-xml").
				Error("XML well-formedness error: "+err.Error()).Build())
			break
		}

		parent := stack[len(stack)-1]

		switch t := tok.(type) {
		case xml.StartElement:
			pos := epub.ByteOffsetToPosition(content, int(offset))
			node := &XMLNode{
				Space:  t.Name.Space,
				Local:  t.Name.Local,
				Offset: offset,
				Line:   pos.Line,
				Col:    pos.Character,
			}
			for _, attr := range t.Attr {
				node.Attrs = append(node.Attrs, XMLAttr{
					Space: attr.Name.Space,
					Local: attr.Name.Local,
					Value: attr.Value,
				})
			}
			parent.Children = append(parent.Children, node)
			stack = append(stack, node)

		case xml.EndElement:
			if len(stack) > 1 {
				stack = stack[:len(stack)-1]
			}

		case xml.CharData:
			text := string(t)
			if len(stack) > 0 {
				current := stack[len(stack)-1]
				current.CharData += text
			}
		}
	}

	return root, diags
}
