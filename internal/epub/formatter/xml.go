// Package formatter provides XML and CSS formatting for EPUB documents.
package formatter

import (
	"bytes"
	"encoding/xml"
	"errors"
	"io"
	"strings"
)

type xmlTokenKind int

const (
	tokXMLDecl xmlTokenKind = iota
	tokProcInst
	tokComment
	tokDirective
	tokStartTag
	tokEndTag
	tokSelfClosing
	tokCharData
)

type xmlToken struct {
	kind xmlTokenKind
	raw  string
	name string // element name for start/end/self-closing tags
}

// FormatXML reformats XML content with consistent indentation.
// It preserves namespace declarations, self-closing tags, and DOCTYPE formatting.
func FormatXML(content []byte, indent string) (string, error) {
	if err := validateXML(content); err != nil {
		return "", err
	}

	tokens := tokenizeRawXML(content)
	return formatTokens(tokens, indent), nil
}

// validateXML checks if the content is well-formed XML using the standard decoder.
func validateXML(content []byte) error {
	decoder := xml.NewDecoder(bytes.NewReader(content))
	var depth int
	for {
		tok, err := decoder.Token()
		if err != nil {
			if errors.Is(err, io.EOF) {
				if depth != 0 {
					return errors.New("XML error: unclosed elements")
				}
				return nil
			}
			return err
		}
		switch tok.(type) {
		case xml.StartElement:
			depth++
		case xml.EndElement:
			depth--
		}
	}
}

// tokenizeRawXML splits XML content into tokens without interpreting namespaces.
// This preserves the original namespace declarations and self-closing tags exactly.
func tokenizeRawXML(content []byte) []xmlToken {
	var tokens []xmlToken
	i := 0
	for i < len(content) {
		if content[i] != '<' {
			end := bytes.IndexByte(content[i:], '<')
			if end < 0 {
				end = len(content) - i
			}
			tokens = append(
				tokens,
				xmlToken{kind: tokCharData, raw: string(content[i : i+end])},
			)
			i += end
			continue
		}

		remaining := content[i:]
		switch {
		case bytes.HasPrefix(remaining, []byte("<?xml ")):
			end := bytes.Index(remaining, []byte("?>"))
			if end < 0 {
				end = len(remaining)
			} else {
				end += 2
			}
			tokens = append(
				tokens,
				xmlToken{kind: tokXMLDecl, raw: string(remaining[:end])},
			)
			i += end

		case bytes.HasPrefix(remaining, []byte("<?")):
			end := bytes.Index(remaining, []byte("?>"))
			if end < 0 {
				end = len(remaining)
			} else {
				end += 2
			}
			tokens = append(
				tokens,
				xmlToken{kind: tokProcInst, raw: string(remaining[:end])},
			)
			i += end

		case bytes.HasPrefix(remaining, []byte("<!--")):
			end := bytes.Index(remaining, []byte("-->"))
			if end < 0 {
				end = len(remaining)
			} else {
				end += 3
			}
			tokens = append(
				tokens,
				xmlToken{kind: tokComment, raw: string(remaining[:end])},
			)
			i += end

		case bytes.HasPrefix(remaining, []byte("<![CDATA[")):
			end := bytes.Index(remaining, []byte("]]>"))
			if end < 0 {
				end = len(remaining)
			} else {
				end += 3
			}
			tokens = append(
				tokens,
				xmlToken{kind: tokCharData, raw: string(remaining[:end])},
			)
			i += end

		case bytes.HasPrefix(remaining, []byte("<!")):
			end := scanTagEnd(content, i)
			tokens = append(
				tokens,
				xmlToken{kind: tokDirective, raw: string(content[i:end])},
			)
			i = end

		case bytes.HasPrefix(remaining, []byte("</")):
			end := scanTagEnd(content, i)
			raw := string(content[i:end])
			tokens = append(
				tokens,
				xmlToken{kind: tokEndTag, raw: raw, name: extractTagName(raw)},
			)
			i = end

		default:
			end := scanTagEnd(content, i)
			raw := string(content[i:end])
			kind := tokStartTag
			if strings.HasSuffix(strings.TrimRight(raw, " \t\n\r"), "/>") {
				kind = tokSelfClosing
			}
			tokens = append(
				tokens,
				xmlToken{kind: kind, raw: raw, name: extractTagName(raw)},
			)
			i = end
		}
	}
	return tokens
}

// scanTagEnd finds the closing '>' of a tag, correctly skipping quoted attribute values.
func scanTagEnd(content []byte, start int) int {
	for i := start + 1; i < len(content); i++ {
		switch content[i] {
		case '"':
			i++
			for i < len(content) && content[i] != '"' {
				i++
			}
		case '\'':
			i++
			for i < len(content) && content[i] != '\'' {
				i++
			}
		case '>':
			return i + 1
		}
	}
	return len(content)
}

// extractTagName returns the element name from a raw tag string.
func extractTagName(raw string) string {
	s := raw
	if strings.HasPrefix(s, "</") {
		s = s[2:]
	} else {
		s = s[1:]
	}
	for i, c := range s {
		if c == ' ' || c == '\t' || c == '\n' || c == '\r' || c == '>' || c == '/' {
			return s[:i]
		}
	}
	return strings.TrimRight(s, ">/ \t\n\r")
}

// normalizeTag collapses whitespace within a tag while preserving quoted attribute values.
func normalizeTag(raw string) string {
	var buf strings.Builder
	buf.Grow(len(raw))
	var inQuote byte
	prevSpace := false

	for i := range len(raw) {
		c := raw[i]
		if inQuote != 0 {
			buf.WriteByte(c)
			if c == inQuote {
				inQuote = 0
			}
			prevSpace = false
		} else if c == '"' || c == '\'' {
			if prevSpace {
				buf.WriteByte(' ')
				prevSpace = false
			}
			buf.WriteByte(c)
			inQuote = c
		} else if c == ' ' || c == '\t' || c == '\n' || c == '\r' {
			prevSpace = true
		} else {
			if prevSpace && c != '>' {
				buf.WriteByte(' ')
			}
			prevSpace = false
			buf.WriteByte(c)
		}
	}

	return buf.String()
}

// formatTokens renders tokens with proper indentation.
func formatTokens(tokens []xmlToken, indent string) string {
	var buf strings.Builder
	depth := 0

	for i := 0; i < len(tokens); i++ {
		tok := tokens[i]
		switch tok.kind {
		case tokXMLDecl:
			buf.WriteString(strings.TrimSpace(tok.raw))
			buf.WriteByte('\n')

		case tokDirective:
			buf.WriteString(strings.TrimSpace(tok.raw))
			buf.WriteByte('\n')

		case tokComment:
			writeIndent(&buf, indent, depth)
			buf.WriteString(strings.TrimSpace(tok.raw))
			buf.WriteByte('\n')

		case tokProcInst:
			writeIndent(&buf, indent, depth)
			buf.WriteString(strings.TrimSpace(tok.raw))
			buf.WriteByte('\n')

		case tokSelfClosing:
			writeIndent(&buf, indent, depth)
			buf.WriteString(normalizeTag(tok.raw))
			buf.WriteByte('\n')

		case tokStartTag:
			if isInlineElement(tokens, i) {
				writeIndent(&buf, indent, depth)
				buf.WriteString(normalizeTag(tok.raw))
				i++
				if i < len(tokens) && tokens[i].kind == tokCharData {
					buf.WriteString(strings.TrimSpace(tokens[i].raw))
					i++
				}
				if i < len(tokens) {
					buf.WriteString(strings.TrimSpace(tokens[i].raw))
				}
				buf.WriteByte('\n')
			} else {
				writeIndent(&buf, indent, depth)
				buf.WriteString(normalizeTag(tok.raw))
				buf.WriteByte('\n')
				depth++
			}

		case tokEndTag:
			depth--
			if depth < 0 {
				depth = 0
			}
			writeIndent(&buf, indent, depth)
			buf.WriteString(strings.TrimSpace(tok.raw))
			buf.WriteByte('\n')

		case tokCharData:
			text := strings.TrimSpace(tok.raw)
			if text != "" {
				writeIndent(&buf, indent, depth)
				buf.WriteString(text)
				buf.WriteByte('\n')
			}
		}
	}

	result := buf.String()
	if !strings.HasSuffix(result, "\n") {
		result += "\n"
	}
	return result
}

// isInlineElement checks if a start tag at position i contains only text content.
func isInlineElement(tokens []xmlToken, i int) bool {
	if i+1 < len(tokens) && tokens[i+1].kind == tokEndTag &&
		tokens[i+1].name == tokens[i].name {
		return true
	}
	if i+2 < len(tokens) && tokens[i+1].kind == tokCharData &&
		tokens[i+2].kind == tokEndTag &&
		tokens[i+2].name == tokens[i].name {
		return true
	}
	return false
}
