package lsp

import (
	"encoding/json"
	"log/slog"
	"strings"

	"github.com/toba/epub-lsp/internal/epub"
	"github.com/toba/epub-lsp/internal/epub/parser"
)

// HandleCodeAction processes textDocument/codeAction requests.
func HandleCodeAction(data []byte, ws WorkspaceReader) []byte {
	var req RequestMessage[CodeActionParams]
	if err := json.Unmarshal(data, &req); err != nil {
		slog.Error("error unmarshalling codeAction: " + err.Error())
		return marshalResponse(req.Id, []CodeAction{})
	}

	uri := req.Params.TextDocument.Uri
	content := ws.GetContent(uri)
	if content == nil {
		return marshalResponse(req.Id, []CodeAction{})
	}

	var actions []CodeAction

	for i := range req.Params.Context.Diagnostics {
		action := codeActionForDiagnostic(
			uri,
			content,
			&req.Params.Context.Diagnostics[i],
		)
		if action != nil {
			actions = append(actions, *action)
		}
	}

	return marshalResponse(req.Id, actions)
}

func codeActionForDiagnostic(uri string, content []byte, diag *Diagnostic) *CodeAction {
	switch diag.Code {
	case "ACC_001":
		// Missing accessMode metadata
		return insertMetaAction(uri, content, diag,
			"Add schema:accessMode metadata",
			`<meta property="schema:accessMode">textual</meta>`)
	case "ACC_002":
		// Missing accessModeSufficient
		return insertMetaAction(uri, content, diag,
			"Add schema:accessModeSufficient metadata",
			`<meta property="schema:accessModeSufficient">textual</meta>`)
	case "ACC_003":
		// Missing accessibilityFeature
		return insertMetaAction(uri, content, diag,
			"Add schema:accessibilityFeature metadata",
			`<meta property="schema:accessibilityFeature">structuralNavigation</meta>`)
	case "ACC_004":
		// Missing accessibilityHazard
		return insertMetaAction(uri, content, diag,
			"Add schema:accessibilityHazard metadata",
			`<meta property="schema:accessibilityHazard">none</meta>`)
	case "ACC_005":
		// Missing accessibilitySummary
		return insertMetaAction(
			uri,
			content,
			diag,
			"Add schema:accessibilitySummary metadata",
			`<meta property="schema:accessibilitySummary">This publication meets WCAG 2.0 Level AA.</meta>`,
		)
	case "HTM_004":
		// Missing alt attribute on img
		return addAttributeAction(uri, content, diag,
			"Add alt attribute",
			"alt", `""`)
	case "HTM_046":
		// Missing role attribute
		return addRoleAction(uri, content, diag)
	}
	return nil
}

func insertMetaAction(
	uri string,
	content []byte,
	diag *Diagnostic,
	title, metaElement string,
) *CodeAction {
	// Find the closing </metadata> tag to insert before it
	root, xmlDiags := parser.Parse(content)
	if len(xmlDiags) > 0 {
		return nil
	}

	metadata := root.FindFirst("metadata")
	if metadata == nil {
		return nil
	}

	// Insert before closing </metadata>
	// Find the </metadata> position in raw content
	insertOffset := findClosingTagOffset(content, int(metadata.Offset), "metadata")
	if insertOffset < 0 {
		return nil
	}

	insertPos := epub.ByteOffsetToPosition(content, insertOffset)
	lp := lspPos(insertPos)

	// Determine indentation from context
	indent := detectIndent(content, insertOffset)

	return &CodeAction{
		Title:       title,
		Kind:        "quickfix",
		Diagnostics: []Diagnostic{*diag},
		Edit: &WorkspaceEdit{
			Changes: map[string][]TextEdit{
				uri: {
					{
						Range:   Range{Start: lp, End: lp},
						NewText: indent + metaElement + "\n",
					},
				},
			},
		},
	}
}

func addAttributeAction(
	uri string,
	content []byte,
	diag *Diagnostic,
	title, attrName, attrValue string,
) *CodeAction {
	// Find the element at the diagnostic position
	//nolint:gosec // LSP line/character numbers fit in int
	diagPos := epub.Position{
		Line:      int(diag.Range.Start.Line),
		Character: int(diag.Range.Start.Character),
	}
	offset := epub.PositionToByteOffset(content, diagPos)
	if offset < 0 {
		return nil
	}

	// Find the > of the start tag
	insertOffset := -1
	for i := offset; i < len(content); i++ {
		if content[i] == '>' {
			insertOffset = i
			break
		}
		if content[i] == '/' && i+1 < len(content) && content[i+1] == '>' {
			insertOffset = i
			break
		}
	}
	if insertOffset < 0 {
		return nil
	}

	insertPos := epub.ByteOffsetToPosition(content, insertOffset)
	lp := lspPos(insertPos)

	return &CodeAction{
		Title:       title,
		Kind:        "quickfix",
		Diagnostics: []Diagnostic{*diag},
		Edit: &WorkspaceEdit{
			Changes: map[string][]TextEdit{
				uri: {
					{
						Range:   Range{Start: lp, End: lp},
						NewText: " " + attrName + "=" + attrValue,
					},
				},
			},
		},
	}
}

func addRoleAction(uri string, content []byte, diag *Diagnostic) *CodeAction {
	// Try to determine the appropriate role from the diagnostic message
	role := "doc-chapter" // default
	msg := strings.ToLower(diag.Message)
	if strings.Contains(msg, "noteref") {
		role = "doc-noteref"
	} else if strings.Contains(msg, "footnote") {
		role = "doc-footnote"
	} else if strings.Contains(msg, "endnote") {
		role = "doc-endnote"
	} else if strings.Contains(msg, "chapter") {
		role = "doc-chapter"
	}

	return addAttributeAction(uri, content, diag,
		"Add role=\""+role+"\" attribute",
		"role", `"`+role+`"`)
}

// findClosingTagOffset finds the byte offset of </tagName> in content
// starting from the element's start offset.
func findClosingTagOffset(content []byte, startOffset int, tagName string) int {
	closing := "</" + tagName
	for i := startOffset; i < len(content)-len(closing); i++ {
		if string(content[i:i+len(closing)]) == closing {
			return i
		}
	}
	return -1
}

// detectIndent returns the whitespace indentation at the given offset's line.
func detectIndent(content []byte, offset int) string {
	// Walk backward to find the start of the line
	lineStart := offset
	for lineStart > 0 && content[lineStart-1] != '\n' {
		lineStart--
	}

	// Collect leading whitespace
	var indent strings.Builder
	for i := lineStart; i < offset; i++ {
		if content[i] == ' ' || content[i] == '\t' {
			indent.WriteByte(content[i])
		} else {
			break
		}
	}

	return indent.String()
}
