package lsp

import (
	"encoding/json"
	"log/slog"
	"strings"
	"unicode"
)

// Token type indices matching SemanticTokenTypes legend.
const (
	tokenKeyword  = 0
	tokenVariable = 1
	tokenFunction = 2
	tokenProperty = 3
	tokenString   = 4
	tokenNumber   = 5
	tokenOperator = 6
	tokenComment  = 7
)

// templateKeywords are Go template action keywords.
var templateKeywords = map[string]bool{
	"if": true, "else": true, "end": true, "range": true,
	"with": true, "define": true, "template": true, "block": true,
	"nil": true, "not": true, "and": true, "or": true,
}

// semanticToken represents a single token before delta encoding.
type semanticToken struct {
	line      uint
	startChar uint
	length    uint
	tokenType uint
}

// templateBlock represents a {{ ... }} block found in the source.
type templateBlock struct {
	// innerStart is the byte offset of the first character after the opening delimiter.
	innerStart int
	// innerEnd is the byte offset of the closing delimiter (exclusive of inner content).
	innerEnd int
	// delimStart is the byte offset of the opening {{ (for comment/operator tokens).
	delimStart int
	// delimEnd is the byte offset after the closing }} (for comment/operator tokens).
	delimEnd int
}

// HandleSemanticTokens processes textDocument/semanticTokens/full requests.
func HandleSemanticTokens(data []byte, ws WorkspaceReader) []byte {
	var req RequestMessage[SemanticTokensParams]
	if err := json.Unmarshal(data, &req); err != nil {
		slog.Error("error unmarshalling semantic tokens: " + err.Error())
		return marshalNullResponse(req.Id)
	}

	content := ws.GetContent(req.Params.TextDocument.Uri)
	if content == nil {
		return marshalResponse(req.Id, SemanticTokensResult{Data: []uint{}})
	}

	tokens := tokenizeTemplates(content)
	encoded := deltaEncode(tokens)

	return marshalResponse(req.Id, SemanticTokensResult{Data: encoded})
}

// findTemplateBlocks scans content for {{ ... }} delimiters and returns their positions.
func findTemplateBlocks(content []byte) []templateBlock {
	var blocks []templateBlock
	text := string(content)
	i := 0

	for i < len(text)-1 {
		// Find opening {{
		idx := strings.Index(text[i:], "{{")
		if idx < 0 {
			break
		}
		openPos := i + idx

		// Find closing }}
		searchFrom := openPos + 2

		// Handle comments: {{/* ... */}}
		innerStart := openPos + 2
		if innerStart < len(text) && text[innerStart] == '-' {
			innerStart++
		}

		closeIdx := strings.Index(text[searchFrom:], "}}")
		if closeIdx < 0 {
			break
		}
		closePos := searchFrom + closeIdx

		innerEnd := closePos
		if innerEnd > 0 && text[innerEnd-1] == '-' {
			innerEnd--
		}

		blocks = append(blocks, templateBlock{
			innerStart: innerStart,
			innerEnd:   innerEnd,
			delimStart: openPos,
			delimEnd:   closePos + 2,
		})

		i = closePos + 2
	}

	return blocks
}

// byteOffsetToLineChar converts a byte offset to (line, character) using the content.
func byteOffsetToLineChar(content []byte, offset int) (uint, uint) {
	line := uint(0)
	lineStart := 0

	for i := 0; i < offset && i < len(content); i++ {
		if content[i] == '\n' {
			line++
			lineStart = i + 1
		}
	}

	return line, uint(offset - lineStart) //nolint:gosec // offset >= lineStart
}

// tokenizeTemplates extracts semantic tokens from Go template blocks in content.
func tokenizeTemplates(content []byte) []semanticToken {
	blocks := findTemplateBlocks(content)
	var tokens []semanticToken
	text := string(content)

	for _, block := range blocks {
		inner := text[block.innerStart:block.innerEnd]
		trimmed := strings.TrimSpace(inner)

		// Comment blocks: {{/* ... */}}
		if strings.HasPrefix(trimmed, "/*") {
			line, char := byteOffsetToLineChar(content, block.delimStart)
			// Token spans from {{ to }}
			//nolint:gosec // delimEnd > delimStart
			length := uint(block.delimEnd - block.delimStart)
			tokens = append(tokens, semanticToken{
				line: line, startChar: char, length: length, tokenType: tokenComment,
			})
			continue
		}

		// Opening {{ delimiter as operator
		openLen := max(block.innerStart-block.delimStart,
			// includes trim marker
			2)
		oLine, oChar := byteOffsetToLineChar(content, block.delimStart)
		tokens = append(tokens, semanticToken{
			line: oLine, startChar: oChar, length: uint(openLen), tokenType: tokenOperator, //nolint:gosec // openLen >= 2
		})

		// Tokenize inner content
		tokens = tokenizeInner(content, block.innerStart, inner, tokens)

		// Closing }} delimiter as operator
		closeStart := block.innerEnd
		if text[block.innerEnd] == '-' {
			closeStart = block.innerEnd
		}
		closeLen := block.delimEnd - closeStart
		cLine, cChar := byteOffsetToLineChar(content, closeStart)
		tokens = append(tokens, semanticToken{
			line: cLine, startChar: cChar, length: uint(closeLen), tokenType: tokenOperator, //nolint:gosec // closeLen >= 2
		})
	}

	return tokens
}

// tokenizeInner tokenizes the content inside a template block.
func tokenizeInner(
	content []byte,
	baseOffset int,
	inner string,
	tokens []semanticToken,
) []semanticToken {
	i := 0

	for i < len(inner) {
		ch := inner[i]

		// Skip whitespace
		if ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' {
			i++
			continue
		}

		// String literals
		if ch == '"' || ch == '`' {
			start := i
			quote := ch
			i++
			for i < len(inner) && inner[i] != quote {
				if quote == '"' && inner[i] == '\\' {
					i++ // skip escaped char
				}
				i++
			}
			if i < len(inner) {
				i++ // closing quote
			}
			line, char := byteOffsetToLineChar(content, baseOffset+start)
			tokens = append(tokens, semanticToken{
				line: line, startChar: char, length: uint(i - start), tokenType: tokenString, //nolint:gosec // i >= start
			})
			continue
		}

		// Numbers
		if ch >= '0' && ch <= '9' {
			start := i
			for i < len(inner) && (inner[i] >= '0' && inner[i] <= '9' || inner[i] == '.') {
				i++
			}
			line, char := byteOffsetToLineChar(content, baseOffset+start)
			tokens = append(tokens, semanticToken{
				line: line, startChar: char, length: uint(i - start), tokenType: tokenNumber, //nolint:gosec // i >= start
			})
			continue
		}

		// Operators: |, :=, =, ==, !=, <, >, <=, >=, &&, ||
		if ch == '|' || ch == '=' || ch == '!' || ch == '<' || ch == '>' || ch == '&' {
			start := i
			i++
			if i < len(inner) {
				next := inner[i]
				if (ch == ':' && next == '=') ||
					(ch == '=' && next == '=') ||
					(ch == '!' && next == '=') ||
					(ch == '<' && next == '=') ||
					(ch == '>' && next == '=') ||
					(ch == '&' && next == '&') ||
					(ch == '|' && next == '|') {
					i++
				}
			}
			line, char := byteOffsetToLineChar(content, baseOffset+start)
			tokens = append(tokens, semanticToken{
				line: line, startChar: char, length: uint(i - start), tokenType: tokenOperator, //nolint:gosec // i >= start
			})
			continue
		}

		// := operator
		if ch == ':' && i+1 < len(inner) && inner[i+1] == '=' {
			line, char := byteOffsetToLineChar(content, baseOffset+i)
			tokens = append(tokens, semanticToken{
				line: line, startChar: char, length: 2, tokenType: tokenOperator,
			})
			i += 2
			continue
		}

		// Variables: $ or .Field
		if ch == '$' {
			start := i
			i++
			for i < len(inner) && (isIdentChar(inner[i])) {
				i++
			}
			line, char := byteOffsetToLineChar(content, baseOffset+start)
			tokens = append(tokens, semanticToken{
				line: line, startChar: char, length: uint(i - start), tokenType: tokenVariable, //nolint:gosec // i >= start
			})
			continue
		}

		if ch == '.' {
			// Check if followed by an identifier (field access)
			if i+1 < len(inner) && isIdentStart(inner[i+1]) {
				start := i
				i++
				for i < len(inner) && isIdentChar(inner[i]) {
					i++
				}
				line, char := byteOffsetToLineChar(content, baseOffset+start)
				tokens = append(tokens, semanticToken{
					line: line, startChar: char, length: uint(i - start), tokenType: tokenProperty, //nolint:gosec // i >= start
				})
				continue
			}
			// Lone dot (current context)
			line, char := byteOffsetToLineChar(content, baseOffset+i)
			tokens = append(tokens, semanticToken{
				line: line, startChar: char, length: 1, tokenType: tokenVariable,
			})
			i++
			continue
		}

		// Identifiers: keywords or functions
		if isIdentStart(ch) {
			start := i
			for i < len(inner) && isIdentChar(inner[i]) {
				i++
			}
			word := inner[start:i]

			tokenType := uint(tokenFunction)
			if templateKeywords[word] {
				tokenType = tokenKeyword
			}

			line, char := byteOffsetToLineChar(content, baseOffset+start)
			tokens = append(tokens, semanticToken{
				line: line, startChar: char, length: uint(len(word)), tokenType: tokenType,
			})
			continue
		}

		// Skip parentheses, commas, and other punctuation
		i++
	}

	return tokens
}

// isIdentStart returns true if c can start an identifier.
func isIdentStart(c byte) bool {
	return c == '_' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
		c > 127 && unicode.IsLetter(rune(c))
}

// isIdentChar returns true if c can continue an identifier.
func isIdentChar(c byte) bool {
	return isIdentStart(c) || (c >= '0' && c <= '9')
}

// deltaEncode converts absolute token positions to LSP delta-encoded format.
// Each token is encoded as 5 uints: deltaLine, deltaStartChar, length, tokenType, tokenModifiers.
func deltaEncode(tokens []semanticToken) []uint {
	if len(tokens) == 0 {
		return []uint{}
	}

	data := make([]uint, 0, len(tokens)*5)
	prevLine := uint(0)
	prevChar := uint(0)

	for _, tok := range tokens {
		deltaLine := tok.line - prevLine
		deltaChar := tok.startChar
		if deltaLine == 0 {
			deltaChar = tok.startChar - prevChar
		}

		data = append(data, deltaLine, deltaChar, tok.length, tok.tokenType, 0)

		prevLine = tok.line
		prevChar = tok.startChar
	}

	return data
}
