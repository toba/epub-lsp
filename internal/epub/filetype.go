package epub

import (
	"bytes"
	"path/filepath"
	"strings"
)

// FileType identifies the kind of EPUB source file.
type FileType int

const (
	FileTypeUnknown FileType = iota
	FileTypeOPF
	FileTypeXHTML
	FileTypeNav
	FileTypeCSS
	FileTypeNCX
)

// DetectFileType determines the file type from extension and content.
// Content sniffing is used to detect navigation documents (epub:type="toc").
func DetectFileType(uri string, content []byte) FileType {
	ext := strings.ToLower(filepath.Ext(uri))

	switch ext {
	case ".opf":
		return FileTypeOPF
	case ".css":
		return FileTypeCSS
	case ".ncx":
		return FileTypeNCX
	case ".xhtml", ".html":
		if isNavDocument(content) {
			return FileTypeNav
		}
		return FileTypeXHTML
	}

	return FileTypeUnknown
}

// isNavDocument checks if XHTML content is a navigation document.
func isNavDocument(content []byte) bool {
	return bytes.Contains(content, []byte(`epub:type="toc"`)) ||
		bytes.Contains(content, []byte(`epub:type='toc'`))
}

// String returns a human-readable name for the file type.
func (ft FileType) String() string {
	switch ft {
	case FileTypeOPF:
		return "OPF"
	case FileTypeXHTML:
		return "XHTML"
	case FileTypeNav:
		return "Nav"
	case FileTypeCSS:
		return "CSS"
	case FileTypeNCX:
		return "NCX"
	default:
		return "Unknown"
	}
}
