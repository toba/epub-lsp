package lsp

import (
	"encoding/json"
	"log/slog"
	"slices"
	"strings"

	"github.com/toba/epub-lsp/internal/epub"
	"github.com/toba/epub-lsp/internal/epub/parser"
)

// autoFixableCodes lists diagnostic codes that can be batch-fixed via source.fixAll.
var autoFixableCodes = map[string]bool{
	"metadata-accessmode":           true,
	"metadata-accessmodesufficient": true,
	"metadata-accessibilityfeature": true,
	"metadata-accessibilityhazard":  true,
	"metadata-accessibilitysummary": true,
	"HTM_008":                       true,
	"epub-type-has-matching-role":   true,
}

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

	if slices.Contains(req.Params.Context.Only, "source.fixAll") {
		actions := handleFixAll(uri, content, ws)
		return marshalResponse(req.Id, actions)
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

func handleFixAll(uri string, content []byte, ws WorkspaceReader) []CodeAction {
	storedDiags := ws.GetDiagnostics(uri)
	if len(storedDiags) == 0 {
		return nil
	}

	var edits []TextEdit
	var fixedDiags []Diagnostic

	for _, d := range storedDiags {
		if !autoFixableCodes[d.Code] {
			continue
		}
		lspDiag := Diagnostic{
			Range: Range{
				Start: Position{
					Line:      intToUint(d.Range.Start.Line),
					Character: intToUint(d.Range.Start.Character),
				},
				End: Position{
					Line:      intToUint(d.Range.End.Line),
					Character: intToUint(d.Range.End.Character),
				},
			},
			Message:  d.Message,
			Severity: d.Severity,
			Code:     d.Code,
			Source:   d.Source,
		}
		action := codeActionForDiagnostic(uri, content, &lspDiag)
		if action == nil || action.Edit == nil {
			continue
		}
		for _, fileEdits := range action.Edit.Changes {
			edits = append(edits, fileEdits...)
		}
		fixedDiags = append(fixedDiags, lspDiag)
	}

	if len(edits) == 0 {
		return nil
	}

	return []CodeAction{
		{
			Title:       "Fix all auto-fixable issues",
			Kind:        "source.fixAll",
			Diagnostics: fixedDiags,
			Edit: &WorkspaceEdit{
				Changes: map[string][]TextEdit{
					uri: edits,
				},
			},
		},
	}
}

func codeActionForDiagnostic(uri string, content []byte, diag *Diagnostic) *CodeAction {
	switch diag.Code {
	case "metadata-accessmode":
		return insertMetaAction(uri, content, diag,
			"Add schema:accessMode metadata",
			`<meta property="schema:accessMode">textual</meta>`)
	case "metadata-accessmodesufficient":
		return insertMetaAction(uri, content, diag,
			"Add schema:accessModeSufficient metadata",
			`<meta property="schema:accessModeSufficient">textual</meta>`)
	case "metadata-accessibilityfeature":
		return insertMetaAction(uri, content, diag,
			"Add schema:accessibilityFeature metadata",
			`<meta property="schema:accessibilityFeature">structuralNavigation</meta>`)
	case "metadata-accessibilityhazard":
		return insertMetaAction(uri, content, diag,
			"Add schema:accessibilityHazard metadata",
			`<meta property="schema:accessibilityHazard">none</meta>`)
	case "metadata-accessibilitysummary":
		return insertMetaAction(
			uri,
			content,
			diag,
			"Add schema:accessibilitySummary metadata",
			`<meta property="schema:accessibilitySummary">This publication meets WCAG 2.0 Level AA.</meta>`,
		)
	case "HTM_008":
		// Missing alt attribute on img
		return addAttributeAction(uri, content, diag,
			"Add alt attribute",
			"alt", `""`)
	case "epub-type-has-matching-role":
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
