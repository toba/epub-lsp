package accessibility

import (
	"strconv"
	"strings"

	"github.com/toba/epub-lsp/internal/epub"
	"github.com/toba/epub-lsp/internal/epub/parser"
	"github.com/toba/epub-lsp/internal/epub/validator"
)

// epubTypeToRole maps epub:type values to their corresponding ARIA roles.
var epubTypeToRole = map[string]string{
	"abstract":        "doc-abstract",
	"acknowledgments": "doc-acknowledgments",
	"afterword":       "doc-afterword",
	"appendix":        "doc-appendix",
	"biblioentry":     "doc-biblioentry",
	"bibliography":    "doc-bibliography",
	"biblioref":       "doc-biblioref",
	"chapter":         "doc-chapter",
	"colophon":        "doc-colophon",
	"conclusion":      "doc-conclusion",
	"cover":           "doc-cover",
	"credit":          "doc-credit",
	"credits":         "doc-credits",
	"dedication":      "doc-dedication",
	"endnote":         "doc-endnote",
	"endnotes":        "doc-endnotes",
	"epigraph":        "doc-epigraph",
	"epilogue":        "doc-epilogue",
	"errata":          "doc-errata",
	"footnote":        "doc-footnote",
	"foreword":        "doc-foreword",
	"glossary":        "doc-glossary",
	"glossdef":        "definition",
	"glossref":        "doc-glossref",
	"glossterm":       "term",
	"index":           "doc-index",
	"introduction":    "doc-introduction",
	"noteref":         "doc-noteref",
	"notice":          "doc-notice",
	"pagebreak":       "doc-pagebreak",
	"page-list":       "doc-pagelist",
	"part":            "doc-part",
	"preface":         "doc-preface",
	"prologue":        "doc-prologue",
	"pullquote":       "doc-pullquote",
	"qna":             "doc-qna",
	"subtitle":        "doc-subtitle",
	"tip":             "doc-tip",
	"toc":             "doc-toc",
}

// StructureValidator checks epub:type / ARIA role mapping and accessibility
// rules in XHTML content documents.
type StructureValidator struct{}

func (v *StructureValidator) FileTypes() []epub.FileType {
	return []epub.FileType{epub.FileTypeXHTML, epub.FileTypeNav}
}

func (v *StructureValidator) Validate(
	_ string,
	content []byte,
	_ *validator.WorkspaceContext,
) []epub.Diagnostic {
	root, xmlDiags := parser.Parse(content)
	if len(xmlDiags) > 0 {
		return nil
	}

	var diags []epub.Diagnostic //nolint:prealloc // size unknown
	diags = append(diags, checkEpubTypeRoles(content, root)...)
	diags = append(diags, checkPageBreakLabels(content, root)...)
	diags = append(diags, checkHeadingLevels(content, root)...)
	diags = append(diags, checkTableCaptions(content, root)...)
	diags = append(diags, checkFormLabels(content, root)...)

	return diags
}

// checkEpubTypeRoles checks that elements with epub:type have a matching ARIA role.
func checkEpubTypeRoles(content []byte, root *parser.XMLNode) []epub.Diagnostic {
	var diags []epub.Diagnostic
	walkEpubTypes(root, func(node *parser.XMLNode, epubType string) {
		for token := range strings.FieldsSeq(epubType) {
			expectedRole, ok := epubTypeToRole[token]
			if !ok {
				continue
			}
			actualRole := node.Attr("role")
			if actualRole == "" || !containsToken(actualRole, expectedRole) {
				pos := epub.ByteOffsetToPosition(content, int(node.Offset))
				diags = append(diags, epub.Diagnostic{
					Code:     "epub-type-has-matching-role",
					Severity: epub.SeverityWarning,
					Message:  "epub:type=\"" + token + "\" should have role=\"" + expectedRole + "\"",
					Source:   source,
					Range:    epub.Range{Start: pos, End: pos},
				})
			}
		}
	})
	return diags
}

// checkPageBreakLabels checks that pagebreak elements have accessible labels.
func checkPageBreakLabels(content []byte, root *parser.XMLNode) []epub.Diagnostic {
	var diags []epub.Diagnostic
	walkEpubTypes(root, func(node *parser.XMLNode, epubType string) {
		if !containsToken(epubType, "pagebreak") {
			return
		}
		// Must have aria-label, title, or text content
		ariaLabel := node.Attr("aria-label")
		title := node.Attr("title")
		text := strings.TrimSpace(node.CharData)

		if ariaLabel == "" && title == "" && text == "" {
			pos := epub.ByteOffsetToPosition(content, int(node.Offset))
			diags = append(diags, epub.Diagnostic{
				Code:     "pagebreak-label",
				Severity: epub.SeverityWarning,
				Message:  "pagebreak element missing accessible label (aria-label, title, or text content)",
				Source:   source,
				Range:    epub.Range{Start: pos, End: pos},
			})
		}
	})
	return diags
}

// checkHeadingLevels checks that heading levels don't skip (e.g. h1 → h3).
func checkHeadingLevels(content []byte, root *parser.XMLNode) []epub.Diagnostic {
	var diags []epub.Diagnostic
	var headings []headingInfo

	collectHeadings(root, &headings)

	for i := 1; i < len(headings); i++ {
		prev := headings[i-1]
		curr := headings[i]
		if curr.level > prev.level+1 {
			pos := epub.ByteOffsetToPosition(content, int(curr.offset))
			diags = append(diags, epub.Diagnostic{
				Code:     "heading-order",
				Severity: epub.SeverityWarning,
				Message: "heading level skipped from h" + strconv.Itoa(
					prev.level,
				) + " to h" + strconv.Itoa(
					curr.level,
				),
				Source: source,
				Range:  epub.Range{Start: pos, End: pos},
			})
		}
	}

	return diags
}

type headingInfo struct {
	level  int
	offset int64
}

func collectHeadings(node *parser.XMLNode, headings *[]headingInfo) {
	for _, child := range node.Children {
		if level := headingLevel(child.Local); level > 0 {
			*headings = append(*headings, headingInfo{level: level, offset: child.Offset})
		}
		collectHeadings(child, headings)
	}
}

func headingLevel(local string) int {
	switch local {
	case "h1":
		return 1
	case "h2":
		return 2
	case "h3":
		return 3
	case "h4":
		return 4
	case "h5":
		return 5
	case "h6":
		return 6
	}
	return 0
}

// checkTableCaptions checks that tables have a <caption> or aria-label.
func checkTableCaptions(content []byte, root *parser.XMLNode) []epub.Diagnostic {
	var diags []epub.Diagnostic

	tables := root.FindAll("table")
	for _, table := range tables {
		caption := table.FindFirst("caption")
		ariaLabel := table.Attr("aria-label")
		ariaLabelledBy := table.Attr("aria-labelledby")

		if caption == nil && ariaLabel == "" && ariaLabelledBy == "" {
			pos := epub.ByteOffsetToPosition(content, int(table.Offset))
			diags = append(diags, epub.Diagnostic{
				Code:     "table-caption",
				Severity: epub.SeverityWarning,
				Message:  "<table> missing <caption>, aria-label, or aria-labelledby",
				Source:   source,
				Range:    epub.Range{Start: pos, End: pos},
			})
		}
	}

	return diags
}

// checkFormLabels checks that form inputs have associated labels.
func checkFormLabels(content []byte, root *parser.XMLNode) []epub.Diagnostic {
	var diags []epub.Diagnostic

	// Collect all label for= values
	labelFor := make(map[string]bool)
	labels := root.FindAll("label")
	for _, label := range labels {
		if forVal := label.Attr("for"); forVal != "" {
			labelFor[forVal] = true
		}
	}

	inputs := root.FindAll("input")
	for _, input := range inputs {
		inputType := input.Attr("type")
		// Hidden, submit, button, reset, image don't need labels
		switch inputType {
		case "hidden", "submit", "button", "reset", "image":
			continue
		}

		id := input.Attr("id")
		ariaLabel := input.Attr("aria-label")
		ariaLabelledBy := input.Attr("aria-labelledby")
		title := input.Attr("title")

		hasLabel := ariaLabel != "" || ariaLabelledBy != "" || title != ""
		if !hasLabel && id != "" {
			hasLabel = labelFor[id]
		}

		if !hasLabel {
			pos := epub.ByteOffsetToPosition(content, int(input.Offset))
			diags = append(diags, epub.Diagnostic{
				Code:     "input-label",
				Severity: epub.SeverityWarning,
				Message:  "<input> missing associated label",
				Source:   source,
				Range:    epub.Range{Start: pos, End: pos},
			})
		}
	}

	// Check select and textarea too
	for _, tagName := range []string{"select", "textarea"} {
		elems := root.FindAll(tagName)
		for _, elem := range elems {
			id := elem.Attr("id")
			ariaLabel := elem.Attr("aria-label")
			ariaLabelledBy := elem.Attr("aria-labelledby")
			title := elem.Attr("title")

			hasLabel := ariaLabel != "" || ariaLabelledBy != "" || title != ""
			if !hasLabel && id != "" {
				hasLabel = labelFor[id]
			}

			if !hasLabel {
				pos := epub.ByteOffsetToPosition(content, int(elem.Offset))
				diags = append(diags, epub.Diagnostic{
					Code:     "input-label",
					Severity: epub.SeverityWarning,
					Message:  "<" + tagName + "> missing associated label",
					Source:   source,
					Range:    epub.Range{Start: pos, End: pos},
				})
			}
		}
	}

	return diags
}

// walkEpubTypes calls fn for every element with an epub:type attribute.
func walkEpubTypes(node *parser.XMLNode, fn func(node *parser.XMLNode, epubType string)) {
	for _, child := range node.Children {
		epubType := child.AttrNS(epubNS, "type")
		if epubType != "" {
			fn(child, epubType)
		}
		walkEpubTypes(child, fn)
	}
}
