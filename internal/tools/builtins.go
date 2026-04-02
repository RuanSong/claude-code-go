package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/claude-code-go/claude/internal/engine"
	"github.com/claude-code-go/claude/pkg/schema"
)

// BashTool executes shell commands
type BashTool struct{}

func (b *BashTool) Name() string { return "Bash" }

func (b *BashTool) Description() string {
	return "Execute shell commands in the terminal"
}

func (b *BashTool) InputSchema() schema.Schema {
	return schema.Object{
		Properties: map[string]schema.Schema{
			"command":     schema.String{},
			"timeout_ms":  schema.Integer{},
			"working_dir": schema.String{},
		},
		Required: []string{"command"},
	}
}

func (b *BashTool) Permission() engine.PermissionMode {
	return engine.PermissionElevated
}

func (b *BashTool) Execute(ctx context.Context, input json.RawMessage, execCtx engine.ToolExecContext) (*engine.ToolResult, error) {
	var req struct {
		Command    string `json:"command"`
		TimeoutMs  int    `json:"timeout_ms"`
		WorkingDir string `json:"working_dir"`
	}

	if err := json.Unmarshal(input, &req); err != nil {
		return nil, fmt.Errorf("parse input: %w", err)
	}

	if req.TimeoutMs <= 0 {
		req.TimeoutMs = 60000 // Default 60 second timeout
	}

	cmd := exec.CommandContext(ctx, "sh", "-c", req.Command)
	if req.WorkingDir != "" {
		cmd.Dir = req.WorkingDir
	}

	// Execute with timeout
	done := make(chan error, 1)
	go func() {
		done <- nil
	}()

	select {
	case <-ctx.Done():
		return &engine.ToolResult{
			Content: []engine.ContentBlock{&engine.TextBlock{Text: "Command cancelled"}},
			IsError: true,
		}, nil
	case <-time.After(time.Duration(req.TimeoutMs) * time.Millisecond):
		return &engine.ToolResult{
			Content: []engine.ContentBlock{&engine.TextBlock{Text: "Command timed out"}},
			IsError: true,
		}, nil
	default:
		output, err := cmd.CombinedOutput()
		if err != nil {
			return &engine.ToolResult{
				Content: []engine.ContentBlock{&engine.TextBlock{Text: string(output)}},
				IsError: true,
			}, nil
		}
		return &engine.ToolResult{
			Content: []engine.ContentBlock{&engine.TextBlock{Text: string(output)}},
		}, nil
	}
}

// ReadTool reads files
type ReadTool struct{}

func (r *ReadTool) Name() string { return "Read" }

func (r *ReadTool) Description() string {
	return "Read contents of a file"
}

func (r *ReadTool) InputSchema() schema.Schema {
	return schema.Object{
		Properties: map[string]schema.Schema{
			"file_path": schema.String{},
		},
		Required: []string{"file_path"},
	}
}

func (r *ReadTool) Permission() engine.PermissionMode {
	return engine.PermissionReadonly
}

func (r *ReadTool) Execute(ctx context.Context, input json.RawMessage, execCtx engine.ToolExecContext) (*engine.ToolResult, error) {
	var req struct {
		FilePath string `json:"file_path"`
	}

	if err := json.Unmarshal(input, &req); err != nil {
		return nil, fmt.Errorf("parse input: %w", err)
	}

	// Use exec to read file since we don't have os.ReadFile wrapper yet
	cmd := exec.CommandContext(ctx, "cat", req.FilePath)
	output, err := cmd.Output()
	if err != nil {
		return &engine.ToolResult{
			Content: []engine.ContentBlock{&engine.TextBlock{Text: fmt.Sprintf("Error reading file: %v", err)}},
			IsError: true,
		}, nil
	}

	return &engine.ToolResult{
		Content: []engine.ContentBlock{&engine.TextBlock{Text: string(output)}},
	}, nil
}

// GlobTool finds files by pattern
type GlobTool struct{}

func (g *GlobTool) Name() string { return "Glob" }

func (g *GlobTool) Description() string {
	return "Find files matching a glob pattern"
}

func (g *GlobTool) InputSchema() schema.Schema {
	return schema.Object{
		Properties: map[string]schema.Schema{
			"pattern": schema.String{},
			"root":    schema.String{},
		},
		Required: []string{"pattern"},
	}
}

func (g *GlobTool) Permission() engine.PermissionMode {
	return engine.PermissionReadonly
}

func (g *GlobTool) Execute(ctx context.Context, input json.RawMessage, execCtx engine.ToolExecContext) (*engine.ToolResult, error) {
	var req struct {
		Pattern string `json:"pattern"`
		Root    string `json:"root"`
	}

	if err := json.Unmarshal(input, &req); err != nil {
		return nil, fmt.Errorf("parse input: %w", err)
	}

	// Use find command as a simple glob implementation
	root := "."
	if req.Root != "" {
		root = req.Root
	}

	cmd := exec.CommandContext(ctx, "find", root, "-name", req.Pattern)
	output, err := cmd.Output()
	if err != nil {
		return &engine.ToolResult{
			Content: []engine.ContentBlock{&engine.TextBlock{Text: fmt.Sprintf("Error finding files: %v", err)}},
			IsError: true,
		}, nil
	}

	return &engine.ToolResult{
		Content: []engine.ContentBlock{&engine.TextBlock{Text: string(output)}},
	}, nil
}

// GrepTool searches file contents
type GrepTool struct{}

func (g *GrepTool) Name() string { return "Grep" }

func (g *GrepTool) Description() string {
	return "Search for patterns in files"
}

func (g *GrepTool) InputSchema() schema.Schema {
	return schema.Object{
		Properties: map[string]schema.Schema{
			"pattern":   schema.String{},
			"path":      schema.String{},
			"recursive": schema.Boolean{},
		},
		Required: []string{"pattern", "path"},
	}
}

func (g *GrepTool) Permission() engine.PermissionMode {
	return engine.PermissionReadonly
}

func (g *GrepTool) Execute(ctx context.Context, input json.RawMessage, execCtx engine.ToolExecContext) (*engine.ToolResult, error) {
	var req struct {
		Pattern   string `json:"pattern"`
		Path      string `json:"path"`
		Recursive bool   `json:"recursive"`
	}

	if err := json.Unmarshal(input, &req); err != nil {
		return nil, fmt.Errorf("parse input: %w", err)
	}

	args := []string{}
	if req.Recursive {
		args = append(args, "-r")
	}
	args = append(args, req.Pattern, req.Path)

	cmd := exec.CommandContext(ctx, "grep", args...)
	output, err := cmd.Output()
	if err != nil {
		return &engine.ToolResult{
			Content: []engine.ContentBlock{&engine.TextBlock{Text: fmt.Sprintf("No matches found or error: %v", err)}},
			IsError: err != nil,
		}, nil
	}

	return &engine.ToolResult{
		Content: []engine.ContentBlock{&engine.TextBlock{Text: string(output)}},
	}, nil
}

// WriteTool writes content to files
type WriteTool struct{}

func (w *WriteTool) Name() string { return "Write" }

func (w *WriteTool) Description() string {
	return "Write content to a file (creates or overwrites)"
}

func (w *WriteTool) InputSchema() schema.Schema {
	return schema.Object{
		Properties: map[string]schema.Schema{
			"file_path": schema.String{},
			"content":   schema.String{},
		},
		Required: []string{"file_path", "content"},
	}
}

func (w *WriteTool) Permission() engine.PermissionMode {
	return engine.PermissionElevated
}

func (w *WriteTool) Execute(ctx context.Context, input json.RawMessage, execCtx engine.ToolExecContext) (*engine.ToolResult, error) {
	var req struct {
		FilePath string `json:"file_path"`
		Content  string `json:"content"`
	}

	if err := json.Unmarshal(input, &req); err != nil {
		return nil, fmt.Errorf("parse input: %w", err)
	}

	// Ensure parent directory exists
	dir := filepath.Dir(req.FilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return &engine.ToolResult{
			Content: []engine.ContentBlock{&engine.TextBlock{Text: fmt.Sprintf("Error creating directory: %v", err)}},
			IsError: true,
		}, nil
	}

	// Write file using os.WriteFile
	if err := os.WriteFile(req.FilePath, []byte(req.Content), 0644); err != nil {
		return &engine.ToolResult{
			Content: []engine.ContentBlock{&engine.TextBlock{Text: fmt.Sprintf("Error writing file: %v", err)}},
			IsError: true,
		}, nil
	}

	return &engine.ToolResult{
		Content: []engine.ContentBlock{&engine.TextBlock{Text: fmt.Sprintf("Written to %s", req.FilePath)}},
	}, nil
}

// Built-in tools
func GetBuiltInTools() []engine.Tool {
	return []engine.Tool{
		&BashTool{},
		&ReadTool{},
		&GlobTool{},
		&GrepTool{},
		&WriteTool{},
	}
}
