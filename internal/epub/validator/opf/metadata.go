package opf

import (
	"github.com/toba/epub-lsp/internal/epub"
	"github.com/toba/epub-lsp/internal/epub/parser"
)

func validateMetadata(content []byte, pkg *parser.XMLNode) []epub.Diagnostic {
	var diags []epub.Diagnostic

	metadata := pkg.FindFirst("metadata")
	if metadata == nil {
		diags = append(diags, epub.NewDiag(content, int(pkg.Offset), source).
			Code("OPF_030").Error("missing required <metadata> element").Build())
		return diags
	}

	// Check dc:identifier
	identifiers := metadata.FindAllNS(epub.NSDC, "identifier")
	if len(identifiers) == 0 {
		diags = append(diags, epub.NewDiag(content, int(metadata.Offset), source).
			Code("OPF_030").Error("missing required <dc:identifier> in metadata").Build())
	}

	// Check unique-identifier references a valid dc:identifier
	uniqueID := pkg.Attr("unique-identifier")
	if uniqueID != "" && len(identifiers) > 0 {
		found := false
		for _, id := range identifiers {
			if id.Attr("id") == uniqueID {
				found = true
				break
			}
		}
		if !found {
			diags = append(diags, epub.NewDiag(content, int(pkg.Offset), source).
				Code("OPF_031").
				Error("unique-identifier \""+uniqueID+"\" does not match any dc:identifier/@id").
				Build())
		}
	}

	// Check dc:title
	titles := metadata.FindAllNS(epub.NSDC, "title")
	if len(titles) == 0 {
		diags = append(diags, epub.NewDiag(content, int(metadata.Offset), source).
			Code("OPF_032").Error("missing required <dc:title> in metadata").Build())
	}

	// Check dc:language
	languages := metadata.FindAllNS(epub.NSDC, "language")
	if len(languages) == 0 {
		diags = append(diags, epub.NewDiag(content, int(metadata.Offset), source).
			Code("OPF_034").Error("missing required <dc:language> in metadata").Build())
	}

	return diags
}
