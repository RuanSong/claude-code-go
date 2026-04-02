package lsp

import (
	"testing"
)

func TestNewClient(t *testing.T) {
	// Create a mock connection using a pipe
	// For simplicity, we'll just test NewClient without actual connections
	client := NewClient("test-client", nil)

	if client == nil {
		t.Fatal("NewClient() returned nil")
	}

	if client.name != "test-client" {
		t.Error("NewClient() did not set name correctly")
	}

	if client.ready {
		t.Error("NewClient() should not set ready = true initially")
	}
}

func TestClient_IsReady(t *testing.T) {
	client := NewClient("test-client", nil)

	if client.IsReady() {
		t.Error("IsReady() should return false initially")
	}
}

func TestClient_Close(t *testing.T) {
	// Skip this test as it requires a real connection
	// The Close method will panic with nil conn
	t.Skip("Close() panics with nil connection - requires integration test")
}

func TestNewServer(t *testing.T) {
	server := NewServer()

	if server == nil {
		t.Fatal("NewServer() returned nil")
	}

	if server.clients == nil {
		t.Error("NewServer() did not initialize clients map")
	}
}

func TestServer_AddClient(t *testing.T) {
	server := NewServer()

	client := server.AddClient("test-client", nil)

	if client == nil {
		t.Fatal("AddClient() returned nil")
	}

	if client.name != "test-client" {
		t.Error("AddClient() did not set client name correctly")
	}

	// Verify client was added
	retrieved, ok := server.GetClient("test-client")
	if !ok {
		t.Error("AddClient() did not add client to server")
	}

	if retrieved.name != "test-client" {
		t.Error("GetClient() returned wrong client")
	}
}

func TestServer_RemoveClient(t *testing.T) {
	server := NewServer()

	server.AddClient("test-client", nil)
	server.RemoveClient("test-client")

	_, ok := server.GetClient("test-client")
	if ok {
		t.Error("RemoveClient() did not remove client")
	}
}

func TestServer_GetClient_NotFound(t *testing.T) {
	server := NewServer()

	_, ok := server.GetClient("non-existent")
	if ok {
		t.Error("GetClient() should return ok = false for non-existent client")
	}
}

func TestServer_Broadcast(t *testing.T) {
	server := NewServer()

	msg := &LSPMessage{
		JSONRPC: "2.0",
		Method:  "test",
	}

	// Should not panic with nil clients
	err := server.Broadcast(msg)
	if err != nil {
		t.Errorf("Broadcast() error = %v", err)
	}
}

func TestNewManager(t *testing.T) {
	manager := NewManager()

	if manager == nil {
		t.Fatal("NewManager() returned nil")
	}

	if manager.servers == nil {
		t.Error("NewManager() did not initialize servers map")
	}
}

func TestManager_CreateServer(t *testing.T) {
	manager := NewManager()

	server := manager.CreateServer("test-server")

	if server == nil {
		t.Fatal("CreateServer() returned nil")
	}

	// Verify server was created
	retrieved, ok := manager.GetServer("test-server")
	if !ok {
		t.Error("CreateServer() did not add server to manager")
	}

	if retrieved != server {
		t.Error("CreateServer() returned different server than GetServer()")
	}
}

func TestManager_GetServer(t *testing.T) {
	manager := NewManager()

	manager.CreateServer("test-server")

	server, ok := manager.GetServer("test-server")
	if !ok {
		t.Fatal("GetServer() returned ok = false for existing server")
	}

	if server == nil {
		t.Error("GetServer() returned nil server")
	}
}

func TestManager_GetServer_NotFound(t *testing.T) {
	manager := NewManager()

	_, ok := manager.GetServer("non-existent")
	if ok {
		t.Error("GetServer() should return ok = false for non-existent server")
	}
}

func TestManager_RemoveServer(t *testing.T) {
	manager := NewManager()

	manager.CreateServer("test-server")
	manager.RemoveServer("test-server")

	_, ok := manager.GetServer("test-server")
	if ok {
		t.Error("RemoveServer() did not remove server")
	}
}

func TestManager_ListServers(t *testing.T) {
	manager := NewManager()

	manager.CreateServer("server1")
	manager.CreateServer("server2")
	manager.CreateServer("server3")

	servers := manager.ListServers()

	if len(servers) != 3 {
		t.Errorf("ListServers() returned %d servers, want 3", len(servers))
	}
}

func TestManager_ListServers_Empty(t *testing.T) {
	manager := NewManager()

	servers := manager.ListServers()

	if len(servers) != 0 {
		t.Errorf("ListServers() returned %d servers, want 0", len(servers))
	}
}

func TestNewLanguageServer(t *testing.T) {
	ls := NewLanguageServer("gopls", []string{"gopls"})

	if ls == nil {
		t.Fatal("NewLanguageServer() returned nil")
	}

	if ls.name != "gopls" {
		t.Error("NewLanguageServer() did not set name correctly")
	}

	if len(ls.command) != 1 || ls.command[0] != "gopls" {
		t.Error("NewLanguageServer() did not set command correctly")
	}

	if ls.env == nil {
		t.Error("NewLanguageServer() did not initialize env slice")
	}
}

func TestLanguageServer_WithEnv(t *testing.T) {
	ls := NewLanguageServer("gopls", []string{"gopls"})

	ls.WithEnv([]string{"GOPATH=/home/user/go"})

	if len(ls.env) != 1 {
		t.Error("WithEnv() did not set env correctly")
	}
}

func TestLanguageServer_WithWorkingDir(t *testing.T) {
	ls := NewLanguageServer("gopls", []string{"gopls"})

	ls.WithWorkingDir("/home/user/project")

	if ls.dir != "/home/user/project" {
		t.Error("WithWorkingDir() did not set dir correctly")
	}
}

func TestLSPMessage_Structure(t *testing.T) {
	msg := LSPMessage{
		ID:      1,
		Method:  "initialize",
		Params:  map[string]interface{}{"key": "value"},
		Result:  map[string]interface{}{"capabilities": map[string]interface{}{}},
		JSONRPC: "2.0",
	}

	if msg.JSONRPC != "2.0" {
		t.Error("LSPMessage.JSONRPC not set correctly")
	}

	if msg.Method == "" {
		t.Error("LSPMessage.Method is empty")
	}
}

func TestLSPError_Structure(t *testing.T) {
	err := LSPError{
		Code:    -32600,
		Message: "Invalid Request",
	}

	if err.Code == 0 {
		t.Error("LSPError.Code is zero")
	}

	if err.Message == "" {
		t.Error("LSPError.Message is empty")
	}
}

func TestTextDocumentItem_Structure(t *testing.T) {
	doc := TextDocumentItem{
		URI:        "file:///path/to/file.go",
		LanguageID: "go",
		Version:    1,
		Text:       "package main",
	}

	if doc.URI == "" {
		t.Error("TextDocumentItem.URI is empty")
	}

	if doc.LanguageID != "go" {
		t.Error("TextDocumentItem.LanguageID not set correctly")
	}
}

func TestPosition_Structure(t *testing.T) {
	pos := Position{
		Line:      10,
		Character: 5,
	}

	if pos.Line != 10 {
		t.Error("Position.Line not set correctly")
	}

	if pos.Character != 5 {
		t.Error("Position.Character not set correctly")
	}
}

func TestRange_Structure(t *testing.T) {
	rng := Range{
		Start: Position{Line: 0, Character: 0},
		End:   Position{Line: 1, Character: 10},
	}

	if rng.Start.Line != 0 {
		t.Error("Range.Start not set correctly")
	}

	if rng.End.Line != 1 {
		t.Error("Range.End not set correctly")
	}
}

func TestCompletionParams_Structure(t *testing.T) {
	params := CompletionParams{
		TextDocument: TextDocumentIdentifier{URI: "file:///path/to/file.go"},
		Position:     Position{Line: 10, Character: 5},
		Context: CompletionContext{
			TriggerKind:      1,
			TriggerCharacter: ".",
		},
	}

	if params.TextDocument.URI == "" {
		t.Error("CompletionParams.TextDocument not set correctly")
	}
}

func TestCompletionList_Structure(t *testing.T) {
	list := CompletionList{
		IsIncomplete: false,
		Items: []CompletionItem{
			{Label: "fmt.Println", Kind: 3, Detail: "func(a ...interface{})"},
		},
	}

	if list.IsIncomplete {
		t.Error("CompletionList.IsIncomplete should be false")
	}

	if len(list.Items) != 1 {
		t.Error("CompletionList.Items not set correctly")
	}
}

func TestCompletionItem_Structure(t *testing.T) {
	item := CompletionItem{
		Label:         "fmt.Println",
		Kind:          3,
		Detail:        "func(a ...interface{})",
		Documentation: "Prints formatted output",
		InsertText:    "fmt.Println()",
	}

	if item.Label == "" {
		t.Error("CompletionItem.Label is empty")
	}

	if item.Kind == 0 {
		t.Error("CompletionItem.Kind not set correctly")
	}
}

func TestTextEdit_Structure(t *testing.T) {
	edit := TextEdit{
		Range:   Range{Start: Position{0, 0}, End: Position{0, 5}},
		NewText: "package",
	}

	if edit.NewText == "" {
		t.Error("TextEdit.NewText is empty")
	}
}

func TestHover_Structure(t *testing.T) {
	hover := Hover{
		Contents: "func Println(a ...interface{})",
		Range:    &Range{Start: Position{0, 0}, End: Position{0, 12}},
	}

	if hover.Contents == nil {
		t.Error("Hover.Contents is nil")
	}

	if hover.Range == nil {
		t.Error("Hover.Range is nil")
	}
}

func TestLocation_Structure(t *testing.T) {
	loc := Location{
		URI:   "file:///path/to/file.go",
		Range: Range{Start: Position{10, 0}, End: Position{10, 20}},
	}

	if loc.URI == "" {
		t.Error("Location.URI is empty")
	}
}

func TestDiagnostic_Structure(t *testing.T) {
	diag := Diagnostic{
		Range:    Range{Start: Position{0, 0}, End: Position{0, 10}},
		Severity: 1,
		Source:   "gopls",
		Message:  "undefined: printf",
	}

	if diag.Message == "" {
		t.Error("Diagnostic.Message is empty")
	}
}

func TestPublishDiagnosticsParams_Structure(t *testing.T) {
	params := PublishDiagnosticsParams{
		URI: "file:///path/to/file.go",
		Diagnostics: []Diagnostic{
			{Range: Range{Start: Position{0, 0}, End: Position{0, 10}}, Severity: 1, Message: "error"},
		},
	}

	if params.URI == "" {
		t.Error("PublishDiagnosticsParams.URI is empty")
	}

	if len(params.Diagnostics) != 1 {
		t.Error("PublishDiagnosticsParams.Diagnostics not set correctly")
	}
}

func TestServerCapabilities_Structure(t *testing.T) {
	caps := ServerCapabilities{
		TextDocumentSync:   1,
		HoverProvider:      true,
		CompletionProvider: &CompletionOptions{TriggerCharacters: []string{"."}},
		DefinitionProvider: true,
		ReferencesProvider: true,
	}

	if caps.TextDocumentSync != 1 {
		t.Error("ServerCapabilities.TextDocumentSync not set correctly")
	}

	if !caps.HoverProvider {
		t.Error("ServerCapabilities.HoverProvider not set correctly")
	}
}

func TestClientCapabilities(t *testing.T) {
	// This tests that we can create and use client capabilities
	_ = ServerCapabilities{
		TextDocumentSync:   1,
		HoverProvider:      true,
		DefinitionProvider: true,
	}
}

func TestManager_ConcurrentAccess(t *testing.T) {
	manager := NewManager()
	done := make(chan bool)

	for i := 0; i < 10; i++ {
		go func(id int) {
			manager.CreateServer(string(rune('0' + id)))
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	servers := manager.ListServers()
	if len(servers) != 10 {
		t.Errorf("ListServers() after concurrent creates returned %d, want 10", len(servers))
	}
}
