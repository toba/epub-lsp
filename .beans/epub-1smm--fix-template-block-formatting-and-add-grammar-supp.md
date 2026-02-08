---
# epub-1smm
title: Fix template block formatting and add grammar support
status: completed
type: feature
priority: normal
created_at: 2026-02-08T07:26:28Z
updated_at: 2026-02-08T07:29:07Z
---

Go template directives in EPUB XHTML files need two fixes:

1. **Formatter**: Multi-line charData containing template directives ({{if}}, {{end}}, etc.) should be split into individual lines with proper template-aware indentation.
2. **Grammar**: Add gotmpl tree-sitter grammar injection to gubby extension for template syntax highlighting in Zed.

## Checklist

- [x] Add regex patterns for template block detection in xml.go
- [x] Rewrite tokCharData case to split multi-line content and adjust depth for template blocks
- [x] Add formatter tests for template block nesting
- [x] Add gotmpl grammar to gubby extension.toml
- [x] Create gotmpl language config, highlights.scm, brackets.scm
- [x] Create epub-xhtml injections.scm for gotmpl injection
- [x] Run tests and lint