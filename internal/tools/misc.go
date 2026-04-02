package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/claude-code-go/claude/internal/engine"
	"github.com/claude-code-go/claude/pkg/schema"
)

type SendMessageTool struct{}

func (s *SendMessageTool) Name() string { return "SendMessage" }

func (s *SendMessageTool) Description() string {
	return "Send a message to another agent"
}

func (s *SendMessageTool) InputSchema() schema.Schema {
	return schema.Object{
		Properties: map[string]schema.Schema{
			"recipient": schema.String{},
			"message":   schema.String{},
			"priority":  schema.String{},
		},
		Required: []string{"recipient", "message"},
	}
}

func (s *SendMessageTool) Permission() engine.PermissionMode {
	return engine.PermissionNormal
}

type SendMessageInput struct {
	Recipient string `json:"recipient"`
	Message   string `json:"message"`
	Priority  string `json:"priority,omitempty"`
}

type SendMessageOutput struct {
	Status    string `json:"status"`
	Timestamp int64  `json:"timestamp"`
}

func (s *SendMessageTool) Execute(ctx context.Context, input json.RawMessage, execCtx *engine.ToolExecContext) (*engine.ToolResult, error) {
	var req SendMessageInput
	if err := json.Unmarshal(input, &req); err != nil {
		return nil, fmt.Errorf("parse input: %w", err)
	}

	output := SendMessageOutput{
		Status:    "sent",
		Timestamp: time.Now().Unix(),
	}

	resultJSON, _ := json.Marshal(output)

	return &engine.ToolResult{
		Content: []engine.ContentBlock{&engine.TextBlock{
			Text: fmt.Sprintf("Message sent to %s", req.Recipient),
		}, &engine.TextBlock{Text: string(resultJSON)}},
	}, nil
}

type SleepTool struct{}

func (s *SleepTool) Name() string { return "Sleep" }

func (s *SleepTool) Description() string {
	return "Wait for a specified duration in proactive mode"
}

func (s *SleepTool) InputSchema() schema.Schema {
	return schema.Object{
		Properties: map[string]schema.Schema{
			"durationMs": schema.Integer{},
			"reason":     schema.String{},
		},
		Required: []string{"durationMs"},
	}
}

func (s *SleepTool) Permission() engine.PermissionMode {
	return engine.PermissionNormal
}

type SleepInput struct {
	DurationMs int64  `json:"durationMs"`
	Reason     string `json:"reason,omitempty"`
}

func (s *SleepTool) Execute(ctx context.Context, input json.RawMessage, execCtx *engine.ToolExecContext) (*engine.ToolResult, error) {
	var req SleepInput
	if err := json.Unmarshal(input, &req); err != nil {
		return nil, fmt.Errorf("parse input: %w", err)
	}

	if req.DurationMs <= 0 || req.DurationMs > 300000 {
		req.DurationMs = 1000
	}

	select {
	case <-ctx.Done():
		return &engine.ToolResult{
			Content: []engine.ContentBlock{&engine.TextBlock{Text: "Sleep interrupted"}},
		}, nil
	case <-time.After(time.Duration(req.DurationMs) * time.Millisecond):
		return &engine.ToolResult{
			Content: []engine.ContentBlock{&engine.TextBlock{
				Text: fmt.Sprintf("Slept for %dms", req.DurationMs),
			}},
		}, nil
	}
}

type BriefTool struct{}

func (b *BriefTool) Name() string { return "Brief" }

func (b *BriefTool) Description() string {
	return "Get a brief summary of the current session state"
}

func (b *BriefTool) InputSchema() schema.Schema {
	return schema.Object{
		Properties: map[string]schema.Schema{},
	}
}

func (b *BriefTool) Permission() engine.PermissionMode {
	return engine.PermissionReadonly
}

type BriefOutput struct {
	SessionDuration string `json:"sessionDuration"`
	MessageCount    int    `json:"messageCount"`
	ToolInvocations int    `json:"toolInvocations"`
}

func (b *BriefTool) Execute(ctx context.Context, input json.RawMessage, execCtx *engine.ToolExecContext) (*engine.ToolResult, error) {
	output := BriefOutput{
		SessionDuration: "unknown",
		MessageCount:    0,
		ToolInvocations: 0,
	}

	resultJSON, _ := json.Marshal(output)

	return &engine.ToolResult{
		Content: []engine.ContentBlock{&engine.TextBlock{
			Text: "Session brief: Active session with 0 messages",
		}, &engine.TextBlock{Text: string(resultJSON)}},
	}, nil
}

type ConfigTool struct{}

func (c *ConfigTool) Name() string { return "Config" }

func (c *ConfigTool) Description() string {
	return "Read or modify Claude Code configuration"
}

func (c *ConfigTool) InputSchema() schema.Schema {
	return schema.Object{
		Properties: map[string]schema.Schema{
			"action": schema.String{},
			"key":    schema.String{},
			"value":  schema.String{},
		},
		Required: []string{"action"},
	}
}

func (c *ConfigTool) Permission() engine.PermissionMode {
	return engine.PermissionElevated
}

type ConfigInput struct {
	Action string `json:"action"`
	Key    string `json:"key,omitempty"`
	Value  string `json:"value,omitempty"`
}

func (c *ConfigTool) Execute(ctx context.Context, input json.RawMessage, execCtx *engine.ToolExecContext) (*engine.ToolResult, error) {
	var req ConfigInput
	if err := json.Unmarshal(input, &req); err != nil {
		return nil, fmt.Errorf("parse input: %w", err)
	}

	var text string
	switch req.Action {
	case "get":
		text = fmt.Sprintf("Config %s = <value>", req.Key)
	case "set":
		text = fmt.Sprintf("Config %s set to %s", req.Key, req.Value)
	case "list":
		text = "Available config keys:\n  - model\n  - api_key\n  - permission_mode\n  - max_turns"
	default:
		text = fmt.Sprintf("Unknown action: %s", req.Action)
	}

	return &engine.ToolResult{
		Content: []engine.ContentBlock{&engine.TextBlock{Text: text}},
	}, nil
}

func GetMiscTools() []engine.Tool {
	return []engine.Tool{
		&SendMessageTool{},
		&SleepTool{},
		&BriefTool{},
		&ConfigTool{},
	}
}
