package epub

import "testing"

func TestDetectFileType(t *testing.T) {
	tests := []struct {
		name    string
		uri     string
		content []byte
		want    FileType
	}{
		{"OPF file", "package.opf", nil, FileTypeOPF},
		{"CSS file", "style.css", nil, FileTypeCSS},
		{"NCX file", "toc.ncx", nil, FileTypeNCX},
		{"XHTML file", "chapter1.xhtml", nil, FileTypeXHTML},
		{"HTML file", "chapter1.html", nil, FileTypeXHTML},
		{"Nav document", "nav.xhtml", []byte(`<nav epub:type="toc">`), FileTypeNav},
		{
			"Nav document single quotes",
			"nav.xhtml",
			[]byte(`<nav epub:type='toc'>`),
			FileTypeNav,
		},
		{"Unknown file", "image.png", nil, FileTypeUnknown},
		{"Case insensitive", "PACKAGE.OPF", nil, FileTypeOPF},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetectFileType(tt.uri, tt.content)
			if got != tt.want {
				t.Errorf("DetectFileType(%q) = %v, want %v", tt.uri, got, tt.want)
			}
		})
	}
}

func TestFileTypeString(t *testing.T) {
	tests := []struct {
		ft   FileType
		want string
	}{
		{FileTypeOPF, "OPF"},
		{FileTypeXHTML, "XHTML"},
		{FileTypeNav, "Nav"},
		{FileTypeCSS, "CSS"},
		{FileTypeNCX, "NCX"},
		{FileTypeUnknown, "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.ft.String(); got != tt.want {
				t.Errorf("FileType.String() = %q, want %q", got, tt.want)
			}
		})
	}
}
