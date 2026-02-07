package lsp

// schemaPropertyDocs maps schema.org meta properties to documentation.
var schemaPropertyDocs = map[string]string{
	"schema:accessMode": "**schema:accessMode**\n\nA human sensory perceptual system or cognitive faculty " +
		"necessary to process or perceive the content.\n\n" +
		"Valid values: `auditory`, `chartOnVisual`, `chemOnVisual`, `colorDependent`, " +
		"`diagramOnVisual`, `mathOnVisual`, `musicOnVisual`, `tactile`, `textOnVisual`, `textual`, `visual`",

	"schema:accessModeSufficient": "**schema:accessModeSufficient**\n\nA list of single or combined access modes " +
		"that are sufficient to understand all the intellectual content of a resource.\n\n" +
		"Valid values: comma-separated combinations of `auditory`, `tactile`, `textual`, `visual`",

	"schema:accessibilityFeature": "**schema:accessibilityFeature**\n\nContent features of the resource, " +
		"such as accessible media, alternatives and supported enhancements for accessibility.\n\n" +
		"Common values: `alternativeText`, `annotations`, `audioDescription`, `bookmarks`, " +
		"`braille`, `captions`, `ChemML`, `describedMath`, `displayTransformability`, " +
		"`highContrastAudio`, `highContrastDisplay`, `index`, `largePrint`, `latex`, " +
		"`longDescription`, `MathML`, `none`, `printPageNumbers`, `readingOrder`, " +
		"`rubyAnnotations`, `signLanguage`, `structuralNavigation`, `synchronizedAudioText`, " +
		"`tableOfContents`, `taggedPDF`, `tactileGraphic`, `tactileObject`, `timingControl`, " +
		"`transcript`, `ttsMarkup`, `unlocked`",

	"schema:accessibilityHazard": "**schema:accessibilityHazard**\n\nA characteristic of the described " +
		"resource that is physiologically dangerous to some users.\n\n" +
		"Valid values: `flashing`, `noFlashingHazard`, `motionSimulation`, " +
		"`noMotionSimulationHazard`, `sound`, `noSoundHazard`, `none`, `unknown`",

	"schema:accessibilitySummary": "**schema:accessibilitySummary**\n\nA human-readable summary of " +
		"specific accessibility features or deficiencies of the publication.",
}

// epubTypeDocs maps epub:type values to documentation with expected ARIA roles.
var epubTypeDocs = map[string]string{
	"toc":          "**toc** — Table of Contents\n\nExpected ARIA role: `doc-toc`\n\nA navigation list of references to the content.",
	"landmarks":    "**landmarks** — Landmarks\n\nExpected ARIA role: `directory`\n\nA list of navigation links to key structural sections.",
	"page-list":    "**page-list** — Page List\n\nExpected ARIA role: `doc-pagelist`\n\nA list of references to static page break locations.",
	"cover":        "**cover** — Cover\n\nExpected ARIA role: `doc-cover`\n\nThe cover image of the publication.",
	"titlepage":    "**titlepage** — Title Page\n\nA page at the beginning of the publication displaying the title.",
	"frontmatter":  "**frontmatter** — Front Matter\n\nPreliminary material to the main content, e.g. preface, dedication.",
	"bodymatter":   "**bodymatter** — Body Matter\n\nThe main content of the publication.",
	"backmatter":   "**backmatter** — Back Matter\n\nClosing material to the main content, e.g. appendices, glossary.",
	"chapter":      "**chapter** — Chapter\n\nExpected ARIA role: `doc-chapter`\n\nA major structural division of a piece of writing.",
	"part":         "**part** — Part\n\nExpected ARIA role: `doc-part`\n\nA major structural division of a piece of writing, larger than a chapter.",
	"footnote":     "**footnote** — Footnote\n\nExpected ARIA role: `doc-footnote`\n\nAncillary information placed at the bottom of a page.",
	"endnote":      "**endnote** — Endnote\n\nExpected ARIA role: `doc-endnote`\n\nAncillary information placed at the end of a work or section.",
	"noteref":      "**noteref** — Note Reference\n\nExpected ARIA role: `doc-noteref`\n\nA reference to a footnote or endnote.",
	"bibliography": "**bibliography** — Bibliography\n\nExpected ARIA role: `doc-bibliography`\n\nA list of works cited.",
	"glossary":     "**glossary** — Glossary\n\nExpected ARIA role: `doc-glossary`\n\nA collection of terms and their definitions.",
	"index":        "**index** — Index\n\nExpected ARIA role: `doc-index`\n\nA navigational aid with references to content.",
	"preface":      "**preface** — Preface\n\nExpected ARIA role: `doc-preface`\n\nAn introductory section preceding the main body.",
	"foreword":     "**foreword** — Foreword\n\nExpected ARIA role: `doc-foreword`\n\nAn introductory section usually by someone other than the author.",
	"appendix":     "**appendix** — Appendix\n\nExpected ARIA role: `doc-appendix`\n\nSupplementary material at the end of the main content.",
	"dedication":   "**dedication** — Dedication\n\nExpected ARIA role: `doc-dedication`\n\nAn inscription addressed to one or more persons.",
	"epigraph":     "**epigraph** — Epigraph\n\nExpected ARIA role: `doc-epigraph`\n\nA quotation set at the beginning of a work or section.",
	"abstract":     "**abstract** — Abstract\n\nExpected ARIA role: `doc-abstract`\n\nA short summary of the work.",
	"colophon":     "**colophon** — Colophon\n\nExpected ARIA role: `doc-colophon`\n\nA brief description of publishing details.",
	"pagebreak":    "**pagebreak** — Page Break\n\nExpected ARIA role: `doc-pagebreak`\n\nA location representing a page break from a static page source.",
}

// dcElementDocs maps Dublin Core element names to documentation.
var dcElementDocs = map[string]string{
	"title":       "**dc:title**\n\nThe title of the publication. Every EPUB must have at least one `dc:title`.",
	"creator":     "**dc:creator**\n\nThe name of a person or organization responsible for creating the content.",
	"language":    "**dc:language**\n\nThe language of the publication content (BCP 47 tag, e.g. `en`, `fr`). Required.",
	"identifier":  "**dc:identifier**\n\nA unique identifier for the publication (e.g. ISBN, UUID). Required.",
	"publisher":   "**dc:publisher**\n\nThe name of the entity responsible for making the publication available.",
	"date":        "**dc:date**\n\nThe date of publication in the form YYYY or YYYY-MM-DD.",
	"description": "**dc:description**\n\nA free-text description of the content of the publication.",
	"rights":      "**dc:rights**\n\nA statement about rights held over the publication.",
	"subject":     "**dc:subject**\n\nThe topic or subject of the content.",
	"contributor": "**dc:contributor**\n\nA person or organization that contributed to the content.",
	"type":        "**dc:type**\n\nThe nature or genre of the content (e.g. `dictionary`, `annotation`).",
	"format":      "**dc:format**\n\nThe file format or physical medium of the publication.",
	"source":      "**dc:source**\n\nA related resource from which the publication is derived.",
	"relation":    "**dc:relation**\n\nA related resource.",
	"coverage":    "**dc:coverage**\n\nThe spatial or temporal coverage of the content.",
}
