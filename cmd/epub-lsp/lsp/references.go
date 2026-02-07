package lsp

import (
	"encoding/json"
	"log/slog"
	"strings"

	"github.com/toba/epub-lsp/internal/epub"
	"github.com/toba/epub-lsp/internal/epub/parser"
)

// HandleReferences processes textDocument/references requests.
func HandleReferences(data []byte, ws WorkspaceReader) []byte {
	var req RequestMessage[ReferenceParams]
	if err := json.Unmarshal(data, &req); err != nil {
		slog.Error("error unmarshalling references: " + err.Error())
		return marshalResponse(req.Id, []Location{})
	}

	uri := req.Params.TextDocument.Uri
	content := ws.GetContent(uri)
	if content == nil {
		return marshalResponse(req.Id, []Location{})
	}

	pos := posToEpub(req.Params.Position)
	offset := epub.PositionToByteOffset(content, pos)
	if offset < 0 {
		return marshalResponse(req.Id, []Location{})
	}

	root, xmlDiags := parser.Parse(content)
	if len(xmlDiags) > 0 {
		return marshalResponse(req.Id, []Location{})
	}

	result := parser.LocateAtPosition(root, content, offset)
	if result == nil {
		return marshalResponse(req.Id, []Location{})
	}

	fileType := ws.GetFileType(uri)
	var locations []Location

	switch fileType {
	case epub.FileTypeOPF:
		locations = referencesInOPF(result, uri, ws)
	case epub.FileTypeXHTML, epub.FileTypeNav:
		locations = referencesInXHTML(result, uri, ws)
	}

	return marshalResponse(req.Id, locations)
}

func referencesInOPF(
	result *parser.LocateResult,
	uri string,
	ws WorkspaceReader,
) []Location {
	node := result.Node

	// On <item id="x"> → find all <itemref idref="x"> and href references
	if node.Local == "item" {
		id := node.Attr("id")
		href := node.Attr("href")
		if id == "" {
			return nil
		}
		return findManifestItemReferences(id, href, uri, ws)
	}

	return nil
}

func referencesInXHTML(
	result *parser.LocateResult,
	uri string,
	ws WorkspaceReader,
) []Location {
	node := result.Node

	// On element with id="x" → find all href="...#x" references
	if result.Attr != nil && result.Attr.Local == "id" && result.InValue {
		return findIDReferences(result.Attr.Value, uri, ws)
	}

	// If on the element itself and it has an id, also look for references
	id := node.Attr("id")
	if id != "" {
		return findIDReferences(id, uri, ws)
	}

	return nil
}

func findManifestItemReferences(id, href, opfURI string, ws WorkspaceReader) []Location {
	var locations []Location

	// Search in OPF for <itemref idref="id">
	opfContent := ws.GetContent(opfURI)
	if opfContent != nil {
		opfRoot, diags := parser.Parse(opfContent)
		if len(diags) == 0 {
			for _, itemref := range opfRoot.FindAll("itemref") {
				if itemref.Attr("idref") == id {
					pos := epub.ByteOffsetToPosition(opfContent, int(itemref.Offset))
					locations = append(locations, Location{
						URI:   opfURI,
						Range: Range{Start: lspPos(pos), End: lspPos(pos)},
					})
				}
			}
		}
	}

	// Search all files for href references to this item's href
	if href != "" {
		for fileURI, content := range ws.GetAllFiles() {
			ft := ws.GetFileType(fileURI)
			if ft != epub.FileTypeXHTML && ft != epub.FileTypeNav {
				continue
			}
			locations = append(
				locations,
				findHrefReferencesInFile(fileURI, content, href)...)
		}
	}

	return locations
}

func findIDReferences(id, sourceURI string, ws WorkspaceReader) []Location {
	var locations []Location

	// Determine the filename for this URI to match against hrefs
	sourcePath := dirFromURI(sourceURI)
	_ = sourcePath

	for fileURI, content := range ws.GetAllFiles() {
		ft := ws.GetFileType(fileURI)
		if ft != epub.FileTypeXHTML && ft != epub.FileTypeNav && ft != epub.FileTypeOPF {
			continue
		}

		root, diags := parser.Parse(content)
		if len(diags) > 0 {
			continue
		}

		// Search for href attributes containing #id
		findHrefWithFragment(root, content, fileURI, id, &locations)
	}

	return locations
}

func findHrefReferencesInFile(fileURI string, content []byte, href string) []Location {
	root, diags := parser.Parse(content)
	if len(diags) > 0 {
		return nil
	}

	var locations []Location

	for _, a := range root.FindAll("a") {
		aHref := a.Attr("href")
		stripped := epub.StripFragment(aHref)
		if stripped == href || strings.HasSuffix(stripped, "/"+href) {
			pos := epub.ByteOffsetToPosition(content, int(a.Offset))
			locations = append(locations, Location{
				URI:   fileURI,
				Range: Range{Start: lspPos(pos), End: lspPos(pos)},
			})
		}
	}

	return locations
}

func findHrefWithFragment(
	node *parser.XMLNode,
	content []byte,
	fileURI, id string,
	locations *[]Location,
) {
	for _, attr := range node.Attrs {
		if attr.Local == "href" {
			_, fragment, hasFragment := strings.Cut(attr.Value, "#")
			if hasFragment && fragment == id {
				pos := epub.ByteOffsetToPosition(content, int(node.Offset))
				*locations = append(*locations, Location{
					URI:   fileURI,
					Range: Range{Start: lspPos(pos), End: lspPos(pos)},
				})
			}
		}
	}
	for _, child := range node.Children {
		findHrefWithFragment(child, content, fileURI, id, locations)
	}
}
