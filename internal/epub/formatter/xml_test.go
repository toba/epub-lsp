package formatter

import (
	"strings"
	"testing"
)

func TestFormatXML_BasicIndentation(t *testing.T) {
	input := []byte(
		`<?xml version="1.0"?><package><metadata><title>Test</title></metadata></package>`,
	)
	result, err := FormatXML(input, "  ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should contain indentation
	if !strings.Contains(result, "  ") {
		t.Error("expected indented output")
	}

	// Should end with newline
	if !strings.HasSuffix(result, "\n") {
		t.Error("expected trailing newline")
	}
}

func TestFormatXML_PreservesDeclaration(t *testing.T) {
	input := []byte(`<?xml version="1.0" encoding="UTF-8"?><root><child/></root>`)
	result, err := FormatXML(input, "  ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.HasPrefix(result, `<?xml version="1.0" encoding="UTF-8"?>`) {
		t.Error("expected XML declaration to be preserved")
	}
}

func TestFormatXML_TabIndent(t *testing.T) {
	input := []byte(`<root><child>text</child></root>`)
	result, err := FormatXML(input, "\t")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "\t") {
		t.Error("expected tab indentation")
	}
}

func TestFormatXML_InvalidXML(t *testing.T) {
	input := []byte(`<root><unclosed>`)
	_, err := FormatXML(input, "  ")
	if err == nil {
		t.Error("expected error for invalid XML")
	}
}

func TestFormatXML_EmptyDocument(t *testing.T) {
	input := []byte(`<root/>`)
	result, err := FormatXML(input, "  ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "root") {
		t.Error("expected root element in output")
	}
}

func TestFormatXML_NamespacePreservation(t *testing.T) {
	input := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE html>
<html xmlns="http://www.w3.org/1999/xhtml" xmlns:epub="http://www.idpf.org/2007/ops" xml:lang="en" lang="en">
<head>
<meta charset="utf-8"/>
<title>Test Title</title>
<link rel="stylesheet" type="text/css" href="../styles/book.css"/>
</head>
<body>
<div class="map-page">
<h2>Locations</h2>
<img src="../images/map.png" alt="Map of photo locations"/>
</div>
</body>
</html>`)

	result, err := FormatXML(input, "   ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should NOT duplicate xmlns on child elements
	count := strings.Count(result, `xmlns="http://www.w3.org/1999/xhtml"`)
	if count != 1 {
		t.Errorf("expected xmlns to appear once, got %d times\n%s", count, result)
	}

	// Should preserve xmlns:epub correctly (not mangle to _xmlns:epub)
	if !strings.Contains(result, `xmlns:epub="http://www.idpf.org/2007/ops"`) {
		t.Errorf("expected xmlns:epub to be preserved\n%s", result)
	}

	// Should preserve self-closing tags
	if !strings.Contains(result, `<meta charset="utf-8"/>`) {
		t.Errorf("expected self-closing meta tag\n%s", result)
	}

	// Should NOT expand self-closing to open/close
	if strings.Contains(result, "</meta>") {
		t.Errorf("should not expand self-closing meta\n%s", result)
	}
	if strings.Contains(result, "</img>") {
		t.Errorf("should not expand self-closing img\n%s", result)
	}
	if strings.Contains(result, "</link>") {
		t.Errorf("should not expand self-closing link\n%s", result)
	}

	// Should have newline after DOCTYPE
	if strings.Contains(result, "<!DOCTYPE html><") {
		t.Errorf("expected newline after DOCTYPE\n%s", result)
	}

	// Should NOT contain mangled namespace prefixes
	if strings.Contains(result, "_xmlns") {
		t.Errorf("should not mangle xmlns prefix\n%s", result)
	}
}

func TestFormatXML_XHTMLFullDocument(t *testing.T) {
	input := []byte(
		`<?xml version="1.0" encoding="UTF-8"?><!DOCTYPE html><html xmlns="http://www.w3.org/1999/xhtml" xmlns:epub="http://www.idpf.org/2007/ops" xml:lang="en" lang="en"><head><meta charset="utf-8"/><title>Test</title></head><body><div class="page"><h2>Hello</h2></div></body></html>`,
	)

	result, err := FormatXML(input, "   ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE html>
<html xmlns="http://www.w3.org/1999/xhtml" xmlns:epub="http://www.idpf.org/2007/ops" xml:lang="en" lang="en">
   <head>
      <meta charset="utf-8"/>
      <title>Test</title>
   </head>
   <body>
      <div class="page">
         <h2>Hello</h2>
      </div>
   </body>
</html>
`

	if result != expected {
		t.Errorf("output mismatch\nexpected:\n%s\ngot:\n%s", expected, result)
	}
}

func TestFormatXML_Comment(t *testing.T) {
	input := []byte(`<root><!-- a comment --><child/></root>`)
	result, err := FormatXML(input, "  ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "<!-- a comment -->") {
		t.Errorf("expected comment to be preserved\n%s", result)
	}
}

func TestFormatXML_SelfClosingPreserved(t *testing.T) {
	input := []byte(`<root><br/><hr/><img src="test.png"/></root>`)
	result, err := FormatXML(input, "  ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "<br/>") {
		t.Errorf("expected self-closing br\n%s", result)
	}
	if !strings.Contains(result, "<hr/>") {
		t.Errorf("expected self-closing hr\n%s", result)
	}
	if !strings.Contains(result, `<img src="test.png"/>`) {
		t.Errorf("expected self-closing img\n%s", result)
	}
}
