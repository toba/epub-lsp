// Command epub-lsp provides a Language Server Protocol server for EPUB validation.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"maps"
	"os"
	"strings"
	"sync"

	"go.lsp.dev/protocol"

	"github.com/toba/epub-lsp/cmd/epub-lsp/lsp"
	"github.com/toba/epub-lsp/internal/epub"
	"github.com/toba/epub-lsp/internal/epub/validator"
	"github.com/toba/epub-lsp/internal/epub/validator/accessibility"
	"github.com/toba/epub-lsp/internal/epub/validator/css"
	"github.com/toba/epub-lsp/internal/epub/validator/nav"
	"github.com/toba/epub-lsp/internal/epub/validator/opf"
	"github.com/toba/epub-lsp/internal/epub/validator/resource"
	"github.com/toba/epub-lsp/internal/epub/validator/xhtml"
	"github.com/toba/lsp/pathutil"
	"github.com/toba/lsp/server"
)

// version is set by goreleaser at build time.
var version = "dev"

const serverName = "epub-lsp"

// TargetFileExtensions lists the file extensions this LSP supports.
var TargetFileExtensions = []string{
	"opf", "xhtml", "html", "css", "ncx",
}

func main() {
	versionFlag := flag.Bool("version", false, "print the LSP version")
	flag.Parse()

	if *versionFlag {
		fmt.Printf("%s -- version %s\n", serverName, version)
		os.Exit(0)
	}

	registry := validator.NewRegistry()
	registry.Register(&opf.Validator{})
	registry.Register(&xhtml.Validator{})
	registry.Register(&nav.Validator{})
	registry.Register(&css.Validator{})
	registry.Register(&resource.ManifestValidator{})
	registry.Register(&resource.ContentValidator{})
	registry.Register(&accessibility.MetadataValidator{})
	registry.Register(&accessibility.PageValidator{})
	registry.Register(&accessibility.OPFAccessibilityValidator{})
	registry.Register(&accessibility.StructureValidator{})

	handler := &epubHandler{
		registry: registry,
		store: &workspaceStore{
			RawFiles:    make(map[string][]byte),
			FileTypes:   make(map[string]epub.FileType),
			Diagnostics: make(map[string][]epub.Diagnostic),
		},
	}

	s := &server.Server{
		Name:    serverName,
		Version: version,
		Handler: handler,
	}

	if err := s.Run(context.Background()); err != nil {
		slog.Error("server error", "err", err)
		os.Exit(1)
	}
}

// epubHandler implements server.Handler and optional handler interfaces.
type epubHandler struct {
	registry *validator.Registry
	store    *workspaceStore
}

// workspaceStore holds the state for a workspace.
type workspaceStore struct {
	mu          sync.RWMutex
	RootPath    string
	RawFiles    map[string][]byte
	FileTypes   map[string]epub.FileType
	Diagnostics map[string][]epub.Diagnostic
	Manifest    *validator.ManifestInfo
	Settings    *lsp.ServerSettings
}

func (s *workspaceStore) GetContent(uri string) []byte {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.RawFiles[uri]
}

func (s *workspaceStore) GetFileType(uri string) epub.FileType {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.FileTypes[uri]
}

func (s *workspaceStore) GetManifest() *validator.ManifestInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Manifest
}

func (s *workspaceStore) GetDiagnostics(uri string) []epub.Diagnostic {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Diagnostics[uri]
}

func (s *workspaceStore) GetAllFiles() map[string][]byte {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make(map[string][]byte, len(s.RawFiles))
	maps.Copy(result, s.RawFiles)
	return result
}

func (s *workspaceStore) GetRootPath() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.RootPath
}

func (s *workspaceStore) GetSettings() *lsp.ServerSettings {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Settings
}

// --- server.Handler ---

func (h *epubHandler) Initialize(
	_ context.Context,
	params *protocol.InitializeParams,
) (protocol.ServerCapabilities, error) {
	// Extract root URI from workspace folders, falling back to RootURI.
	var rootURI string
	if len(params.WorkspaceFolders) > 0 {
		rootURI = params.WorkspaceFolders[0].URI
	} else {
		rootURI = string(
			params.RootURI,
		) //nolint:staticcheck // fallback for older clients
	}
	h.store.mu.Lock()
	h.store.RootPath = pathutil.URIToFilePath(rootURI)

	// Extract settings from initialization options
	if params.InitializationOptions != nil {
		raw, err := json.Marshal(params.InitializationOptions)
		if err == nil {
			var settings lsp.ServerSettings
			if json.Unmarshal(raw, &settings) == nil {
				h.store.Settings = &settings
			}
		}
	}
	h.store.mu.Unlock()

	return protocol.ServerCapabilities{
		DocumentLinkProvider:   &protocol.DocumentLinkOptions{},
		DocumentSymbolProvider: true,
		DefinitionProvider:     true,
		ReferencesProvider:     true,
		HoverProvider:          true,
		CodeActionProvider: &protocol.CodeActionOptions{
			CodeActionKinds: []protocol.CodeActionKind{
				protocol.QuickFix,
				"source.fixAll",
			},
		},
		CompletionProvider: &protocol.CompletionOptions{
			TriggerCharacters: []string{"<", "\"", ":", " "},
		},
		DocumentFormattingProvider: true,
		SemanticTokensProvider: map[string]any{
			"legend": map[string]any{
				"tokenTypes":     lsp.SemanticTokenTypes,
				"tokenModifiers": lsp.SemanticTokenModifiers,
			},
			"full": true,
		},
	}, nil
}

func (h *epubHandler) Diagnostics(
	_ context.Context,
	uri protocol.DocumentURI,
	content string,
) ([]protocol.Diagnostic, error) {
	uriStr := string(uri)

	if !hasTargetExtension(uriStr) {
		return nil, nil
	}

	contentBytes := []byte(content)

	h.store.mu.Lock()

	// Update stored content
	h.store.RawFiles[uriStr] = contentBytes

	// Detect file type
	fileType := epub.DetectFileType(uriStr, contentBytes)
	h.store.FileTypes[uriStr] = fileType
	opfChanged := fileType == epub.FileTypeOPF

	// Build workspace context
	ctx := &validator.WorkspaceContext{
		RootPath:              h.store.RootPath,
		Files:                 h.store.RawFiles,
		FileTypes:             h.store.FileTypes,
		AccessibilitySeverity: accessibilitySeverity(h.store.Settings),
	}

	// Update manifest info from any OPF files
	for u, c := range h.store.RawFiles {
		if h.store.FileTypes[u] == epub.FileTypeOPF {
			if m := opf.ParseManifest(c); m != nil {
				ctx.Manifest = m
				h.store.Manifest = m
				break
			}
		}
	}

	// Resolve file types for all files if needed
	for u, c := range h.store.RawFiles {
		if h.store.FileTypes[u] == epub.FileTypeUnknown {
			h.store.FileTypes[u] = epub.DetectFileType(u, c)
		}
	}

	h.store.mu.Unlock()

	// Validate the changed file
	diags := h.registry.ValidateFile(uriStr, contentBytes, fileType, ctx)

	h.store.mu.Lock()
	h.store.Diagnostics[uriStr] = diags
	h.store.mu.Unlock()

	// If OPF changed, we should re-validate other files too, but the
	// server harness only calls Diagnostics for the changed file. For now
	// we handle the single-file case. Cross-file re-validation on OPF
	// change would require a more advanced integration.
	_ = opfChanged

	// Convert to protocol diagnostics
	result := make([]protocol.Diagnostic, len(diags))
	for i, d := range diags {
		result[i] = protocol.Diagnostic{
			Range:    epubRangeToProtocol(d.Range),
			Message:  d.Message,
			Severity: protocol.DiagnosticSeverity(d.Severity),
			Code:     d.Code,
			Source:   d.Source,
		}
	}

	return result, nil
}

func (h *epubHandler) Shutdown(_ context.Context) error {
	return nil
}

// --- Optional handlers using JSON round-trip bridge ---

// roundTrip marshals protocol params into a JSON-RPC request, calls the
// existing handler, and unmarshals the result from the JSON-RPC response.
func roundTrip[P any, R any](
	id int,
	method string,
	params P,
	handler func([]byte, lsp.WorkspaceReader) []byte,
	ws lsp.WorkspaceReader,
) (R, error) {
	req := struct {
		JsonRpc string `json:"jsonrpc"`
		Id      int    `json:"id"`
		Method  string `json:"method"`
		Params  P      `json:"params"`
	}{
		JsonRpc: "2.0",
		Id:      id,
		Method:  method,
		Params:  params,
	}

	data, err := json.Marshal(req)
	if err != nil {
		var zero R
		return zero, err
	}

	respData := handler(data, ws)

	var resp struct {
		Result R `json:"result"`
	}
	if err := json.Unmarshal(respData, &resp); err != nil {
		var zero R
		return zero, err
	}

	return resp.Result, nil
}

func (h *epubHandler) Hover(
	_ context.Context,
	params *protocol.HoverParams,
) (*protocol.Hover, error) { //nolint:unparam // interface method
	type hoverParams struct {
		TextDocument struct {
			Uri string `json:"uri"`
		} `json:"textDocument"`
		Position struct {
			Line      uint32 `json:"line"`
			Character uint32 `json:"character"`
		} `json:"position"`
	}
	p := hoverParams{}
	p.TextDocument.Uri = string(params.TextDocument.URI)
	p.Position.Line = params.Position.Line
	p.Position.Character = params.Position.Character

	result, err := roundTrip[hoverParams, *protocol.Hover](
		1,
		"textDocument/hover",
		p,
		lsp.HandleHover,
		h.store,
	)
	if err != nil {
		return nil, nil //nolint:nilerr // hover errors should return nil, not fail
	}
	return result, nil
}

func (h *epubHandler) Completion(
	_ context.Context,
	params *protocol.CompletionParams,
) (*protocol.CompletionList, error) { //nolint:unparam // interface method
	type completionParams struct {
		TextDocument struct {
			Uri string `json:"uri"`
		} `json:"textDocument"`
		Position struct {
			Line      uint32 `json:"line"`
			Character uint32 `json:"character"`
		} `json:"position"`
	}
	p := completionParams{}
	p.TextDocument.Uri = string(params.TextDocument.URI)
	p.Position.Line = params.Position.Line
	p.Position.Character = params.Position.Character

	result, err := roundTrip[completionParams, *protocol.CompletionList](
		1,
		"textDocument/completion",
		p,
		lsp.HandleCompletion,
		h.store,
	)
	if err != nil {
		return nil, nil //nolint:nilerr // completion errors should return nil
	}
	return result, nil
}

func (h *epubHandler) Definition(
	_ context.Context,
	params *protocol.DefinitionParams,
) ([]protocol.Location, error) { //nolint:unparam // interface method
	type definitionParams struct {
		TextDocument struct {
			Uri string `json:"uri"`
		} `json:"textDocument"`
		Position struct {
			Line      uint32 `json:"line"`
			Character uint32 `json:"character"`
		} `json:"position"`
	}
	p := definitionParams{}
	p.TextDocument.Uri = string(params.TextDocument.URI)
	p.Position.Line = params.Position.Line
	p.Position.Character = params.Position.Character

	result, err := roundTrip[definitionParams, []protocol.Location](
		1,
		"textDocument/definition",
		p,
		lsp.HandleDefinition,
		h.store,
	)
	if err != nil {
		return nil, nil //nolint:nilerr // definition errors should return nil
	}
	return result, nil
}

func (h *epubHandler) Formatting(
	_ context.Context,
	params *protocol.DocumentFormattingParams,
) ([]protocol.TextEdit, error) { //nolint:unparam // interface method
	type formattingParams struct {
		TextDocument struct {
			Uri string `json:"uri"`
		} `json:"textDocument"`
		Options struct {
			TabSize      int  `json:"tabSize"`
			InsertSpaces bool `json:"insertSpaces"`
		} `json:"options"`
	}
	p := formattingParams{}
	p.TextDocument.Uri = string(params.TextDocument.URI)
	p.Options.TabSize = int(params.Options.TabSize)
	p.Options.InsertSpaces = params.Options.InsertSpaces

	result, err := roundTrip[formattingParams, []protocol.TextEdit](
		1,
		"textDocument/formatting",
		p,
		lsp.HandleFormatting,
		h.store,
	)
	if err != nil {
		return nil, nil //nolint:nilerr // formatting errors should return nil
	}
	return result, nil
}

func (h *epubHandler) CodeAction(
	_ context.Context,
	params *protocol.CodeActionParams,
) ([]protocol.CodeAction, error) { //nolint:unparam // interface method
	// Marshal the protocol params directly - the JSON shape is compatible
	data, err := json.Marshal(struct {
		JsonRpc string                     `json:"jsonrpc"`
		Id      int                        `json:"id"`
		Method  string                     `json:"method"`
		Params  *protocol.CodeActionParams `json:"params"`
	}{
		JsonRpc: "2.0",
		Id:      1,
		Method:  "textDocument/codeAction",
		Params:  params,
	})
	if err != nil {
		return nil, nil //nolint:nilerr // code action errors should return nil
	}

	respData := lsp.HandleCodeAction(data, h.store)

	var resp struct {
		Result []protocol.CodeAction `json:"result"`
	}
	if err := json.Unmarshal(respData, &resp); err != nil {
		return nil, nil //nolint:nilerr // unmarshal errors should return nil
	}
	return resp.Result, nil
}

func (h *epubHandler) References(
	_ context.Context,
	params *protocol.ReferenceParams,
) ([]protocol.Location, error) { //nolint:unparam // interface method
	type referenceParams struct {
		TextDocument struct {
			Uri string `json:"uri"`
		} `json:"textDocument"`
		Position struct {
			Line      uint32 `json:"line"`
			Character uint32 `json:"character"`
		} `json:"position"`
		Context struct {
			IncludeDeclaration bool `json:"includeDeclaration"`
		} `json:"context"`
	}
	p := referenceParams{}
	p.TextDocument.Uri = string(params.TextDocument.URI)
	p.Position.Line = params.Position.Line
	p.Position.Character = params.Position.Character
	p.Context.IncludeDeclaration = params.Context.IncludeDeclaration

	result, err := roundTrip[referenceParams, []protocol.Location](
		1,
		"textDocument/references",
		p,
		lsp.HandleReferences,
		h.store,
	)
	if err != nil {
		return nil, nil //nolint:nilerr // references errors should return nil
	}
	return result, nil
}

func (h *epubHandler) DocumentSymbol(
	_ context.Context,
	params *protocol.DocumentSymbolParams,
) ([]any, error) { //nolint:unparam // interface method
	type docSymbolParams struct {
		TextDocument struct {
			Uri string `json:"uri"`
		} `json:"textDocument"`
	}
	p := docSymbolParams{}
	p.TextDocument.Uri = string(params.TextDocument.URI)

	result, err := roundTrip[docSymbolParams, []any](
		1,
		"textDocument/documentSymbol",
		p,
		lsp.HandleDocumentSymbol,
		h.store,
	)
	if err != nil {
		return nil, nil //nolint:nilerr // document symbol errors should return nil
	}
	return result, nil
}

func (h *epubHandler) DocumentLink(
	_ context.Context,
	params *protocol.DocumentLinkParams,
) ([]protocol.DocumentLink, error) { //nolint:unparam // interface method
	type docLinkParams struct {
		TextDocument struct {
			Uri string `json:"uri"`
		} `json:"textDocument"`
	}
	p := docLinkParams{}
	p.TextDocument.Uri = string(params.TextDocument.URI)

	result, err := roundTrip[docLinkParams, []protocol.DocumentLink](
		1,
		"textDocument/documentLink",
		p,
		lsp.HandleDocumentLink,
		h.store,
	)
	if err != nil {
		return nil, nil //nolint:nilerr // document link errors should return nil
	}
	return result, nil
}

func (h *epubHandler) SemanticTokensFull(
	_ context.Context,
	params *protocol.SemanticTokensParams,
) (*protocol.SemanticTokens, error) { //nolint:unparam // interface method
	type semTokenParams struct {
		TextDocument struct {
			Uri string `json:"uri"`
		} `json:"textDocument"`
	}
	p := semTokenParams{}
	p.TextDocument.Uri = string(params.TextDocument.URI)

	result, err := roundTrip[semTokenParams, *protocol.SemanticTokens](
		1,
		"textDocument/semanticTokens/full",
		p,
		lsp.HandleSemanticTokens,
		h.store,
	)
	if err != nil {
		return nil, nil //nolint:nilerr // semantic token errors should return nil
	}
	return result, nil
}

// --- Conversion helpers ---

// intToU32 converts an int to uint32, clamping negatives to 0.
func intToU32(n int) uint32 {
	if n < 0 {
		return 0
	}
	return uint32(n) //nolint:gosec // line/character numbers fit in uint32
}

// epubRangeToProtocol converts an epub.Range to a protocol.Range.
func epubRangeToProtocol(r epub.Range) protocol.Range {
	return protocol.Range{
		Start: protocol.Position{
			Line:      intToU32(r.Start.Line),
			Character: intToU32(r.Start.Character),
		},
		End: protocol.Position{
			Line:      intToU32(r.End.Line),
			Character: intToU32(r.End.Character),
		},
	}
}

// --- Utilities ---

// hasTargetExtension checks if a URI has one of the target file extensions.
func hasTargetExtension(uri string) bool {
	lower := strings.ToLower(uri)
	for _, ext := range TargetFileExtensions {
		if strings.HasSuffix(lower, "."+ext) {
			return true
		}
	}
	return false
}

// accessibilitySeverity maps the settings string to an epub severity constant.
func accessibilitySeverity(settings *lsp.ServerSettings) int {
	if settings == nil {
		return epub.SeverityWarning
	}
	switch settings.Accessibility {
	case "ignore":
		return 0
	case "error":
		return epub.SeverityError
	default:
		return epub.SeverityWarning
	}
}
