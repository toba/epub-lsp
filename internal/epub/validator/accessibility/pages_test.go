package accessibility

import (
	"testing"

	"github.com/toba/epub-lsp/internal/epub"
	"github.com/toba/epub-lsp/internal/epub/testutil"
	"github.com/toba/epub-lsp/internal/epub/validator"
)

// defaultSeverity provides the default accessibility severity for tests.
const defaultSeverity = epub.SeverityWarning

func makeOPFWithFeature(feature string) []byte {
	return []byte(`<?xml version="1.0" encoding="UTF-8"?>
<package xmlns="http://www.idpf.org/2007/opf" unique-identifier="uid" version="3.0">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
    <dc:identifier id="uid">urn:isbn:123</dc:identifier>
    <dc:title>Test</dc:title>
    <dc:language>en</dc:language>
    <meta property="schema:accessibilityFeature">` + feature + `</meta>
  </metadata>
  <manifest/>
  <spine/>
</package>`)
}

func navWithPageList() []byte {
	return []byte(`<?xml version="1.0" encoding="UTF-8"?>
<html xmlns="http://www.w3.org/1999/xhtml" xmlns:epub="http://www.idpf.org/2007/ops" lang="en">
<head><title>Nav</title></head>
<body>
  <nav epub:type="toc">
    <ol><li><a href="ch1.xhtml">Ch1</a></li></ol>
  </nav>
  <nav epub:type="page-list">
    <ol>
      <li><a href="ch1.xhtml#pg1">1</a></li>
      <li><a href="ch1.xhtml#pg2">2</a></li>
    </ol>
  </nav>
</body>
</html>`)
}

func contentWithPageBreaks() []byte {
	return []byte(`<?xml version="1.0" encoding="UTF-8"?>
<html xmlns="http://www.w3.org/1999/xhtml" xmlns:epub="http://www.idpf.org/2007/ops" lang="en">
<head><title>Ch1</title></head>
<body>
  <span epub:type="pagebreak" id="pg1" aria-label="1"/>
  <p>Content</p>
  <span epub:type="pagebreak" id="pg2" aria-label="2"/>
</body>
</html>`)
}

func TestPageValidator_PrintPageNumbers_NoPageList(t *testing.T) {
	opfContent := makeOPFWithFeature("printPageNumbers")

	ctx := &validator.WorkspaceContext{
		Manifest: &validator.ManifestInfo{
			Metadata: validator.MetadataInfo{
				AccessibilityFeatures: []string{"printPageNumbers"},
			},
		},
		Files: map[string][]byte{
			"file:///book/OEBPS/package.opf": opfContent,
			"file:///book/OEBPS/ch1.xhtml":   contentWithPageBreaks(),
		},
		FileTypes: map[string]epub.FileType{
			"file:///book/OEBPS/package.opf": epub.FileTypeOPF,
			"file:///book/OEBPS/ch1.xhtml":   epub.FileTypeXHTML,
		},
		AccessibilitySeverity: defaultSeverity,
	}

	v := &PageValidator{}
	diags := v.Validate("file:///book/OEBPS/package.opf", opfContent, ctx)

	testutil.ExpectCode(t, testutil.DiagCodes(diags), "printPageNumbers-nopagelist")
}

func TestPageValidator_PrintPageNumbers_NoPageBreaks(t *testing.T) {
	opfContent := makeOPFWithFeature("printPageNumbers")

	plainContent := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<html xmlns="http://www.w3.org/1999/xhtml" lang="en">
<head><title>Ch1</title></head>
<body><p>Content</p></body>
</html>`)

	ctx := &validator.WorkspaceContext{
		Manifest: &validator.ManifestInfo{
			Metadata: validator.MetadataInfo{
				AccessibilityFeatures: []string{"printPageNumbers"},
			},
		},
		Files: map[string][]byte{
			"file:///book/OEBPS/package.opf": opfContent,
			"file:///book/OEBPS/nav.xhtml":   navWithPageList(),
			"file:///book/OEBPS/ch1.xhtml":   plainContent,
		},
		FileTypes: map[string]epub.FileType{
			"file:///book/OEBPS/package.opf": epub.FileTypeOPF,
			"file:///book/OEBPS/nav.xhtml":   epub.FileTypeNav,
			"file:///book/OEBPS/ch1.xhtml":   epub.FileTypeXHTML,
		},
		AccessibilitySeverity: defaultSeverity,
	}

	v := &PageValidator{}
	diags := v.Validate("file:///book/OEBPS/package.opf", opfContent, ctx)

	testutil.ExpectCode(t, testutil.DiagCodes(diags), "printPageNumbers-nopagebreaks")
}

func TestPageValidator_PrintPageNumbers_AllPresent(t *testing.T) {
	opfContent := makeOPFWithFeature("printPageNumbers")

	ctx := &validator.WorkspaceContext{
		Manifest: &validator.ManifestInfo{
			Metadata: validator.MetadataInfo{
				AccessibilityFeatures: []string{"printPageNumbers"},
				HasDCSource:           true,
			},
		},
		Files: map[string][]byte{
			"file:///book/OEBPS/package.opf": opfContent,
			"file:///book/OEBPS/nav.xhtml":   navWithPageList(),
			"file:///book/OEBPS/ch1.xhtml":   contentWithPageBreaks(),
		},
		FileTypes: map[string]epub.FileType{
			"file:///book/OEBPS/package.opf": epub.FileTypeOPF,
			"file:///book/OEBPS/nav.xhtml":   epub.FileTypeNav,
			"file:///book/OEBPS/ch1.xhtml":   epub.FileTypeXHTML,
		},
		AccessibilitySeverity: defaultSeverity,
	}

	v := &PageValidator{}
	diags := v.Validate("file:///book/OEBPS/package.opf", opfContent, ctx)

	codes := testutil.DiagCodes(diags)
	if codes["printPageNumbers-nopagelist"] {
		t.Error("unexpected printPageNumbers-nopagelist")
	}
	if codes["printPageNumbers-nopagebreaks"] {
		t.Error("unexpected printPageNumbers-nopagebreaks")
	}
}

func TestPageValidator_PageListWithoutDCSource(t *testing.T) {
	opfContent := makeOPFWithFeature("structuralNavigation")

	ctx := &validator.WorkspaceContext{
		Manifest: &validator.ManifestInfo{
			Metadata: validator.MetadataInfo{
				AccessibilityFeatures: []string{"structuralNavigation"},
				HasDCSource:           false,
			},
		},
		Files: map[string][]byte{
			"file:///book/OEBPS/package.opf": opfContent,
			"file:///book/OEBPS/nav.xhtml":   navWithPageList(),
		},
		FileTypes: map[string]epub.FileType{
			"file:///book/OEBPS/package.opf": epub.FileTypeOPF,
			"file:///book/OEBPS/nav.xhtml":   epub.FileTypeNav,
		},
		AccessibilitySeverity: defaultSeverity,
	}

	v := &PageValidator{}
	diags := v.Validate("file:///book/OEBPS/package.opf", opfContent, ctx)

	testutil.ExpectCode(t, testutil.DiagCodes(diags), "epub-pagesource")
}

func TestPageValidator_BrokenPageListRef(t *testing.T) {
	opfContent := makeOPFWithFeature("printPageNumbers")

	// Content without pg2 id
	contentMissingID := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<html xmlns="http://www.w3.org/1999/xhtml" xmlns:epub="http://www.idpf.org/2007/ops" lang="en">
<head><title>Ch1</title></head>
<body>
  <span epub:type="pagebreak" id="pg1" aria-label="1"/>
  <p>Content</p>
</body>
</html>`)

	ctx := &validator.WorkspaceContext{
		Manifest: &validator.ManifestInfo{
			Metadata: validator.MetadataInfo{
				AccessibilityFeatures: []string{"printPageNumbers"},
				HasDCSource:           true,
			},
		},
		Files: map[string][]byte{
			"file:///book/OEBPS/package.opf": opfContent,
			"file:///book/OEBPS/nav.xhtml":   navWithPageList(),
			"file:///book/OEBPS/ch1.xhtml":   contentMissingID,
		},
		FileTypes: map[string]epub.FileType{
			"file:///book/OEBPS/package.opf": epub.FileTypeOPF,
			"file:///book/OEBPS/nav.xhtml":   epub.FileTypeNav,
			"file:///book/OEBPS/ch1.xhtml":   epub.FileTypeXHTML,
		},
		AccessibilitySeverity: defaultSeverity,
	}

	v := &PageValidator{}
	diags := v.Validate("file:///book/OEBPS/package.opf", opfContent, ctx)

	testutil.ExpectCode(t, testutil.DiagCodes(diags), "epub-pagelist-broken")
}

func TestPageValidator_NilContext(t *testing.T) {
	v := &PageValidator{}
	diags := v.Validate("package.opf", []byte("<package/>"), nil)

	if len(diags) != 0 {
		t.Errorf("expected no diagnostics with nil context, got %d", len(diags))
	}
}
