---
# 28o-ieo
title: Adopt shared github.com/toba/lsp module
status: completed
type: feature
priority: normal
created_at: 2026-03-21T17:18:33Z
updated_at: 2026-03-21T17:32:11Z
sync:
    github:
        issue_number: "10"
        synced_at: "2026-03-21T17:45:57Z"
---

Replace duplicated LSP infrastructure code with the shared github.com/toba/lsp module. This eliminates ~500-700 lines of copy-pasted code that is identical across all four toba LSP projects.

## Packages to adopt

- `github.com/toba/lsp/transport` — replaces hand-rolled `parsing.go` (ReceiveInput, decode, Encode, SendToLspClient, SendOutput)
- `github.com/toba/lsp/logging` — replaces createLogFile(), configureLogging(), MaxLogFileSize/DirPermissions/FilePermissions constants
- `github.com/toba/lsp/pathutil` — replaces uriToFilePath(), filePathToUri(), convertKeysFromFilePathToUri()
- `github.com/toba/lsp/position` — replaces OffsetToLineChar/offsetToLineCol, LineCharToOffset, intToUint, uintToInt

## Steps

1. Add `github.com/toba/lsp` dependency
2. Replace transport layer: `lsp.ReceiveInput` → `transport.NewScanner`, `lsp.SendToLspClient` → `transport.Send`, `lsp.Encode` → `transport.Encode`
3. Replace logging: `createLogFile()`/`configureLogging()` → `logging.Configure(appName)`
4. Replace path utilities: `uriToFilePath` → `pathutil.URIToFilePath`, `filePathToUri` → `pathutil.FilePathToURI`
5. Replace position utilities: `OffsetToLineChar` → `position.OffsetToLineCol`, etc.
6. Delete the replaced local code (parsing.go, logging functions, URI functions, position functions)
7. Remove now-unused constants (ContentLengthHeader, HeaderDelimiter, etc.)
8. Run tests and linter


## Summary of Changes

Replaced duplicated LSP infrastructure code with the shared `github.com/toba/lsp` module (v0.1.0):

- **transport**: Deleted `parsing.go`, replaced `ReceiveInput`/`SendToLspClient`/`Encode` with `transport.NewScanner`/`transport.Send`/`transport.Encode`
- **logging**: Replaced `configureLogging()`/`createLogFile()` with `logging.Configure(serverName)`
- **pathutil**: Replaced local `uriToFilePath()` with `pathutil.URIToFilePath`
- **position**: Replaced local `intToUint()` across 4 files with `position.IntToUint`
- Removed dead constants: `ContentLengthHeader`, `HeaderDelimiter`, `LineDelimiter`, `DirPermissions`, `FilePermissions`, `MaxLogFileSize`
- Updated `parsing_test.go` to test via `transport` package
- All tests pass, lint clean
