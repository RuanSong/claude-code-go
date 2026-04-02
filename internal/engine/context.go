package engine

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
)

// MessageType represents the type of a message
type MessageType int

const (
	MessageTypeUser MessageType = iota
	MessageTypeAssistant
	MessageTypeToolUse
	MessageTypeToolResult
	MessageTypeSystem
)

func (m MessageType) String() string {
	switch m {
	case MessageTypeUser:
		return "user"
	case MessageTypeAssistant:
		return "assistant"
	case MessageTypeToolUse:
		return "tool_use"
	case MessageTypeToolResult:
		return "tool_result"
	case MessageTypeSystem:
		return "system"
	default:
		return "unknown"
	}
}

// Message represents a conversation message
type Message struct {
	Type    MessageType    `json:"type"`
	Role    string         `json:"role"`
	Content []ContentBlock `json:"content"`
}

// Role constants
const (
	RoleUser      = "user"
	RoleAssistant = "assistant"
	RoleSystem    = "system"
	RoleTool      = "tool"
)

// NewUserMessage creates a new user message
func NewUserMessage(content string) *Message {
	return &Message{
		Type: MessageTypeUser,
		Role: RoleUser,
		Content: []ContentBlock{
			&TextBlock{Type: "text", Text: content},
		},
	}
}

// NewAssistantMessage creates a new assistant message
func NewAssistantMessage(content []ContentBlock) *Message {
	return &Message{
		Type:    MessageTypeAssistant,
		Role:    RoleAssistant,
		Content: content,
	}
}

// NewToolUseMessage creates a new tool use message
func NewToolUseMessage(toolID, toolName string, input json.RawMessage) *Message {
	return &Message{
		Type: MessageTypeToolUse,
		Role: RoleAssistant,
		Content: []ContentBlock{
			&ToolUseBlock{
				Type:  "tool_use",
				ID:    toolID,
				Name:  toolName,
				Input: input,
			},
		},
	}
}

// NewToolResultMessage creates a new tool result message
func NewToolResultMessage(toolUseID string, content any, isError bool) *Message {
	return &Message{
		Type: MessageTypeToolResult,
		Role: RoleTool,
		Content: []ContentBlock{
			&ToolResultBlock{
				Type:      "tool_result",
				ToolUseID: toolUseID,
				Content:   content,
				IsError:   isError,
			},
		},
	}
}

// HasToolCalls checks if the message contains any tool use blocks
func (m *Message) HasToolCalls() bool {
	for _, block := range m.Content {
		if _, ok := block.(*ToolUseBlock); ok {
			return true
		}
	}
	return false
}

// GetToolCalls returns all tool use blocks in the message
func (m *Message) GetToolCalls() []*ToolUseBlock {
	toolCalls := make([]*ToolUseBlock, 0)
	for _, block := range m.Content {
		if toolBlock, ok := block.(*ToolUseBlock); ok {
			toolCalls = append(toolCalls, toolBlock)
		}
	}
	return toolCalls
}

// GetTextContent returns the text content of the message
func (m *Message) GetTextContent() string {
	var builder strings.Builder
	for _, block := range m.Content {
		if tb, ok := block.(*TextBlock); ok {
			builder.WriteString(tb.Text)
		}
	}
	return builder.String()
}

// TokenUsage tracks token consumption
type TokenUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
	TotalTokens  int `json:"total_tokens"`
}

// ContextManager manages conversation context and message history
type ContextManager struct {
	mu           sync.RWMutex
	messages     []*Message
	maxTokens    int
	currentUsage TokenUsage
	systemPrompt string
}

func NewContextManager(systemPrompt string, maxTokens int) *ContextManager {
	return &ContextManager{
		messages:     make([]*Message, 0),
		maxTokens:    maxTokens,
		systemPrompt: systemPrompt,
	}
}

func (m *ContextManager) AddMessage(msg *Message) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.messages = append(m.messages, msg)
}

func (m *ContextManager) GetMessages() []*Message {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.messages
}

func (m *ContextManager) GetSystemPrompt() string {
	return m.systemPrompt
}

func (m *ContextManager) GetUsage() TokenUsage {
	return m.currentUsage
}

func (m *ContextManager) UpdateUsage(usage TokenUsage) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.currentUsage = usage
}

func (m *ContextManager) CountTokens() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.estimateTokens()
}

func (m *ContextManager) estimateTokens() int {
	total := 0
	for _, msg := range m.messages {
		data, _ := json.Marshal(msg)
		total += len(string(data)) / 4
	}
	if m.systemPrompt != "" {
		total += len(m.systemPrompt) / 4
	}
	return total
}

func (m *ContextManager) NeedsCompaction() bool {
	return m.CountTokens() > m.maxTokens
}

func (m *ContextManager) Compact(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	// TODO: implement actual compaction logic
	// For now, keep only the last half of messages
	keepCount := len(m.messages) / 2
	if keepCount < 2 {
		keepCount = 2
	}
	m.messages = m.messages[len(m.messages)-keepCount:]
	return nil
}

func (m *ContextManager) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.messages = make([]*Message, 0)
	m.currentUsage = TokenUsage{}
}

// AssistantMessage represents an LLM response
type AssistantMessage struct {
	ID         string         `json:"id"`
	Type       string         `json:"type"`
	Role       string         `json:"role"`
	Content    []ContentBlock `json:"content"`
	Model      string         `json:"model"`
	Usage      TokenUsage     `json:"usage"`
	StopReason string         `json:"stop_reason"`
}

func (a *AssistantMessage) HasToolCalls() bool {
	for _, block := range a.Content {
		if tu, ok := block.(*ToolUseBlock); ok && tu != nil {
			return true
		}
	}
	return false
}

func (a *AssistantMessage) GetToolCalls() []*ToolUseBlock {
	calls := make([]*ToolUseBlock, 0)
	for _, block := range a.Content {
		if tu, ok := block.(*ToolUseBlock); ok {
			calls = append(calls, tu)
		}
	}
	return calls
}

func (a *AssistantMessage) GetTextContent() string {
	var builder strings.Builder
	for _, block := range a.Content {
		if tb, ok := block.(*TextBlock); ok {
			builder.WriteString(tb.Text)
		}
	}
	return builder.String()
}

// ToolCall represents a single tool call from the LLM
type ToolCall struct {
	ID   string
	Name string
	Args json.RawMessage
}

// Config holds engine configuration
type Config struct {
	Model        string
	SystemPrompt string
	MaxTokens    int
	MaxTurns     int
	MaxBudgetUSD float64
	Thinking     bool
	Cwd          string
	Tools        *ToolRegistry
	Commands     *CommandRegistry
}
