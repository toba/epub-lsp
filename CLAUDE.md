# EPUB LSP

Language Server Protocol implementation for EPUB validation. Pure Go, no external dependencies.

## Guidelines

- Be concise
- When fixing or investigating code issues, ALWAYS create a failing test FIRST to demonstrate understanding of the problem THEN change code and confirm the test passes
- Run `golangci-lint run --fix` after modifying Go code
- Run `go test ./...` after changes
- **NEVER commit without explicit user request**

## Building

```bash
go build ./cmd/epub-lsp
```

## Testing

```bash
go test ./...
```

## Project Layout

- `cmd/epub-lsp/` - LSP server entry point and JSON-RPC message loop
- `cmd/epub-lsp/lsp/` - Protocol types, message framing, request handlers
- `internal/epub/` - Core types: `Diagnostic`, `FileType`, `Position`, `DiagBuilder`, namespace constants, URL utilities
- `internal/epub/parser/` - XML parser (namespace-aware, offset-tracking) and CSS tokenizer
- `internal/epub/testutil/` - Shared test helpers (`HasCode`, `DiagCodes`, `ExpectCode`, `SeverityName`)
- `internal/epub/validator/` - `Registry`, `Validator` interface, `WorkspaceContext`
- `internal/epub/validator/opf/` - OPF package validation (metadata, manifest, spine) and `ParseOPFMetadata`/`ParseManifest` helpers
- `internal/epub/validator/xhtml/` - XHTML namespace and structure checks
- `internal/epub/validator/nav/` - Navigation document validation
- `internal/epub/validator/css/` - CSS property and syntax checks
- `internal/epub/validator/resource/` - Cross-file manifest and content reference checks
- `internal/epub/validator/accessibility/` - Accessibility metadata, structure, pages, and OPF checks

## Key Patterns

- Validators implement the `Validator` interface and register with `validator.Registry`
- `DiagBuilder` fluent API: `epub.NewDiag(content, offset, source).Code("X").Error("msg").Build()`
- Namespace constants live in `internal/epub/namespace.go` (`NSEpub`, `NSDC`, `NSXHTML`, `NSXML`)
- URL helpers live in `internal/epub/urlutil.go` (`IsRemoteURL`, `StripFragment`, `ContainsToken`)
- Test helpers live in `internal/epub/testutil/` - use these instead of defining per-package helpers
- Cross-file data flows through `WorkspaceContext` (manifest info, file map, file types)
- Files are validated concurrently per workspace change via `sync.WaitGroup`

## Releasing

Pushing a version tag (`v*`) triggers GoReleaser via GitHub Actions, producing binaries for linux/darwin/windows on amd64/arm64. The companion Zed extension lives in the sibling `gubby` repo.
