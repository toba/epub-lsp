// Package validator provides the EPUB validation framework.
package validator

import (
	"slices"

	"github.com/toba/epub-lsp/internal/epub"
)

// Validator validates EPUB source files of specific types.
type Validator interface {
	FileTypes() []epub.FileType
	Validate(uri string, content []byte, ctx *WorkspaceContext) []epub.Diagnostic
}

// ManifestItem represents a single item in the OPF manifest.
type ManifestItem struct {
	ID        string
	Href      string
	MediaType string
}

// SpineItem represents a single itemref in the OPF spine.
type SpineItem struct {
	IDRef  string
	Linear bool
}

// MetadataInfo holds parsed OPF metadata relevant to accessibility validation.
type MetadataInfo struct {
	// AccessModes lists schema:accessMode values.
	AccessModes []string
	// AccessModeSufficient lists schema:accessModeSufficient values.
	AccessModeSufficient []string
	// AccessibilityFeatures lists schema:accessibilityFeature values.
	AccessibilityFeatures []string
	// AccessibilityHazards lists schema:accessibilityHazard values.
	AccessibilityHazards []string
	// AccessibilitySummary is the schema:accessibilitySummary value.
	AccessibilitySummary string
	// HasDCSource is true if a dc:source element exists.
	HasDCSource bool
	// HasTitle is true if dc:title exists.
	HasTitle bool
	// HasLanguage is true if dc:language exists.
	HasLanguage bool
}

// ManifestInfo holds parsed OPF manifest, spine, and metadata.
type ManifestInfo struct {
	Items    []ManifestItem
	Spine    []SpineItem
	Metadata MetadataInfo
}

// WorkspaceContext provides cross-file information for validators.
type WorkspaceContext struct {
	RootPath  string
	Files     map[string][]byte
	FileTypes map[string]epub.FileType
	Manifest  *ManifestInfo
	// AccessibilitySeverity controls accessibility diagnostic severity.
	// 0 = ignore (skip checks), 1 = error, 2 = warning (default).
	AccessibilitySeverity int
}

// Registry holds all registered validators and dispatches validation.
type Registry struct {
	validators []Validator
}

// NewRegistry creates a new validator registry.
func NewRegistry() *Registry {
	return &Registry{}
}

// Register adds a validator to the registry.
func (r *Registry) Register(v Validator) {
	r.validators = append(r.validators, v)
}

// ValidateFile runs all validators that match the given file type.
func (r *Registry) ValidateFile(
	uri string,
	content []byte,
	fileType epub.FileType,
	ctx *WorkspaceContext,
) []epub.Diagnostic {
	var diags []epub.Diagnostic

	for _, v := range r.validators {
		if slices.Contains(v.FileTypes(), fileType) {
			diags = append(diags, v.Validate(uri, content, ctx)...)
		}
	}

	return diags
}
