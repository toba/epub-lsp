package accessibility

import (
	"testing"

	"github.com/toba/epub-lsp/internal/epub"
)

func TestMetadataValidator_FullyAccessible(t *testing.T) {
	content := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<package xmlns="http://www.idpf.org/2007/opf" unique-identifier="uid" version="3.0">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
    <dc:identifier id="uid">urn:isbn:123</dc:identifier>
    <dc:title>Test</dc:title>
    <dc:language>en</dc:language>
    <meta property="schema:accessMode">textual</meta>
    <meta property="schema:accessModeSufficient">textual</meta>
    <meta property="schema:accessibilityFeature">structuralNavigation</meta>
    <meta property="schema:accessibilityHazard">none</meta>
    <meta property="schema:accessibilitySummary">Accessible</meta>
  </metadata>
  <manifest/>
  <spine/>
</package>`)

	v := &MetadataValidator{}
	diags := v.Validate("package.opf", content, nil)

	for _, d := range diags {
		if d.Severity <= epub.SeverityWarning {
			t.Errorf("unexpected error/warning: [%s] %s", d.Code, d.Message)
		}
	}
}

func TestMetadataValidator_AllMissing(t *testing.T) {
	content := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<package xmlns="http://www.idpf.org/2007/opf" unique-identifier="uid" version="3.0">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
    <dc:identifier id="uid">urn:isbn:123</dc:identifier>
    <dc:title>Test</dc:title>
    <dc:language>en</dc:language>
  </metadata>
  <manifest/>
  <spine/>
</package>`)

	v := &MetadataValidator{}
	diags := v.Validate("package.opf", content, nil)

	codes := diagCodes(diags)
	expectCode(t, codes, "metadata-accessmode")
	expectCode(t, codes, "metadata-accessibilityfeature")
	expectCode(t, codes, "metadata-accessibilityhazard")
	expectCode(t, codes, "metadata-accessibilitysummary")
	expectCode(t, codes, "metadata-accessmodesufficient")
}

func TestMetadataValidator_InvalidAccessMode(t *testing.T) {
	content := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<package xmlns="http://www.idpf.org/2007/opf" unique-identifier="uid" version="3.0">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
    <dc:identifier id="uid">urn:isbn:123</dc:identifier>
    <dc:title>Test</dc:title>
    <dc:language>en</dc:language>
    <meta property="schema:accessMode">invalid_mode</meta>
    <meta property="schema:accessModeSufficient">textual</meta>
    <meta property="schema:accessibilityFeature">structuralNavigation</meta>
    <meta property="schema:accessibilityHazard">none</meta>
    <meta property="schema:accessibilitySummary">Test</meta>
  </metadata>
  <manifest/>
  <spine/>
</package>`)

	v := &MetadataValidator{}
	diags := v.Validate("package.opf", content, nil)

	expectCode(t, diagCodes(diags), "metadata-accessmode-invalid")
}

func TestMetadataValidator_InvalidFeature(t *testing.T) {
	content := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<package xmlns="http://www.idpf.org/2007/opf" unique-identifier="uid" version="3.0">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
    <dc:identifier id="uid">urn:isbn:123</dc:identifier>
    <dc:title>Test</dc:title>
    <dc:language>en</dc:language>
    <meta property="schema:accessMode">textual</meta>
    <meta property="schema:accessModeSufficient">textual</meta>
    <meta property="schema:accessibilityFeature">madeUpFeature</meta>
    <meta property="schema:accessibilityHazard">none</meta>
  </metadata>
  <manifest/>
  <spine/>
</package>`)

	v := &MetadataValidator{}
	diags := v.Validate("package.opf", content, nil)

	expectCode(t, diagCodes(diags), "metadata-accessibilityfeature-invalid")
}

func TestMetadataValidator_ContradictoryHazards(t *testing.T) {
	content := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<package xmlns="http://www.idpf.org/2007/opf" unique-identifier="uid" version="3.0">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
    <dc:identifier id="uid">urn:isbn:123</dc:identifier>
    <dc:title>Test</dc:title>
    <dc:language>en</dc:language>
    <meta property="schema:accessMode">textual</meta>
    <meta property="schema:accessModeSufficient">textual</meta>
    <meta property="schema:accessibilityFeature">structuralNavigation</meta>
    <meta property="schema:accessibilityHazard">noFlashingHazard</meta>
    <meta property="schema:accessibilityHazard">flashing</meta>
  </metadata>
  <manifest/>
  <spine/>
</package>`)

	v := &MetadataValidator{}
	diags := v.Validate("package.opf", content, nil)

	// Should have a contradictory hazard error
	hasContradiction := false
	for _, d := range diags {
		if d.Code == "metadata-accessibilityhazard-invalid" &&
			d.Message == `contradictory hazard values: "noFlashingHazard" and "flashing"` {
			hasContradiction = true
			break
		}
	}
	if !hasContradiction {
		t.Error("expected contradictory hazard diagnostic")
	}
}

func TestMetadataValidator_InvalidHazard(t *testing.T) {
	content := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<package xmlns="http://www.idpf.org/2007/opf" unique-identifier="uid" version="3.0">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
    <dc:identifier id="uid">urn:isbn:123</dc:identifier>
    <dc:title>Test</dc:title>
    <dc:language>en</dc:language>
    <meta property="schema:accessMode">textual</meta>
    <meta property="schema:accessModeSufficient">textual</meta>
    <meta property="schema:accessibilityFeature">structuralNavigation</meta>
    <meta property="schema:accessibilityHazard">madeUpHazard</meta>
  </metadata>
  <manifest/>
  <spine/>
</package>`)

	v := &MetadataValidator{}
	diags := v.Validate("package.opf", content, nil)

	expectCode(t, diagCodes(diags), "metadata-accessibilityhazard-invalid")
}

// helpers

func diagCodes(diags []epub.Diagnostic) map[string]bool {
	codes := make(map[string]bool)
	for _, d := range diags {
		if d.Code != "" {
			codes[d.Code] = true
		}
	}
	return codes
}

func expectCode(t *testing.T, codes map[string]bool, code string) {
	t.Helper()
	if !codes[code] {
		t.Errorf("expected diagnostic code %s", code)
	}
}
