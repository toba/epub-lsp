package resource

import (
	"testing"

	"github.com/toba/epub-lsp/internal/epub"
	"github.com/toba/epub-lsp/internal/epub/validator"
)

func TestManifestValidator_MissingFile(t *testing.T) {
	content := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<package xmlns="http://www.idpf.org/2007/opf" unique-identifier="uid" version="3.0">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
    <dc:identifier id="uid">urn:isbn:123</dc:identifier>
    <dc:title>Test</dc:title>
    <dc:language>en</dc:language>
  </metadata>
  <manifest>
    <item id="ch1" href="chapter1.xhtml" media-type="application/xhtml+xml"/>
    <item id="missing" href="nonexistent.xhtml" media-type="application/xhtml+xml"/>
  </manifest>
  <spine>
    <itemref idref="ch1"/>
  </spine>
</package>`)

	ctx := &validator.WorkspaceContext{
		Files: map[string][]byte{
			"file:///book/OEBPS/package.opf":    content,
			"file:///book/OEBPS/chapter1.xhtml": []byte("<html/>"),
		},
	}

	v := &ManifestValidator{}
	diags := v.Validate("file:///book/OEBPS/package.opf", content, ctx)

	if !hasCode(diags, "RSC_007") {
		t.Error("expected RSC_007 for missing file")
	}

	// chapter1.xhtml should not trigger RSC_007
	rsc007Count := 0
	for _, d := range diags {
		if d.Code == "RSC_007" {
			rsc007Count++
		}
	}
	if rsc007Count != 1 {
		t.Errorf("expected exactly 1 RSC_007, got %d", rsc007Count)
	}
}

func TestManifestValidator_AllFilesExist(t *testing.T) {
	content := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<package xmlns="http://www.idpf.org/2007/opf" unique-identifier="uid" version="3.0">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
    <dc:identifier id="uid">urn:isbn:123</dc:identifier>
    <dc:title>Test</dc:title>
    <dc:language>en</dc:language>
  </metadata>
  <manifest>
    <item id="ch1" href="chapter1.xhtml" media-type="application/xhtml+xml"/>
    <item id="css" href="style.css" media-type="text/css"/>
  </manifest>
  <spine>
    <itemref idref="ch1"/>
  </spine>
</package>`)

	ctx := &validator.WorkspaceContext{
		Files: map[string][]byte{
			"file:///book/OEBPS/package.opf":    content,
			"file:///book/OEBPS/chapter1.xhtml": []byte("<html/>"),
			"file:///book/OEBPS/style.css":      []byte("body {}"),
		},
	}

	v := &ManifestValidator{}
	diags := v.Validate("file:///book/OEBPS/package.opf", content, ctx)

	if hasCode(diags, "RSC_007") {
		t.Error("unexpected RSC_007 when all files exist")
	}
}

func TestManifestValidator_NilContext(t *testing.T) {
	content := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<package xmlns="http://www.idpf.org/2007/opf" unique-identifier="uid" version="3.0">
  <manifest>
    <item id="ch1" href="chapter1.xhtml" media-type="application/xhtml+xml"/>
  </manifest>
</package>`)

	v := &ManifestValidator{}
	diags := v.Validate("package.opf", content, nil)

	if len(diags) != 0 {
		t.Errorf("expected no diagnostics with nil context, got %d", len(diags))
	}
}

func TestContentValidator_ResourceNotInManifest(t *testing.T) {
	content := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<html xmlns="http://www.w3.org/1999/xhtml" lang="en">
<head>
  <title>Test</title>
  <link rel="stylesheet" href="style.css"/>
</head>
<body>
  <img src="cover.jpg" alt="Cover"/>
  <img src="photo.png" alt="Photo"/>
</body>
</html>`)

	ctx := &validator.WorkspaceContext{
		Manifest: &validator.ManifestInfo{
			Items: []validator.ManifestItem{
				{ID: "ch1", Href: "chapter1.xhtml", MediaType: "application/xhtml+xml"},
				{ID: "css", Href: "style.css", MediaType: "text/css"},
				{ID: "cover", Href: "cover.jpg", MediaType: "image/jpeg"},
			},
		},
	}

	v := &ContentValidator{}
	diags := v.Validate("file:///book/OEBPS/chapter1.xhtml", content, ctx)

	// photo.png is not in manifest
	if !hasCode(diags, "RSC_008") {
		t.Error("expected RSC_008 for photo.png not in manifest")
	}

	// Only one RSC_008 (cover.jpg and style.css are in manifest)
	rsc008Count := 0
	for _, d := range diags {
		if d.Code == "RSC_008" {
			rsc008Count++
		}
	}
	if rsc008Count != 1 {
		t.Errorf("expected exactly 1 RSC_008, got %d", rsc008Count)
	}
}

func TestContentValidator_AllResourcesInManifest(t *testing.T) {
	content := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<html xmlns="http://www.w3.org/1999/xhtml" lang="en">
<head>
  <title>Test</title>
  <link rel="stylesheet" href="style.css"/>
</head>
<body>
  <img src="cover.jpg" alt="Cover"/>
</body>
</html>`)

	ctx := &validator.WorkspaceContext{
		Manifest: &validator.ManifestInfo{
			Items: []validator.ManifestItem{
				{ID: "ch1", Href: "chapter1.xhtml", MediaType: "application/xhtml+xml"},
				{ID: "css", Href: "style.css", MediaType: "text/css"},
				{ID: "cover", Href: "cover.jpg", MediaType: "image/jpeg"},
			},
		},
	}

	v := &ContentValidator{}
	diags := v.Validate("file:///book/OEBPS/chapter1.xhtml", content, ctx)

	if hasCode(diags, "RSC_008") {
		t.Error("unexpected RSC_008 when all resources are in manifest")
	}
}

func TestContentValidator_SkipsRemoteAndDataURIs(t *testing.T) {
	content := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<html xmlns="http://www.w3.org/1999/xhtml" lang="en">
<head><title>Test</title></head>
<body>
  <img src="https://example.com/remote.jpg" alt="Remote"/>
  <img src="data:image/png;base64,abc" alt="Data"/>
</body>
</html>`)

	ctx := &validator.WorkspaceContext{
		Manifest: &validator.ManifestInfo{
			Items: []validator.ManifestItem{},
		},
	}

	v := &ContentValidator{}
	diags := v.Validate("chapter.xhtml", content, ctx)

	if hasCode(diags, "RSC_008") {
		t.Error("unexpected RSC_008 for remote/data URIs")
	}
}

func TestContentValidator_NilManifest(t *testing.T) {
	content := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<html xmlns="http://www.w3.org/1999/xhtml" lang="en">
<head><title>Test</title></head>
<body><img src="cover.jpg" alt="Cover"/></body>
</html>`)

	v := &ContentValidator{}
	diags := v.Validate("chapter.xhtml", content, nil)

	if len(diags) != 0 {
		t.Errorf("expected no diagnostics with nil context, got %d", len(diags))
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
