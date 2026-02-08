// Package accessibility validates EPUB accessibility metadata and structure
// per the Ace (daisy/ace) rule set.
package accessibility

import (
	"slices"
	"strings"

	"github.com/toba/epub-lsp/internal/epub"
	"github.com/toba/epub-lsp/internal/epub/parser"
	"github.com/toba/epub-lsp/internal/epub/validator"
	"github.com/toba/epub-lsp/internal/epub/validator/opf"
)

const source = "epub-accessibility"

// validAccessModes lists the valid schema:accessMode values.
var validAccessModes = []string{
	"auditory", "chartOnVisual", "chemOnVisual", "colorDependent",
	"diagramOnVisual", "mathOnVisual", "musicOnVisual", "tactile",
	"textOnVisual", "textual", "visual",
}

// validAccessibilityFeatures lists the valid schema:accessibilityFeature values.
var validAccessibilityFeatures = []string{
	"alternativeText", "annotations", "audioDescription", "bookmarks",
	"braille", "captions", "ChemML", "describedMath", "displayTransformability",
	"displayTransformability/font-size", "displayTransformability/font-family",
	"displayTransformability/line-height", "displayTransformability/word-spacing",
	"displayTransformability/letter-spacing", "displayTransformability/color",
	"displayTransformability/background-color",
	"highContrastAudio", "highContrastDisplay", "index", "largePrint",
	"latex", "longDescription", "MathML", "none", "printPageNumbers",
	"readingOrder", "rubyAnnotations", "signLanguage",
	"structuralNavigation", "synchronizedAudioText", "tableOfContents",
	"taggedPDF", "tactileGraphic", "tactileObject", "timingControl",
	"transcript", "ttsMarkup", "unlocked",
	"ARIA", "fullRubyAnnotations",
	"pageBreakMarkers", "pageNavigation",
}

// validAccessibilityHazards lists the valid schema:accessibilityHazard values.
var validAccessibilityHazards = []string{
	"flashing", "noFlashingHazard",
	"motionSimulation", "noMotionSimulationHazard",
	"sound", "noSoundHazard",
	"none", "unknown",
}

// hazardContradictions maps each "no" hazard to its contradicting hazard.
var hazardContradictions = map[string]string{
	"noFlashingHazard":         "flashing",
	"noMotionSimulationHazard": "motionSimulation",
	"noSoundHazard":            "sound",
}

// MetadataValidator checks OPF accessibility metadata.
type MetadataValidator struct{}

func (v *MetadataValidator) FileTypes() []epub.FileType {
	return []epub.FileType{epub.FileTypeOPF}
}

func (v *MetadataValidator) Validate(
	_ string,
	content []byte,
	ctx *validator.WorkspaceContext,
) []epub.Diagnostic {
	if ctx != nil && ctx.AccessibilitySeverity == 0 {
		return nil
	}
	_, metadata := opf.ParseOPFMetadata(content)
	if metadata == nil {
		return nil
	}
	diags := validateAccessibilityMetadata(content, metadata)
	if ctx != nil && ctx.AccessibilitySeverity != 0 {
		for i := range diags {
			diags[i].Severity = ctx.AccessibilitySeverity
		}
	}
	return diags
}

func validateAccessibilityMetadata(
	content []byte,
	metadata *parser.XMLNode,
) []epub.Diagnostic {
	var diags []epub.Diagnostic

	pos := epub.ByteOffsetToPosition(content, int(metadata.Offset))
	rng := epub.Range{Start: pos, End: pos}

	// Collect metadata values
	var (
		accessModes          []string
		accessModeSufficient []string
		features             []string
		hazards              []string
		hasSummary           bool
	)

	for _, child := range metadata.Children {
		if child.Local != "meta" {
			continue
		}
		property := child.Attr("property")
		if property == "" {
			continue
		}
		value := strings.TrimSpace(child.CharData)
		switch property {
		case "schema:accessMode":
			accessModes = append(accessModes, value)
			if !slices.Contains(validAccessModes, value) {
				diags = append(diags, epub.NewDiag(content, int(child.Offset), source).
					Code("metadata-accessmode-invalid").
					Error("invalid access mode value: \""+value+"\"").Build())
			}

		case "schema:accessModeSufficient":
			accessModeSufficient = append(accessModeSufficient, value)

		case "schema:accessibilityFeature":
			features = append(features, value)
			if !slices.Contains(validAccessibilityFeatures, value) {
				diags = append(diags, epub.NewDiag(content, int(child.Offset), source).
					Code("metadata-accessibilityfeature-invalid").
					Error("invalid accessibility feature value: \""+value+"\"").Build())
			}

		case "schema:accessibilityHazard":
			hazards = append(hazards, value)
			if !slices.Contains(validAccessibilityHazards, value) {
				diags = append(diags, epub.NewDiag(content, int(child.Offset), source).
					Code("metadata-accessibilityhazard-invalid").
					Error("invalid accessibility hazard value: \""+value+"\"").Build())
			}

		case "schema:accessibilitySummary":
			hasSummary = true
		}
	}

	// Check for missing metadata
	if len(accessModes) == 0 {
		diags = append(diags, epub.Diagnostic{
			Code:     "metadata-accessmode",
			Severity: epub.SeverityWarning,
			Message:  "missing schema:accessMode metadata",
			Source:   source,
			Range:    rng,
		})
	}

	if len(features) == 0 {
		diags = append(diags, epub.Diagnostic{
			Code:     "metadata-accessibilityfeature",
			Severity: epub.SeverityWarning,
			Message:  "missing schema:accessibilityFeature metadata",
			Source:   source,
			Range:    rng,
		})
	}

	if len(hazards) == 0 {
		diags = append(diags, epub.Diagnostic{
			Code:     "metadata-accessibilityhazard",
			Severity: epub.SeverityWarning,
			Message:  "missing schema:accessibilityHazard metadata",
			Source:   source,
			Range:    rng,
		})
	}

	if !hasSummary {
		diags = append(diags, epub.Diagnostic{
			Code:     "metadata-accessibilitysummary",
			Severity: epub.SeverityInfo,
			Message:  "missing schema:accessibilitySummary metadata",
			Source:   source,
			Range:    rng,
		})
	}

	if len(accessModeSufficient) == 0 {
		diags = append(diags, epub.Diagnostic{
			Code:     "metadata-accessmodesufficient",
			Severity: epub.SeverityWarning,
			Message:  "missing schema:accessModeSufficient metadata",
			Source:   source,
			Range:    rng,
		})
	}

	// Check for contradictory hazards
	hazardSet := make(map[string]bool)
	for _, h := range hazards {
		hazardSet[h] = true
	}

	for noHazard, hazard := range hazardContradictions {
		if hazardSet[noHazard] && hazardSet[hazard] {
			diags = append(diags, epub.Diagnostic{
				Code:     "metadata-accessibilityhazard-invalid",
				Severity: epub.SeverityError,
				Message:  "contradictory hazard values: \"" + noHazard + "\" and \"" + hazard + "\"",
				Source:   source,
				Range:    rng,
			})
		}
	}

	return diags
}
