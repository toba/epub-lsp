package parser

import (
	"unicode/utf8"

	"github.com/toba/epub-lsp/internal/epub"
)

// CSSToken represents a token from a CSS file.
type CSSToken struct {
	Type   CSSTokenType
	Value  string
	Offset int
	Line   int
	Col    int
}

// CSSTokenType identifies the kind of CSS token.
type CSSTokenType int

const (
	CSSTokenProperty CSSTokenType = iota
	CSSTokenValue
	CSSTokenAtRule
	CSSTokenBraceOpen
	CSSTokenBraceClose
	CSSTokenColon
	CSSTokenSemicolon
	CSSTokenComment
	CSSTokenError
	CSSTokenEOF
)

// CSSTokenizer scans CSS content into tokens.
type CSSTokenizer struct {
	content []byte
	pos     int
	line    int
	col     int
	pending *CSSToken
}

// NewCSSTokenizer creates a tokenizer for the given CSS content.
func NewCSSTokenizer(content []byte) *CSSTokenizer {
	return &CSSTokenizer{content: content}
}

// Unread pushes a token back so it will be returned by the next call to Next.
func (t *CSSTokenizer) Unread(tok CSSToken) {
	t.pending = &tok
}

// Next returns the next token from the CSS content.
func (t *CSSTokenizer) Next() CSSToken {
	if t.pending != nil {
		tok := *t.pending
		t.pending = nil
		return tok
	}

	t.skipWhitespace()

	if t.pos >= len(t.content) {
		return CSSToken{Type: CSSTokenEOF, Offset: t.pos, Line: t.line, Col: t.col}
	}

	ch := t.content[t.pos]

	switch {
	case ch == '/' && t.pos+1 < len(t.content) && t.content[t.pos+1] == '*':
		return t.scanComment()
	case ch == '@':
		return t.scanAtRule()
	case ch == '{':
		tok := CSSToken{
			Type:   CSSTokenBraceOpen,
			Value:  "{",
			Offset: t.pos,
			Line:   t.line,
			Col:    t.col,
		}
		t.advance()
		return tok
	case ch == '}':
		tok := CSSToken{
			Type:   CSSTokenBraceClose,
			Value:  "}",
			Offset: t.pos,
			Line:   t.line,
			Col:    t.col,
		}
		t.advance()
		return tok
	case ch == ':':
		tok := CSSToken{
			Type:   CSSTokenColon,
			Value:  ":",
			Offset: t.pos,
			Line:   t.line,
			Col:    t.col,
		}
		t.advance()
		return tok
	case ch == ';':
		tok := CSSToken{
			Type:   CSSTokenSemicolon,
			Value:  ";",
			Offset: t.pos,
			Line:   t.line,
			Col:    t.col,
		}
		t.advance()
		return tok
	default:
		return t.scanIdent()
	}
}

func (t *CSSTokenizer) advance() {
	if t.pos < len(t.content) {
		if t.content[t.pos] == '\n' {
			t.line++
			t.col = 0
		} else {
			t.col++
		}
		t.pos++
	}
}

func (t *CSSTokenizer) skipWhitespace() {
	for t.pos < len(t.content) {
		ch := t.content[t.pos]
		if ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' {
			t.advance()
		} else {
			break
		}
	}
}

func (t *CSSTokenizer) scanComment() CSSToken {
	start := t.pos
	startLine := t.line
	startCol := t.col
	t.advance() // /
	t.advance() // *

	for t.pos < len(t.content) {
		if t.content[t.pos] == '*' && t.pos+1 < len(t.content) &&
			t.content[t.pos+1] == '/' {
			t.advance() // *
			t.advance() // /
			return CSSToken{
				Type:   CSSTokenComment,
				Value:  string(t.content[start:t.pos]),
				Offset: start,
				Line:   startLine,
				Col:    startCol,
			}
		}
		t.advance()
	}

	// Unclosed comment
	return CSSToken{
		Type:   CSSTokenError,
		Value:  "unclosed comment",
		Offset: start,
		Line:   startLine,
		Col:    startCol,
	}
}

func (t *CSSTokenizer) scanAtRule() CSSToken {
	start := t.pos
	startLine := t.line
	startCol := t.col
	t.advance() // @

	for t.pos < len(t.content) {
		ch := t.content[t.pos]
		if ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' || ch == '{' || ch == ';' {
			break
		}
		t.advance()
	}

	return CSSToken{
		Type:   CSSTokenAtRule,
		Value:  string(t.content[start:t.pos]),
		Offset: start,
		Line:   startLine,
		Col:    startCol,
	}
}

func (t *CSSTokenizer) scanIdent() CSSToken {
	start := t.pos
	startLine := t.line
	startCol := t.col

	for t.pos < len(t.content) {
		ch := t.content[t.pos]
		if ch == '{' || ch == '}' || ch == ':' || ch == ';' || ch == '/' {
			break
		}
		if ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' {
			break
		}
		t.advance()
	}

	value := string(t.content[start:t.pos])
	return CSSToken{
		Type:   CSSTokenProperty,
		Value:  value,
		Offset: start,
		Line:   startLine,
		Col:    startCol,
	}
}

// CSSPropertyDecl represents a CSS property declaration found by scanning.
type CSSPropertyDecl struct {
	Property string
	Value    string
	Offset   int
	Line     int
	Col      int
}

// CSSAtRule represents an @-rule found by scanning.
type CSSAtRule struct {
	Name   string
	Offset int
	Line   int
	Col    int
}

// ScanCSS extracts property declarations and @-rules from CSS content.
// It also returns diagnostics for parse errors and encoding issues.
func ScanCSS(content []byte) ([]CSSPropertyDecl, []CSSAtRule, []epub.Diagnostic) {
	var props []CSSPropertyDecl
	var atRules []CSSAtRule
	var diags []epub.Diagnostic

	// Check UTF-8 encoding
	if !utf8.Valid(content) {
		diags = append(diags, epub.Diagnostic{
			Code:     "CSS_003",
			Severity: epub.SeverityError,
			Message:  "CSS file is not valid UTF-8",
			Source:   "epub-css",
			Range:    epub.Range{},
		})
		return props, atRules, diags
	}

	tok := NewCSSTokenizer(content)
	braceDepth := 0

	for {
		t := tok.Next()
		if t.Type == CSSTokenEOF {
			break
		}

		switch t.Type {
		case CSSTokenError:
			pos := epub.Position{Line: t.Line, Character: t.Col}
			diags = append(diags, epub.Diagnostic{
				Code:     "CSS_008",
				Severity: epub.SeverityError,
				Message:  "CSS parse error: " + t.Value,
				Source:   "epub-css",
				Range:    epub.Range{Start: pos, End: pos},
			})

		case CSSTokenAtRule:
			atRules = append(atRules, CSSAtRule{
				Name:   t.Value,
				Offset: t.Offset,
				Line:   t.Line,
				Col:    t.Col,
			})

		case CSSTokenBraceOpen:
			braceDepth++

		case CSSTokenBraceClose:
			braceDepth--
			if braceDepth < 0 {
				pos := epub.Position{Line: t.Line, Character: t.Col}
				diags = append(diags, epub.Diagnostic{
					Code:     "CSS_008",
					Severity: epub.SeverityError,
					Message:  "CSS parse error: unexpected '}'",
					Source:   "epub-css",
					Range:    epub.Range{Start: pos, End: pos},
				})
				braceDepth = 0
			}

		case CSSTokenProperty:
			if braceDepth > 0 {
				propName := t.Value
				propLine := t.Line
				propCol := t.Col
				propOffset := t.Offset

				// Look for colon then value
				next := tok.Next()
				if next.Type == CSSTokenColon {
					// Scan the value up to semicolon or brace close
					valParts := ""
					for {
						vt := tok.Next()
						if vt.Type == CSSTokenSemicolon ||
							vt.Type == CSSTokenBraceClose ||
							vt.Type == CSSTokenEOF {
							if vt.Type == CSSTokenBraceClose {
								braceDepth--
							}
							break
						}
						if valParts != "" {
							valParts += " "
						}
						valParts += vt.Value
					}
					props = append(props, CSSPropertyDecl{
						Property: propName,
						Value:    valParts,
						Offset:   propOffset,
						Line:     propLine,
						Col:      propCol,
					})
				} else {
					// Not a property declaration (likely a selector); push back
					tok.Unread(next)
				}
			}
		}
	}

	if braceDepth > 0 {
		pos := epub.ByteOffsetToPosition(content, len(content))
		diags = append(diags, epub.Diagnostic{
			Code:     "CSS_008",
			Severity: epub.SeverityError,
			Message:  "CSS parse error: unclosed '{'",
			Source:   "epub-css",
			Range:    epub.Range{Start: pos, End: pos},
		})
	}

	return props, atRules, diags
}
