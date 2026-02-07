package lsp

import (
	"encoding/json"
	"log/slog"
	"net/url"
	"strings"

	"github.com/toba/epub-lsp/internal/epub"
	"github.com/toba/epub-lsp/internal/epub/parser"
)

// HandleDefinition processes textDocument/definition requests.
func HandleDefinition(data []byte, ws WorkspaceReader) []byte {
	var req RequestMessage[DefinitionParams]
	if err := json.Unmarshal(data, &req); err != nil {
		slog.Error("error unmarshalling definition: " + err.Error())
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

	fileType := ws.GetFileType(uri)
	var locations []Location

	root, xmlDiags := parser.Parse(content)
	if len(xmlDiags) > 0 {
		return marshalResponse(req.Id, []Location{})
	}

	result := parser.LocateAtPosition(root, content, offset)
	if result == nil {
		return marshalResponse(req.Id, []Location{})
	}

	switch fileType {
	case epub.FileTypeOPF:
		locations = definitionInOPF(result, content, uri, root, ws)
	case epub.FileTypeXHTML, epub.FileTypeNav:
		locations = definitionInXHTML(result, content, uri, ws)
	}

	return marshalResponse(req.Id, locations)
}

func definitionInOPF(
	result *parser.LocateResult,
	content []byte,
	uri string,
	root *parser.XMLNode,
	ws WorkspaceReader,
) []Location {
	if result.Attr == nil || !result.InValue {
		return nil
	}

	node := result.Node
	attr := result.Attr

	// <itemref idref="x"> → jump to <item id="x">
	if node.Local == "itemref" && attr.Local == "idref" {
		return findManifestItemByID(root, content, uri, attr.Value)
	}

	// <item href="x"> → jump to the referenced file
	if node.Local == "item" && attr.Local == "href" {
		baseDir := dirFromURI(uri)
		target := resolveToFileURI(baseDir, attr.Value, uri)
		targetContent := ws.GetContent(target)
		if targetContent != nil {
			return []Location{{URI: target, Range: Range{}}}
		}
		// Try finding in all files by path suffix
		for fileURI := range ws.GetAllFiles() {
			if pathEndsWith(fileURI, attr.Value) {
				return []Location{{URI: fileURI, Range: Range{}}}
			}
		}
	}

	// unique-identifier="x" → jump to dc:identifier id="x"
	if node.Local == "package" && attr.Local == "unique-identifier" {
		return findElementByID(root, content, uri, attr.Value)
	}

	return nil
}

func definitionInXHTML(
	result *parser.LocateResult,
	content []byte,
	uri string,
	ws WorkspaceReader,
) []Location {
	if result.Attr == nil || !result.InValue {
		return nil
	}

	attr := result.Attr

	// <a href="file.xhtml#id"> → jump to element with matching id in target file
	if attr.Local == "href" && !epub.IsRemoteURL(attr.Value) {
		return resolveHrefTarget(attr.Value, content, uri, ws)
	}

	return nil
}

func resolveHrefTarget(href string, _ []byte, uri string, ws WorkspaceReader) []Location {
	filePart, fragment, hasFragment := strings.Cut(href, "#")

	baseDir := dirFromURI(uri)

	var targetURI string
	var targetContent []byte

	if filePart == "" {
		// Same-file reference
		targetURI = uri
		targetContent = ws.GetContent(uri)
	} else {
		target := resolveToFileURI(baseDir, filePart, uri)
		targetContent = ws.GetContent(target)
		if targetContent != nil {
			targetURI = target
		} else {
			// Try all files
			for fileURI, c := range ws.GetAllFiles() {
				if pathEndsWith(fileURI, filePart) {
					targetURI = fileURI
					targetContent = c
					break
				}
			}
		}
	}

	if targetContent == nil {
		return nil
	}

	if !hasFragment || fragment == "" {
		return []Location{{URI: targetURI, Range: Range{}}}
	}

	// Find element with matching id in target file
	targetRoot, diags := parser.Parse(targetContent)
	if len(diags) > 0 {
		return nil
	}

	return findElementByID(targetRoot, targetContent, targetURI, fragment)
}

func findManifestItemByID(
	root *parser.XMLNode,
	content []byte,
	uri, id string,
) []Location {
	items := root.FindAll("item")
	for _, item := range items {
		if item.Attr("id") == id {
			pos := epub.ByteOffsetToPosition(content, int(item.Offset))
			return []Location{{
				URI:   uri,
				Range: Range{Start: lspPos(pos), End: lspPos(pos)},
			}}
		}
	}
	return nil
}

func findElementByID(root *parser.XMLNode, content []byte, uri, id string) []Location {
	return findElementByIDRecursive(root, content, uri, id)
}

func findElementByIDRecursive(
	node *parser.XMLNode,
	content []byte,
	uri, id string,
) []Location {
	if node.Attr("id") == id {
		pos := epub.ByteOffsetToPosition(content, int(node.Offset))
		return []Location{{
			URI:   uri,
			Range: Range{Start: lspPos(pos), End: lspPos(pos)},
		}}
	}
	for _, child := range node.Children {
		if locs := findElementByIDRecursive(child, content, uri, id); len(locs) > 0 {
			return locs
		}
	}
	return nil
}

// pathEndsWith checks if a URI path ends with the given suffix.
func pathEndsWith(uri, suffix string) bool {
	if u, err := url.Parse(uri); err == nil {
		p := u.Path
		return p == suffix || strings.HasSuffix(p, "/"+suffix)
	}
	return strings.HasSuffix(uri, "/"+suffix)
}
