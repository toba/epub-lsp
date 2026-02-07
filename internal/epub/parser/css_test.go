package parser

import (
	"testing"
)

func TestScanCSS_ValidCSS(t *testing.T) {
	content := []byte(`
body {
  margin: 0;
  padding: 10px;
  font-family: serif;
}

h1 {
  color: #333;
}
`)

	props, _, diags := ScanCSS(content)
	if len(diags) != 0 {
		t.Errorf("expected no diagnostics, got %d:", len(diags))
		for _, d := range diags {
			t.Errorf("  [%s] %s", d.Code, d.Message)
		}
	}

	if len(props) != 4 {
		t.Errorf("expected 4 property declarations, got %d", len(props))
	}
}

func TestScanCSS_UnclosedBrace(t *testing.T) {
	content := []byte(`body { margin: 0;`)

	_, _, diags := ScanCSS(content)
	hasError := false
	for _, d := range diags {
		if d.Code == "CSS_008" {
			hasError = true
			break
		}
	}
	if !hasError {
		t.Error("expected CSS_008 for unclosed brace")
	}
}

func TestScanCSS_UnclosedComment(t *testing.T) {
	content := []byte(`/* unclosed comment
body { margin: 0; }`)

	_, _, diags := ScanCSS(content)
	hasError := false
	for _, d := range diags {
		if d.Code == "CSS_008" {
			hasError = true
			break
		}
	}
	if !hasError {
		t.Error("expected CSS_008 for unclosed comment")
	}
}

func TestScanCSS_ExtraBraceClose(t *testing.T) {
	content := []byte(`body { margin: 0; }}`)

	_, _, diags := ScanCSS(content)
	hasError := false
	for _, d := range diags {
		if d.Code == "CSS_008" {
			hasError = true
			break
		}
	}
	if !hasError {
		t.Error("expected CSS_008 for extra closing brace")
	}
}

func TestScanCSS_InvalidUTF8(t *testing.T) {
	content := []byte{0xff, 0xfe, 0x62, 0x6f, 0x64, 0x79}

	_, _, diags := ScanCSS(content)
	hasError := false
	for _, d := range diags {
		if d.Code == "CSS_003" {
			hasError = true
			break
		}
	}
	if !hasError {
		t.Error("expected CSS_003 for invalid UTF-8")
	}
}

func TestScanCSS_AtRules(t *testing.T) {
	content := []byte(`
@font-face {
  font-family: "MyFont";
  src: url("myfont.woff2");
}

@media screen {
  body { margin: 0; }
}
`)

	_, atRules, diags := ScanCSS(content)
	if len(diags) != 0 {
		t.Errorf("expected no diagnostics, got %d", len(diags))
	}

	if len(atRules) != 2 {
		t.Errorf("expected 2 at-rules, got %d", len(atRules))
	}

	if len(atRules) >= 1 && atRules[0].Name != "@font-face" {
		t.Errorf("expected @font-face, got %q", atRules[0].Name)
	}
	if len(atRules) >= 2 && atRules[1].Name != "@media" {
		t.Errorf("expected @media, got %q", atRules[1].Name)
	}
}

func TestScanCSS_PropertyValues(t *testing.T) {
	content := []byte(`
div {
  position: fixed;
  direction: rtl;
}
`)

	props, _, diags := ScanCSS(content)
	if len(diags) != 0 {
		t.Errorf("expected no diagnostics, got %d", len(diags))
	}

	if len(props) != 2 {
		t.Fatalf("expected 2 properties, got %d", len(props))
	}

	if props[0].Property != "position" || props[0].Value != "fixed" {
		t.Errorf(
			"expected position: fixed, got %s: %s",
			props[0].Property,
			props[0].Value,
		)
	}
	if props[1].Property != "direction" || props[1].Value != "rtl" {
		t.Errorf("expected direction: rtl, got %s: %s", props[1].Property, props[1].Value)
	}
}
