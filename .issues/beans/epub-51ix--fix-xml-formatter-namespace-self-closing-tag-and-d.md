---
# epub-51ix
title: Fix XML formatter namespace, self-closing tag, and DOCTYPE handling
status: completed
type: bug
priority: normal
created_at: 2026-02-08T04:08:51Z
updated_at: 2026-02-08T07:05:55Z
---

The FormatXML function uses Go's encoding/xml.Encoder for output, which has fundamental issues:

1. Repeats xmlns= on every child element
2. Mangles prefixed namespace declarations (xmlns:epub → xmlns:_xmlns _xmlns:epub)
3. Never produces self-closing tags (<meta/> → <meta></meta>)
4. Doesn't add newline after <!DOCTYPE html>

## Root Cause
encoding/xml.Encoder doesn't handle namespace declarations properly and doesn't support void/self-closing elements.

## Fix
Rewrite FormatXML to use xml.Decoder for tokenizing but write output manually with proper:
- Namespace declaration preservation (no duplication, no mangling)
- Self-closing tag support for void XHTML elements
- DOCTYPE newline handling

## Checklist
- [ ] Create failing test demonstrating the issues
- [ ] Rewrite FormatXML to emit output manually instead of using xml.Encoder
- [ ] Handle namespace declarations properly (preserve original, no duplication)
- [ ] Support self-closing tags for void elements
- [ ] Add newline after DOCTYPE directives
- [ ] Run tests and golangci-lint