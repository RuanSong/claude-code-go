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

// OpenAIClient OpenAI 兼容客户端
// 支持所有使用 OpenAI Chat Completions API 的提供商
type OpenAIClient struct {
	config     ProviderConfig
	httpClient *http.Client
}

// NewOpenAIClient 创建 OpenAI 兼容客户端
func NewOpenAIClient(config ProviderConfig) *OpenAIClient {
	return &OpenAIClient{
		config: config,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

// Name 返回提供者名称
func (c *OpenAIClient) Name() string {
	return c.config.Name
}

// Config 返回提供者配置
func (c *OpenAIClient) Config() ProviderConfig {
	return c.config
}

// SupportsTools 是否支持工具调用
func (c *OpenAIClient) SupportsTools() bool {
	return c.config.Capabilities.SupportsTools
}

// SupportsVision 是否支持视觉理解
func (c *OpenAIClient) SupportsVision() bool {
	return c.config.Capabilities.SupportsVision
}

// CreateMessage 创建消息 (非流式)
func (c *OpenAIClient) CreateMessage(ctx context.Context, req *MessageRequest) (*MessageResponse, error) {
	model := req.Model
	if model == "" {
		model = c.config.DefaultModel
	}

	openaiReq := c.convertToOpenAIRequest(model, req)

	body, err := json.Marshal(openaiReq)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/chat/completions", strings.TrimSuffix(c.config.BaseURL, "/"))

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

	var openaiResp OpenAIResponse
	if err := json.Unmarshal(respBody, &openaiResp); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	return c.convertToMessageResponse(&openaiResp)
}

// CreateMessageStream 创建消息 (流式)
func (c *OpenAIClient) CreateMessageStream(ctx context.Context, req *MessageRequest, handler func(*StreamEvent) error) error {
	model := req.Model
	if model == "" {
		model = c.config.DefaultModel
	}

	openaiReq := c.convertToOpenAIRequest(model, req)
	openaiReq.Stream = true

	body, err := json.Marshal(openaiReq)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/chat/completions", strings.TrimSuffix(c.config.BaseURL, "/"))

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

// convertToOpenAIRequest 转换为 OpenAI 请求格式
func (c *OpenAIClient) convertToOpenAIRequest(model string, req *MessageRequest) OpenAIRequest {
	openaiReq := OpenAIRequest{
		Model: model,
	}

	// 转换消息
	for _, msg := range req.Messages {
		openaiMsg := OpenAIMessage{
			Role: msg.Role,
		}

		// 处理 content
		switch content := msg.Content.(type) {
		case string:
			openaiMsg.Content = content
		case []ContentBlock:
			openaiMsg.Content = c.convertContentBlocks(content)
		}
		openaiReq.Messages = append(openaiReq.Messages, openaiMsg)
	}

	// 系统提示
	if req.SystemPrompt != "" {
		openaiReq.Messages = append([]OpenAIMessage{
			{Role: "system", Content: req.SystemPrompt},
		}, openaiReq.Messages...)
	}

	// 其他参数
	if req.MaxTokens > 0 {
		openaiReq.MaxTokens = req.MaxTokens
	}
	if req.Temperature > 0 {
		openaiReq.Temperature = req.Temperature
	}

	// 工具
	if len(req.Tools) > 0 {
		openaiReq.Tools = c.convertTools(req.Tools)
	}

	// 工具选择
	if req.ToolChoice != "" {
		openaiReq.ToolChoice = req.ToolChoice
	}

	return openaiReq
}

// convertContentBlocks 转换内容块
func (c *OpenAIClient) convertContentBlocks(blocks []ContentBlock) any {
	result := make([]any, 0, len(blocks))
	for _, block := range blocks {
		switch b := block.(type) {
		case *TextContent:
			result = append(result, map[string]any{
				"type": "text",
				"text": b.Text,
			})
		case *ImageContent:
			result = append(result, map[string]any{
				"type": "image_url",
				"image_url": map[string]any{
					"url": b.ImageURL.URL,
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

// convertTools 转换工具定义
func (c *OpenAIClient) convertTools(tools []ToolDef) []OpenAITool {
	result := make([]OpenAITool, 0, len(tools))
	for _, tool := range tools {
		result = append(result, OpenAITool{
			Type: "function",
			Function: OpenAIFunction{
				Name:        tool.Name,
				Description: tool.Description,
				Parameters:  tool.Parameters,
			},
		})
	}
	return result
}

// convertToMessageResponse 转换响应
func (c *OpenAIClient) convertToMessageResponse(openaiResp *OpenAIResponse) (*MessageResponse, error) {
	if len(openaiResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	choice := openaiResp.Choices[0]

	resp := &MessageResponse{
		ID:         openaiResp.ID,
		Type:       "message",
		Model:      openaiResp.Model,
		StopReason: choice.FinishReason,
		Usage:      Usage{},
	}

	// 转换角色
	if choice.Message.Role != "" {
		resp.Role = choice.Message.Role
	} else {
		resp.Role = "assistant"
	}

	// 转换内容
	if choice.Message.Content != "" {
		resp.Content = []any{map[string]any{"type": "text", "text": choice.Message.Content}}
	}

	// 转换使用量
	if openaiResp.Usage.InputTokens > 0 {
		resp.Usage.InputTokens = openaiResp.Usage.InputTokens
	}
	if openaiResp.Usage.OutputTokens > 0 {
		resp.Usage.OutputTokens = openaiResp.Usage.OutputTokens
	}

	return resp, nil
}

// parseSSEStream 解析 SSE 流
func (c *OpenAIClient) parseSSEStream(ctx context.Context, reader io.Reader, handler func(*StreamEvent) error) error {
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

				var chunk OpenAIChunk
				if err := json.Unmarshal([]byte(data), &chunk); err != nil {
					continue
				}

				event := c.convertToStreamEvent(&chunk)
				if err := handler(event); err != nil {
					return err
				}
			}
		}
	}
}

// convertToStreamEvent 转换流式事件
func (c *OpenAIClient) convertToStreamEvent(chunk *OpenAIChunk) *StreamEvent {
	event := &StreamEvent{
		Type: chunk.Object,
	}

	if chunk.ID != "" {
		event.Message = &Message{Content: ""}
	}

	if len(chunk.Choices) > 0 {
		choice := chunk.Choices[0]

		if choice.Delta.Content != "" {
			event.Delta = &StreamDelta{
				Type: "content_block_delta",
				Text: choice.Delta.Content,
			}
		}

		if choice.Delta.ToolCall != nil {
			event.Delta = &StreamDelta{
				Type:        "content_block_delta",
				PartialJSON: choice.Delta.ToolCall.Function.Arguments,
			}
		}

		if choice.FinishReason != "" {
			event.Delta = &StreamDelta{
				Type: "message_delta",
			}
		}
	}

	return event
}

// setHeaders 设置请求头
func (c *OpenAIClient) setHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	if c.config.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.config.APIKey)
	}
}

// OpenAIRequest OpenAI 请求格式
type OpenAIRequest struct {
	Model        string          `json:"model"`
	Messages     []OpenAIMessage `json:"messages"`
	SystemPrompt string          `json:"-"`
	MaxTokens    int             `json:"max_tokens,omitempty"`
	Temperature  float64         `json:"temperature,omitempty"`
	Tools        []OpenAITool    `json:"tools,omitempty"`
	ToolChoice   string          `json:"tool_choice,omitempty"`
	Stream       bool            `json:"stream,omitempty"`
	Extra        map[string]any  `json:"extra,omitempty"`
}

// OpenAIMessage OpenAI 消息格式
type OpenAIMessage struct {
	Role    string `json:"role"`
	Content any    `json:"content"`
}

// OpenAITool OpenAI 工具格式
type OpenAITool struct {
	Type     string         `json:"type"`
	Function OpenAIFunction `json:"function"`
}

// OpenAIFunction OpenAI 函数定义
type OpenAIFunction struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  json.RawMessage `json:"parameters,omitempty"`
}

// OpenAIResponse OpenAI 响应格式
type OpenAIResponse struct {
	ID      string         `json:"id"`
	Object  string         `json:"object"`
	Created int64          `json:"created"`
	Model   string         `json:"model"`
	Choices []OpenAIChoice `json:"choices"`
	Usage   OpenAIUsage    `json:"usage"`
}

// OpenAIChoice OpenAI 选择
type OpenAIChoice struct {
	Index        int                   `json:"index"`
	Message      OpenAIResponseMessage `json:"message"`
	FinishReason string                `json:"finish_reason"`
	Delta        OpenAIDelta           `json:"delta,omitempty"`
}

// OpenAIResponseMessage OpenAI 响应消息
type OpenAIResponseMessage struct {
	Role    string `json:"role,omitempty"`
	Content string `json:"content,omitempty"`
}

// OpenAIDelta OpenAI 增量
type OpenAIDelta struct {
	Content  string          `json:"content,omitempty"`
	Role     string          `json:"role,omitempty"`
	ToolCall *OpenAIToolCall `json:"tool_call,omitempty"`
}

// OpenAIToolCall OpenAI 工具调用
type OpenAIToolCall struct {
	Index    int                `json:"index,omitempty"`
	ID       string             `json:"id,omitempty"`
	Type     string             `json:"type,omitempty"`
	Function OpenAIFunctionCall `json:"function,omitempty"`
}

// OpenAIFunctionCall OpenAI 函数调用
type OpenAIFunctionCall struct {
	Name      string `json:"name,omitempty"`
	Arguments string `json:"arguments,omitempty"`
}

// OpenAIUsage OpenAI 使用量
type OpenAIUsage struct {
	PromptTokens     int `json:"prompt_tokens,omitempty"`
	CompletionTokens int `json:"completion_tokens,omitempty"`
	TotalTokens      int `json:"total_tokens,omitempty"`
	InputTokens      int `json:"input_tokens,omitempty"`
	OutputTokens     int `json:"output_tokens,omitempty"`
}

// OpenAIChunk OpenAI 流式块
type OpenAIChunk struct {
	ID      string         `json:"id"`
	Object  string         `json:"object"`
	Created int64          `json:"created"`
	Model   string         `json:"model"`
	Choices []OpenAIChoice `json:"choices"`
}
