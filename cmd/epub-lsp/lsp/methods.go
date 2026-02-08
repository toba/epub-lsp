package lsp

import (
	"encoding/json"
	"errors"
	"log/slog"
	"strconv"

	"github.com/toba/epub-lsp/internal/epub"
	"github.com/toba/epub-lsp/internal/epub/validator"
)

// WorkspaceReader provides read access to workspace state for LSP handlers.
type WorkspaceReader interface {
	GetContent(uri string) []byte
	GetFileType(uri string) epub.FileType
	GetManifest() *validator.ManifestInfo
	GetDiagnostics(uri string) []epub.Diagnostic
	GetAllFiles() map[string][]byte
	GetRootPath() string
	GetSettings() *ServerSettings
}

// ID represents a JSON-RPC request ID that can be either a string or number.
type ID int

func (id *ID) UnmarshalJSON(data []byte) error {
	length := len(data)
	if data[0] == '"' && data[length-1] == '"' {
		data = data[1 : length-1]
	}

	number, err := strconv.Atoi(string(data))
	if err != nil {
		return errors.New("'ID' expected either a string or an integer")
	}

	*id = ID(number)
	return nil
}

func (id *ID) MarshalJSON() ([]byte, error) {
	val := strconv.Itoa(int(*id))
	return []byte(val), nil
}

// RequestMessage represents a JSON-RPC request.
type RequestMessage[T any] struct {
	JsonRpc string `json:"jsonrpc"`
	Id      ID     `json:"id"`
	Method  string `json:"method"`
	Params  T      `json:"params"`
}

// ResponseMessage represents a JSON-RPC response.
type ResponseMessage[T any] struct {
	JsonRpc string         `json:"jsonrpc"`
	Id      ID             `json:"id"`
	Result  T              `json:"result"`
	Error   *ResponseError `json:"error"`
}

// ResponseError represents a JSON-RPC error.
type ResponseError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// NotificationMessage represents a JSON-RPC notification (no response expected).
type NotificationMessage[T any] struct {
	JsonRpc string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  T      `json:"params"`
}

// ServerSettings holds configuration options sent by the editor.
type ServerSettings struct {
	Accessibility string `json:"accessibility"`
}

// InitializeParams holds parameters for the initialize request.
type InitializeParams struct {
	ProcessId             int             `json:"processId"`
	Capabilities          map[string]any  `json:"capabilities"`
	ClientInfo            ClientInfo      `json:"clientInfo"`
	RootUri               string          `json:"rootUri"`
	WorkspaceFolders      any             `json:"workspaceFolders"`
	InitializationOptions *ServerSettings `json:"initializationOptions"`
}

// ClientInfo describes the connecting editor.
type ClientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// CodeActionOptions describes code action capabilities.
type CodeActionOptions struct {
	CodeActionKinds []string `json:"codeActionKinds,omitempty"`
}

// ServerCapabilities describes the capabilities this server supports.
type ServerCapabilities struct {
	TextDocumentSync           int                    `json:"textDocumentSync"`
	DocumentLinkProvider       bool                   `json:"documentLinkProvider,omitempty"`
	DocumentSymbolProvider     bool                   `json:"documentSymbolProvider,omitempty"`
	DefinitionProvider         bool                   `json:"definitionProvider,omitempty"`
	ReferencesProvider         bool                   `json:"referencesProvider,omitempty"`
	HoverProvider              bool                   `json:"hoverProvider,omitempty"`
	CodeActionProvider         *CodeActionOptions     `json:"codeActionProvider,omitempty"`
	CompletionProvider         *CompletionOptions     `json:"completionProvider,omitempty"`
	DocumentFormattingProvider bool                   `json:"documentFormattingProvider,omitempty"`
	SemanticTokensProvider     *SemanticTokensOptions `json:"semanticTokensProvider,omitempty"`
}

// SemanticTokensLegend describes the token types and modifiers used by semantic tokens.
type SemanticTokensLegend struct {
	TokenTypes     []string `json:"tokenTypes"`
	TokenModifiers []string `json:"tokenModifiers"`
}

// SemanticTokensOptions describes semantic token capabilities.
type SemanticTokensOptions struct {
	Legend SemanticTokensLegend `json:"legend"`
	Full   bool                 `json:"full"`
}

// SemanticTokensParams holds parameters for textDocument/semanticTokens/full.
type SemanticTokensParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

// SemanticTokensResult holds the encoded semantic tokens.
type SemanticTokensResult struct {
	Data []uint `json:"data"`
}

// SemanticTokenTypes defines the token type legend for Go template syntax.
var SemanticTokenTypes = []string{
	"keyword",  // 0
	"variable", // 1
	"function", // 2
	"property", // 3
	"string",   // 4
	"number",   // 5
	"operator", // 6
	"comment",  // 7
}

// SemanticTokenModifiers defines the token modifier legend (currently empty).
var SemanticTokenModifiers = []string{}

// CompletionOptions describes completion capabilities.
type CompletionOptions struct {
	TriggerCharacters []string `json:"triggerCharacters,omitempty"`
}

// InitializeResult is the response to the initialize request.
type InitializeResult struct {
	Capabilities ServerCapabilities `json:"capabilities"`
	ServerInfo   ServerInfo         `json:"serverInfo"`
}

// ServerInfo describes this server.
type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// PublishDiagnosticsParams holds parameters for publishing diagnostics.
type PublishDiagnosticsParams struct {
	Uri         string       `json:"uri"`
	Diagnostics []Diagnostic `json:"diagnostics"`
}

// Diagnostic represents a diagnostic message.
type Diagnostic struct {
	Range    Range  `json:"range"`
	Message  string `json:"message"`
	Severity int    `json:"severity"`
	Code     string `json:"code,omitempty"`
	Source   string `json:"source,omitempty"`
}

// Position represents a position in a text document.
type Position struct {
	Line      uint `json:"line"`
	Character uint `json:"character"`
}

// Range represents a range in a text document.
type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

// TextDocumentItem represents a text document.
type TextDocumentItem struct {
	Uri        string `json:"uri"`
	Version    int    `json:"version"`
	LanguageId string `json:"languageId"`
	Text       string `json:"text"`
}

// TextDocumentIdentifier identifies a text document.
type TextDocumentIdentifier struct {
	Uri string `json:"uri"`
}

// ProcessInitializeRequest handles the initialize request.
func ProcessInitializeRequest(
	data []byte,
	lspName, lspVersion string,
) (response []byte, rootURI string, settings *ServerSettings) {
	req := RequestMessage[InitializeParams]{}

	err := json.Unmarshal(data, &req)
	if err != nil {
		msg := "error while unmarshalling data during 'initialize' phase: " + err.Error()
		slog.Error(msg)
		panic(msg)
	}

	res := ResponseMessage[InitializeResult]{
		JsonRpc: JSONRPCVersion,
		Id:      req.Id,
		Result: InitializeResult{
			Capabilities: ServerCapabilities{
				TextDocumentSync:       TextDocumentSyncFull,
				DocumentLinkProvider:   true,
				DocumentSymbolProvider: true,
				DefinitionProvider:     true,
				ReferencesProvider:     true,
				HoverProvider:          true,
				CodeActionProvider: &CodeActionOptions{
					CodeActionKinds: []string{"quickfix", "source.fixAll"},
				},
				CompletionProvider: &CompletionOptions{
					TriggerCharacters: []string{"<", "\"", ":", " "},
				},
				DocumentFormattingProvider: true,
				SemanticTokensProvider: &SemanticTokensOptions{
					Legend: SemanticTokensLegend{
						TokenTypes:     SemanticTokenTypes,
						TokenModifiers: SemanticTokenModifiers,
					},
					Full: true,
				},
			},
			ServerInfo: ServerInfo{
				Name:    lspName,
				Version: lspVersion,
			},
		},
	}

	response, err = json.Marshal(res)
	if err != nil {
		msg := "error while marshalling data during 'initialize' phase: " + err.Error()
		slog.Error(msg)
		panic(msg)
	}

	return response, req.Params.RootUri, req.Params.InitializationOptions
}

// ProcessShutdownRequest handles the shutdown request.
func ProcessShutdownRequest(jsonVersion string, requestId ID) []byte {
	response := ResponseMessage[any]{
		JsonRpc: jsonVersion,
		Id:      requestId,
		Result:  nil,
		Error:   nil,
	}

	responseText, err := json.Marshal(response)
	if err != nil {
		msg := "error while marshalling shutdown response: " + err.Error()
		slog.Error(msg)
		panic(msg)
	}

	return responseText
}

// ProcessIllegalRequestAfterShutdown returns an error for requests after shutdown.
func ProcessIllegalRequestAfterShutdown(jsonVersion string, requestId ID) []byte {
	response := ResponseMessage[any]{
		JsonRpc: jsonVersion,
		Id:      requestId,
		Result:  nil,
		Error: &ResponseError{
			Code:    ErrorInvalidRequest,
			Message: "illegal request while server shutting down",
		},
	}

	responseText, err := json.Marshal(response)
	if err != nil {
		msg := "error while marshalling error response: " + err.Error()
		slog.Error(msg)
		panic(msg)
	}

	return responseText
}

// DidOpenTextDocumentParams holds parameters for textDocument/didOpen.
type DidOpenTextDocumentParams struct {
	TextDocument TextDocumentItem `json:"textDocument"`
}

// ProcessDidOpenTextDocumentNotification handles textDocument/didOpen.
func ProcessDidOpenTextDocumentNotification(
	data []byte,
) (fileURI string, fileContent []byte) {
	request := RequestMessage[DidOpenTextDocumentParams]{}

	err := json.Unmarshal(data, &request)
	if err != nil {
		msg := "error while unmarshalling 'textDocument/didOpen': " + err.Error()
		slog.Error(msg)
		panic(msg)
	}

	return request.Params.TextDocument.Uri, []byte(request.Params.TextDocument.Text)
}

// TextDocumentContentChangeEvent represents a content change event.
type TextDocumentContentChangeEvent struct {
	Text string `json:"text"`
}

// DidChangeTextDocumentParams holds parameters for textDocument/didChange.
type DidChangeTextDocumentParams struct {
	TextDocument   TextDocumentItem                 `json:"textDocument"`
	ContentChanges []TextDocumentContentChangeEvent `json:"contentChanges"`
}

// ProcessDidChangeTextDocumentNotification handles textDocument/didChange.
func ProcessDidChangeTextDocumentNotification(
	data []byte,
) (fileURI string, fileContent []byte) {
	var request RequestMessage[DidChangeTextDocumentParams]

	err := json.Unmarshal(data, &request)
	if err != nil {
		msg := "error while unmarshalling 'textDocument/didChange': " + err.Error()
		slog.Error(msg)
		panic(msg)
	}

	changes := request.Params.ContentChanges
	if len(changes) == 0 {
		slog.Warn("'contentChanges' field is empty")
		return "", nil
	}

	return request.Params.TextDocument.Uri, []byte(changes[0].Text)
}

// --- New LSP types for interactive features ---

// DocumentLinkParams holds parameters for textDocument/documentLink.
type DocumentLinkParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

// DocumentLink represents a link in a document.
type DocumentLink struct {
	Range  Range  `json:"range"`
	Target string `json:"target"`
}

// DocumentSymbolParams holds parameters for textDocument/documentSymbol.
type DocumentSymbolParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

// SymbolKind identifies the kind of a symbol.
type SymbolKind int

// Symbol kind constants.
const (
	SymbolKindFile        SymbolKind = 1
	SymbolKindModule      SymbolKind = 2
	SymbolKindNamespace   SymbolKind = 3
	SymbolKindPackage     SymbolKind = 4
	SymbolKindClass       SymbolKind = 5
	SymbolKindMethod      SymbolKind = 6
	SymbolKindProperty    SymbolKind = 7
	SymbolKindField       SymbolKind = 8
	SymbolKindConstructor SymbolKind = 9
	SymbolKindEnum        SymbolKind = 10
	SymbolKindInterface   SymbolKind = 11
	SymbolKindFunction    SymbolKind = 12
	SymbolKindVariable    SymbolKind = 13
	SymbolKindConstant    SymbolKind = 14
	SymbolKindString      SymbolKind = 15
	SymbolKindNumber      SymbolKind = 16
	SymbolKindBoolean     SymbolKind = 17
	SymbolKindArray       SymbolKind = 18
	SymbolKindObject      SymbolKind = 19
	SymbolKindKey         SymbolKind = 20
	SymbolKindNull        SymbolKind = 21
	SymbolKindStruct      SymbolKind = 23
	SymbolKindEvent       SymbolKind = 24
)

// DocumentSymbol represents a symbol in a document.
type DocumentSymbol struct {
	Name           string           `json:"name"`
	Detail         string           `json:"detail,omitempty"`
	Kind           SymbolKind       `json:"kind"`
	Range          Range            `json:"range"`
	SelectionRange Range            `json:"selectionRange"`
	Children       []DocumentSymbol `json:"children,omitempty"`
}

// DefinitionParams holds parameters for textDocument/definition.
type DefinitionParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
}

// Location represents a location in a document.
type Location struct {
	URI   string `json:"uri"`
	Range Range  `json:"range"`
}

// ReferenceParams holds parameters for textDocument/references.
type ReferenceParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
	Context      ReferenceContext       `json:"context"`
}

// ReferenceContext controls what references are returned.
type ReferenceContext struct {
	IncludeDeclaration bool `json:"includeDeclaration"`
}

// HoverParams holds parameters for textDocument/hover.
type HoverParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
}

// Hover represents hover information.
type Hover struct {
	Contents MarkupContent `json:"contents"`
	Range    *Range        `json:"range,omitempty"`
}

// MarkupContent represents documentation content.
type MarkupContent struct {
	Kind  string `json:"kind"`
	Value string `json:"value"`
}

// CodeActionParams holds parameters for textDocument/codeAction.
type CodeActionParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Range        Range                  `json:"range"`
	Context      CodeActionContext      `json:"context"`
}

// CodeActionContext carries code action request context.
type CodeActionContext struct {
	Diagnostics []Diagnostic `json:"diagnostics"`
	Only        []string     `json:"only,omitempty"`
}

// CodeAction represents a code action.
type CodeAction struct {
	Title       string         `json:"title"`
	Kind        string         `json:"kind,omitempty"`
	Diagnostics []Diagnostic   `json:"diagnostics,omitempty"`
	Edit        *WorkspaceEdit `json:"edit,omitempty"`
}

// WorkspaceEdit represents changes to workspace resources.
type WorkspaceEdit struct {
	Changes map[string][]TextEdit `json:"changes"`
}

// TextEdit represents a text edit.
type TextEdit struct {
	Range   Range  `json:"range"`
	NewText string `json:"newText"`
}

// CompletionParams holds parameters for textDocument/completion.
type CompletionParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
}

// CompletionItem represents a completion suggestion.
type CompletionItem struct {
	Label         string `json:"label"`
	Kind          int    `json:"kind,omitempty"`
	Detail        string `json:"detail,omitempty"`
	Documentation string `json:"documentation,omitempty"`
	InsertText    string `json:"insertText,omitempty"`
}

// Completion kind constants.
const (
	CompletionKindText     = 1
	CompletionKindProperty = 10
	CompletionKindValue    = 12
	CompletionKindEnum     = 13
	CompletionKindKeyword  = 14
)

// CompletionList represents a list of completion items.
type CompletionList struct {
	IsIncomplete bool             `json:"isIncomplete"`
	Items        []CompletionItem `json:"items"`
}

// DocumentFormattingParams holds parameters for textDocument/formatting.
type DocumentFormattingParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Options      FormattingOptions      `json:"options"`
}

// FormattingOptions describes formatting options.
type FormattingOptions struct {
	TabSize      int  `json:"tabSize"`
	InsertSpaces bool `json:"insertSpaces"`
}
