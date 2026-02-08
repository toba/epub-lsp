package lsp

import (
	"testing"
)

func TestFindTemplateBlocks(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		blocks int
	}{
		{"no templates", "<div>hello</div>", 0},
		{"single block", `{{ .Title }}`, 1},
		{"two blocks", `{{ .A }} and {{ .B }}`, 2},
		{"trim markers", `{{- .Title -}}`, 1},
		{"comment", `{{/* a comment */}}`, 1},
		{"empty content", "", 0},
		{"unclosed block", `{{ .Title`, 0},
		{"adjacent blocks", `{{ .A }}{{ .B }}`, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			blocks := findTemplateBlocks([]byte(tt.input))
			if len(blocks) != tt.blocks {
				t.Errorf("expected %d blocks, got %d", tt.blocks, len(blocks))
			}
		})
	}
}

func TestFindTemplateBlockPositions(t *testing.T) {
	input := []byte(`{{ .Title }}`)
	blocks := findTemplateBlocks(input)
	if len(blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(blocks))
	}

	b := blocks[0]
	inner := string(input[b.innerStart:b.innerEnd])
	if inner != " .Title " {
		t.Errorf("expected inner %q, got %q", " .Title ", inner)
	}
	if b.delimStart != 0 {
		t.Errorf("expected delimStart 0, got %d", b.delimStart)
	}
	if b.delimEnd != 12 {
		t.Errorf("expected delimEnd 12, got %d", b.delimEnd)
	}
}

func TestFindTemplateBlockTrimMarkers(t *testing.T) {
	input := []byte(`{{- .Title -}}`)
	blocks := findTemplateBlocks(input)
	if len(blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(blocks))
	}

	b := blocks[0]
	inner := string(input[b.innerStart:b.innerEnd])
	// innerStart skips {{-, innerEnd stops before -}}
	if inner != " .Title " {
		t.Errorf("expected inner %q, got %q", " .Title ", inner)
	}
}

func TestTokenizeKeywords(t *testing.T) {
	input := []byte(`{{ if .Ready }}yes{{ end }}`)
	tokens := tokenizeTemplates(input)

	var keywords []string
	for _, tok := range tokens {
		if tok.tokenType == tokenKeyword {
			keywords = append(
				keywords,
				string(
					input[lineCharToOffset(input, tok.line, tok.startChar):lineCharToOffset(input, tok.line, tok.startChar)+tok.length],
				),
			)
		}
	}

	expected := []string{"if", "end"}
	if len(keywords) != len(expected) {
		t.Fatalf(
			"expected %d keywords, got %d: %v",
			len(expected),
			len(keywords),
			keywords,
		)
	}
	for i, kw := range expected {
		if keywords[i] != kw {
			t.Errorf("keyword %d: expected %q, got %q", i, kw, keywords[i])
		}
	}
}

func TestTokenizeVariables(t *testing.T) {
	input := []byte(`{{ $x := .Field }}`)
	tokens := tokenizeTemplates(input)

	var vars []uint
	for _, tok := range tokens {
		if tok.tokenType == tokenVariable {
			vars = append(vars, tok.tokenType)
		}
	}

	if len(vars) != 1 {
		t.Errorf("expected 1 variable token ($x), got %d", len(vars))
	}
}

func TestTokenizeProperties(t *testing.T) {
	input := []byte(`{{ .Field }}`)
	tokens := tokenizeTemplates(input)

	var props []uint
	for _, tok := range tokens {
		if tok.tokenType == tokenProperty {
			props = append(props, tok.tokenType)
		}
	}

	if len(props) != 1 {
		t.Errorf("expected 1 property token (.Field), got %d", len(props))
	}
}

func TestTokenizeFunctions(t *testing.T) {
	input := []byte(`{{ len .Items }}`)
	tokens := tokenizeTemplates(input)

	var funcs []string
	for _, tok := range tokens {
		if tok.tokenType == tokenFunction {
			funcs = append(
				funcs,
				string(
					input[lineCharToOffset(input, tok.line, tok.startChar):lineCharToOffset(input, tok.line, tok.startChar)+tok.length],
				),
			)
		}
	}

	if len(funcs) != 1 || funcs[0] != "len" {
		t.Errorf("expected [len], got %v", funcs)
	}
}

func TestTokenizeStrings(t *testing.T) {
	input := []byte(`{{ "hello" }}`)
	tokens := tokenizeTemplates(input)

	found := false
	for _, tok := range tokens {
		if tok.tokenType == tokenString {
			found = true
		}
	}
	if !found {
		t.Error("expected string token for \"hello\"")
	}
}

func TestTokenizeNumbers(t *testing.T) {
	input := []byte(`{{ 42 }}`)
	tokens := tokenizeTemplates(input)

	found := false
	for _, tok := range tokens {
		if tok.tokenType == tokenNumber {
			found = true
		}
	}
	if !found {
		t.Error("expected number token for 42")
	}
}

func TestTokenizeOperators(t *testing.T) {
	input := []byte(`{{ $x := .A | len }}`)
	tokens := tokenizeTemplates(input)

	var ops int
	for _, tok := range tokens {
		if tok.tokenType == tokenOperator {
			ops++
		}
	}
	// Expect: {{, :=, |, }}
	if ops != 4 {
		t.Errorf("expected 4 operator tokens ({{, :=, |, }}), got %d", ops)
	}
}

func TestTokenizeComment(t *testing.T) {
	input := []byte(`{{/* this is a comment */}}`)
	tokens := tokenizeTemplates(input)

	if len(tokens) != 1 {
		t.Fatalf("expected 1 token for comment block, got %d", len(tokens))
	}
	if tokens[0].tokenType != tokenComment {
		t.Errorf(
			"expected comment token type %d, got %d",
			tokenComment,
			tokens[0].tokenType,
		)
	}
	if tokens[0].length != 27 {
		t.Errorf("expected length 27, got %d", tokens[0].length)
	}
}

func TestTokenizeMixedContent(t *testing.T) {
	input := []byte(`<div class="title">{{ .Title }}</div>`)
	tokens := tokenizeTemplates(input)

	// Should only have tokens from the template block, not from XML
	for _, tok := range tokens {
		offset := lineCharToOffset(input, tok.line, tok.startChar)
		text := string(input[offset : offset+tok.length])
		// None of the tokens should be XML content
		if text == "<div" || text == "class" || text == "</div>" {
			t.Errorf("unexpected XML token: %q", text)
		}
	}

	// Should have tokens: {{, .Title, }}
	if len(tokens) < 3 {
		t.Errorf("expected at least 3 tokens, got %d", len(tokens))
	}
}

func TestDeltaEncode(t *testing.T) {
	tokens := []semanticToken{
		{line: 0, startChar: 0, length: 2, tokenType: tokenOperator},
		{line: 0, startChar: 3, length: 5, tokenType: tokenKeyword},
		{line: 1, startChar: 2, length: 3, tokenType: tokenVariable},
	}

	data := deltaEncode(tokens)

	expected := []uint{
		0, 0, 2, tokenOperator, 0, // first token: deltaLine=0, deltaChar=0
		0, 3, 5, tokenKeyword, 0, // same line: deltaLine=0, deltaChar=3
		1, 2, 3, tokenVariable, 0, // next line: deltaLine=1, deltaChar=2 (absolute)
	}

	if len(data) != len(expected) {
		t.Fatalf("expected %d values, got %d", len(expected), len(data))
	}
	for i, v := range expected {
		if data[i] != v {
			t.Errorf("data[%d]: expected %d, got %d", i, v, data[i])
		}
	}
}

func TestDeltaEncodeEmpty(t *testing.T) {
	data := deltaEncode(nil)
	if len(data) != 0 {
		t.Errorf("expected empty data, got %v", data)
	}
}

func TestTokenizeMultilineTemplate(t *testing.T) {
	input := []byte("{{ if .A }}\nhello\n{{ end }}")
	tokens := tokenizeTemplates(input)

	// Find the "end" keyword - it should be on line 2
	for _, tok := range tokens {
		if tok.tokenType == tokenKeyword {
			text := extractToken(input, tok)
			if text == "end" && tok.line != 2 {
				t.Errorf("expected 'end' on line 2, got line %d", tok.line)
			}
		}
	}
}

func TestHandleSemanticTokens(t *testing.T) {
	ws := newMockWorkspace()
	uri := "file:///test.xhtml"
	ws.files[uri] = []byte(`<p>{{ .Title }}</p>`)

	data := makeRequest(t, 1, MethodSemanticTokensFull, SemanticTokensParams{
		TextDocument: TextDocumentIdentifier{Uri: uri},
	})

	response := HandleSemanticTokens(data, ws)
	result := unmarshalResult[SemanticTokensResult](t, response)

	if len(result.Data) == 0 {
		t.Error("expected non-empty semantic tokens data")
	}

	// Should have 3 tokens: {{, .Title, }}
	// Each token is 5 values
	if len(result.Data)%5 != 0 {
		t.Errorf("data length %d is not a multiple of 5", len(result.Data))
	}
}

func TestHandleSemanticTokensEmptyFile(t *testing.T) {
	ws := newMockWorkspace()
	uri := "file:///test.xhtml"
	ws.files[uri] = []byte(`<p>no templates here</p>`)

	data := makeRequest(t, 1, MethodSemanticTokensFull, SemanticTokensParams{
		TextDocument: TextDocumentIdentifier{Uri: uri},
	})

	response := HandleSemanticTokens(data, ws)
	result := unmarshalResult[SemanticTokensResult](t, response)

	if len(result.Data) != 0 {
		t.Errorf("expected empty data for file without templates, got %v", result.Data)
	}
}

func TestHandleSemanticTokensMissingFile(t *testing.T) {
	ws := newMockWorkspace()

	data := makeRequest(t, 1, MethodSemanticTokensFull, SemanticTokensParams{
		TextDocument: TextDocumentIdentifier{Uri: "file:///missing.xhtml"},
	})

	response := HandleSemanticTokens(data, ws)
	result := unmarshalResult[SemanticTokensResult](t, response)

	if len(result.Data) != 0 {
		t.Errorf("expected empty data for missing file, got %v", result.Data)
	}
}

func TestTokenizeLoneDot(t *testing.T) {
	input := []byte(`{{ . }}`)
	tokens := tokenizeTemplates(input)

	found := false
	for _, tok := range tokens {
		if tok.tokenType == tokenVariable && tok.length == 1 {
			found = true
		}
	}
	if !found {
		t.Error("expected variable token for lone dot")
	}
}

func TestTokenizePipe(t *testing.T) {
	input := []byte(`{{ .Name | printf "%s" }}`)
	tokens := tokenizeTemplates(input)

	types := make([]uint, 0, len(tokens))
	for _, tok := range tokens {
		types = append(types, tok.tokenType)
	}

	// Should contain: operator({{), property(.Name), operator(|), function(printf), string("%s"), operator(}})
	expectedTypes := []uint{
		tokenOperator,
		tokenProperty,
		tokenOperator,
		tokenFunction,
		tokenString,
		tokenOperator,
	}
	if len(types) != len(expectedTypes) {
		t.Fatalf("expected %d tokens, got %d: %v", len(expectedTypes), len(types), types)
	}
	for i, et := range expectedTypes {
		if types[i] != et {
			t.Errorf("token %d: expected type %d, got %d", i, et, types[i])
		}
	}
}

func TestTokenizeBacktickString(t *testing.T) {
	input := []byte("{{ `raw string` }}")
	tokens := tokenizeTemplates(input)

	found := false
	for _, tok := range tokens {
		if tok.tokenType == tokenString {
			text := extractToken(input, tok)
			if text == "`raw string`" {
				found = true
			}
		}
	}
	if !found {
		t.Error("expected string token for backtick string")
	}
}

// lineCharToOffset converts line/char back to byte offset for verification.
func lineCharToOffset(content []byte, line, char uint) uint {
	l := uint(0)
	for i := range uint(len(content)) {
		if l == line {
			return i + char
		}
		if content[i] == '\n' {
			l++
		}
	}
	return uint(len(content))
}

// extractToken extracts the token text from content for test verification.
func extractToken(content []byte, tok semanticToken) string {
	offset := lineCharToOffset(content, tok.line, tok.startChar)
	end := min(offset+tok.length, uint(len(content)))
	return string(content[offset:end])
}
