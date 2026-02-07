package formatter

import (
	"strings"
	"testing"
)

func TestFormatCSS_BasicFormatting(t *testing.T) {
	input := []byte(`body{color:red;font-size:12px;}`)
	result, err := FormatCSS(input, "  ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should contain indentation
	if !strings.Contains(result, "  color") {
		t.Errorf("expected indented property, got:\n%s", result)
	}

	// Should end with newline
	if !strings.HasSuffix(result, "\n") {
		t.Error("expected trailing newline")
	}
}

func TestFormatCSS_AtRules(t *testing.T) {
	input := []byte(`@charset "utf-8";
@font-face{font-family:"Test";}`)
	result, err := FormatCSS(input, "  ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "@charset") {
		t.Error("expected @charset in output")
	}
	if !strings.Contains(result, "@font-face") {
		t.Error("expected @font-face in output")
	}
}

func TestFormatCSS_TabIndent(t *testing.T) {
	input := []byte(`p{margin:0;}`)
	result, err := FormatCSS(input, "\t")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "\tmargin") {
		t.Errorf("expected tab-indented property, got:\n%s", result)
	}
}

func TestFormatCSS_MultipleSelectors(t *testing.T) {
	input := []byte(`body{color:red;}p{margin:0;}`)
	result, err := FormatCSS(input, "  ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "body") || !strings.Contains(result, "p") {
		t.Errorf("expected both selectors in output, got:\n%s", result)
	}
}

func TestFormatCSS_Empty(t *testing.T) {
	input := []byte(``)
	result, err := FormatCSS(input, "  ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != "\n" {
		t.Errorf("expected just a newline for empty input, got: %q", result)
	}
}

func TestFormatCSS_Comments(t *testing.T) {
	input := []byte(`/* header styles */
body{color:red;}`)
	result, err := FormatCSS(input, "  ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "/* header styles */") {
		t.Error("expected comment to be preserved")
	}
}
