package nav

import (
	"testing"

	"github.com/toba/epub-lsp/internal/epub"
	"github.com/toba/epub-lsp/internal/epub/testutil"
	"github.com/toba/epub-lsp/internal/epub/validator"
)

func TestValidNav(t *testing.T) {
	content := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<html xmlns="http://www.w3.org/1999/xhtml" xmlns:epub="http://www.idpf.org/2007/ops" lang="en">
<head><title>Navigation</title></head>
<body>
  <nav epub:type="toc">
    <h2>Table of Contents</h2>
    <ol>
      <li><a href="chapter1.xhtml">Chapter 1</a></li>
      <li><a href="chapter2.xhtml">Chapter 2</a></li>
    </ol>
  </nav>
  <nav epub:type="landmarks">
    <ol>
      <li><a epub:type="toc" href="#toc">Table of Contents</a></li>
    </ol>
  </nav>
</body>
</html>`)

	v := &Validator{}
	diags := v.Validate("nav.xhtml", content, nil)

	// Should only have info-level or no diagnostics (no page-list is just info)
	for _, d := range diags {
		if d.Severity <= epub.SeverityWarning {
			t.Errorf("unexpected error/warning: [%s] %s", d.Code, d.Message)
		}
	}
}

func TestMissingTocNav(t *testing.T) {
	content := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<html xmlns="http://www.w3.org/1999/xhtml" xmlns:epub="http://www.idpf.org/2007/ops" lang="en">
<head><title>Navigation</title></head>
<body>
  <nav epub:type="landmarks">
    <ol>
      <li><a href="chapter1.xhtml">Start</a></li>
    </ol>
  </nav>
</body>
</html>`)

	v := &Validator{}
	diags := v.Validate("nav.xhtml", content, nil)

	if !testutil.HasCode(diags, "NAV_003") {
		t.Error("expected NAV_003 for missing toc nav")
	}
}

func TestMissingOlInTocNav(t *testing.T) {
	content := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<html xmlns="http://www.w3.org/1999/xhtml" xmlns:epub="http://www.idpf.org/2007/ops" lang="en">
<head><title>Navigation</title></head>
<body>
  <nav epub:type="toc">
    <h2>Table of Contents</h2>
    <ul>
      <li><a href="chapter1.xhtml">Chapter 1</a></li>
    </ul>
  </nav>
</body>
</html>`)

	v := &Validator{}
	diags := v.Validate("nav.xhtml", content, nil)

	hasOlWarning := false
	for _, d := range diags {
		if d.Severity == epub.SeverityWarning &&
			d.Message == "toc nav is missing required <ol> element" {
			hasOlWarning = true
			break
		}
	}
	if !hasOlWarning {
		t.Error("expected warning for missing <ol> in toc nav")
	}
}

func TestRemoteNavLink(t *testing.T) {
	content := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<html xmlns="http://www.w3.org/1999/xhtml" xmlns:epub="http://www.idpf.org/2007/ops" lang="en">
<head><title>Navigation</title></head>
<body>
  <nav epub:type="toc">
    <ol>
      <li><a href="chapter1.xhtml">Chapter 1</a></li>
      <li><a href="https://example.com/chapter2">Chapter 2</a></li>
    </ol>
  </nav>
</body>
</html>`)

	v := &Validator{}
	diags := v.Validate("nav.xhtml", content, nil)

	if !testutil.HasCode(diags, "NAV_010") {
		t.Error("expected NAV_010 for remote link")
	}
}

func TestNoPageListOrLandmarks(t *testing.T) {
	content := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<html xmlns="http://www.w3.org/1999/xhtml" xmlns:epub="http://www.idpf.org/2007/ops" lang="en">
<head><title>Navigation</title></head>
<body>
  <nav epub:type="toc">
    <ol>
      <li><a href="chapter1.xhtml">Chapter 1</a></li>
    </ol>
  </nav>
</body>
</html>`)

	v := &Validator{}
	diags := v.Validate("nav.xhtml", content, nil)

	hasInfo := false
	for _, d := range diags {
		if d.Severity == epub.SeverityInfo &&
			d.Message == "navigation document has no page-list or landmarks nav" {
			hasInfo = true
			break
		}
	}
	if !hasInfo {
		t.Error("expected info diagnostic for missing page-list/landmarks")
	}
}

func TestTocSpineOrderMismatch(t *testing.T) {
	content := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<html xmlns="http://www.w3.org/1999/xhtml" xmlns:epub="http://www.idpf.org/2007/ops" lang="en">
<head><title>Navigation</title></head>
<body>
  <nav epub:type="toc">
    <ol>
      <li><a href="chapter2.xhtml">Chapter 2</a></li>
      <li><a href="chapter1.xhtml">Chapter 1</a></li>
    </ol>
  </nav>
</body>
</html>`)

	ctx := &validator.WorkspaceContext{
		Manifest: &validator.ManifestInfo{
			Items: []validator.ManifestItem{
				{ID: "ch1", Href: "chapter1.xhtml", MediaType: "application/xhtml+xml"},
				{ID: "ch2", Href: "chapter2.xhtml", MediaType: "application/xhtml+xml"},
			},
			Spine: []validator.SpineItem{
				{IDRef: "ch1", Linear: true},
				{IDRef: "ch2", Linear: true},
			},
		},
	}

	v := &Validator{}
	diags := v.Validate("nav.xhtml", content, ctx)

	if !testutil.HasCode(diags, "NAV_011") {
		t.Error("expected NAV_011 for TOC/spine order mismatch")
	}
}

func TestTocSpineOrderMatch(t *testing.T) {
	content := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<html xmlns="http://www.w3.org/1999/xhtml" xmlns:epub="http://www.idpf.org/2007/ops" lang="en">
<head><title>Navigation</title></head>
<body>
  <nav epub:type="toc">
    <ol>
      <li><a href="chapter1.xhtml">Chapter 1</a></li>
      <li><a href="chapter2.xhtml">Chapter 2</a></li>
    </ol>
  </nav>
</body>
</html>`)

	ctx := &validator.WorkspaceContext{
		Manifest: &validator.ManifestInfo{
			Items: []validator.ManifestItem{
				{ID: "ch1", Href: "chapter1.xhtml", MediaType: "application/xhtml+xml"},
				{ID: "ch2", Href: "chapter2.xhtml", MediaType: "application/xhtml+xml"},
			},
			Spine: []validator.SpineItem{
				{IDRef: "ch1", Linear: true},
				{IDRef: "ch2", Linear: true},
			},
		},
	}

	v := &Validator{}
	diags := v.Validate("nav.xhtml", content, ctx)

	if testutil.HasCode(diags, "NAV_011") {
		t.Error("unexpected NAV_011 when TOC matches spine order")
	}
}
