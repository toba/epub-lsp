# EPUB LSP

A Language Server Protocol (LSP) implementation for EPUB source files. Validates XHTML content documents, OPF package files, navigation documents, and CSS with rules ported from [epubcheck](https://github.com/w3c/epubcheck) and [ace](https://github.com/daisy/ace). No Java or Node.js dependencies.

## Features

- **Diagnostics**: Real-time validation of EPUB source files as you type
  - OPF package document validation (metadata, manifest, spine)
  - XHTML content document validation (namespaces, structure, accessibility)
  - Navigation document validation
  - CSS validation

## Installation

### Download Binary

Download prebuilt binaries from [GitHub Releases](https://github.com/toba/epub-lsp/releases).

### Install from Source

```bash
go install github.com/toba/epub-lsp/cmd/epub-lsp@latest
```

### Build from Source

```bash
git clone https://github.com/toba/epub-lsp.git
cd epub-lsp
go build -o epub-lsp ./cmd/epub-lsp
go test ./...
```

## Supported File Types

| Extension | Type |
|-----------|------|
| `.opf` | OPF package document |
| `.xhtml` | XHTML content document |
| `.html` | HTML content document |
| `.css` | CSS stylesheet |
| `.ncx` | NCX navigation (EPUB 2) |

## License

MIT License - see [LICENSE](LICENSE) for details.
