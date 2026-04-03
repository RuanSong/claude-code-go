package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// AnthropicClient Anthropic 兼容客户端
// 支持使用 Anthropic /messages 端点的提供商 (如 OpenCode Go)
type AnthropicClient struct {
	config     ProviderConfig
	httpClient *http.Client
}

// NewAnthropicClient 创建 Anthropic 兼容客户端
func NewAnthropicClient(config ProviderConfig) *AnthropicClient {
	return &AnthropicClient{
		config: config,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

// Name 返回提供者名称
func (c *AnthropicClient) Name() string {
	return c.config.Name
}

// Config 返回提供者配置
func (c *AnthropicClient) Config() ProviderConfig {
	return c.config
}

// SupportsTools 是否支持工具调用
func (c *AnthropicClient) SupportsTools() bool {
	return c.config.Capabilities.SupportsTools
}

// SupportsVision 是否支持视觉理解
func (c *AnthropicClient) SupportsVision() bool {
	return c.config.Capabilities.SupportsVision
}

// CreateMessage 创建消息 (非流式)
func (c *AnthropicClient) CreateMessage(ctx context.Context, req *MessageRequest) (*MessageResponse, error) {
	model := req.Model
	if model == "" {
		model = c.config.DefaultModel
	}

	anthropicReq := c.convertToAnthropicRequest(model, req)

	body, err := json.Marshal(anthropicReq)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/messages", strings.TrimSuffix(c.config.BaseURL, "/"))

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	c.setHeaders(httpReq)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, respBody)
	}

	var anthropicResp AnthropicResponse
	if err := json.Unmarshal(respBody, &anthropicResp); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	return c.convertToMessageResponse(&anthropicResp)
}

// CreateMessageStream 创建消息 (流式)
func (c *AnthropicClient) CreateMessageStream(ctx context.Context, req *MessageRequest, handler func(*StreamEvent) error) error {
	model := req.Model
	if model == "" {
		model = c.config.DefaultModel
	}

	anthropicReq := c.convertToAnthropicRequest(model, req)
	anthropicReq.Stream = true

	body, err := json.Marshal(anthropicReq)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/messages", strings.TrimSuffix(c.config.BaseURL, "/"))

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	c.setHeaders(httpReq)
	httpReq.Header.Set("Accept", "text/event-stream")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	return c.parseSSEStream(ctx, resp.Body, handler)
}

// convertToAnthropicRequest 转换为 Anthropic 请求格式
func (c *AnthropicClient) convertToAnthropicRequest(model string, req *MessageRequest) AnthropicRequest {
	anthropicReq := AnthropicRequest{
		Model: model,
	}

	// 转换消息
	for _, msg := range req.Messages {
		content := c.convertMessageContent(msg.Content)
		anthropicReq.Messages = append(anthropicReq.Messages, AnthropicMessage{
			Role:    msg.Role,
			Content: content,
		})
	}

	// 系统提示
	if req.SystemPrompt != "" {
		anthropicReq.SystemPrompt = req.SystemPrompt
	}

	// 其他参数
	if req.MaxTokens > 0 {
		anthropicReq.MaxTokens = req.MaxTokens
	}
	if req.Temperature > 0 {
		anthropicReq.Temperature = req.Temperature
	}

	// 工具
	if len(req.Tools) > 0 {
		anthropicReq.Tools = c.convertTools(req.Tools)
	}

	// 工具选择
	if req.ToolChoice != "" {
		anthropicReq.ToolChoice = &AnthropicToolChoice{
			Type: "function",
			Name: req.ToolChoice,
		}
	}

	// 思考配置
	if req.Thinking != nil {
		anthropicReq.Thinking = req.Thinking
	}

	return anthropicReq
}

// convertMessageContent 转换消息内容
func (c *AnthropicClient) convertMessageContent(content any) any {
	switch v := content.(type) {
	case string:
		return v
	case []ContentBlock:
		result := make([]any, 0, len(v))
		for _, block := range v {
			switch b := block.(type) {
			case *TextContent:
				result = append(result, map[string]any{
					"type": "text",
					"text": b.Text,
				})
			case *ImageContent:
				result = append(result, map[string]any{
					"type": "image",
					"source": map[string]any{
						"type":       "url",
						"media_type": "image/jpeg",
						"data":       b.ImageURL.URL,
					},
				})
			case *ToolUseContent:
				result = append(result, map[string]any{
					"type": "tool_use",
					"id":   b.ID,
					"name": b.Name,
				})
			}
		}
		return result
	}
	return content
}

// convertTools 转换工具定义
func (c *AnthropicClient) convertTools(tools []ToolDef) []AnthropicTool {
	result := make([]AnthropicTool, 0, len(tools))
	for _, tool := range tools {
		result = append(result, AnthropicTool{
			Name:        tool.Name,
			Description: tool.Description,
			InputSchema: tool.Parameters,
		})
	}
	return result
}

// convertToMessageResponse 转换响应
func (c *AnthropicClient) convertToMessageResponse(resp *AnthropicResponse) (*MessageResponse, error) {
	result := &MessageResponse{
		ID:         resp.ID,
		Type:       resp.Type,
		Role:       resp.Role,
		Model:      resp.Model,
		StopReason: resp.StopReason,
		Content:    make([]any, 0),
		Usage:      Usage{},
	}

	// 转换内容块
	for _, block := range resp.Content {
		if block.Type == "text" {
			result.Content = append(result.Content, map[string]any{
				"type": "text",
				"text": block.Text,
			})
		} else if block.Type == "tool_use" {
			result.Content = append(result.Content, map[string]any{
				"type":  "tool_use",
				"id":    block.ID,
				"name":  block.Name,
				"input": block.Input,
			})
		}
	}

	// 转换使用量
	result.Usage.InputTokens = resp.Usage.InputTokens
	result.Usage.OutputTokens = resp.Usage.OutputTokens

	return result, nil
}

// parseSSEStream 解析 SSE 流
func (c *AnthropicClient) parseSSEStream(ctx context.Context, reader io.Reader, handler func(*StreamEvent) error) error {
	buf := make([]byte, 4096)
	line := ""

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		n, err := reader.Read(buf)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return fmt.Errorf("read error: %w", err)
		}

		line += string(buf[:n])

		for strings.Contains(line, "\n") {
			parts := strings.SplitN(line, "\n", 2)
			eventLine := parts[0]
			line = parts[1]

			if strings.HasPrefix(eventLine, "data: ") {
				data := strings.TrimPrefix(eventLine, "data: ")
				if data == "[DONE]" {
					return nil
				}

				var event AnthropicStreamEvent
				if err := json.Unmarshal([]byte(data), &event); err != nil {
					continue
				}

				streamEvent := c.convertToStreamEvent(&event)
				if err := handler(streamEvent); err != nil {
					return err
				}
			}
		}
	}
}

// convertToStreamEvent 转换流式事件
func (c *AnthropicClient) convertToStreamEvent(event *AnthropicStreamEvent) *StreamEvent {
	result := &StreamEvent{
		Type:  event.Type,
		Index: event.Index,
	}

	switch event.Type {
	case "content_block_start":
		if len(event.ContentBlock) > 0 {
			cb := event.ContentBlock[0]
			if cb.Type == "text" {
				result.ContentBlock = newTextContentBlock("")
			}
		}
	case "content_block_delta":
		if event.Delta.Type == "text_delta" {
			result.Delta = &StreamDelta{
				Type: "content_block_delta",
				Text: event.Delta.Text,
			}
		} else if event.Delta.Type == "input_json_delta" {
			result.Delta = &StreamDelta{
				Type:        "content_block_delta",
				PartialJSON: event.Delta.PartialJSON,
			}
		}
	case "message_delta":
		result.Delta = &StreamDelta{
			Type: "message_delta",
		}
		if event.Usage != nil {
			result.Delta.Usage = &Usage{
				OutputTokens: event.Usage.OutputTokens,
			}
		}
	}

	return result
}

func newTextContentBlock(text string) ContentBlock {
	return TextContent{Type: "text", Text: text}
}

// setHeaders 设置请求头
func (c *AnthropicClient) setHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.config.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")
}

// AnthropicRequest Anthropic 请求格式
type AnthropicRequest struct {
	Model        string               `json:"model"`
	Messages     []AnthropicMessage   `json:"messages"`
	SystemPrompt string               `json:"system,omitempty"`
	MaxTokens    int                  `json:"max_tokens"`
	Temperature  float64              `json:"temperature,omitempty"`
	Tools        []AnthropicTool      `json:"tools,omitempty"`
	ToolChoice   *AnthropicToolChoice `json:"tool_choice,omitempty"`
	Thinking     *ThinkingConfig      `json:"thinking,omitempty"`
	Stream       bool                 `json:"stream,omitempty"`
}

// AnthropicMessage Anthropic 消息格式
type AnthropicMessage struct {
	Role    string `json:"role"`
	Content any    `json:"content"`
}

// AnthropicTool Anthropic 工具格式
type AnthropicTool struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"input_schema"`
}

// AnthropicToolChoice Anthropic 工具选择
type AnthropicToolChoice struct {
	Type string `json:"type"`
	Name string `json:"name,omitempty"`
}

// AnthropicResponse Anthropic 响应格式
type AnthropicResponse struct {
	ID           string                  `json:"id"`
	Type         string                  `json:"type"`
	Role         string                  `json:"role"`
	Content      []AnthropicContentBlock `json:"content"`
	Model        string                  `json:"model"`
	StopReason   string                  `json:"stop_reason"`
	StopSequence any                     `json:"stop_sequence,omitempty"`
	Usage        AnthropicUsage          `json:"usage"`
}

// AnthropicContentBlock Anthropic 内容块
type AnthropicContentBlock struct {
	Type  string          `json:"type"`
	Text  string          `json:"text,omitempty"`
	ID    string          `json:"id,omitempty"`
	Name  string          `json:"name,omitempty"`
	Input json.RawMessage `json:"input,omitempty"`
}

// AnthropicUsage Anthropic 使用量
type AnthropicUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// AnthropicStreamEvent Anthropic 流式事件
type AnthropicStreamEvent struct {
	Type         string                  `json:"type"`
	Index        int                     `json:"index,omitempty"`
	ContentBlock []AnthropicContentBlock `json:"content_block,omitempty"`
	Delta        *AnthropicDelta         `json:"delta,omitempty"`
	Usage        *AnthropicUsage         `json:"usage,omitempty"`
}

// AnthropicDelta Anthropic 增量
type AnthropicDelta struct {
	Type        string `json:"type"`
	Text        string `json:"text,omitempty"`
	PartialJSON string `json:"partial_json,omitempty"`
}
