---
# til-z3n
title: Migrate to toba/lsp server harness
status: completed
type: task
priority: high
created_at: 2026-03-21T17:45:43Z
updated_at: 2026-03-21T18:06:18Z
sync:
    github:
        issue_number: "12"
        synced_at: "2026-03-21T18:06:34Z"
---

Replace the hand-rolled LSP main loop with the new `github.com/toba/lsp/server` package (v0.2.0+).

## What changes

The `server` package handles all lifecycle boilerplate:
- JSON-RPC transport via `go.lsp.dev/jsonrpc2`
- `initialize` / `initialized` / `shutdown` / `exit` lifecycle
- Document state management (open/change/close)
- Diagnostic publishing with debouncing
- Optional handler delegation (Hover, Completion, Definition, Formatting, CodeAction, References, Rename, DocumentSymbol)

## Steps

- [ ] Add `github.com/toba/lsp v0.2.0` dependency
- [ ] Implement `server.Handler` interface (Initialize, Diagnostics, Shutdown)
- [ ] Implement any optional handler interfaces (e.g. `server.HoverHandler`, `server.CompletionHandler`)
- [ ] Replace main loop with `server.Server{Name: "epub-lsp", Version: version, Handler: h}.Run(ctx)`
- [ ] Remove hand-rolled JSON-RPC dispatch, document store, and diagnostic goroutine
- [ ] Remove direct dependency on `toba/lsp/transport` if no longer needed
- [ ] Run tests and linter
- [ ] Verify in editor (VS Code or Zed)

## Summary of Changes\n\nReplaced the hand-rolled LSP main loop (transport.Scanner + manual JSON-RPC dispatch) with the toba/lsp/server harness. The epubHandler type implements server.Handler (Initialize, Diagnostics, Shutdown) plus all optional handler interfaces (HoverHandler, CompletionHandler, DefinitionHandler, FormattingHandler, CodeActionHandler, ReferencesHandler, DocumentSymbolHandler, DocumentLinkHandler, SemanticTokensFullHandler). Existing handler functions in cmd/epub-lsp/lsp/ are preserved via JSON round-trip bridging. Added DocumentLinkHandler and SemanticTokensFullHandler interfaces to the toba/lsp/server package. Updated go.mod to toba/lsp v0.2.1 with go.lsp.dev/protocol dependency.
