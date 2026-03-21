---
# dsg-uj7
title: Migrate to go.lsp.dev/protocol + go.lsp.dev/jsonrpc2
status: completed
type: feature
priority: normal
created_at: 2026-03-21T17:11:18Z
updated_at: 2026-03-21T18:06:23Z
sync:
    github:
        issue_number: "11"
        synced_at: "2026-03-21T18:06:34Z"
---

Replace the hand-rolled LSP protocol types and JSON-RPC transport with the standard go.lsp.dev/protocol and go.lsp.dev/jsonrpc2 packages. This eliminates maintaining custom LSP struct definitions and transport code, and gives full spec-compliant types for all LSP methods.

## Steps
- Add go.lsp.dev/protocol and go.lsp.dev/jsonrpc2 as dependencies
- Replace all custom LSP protocol types with imports from go.lsp.dev/protocol
- Replace the custom JSON-RPC transport with go.lsp.dev/jsonrpc2 stream handling
- Update all handler functions to use the standard protocol types
- Remove custom type definitions that are now redundant
- Run tests and linter

## Summary of Changes\n\nSuperseded by til-z3n (server harness migration). The server harness uses go.lsp.dev/protocol and go.lsp.dev/jsonrpc2 internally, so this migration is complete.
