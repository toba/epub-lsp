// Package css validates CSS stylesheets for EPUB compliance.
package css

import (
	"strings"

	"github.com/toba/epub-lsp/internal/epub"
	"github.com/toba/epub-lsp/internal/epub/parser"
	"github.com/toba/epub-lsp/internal/epub/validator"
)

const source = "epub-css"

// allowedFontFormats lists the standard font formats for EPUB.
var allowedFontFormats = map[string]bool{
	"woff":              true,
	"woff2":             true,
	"opentype":          true,
	"truetype":          true,
	"embedded-opentype": false, // not standard for EPUB
}

// allowedFontExtensions lists standard font file extensions for EPUB.
var allowedFontExtensions = map[string]bool{
	".woff":  true,
	".woff2": true,
	".otf":   true,
	".ttf":   true,
}

// Validator validates CSS stylesheets.
type Validator struct{}

func (v *Validator) FileTypes() []epub.FileType {
	return []epub.FileType{epub.FileTypeCSS}
}

func (v *Validator) Validate(
	_ string,
	content []byte,
	_ *validator.WorkspaceContext,
) []epub.Diagnostic {
	props, atRules, diags := parser.ScanCSS(content)

	// Check properties
	for _, prop := range props {
		pos := epub.Position{Line: prop.Line, Character: prop.Col}
		rng := epub.Range{Start: pos, End: pos}

		switch prop.Property {
		case "direction", "unicode-bidi":
			diags = append(diags, epub.Diagnostic{
				Code:     "CSS_001",
				Severity: epub.SeverityError,
				Message:  "CSS property \"" + prop.Property + "\" must not be used in EPUB content documents",
				Source:   source,
				Range:    rng,
			})

		case "position":
			val := strings.TrimSpace(prop.Value)
			switch val {
			case "fixed":
				diags = append(diags, epub.Diagnostic{
					Code:     "CSS_006",
					Severity: epub.SeverityWarning,
					Message:  "position: fixed is not well supported in EPUB reading systems",
					Source:   source,
					Range:    rng,
				})
			case "absolute":
				diags = append(diags, epub.Diagnostic{
					Code:     "CSS_017",
					Severity: epub.SeverityWarning,
					Message:  "position: absolute may not be well supported in EPUB reading systems",
					Source:   source,
					Range:    rng,
				})
			}
		}
	}

	// Check @font-face for non-standard font types
	for _, atRule := range atRules {
		if atRule.Name != "@font-face" {
			continue
		}

		// Look for src properties that follow this @font-face
		for _, prop := range props {
			if prop.Property != "src" || prop.Offset < atRule.Offset {
				continue
			}
			checkFontSrc(prop, &diags)
		}
	}

	return diags
}

func checkFontSrc(prop parser.CSSPropertyDecl, diags *[]epub.Diagnostic) {
	val := strings.ToLower(prop.Value)
	pos := epub.Position{Line: prop.Line, Character: prop.Col}
	rng := epub.Range{Start: pos, End: pos}

	// Check format() hints
	if idx := strings.Index(val, "format("); idx >= 0 {
		end := strings.Index(val[idx:], ")")
		if end > 0 {
			format := val[idx+7 : idx+end]
			format = strings.Trim(format, `"' `)
			if _, known := allowedFontFormats[format]; !known {
				*diags = append(*diags, epub.Diagnostic{
					Code:     "CSS_007",
					Severity: epub.SeverityWarning,
					Message:  "non-standard font format: \"" + format + "\"",
					Source:   source,
					Range:    rng,
				})
			}
		}
		return
	}

	// Check URL extension if no format() hint
	if idx := strings.Index(val, "url("); idx >= 0 {
		end := strings.Index(val[idx:], ")")
		if end > 0 {
			urlVal := val[idx+4 : idx+end]
			urlVal = strings.Trim(urlVal, `"' `)
			hasValidExt := false
			for ext := range allowedFontExtensions {
				if strings.HasSuffix(urlVal, ext) {
					hasValidExt = true
					break
				}
			}
			if !hasValidExt && urlVal != "" {
				*diags = append(*diags, epub.Diagnostic{
					Code:     "CSS_007",
					Severity: epub.SeverityWarning,
					Message:  "non-standard font type in @font-face src",
					Source:   source,
					Range:    rng,
				})
			}
		}
	}
}
