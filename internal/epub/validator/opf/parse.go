package opf

import (
	"strings"

	"github.com/toba/epub-lsp/internal/epub/parser"
	"github.com/toba/epub-lsp/internal/epub/validator"
)

// ParseManifest parses an OPF document's manifest, spine, and metadata into ManifestInfo.
// Returns nil if the content cannot be parsed or has no package element.
func ParseManifest(content []byte) *validator.ManifestInfo {
	root, diags := parser.Parse(content)
	if len(diags) > 0 {
		return nil
	}

	pkg := root.FindFirst("package")
	if pkg == nil {
		return nil
	}

	info := &validator.ManifestInfo{}

	// Parse manifest items
	manifest := pkg.FindFirst("manifest")
	if manifest != nil {
		for _, item := range manifest.Children {
			if item.Local != "item" {
				continue
			}
			info.Items = append(info.Items, validator.ManifestItem{
				ID:        item.Attr("id"),
				Href:      item.Attr("href"),
				MediaType: item.Attr("media-type"),
			})
		}
	}

	// Parse spine
	spine := pkg.FindFirst("spine")
	if spine != nil {
		for _, itemref := range spine.Children {
			if itemref.Local != "itemref" {
				continue
			}
			linear := itemref.Attr("linear")
			info.Spine = append(info.Spine, validator.SpineItem{
				IDRef:  itemref.Attr("idref"),
				Linear: linear != "no",
			})
		}
	}

	// Parse metadata
	metadata := pkg.FindFirst("metadata")
	if metadata != nil {
		parseMetadataInfo(metadata, &info.Metadata)
	}

	return info
}

func parseMetadataInfo(metadata *parser.XMLNode, meta *validator.MetadataInfo) {
	// Check dc:title and dc:language
	if len(metadata.FindAllNS(dcNS, "title")) > 0 {
		meta.HasTitle = true
	}
	if len(metadata.FindAllNS(dcNS, "language")) > 0 {
		meta.HasLanguage = true
	}
	if len(metadata.FindAllNS(dcNS, "source")) > 0 {
		meta.HasDCSource = true
	}

	// Parse <meta> elements for schema.org accessibility properties
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
			meta.AccessModes = append(meta.AccessModes, value)
		case "schema:accessModeSufficient":
			meta.AccessModeSufficient = append(meta.AccessModeSufficient, value)
		case "schema:accessibilityFeature":
			meta.AccessibilityFeatures = append(meta.AccessibilityFeatures, value)
		case "schema:accessibilityHazard":
			meta.AccessibilityHazards = append(meta.AccessibilityHazards, value)
		case "schema:accessibilitySummary":
			meta.AccessibilitySummary = value
		}
	}
}
