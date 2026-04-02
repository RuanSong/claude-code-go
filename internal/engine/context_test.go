package engine

import (
	"context"
	"encoding/json"
	"testing"
)

func TestNewContextManager(t *testing.T) {
	cm := NewContextManager("You are a helpful assistant.", 1000)

	if cm.systemPrompt != "You are a helpful assistant." {
		t.Errorf("Expected system prompt 'You are a helpful assistant.', got '%s'", cm.systemPrompt)
	}
	if cm.maxTokens != 1000 {
		t.Errorf("Expected maxTokens 1000, got %d", cm.maxTokens)
	}
}

func TestAddMessage(t *testing.T) {
	cm := NewContextManager("", 1000)

	msg := NewUserMessage("Hello")
	cm.AddMessage(msg)

	messages := cm.GetMessages()
	if len(messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(messages))
	}
}

func TestGetSystemPrompt(t *testing.T) {
	cm := NewContextManager("Test prompt", 1000)

	prompt := cm.GetSystemPrompt()
	if prompt != "Test prompt" {
		t.Errorf("Expected 'Test prompt', got '%s'", prompt)
	}
}

func TestUpdateUsage(t *testing.T) {
	cm := NewContextManager("", 1000)

	cm.UpdateUsage(TokenUsage{InputTokens: 100, OutputTokens: 50, TotalTokens: 150})

	usage := cm.GetUsage()
	if usage.InputTokens != 100 {
		t.Errorf("Expected InputTokens 100, got %d", usage.InputTokens)
	}
	if usage.OutputTokens != 50 {
		t.Errorf("Expected OutputTokens 50, got %d", usage.OutputTokens)
	}
}

func TestCountTokens(t *testing.T) {
	cm := NewContextManager("System prompt here", 1000)

	msg := NewUserMessage("Hello world")
	cm.AddMessage(msg)

	tokens := cm.CountTokens()
	if tokens <= 0 {
		t.Error("Expected positive token count")
	}
}

func TestNeedsCompaction(t *testing.T) {
	cm := NewContextManager("", 0) // 0 maxTokens

	// With 0 tokens and 0 maxTokens, NeedsCompaction should be false (0 > 0 is false)
	if cm.NeedsCompaction() {
		t.Error("Expected NeedsCompaction to be false with 0 tokens and 0 maxTokens")
	}

	// Add some content
	cm.AddMessage(NewUserMessage("Hello"))
	cm.AddMessage(NewAssistantMessage([]ContentBlock{&TextBlock{Text: "Hi"}}))

	// Now we should need compaction since tokens > 0 > maxTokens
	if !cm.NeedsCompaction() {
		t.Error("Expected NeedsCompaction to be true with content and 0 maxTokens")
	}
}

func TestCompact(t *testing.T) {
	cm := NewContextManager("", 1000)

	// Add multiple messages
	for i := 0; i < 10; i++ {
		msg := NewUserMessage("Message")
		cm.AddMessage(msg)
	}

	initialCount := len(cm.GetMessages())
	if initialCount != 10 {
		t.Errorf("Expected 10 messages, got %d", initialCount)
	}

	err := cm.Compact(context.Background())
	if err != nil {
		t.Errorf("Compact failed: %v", err)
	}

	// After compaction, should have fewer messages
	finalCount := len(cm.GetMessages())
	if finalCount >= initialCount {
		t.Errorf("Expected compaction to reduce message count, got %d -> %d", initialCount, finalCount)
	}
}

func TestClear(t *testing.T) {
	cm := NewContextManager("", 1000)

	msg := NewUserMessage("Hello")
	cm.AddMessage(msg)
	cm.UpdateUsage(TokenUsage{InputTokens: 100})

	cm.Clear()

	if len(cm.GetMessages()) != 0 {
		t.Error("Expected 0 messages after clear")
	}

	usage := cm.GetUsage()
	if usage.InputTokens != 0 {
		t.Error("Expected 0 tokens after clear")
	}
}

func TestNewUserMessage(t *testing.T) {
	msg := NewUserMessage("Hello, world!")

	if msg.Type != MessageTypeUser {
		t.Errorf("Expected MessageTypeUser, got %v", msg.Type)
	}
	if msg.Role != RoleUser {
		t.Errorf("Expected RoleUser, got %s", msg.Role)
	}
	if len(msg.Content) != 1 {
		t.Errorf("Expected 1 content block, got %d", len(msg.Content))
	}
}

func TestNewAssistantMessage(t *testing.T) {
	content := []ContentBlock{&TextBlock{Type: "text", Text: "Hello"}}
	msg := NewAssistantMessage(content)

	if msg.Type != MessageTypeAssistant {
		t.Errorf("Expected MessageTypeAssistant, got %v", msg.Type)
	}
	if msg.Role != RoleAssistant {
		t.Errorf("Expected RoleAssistant, got %s", msg.Role)
	}
}

func TestNewToolUseMessage(t *testing.T) {
	input := json.RawMessage(`{"key": "value"}`)
	msg := NewToolUseMessage("tool_123", "Bash", input)

	if msg.Type != MessageTypeToolUse {
		t.Errorf("Expected MessageTypeToolUse, got %v", msg.Type)
	}
	if msg.Role != RoleAssistant {
		t.Errorf("Expected RoleAssistant, got %s", msg.Role)
	}
}

func TestNewToolResultMessage(t *testing.T) {
	msg := NewToolResultMessage("tool_123", "Result content", false)

	if msg.Type != MessageTypeToolResult {
		t.Errorf("Expected MessageTypeToolResult, got %v", msg.Type)
	}
	if msg.Role != RoleTool {
		t.Errorf("Expected RoleTool, got %s", msg.Role)
	}
}

func TestAssistantMessageHasToolCalls(t *testing.T) {
	// Without tool calls
	content := []ContentBlock{&TextBlock{Type: "text", Text: "Hello"}}
	msg := NewAssistantMessage(content)
	if msg.HasToolCalls() {
		t.Error("Expected no tool calls")
	}

	// With tool call
	input := json.RawMessage(`{}`)
	content = []ContentBlock{&ToolUseBlock{Type: "tool_use", ID: "123", Name: "Bash", Input: input}}
	msg = NewAssistantMessage(content)
	if !msg.HasToolCalls() {
		t.Error("Expected tool calls")
	}
}

func TestAssistantMessageGetToolCalls(t *testing.T) {
	input := json.RawMessage(`{"command": "ls"}`)
	content := []ContentBlock{
		&ToolUseBlock{Type: "tool_use", ID: "1", Name: "Bash", Input: input},
		&ToolUseBlock{Type: "tool_use", ID: "2", Name: "Read", Input: input},
	}
	msg := NewAssistantMessage(content)

	calls := msg.GetToolCalls()
	if len(calls) != 2 {
		t.Errorf("Expected 2 tool calls, got %d", len(calls))
	}
}

func TestAssistantMessageGetTextContent(t *testing.T) {
	content := []ContentBlock{
		&TextBlock{Type: "text", Text: "Hello "},
		&TextBlock{Type: "text", Text: "World"},
	}
	msg := NewAssistantMessage(content)

	text := msg.GetTextContent()
	if text != "Hello World" {
		t.Errorf("Expected 'Hello World', got '%s'", text)
	}
}

func TestMessageTypeString(t *testing.T) {
	tests := []struct {
		msgType  MessageType
		expected string
	}{
		{MessageTypeUser, "user"},
		{MessageTypeAssistant, "assistant"},
		{MessageTypeToolUse, "tool_use"},
		{MessageTypeToolResult, "tool_result"},
		{MessageTypeSystem, "system"},
	}

	for _, tt := range tests {
		if tt.msgType.String() != tt.expected {
			t.Errorf("Expected %v.String() to be '%s', got '%s'", tt.msgType, tt.expected, tt.msgType.String())
		}
	}
}

func TestTokenUsage(t *testing.T) {
	usage := TokenUsage{
		InputTokens:  100,
		OutputTokens: 50,
		TotalTokens:  150,
	}

	data, err := json.Marshal(usage)
	if err != nil {
		t.Errorf("Marshal failed: %v", err)
	}

	var unmarshaled TokenUsage
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Errorf("Unmarshal failed: %v", err)
	}

	if unmarshaled.InputTokens != 100 {
		t.Errorf("Expected InputTokens 100, got %d", unmarshaled.InputTokens)
	}
}
