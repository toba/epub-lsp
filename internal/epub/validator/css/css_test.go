package css

import (
	"testing"

	"github.com/toba/epub-lsp/internal/epub"
)

func TestValidCSS(t *testing.T) {
	content := []byte(`
body {
  margin: 0;
  padding: 10px;
  font-family: serif;
}
`)

	v := &Validator{}
	diags := v.Validate("style.css", content, nil)

	if len(diags) != 0 {
		t.Errorf("expected no diagnostics for valid CSS, got %d:", len(diags))
		for _, d := range diags {
			t.Errorf("  [%s] %s", d.Code, d.Message)
		}
	}
}

func TestDirectionProperty(t *testing.T) {
	content := []byte(`
p {
  direction: rtl;
}
`)

	v := &Validator{}
	diags := v.Validate("style.css", content, nil)

	if !hasCode(diags, "CSS_001") {
		t.Error("expected CSS_001 for direction property")
	}
}

func TestUnicodeBidiProperty(t *testing.T) {
	content := []byte(`
span {
  unicode-bidi: embed;
}
`)

	v := &Validator{}
	diags := v.Validate("style.css", content, nil)

	if !hasCode(diags, "CSS_001") {
		t.Error("expected CSS_001 for unicode-bidi property")
	}
}

func TestPositionFixed(t *testing.T) {
	content := []byte(`
.sidebar {
  position: fixed;
}
`)

	v := &Validator{}
	diags := v.Validate("style.css", content, nil)

	if !hasCode(diags, "CSS_006") {
		t.Error("expected CSS_006 for position: fixed")
	}
}

func TestPositionAbsolute(t *testing.T) {
	content := []byte(`
.overlay {
  position: absolute;
}
`)

	v := &Validator{}
	diags := v.Validate("style.css", content, nil)

	if !hasCode(diags, "CSS_017") {
		t.Error("expected CSS_017 for position: absolute")
	}
}

func TestNonStandardFontFormat(t *testing.T) {
	content := []byte(`
@font-face {
  font-family: "MyFont";
  src: url("font.svg") format("svg");
}
`)

	v := &Validator{}
	diags := v.Validate("style.css", content, nil)

	if !hasCode(diags, "CSS_007") {
		t.Error("expected CSS_007 for non-standard font format")
	}
}

func TestStandardFontFormat(t *testing.T) {
	content := []byte(`
@font-face {
  font-family: "MyFont";
  src: url("font.woff2") format("woff2");
}
`)

	v := &Validator{}
	diags := v.Validate("style.css", content, nil)

	for _, d := range diags {
		if d.Code == "CSS_007" {
			t.Error("unexpected CSS_007 for standard woff2 font")
		}
	}
}

func TestInvalidUTF8CSS(t *testing.T) {
	content := []byte{0xff, 0xfe, 0x62, 0x6f, 0x64, 0x79, 0x20, 0x7b, 0x7d}

	v := &Validator{}
	diags := v.Validate("style.css", content, nil)

	if !hasCode(diags, "CSS_003") {
		t.Error("expected CSS_003 for invalid UTF-8")
	}
}

func TestUnclosedBrace(t *testing.T) {
	content := []byte(`body { margin: 0;`)

	v := &Validator{}
	diags := v.Validate("style.css", content, nil)

	if !hasCode(diags, "CSS_008") {
		t.Error("expected CSS_008 for unclosed brace")
	}
}

func TestMultipleIssues(t *testing.T) {
	content := []byte(`
.box {
  direction: ltr;
  position: fixed;
  unicode-bidi: isolate;
}
`)

	v := &Validator{}
	diags := v.Validate("style.css", content, nil)

	css001Count := 0
	css006Count := 0
	for _, d := range diags {
		switch d.Code {
		case "CSS_001":
			css001Count++
		case "CSS_006":
			css006Count++
		}
	}

	if css001Count != 2 {
		t.Errorf("expected 2 CSS_001 diagnostics, got %d", css001Count)
	}
	if css006Count != 1 {
		t.Errorf("expected 1 CSS_006 diagnostic, got %d", css006Count)
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
