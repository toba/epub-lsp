package xhtml

import (
	"testing"

	"github.com/toba/epub-lsp/internal/epub"
)

func TestValidXHTML(t *testing.T) {
	content := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<html xmlns="http://www.w3.org/1999/xhtml" lang="en" xml:lang="en">
<head><title>Test</title></head>
<body>
  <p>Hello</p>
  <img src="cover.jpg" alt="Cover"/>
</body>
</html>`)

	v := &Validator{}
	diags := v.Validate("chapter.xhtml", content, nil)

	if len(diags) != 0 {
		t.Errorf("expected no diagnostics for valid XHTML, got %d:", len(diags))
		for _, d := range diags {
			t.Errorf("  [%s] %s: %s", d.Code, severityName(d.Severity), d.Message)
		}
	}
}

func TestMissingXHTMLNamespace(t *testing.T) {
	content := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<html lang="en">
<head><title>Test</title></head>
<body><p>Hello</p></body>
</html>`)

	v := &Validator{}
	diags := v.Validate("chapter.xhtml", content, nil)

	if !hasCode(diags, "HTM_049") {
		t.Error("expected HTM_049 for missing XHTML namespace")
	}
}

func TestLangMismatch(t *testing.T) {
	content := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<html xmlns="http://www.w3.org/1999/xhtml" lang="en" xml:lang="fr">
<head><title>Test</title></head>
<body><p>Hello</p></body>
</html>`)

	v := &Validator{}
	diags := v.Validate("chapter.xhtml", content, nil)

	if !hasCode(diags, "HTM_017") {
		t.Error("expected HTM_017 for lang mismatch")
	}
}

func TestMissingLang(t *testing.T) {
	content := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<html xmlns="http://www.w3.org/1999/xhtml">
<head><title>Test</title></head>
<body><p>Hello</p></body>
</html>`)

	v := &Validator{}
	diags := v.Validate("chapter.xhtml", content, nil)

	hasMissingLang := false
	for _, d := range diags {
		if d.Severity == epub.SeverityInfo &&
			d.Message == "missing lang attribute on <html> element" {
			hasMissingLang = true
			break
		}
	}
	if !hasMissingLang {
		t.Error("expected info diagnostic for missing lang attribute")
	}
}

func TestMissingImgAlt(t *testing.T) {
	content := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<html xmlns="http://www.w3.org/1999/xhtml" lang="en">
<head><title>Test</title></head>
<body>
  <img src="photo.jpg"/>
  <img src="icon.png" alt="Icon"/>
  <img src="banner.jpg"/>
</body>
</html>`)

	v := &Validator{}
	diags := v.Validate("chapter.xhtml", content, nil)

	count := 0
	for _, d := range diags {
		if d.Code == "HTM_008" {
			count++
		}
	}
	if count != 2 {
		t.Errorf("expected 2 HTM_008 diagnostics, got %d", count)
	}
}

func TestMalformedXHTML(t *testing.T) {
	content := []byte(`<html><body><p>unclosed`)

	v := &Validator{}
	diags := v.Validate("chapter.xhtml", content, nil)

	if len(diags) == 0 {
		t.Error("expected diagnostics for malformed XHTML")
	}
}

// helpers

func hasCode(diags []epub.Diagnostic, code string) bool {
	for _, d := range diags {
		if d.Code == code {
			return true
		}
	}
	return false
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
