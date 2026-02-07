// Package opf validates OPF package documents.
package opf

import (
	"github.com/toba/epub-lsp/internal/epub"
	"github.com/toba/epub-lsp/internal/epub/parser"
	"github.com/toba/epub-lsp/internal/epub/validator"
)

const source = "epub-opf"

// Validator validates OPF package documents.
type Validator struct{}

func (v *Validator) FileTypes() []epub.FileType {
	return []epub.FileType{epub.FileTypeOPF}
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

	pkg := root.FindFirst("package")
	if pkg == nil {
		return diags
	}

	diags = append(diags, validateMetadata(content, pkg)...)
	diags = append(diags, validateManifest(content, pkg)...)
	diags = append(diags, validateSpine(content, pkg)...)

	return diags
}
