---
# epub-ulle
title: Add syntax highlighting for EPUB files
status: completed
type: feature
priority: normal
created_at: 2026-02-07T23:56:01Z
updated_at: 2026-02-08T00:01:03Z
---

## Summary
Add tree-sitter XML grammar to gubby extension for baseline XML coloring, and semantic tokens to epub-lsp for Go template highlighting.

## Checklist
- [x] Part 1a: Add tree-sitter-xml grammar to gubby extension.toml
- [x] Part 1b: Update language config.toml files with grammar, brackets, comments
- [x] Part 1c: Add highlights.scm, indents.scm, outline.scm to each language dir
- [x] Part 2a: Add semantic token types and capabilities to protocol.go
- [x] Part 2b: Add handler dispatch in main.go
- [x] Part 2c: Implement semantic tokens handler and tokenizer
- [x] Part 2d: Add tests for semantic tokens
- [x] Run tests and linter