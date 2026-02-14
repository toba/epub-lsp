---
# epub-obrt
title: Add 8 high-value LSP capabilities
status: completed
type: epic
priority: normal
created_at: 2026-02-07T23:10:51Z
updated_at: 2026-02-07T23:24:55Z
---

Add interactive LSP features: documentLink, documentSymbol, definition, references, hover, codeAction, completion, formatting. Includes shared infrastructure (WorkspaceReader, LocateAtPosition, PositionToByteOffset).

## Checklist
- [x] Shared infrastructure (WorkspaceReader, PositionToByteOffset, LocateAtPosition, ServerCapabilities, method constants)
- [x] F1: textDocument/documentLink
- [x] F2: textDocument/documentSymbol
- [x] F3: textDocument/definition
- [x] F4: textDocument/references
- [x] F5: textDocument/hover
- [x] F6: textDocument/codeAction
- [x] F7: textDocument/completion
- [x] F8: textDocument/formatting