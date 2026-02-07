package formatter

import (
	"strings"

	"github.com/toba/epub-lsp/internal/epub/parser"
)

// FormatCSS reformats CSS content with consistent indentation.
func FormatCSS(content []byte, indent string) (string, error) {
	tok := parser.NewCSSTokenizer(content)
	var buf strings.Builder
	depth := 0
	needNewline := false

	for {
		t := tok.Next()
		if t.Type == parser.CSSTokenEOF {
			break
		}

		switch t.Type {
		case parser.CSSTokenAtRule:
			if needNewline {
				buf.WriteByte('\n')
			}
			writeIndent(&buf, indent, depth)
			buf.WriteString(t.Value)
			needNewline = false

			// Peek at next token
			next := tok.Next()
			switch next.Type {
			case parser.CSSTokenBraceOpen:
				buf.WriteString(" {\n")
				depth++
				needNewline = false
			case parser.CSSTokenSemicolon:
				buf.WriteString(";\n")
				needNewline = false
			default:
				buf.WriteByte(' ')
				tok.Unread(next)
			}

		case parser.CSSTokenProperty:
			if depth > 0 {
				// Inside a block: property declaration
				if needNewline {
					buf.WriteByte('\n')
				}
				writeIndent(&buf, indent, depth)
				buf.WriteString(t.Value)

				// Expect colon
				next := tok.Next()
				if next.Type == parser.CSSTokenColon {
					buf.WriteString(": ")
					// Collect value tokens until semicolon or close brace
					valueParts := collectValue(tok, &depth)
					buf.WriteString(valueParts)
					needNewline = false
				} else {
					tok.Unread(next)
					needNewline = true
				}
			} else {
				// Outside a block: selector
				if needNewline {
					buf.WriteByte('\n')
					needNewline = false
				}
				writeIndent(&buf, indent, depth)
				buf.WriteString(t.Value)

				// Collect remaining selector parts until brace open
				for {
					next := tok.Next()
					if next.Type == parser.CSSTokenBraceOpen {
						buf.WriteString(" {\n")
						depth++
						needNewline = false
						break
					}
					if next.Type == parser.CSSTokenEOF {
						break
					}
					buf.WriteByte(' ')
					buf.WriteString(next.Value)
				}
			}

		case parser.CSSTokenBraceClose:
			if depth > 0 {
				depth--
			}
			if needNewline {
				buf.WriteByte('\n')
			}
			writeIndent(&buf, indent, depth)
			buf.WriteString("}\n")
			needNewline = false

		case parser.CSSTokenComment:
			if needNewline {
				buf.WriteByte('\n')
			}
			writeIndent(&buf, indent, depth)
			buf.WriteString(t.Value)
			buf.WriteByte('\n')
			needNewline = false
		}
	}

	result := buf.String()
	if !strings.HasSuffix(result, "\n") {
		result += "\n"
	}

	return result, nil
}

func writeIndent(buf *strings.Builder, indent string, depth int) {
	for range depth {
		buf.WriteString(indent)
	}
}

func collectValue(tok *parser.CSSTokenizer, depth *int) string {
	var parts []string
	for {
		t := tok.Next()
		switch t.Type {
		case parser.CSSTokenSemicolon:
			return strings.Join(parts, " ") + ";\n"
		case parser.CSSTokenBraceClose:
			*depth--
			return strings.Join(parts, " ") + ";\n}\n"
		case parser.CSSTokenEOF:
			return strings.Join(parts, " ") + ";\n"
		default:
			parts = append(parts, t.Value)
		}
	}
}
