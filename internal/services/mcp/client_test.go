package mcp

import (
	"context"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	config := &ClientConfig{
		Name:    "test-client",
		Version: "1.0.0",
	}

	client := NewClient(config)

	if client == nil {
		t.Fatal("NewClient() returned nil")
	}

	if client.config != config {
		t.Error("NewClient() did not set config correctly")
	}

	if client.tools == nil {
		t.Error("NewClient() did not initialize tools map")
	}
}

func TestNewClient_DefaultHTTPClient(t *testing.T) {
	config := &ClientConfig{
		Name: "test-client",
	}

	client := NewClient(config)

	if client.config.HTTPClient == nil {
		t.Error("NewClient() did not set default HTTPClient")
	}

	if client.config.Timeout == 0 {
		t.Error("NewClient() did not set default timeout")
	}
}

func TestClient_Connect(t *testing.T) {
	tests := []struct {
		name      string
		serverURL string
		wantErr   bool
	}{
		{
			name:      "valid URL",
			serverURL: "http://localhost:8080",
			wantErr:   false,
		},
		{
			name:      "valid URL with path",
			serverURL: "http://localhost:8080/mcp",
			wantErr:   false,
		},
		{
			name:      "invalid URL",
			serverURL: "://invalid",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(&ClientConfig{Name: "test"})
			err := client.Connect(context.Background(), tt.serverURL)
			if (err != nil) != tt.wantErr {
				t.Errorf("Connect() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClient_Connect_StoresConfig(t *testing.T) {
	config := &ClientConfig{
		Name: "test-client",
	}

	client := NewClient(config)
	err := client.Connect(context.Background(), "http://localhost:8080")

	if err != nil {
		t.Errorf("Connect() unexpected error: %v", err)
	}
}

func TestClient_ListTools(t *testing.T) {
	client := NewClient(&ClientConfig{Name: "test"})

	// Register some tools
	client.RegisterTool(MCPTool{
		Name:        "tool1",
		Description: "First tool",
		InputSchema: map[string]interface{}{},
	})
	client.RegisterTool(MCPTool{
		Name:        "tool2",
		Description: "Second tool",
		InputSchema: map[string]interface{}{},
	})

	result, err := client.ListTools(context.Background())
	if err != nil {
		t.Fatalf("ListTools() error = %v", err)
	}

	if len(result.Tools) != 2 {
		t.Errorf("ListTools() returned %d tools, want 2", len(result.Tools))
	}
}

func TestClient_ListTools_Empty(t *testing.T) {
	client := NewClient(&ClientConfig{Name: "test"})

	result, err := client.ListTools(context.Background())
	if err != nil {
		t.Fatalf("ListTools() error = %v", err)
	}

	if len(result.Tools) != 0 {
		t.Errorf("ListTools() returned %d tools, want 0", len(result.Tools))
	}
}

func TestClient_CallTool(t *testing.T) {
	client := NewClient(&ClientConfig{Name: "test"})

	client.RegisterTool(MCPTool{
		Name:        "test-tool",
		Description: "A test tool",
		InputSchema: map[string]interface{}{"prompt": "string"},
	})

	args := map[string]interface{}{
		"prompt": "Hello world",
	}

	resp, err := client.CallTool(context.Background(), "test-tool", args)
	if err != nil {
		t.Fatalf("CallTool() error = %v", err)
	}

	if !resp.Success {
		t.Error("CallTool() returned Success = false")
	}

	if resp.Result == nil {
		t.Error("CallTool() returned nil Result")
	}
}

func TestClient_CallTool_NotFound(t *testing.T) {
	client := NewClient(&ClientConfig{Name: "test"})

	resp, err := client.CallTool(context.Background(), "non-existent", nil)
	if err != nil {
		t.Fatalf("CallTool() error = %v", err)
	}

	if resp.Success {
		t.Error("CallTool() returned Success = true for non-existent tool")
	}

	if resp.Error == nil {
		t.Error("CallTool() returned nil Error for non-existent tool")
	}

	if resp.Error.Code != -32601 {
		t.Errorf("CallTool() Error.Code = %d, want -32601", resp.Error.Code)
	}
}

func TestClient_RegisterTool(t *testing.T) {
	client := NewClient(&ClientConfig{Name: "test"})

	tool := MCPTool{
		Name:        "new-tool",
		Description: "A new tool",
		InputSchema: map[string]interface{}{},
	}

	client.RegisterTool(tool)

	// Verify tool was registered
	result, err := client.ListTools(context.Background())
	if err != nil {
		t.Fatalf("ListTools() error = %v", err)
	}

	found := false
	for _, t := range result.Tools {
		if t.Name == "new-tool" {
			found = true
			break
		}
	}

	if !found {
		t.Error("RegisterTool() did not register tool")
	}
}

func TestClient_Close(t *testing.T) {
	client := NewClient(&ClientConfig{Name: "test"})

	client.RegisterTool(MCPTool{
		Name:        "tool1",
		Description: "First tool",
	})

	err := client.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}

	// Verify tools were cleared
	result, err := client.ListTools(context.Background())
	if err != nil {
		t.Fatalf("ListTools() error = %v", err)
	}

	if len(result.Tools) != 0 {
		t.Errorf("Close() did not clear tools, still have %d", len(result.Tools))
	}
}

func TestNewServerManager(t *testing.T) {
	manager := NewServerManager()

	if manager == nil {
		t.Fatal("NewServerManager() returned nil")
	}

	if manager.clients == nil {
		t.Error("NewServerManager() did not initialize clients map")
	}
}

func TestServerManager_AddServer(t *testing.T) {
	manager := NewServerManager()

	config := &ClientConfig{Name: "server1"}
	client := manager.AddServer("server1", config)

	if client == nil {
		t.Fatal("AddServer() returned nil")
	}

	if client.config.Name != "server1" {
		t.Error("AddServer() did not create client with correct config")
	}
}

func TestServerManager_GetServer(t *testing.T) {
	manager := NewServerManager()

	// Add a server
	config := &ClientConfig{Name: "server1"}
	manager.AddServer("server1", config)

	// Get the server
	client, ok := manager.GetServer("server1")
	if !ok {
		t.Fatal("GetServer() returned ok = false for existing server")
	}

	if client.config.Name != "server1" {
		t.Error("GetServer() returned wrong server")
	}
}

func TestServerManager_GetServer_NotFound(t *testing.T) {
	manager := NewServerManager()

	_, ok := manager.GetServer("non-existent")
	if ok {
		t.Error("GetServer() returned ok = true for non-existent server")
	}
}

func TestServerManager_RemoveServer(t *testing.T) {
	manager := NewServerManager()

	// Add a server
	config := &ClientConfig{Name: "server1"}
	manager.AddServer("server1", config)

	// Remove the server
	manager.RemoveServer("server1")

	// Verify it's gone
	_, ok := manager.GetServer("server1")
	if ok {
		t.Error("RemoveServer() did not remove server")
	}
}

func TestServerManager_ListServers(t *testing.T) {
	manager := NewServerManager()

	// Add some servers
	manager.AddServer("server1", &ClientConfig{Name: "server1"})
	manager.AddServer("server2", &ClientConfig{Name: "server2"})
	manager.AddServer("server3", &ClientConfig{Name: "server3"})

	servers := manager.ListServers()

	if len(servers) != 3 {
		t.Errorf("ListServers() returned %d servers, want 3", len(servers))
	}
}

func TestServerManager_ListServers_Empty(t *testing.T) {
	manager := NewServerManager()

	servers := manager.ListServers()

	if len(servers) != 0 {
		t.Errorf("ListServers() returned %d servers, want 0", len(servers))
	}
}

func TestNewMCPProtocol(t *testing.T) {
	protocol := NewMCPProtocol()

	if protocol == nil {
		t.Fatal("NewMCPProtocol() returned nil")
	}

	if protocol.manager == nil {
		t.Error("NewMCPProtocol() did not initialize manager")
	}
}

func TestMCPProtocol_Initialize(t *testing.T) {
	protocol := NewMCPProtocol()

	servers := []map[string]interface{}{
		{"name": "server1", "version": "1.0.0"},
		{"name": "server2"},
	}

	err := protocol.Initialize(context.Background(), servers)
	if err != nil {
		t.Errorf("Initialize() error = %v", err)
	}
}

func TestMCPProtocol_CallTool_ServerNotFound(t *testing.T) {
	protocol := NewMCPProtocol()

	resp, err := protocol.CallTool(context.Background(), "non-existent", "tool", nil)
	if err != nil {
		t.Fatalf("CallTool() error = %v", err)
	}

	if resp.Success {
		t.Error("CallTool() returned Success = true for non-existent server")
	}
}

func TestMCPProtocol_ListTools_ServerNotFound(t *testing.T) {
	protocol := NewMCPProtocol()

	_, err := protocol.ListTools(context.Background(), "non-existent")
	if err == nil {
		t.Error("ListTools() did not return error for non-existent server")
	}
}

func TestMCPProtocol_MarshalJSON(t *testing.T) {
	protocol := NewMCPProtocol()

	data, err := protocol.MarshalJSON()
	if err != nil {
		t.Errorf("MarshalJSON() error = %v", err)
		return
	}

	if len(data) == 0 {
		t.Error("MarshalJSON() returned empty data")
	}
}

func TestClientConfig_HTTPClientTimeout(t *testing.T) {
	tests := []struct {
		name    string
		timeout time.Duration
	}{
		{
			name:    "30 second default",
			timeout: 30 * time.Second,
		},
		{
			name:    "custom timeout",
			timeout: 60 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &ClientConfig{
				Name:    "test",
				Timeout: tt.timeout,
			}

			client := NewClient(config)

			if client.config.Timeout != tt.timeout {
				t.Errorf("Client config.Timeout = %v, want %v", client.config.Timeout, tt.timeout)
			}
		})
	}
}

func TestMCPTool_Structure(t *testing.T) {
	tool := MCPTool{
		Name:        "test-tool",
		Description: "A test tool",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"prompt": map[string]interface{}{
					"type": "string",
				},
			},
		},
	}

	if tool.Name == "" {
		t.Error("MCPTool.Name is empty")
	}

	if tool.Description == "" {
		t.Error("MCPTool.Description is empty")
	}

	if tool.InputSchema == nil {
		t.Error("MCPTool.InputSchema is nil")
	}
}

func TestToolCallRequest_Structure(t *testing.T) {
	req := ToolCallRequest{
		Method: "tools/call",
		Params: map[string]interface{}{
			"name": "test",
		},
	}

	if req.Method == "" {
		t.Error("ToolCallRequest.Method is empty")
	}

	if req.Params == nil {
		t.Error("ToolCallRequest.Params is nil")
	}
}

func TestToolCallResponse_Structure(t *testing.T) {
	resp := ToolCallResponse{
		Success: true,
		Result: map[string]interface{}{
			"output": "test output",
		},
	}

	if !resp.Success {
		t.Error("ToolCallResponse.Success = false, want true")
	}

	if resp.Result == nil {
		t.Error("ToolCallResponse.Result is nil")
	}
}

func TestErrorResponse_Structure(t *testing.T) {
	errResp := ErrorResponse{
		Code:    -32601,
		Message: "Method not found",
	}

	if errResp.Code == 0 {
		t.Error("ErrorResponse.Code is zero")
	}

	if errResp.Message == "" {
		t.Error("ErrorResponse.Message is empty")
	}
}

func TestListToolsResult_Structure(t *testing.T) {
	result := ListToolsResult{
		Tools: []MCPTool{
			{Name: "tool1"},
			{Name: "tool2"},
		},
	}

	if len(result.Tools) != 2 {
		t.Errorf("ListToolsResult.Tools has %d tools, want 2", len(result.Tools))
	}
}

func TestServerManager_ConcurrentAccess(t *testing.T) {
	manager := NewServerManager()
	config := &ClientConfig{Name: "test"}

	// Concurrent add
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			manager.AddServer(string(rune('0'+id)), config)
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	servers := manager.ListServers()
	if len(servers) != 10 {
		t.Errorf("ListServers() after concurrent adds returned %d, want 10", len(servers))
	}
}
