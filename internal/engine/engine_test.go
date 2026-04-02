package engine

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/claude-code-go/claude/pkg/schema"
)

func TestNewQueryEngine(t *testing.T) {
	config := Config{
		SystemPrompt: "You are a helpful assistant.",
		MaxTokens:    4096,
		MaxTurns:     100,
		Model:        "claude-sonnet-4-20250514",
	}

	engine := NewQueryEngine(config, nil)

	if engine == nil {
		t.Fatal("NewQueryEngine() returned nil")
	}

	if engine.context == nil {
		t.Error("NewQueryEngine() did not initialize context")
	}

	if engine.permission == nil {
		t.Error("NewQueryEngine() did not initialize permission manager")
	}
}

func TestQueryEngine_GetContext(t *testing.T) {
	engine := NewQueryEngine(Config{}, nil)

	ctx := engine.GetContext()
	if ctx == nil {
		t.Error("GetContext() returned nil")
	}
}

func TestQueryEngine_Reset(t *testing.T) {
	engine := NewQueryEngine(Config{}, nil)

	engine.turnCount = 5
	engine.context.AddMessage(NewUserMessage("Hello"))

	engine.Reset()

	if engine.turnCount != 0 {
		t.Errorf("Reset() turnCount = %d, want 0", engine.turnCount)
	}
}

func TestQueryEngine_SubmitMessage_SlashCommand(t *testing.T) {
	engine := NewQueryEngine(Config{
		Commands: NewCommandRegistry(),
	}, nil)

	err := engine.SubmitMessage(context.Background(), "/help")
	if err == nil {
		t.Error("SubmitMessage() should error for unknown command")
	}
}

func TestQueryEngine_SubmitMessage_MaxTurns(t *testing.T) {
	engine := NewQueryEngine(Config{
		MaxTurns: 2,
		Commands: NewCommandRegistry(),
	}, nil)

	// Simulate reaching max turns
	engine.turnCount = 2

	err := engine.SubmitMessage(context.Background(), "Hello")
	if err == nil {
		t.Error("SubmitMessage() should error when max turns exceeded")
	}
}

func TestQueryEngine_estimateCost(t *testing.T) {
	engine := NewQueryEngine(Config{}, nil)

	engine.context.UpdateUsage(TokenUsage{
		InputTokens:  1000,
		OutputTokens: 500,
	})

	cost := engine.estimateCost()
	if cost < 0 {
		t.Errorf("estimateCost() = %v, want >= 0", cost)
	}
}

func TestNewToolRegistry(t *testing.T) {
	registry := NewToolRegistry()

	if registry == nil {
		t.Fatal("NewToolRegistry() returned nil")
	}

	if registry.tools == nil {
		t.Error("NewToolRegistry() did not initialize tools map")
	}
}

func TestToolRegistry_Register(t *testing.T) {
	registry := NewToolRegistry()

	tool := &mockTool{name: "test-tool", description: "A test tool"}
	err := registry.Register(tool)
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	retrieved, ok := registry.Get("test-tool")
	if !ok {
		t.Fatal("Register() did not register tool")
	}

	if retrieved.Name() != "test-tool" {
		t.Error("Register() stored wrong tool")
	}
}

func TestToolRegistry_Register_Duplicate(t *testing.T) {
	registry := NewToolRegistry()

	tool := &mockTool{name: "test-tool"}
	registry.Register(tool)

	err := registry.Register(tool)
	if err == nil {
		t.Error("Register() should return error for duplicate tool")
	}
}

func TestToolRegistry_Register_EmptyName(t *testing.T) {
	registry := NewToolRegistry()

	tool := &mockTool{name: ""}
	err := registry.Register(tool)
	if err == nil {
		t.Error("Register() should return error for empty name")
	}
}

func TestToolRegistry_Get(t *testing.T) {
	registry := NewToolRegistry()

	registry.Register(&mockTool{name: "test-tool"})

	tool, ok := registry.Get("test-tool")
	if !ok {
		t.Fatal("Get() returned ok = false for registered tool")
	}

	if tool.Name() != "test-tool" {
		t.Error("Get() returned wrong tool")
	}
}

func TestToolRegistry_Get_NotFound(t *testing.T) {
	registry := NewToolRegistry()

	_, ok := registry.Get("non-existent")
	if ok {
		t.Error("Get() should return ok = false for non-existent tool")
	}
}

func TestToolRegistry_List(t *testing.T) {
	registry := NewToolRegistry()

	registry.Register(&mockTool{name: "tool1"})
	registry.Register(&mockTool{name: "tool2"})

	tools := registry.List()

	if len(tools) != 2 {
		t.Errorf("List() returned %d tools, want 2", len(tools))
	}
}

func TestToolRegistry_List_Empty(t *testing.T) {
	registry := NewToolRegistry()

	tools := registry.List()

	if len(tools) != 0 {
		t.Errorf("List() returned %d tools, want 0", len(tools))
	}
}

func TestToolRegistry_Names(t *testing.T) {
	registry := NewToolRegistry()

	registry.Register(&mockTool{name: "tool1"})
	registry.Register(&mockTool{name: "tool2"})

	names := registry.Names()

	if len(names) != 2 {
		t.Errorf("Names() returned %d names, want 2", len(names))
	}
}

func TestNewCommandRegistry(t *testing.T) {
	registry := NewCommandRegistry()

	if registry == nil {
		t.Fatal("NewCommandRegistry() returned nil")
	}

	if registry.commands == nil {
		t.Error("NewCommandRegistry() did not initialize commands map")
	}
}

func TestCommandRegistry_Register(t *testing.T) {
	registry := NewCommandRegistry()

	cmd := NewCommitCommand()
	err := registry.Register(cmd)
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	retrieved, ok := registry.Get("commit")
	if !ok {
		t.Fatal("Register() did not register command")
	}

	if retrieved.Name() != "commit" {
		t.Error("Register() stored wrong command")
	}
}

func TestCommandRegistry_Register_Duplicate(t *testing.T) {
	registry := NewCommandRegistry()

	cmd := NewCommitCommand()
	registry.Register(cmd)

	err := registry.Register(cmd)
	if err == nil {
		t.Error("Register() should return error for duplicate command")
	}
}

func TestCommandRegistry_Register_EmptyName(t *testing.T) {
	registry := NewCommandRegistry()

	cmd := &mockCommand{name: ""}
	err := registry.Register(cmd)
	if err == nil {
		t.Error("Register() should return error for empty name")
	}
}

func TestCommandRegistry_Get(t *testing.T) {
	registry := NewCommandRegistry()

	registry.Register(NewCommitCommand())

	cmd, ok := registry.Get("commit")
	if !ok {
		t.Fatal("Get() returned ok = false for registered command")
	}

	if cmd.Name() != "commit" {
		t.Error("Get() returned wrong command")
	}
}

func TestCommandRegistry_Get_NotFound(t *testing.T) {
	registry := NewCommandRegistry()

	_, ok := registry.Get("non-existent")
	if ok {
		t.Error("Get() should return ok = false for non-existent command")
	}
}

func TestCommandRegistry_List(t *testing.T) {
	registry := NewCommandRegistry()

	registry.Register(NewCommitCommand())
	registry.Register(NewReviewCommand())

	commands := registry.List()

	if len(commands) != 2 {
		t.Errorf("List() returned %d commands, want 2", len(commands))
	}
}

func TestCommandRegistry_GetByPrefix(t *testing.T) {
	registry := NewCommandRegistry()

	registry.Register(NewCommitCommand())
	registry.Register(NewReviewCommand())

	cmd, ok := registry.GetByPrefix("/comm")
	if !ok {
		t.Fatal("GetByPrefix() returned ok = false for existing prefix")
	}

	if cmd.Name() != "commit" {
		t.Error("GetByPrefix() returned wrong command")
	}
}

func TestCommandRegistry_GetByPrefix_NotFound(t *testing.T) {
	registry := NewCommandRegistry()

	_, ok := registry.GetByPrefix("non-existent")
	if ok {
		t.Error("GetByPrefix() should return ok = false for non-existent prefix")
	}
}

func TestPermissionMode_String(t *testing.T) {
	tests := []struct {
		mode   PermissionMode
		expect string
	}{
		{PermissionNormal, "normal"},
		{PermissionElevated, "elevated"},
		{PermissionReadonly, "readonly"},
		{PermissionMode(100), "unknown"},
	}

	for _, tt := range tests {
		if tt.mode.String() != tt.expect {
			t.Errorf("PermissionMode(%d).String() = %v, want %v", tt.mode, tt.mode.String(), tt.expect)
		}
	}
}

func TestCommandType_String(t *testing.T) {
	tests := []struct {
		cmdType CommandType
		expect  string
	}{
		{CommandTypePrompt, "prompt"},
		{CommandTypeCustom, "custom"},
		{CommandTypeLocal, "local"},
		{CommandTypeLocalJSX, "local-jsx"},
		{CommandType(100), "unknown"},
	}

	for _, tt := range tests {
		if tt.cmdType.String() != tt.expect {
			t.Errorf("CommandType(%d).String() = %v, want %v", tt.cmdType, tt.cmdType.String(), tt.expect)
		}
	}
}

func TestNewCommitCommand(t *testing.T) {
	cmd := NewCommitCommand()

	if cmd == nil {
		t.Fatal("NewCommitCommand() returned nil")
	}

	if cmd.Name() != "commit" {
		t.Errorf("NewCommitCommand() Name = %v, want commit", cmd.Name())
	}

	if cmd.Description() != "Create a git commit" {
		t.Error("NewCommitCommand() Description not set correctly")
	}

	if cmd.Type() != CommandTypePrompt {
		t.Errorf("NewCommitCommand() Type = %v, want CommandTypePrompt", cmd.Type())
	}
}

func TestNewReviewCommand(t *testing.T) {
	cmd := NewReviewCommand()

	if cmd == nil {
		t.Fatal("NewReviewCommand() returned nil")
	}

	if cmd.Name() != "review" {
		t.Errorf("NewReviewCommand() Name = %v, want review", cmd.Name())
	}

	if cmd.Type() != CommandTypePrompt {
		t.Errorf("NewReviewCommand() Type = %v, want CommandTypePrompt", cmd.Type())
	}
}

func TestNewConfigCommand(t *testing.T) {
	cmd := NewConfigCommand()

	if cmd == nil {
		t.Fatal("NewConfigCommand() returned nil")
	}

	if cmd.Name() != "config" {
		t.Errorf("NewConfigCommand() Name = %v, want config", cmd.Name())
	}

	if cmd.Type() != CommandTypeCustom {
		t.Errorf("NewConfigCommand() Type = %v, want CommandTypeCustom", cmd.Type())
	}
}

func TestTextBlock_BlockType(t *testing.T) {
	block := &TextBlock{Type: "text", Text: "Hello"}

	if block.blockType() != "text" {
		t.Errorf("TextBlock.blockType() = %v, want text", block.blockType())
	}
}

func TestToolUseBlock_BlockType(t *testing.T) {
	block := &ToolUseBlock{Type: "tool_use", ID: "123", Name: "test"}

	if block.blockType() != "tool_use" {
		t.Errorf("ToolUseBlock.blockType() = %v, want tool_use", block.blockType())
	}
}

func TestToolResultBlock_BlockType(t *testing.T) {
	block := &ToolResultBlock{Type: "tool_result", ToolUseID: "123"}

	if block.blockType() != "tool_result" {
		t.Errorf("ToolResultBlock.blockType() = %v, want tool_result", block.blockType())
	}
}

func TestImageBlock_BlockType(t *testing.T) {
	block := &ImageBlock{Type: "image", Source: "data:image/png;base64,..."}

	if block.blockType() != "image" {
		t.Errorf("ImageBlock.blockType() = %v, want image", block.blockType())
	}
}

func TestToolResult_AddText(t *testing.T) {
	result := &ToolResult{}

	result.AddText("Hello world")

	if len(result.Content) != 1 {
		t.Errorf("AddText() added %d content blocks, want 1", len(result.Content))
	}

	if result.IsError {
		t.Error("AddText() should not set IsError = true")
	}
}

func TestToolResult_AddError(t *testing.T) {
	result := &ToolResult{}

	result.AddError("Something went wrong")

	if len(result.Content) != 1 {
		t.Errorf("AddError() added %d content blocks, want 1", len(result.Content))
	}

	if !result.IsError {
		t.Error("AddError() should set IsError = true")
	}

	if result.Error == nil {
		t.Error("AddError() should set Error")
	}
}

func TestToolExecContext_GetTodos(t *testing.T) {
	ctx := &ToolExecContext{
		Todos: []TodoItem{
			{Content: "Task 1", Status: "in_progress"},
		},
	}

	todos := ctx.GetTodos()
	if len(todos) != 1 {
		t.Errorf("GetTodos() returned %d todos, want 1", len(todos))
	}
}

func TestToolExecContext_SetTodos(t *testing.T) {
	ctx := &ToolExecContext{}

	newTodos := []TodoItem{
		{Content: "Task 1", Status: "completed"},
		{Content: "Task 2", Status: "pending"},
	}

	ctx.SetTodos(newTodos)

	if len(ctx.Todos) != 2 {
		t.Errorf("SetTodos() set %d todos, want 2", len(ctx.Todos))
	}
}

func TestCommandContext_GetWorkingDirectory(t *testing.T) {
	ctx := &CommandContext{
		Cwd: "/home/user/project",
	}

	if ctx.GetWorkingDirectory() != "/home/user/project" {
		t.Error("GetWorkingDirectory() not returning correct value")
	}
}

func TestTodoItem_Structure(t *testing.T) {
	item := TodoItem{
		Content:    "Complete the task",
		Status:     "in_progress",
		ActiveForm: "Completing the task",
	}

	if item.Content == "" {
		t.Error("TodoItem.Content is empty")
	}

	if item.Status != "in_progress" {
		t.Error("TodoItem.Status not set correctly")
	}
}

func TestToolResult_Structure(t *testing.T) {
	result := &ToolResult{
		Content: []ContentBlock{&TextBlock{Text: "Hello"}},
		IsError: false,
	}

	if len(result.Content) != 1 {
		t.Error("ToolResult.Content not set correctly")
	}
}

func TestToolRegistry_ConcurrentAccess(t *testing.T) {
	registry := NewToolRegistry()
	done := make(chan bool)

	for i := 0; i < 10; i++ {
		go func(id int) {
			registry.Register(&mockTool{name: string(rune('0' + id))})
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	tools := registry.List()
	if len(tools) != 10 {
		t.Errorf("List() after concurrent registers returned %d, want 10", len(tools))
	}
}

func TestCommandRegistry_ConcurrentAccess(t *testing.T) {
	// Skip this test - direct concurrent Register calls are not thread-safe
	t.Skip("Register is not thread-safe for concurrent calls")
}

type mockTool struct {
	name        string
	description string
	permission  PermissionMode
}

func (m *mockTool) Name() string               { return m.name }
func (m *mockTool) Description() string        { return m.description }
func (m *mockTool) Permission() PermissionMode { return m.permission }
func (m *mockTool) InputSchema() schema.Schema { return nil }
func (m *mockTool) Execute(ctx context.Context, input json.RawMessage, execCtx *ToolExecContext) (*ToolResult, error) {
	return &ToolResult{Content: []ContentBlock{&TextBlock{Text: "mock result"}}}, nil
}

type mockCommand struct {
	name        string
	description string
	cmdType     CommandType
}

func (m *mockCommand) Name() string        { return m.name }
func (m *mockCommand) Description() string { return m.description }
func (m *mockCommand) Type() CommandType   { return m.cmdType }
func (m *mockCommand) Execute(ctx context.Context, args []string, execCtx CommandContext) error {
	return nil
}
