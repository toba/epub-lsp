---
# epub-n650
title: Implement Go optimization plan
status: completed
type: task
priority: normal
created_at: 2026-02-07T22:44:25Z
updated_at: 2026-02-07T22:58:22Z
---

Consolidate shared code, eliminate magic strings, and add safe concurrency for file validation across the epub-lsp codebase. 8 steps covering test helpers, namespace constants, URL utilities, diagnostic builder, OPF parse helper, LSP severity cleanup, form label consolidation, and concurrent validation.