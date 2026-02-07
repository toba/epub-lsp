package xhtml

import (
	"github.com/toba/epub-lsp/internal/epub"
	"github.com/toba/epub-lsp/internal/epub/parser"
)

const xhtmlNS = "http://www.w3.org/1999/xhtml"

func validateNamespaces(content []byte, root *parser.XMLNode) []epub.Diagnostic {
	var diags []epub.Diagnostic

	html := root.FindFirst("html")
	if html == nil {
		return diags
	}

	// Check XHTML namespace
	if html.Space != xhtmlNS {
		pos := epub.ByteOffsetToPosition(content, int(html.Offset))
		diags = append(diags, epub.Diagnostic{
			Code:     "HTM_049",
			Severity: epub.SeverityError,
			Message:  `missing XHTML namespace (xmlns="http://www.w3.org/1999/xhtml")`,
			Source:   source,
			Range:    epub.Range{Start: pos, End: pos},
		})
	}

	// Check lang attributes match
	xmlLang := html.AttrNS("http://www.w3.org/XML/1998/namespace", "lang")
	lang := html.Attr("lang")

	if xmlLang != "" && lang != "" && xmlLang != lang {
		pos := epub.ByteOffsetToPosition(content, int(html.Offset))
		diags = append(diags, epub.Diagnostic{
			Code:     "HTM_017",
			Severity: epub.SeverityWarning,
			Message:  `xml:lang ("` + xmlLang + `") and lang ("` + lang + `") values don't match`,
			Source:   source,
			Range:    epub.Range{Start: pos, End: pos},
		})
	}

	// Check for missing lang attribute
	if lang == "" && xmlLang == "" {
		pos := epub.ByteOffsetToPosition(content, int(html.Offset))
		diags = append(diags, epub.Diagnostic{
			Severity: epub.SeverityInfo,
			Message:  `missing lang attribute on <html> element`,
			Source:   source,
			Range:    epub.Range{Start: pos, End: pos},
		})
	}

	return diags
}
