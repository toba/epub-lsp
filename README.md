# EPUB LSP

A Language Server Protocol implementation for EPUB source files. Validates XHTML content documents, OPF package files, navigation documents, and CSS with rules based on [epubcheck](https://github.com/w3c/epubcheck) and [ace](https://github.com/daisy/ace). Pure Go, no Java or Node.js dependencies.

## Installation

### Download Binary

Download prebuilt binaries from [GitHub Releases](https://github.com/toba/epub-lsp/releases).

Available for Linux, macOS, and Windows (amd64/arm64).

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

## Editor Integration

epub-lsp communicates over stdin/stdout using JSON-RPC per the LSP specification. Point your editor's LSP client at the `epub-lsp` binary for `.opf`, `.xhtml`, `.html`, and `.css` files.

A Zed extension is available at [gubby](https://github.com/toba/gubby).

## Supported File Types

| Extension | Type | Detection |
|-----------|------|-----------|
| `.opf` | OPF package document | Extension |
| `.xhtml` | XHTML content document | Extension |
| `.html` | HTML content document | Extension |
| `.css` | CSS stylesheet | Extension |
| `.ncx` | NCX navigation (EPUB 2) | Extension |

Navigation documents (`.xhtml`/`.html` containing `epub:type="toc"`) are detected via content sniffing and receive additional nav-specific validation.

## Validators

### OPF Package Document

- Required metadata: `dc:identifier`, `dc:title`, `dc:language`
- `unique-identifier` must reference a valid `dc:identifier/@id`
- Manifest integrity: unique IDs, valid media-types, no duplicate hrefs
- Spine itemrefs must reference existing manifest items

### XHTML Content Document

- XHTML namespace (`xmlns="http://www.w3.org/1999/xhtml"`) required
- `xml:lang` and `lang` consistency
- `<img>` elements must have `alt` attribute

### Navigation Document

- `<nav epub:type="toc">` required with `<ol>` child
- No remote links allowed in navigation
- Optional page-list and landmarks detection
- TOC link order vs spine order consistency

### CSS Stylesheet

- Forbidden properties: `direction`, `unicode-bidi`
- Position warnings: `fixed`, `absolute`
- `@font-face` format validation (woff, woff2, opentype, truetype)
- UTF-8 encoding check
- Unclosed brace detection

### Cross-File Resource Validation

- Manifest items reference files that exist in the workspace
- Resources referenced in content (`<img>`, `<link>`, `<audio>`, `<video>`, `<source>`) exist in the OPF manifest

### Accessibility (based on DAISY Ace rules)

- **Metadata**: `schema:accessMode`, `schema:accessibilityFeature`, `schema:accessibilityHazard`, `schema:accessibilitySummary`, `schema:accessModeSufficient` with value validation and contradictory hazard detection
- **OPF**: `dc:title` and `dc:language` presence
- **Page navigation**: `printPageNumbers` requires page-list nav and pagebreak markers; page-list requires `dc:source`; page-list references validated against content IDs
- **Structure**: `epub:type` to ARIA role mapping, pagebreak labels, heading level ordering, table captions, form input labels

## Architecture

```
cmd/epub-lsp/           LSP server (stdin/stdout JSON-RPC)
  lsp/                  Protocol types, message framing, handlers
internal/epub/          Core types (Diagnostic, FileType, Position)
  parser/               XML and CSS parsers with offset tracking
  testutil/             Shared test helpers
  validator/            Registry and Validator interface
    opf/                OPF package validation + OPF parsing
    xhtml/              XHTML namespace and structure checks
    nav/                Navigation document validation
    css/                CSS property and syntax checks
    resource/           Cross-file manifest and content reference checks
    accessibility/      Accessibility metadata, structure, and page checks
```

Validators register with a central `Registry` and are dispatched by file type. Files within a workspace are validated concurrently. Cross-file context (manifest items, spine order, file contents) is passed via `WorkspaceContext`.

## License

MIT License - see [LICENSE](LICENSE) for details.
