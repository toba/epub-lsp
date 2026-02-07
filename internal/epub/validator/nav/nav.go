// Package nav validates EPUB navigation documents.
package nav

import (
	"strings"

	"github.com/toba/epub-lsp/internal/epub"
	"github.com/toba/epub-lsp/internal/epub/parser"
	"github.com/toba/epub-lsp/internal/epub/validator"
)

const (
	source  = "epub-nav"
	epubNS  = "http://www.idpf.org/2007/ops"
	xhtmlNS = "http://www.w3.org/1999/xhtml"
)

// Validator validates EPUB navigation documents.
type Validator struct{}

func (v *Validator) FileTypes() []epub.FileType {
	return []epub.FileType{epub.FileTypeNav}
}

func (v *Validator) Validate(
	_ string,
	content []byte,
	ctx *validator.WorkspaceContext,
) []epub.Diagnostic {
	root, diags := parser.Parse(content)
	if len(diags) > 0 {
		return diags
	}

	diags = append(diags, validateTocNav(content, root)...)
	diags = append(diags, validateNavLinks(content, root)...)
	diags = append(diags, validateNavTypes(content, root)...)

	if ctx != nil && ctx.Manifest != nil {
		diags = append(diags, validateTocSpineOrder(content, root, ctx)...)
	}

	return diags
}

// findNavElements returns all <nav> elements from the document.
func findNavElements(root *parser.XMLNode) []*parser.XMLNode {
	return root.FindAll("nav")
}

// getEpubType returns the epub:type attribute value of a node.
func getEpubType(node *parser.XMLNode) string {
	return node.AttrNS(epubNS, "type")
}

// validateTocNav checks that a <nav epub:type="toc"> element exists and has an <ol>.
func validateTocNav(content []byte, root *parser.XMLNode) []epub.Diagnostic {
	var diags []epub.Diagnostic

	navs := findNavElements(root)
	var tocNav *parser.XMLNode

	for _, nav := range navs {
		if getEpubType(nav) == "toc" {
			tocNav = nav
			break
		}
	}

	if tocNav == nil {
		// Report at document root
		html := root.FindFirst("html")
		var pos epub.Position
		if html != nil {
			pos = epub.ByteOffsetToPosition(content, int(html.Offset))
		}
		diags = append(diags, epub.Diagnostic{
			Code:     "NAV_003",
			Severity: epub.SeverityError,
			Message:  `no <nav epub:type="toc"> element found`,
			Source:   source,
			Range:    epub.Range{Start: pos, End: pos},
		})
		return diags
	}

	// Check for <ol> inside toc nav
	ol := tocNav.FindFirst("ol")
	if ol == nil {
		pos := epub.ByteOffsetToPosition(content, int(tocNav.Offset))
		diags = append(diags, epub.Diagnostic{
			Severity: epub.SeverityWarning,
			Message:  "toc nav is missing required <ol> element",
			Source:   source,
			Range:    epub.Range{Start: pos, End: pos},
		})
	}

	return diags
}

// validateNavLinks checks that nav links don't reference remote resources.
func validateNavLinks(content []byte, root *parser.XMLNode) []epub.Diagnostic {
	var diags []epub.Diagnostic

	links := root.FindAll("a")
	for _, a := range links {
		href := a.Attr("href")
		if href == "" {
			continue
		}
		if strings.HasPrefix(href, "http://") || strings.HasPrefix(href, "https://") {
			pos := epub.ByteOffsetToPosition(content, int(a.Offset))
			diags = append(diags, epub.Diagnostic{
				Code:     "NAV_010",
				Severity: epub.SeverityError,
				Message:  "nav links to remote resource: " + href,
				Source:   source,
				Range:    epub.Range{Start: pos, End: pos},
			})
		}
	}

	return diags
}

// validateNavTypes checks for informational nav types.
func validateNavTypes(content []byte, root *parser.XMLNode) []epub.Diagnostic {
	var diags []epub.Diagnostic

	navs := findNavElements(root)
	hasPageList := false
	hasLandmarks := false

	for _, nav := range navs {
		epubType := getEpubType(nav)
		if epubType == "page-list" {
			hasPageList = true
		}
		if epubType == "landmarks" {
			hasLandmarks = true
		}
	}

	if !hasPageList && !hasLandmarks {
		html := root.FindFirst("html")
		var pos epub.Position
		if html != nil {
			pos = epub.ByteOffsetToPosition(content, int(html.Offset))
		}
		diags = append(diags, epub.Diagnostic{
			Severity: epub.SeverityInfo,
			Message:  "navigation document has no page-list or landmarks nav",
			Source:   source,
			Range:    epub.Range{Start: pos, End: pos},
		})
	}

	return diags
}

// validateTocSpineOrder checks that TOC link order matches spine order.
func validateTocSpineOrder(
	content []byte,
	root *parser.XMLNode,
	ctx *validator.WorkspaceContext,
) []epub.Diagnostic {
	var diags []epub.Diagnostic

	navs := findNavElements(root)
	var tocNav *parser.XMLNode
	for _, nav := range navs {
		if getEpubType(nav) == "toc" {
			tocNav = nav
			break
		}
	}

	if tocNav == nil || ctx.Manifest == nil {
		return diags
	}

	// Extract hrefs from toc nav links (in order)
	tocHrefs := extractNavHrefs(tocNav)
	if len(tocHrefs) == 0 {
		return diags
	}

	// Build spine order: idref -> index, then manifest id -> href
	idToHref := make(map[string]string)
	for _, item := range ctx.Manifest.Items {
		idToHref[item.ID] = item.Href
	}

	spineHrefOrder := make([]string, 0, len(ctx.Manifest.Spine))
	for _, s := range ctx.Manifest.Spine {
		if href, ok := idToHref[s.IDRef]; ok {
			spineHrefOrder = append(spineHrefOrder, href)
		}
	}

	// Map spine href to index
	spineIndex := make(map[string]int)
	for i, href := range spineHrefOrder {
		spineIndex[href] = i
	}

	// Check that toc hrefs are in non-decreasing spine order
	lastSpineIdx := -1
	for _, tocHref := range tocHrefs {
		// Strip fragment
		base := tocHref
		if idx := strings.Index(base, "#"); idx >= 0 {
			base = base[:idx]
		}

		if si, ok := spineIndex[base]; ok {
			if si < lastSpineIdx {
				pos := epub.ByteOffsetToPosition(content, int(tocNav.Offset))
				diags = append(diags, epub.Diagnostic{
					Code:     "NAV_011",
					Severity: epub.SeverityWarning,
					Message:  "TOC link order doesn't match spine order",
					Source:   source,
					Range:    epub.Range{Start: pos, End: pos},
				})
				break
			}
			lastSpineIdx = si
		}
	}

	return diags
}

// extractNavHrefs returns all href values from <a> elements within a nav, in order.
func extractNavHrefs(nav *parser.XMLNode) []string {
	var hrefs []string
	var walk func(node *parser.XMLNode)
	walk = func(node *parser.XMLNode) {
		for _, child := range node.Children {
			if child.Local == "a" {
				href := child.Attr("href")
				if href != "" {
					hrefs = append(hrefs, href)
				}
			}
			walk(child)
		}
	}
	walk(nav)
	return hrefs
}
