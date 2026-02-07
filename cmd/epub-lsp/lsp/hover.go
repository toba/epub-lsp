package lsp

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/toba/epub-lsp/internal/epub"
	"github.com/toba/epub-lsp/internal/epub/parser"
)

// HandleHover processes textDocument/hover requests.
func HandleHover(data []byte, ws WorkspaceReader) []byte {
	var req RequestMessage[HoverParams]
	if err := json.Unmarshal(data, &req); err != nil {
		slog.Error("error unmarshalling hover: " + err.Error())
		return marshalNullResponse(req.Id)
	}

	uri := req.Params.TextDocument.Uri
	content := ws.GetContent(uri)
	if content == nil {
		return marshalNullResponse(req.Id)
	}

	pos := posToEpub(req.Params.Position)
	offset := epub.PositionToByteOffset(content, pos)
	if offset < 0 {
		return marshalNullResponse(req.Id)
	}

	root, xmlDiags := parser.Parse(content)
	if len(xmlDiags) > 0 {
		return marshalNullResponse(req.Id)
	}

	result := parser.LocateAtPosition(root, content, offset)
	if result == nil {
		return marshalNullResponse(req.Id)
	}

	fileType := ws.GetFileType(uri)
	var hover *Hover

	switch fileType {
	case epub.FileTypeOPF:
		hover = hoverOPF(result, ws)
	case epub.FileTypeXHTML, epub.FileTypeNav:
		hover = hoverXHTML(result)
	}

	if hover == nil {
		return marshalNullResponse(req.Id)
	}

	return marshalResponse(req.Id, hover)
}

func hoverOPF(result *parser.LocateResult, ws WorkspaceReader) *Hover {
	node := result.Node

	// <itemref idref="x"> → show manifest item details
	if node.Local == "itemref" && result.Attr != nil && result.Attr.Local == "idref" &&
		result.InValue {
		manifest := ws.GetManifest()
		if manifest != nil {
			for _, item := range manifest.Items {
				if item.ID == result.Attr.Value {
					text := fmt.Sprintf(
						"**Manifest Item**\n- **ID:** %s\n- **Href:** %s\n- **Media-Type:** %s",
						item.ID,
						item.Href,
						item.MediaType,
					)
					return &Hover{Contents: MarkupContent{Kind: "markdown", Value: text}}
				}
			}
		}
	}

	// <meta property="schema:..."> → show property docs
	if node.Local == "meta" {
		prop := node.Attr("property")
		if doc, ok := schemaPropertyDocs[prop]; ok {
			return &Hover{Contents: MarkupContent{Kind: "markdown", Value: doc}}
		}
	}

	// epub:type values → show ARIA role mapping
	if result.Attr != nil && result.Attr.Local == "type" &&
		result.Attr.Space == epub.NSEpub &&
		result.InValue {
		value := result.Attr.Value
		// Check each token in the value
		for token := range strings.FieldsSeq(value) {
			if doc, ok := epubTypeDocs[token]; ok {
				return &Hover{Contents: MarkupContent{Kind: "markdown", Value: doc}}
			}
		}
	}

	// dc:* elements → show Dublin Core docs
	if node.Space == epub.NSDC {
		if doc, ok := dcElementDocs[node.Local]; ok {
			return &Hover{Contents: MarkupContent{Kind: "markdown", Value: doc}}
		}
	}

	return nil
}

func hoverXHTML(result *parser.LocateResult) *Hover {
	// epub:type values
	if result.Attr != nil && result.Attr.Local == "type" &&
		result.Attr.Space == epub.NSEpub &&
		result.InValue {
		for token := range strings.FieldsSeq(result.Attr.Value) {
			if doc, ok := epubTypeDocs[token]; ok {
				return &Hover{Contents: MarkupContent{Kind: "markdown", Value: doc}}
			}
		}
	}
	return nil
}

func marshalNullResponse(id ID) []byte {
	res := ResponseMessage[any]{
		JsonRpc: JSONRPCVersion,
		Id:      id,
		Result:  nil,
	}
	data, err := json.Marshal(res)
	if err != nil {
		slog.Error("error marshalling null response: " + err.Error())
		return nil
	}
	return data
}
