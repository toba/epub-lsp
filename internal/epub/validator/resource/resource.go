// Package resource validates cross-file resource references in EPUB.
package resource

import (
	"net/url"
	"path"
	"strings"

	"github.com/toba/epub-lsp/internal/epub"
	"github.com/toba/epub-lsp/internal/epub/parser"
	"github.com/toba/epub-lsp/internal/epub/validator"
)

const source = "epub-resource"

// ManifestValidator checks that manifest hrefs reference existing files.
// It runs on OPF files.
type ManifestValidator struct{}

func (v *ManifestValidator) FileTypes() []epub.FileType {
	return []epub.FileType{epub.FileTypeOPF}
}

func (v *ManifestValidator) Validate(
	uri string,
	content []byte,
	ctx *validator.WorkspaceContext,
) []epub.Diagnostic {
	if ctx == nil || ctx.Files == nil {
		return nil
	}

	root, xmlDiags := parser.Parse(content)
	if len(xmlDiags) > 0 {
		return nil // XML errors handled by the OPF validator
	}

	pkg := root.FindFirst("package")
	if pkg == nil {
		return nil
	}

	manifest := pkg.FindFirst("manifest")
	if manifest == nil {
		return nil
	}

	// Determine the OPF directory for resolving relative hrefs
	opfDir := dirFromURI(uri)

	var diags []epub.Diagnostic

	for _, item := range manifest.Children {
		if item.Local != "item" {
			continue
		}

		href := item.Attr("href")
		if href == "" {
			continue
		}

		// Skip remote resources
		if epub.IsRemoteURL(href) {
			continue
		}

		// Resolve relative href against OPF directory
		resolvedURI := resolveHref(opfDir, href)

		if !fileExistsInWorkspace(resolvedURI, ctx.Files) {
			diags = append(diags, epub.NewDiag(content, int(item.Offset), source).
				Code("RSC_007").
				Error("manifest item references missing file: "+href).Build())
		}
	}

	return diags
}

// ContentValidator checks that resources referenced in content documents
// are listed in the manifest. It runs on XHTML and Nav files.
type ContentValidator struct{}

func (v *ContentValidator) FileTypes() []epub.FileType {
	return []epub.FileType{epub.FileTypeXHTML, epub.FileTypeNav}
}

func (v *ContentValidator) Validate(
	uri string,
	content []byte,
	ctx *validator.WorkspaceContext,
) []epub.Diagnostic {
	if ctx == nil || ctx.Manifest == nil {
		return nil
	}

	root, xmlDiags := parser.Parse(content)
	if len(xmlDiags) > 0 {
		return nil
	}

	// Build set of manifest hrefs
	manifestHrefs := make(map[string]bool)
	for _, item := range ctx.Manifest.Items {
		manifestHrefs[item.Href] = true
	}

	contentDir := dirFromURI(uri)

	var diags []epub.Diagnostic

	// Check <img src="...">
	imgs := root.FindAll("img")
	for _, img := range imgs {
		src := img.Attr("src")
		if src == "" {
			continue
		}
		if epub.IsRemoteURL(src) || strings.HasPrefix(src, "data:") {
			continue
		}
		checkResourceInManifest(content, img, src, contentDir, manifestHrefs, &diags)
	}

	// Check <link href="..."> (typically CSS)
	links := root.FindAll("link")
	for _, link := range links {
		href := link.Attr("href")
		if href == "" {
			continue
		}
		if epub.IsRemoteURL(href) {
			continue
		}
		checkResourceInManifest(content, link, href, contentDir, manifestHrefs, &diags)
	}

	// Check <image> and <source> elements (for SVG/audio/video in XHTML)
	for _, tagName := range []string{"image", "source", "audio", "video"} {
		elems := root.FindAll(tagName)
		for _, elem := range elems {
			src := elem.Attr("src")
			if src == "" {
				src = elem.Attr("href")
			}
			if src == "" || epub.IsRemoteURL(src) ||
				strings.HasPrefix(src, "data:") {
				continue
			}
			checkResourceInManifest(content, elem, src, contentDir, manifestHrefs, &diags)
		}
	}

	return diags
}

func checkResourceInManifest(
	content []byte,
	node *parser.XMLNode,
	ref string,
	contentDir string,
	manifestHrefs map[string]bool,
	diags *[]epub.Diagnostic,
) {
	ref = epub.StripFragment(ref)
	if ref == "" {
		return
	}

	// We need to resolve relative to the OPF location to match manifest hrefs.
	// The manifest hrefs are relative to the OPF, so we need the path
	// of this content file relative to the OPF.
	// If we don't have an OPF path, we try a simpler approach.
	resolved := resolveHref(contentDir, ref)

	// Try to match against manifest hrefs.
	// Manifest hrefs are relative to OPF. Content refs are relative to content file.
	// We need a common resolution. Check if the resolved href ends with any manifest href.
	found := false
	for manifestHref := range manifestHrefs {
		if pathEndsWith(resolved, manifestHref) {
			found = true
			break
		}
	}

	// Also try the raw ref (in case content and OPF are in the same directory)
	if !found && manifestHrefs[ref] {
		found = true
	}

	if !found {
		*diags = append(*diags, epub.NewDiag(content, int(node.Offset), source).
			Code("RSC_008").Warning("resource not found in manifest: "+ref).Build())
	}
}

// dirFromURI returns the directory portion of a URI.
func dirFromURI(uri string) string {
	// Try to parse as URL first
	if u, err := url.Parse(uri); err == nil && u.Path != "" {
		return path.Dir(u.Path)
	}
	idx := strings.LastIndex(uri, "/")
	if idx >= 0 {
		return uri[:idx]
	}
	return ""
}

// resolveHref resolves a relative href against a base directory.
func resolveHref(baseDir, href string) string {
	if href == "" {
		return ""
	}
	// URL-decode the href
	if decoded, err := url.PathUnescape(href); err == nil {
		href = decoded
	}
	if path.IsAbs(href) {
		return href
	}
	return path.Clean(baseDir + "/" + href)
}

// fileExistsInWorkspace checks if a file URI exists in the workspace files.
func fileExistsInWorkspace(resolvedPath string, files map[string][]byte) bool {
	for fileURI := range files {
		if u, err := url.Parse(fileURI); err == nil {
			if u.Path == resolvedPath || strings.HasSuffix(u.Path, resolvedPath) {
				return true
			}
		}
		if strings.HasSuffix(fileURI, resolvedPath) {
			return true
		}
	}
	return false
}

// pathEndsWith checks if a full path ends with a suffix path.
func pathEndsWith(full, suffix string) bool {
	if full == suffix {
		return true
	}
	return strings.HasSuffix(full, "/"+suffix)
}
