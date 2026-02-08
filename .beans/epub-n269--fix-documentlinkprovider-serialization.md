---
# epub-n269
title: Fix documentLinkProvider serialization
status: completed
type: bug
priority: normal
created_at: 2026-02-08T00:05:52Z
updated_at: 2026-02-08T00:06:31Z
---

Zed/gubby expects DocumentLinkOptions struct but epub-lsp sends boolean true. Change DocumentLinkProvider from bool to struct pointer.