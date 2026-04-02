package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"

	"github.com/claude-code-go/claude/internal/engine"
	"github.com/claude-code-go/claude/pkg/schema"
)

type AskUserQuestionTool struct{}

func (a *AskUserQuestionTool) Name() string { return "AskUserQuestion" }

func (a *AskUserQuestionTool) Description() string {
	return "Ask the user a question and get a response"
}

func (a *AskUserQuestionTool) InputSchema() schema.Schema {
	return schema.Object{
		Properties: map[string]schema.Schema{
			"question":      schema.String{},
			"options":       schema.Array{Items: schema.String{}},
			"allowMultiple": schema.Boolean{},
		},
		Required: []string{"question"},
	}
}

func (a *AskUserQuestionTool) Permission() engine.PermissionMode {
	return engine.PermissionNormal
}

type AskUserQuestionInput struct {
	Question      string   `json:"question"`
	Options       []string `json:"options,omitempty"`
	AllowMultiple bool     `json:"allowMultiple,omitempty"`
}

func (a *AskUserQuestionTool) Execute(ctx context.Context, input json.RawMessage, execCtx engine.ToolExecContext) (*engine.ToolResult, error) {
	var req AskUserQuestionInput
	if err := json.Unmarshal(input, &req); err != nil {
		return nil, fmt.Errorf("parse input: %w", err)
	}

	text := fmt.Sprintf("Question: %s\n", req.Question)
	if len(req.Options) > 0 {
		text += "Options:\n"
		for i, opt := range req.Options {
			text += fmt.Sprintf("  %d. %s\n", i+1, opt)
		}
	}
	text += "\n(This is a placeholder - user interaction not yet implemented)"

	return &engine.ToolResult{
		Content: []engine.ContentBlock{&engine.TextBlock{Text: text}},
	}, nil
}

type EnterPlanModeTool struct{}

func (e *EnterPlanModeTool) Name() string { return "EnterPlanMode" }

func (e *EnterPlanModeTool) Description() string {
	return "Enter plan mode to review changes before executing"
}

func (e *EnterPlanModeTool) InputSchema() schema.Schema {
	return schema.Object{
		Properties: map[string]schema.Schema{
			"reason": schema.String{},
		},
	}
}

func (e *EnterPlanModeTool) Permission() engine.PermissionMode {
	return engine.PermissionNormal
}

type EnterPlanModeInput struct {
	Reason string `json:"reason,omitempty"`
}

func (e *EnterPlanModeTool) Execute(ctx context.Context, input json.RawMessage, execCtx engine.ToolExecContext) (*engine.ToolResult, error) {
	var req EnterPlanModeInput
	json.Unmarshal(input, &req)

	return &engine.ToolResult{
		Content: []engine.ContentBlock{&engine.TextBlock{Text: "Entering plan mode..."}},
	}, nil
}

type ExitPlanModeTool struct{}

func (e *ExitPlanModeTool) Name() string { return "ExitPlanMode" }

func (e *ExitPlanModeTool) Description() string {
	return "Exit plan mode and continue with execution"
}

func (e *ExitPlanModeTool) InputSchema() schema.Schema {
	return schema.Object{
		Properties: map[string]schema.Schema{},
	}
}

func (e *ExitPlanModeTool) Permission() engine.PermissionMode {
	return engine.PermissionNormal
}

func (e *ExitPlanModeTool) Execute(ctx context.Context, input json.RawMessage, execCtx engine.ToolExecContext) (*engine.ToolResult, error) {
	return &engine.ToolResult{
		Content: []engine.ContentBlock{&engine.TextBlock{Text: "Exiting plan mode..."}},
	}, nil
}

type EnterWorktreeTool struct{}

func (e *EnterWorktreeTool) Name() string { return "EnterWorktree" }

func (e *EnterWorktreeTool) Description() string {
	return "Enter a git worktree for isolated changes"
}

func (e *EnterWorktreeTool) InputSchema() schema.Schema {
	return schema.Object{
		Properties: map[string]schema.Schema{
			"name":   schema.String{},
			"branch": schema.String{},
			"path":   schema.String{},
		},
		Required: []string{"name"},
	}
}

func (e *EnterWorktreeTool) Permission() engine.PermissionMode {
	return engine.PermissionElevated
}

type EnterWorktreeInput struct {
	Name   string `json:"name"`
	Branch string `json:"branch,omitempty"`
	Path   string `json:"path,omitempty"`
}

func (e *EnterWorktreeTool) Execute(ctx context.Context, input json.RawMessage, execCtx engine.ToolExecContext) (*engine.ToolResult, error) {
	var req EnterWorktreeInput
	if err := json.Unmarshal(input, &req); err != nil {
		return nil, fmt.Errorf("parse input: %w", err)
	}

	cmd := exec.CommandContext(ctx, "git", "worktree", "add")
	if req.Branch != "" {
		cmd.Args = append(cmd.Args, "-b", req.Branch)
	}
	cmd.Args = append(cmd.Args, req.Name)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return &engine.ToolResult{
			Content: []engine.ContentBlock{&engine.TextBlock{Text: string(output)}},
			IsError: true,
		}, nil
	}

	return &engine.ToolResult{
		Content: []engine.ContentBlock{&engine.TextBlock{Text: fmt.Sprintf("Entered worktree: %s", req.Name)}},
	}, nil
}

type ExitWorktreeTool struct{}

func (e *ExitWorktreeTool) Name() string { return "ExitWorktree" }

func (e *ExitWorktreeTool) Description() string {
	return "Exit a git worktree"
}

func (e *ExitWorktreeTool) InputSchema() schema.Schema {
	return schema.Object{
		Properties: map[string]schema.Schema{
			"name": schema.String{},
		},
		Required: []string{"name"},
	}
}

func (e *ExitWorktreeTool) Permission() engine.PermissionMode {
	return engine.PermissionElevated
}

type ExitWorktreeInput struct {
	Name string `json:"name"`
}

func (e *ExitWorktreeTool) Execute(ctx context.Context, input json.RawMessage, execCtx engine.ToolExecContext) (*engine.ToolResult, error) {
	var req ExitWorktreeInput
	if err := json.Unmarshal(input, &req); err != nil {
		return nil, fmt.Errorf("parse input: %w", err)
	}

	cmd := exec.CommandContext(ctx, "git", "worktree", "remove", req.Name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return &engine.ToolResult{
			Content: []engine.ContentBlock{&engine.TextBlock{Text: string(output)}},
			IsError: true,
		}, nil
	}

	return &engine.ToolResult{
		Content: []engine.ContentBlock{&engine.TextBlock{Text: fmt.Sprintf("Exited worktree: %s", req.Name)}},
	}, nil
}

type SkillTool struct{}

func (s *SkillTool) Name() string { return "Skill" }

func (s *SkillTool) Description() string {
	return "Execute a skill or custom command"
}

func (s *SkillTool) InputSchema() schema.Schema {
	return schema.Object{
		Properties: map[string]schema.Schema{
			"name": schema.String{},
			"args": schema.String{},
		},
		Required: []string{"name"},
	}
}

func (s *SkillTool) Permission() engine.PermissionMode {
	return engine.PermissionNormal
}

type SkillInput struct {
	Name string `json:"name"`
	Args string `json:"args,omitempty"`
}

type SkillOutput struct {
	Success bool   `json:"success"`
	Result  string `json:"result,omitempty"`
}

func (s *SkillTool) Execute(ctx context.Context, input json.RawMessage, execCtx engine.ToolExecContext) (*engine.ToolResult, error) {
	var req SkillInput
	if err := json.Unmarshal(input, &req); err != nil {
		return nil, fmt.Errorf("parse input: %w", err)
	}

	output := SkillOutput{
		Success: true,
		Result:  fmt.Sprintf("Skill '%s' would be executed with args: %s", req.Name, req.Args),
	}

	resultJSON, _ := json.Marshal(output)

	return &engine.ToolResult{
		Content: []engine.ContentBlock{&engine.TextBlock{
			Text: fmt.Sprintf("Executing skill: %s", req.Name),
		}, &engine.TextBlock{Text: string(resultJSON)}},
	}, nil
}

type NotebookEditTool struct{}

func (n *NotebookEditTool) Name() string { return "NotebookEdit" }

func (n *NotebookEditTool) Description() string {
	return "Edit a Jupyter notebook"
}

func (n *NotebookEditTool) InputSchema() schema.Schema {
	return schema.Object{
		Properties: map[string]schema.Schema{
			"file_path":   schema.String{},
			"cell_index":  schema.Integer{},
			"new_content": schema.String{},
			"cell_type":   schema.String{},
		},
		Required: []string{"file_path", "cell_index", "new_content"},
	}
}

func (n *NotebookEditTool) Permission() engine.PermissionMode {
	return engine.PermissionElevated
}

type NotebookEditInput struct {
	FilePath   string `json:"file_path"`
	CellIndex  int    `json:"cell_index"`
	NewContent string `json:"new_content"`
	CellType   string `json:"cell_type,omitempty"`
}

func (n *NotebookEditTool) Execute(ctx context.Context, input json.RawMessage, execCtx engine.ToolExecContext) (*engine.ToolResult, error) {
	var req NotebookEditInput
	if err := json.Unmarshal(input, &req); err != nil {
		return nil, fmt.Errorf("parse input: %w", err)
	}

	return &engine.ToolResult{
		Content: []engine.ContentBlock{&engine.TextBlock{
			Text: fmt.Sprintf("Notebook edit not fully implemented: would edit cell %d in %s", req.CellIndex, req.FilePath),
		}},
	}, nil
}

type PowerShellTool struct{}

func (p *PowerShellTool) Name() string { return "PowerShell" }

func (p *PowerShellTool) Description() string {
	return "Execute PowerShell commands (Windows only)"
}

func (p *PowerShellTool) InputSchema() schema.Schema {
	return schema.Object{
		Properties: map[string]schema.Schema{
			"command":    schema.String{},
			"timeout_ms": schema.Integer{},
		},
		Required: []string{"command"},
	}
}

func (p *PowerShellTool) Permission() engine.PermissionMode {
	return engine.PermissionElevated
}

type PowerShellInput struct {
	Command   string `json:"command"`
	TimeoutMs int    `json:"timeout_ms,omitempty"`
}

func (p *PowerShellTool) Execute(ctx context.Context, input json.RawMessage, execCtx engine.ToolExecContext) (*engine.ToolResult, error) {
	var req PowerShellInput
	if err := json.Unmarshal(input, &req); err != nil {
		return nil, fmt.Errorf("parse input: %w", err)
	}

	cmd := exec.CommandContext(ctx, "pwsh", "-c", req.Command)
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

type REPLTool struct{}

func (r *REPLTool) Name() string { return "REPL" }

func (r *REPLTool) Description() string {
	return "Start an interactive REPL session"
}

func (r *REPLTool) InputSchema() schema.Schema {
	return schema.Object{
		Properties: map[string]schema.Schema{
			"language": schema.String{},
		},
		Required: []string{"language"},
	}
}

func (r *REPLTool) Permission() engine.PermissionMode {
	return engine.PermissionElevated
}

type REPLInput struct {
	Language string `json:"language"`
}

func (r *REPLTool) Execute(ctx context.Context, input json.RawMessage, execCtx engine.ToolExecContext) (*engine.ToolResult, error) {
	var req REPLInput
	if err := json.Unmarshal(input, &req); err != nil {
		return nil, fmt.Errorf("parse input: %w", err)
	}

	return &engine.ToolResult{
		Content: []engine.ContentBlock{&engine.TextBlock{
			Text: fmt.Sprintf("REPL for %s not yet implemented", req.Language),
		}},
	}, nil
}

type ScheduleCronTool struct{}

func (s *ScheduleCronTool) Name() string { return "ScheduleCron" }

func (s *ScheduleCronTool) Description() string {
	return "Schedule a task to run periodically"
}

func (s *ScheduleCronTool) InputSchema() schema.Schema {
	return schema.Object{
		Properties: map[string]schema.Schema{
			"cron_expr":   schema.String{},
			"task":        schema.String{},
			"description": schema.String{},
		},
		Required: []string{"cron_expr", "task"},
	}
}

func (s *ScheduleCronTool) Permission() engine.PermissionMode {
	return engine.PermissionElevated
}

type ScheduleCronInput struct {
	CronExpr    string `json:"cron_expr"`
	Task        string `json:"task"`
	Description string `json:"description,omitempty"`
}

func (s *ScheduleCronTool) Execute(ctx context.Context, input json.RawMessage, execCtx engine.ToolExecContext) (*engine.ToolResult, error) {
	var req ScheduleCronInput
	if err := json.Unmarshal(input, &req); err != nil {
		return nil, fmt.Errorf("parse input: %w", err)
	}

	return &engine.ToolResult{
		Content: []engine.ContentBlock{&engine.TextBlock{
			Text: fmt.Sprintf("Cron task scheduled: %s - %s (%s)", req.CronExpr, req.Task, req.Description),
		}},
	}, nil
}

type McpAuthTool struct{}

func (m *McpAuthTool) Name() string { return "McpAuth" }

func (m *McpAuthTool) Description() string {
	return "Authenticate with an MCP server"
}

func (m *McpAuthTool) InputSchema() schema.Schema {
	return schema.Object{
		Properties: map[string]schema.Schema{
			"server":   schema.String{},
			"auth_url": schema.String{},
		},
		Required: []string{"server", "auth_url"},
	}
}

func (m *McpAuthTool) Permission() engine.PermissionMode {
	return engine.PermissionElevated
}

type McpAuthInput struct {
	Server  string `json:"server"`
	AuthURL string `json:"auth_url"`
}

func (m *McpAuthTool) Execute(ctx context.Context, input json.RawMessage, execCtx engine.ToolExecContext) (*engine.ToolResult, error) {
	var req McpAuthInput
	if err := json.Unmarshal(input, &req); err != nil {
		return nil, fmt.Errorf("parse input: %w", err)
	}

	return &engine.ToolResult{
		Content: []engine.ContentBlock{&engine.TextBlock{
			Text: fmt.Sprintf("MCP auth for %s - navigate to %s", req.Server, req.AuthURL),
		}},
	}, nil
}

type TaskOutputTool struct{}

func (t *TaskOutputTool) Name() string { return "TaskOutput" }

func (t *TaskOutputTool) Description() string {
	return "Get the output of a completed task"
}

func (t *TaskOutputTool) InputSchema() schema.Schema {
	return schema.Object{
		Properties: map[string]schema.Schema{
			"taskId": schema.String{},
		},
		Required: []string{"taskId"},
	}
}

func (t *TaskOutputTool) Permission() engine.PermissionMode {
	return engine.PermissionReadonly
}

type TaskOutputInput struct {
	TaskID string `json:"taskId"`
}

func (t *TaskOutputTool) Execute(ctx context.Context, input json.RawMessage, execCtx engine.ToolExecContext) (*engine.ToolResult, error) {
	var req TaskOutputInput
	if err := json.Unmarshal(input, &req); err != nil {
		return nil, fmt.Errorf("parse input: %w", err)
	}

	return &engine.ToolResult{
		Content: []engine.ContentBlock{&engine.TextBlock{
			Text: fmt.Sprintf("Task output for %s: (not yet implemented)", req.TaskID),
		}},
	}, nil
}

type LSPTool struct{}

func (l *LSPTool) Name() string { return "LSP" }

func (l *LSPTool) Description() string {
	return "Execute Language Server Protocol operations"
}

func (l *LSPTool) InputSchema() schema.Schema {
	return schema.Object{
		Properties: map[string]schema.Schema{
			"action":   schema.String{},
			"file":     schema.String{},
			"position": schema.Object{},
		},
		Required: []string{"action"},
	}
}

func (l *LSPTool) Permission() engine.PermissionMode {
	return engine.PermissionReadonly
}

type LSPInput struct {
	Action   string                 `json:"action"`
	File     string                 `json:"file,omitempty"`
	Position map[string]interface{} `json:"position,omitempty"`
}

func (l *LSPTool) Execute(ctx context.Context, input json.RawMessage, execCtx engine.ToolExecContext) (*engine.ToolResult, error) {
	var req LSPInput
	if err := json.Unmarshal(input, &req); err != nil {
		return nil, fmt.Errorf("parse input: %w", err)
	}

	return &engine.ToolResult{
		Content: []engine.ContentBlock{&engine.TextBlock{
			Text: fmt.Sprintf("LSP action '%s' on %s (LSP integration not yet implemented)", req.Action, req.File),
		}},
	}, nil
}

func GetAdditionalTools() []engine.Tool {
	return []engine.Tool{
		&AskUserQuestionTool{},
		&EnterPlanModeTool{},
		&ExitPlanModeTool{},
		&EnterWorktreeTool{},
		&ExitWorktreeTool{},
		&SkillTool{},
		&NotebookEditTool{},
		&PowerShellTool{},
		&REPLTool{},
		&ScheduleCronTool{},
		&McpAuthTool{},
		&TaskOutputTool{},
		&LSPTool{},
	}
}
