package xhtml

import (
	"github.com/toba/epub-lsp/internal/epub"
	"github.com/toba/epub-lsp/internal/epub/parser"
)

func validateNamespaces(content []byte, root *parser.XMLNode) []epub.Diagnostic {
	var diags []epub.Diagnostic

	html := root.FindFirst("html")
	if html == nil {
		return diags
	}

	// Check XHTML namespace
	if html.Space != epub.NSXHTML {
		diags = append(diags, epub.NewDiag(content, int(html.Offset), source).
			Code("HTM_049").
			Error(`missing XHTML namespace (xmlns="http://www.w3.org/1999/xhtml")`).
			Build())
	}

	// Check lang attributes match
	xmlLang := html.AttrNS(epub.NSXML, "lang")
	lang := html.Attr("lang")

	if xmlLang != "" && lang != "" && xmlLang != lang {
		diags = append(diags, epub.NewDiag(content, int(html.Offset), source).
			Code("HTM_017").
			Warning(`xml:lang ("`+xmlLang+`") and lang ("`+lang+`") values don't match`).
			Build())
	}

	// Check for missing lang attribute
	if lang == "" && xmlLang == "" {
		diags = append(diags, epub.NewDiag(content, int(html.Offset), source).
			Info(`missing lang attribute on <html> element`).Build())
	}

	return diags
}
