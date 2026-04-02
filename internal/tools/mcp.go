package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/claude-code-go/claude/internal/engine"
	"github.com/claude-code-go/claude/internal/services/mcp"
	"github.com/claude-code-go/claude/pkg/schema"
)

type MCPTool struct {
	protocol *mcp.MCPProtocol
}

func NewMCPTool(protocol *mcp.MCPProtocol) *MCPTool {
	return &MCPTool{protocol: protocol}
}

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
	return engine.PermissionElevated
}

type MCPInput struct {
	Server string         `json:"server"`
	Tool   string         `json:"tool"`
	Args   map[string]any `json:"args,omitempty"`
}

type MCPOutput struct {
	Result  string                   `json:"result,omitempty"`
	Error   string                   `json:"error,omitempty"`
	Server  string                   `json:"server"`
	Tool    string                   `json:"tool"`
	Success bool                     `json:"success"`
	Content []map[string]interface{} `json:"content,omitempty"`
}

func (m *MCPTool) Execute(ctx context.Context, input json.RawMessage, execCtx *engine.ToolExecContext) (*engine.ToolResult, error) {
	var req MCPInput
	if err := json.Unmarshal(input, &req); err != nil {
		return nil, fmt.Errorf("parse input: %w", err)
	}

	if m.protocol == nil {
		return &engine.ToolResult{
			Content: []engine.ContentBlock{&engine.TextBlock{
				Text: "MCP protocol not initialized",
			}},
		}, nil
	}

	resp, err := m.protocol.CallTool(ctx, req.Server, req.Tool, req.Args)
	if err != nil {
		return &engine.ToolResult{
			Content: []engine.ContentBlock{&engine.TextBlock{
				Text: fmt.Sprintf("Error calling MCP tool: %v", err),
			}},
		}, nil
	}

	output := MCPOutput{
		Server:  req.Server,
		Tool:    req.Tool,
		Success: resp.Success,
	}

	if resp.Success {
		if resp.Result != nil {
			if content, ok := resp.Result["content"]; ok {
				if contentBytes, err := json.Marshal(content); err == nil {
					json.Unmarshal(contentBytes, &output.Content)
				}
			}
			if text, ok := resp.Result["text"].(string); ok {
				output.Result = text
			}
		}
	} else if resp.Error != nil {
		output.Error = resp.Error.Message
	}

	resultJSON, _ := json.Marshal(output)

	return &engine.ToolResult{
		Content: []engine.ContentBlock{&engine.TextBlock{
			Text: fmt.Sprintf("MCP tool %s from server %s", req.Tool, req.Server),
		}, &engine.TextBlock{Text: string(resultJSON)}},
	}, nil
}

type ListMcpResourcesTool struct {
	protocol *mcp.MCPProtocol
}

func NewListMcpResourcesTool(protocol *mcp.MCPProtocol) *ListMcpResourcesTool {
	return &ListMcpResourcesTool{protocol: protocol}
}

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

func (l *ListMcpResourcesTool) Execute(ctx context.Context, input json.RawMessage, execCtx *engine.ToolExecContext) (*engine.ToolResult, error) {
	var req ListMcpResourcesInput
	json.Unmarshal(input, &req)

	if l.protocol == nil {
		return &engine.ToolResult{
			Content: []engine.ContentBlock{&engine.TextBlock{Text: "MCP protocol not initialized"}},
		}, nil
	}

	serverName := req.Server
	resources, err := l.protocol.ListResources(ctx, serverName)
	if err != nil {
		return &engine.ToolResult{
			Content: []engine.ContentBlock{&engine.TextBlock{
				Text: fmt.Sprintf("Error listing resources: %v", err),
			}},
		}, nil
	}

	text := "MCP Resources:\n"
	for _, resource := range resources.Resources {
		text += fmt.Sprintf("  - %s (%s)\n", resource.URI, resource.Name)
		if resource.Description != "" {
			text += fmt.Sprintf("    %s\n", resource.Description)
		}
	}

	if len(resources.Resources) == 0 {
		text += "  (No resources available)\n"
	}

	return &engine.ToolResult{
		Content: []engine.ContentBlock{&engine.TextBlock{Text: text}},
	}, nil
}

type ReadMcpResourceTool struct {
	protocol *mcp.MCPProtocol
}

func NewReadMcpResourceTool(protocol *mcp.MCPProtocol) *ReadMcpResourceTool {
	return &ReadMcpResourceTool{protocol: protocol}
}

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

func (r *ReadMcpResourceTool) Execute(ctx context.Context, input json.RawMessage, execCtx *engine.ToolExecContext) (*engine.ToolResult, error) {
	var req ReadMcpResourceInput
	if err := json.Unmarshal(input, &req); err != nil {
		return nil, fmt.Errorf("parse input: %w", err)
	}

	if r.protocol == nil {
		return &engine.ToolResult{
			Content: []engine.ContentBlock{&engine.TextBlock{
				Text: "MCP protocol not initialized",
			}},
		}, nil
	}

	result, err := r.protocol.ReadResource(ctx, req.Server, req.Resource)
	if err != nil {
		return &engine.ToolResult{
			Content: []engine.ContentBlock{&engine.TextBlock{
				Text: fmt.Sprintf("Error reading resource: %v", err),
			}},
		}, nil
	}

	resultJSON, _ := json.Marshal(result)

	return &engine.ToolResult{
		Content: []engine.ContentBlock{&engine.TextBlock{Text: string(resultJSON)}},
	}, nil
}

func GetMCPTools(protocol *mcp.MCPProtocol) []engine.Tool {
	return []engine.Tool{
		NewMCPTool(protocol),
		NewListMcpResourcesTool(protocol),
		NewReadMcpResourceTool(protocol),
	}
}
