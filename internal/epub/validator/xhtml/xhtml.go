// Package xhtml validates XHTML content documents.
package xhtml

import (
	"github.com/toba/epub-lsp/internal/epub"
	"github.com/toba/epub-lsp/internal/epub/parser"
	"github.com/toba/epub-lsp/internal/epub/validator"
)

const source = "epub-xhtml"

// Validator validates XHTML content documents.
type Validator struct{}

func (v *Validator) FileTypes() []epub.FileType {
	return []epub.FileType{epub.FileTypeXHTML, epub.FileTypeNav}
}

func (v *Validator) Validate(
	_ string,
	content []byte,
	_ *validator.WorkspaceContext,
) []epub.Diagnostic {
	root, diags := parser.Parse(content)
	if len(diags) > 0 {
		return diags
	}

	diags = append(diags, validateNamespaces(content, root)...)
	diags = append(diags, validateStructure(content, root)...)

	return diags
}
