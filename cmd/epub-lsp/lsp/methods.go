package lsp

import (
	"encoding/json"
	"errors"
	"log/slog"
	"strconv"
)

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

// InitializeParams holds parameters for the initialize request.
type InitializeParams struct {
	ProcessId        int            `json:"processId"`
	Capabilities     map[string]any `json:"capabilities"`
	ClientInfo       ClientInfo     `json:"clientInfo"`
	RootUri          string         `json:"rootUri"`
	WorkspaceFolders any            `json:"workspaceFolders"`
}

// ClientInfo describes the connecting editor.
type ClientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// ServerCapabilities describes the capabilities this server supports.
type ServerCapabilities struct {
	TextDocumentSync int `json:"textDocumentSync"`
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
) (response []byte, rootURI string) {
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
				TextDocumentSync: TextDocumentSyncFull,
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

	return response, req.Params.RootUri
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
