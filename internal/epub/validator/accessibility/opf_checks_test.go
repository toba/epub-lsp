package accessibility

import (
	"testing"
)

func TestOPFAccessibility_MissingTitle(t *testing.T) {
	content := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<package xmlns="http://www.idpf.org/2007/opf" unique-identifier="uid" version="3.0">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
    <dc:identifier id="uid">urn:isbn:123</dc:identifier>
    <dc:language>en</dc:language>
  </metadata>
  <manifest/>
  <spine/>
</package>`)

	v := &OPFAccessibilityValidator{}
	diags := v.Validate("package.opf", content, nil)

	expectCode(t, diagCodes(diags), "epub-title")
}

func TestOPFAccessibility_MissingLanguage(t *testing.T) {
	content := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<package xmlns="http://www.idpf.org/2007/opf" unique-identifier="uid" version="3.0">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
    <dc:identifier id="uid">urn:isbn:123</dc:identifier>
    <dc:title>Test</dc:title>
  </metadata>
  <manifest/>
  <spine/>
</package>`)

	v := &OPFAccessibilityValidator{}
	diags := v.Validate("package.opf", content, nil)

	expectCode(t, diagCodes(diags), "epub-lang")
}

func TestOPFAccessibility_Complete(t *testing.T) {
	content := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<package xmlns="http://www.idpf.org/2007/opf" unique-identifier="uid" version="3.0">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
    <dc:identifier id="uid">urn:isbn:123</dc:identifier>
    <dc:title>Test Book</dc:title>
    <dc:language>en</dc:language>
  </metadata>
  <manifest/>
  <spine/>
</package>`)

	v := &OPFAccessibilityValidator{}
	diags := v.Validate("package.opf", content, nil)

	if len(diags) != 0 {
		t.Errorf("expected no diagnostics, got %d:", len(diags))
		for _, d := range diags {
			t.Errorf("  [%s] %s", d.Code, d.Message)
		}
	}
}
