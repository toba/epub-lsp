package opf

import (
	"testing"

	"github.com/toba/epub-lsp/internal/epub"
)

func TestValidOPF(t *testing.T) {
	content := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<package xmlns="http://www.idpf.org/2007/opf" unique-identifier="uid" version="3.0">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
    <dc:identifier id="uid">urn:isbn:123456789</dc:identifier>
    <dc:title>Test Book</dc:title>
    <dc:language>en</dc:language>
  </metadata>
  <manifest>
    <item id="ch1" href="chapter1.xhtml" media-type="application/xhtml+xml"/>
  </manifest>
  <spine>
    <itemref idref="ch1"/>
  </spine>
</package>`)

	v := &Validator{}
	diags := v.Validate("package.opf", content, nil)

	if len(diags) != 0 {
		t.Errorf("expected no diagnostics for valid OPF, got %d:", len(diags))
		for _, d := range diags {
			t.Errorf("  [%s] %s: %s", d.Code, severityName(d.Severity), d.Message)
		}
	}
}

func TestMissingMetadataFields(t *testing.T) {
	content := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<package xmlns="http://www.idpf.org/2007/opf" unique-identifier="uid" version="3.0">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
  </metadata>
  <manifest>
    <item id="ch1" href="chapter1.xhtml" media-type="application/xhtml+xml"/>
  </manifest>
  <spine>
    <itemref idref="ch1"/>
  </spine>
</package>`)

	v := &Validator{}
	diags := v.Validate("package.opf", content, nil)

	codes := diagCodes(diags)
	expectCode(t, codes, "OPF_030") // missing dc:identifier
	expectCode(t, codes, "OPF_032") // missing dc:title
	expectCode(t, codes, "OPF_034") // missing dc:language
}

func TestUniqueIdentifierMismatch(t *testing.T) {
	content := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<package xmlns="http://www.idpf.org/2007/opf" unique-identifier="wrong" version="3.0">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
    <dc:identifier id="uid">urn:isbn:123456789</dc:identifier>
    <dc:title>Test</dc:title>
    <dc:language>en</dc:language>
  </metadata>
  <manifest/>
  <spine/>
</package>`)

	v := &Validator{}
	diags := v.Validate("package.opf", content, nil)

	codes := diagCodes(diags)
	expectCode(t, codes, "OPF_031")
}

func TestMissingSpine(t *testing.T) {
	content := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<package xmlns="http://www.idpf.org/2007/opf" unique-identifier="uid" version="3.0">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
    <dc:identifier id="uid">urn:isbn:123456789</dc:identifier>
    <dc:title>Test</dc:title>
    <dc:language>en</dc:language>
  </metadata>
  <manifest/>
</package>`)

	v := &Validator{}
	diags := v.Validate("package.opf", content, nil)

	codes := diagCodes(diags)
	expectCode(t, codes, "OPF_019")
}

func TestSpineReferencesNonexistentManifestItem(t *testing.T) {
	content := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<package xmlns="http://www.idpf.org/2007/opf" unique-identifier="uid" version="3.0">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
    <dc:identifier id="uid">urn:isbn:123456789</dc:identifier>
    <dc:title>Test</dc:title>
    <dc:language>en</dc:language>
  </metadata>
  <manifest>
    <item id="ch1" href="chapter1.xhtml" media-type="application/xhtml+xml"/>
  </manifest>
  <spine>
    <itemref idref="nonexistent"/>
  </spine>
</package>`)

	v := &Validator{}
	diags := v.Validate("package.opf", content, nil)

	codes := diagCodes(diags)
	expectCode(t, codes, "OPF_003")
}

func TestManifestWarnings(t *testing.T) {
	content := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<package xmlns="http://www.idpf.org/2007/opf" unique-identifier="uid" version="3.0">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
    <dc:identifier id="uid">urn:isbn:123456789</dc:identifier>
    <dc:title>Test</dc:title>
    <dc:language>en</dc:language>
  </metadata>
  <manifest>
    <item id="ch1" href="chapter1.xhtml"/>
    <item id="ch1" href="chapter2.xhtml" media-type="application/xhtml+xml"/>
    <item id="ch3" href="" media-type="application/xhtml+xml"/>
  </manifest>
  <spine>
    <itemref idref="ch1"/>
  </spine>
</package>`)

	v := &Validator{}
	diags := v.Validate("package.opf", content, nil)

	hasWarning := func(msg string) bool {
		for _, d := range diags {
			if d.Severity == epub.SeverityWarning && d.Message == msg {
				return true
			}
		}
		return false
	}

	if !hasWarning("manifest item missing media-type attribute") {
		t.Error("expected warning for missing media-type")
	}
	if !hasWarning(`duplicate manifest item id: "ch1"`) {
		t.Error("expected warning for duplicate id")
	}
	if !hasWarning("manifest item href is empty") {
		t.Error("expected warning for empty href")
	}
}

func TestMalformedXML(t *testing.T) {
	content := []byte(`<package><unclosed>`)

	v := &Validator{}
	diags := v.Validate("package.opf", content, nil)

	if len(diags) == 0 {
		t.Error("expected diagnostics for malformed XML")
	}
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

func severityName(s int) string {
	switch s {
	case epub.SeverityError:
		return "Error"
	case epub.SeverityWarning:
		return "Warning"
	case epub.SeverityInfo:
		return "Info"
	default:
		return "Unknown"
	}
}
