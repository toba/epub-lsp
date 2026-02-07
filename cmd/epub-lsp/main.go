// Command epub-lsp provides a Language Server Protocol server for EPUB validation.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/toba/epub-lsp/cmd/epub-lsp/lsp"
	"github.com/toba/epub-lsp/internal/epub"
	"github.com/toba/epub-lsp/internal/epub/validator"
	"github.com/toba/epub-lsp/internal/epub/validator/accessibility"
	"github.com/toba/epub-lsp/internal/epub/validator/css"
	"github.com/toba/epub-lsp/internal/epub/validator/nav"
	"github.com/toba/epub-lsp/internal/epub/validator/opf"
	"github.com/toba/epub-lsp/internal/epub/validator/resource"
	"github.com/toba/epub-lsp/internal/epub/validator/xhtml"
)

// version is set by goreleaser at build time.
var version = "dev"

const serverName = "epub-lsp"

// TargetFileExtensions lists the file extensions this LSP supports.
var TargetFileExtensions = []string{
	"opf", "xhtml", "html", "css", "ncx",
}

// workspaceStore holds the state for a workspace.
type workspaceStore struct {
	RootPath    string
	RawFiles    map[string][]byte
	FileTypes   map[string]epub.FileType
	Diagnostics map[string][]epub.Diagnostic
}

func main() {
	versionFlag := flag.Bool("version", false, "print the LSP version")
	flag.Parse()

	if *versionFlag {
		fmt.Printf("%s -- version %s\n", serverName, version)
		os.Exit(0)
	}

	configureLogging()
	scanner := lsp.ReceiveInput(os.Stdin)

	storage := &workspaceStore{
		RawFiles:    make(map[string][]byte),
		FileTypes:   make(map[string]epub.FileType),
		Diagnostics: make(map[string][]epub.Diagnostic),
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

	rootPathNotification := make(chan string, 2)
	textChangedNotification := make(chan bool, 2)
	textFromClient := make(map[string][]byte)
	muTextFromClient := new(sync.Mutex)
	muStdout := new(sync.Mutex)

	go processDiagnosticNotification(
		storage,
		registry,
		rootPathNotification,
		textChangedNotification,
		textFromClient,
		muTextFromClient,
		muStdout,
	)

	var request lsp.RequestMessage[any]
	var response []byte
	var isRequestResponse bool
	var isExiting bool

	slog.Info("starting lsp server",
		slog.String("server_name", serverName),
		slog.String("server_version", version),
	)

	for scanner.Scan() {
		data := scanner.Bytes()
		_ = json.Unmarshal(data, &request)

		if isExiting {
			if request.Method == lsp.MethodExit {
				break
			}
			response = lsp.ProcessIllegalRequestAfterShutdown(
				request.JsonRpc,
				request.Id,
			)
			muStdout.Lock()
			lsp.SendToLspClient(os.Stdout, response)
			muStdout.Unlock()
			continue
		}

		slog.Info("request " + request.Method)

		switch request.Method {
		case lsp.MethodInitialize:
			var rootURI string
			response, rootURI = lsp.ProcessInitializeRequest(
				data,
				serverName,
				version,
			)
			notifyTheRootPath(rootPathNotification, rootURI)
			rootPathNotification = nil
			isRequestResponse = true

		case lsp.MethodInitialized:
			isRequestResponse = false

		case lsp.MethodShutdown:
			isExiting = true
			isRequestResponse = true
			response = lsp.ProcessShutdownRequest(request.JsonRpc, request.Id)

		case lsp.MethodDidOpen:
			isRequestResponse = false
			fileURI, fileContent := lsp.ProcessDidOpenTextDocumentNotification(data)
			insertTextDocumentToDiagnostic(
				fileURI,
				fileContent,
				textChangedNotification,
				textFromClient,
				muTextFromClient,
			)

		case lsp.MethodDidChange:
			isRequestResponse = false
			fileURI, fileContent := lsp.ProcessDidChangeTextDocumentNotification(data)
			insertTextDocumentToDiagnostic(
				fileURI,
				fileContent,
				textChangedNotification,
				textFromClient,
				muTextFromClient,
			)

		case lsp.MethodDidClose:
			isRequestResponse = false

		default:
			isRequestResponse = false
		}

		if isRequestResponse {
			muStdout.Lock()
			lsp.SendToLspClient(os.Stdout, response)
			muStdout.Unlock()
		}

		response = nil
	}

	if scanner.Err() != nil {
		msg := "error while closing LSP: " + scanner.Err().Error()
		slog.Error(msg)
		panic(msg)
	}
}

// insertTextDocumentToDiagnostic queues a document for diagnostic processing.
func insertTextDocumentToDiagnostic(
	uri string,
	content []byte,
	textChangedNotification chan bool,
	textFromClient map[string][]byte,
	muTextFromClient *sync.Mutex,
) {
	if uri == "" {
		return
	}

	muTextFromClient.Lock()
	textFromClient[uri] = content

	if len(textChangedNotification) == 0 {
		textChangedNotification <- true
	}

	muTextFromClient.Unlock()
}

// notifyTheRootPath sends the root path to the diagnostic goroutine.
func notifyTheRootPath(rootPathNotification chan string, rootURI string) {
	if rootPathNotification == nil {
		return
	}
	rootPathNotification <- rootURI
	close(rootPathNotification)
}

// processDiagnosticNotification runs diagnostics and publishes results.
func processDiagnosticNotification(
	storage *workspaceStore,
	registry *validator.Registry,
	rootPathNotification chan string,
	textChangedNotification chan bool,
	textFromClient map[string][]byte,
	muTextFromClient *sync.Mutex,
	muStdout *sync.Mutex,
) {
	rootPath, ok := <-rootPathNotification
	if !ok {
		slog.Error("rootPathNotification closed before receiving")
		return
	}

	storage.RootPath = uriToFilePath(rootPath)

	notification := &lsp.NotificationMessage[lsp.PublishDiagnosticsParams]{
		JsonRpc: lsp.JSONRPCVersion,
		Method:  lsp.MethodPublishDiagnostics,
	}

	cloneTextFromClient := make(map[string][]byte)

	for {
		_, ok := <-textChangedNotification
		if !ok {
			return
		}

		if len(textFromClient) == 0 {
			continue
		}

		muTextFromClient.Lock()

		clear(cloneTextFromClient)
		for uri, content := range textFromClient {
			if !hasTargetExtension(uri) {
				continue
			}
			storage.RawFiles[uri] = content
			cloneTextFromClient[uri] = content
		}

		clear(textFromClient)
		for range len(textChangedNotification) {
			<-textChangedNotification
		}

		muTextFromClient.Unlock()

		if len(cloneTextFromClient) == 0 {
			continue
		}

		// Detect file types and parse OPF manifests
		opfChanged := false
		for uri, content := range cloneTextFromClient {
			fileType := epub.DetectFileType(uri, content)
			storage.FileTypes[uri] = fileType
			if fileType == epub.FileTypeOPF {
				opfChanged = true
			}
		}

		// Update manifest info from any OPF files
		ctx := &validator.WorkspaceContext{
			RootPath:  storage.RootPath,
			Files:     storage.RawFiles,
			FileTypes: storage.FileTypes,
		}

		for uri, content := range storage.RawFiles {
			if storage.FileTypes[uri] == epub.FileTypeOPF {
				if m := opf.ParseManifest(content); m != nil {
					ctx.Manifest = m
					break
				}
			}
		}

		// If an OPF file changed, re-validate all open files
		filesToValidate := cloneTextFromClient
		if opfChanged {
			filesToValidate = storage.RawFiles
		}

		// Resolve file types before concurrent validation
		for uri, content := range filesToValidate {
			if storage.FileTypes[uri] == epub.FileTypeUnknown {
				storage.FileTypes[uri] = epub.DetectFileType(uri, content)
			}
		}

		// Validate files concurrently
		type validationResult struct {
			uri   string
			diags []epub.Diagnostic
		}
		results := make(chan validationResult, len(filesToValidate))

		var wg sync.WaitGroup
		for uri, content := range filesToValidate {
			wg.Go(func() {
				diags := registry.ValidateFile(uri, content, storage.FileTypes[uri], ctx)
				results <- validationResult{uri: uri, diags: diags}
			})
		}

		go func() {
			wg.Wait()
			close(results)
		}()

		for r := range results {
			storage.Diagnostics[r.uri] = r.diags
			publishDiagnostics(notification, r.uri, r.diags, muStdout)
		}
	}
}

// publishDiagnostics marshals and sends diagnostics for a single file.
func publishDiagnostics(
	notification *lsp.NotificationMessage[lsp.PublishDiagnosticsParams],
	uri string,
	diags []epub.Diagnostic,
	muStdout *sync.Mutex,
) {
	lspDiags := make([]lsp.Diagnostic, len(diags))
	for i, d := range diags {
		lspDiags[i] = lsp.Diagnostic{
			Range: lsp.Range{
				Start: lsp.Position{
					Line:      intToUint(d.Range.Start.Line),
					Character: intToUint(d.Range.Start.Character),
				},
				End: lsp.Position{
					Line:      intToUint(d.Range.End.Line),
					Character: intToUint(d.Range.End.Character),
				},
			},
			Message:  d.Message,
			Severity: d.Severity,
			Code:     d.Code,
			Source:   d.Source,
		}
	}

	notification.Params = lsp.PublishDiagnosticsParams{
		Uri:         uri,
		Diagnostics: lspDiags,
	}

	response, err := json.Marshal(notification)
	if err != nil {
		slog.Error("unable to marshal diagnostic notification: " + err.Error())
		return
	}

	muStdout.Lock()
	lsp.SendToLspClient(os.Stdout, response)
	muStdout.Unlock()
}

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

// uriToFilePath converts a file URI to an OS path.
func uriToFilePath(uri string) string {
	if uri == "" {
		return ""
	}

	u, err := url.Parse(uri)
	if err != nil {
		slog.Error("unable to parse URI: " + err.Error())
		return uri
	}

	if u.Scheme != "file" {
		return uri
	}

	path := u.Path
	if runtime.GOOS == "windows" {
		if len(path) >= 3 && path[0] == '/' && path[2] == ':' {
			path = path[1:]
		}
	}

	return filepath.FromSlash(path)
}

// configureLogging sets up structured logging.
func configureLogging() {
	file := createLogFile()
	if file == nil {
		file = os.Stderr
	}

	logger := slog.New(slog.NewJSONHandler(file, nil))
	slog.SetDefault(logger)
}

// createLogFile creates or opens the log file.
func createLogFile() *os.File {
	userCachePath, err := os.UserCacheDir()
	if err != nil {
		return os.Stderr
	}

	appCachePath := filepath.Join(userCachePath, "epub-lsp")
	logFilePath := filepath.Join(appCachePath, "epub-lsp.log")

	_ = os.Mkdir(appCachePath, lsp.DirPermissions)

	fileInfo, err := os.Stat(logFilePath)
	if err == nil && fileInfo.Size() >= lsp.MaxLogFileSize {
		//nolint:gosec // safe log file path
		file, err := os.OpenFile(logFilePath, os.O_TRUNC|os.O_WRONLY, lsp.FilePermissions)
		if err != nil {
			return os.Stderr
		}
		return file
	}

	//nolint:gosec // safe log file path
	file, err := os.OpenFile(
		logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, lsp.FilePermissions,
	)
	if err != nil {
		return os.Stderr
	}

	return file
}

// intToUint safely converts int to uint, returning 0 for negative values.
func intToUint(v int) uint {
	if v < 0 {
		return 0
	}
	return uint(v) //nolint:gosec // bounds checked above
}
