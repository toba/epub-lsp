package lsp

import (
	"encoding/json"
	"log/slog"
	"net/url"
	"path"
	"strings"

	"github.com/toba/epub-lsp/internal/epub"
	"github.com/toba/epub-lsp/internal/epub/parser"
)

// HandleDocumentLink processes textDocument/documentLink requests.
func HandleDocumentLink(data []byte, ws WorkspaceReader) []byte {
	var req RequestMessage[DocumentLinkParams]
	if err := json.Unmarshal(data, &req); err != nil {
		slog.Error("error unmarshalling documentLink: " + err.Error())
		return marshalResponse(req.Id, []DocumentLink{})
	}

	uri := req.Params.TextDocument.Uri
	content := ws.GetContent(uri)
	if content == nil {
		return marshalResponse(req.Id, []DocumentLink{})
	}

	fileType := ws.GetFileType(uri)
	var links []DocumentLink

	switch fileType {
	case epub.FileTypeOPF:
		links = extractOPFLinks(content, uri)
	case epub.FileTypeXHTML, epub.FileTypeNav:
		links = extractXHTMLLinks(content, uri)
	}

	return marshalResponse(req.Id, links)
}

func extractOPFLinks(content []byte, uri string) []DocumentLink {
	root, diags := parser.Parse(content)
	if len(diags) > 0 {
		return nil
	}

	baseDir := dirFromURI(uri)
	var links []DocumentLink

	// Find manifest items with href attributes
	items := root.FindAll("item")
	for _, item := range items {
		href := item.Attr("href")
		if href == "" || epub.IsRemoteURL(href) {
			continue
		}
		if r, ok := findAttrValueRange(content, int(item.Offset), "href"); ok {
			target := resolveToFileURI(baseDir, href, uri)
			links = append(links, DocumentLink{Range: r, Target: target})
		}
	}

	return links
}

func extractXHTMLLinks(content []byte, uri string) []DocumentLink {
	root, diags := parser.Parse(content)
	if len(diags) > 0 {
		return nil
	}

	baseDir := dirFromURI(uri)
	var links []DocumentLink

	// <a href="...">
	for _, node := range root.FindAll("a") {
		href := node.Attr("href")
		if href == "" || epub.IsRemoteURL(href) {
			continue
		}
		if r, ok := findAttrValueRange(content, int(node.Offset), "href"); ok {
			target := resolveToFileURI(baseDir, href, uri)
			links = append(links, DocumentLink{Range: r, Target: target})
		}
	}

	// <img src="...">
	for _, node := range root.FindAll("img") {
		src := node.Attr("src")
		if src == "" || epub.IsRemoteURL(src) || strings.HasPrefix(src, "data:") {
			continue
		}
		if r, ok := findAttrValueRange(content, int(node.Offset), "src"); ok {
			target := resolveToFileURI(baseDir, src, uri)
			links = append(links, DocumentLink{Range: r, Target: target})
		}
	}

	// <link href="...">
	for _, node := range root.FindAll("link") {
		href := node.Attr("href")
		if href == "" || epub.IsRemoteURL(href) {
			continue
		}
		if r, ok := findAttrValueRange(content, int(node.Offset), "href"); ok {
			target := resolveToFileURI(baseDir, href, uri)
			links = append(links, DocumentLink{Range: r, Target: target})
		}
	}

	// <source src="...">, <audio src="...">, <video src="...">
	for _, tag := range []string{"source", "audio", "video"} {
		for _, node := range root.FindAll(tag) {
			src := node.Attr("src")
			if src == "" || epub.IsRemoteURL(src) || strings.HasPrefix(src, "data:") {
				continue
			}
			if r, ok := findAttrValueRange(content, int(node.Offset), "src"); ok {
				target := resolveToFileURI(baseDir, src, uri)
				links = append(links, DocumentLink{Range: r, Target: target})
			}
		}
	}

	return links
}

// findAttrValueRange finds the range of an attribute value in raw content
// starting from a tag offset. Returns the range covering just the value text.
func findAttrValueRange(content []byte, tagOffset int, attrName string) (Range, bool) {
	// Search for attrName= within the tag
	searchStart := tagOffset
	tagEnd := findStartTagEndByte(content, tagOffset)

	for i := searchStart; i < tagEnd; i++ {
		if i+len(attrName) >= len(content) {
			break
		}
		if string(content[i:i+len(attrName)]) == attrName {
			// Check it's followed by =
			j := i + len(attrName)
			for j < tagEnd && (content[j] == ' ' || content[j] == '\t') {
				j++
			}
			if j < tagEnd && content[j] == '=' {
				j++
				for j < tagEnd && (content[j] == ' ' || content[j] == '\t') {
					j++
				}
				if j < tagEnd && (content[j] == '"' || content[j] == '\'') {
					quote := content[j]
					valueStart := j + 1
					valueEnd := valueStart
					for valueEnd < len(content) && content[valueEnd] != quote {
						valueEnd++
					}
					startPos := epub.ByteOffsetToPosition(content, valueStart)
					endPos := epub.ByteOffsetToPosition(content, valueEnd)
					return Range{
						Start: Position{
							Line:      intToUint(startPos.Line),
							Character: intToUint(startPos.Character),
						},
						End: Position{
							Line:      intToUint(endPos.Line),
							Character: intToUint(endPos.Character),
						},
					}, true
				}
			}
		}
	}
	return Range{}, false
}

// findStartTagEndByte finds the byte offset of '>' closing the start tag.
func findStartTagEndByte(content []byte, tagStart int) int {
	inString := byte(0)
	for i := tagStart; i < len(content); i++ {
		ch := content[i]
		if inString != 0 {
			if ch == inString {
				inString = 0
			}
			continue
		}
		if ch == '"' || ch == '\'' {
			inString = ch
			continue
		}
		if ch == '>' {
			return i
		}
	}
	return len(content)
}

// dirFromURI extracts the directory portion of a URI.
func dirFromURI(uri string) string {
	if u, err := url.Parse(uri); err == nil && u.Path != "" {
		return path.Dir(u.Path)
	}
	idx := strings.LastIndex(uri, "/")
	if idx >= 0 {
		return uri[:idx]
	}
	return ""
}

// resolveToFileURI resolves a relative href to a file:// URI.
func resolveToFileURI(baseDir, href, originURI string) string {
	if decoded, err := url.PathUnescape(href); err == nil {
		href = decoded
	}
	resolved := path.Clean(baseDir + "/" + epub.StripFragment(href))

	// Reconstruct as file URI using the origin URI's scheme
	if u, err := url.Parse(originURI); err == nil && u.Scheme == "file" {
		return "file://" + resolved
	}
	return resolved
}
