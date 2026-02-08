package lsp

import (
	"encoding/json"
	"maps"
	"testing"

	"github.com/toba/epub-lsp/internal/epub"
	"github.com/toba/epub-lsp/internal/epub/validator"
)

// mockWorkspace implements WorkspaceReader for tests.
type mockWorkspace struct {
	files       map[string][]byte
	fileTypes   map[string]epub.FileType
	diagnostics map[string][]epub.Diagnostic
	manifest    *validator.ManifestInfo
	rootPath    string
	settings    *ServerSettings
}

func (m *mockWorkspace) GetContent(
	uri string,
) []byte {
	return m.files[uri]
}

func (m *mockWorkspace) GetFileType(
	uri string,
) epub.FileType {
	return m.fileTypes[uri]
}
func (m *mockWorkspace) GetManifest() *validator.ManifestInfo { return m.manifest }

func (m *mockWorkspace) GetDiagnostics(
	uri string,
) []epub.Diagnostic {
	return m.diagnostics[uri]
}
func (m *mockWorkspace) GetRootPath() string          { return m.rootPath }
func (m *mockWorkspace) GetSettings() *ServerSettings { return m.settings }
func (m *mockWorkspace) GetAllFiles() map[string][]byte {
	result := make(map[string][]byte, len(m.files))
	maps.Copy(result, m.files)
	return result
}

func newMockWorkspace() *mockWorkspace {
	return &mockWorkspace{
		files:       make(map[string][]byte),
		fileTypes:   make(map[string]epub.FileType),
		diagnostics: make(map[string][]epub.Diagnostic),
	}
}

// makeRequest creates a JSON-RPC request for a given method and params.
func makeRequest[T any](t *testing.T, id int, method string, params T) []byte {
	t.Helper()
	req := RequestMessage[T]{
		JsonRpc: JSONRPCVersion,
		Id:      ID(id),
		Method:  method,
		Params:  params,
	}
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

// unmarshalResult extracts the result from a JSON-RPC response.
func unmarshalResult[T any](t *testing.T, data []byte) T {
	t.Helper()
	var resp ResponseMessage[T]
	if err := json.Unmarshal(data, &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v\nraw: %s", err, data)
	}
	return resp.Result
}
