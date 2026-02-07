package parser

import (
	"bytes"
	"strings"
)

// LocateResult describes what XML construct the cursor is on.
type LocateResult struct {
	Node    *XMLNode
	Attr    *XMLAttr // nil if not on an attribute
	InValue bool     // true if inside attribute value
}

// LocateAtPosition walks an XML tree and the raw content to determine
// which node/attribute the cursor (at the given byte offset) is on.
func LocateAtPosition(root *XMLNode, content []byte, offset int) *LocateResult {
	if offset < 0 || offset >= len(content) {
		return nil
	}

	// Find the deepest node whose tag encompasses the offset
	node := findDeepestNode(root, content, offset)
	if node == nil {
		return nil
	}

	// Check if the offset is within the start tag (where attributes live)
	tagStart := int(node.Offset)
	tagEnd := findStartTagEnd(content, tagStart)

	if offset >= tagStart && offset <= tagEnd {
		// Within the start tag — check attributes
		attr, inValue := locateAttribute(content, tagStart, tagEnd, offset, node)
		if attr != nil {
			return &LocateResult{Node: node, Attr: attr, InValue: inValue}
		}
	}

	return &LocateResult{Node: node}
}

// findDeepestNode returns the deepest XMLNode whose span covers offset.
func findDeepestNode(node *XMLNode, content []byte, offset int) *XMLNode {
	for _, child := range node.Children {
		if int(child.Offset) > offset {
			continue
		}
		if deeper := findDeepestNode(child, content, offset); deeper != nil {
			return deeper
		}
		// Check if offset falls within this child's span
		childEnd := findElementEnd(content, int(child.Offset), child.Local)
		if offset >= int(child.Offset) && offset <= childEnd {
			return child
		}
	}

	if node.Local == "#document" {
		return nil
	}

	return nil
}

// findStartTagEnd finds the position of '>' that closes the start tag.
func findStartTagEnd(content []byte, tagStart int) int {
	inString := byte(0)
	for i := tagStart; i < len(content); i++ {
		ch := content[i]
		if inString != 0 {
			if ch == inString {
				inString = 0
			}
			continue
		}
		if ch == '"' || ch == '\'' {
			inString = ch
			continue
		}
		if ch == '>' {
			return i
		}
	}
	return len(content) - 1
}

// findElementEnd finds the approximate end of an element by looking for
// the closing tag. Falls back to end of start tag if not found.
func findElementEnd(content []byte, tagStart int, local string) int {
	// First find end of start tag
	startTagEnd := findStartTagEnd(content, tagStart)

	// Check if self-closing
	if startTagEnd > 0 && content[startTagEnd-1] == '/' {
		return startTagEnd
	}

	// Look for closing tag
	closeTag := []byte("</" + local)
	idx := bytes.Index(content[startTagEnd:], closeTag)
	if idx >= 0 {
		endIdx := startTagEnd + idx + len(closeTag)
		// Find the > of the close tag
		for endIdx < len(content) {
			if content[endIdx] == '>' {
				return endIdx
			}
			endIdx++
		}
	}

	return startTagEnd
}

// locateAttribute checks if the offset is within an attribute of the tag.
func locateAttribute(
	content []byte,
	tagStart, tagEnd, offset int,
	node *XMLNode,
) (*XMLAttr, bool) {
	// Extract the tag text
	if tagEnd >= len(content) {
		tagEnd = len(content) - 1
	}
	tagText := string(content[tagStart : tagEnd+1])

	for i := range node.Attrs {
		attr := &node.Attrs[i]
		// Find attribute in the tag text
		attrName := attr.Local
		if attr.Space != "" {
			// For namespaced attributes, try the prefix form
			// We search raw text for common patterns
			prefixes := namespacePrefixes(tagText)
			if prefix, ok := prefixes[attr.Space]; ok {
				attrName = prefix + ":" + attr.Local
			}
		}

		// Find name="value" or name='value'
		attrPos := findAttributeInTag(tagText, attrName, attr.Value)
		if attrPos < 0 {
			continue
		}

		absStart := tagStart + attrPos
		// Entire attribute span: name="value"
		attrLen := len(attrName) + 1 + 1 + len(attr.Value) + 1 // name="value"
		absEnd := absStart + attrLen

		if offset >= absStart && offset < absEnd {
			// Determine if we're in the value portion
			valueStart := absStart + len(attrName) + 2 // skip name="
			inValue := offset >= valueStart && offset <= valueStart+len(attr.Value)
			return attr, inValue
		}
	}

	return nil, false
}

// findAttributeInTag finds attribute name="value" in the tag text, returning
// the byte offset of the attribute name within tagText.
func findAttributeInTag(tagText, name, value string) int {
	search := tagText
	base := 0
	for {
		idx := strings.Index(search, name)
		if idx < 0 {
			return -1
		}

		// Verify it's followed by = then a quote
		rest := search[idx+len(name):]
		rest = strings.TrimLeft(rest, " \t\n\r")
		if rest != "" && rest[0] == '=' {
			rest = rest[1:]
			rest = strings.TrimLeft(rest, " \t\n\r")
			if rest != "" && (rest[0] == '"' || rest[0] == '\'') {
				q := rest[0]
				endQ := strings.IndexByte(rest[1:], q)
				if endQ >= 0 {
					foundValue := rest[1 : 1+endQ]
					if foundValue == value {
						return base + idx
					}
				}
			}
		}

		// Keep searching past this match
		advance := idx + len(name)
		search = search[advance:]
		base += advance
	}
}

// namespacePrefixes extracts xmlns:prefix="uri" declarations from tag text.
func namespacePrefixes(tagText string) map[string]string {
	result := make(map[string]string)
	rest := tagText
	for {
		idx := strings.Index(rest, "xmlns:")
		if idx < 0 {
			break
		}
		rest = rest[idx+6:]
		eqIdx := strings.IndexByte(rest, '=')
		if eqIdx < 0 {
			break
		}
		prefix := rest[:eqIdx]
		rest = rest[eqIdx+1:]
		rest = strings.TrimLeft(rest, " \t\n\r")
		if rest == "" {
			break
		}
		q := rest[0]
		if q != '"' && q != '\'' {
			continue
		}
		rest = rest[1:]
		endQ := strings.IndexByte(rest, q)
		if endQ < 0 {
			break
		}
		uri := rest[:endQ]
		result[uri] = prefix
		rest = rest[endQ+1:]
	}
	return result
}
