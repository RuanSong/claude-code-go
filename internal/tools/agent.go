package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/claude-code-go/claude/internal/engine"
	"github.com/claude-code-go/claude/pkg/schema"
)

type Agent struct {
	ID        string
	Name      string
	Status    string
	CreatedAt time.Time
	Messages  []Message
	mu        sync.RWMutex
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

var agentCounter int
var agentMu sync.Mutex
var agents = make(map[string]*Agent)

func nextAgentID() string {
	agentMu.Lock()
	defer agentMu.Unlock()
	agentCounter++
	return fmt.Sprintf("agent-%d", agentCounter)
}

type AgentTool struct{}

func (a *AgentTool) Name() string { return "Agent" }

func (a *AgentTool) Description() string {
	return "Spawn a sub-agent to perform tasks in isolation with its own budget"
}

func (a *AgentTool) InputSchema() schema.Schema {
	return schema.Object{
		Properties: map[string]schema.Schema{
			"prompt":       schema.String{},
			"agentName":    schema.String{},
			"maxTokens":    schema.Integer{},
			"model":        schema.String{},
			"tools":        schema.Array{Items: schema.String{}},
			"budget":       schema.Float{},
			"systemPrompt": schema.String{},
		},
		Required: []string{"prompt"},
	}
}

func (a *AgentTool) Permission() engine.PermissionMode {
	return engine.PermissionElevated
}

type AgentInput struct {
	Prompt       string   `json:"prompt"`
	AgentName    string   `json:"agentName,omitempty"`
	MaxTokens    int      `json:"maxTokens,omitempty"`
	Model        string   `json:"model,omitempty"`
	Tools        []string `json:"tools,omitempty"`
	Budget       float64  `json:"budget,omitempty"`
	SystemPrompt string   `json:"systemPrompt,omitempty"`
}

type AgentOutput struct {
	AgentID   string `json:"agentId"`
	AgentName string `json:"agentName,omitempty"`
	Status    string `json:"status"`
	Result    string `json:"result,omitempty"`
}

func (a *AgentTool) Execute(ctx context.Context, input json.RawMessage, execCtx *engine.ToolExecContext) (*engine.ToolResult, error) {
	var req AgentInput
	if err := json.Unmarshal(input, &req); err != nil {
		return nil, fmt.Errorf("parse input: %w", err)
	}

	agentID := nextAgentID()
	agentName := req.AgentName
	if agentName == "" {
		agentName = fmt.Sprintf("agent-%s", agentID)
	}

	agent := &Agent{
		ID:        agentID,
		Name:      agentName,
		Status:    "running",
		CreatedAt: time.Now(),
		Messages:  []Message{},
	}

	agents[agentID] = agent

	go func() {
		time.Sleep(100 * time.Millisecond)
		agent.mu.Lock()
		agent.Status = "completed"
		agent.Messages = append(agent.Messages, Message{
			Role:    "user",
			Content: req.Prompt,
		})
		agent.Messages = append(agent.Messages, Message{
			Role:    "assistant",
			Content: fmt.Sprintf("Agent '%s' would process: %s", agentName, req.Prompt),
		})
		agent.mu.Unlock()
	}()

	output := AgentOutput{
		AgentID:   agentID,
		AgentName: agentName,
		Status:    "spawned",
	}

	resultJSON, _ := json.Marshal(output)

	return &engine.ToolResult{
		Content: []engine.ContentBlock{&engine.TextBlock{
			Text: fmt.Sprintf("Agent '%s' spawned (ID: %s)", agentName, agentID),
		}, &engine.TextBlock{Text: string(resultJSON)}},
	}, nil
}

type AgentResultTool struct{}

func (a *AgentResultTool) Name() string { return "AgentResult" }

func (a *AgentResultTool) Description() string {
	return "Get the result of a completed agent"
}

func (a *AgentResultTool) InputSchema() schema.Schema {
	return schema.Object{
		Properties: map[string]schema.Schema{
			"agentId": schema.String{},
		},
		Required: []string{"agentId"},
	}
}

func (a *AgentResultTool) Permission() engine.PermissionMode {
	return engine.PermissionReadonly
}

type AgentResultInput struct {
	AgentID string `json:"agentId"`
}

func (a *AgentResultTool) Execute(ctx context.Context, input json.RawMessage, execCtx *engine.ToolExecContext) (*engine.ToolResult, error) {
	var req AgentResultInput
	if err := json.Unmarshal(input, &req); err != nil {
		return nil, fmt.Errorf("parse input: %w", err)
	}

	agent, exists := agents[req.AgentID]
	if !exists {
		return &engine.ToolResult{
			Content: []engine.ContentBlock{&engine.TextBlock{Text: fmt.Sprintf("Agent %s not found", req.AgentID)}},
			IsError: true,
		}, nil
	}

	agent.mu.RLock()
	defer agent.mu.RUnlock()

	result := fmt.Sprintf("Agent %s (Status: %s):\n", agent.Name, agent.Status)
	for _, msg := range agent.Messages {
		result += fmt.Sprintf("[%s] %s\n", msg.Role, msg.Content)
	}

	return &engine.ToolResult{
		Content: []engine.ContentBlock{&engine.TextBlock{Text: result}},
	}, nil
}

func GetAgentTools() []engine.Tool {
	return []engine.Tool{
		&AgentTool{},
		&AgentResultTool{},
	}
}
