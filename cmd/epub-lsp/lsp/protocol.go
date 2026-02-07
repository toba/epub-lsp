// Package lsp implements LSP message types and handlers for EPUB validation.
package lsp

// LSP protocol constants.
const (
	JSONRPCVersion = "2.0"

	SeverityError   = 1
	SeverityWarning = 2
	SeverityInfo    = 3
	SeverityHint    = 4

	TextDocumentSyncFull = 1

	ErrorInvalidRequest = -32600
)

// LSP method names.
const (
	MethodInitialize         = "initialize"
	MethodInitialized        = "initialized"
	MethodShutdown           = "shutdown"
	MethodExit               = "exit"
	MethodDidOpen            = "textDocument/didOpen"
	MethodDidChange          = "textDocument/didChange"
	MethodDidClose           = "textDocument/didClose"
	MethodPublishDiagnostics = "textDocument/publishDiagnostics"
)

// LSP header constants.
const (
	ContentLengthHeader = "Content-Length"
	HeaderDelimiter     = "\r\n\r\n"
	LineDelimiter       = "\r\n"
)

// File and logging constants.
const (
	DirPermissions  = 0750
	FilePermissions = 0600
	MaxLogFileSize  = 5_000_000 // 5MB
)
