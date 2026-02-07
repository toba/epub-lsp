package accessibility

import (
	"slices"
	"strings"

	"github.com/toba/epub-lsp/internal/epub"
	"github.com/toba/epub-lsp/internal/epub/parser"
	"github.com/toba/epub-lsp/internal/epub/validator"
)

const epubNS = "http://www.idpf.org/2007/ops"

// PageValidator checks page navigation consistency across OPF, nav, and content docs.
// It runs on OPF files but reads cross-file context.
type PageValidator struct{}

func (v *PageValidator) FileTypes() []epub.FileType {
	return []epub.FileType{epub.FileTypeOPF}
}

func (v *PageValidator) Validate(
	_ string,
	content []byte,
	ctx *validator.WorkspaceContext,
) []epub.Diagnostic {
	if ctx == nil || ctx.Manifest == nil {
		return nil
	}

	root, xmlDiags := parser.Parse(content)
	if len(xmlDiags) > 0 {
		return nil
	}

	pkg := root.FindFirst("package")
	if pkg == nil {
		return nil
	}

	metadata := pkg.FindFirst("metadata")
	if metadata == nil {
		return nil
	}

	pos := epub.ByteOffsetToPosition(content, int(metadata.Offset))
	rng := epub.Range{Start: pos, End: pos}

	meta := &ctx.Manifest.Metadata
	hasPrintPageNumbers := slices.Contains(meta.AccessibilityFeatures, "printPageNumbers")

	var diags []epub.Diagnostic

	if hasPrintPageNumbers {
		// Check for page-list in nav documents
		hasPageList := navHasPageList(ctx)
		if !hasPageList {
			diags = append(diags, epub.Diagnostic{
				Code:     "printPageNumbers-nopagelist",
				Severity: epub.SeverityError,
				Message:  "printPageNumbers feature declared but no page-list nav found",
				Source:   source,
				Range:    rng,
			})
		}

		// Check for pagebreak markers in content documents
		hasPageBreaks := contentHasPageBreaks(ctx)
		if !hasPageBreaks {
			diags = append(diags, epub.Diagnostic{
				Code:     "printPageNumbers-nopagebreaks",
				Severity: epub.SeverityError,
				Message:  "printPageNumbers feature declared but no epub:type=\"pagebreak\" found in content",
				Source:   source,
				Range:    rng,
			})
		}
	}

	// Check for page list without dc:source in reflowable content
	if navHasPageList(ctx) && !meta.HasDCSource {
		diags = append(diags, epub.Diagnostic{
			Code:     "epub-pagesource",
			Severity: epub.SeverityWarning,
			Message:  "page list present but missing dc:source metadata",
			Source:   source,
			Range:    rng,
		})
	}

	// Check page-list references point to existing element IDs
	diags = append(diags, checkPageListReferences(content, rng, ctx)...)

	return diags
}

// navHasPageList checks if any nav document in the workspace has a page-list nav.
func navHasPageList(ctx *validator.WorkspaceContext) bool {
	for uri, content := range ctx.Files {
		ft := ctx.FileTypes[uri]
		if ft != epub.FileTypeNav {
			continue
		}
		root, diags := parser.Parse(content)
		if len(diags) > 0 {
			continue
		}
		for _, nav := range root.FindAll("nav") {
			if nav.AttrNS(epubNS, "type") == "page-list" {
				return true
			}
		}
	}
	return false
}

// contentHasPageBreaks checks if any XHTML content document has epub:type="pagebreak".
func contentHasPageBreaks(ctx *validator.WorkspaceContext) bool {
	for uri, content := range ctx.Files {
		ft := ctx.FileTypes[uri]
		if ft != epub.FileTypeXHTML && ft != epub.FileTypeNav {
			continue
		}
		if hasPageBreakInContent(content) {
			return true
		}
	}
	return false
}

func hasPageBreakInContent(content []byte) bool {
	root, diags := parser.Parse(content)
	if len(diags) > 0 {
		return false
	}
	return findPageBreak(root)
}

func findPageBreak(node *parser.XMLNode) bool {
	for _, child := range node.Children {
		epubType := child.AttrNS(epubNS, "type")
		if containsToken(epubType, "pagebreak") {
			return true
		}
		if findPageBreak(child) {
			return true
		}
	}
	return false
}

// checkPageListReferences verifies page-list hrefs point to real element IDs.
func checkPageListReferences(
	_ []byte,
	rng epub.Range,
	ctx *validator.WorkspaceContext,
) []epub.Diagnostic {
	var diags []epub.Diagnostic

	// Find page-list nav
	for uri, navContent := range ctx.Files {
		ft := ctx.FileTypes[uri]
		if ft != epub.FileTypeNav {
			continue
		}
		root, xmlDiags := parser.Parse(navContent)
		if len(xmlDiags) > 0 {
			continue
		}
		for _, nav := range root.FindAll("nav") {
			if nav.AttrNS(epubNS, "type") != "page-list" {
				continue
			}
			// Check each link in the page list
			for _, a := range nav.FindAll("a") {
				href := a.Attr("href")
				if href == "" {
					continue
				}
				parts := strings.SplitN(href, "#", 2)
				if len(parts) != 2 || parts[1] == "" {
					continue
				}
				targetFile := parts[0]
				targetID := parts[1]

				if !idExistsInFile(targetFile, targetID, ctx) {
					diags = append(diags, epub.Diagnostic{
						Code:     "epub-pagelist-broken",
						Severity: epub.SeverityError,
						Message:  "page list references nonexistent id \"" + targetID + "\" in " + targetFile,
						Source:   source,
						Range:    rng,
					})
				}
			}
		}
	}

	return diags
}

// idExistsInFile checks if an element with the given id exists in a workspace file.
func idExistsInFile(filename, id string, ctx *validator.WorkspaceContext) bool {
	for uri, content := range ctx.Files {
		if !strings.HasSuffix(uri, filename) {
			continue
		}
		root, diags := parser.Parse(content)
		if len(diags) > 0 {
			continue
		}
		if findElementByID(root, id) {
			return true
		}
	}
	return false
}

func findElementByID(node *parser.XMLNode, id string) bool {
	for _, child := range node.Children {
		if child.Attr("id") == id {
			return true
		}
		if findElementByID(child, id) {
			return true
		}
	}
	return false
}

// containsToken checks if a space-separated token list contains the given token.
func containsToken(tokenList, token string) bool {
	return slices.Contains(strings.Fields(tokenList), token)
}
