package lsp

import (
	"encoding/json"
	"log/slog"
	"strings"

	"github.com/toba/epub-lsp/internal/epub"
	"github.com/toba/epub-lsp/internal/epub/formatter"
)

// HandleFormatting processes textDocument/formatting requests.
func HandleFormatting(data []byte, ws WorkspaceReader) []byte {
	var req RequestMessage[DocumentFormattingParams]
	if err := json.Unmarshal(data, &req); err != nil {
		slog.Error("error unmarshalling formatting: " + err.Error())
		return marshalResponse(req.Id, []TextEdit{})
	}

	uri := req.Params.TextDocument.Uri
	content := ws.GetContent(uri)
	if content == nil {
		return marshalResponse(req.Id, []TextEdit{})
	}

	fileType := ws.GetFileType(uri)
	indent := "  "
	if !req.Params.Options.InsertSpaces {
		indent = "\t"
	} else if req.Params.Options.TabSize > 0 {
		indent = ""
		var indentSb31 strings.Builder
		for range req.Params.Options.TabSize {
			indentSb31.WriteString(" ")
		}
		indent += indentSb31.String()
	}

	var formatted string
	var err error

	switch fileType {
	case epub.FileTypeOPF, epub.FileTypeXHTML, epub.FileTypeNav:
		formatted, err = formatter.FormatXML(content, indent)
	case epub.FileTypeCSS:
		formatted, err = formatter.FormatCSS(content, indent)
	default:
		return marshalResponse(req.Id, []TextEdit{})
	}

	if err != nil {
		slog.Warn("formatting failed: " + err.Error())
		return marshalResponse(req.Id, []TextEdit{})
	}

	if formatted == string(content) {
		return marshalResponse(req.Id, []TextEdit{})
	}

	// Replace entire document
	endPos := epub.ByteOffsetToPosition(content, len(content))

	edits := []TextEdit{
		{
			Range: Range{
				Start: Position{Line: 0, Character: 0},
				End: Position{
					Line:      intToUint(endPos.Line),
					Character: intToUint(endPos.Character),
				},
			},
			NewText: formatted,
		},
	}

	return marshalResponse(req.Id, edits)
}
