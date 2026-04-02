package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/claude-code-go/claude/internal/engine"
	"github.com/claude-code-go/claude/pkg/schema"
)

type MCPTool struct{}

func (m *MCPTool) Name() string { return "MCP" }

func (m *MCPTool) Description() string {
	return "Execute a tool from a Model Context Protocol server"
}

func (m *MCPTool) InputSchema() schema.Schema {
	return schema.Object{
		Properties: map[string]schema.Schema{
			"server": schema.String{},
			"tool":   schema.String{},
			"args":   schema.Object{},
		},
		Required: []string{"server", "tool"},
	}
}

func (m *MCPTool) Permission() engine.PermissionMode {
	return engine.PermissionNormal
}

type MCPInput struct {
	Server string         `json:"server"`
	Tool   string         `json:"tool"`
	Args   map[string]any `json:"args,omitempty"`
}

type MCPOutput struct {
	Result  string `json:"result,omitempty"`
	Error   string `json:"error,omitempty"`
	Server  string `json:"server"`
	Tool    string `json:"tool"`
	Success bool   `json:"success"`
}

func (m *MCPTool) Execute(ctx context.Context, input json.RawMessage, execCtx engine.ToolExecContext) (*engine.ToolResult, error) {
	var req MCPInput
	if err := json.Unmarshal(input, &req); err != nil {
		return nil, fmt.Errorf("parse input: %w", err)
	}

	output := MCPOutput{
		Server:  req.Server,
		Tool:    req.Tool,
		Success: true,
		Result:  fmt.Sprintf("MCP tool %s from %s would be executed here", req.Tool, req.Server),
	}

	resultJSON, _ := json.Marshal(output)

	return &engine.ToolResult{
		Content: []engine.ContentBlock{&engine.TextBlock{
			Text: fmt.Sprintf("MCP tool %s from server %s", req.Tool, req.Server),
		}, &engine.TextBlock{Text: string(resultJSON)}},
	}, nil
}

type ListMcpResourcesTool struct{}

func (l *ListMcpResourcesTool) Name() string { return "ListMcpResources" }

func (l *ListMcpResourcesTool) Description() string {
	return "List available resources from MCP servers"
}

func (l *ListMcpResourcesTool) InputSchema() schema.Schema {
	return schema.Object{
		Properties: map[string]schema.Schema{
			"server": schema.String{},
		},
	}
}

func (l *ListMcpResourcesTool) Permission() engine.PermissionMode {
	return engine.PermissionReadonly
}

type ListMcpResourcesInput struct {
	Server string `json:"server,omitempty"`
}

func (l *ListMcpResourcesTool) Execute(ctx context.Context, input json.RawMessage, execCtx engine.ToolExecContext) (*engine.ToolResult, error) {
	var req ListMcpResourcesInput
	json.Unmarshal(input, &req)

	text := "MCP Resources:\n"
	text += "  (No MCP servers connected)\n"
	text += "\nUse /mcp to manage MCP servers"

	return &engine.ToolResult{
		Content: []engine.ContentBlock{&engine.TextBlock{Text: text}},
	}, nil
}

type ReadMcpResourceTool struct{}

func (r *ReadMcpResourceTool) Name() string { return "ReadMcpResource" }

func (r *ReadMcpResourceTool) Description() string {
	return "Read a resource from an MCP server"
}

func (r *ReadMcpResourceTool) InputSchema() schema.Schema {
	return schema.Object{
		Properties: map[string]schema.Schema{
			"server":   schema.String{},
			"resource": schema.String{},
		},
		Required: []string{"server", "resource"},
	}
}

func (r *ReadMcpResourceTool) Permission() engine.PermissionMode {
	return engine.PermissionReadonly
}

type ReadMcpResourceInput struct {
	Server   string `json:"server"`
	Resource string `json:"resource"`
}

func (r *ReadMcpResourceTool) Execute(ctx context.Context, input json.RawMessage, execCtx engine.ToolExecContext) (*engine.ToolResult, error) {
	var req ReadMcpResourceInput
	if err := json.Unmarshal(input, &req); err != nil {
		return nil, fmt.Errorf("parse input: %w", err)
	}

	return &engine.ToolResult{
		Content: []engine.ContentBlock{&engine.TextBlock{
			Text: fmt.Sprintf("MCP resource %s from server %s", req.Resource, req.Server),
		}},
	}, nil
}

func GetMCPTools() []engine.Tool {
	return []engine.Tool{
		&MCPTool{},
		&ListMcpResourcesTool{},
		&ReadMcpResourceTool{},
	}
}
