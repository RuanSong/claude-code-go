package llm

import (
	"context"
	"os"

	"github.com/claude-code-go/claude/pkg/anthropic"
)

// AnthropicAdapter Anthropic 客户端适配器
// 将现有的 anthropic.Client 适配为 LLMProvider 接口
type AnthropicAdapter struct {
	client *anthropic.Client
	config ProviderConfig
}

// NewAnthropicAdapter 创建 Anthropic 适配器
func NewAnthropicAdapter(client *anthropic.Client, config ProviderConfig) *AnthropicAdapter {
	// 如果 API Key 为空，尝试从环境变量获取
	if config.APIKey == "" && config.APIKeyEnv != "" {
		config.APIKey = os.Getenv(config.APIKeyEnv)
	}
	return &AnthropicAdapter{
		client: client,
		config: config,
	}
}

// Name 返回提供者名称
func (a *AnthropicAdapter) Name() string {
	return a.config.Name
}

// Config 返回提供者配置
func (a *AnthropicAdapter) Config() ProviderConfig {
	return a.config
}

// SupportsTools 是否支持工具调用
func (a *AnthropicAdapter) SupportsTools() bool {
	return a.config.Capabilities.SupportsTools
}

// SupportsVision 是否支持视觉理解
func (a *AnthropicAdapter) SupportsVision() bool {
	return a.config.Capabilities.SupportsVision
}

// CreateMessage 创建消息 (非流式)
func (a *AnthropicAdapter) CreateMessage(ctx context.Context, req *MessageRequest) (*MessageResponse, error) {
	model := req.Model
	if model == "" {
		model = a.config.DefaultModel
	}

	// 转换 MessageRequest 为 anthropic.CreateMessageRequest
	anthropicReq := &anthropic.CreateMessageRequest{
		Model: model,
	}

	// 转换消息
	for _, msg := range req.Messages {
		content := convertContentToAnthropic(msg.Content)
		anthropicReq.Messages = append(anthropicReq.Messages, anthropic.Message{
			Role:    msg.Role,
			Content: content,
		})
	}

	// 系统提示
	if req.SystemPrompt != "" {
		anthropicReq.System = req.SystemPrompt
	}

	// 其他参数
	if req.MaxTokens > 0 {
		anthropicReq.MaxTokens = req.MaxTokens
	}

	// 转换工具
	if len(req.Tools) > 0 {
		anthropicReq.Tools = convertToolsToAnthropic(req.Tools)
	}

	// 发送请求
	resp, err := a.client.CreateMessage(ctx, anthropicReq)
	if err != nil {
		return nil, err
	}

	// 转换响应
	return convertAnthropicResponse(resp), nil
}

// CreateMessageStream 创建消息 (流式)
func (a *AnthropicAdapter) CreateMessageStream(ctx context.Context, req *MessageRequest, handler func(*StreamEvent) error) error {
	model := req.Model
	if model == "" {
		model = a.config.DefaultModel
	}

	// 转换 MessageRequest 为 anthropic.CreateMessageRequest
	anthropicReq := &anthropic.CreateMessageRequest{
		Model: model,
	}

	// 转换消息
	for _, msg := range req.Messages {
		content := convertContentToAnthropic(msg.Content)
		anthropicReq.Messages = append(anthropicReq.Messages, anthropic.Message{
			Role:    msg.Role,
			Content: content,
		})
	}

	// 系统提示
	if req.SystemPrompt != "" {
		anthropicReq.System = req.SystemPrompt
	}

	// 其他参数
	if req.MaxTokens > 0 {
		anthropicReq.MaxTokens = req.MaxTokens
	}

	// 发送流式请求
	return a.client.CreateMessageStream(ctx, anthropicReq, func(event *anthropic.StreamEvent) error {
		// 转换流式事件
		streamEvent := convertAnthropicStreamEvent(event)
		return handler(streamEvent)
	})
}

// convertContentToAnthropic 转换内容为 Anthropic 格式
func convertContentToAnthropic(content any) []anthropic.ContentBlock {
	if content == nil {
		return nil
	}

	switch v := content.(type) {
	case string:
		return []anthropic.ContentBlock{anthropic.NewTextContent(v)}
	case []ContentBlock:
		result := make([]anthropic.ContentBlock, 0, len(v))
		for _, block := range v {
			switch b := block.(type) {
			case *TextContent:
				result = append(result, anthropic.NewTextContent(b.Text))
			case *ToolUseContent:
				result = append(result, anthropic.NewToolUseContent(b.ID, b.Name, b.Input))
			}
		}
		return result
	}
	return nil
}

// convertToolsToAnthropic 转换工具为 Anthropic 格式
func convertToolsToAnthropic(tools []ToolDef) []anthropic.ToolDef {
	result := make([]anthropic.ToolDef, 0, len(tools))
	for _, tool := range tools {
		result = append(result, anthropic.ToolDef{
			Name:        tool.Name,
			Description: tool.Description,
			InputSchema: tool.Parameters,
		})
	}
	return result
}

// convertAnthropicResponse 转换 Anthropic 响应
func convertAnthropicResponse(resp *anthropic.CreateMessageResponse) *MessageResponse {
	content := make([]any, 0, len(resp.Content))
	for _, block := range resp.Content {
		switch b := block.(type) {
		case *anthropic.TextContent:
			content = append(content, map[string]any{
				"type": "text",
				"text": b.Text,
			})
		case *anthropic.ToolUseContent:
			content = append(content, map[string]any{
				"type":  "tool_use",
				"id":    b.ID,
				"name":  b.Name,
				"input": b.Input,
			})
		}
	}

	return &MessageResponse{
		ID:         resp.ID,
		Type:       resp.Type,
		Role:       resp.Role,
		Content:    content,
		Model:      resp.Model,
		StopReason: resp.StopReason,
		Usage: Usage{
			InputTokens:  resp.Usage.InputTokens,
			OutputTokens: resp.Usage.OutputTokens,
		},
	}
}

// convertAnthropicStreamEvent 转换 Anthropic 流式事件
func convertAnthropicStreamEvent(event *anthropic.StreamEvent) *StreamEvent {
	result := &StreamEvent{
		Type:  event.Type,
		Index: event.Index,
	}

	if event.Delta != nil {
		result.Delta = &StreamDelta{
			Type:        event.Delta.Type,
			Text:        event.Delta.Text,
			PartialJSON: event.Delta.PartialJSON,
		}
	}

	return result
}

// Ensure AnthropicAdapter implements LLMProvider at compile time
var _ LLMProvider = (*AnthropicAdapter)(nil)

// NewAnthropicProvider 创建基于现有 anthropic.Client 的 Provider
// 用于兼容现有的 anthropic.Client 实现
func NewAnthropicProvider(apiKey, baseURL, defaultModel string) LLMProvider {
	config := ProviderConfig{
		Name:         "anthropic",
		BaseURL:      baseURL,
		APIKey:       apiKey,
		DefaultModel: defaultModel,
		Capabilities: ProviderCapabilities{
			SupportsTools:  true,
			SupportsVision: true,
		},
	}

	client := anthropic.NewClient(anthropic.Config{
		APIKey:  apiKey,
		BaseURL: baseURL,
		Model:   defaultModel,
	})

	return NewAnthropicAdapter(client, config)
}
