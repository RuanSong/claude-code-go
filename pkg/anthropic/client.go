package anthropic

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

const (
	DefaultBaseURL = "https://api.anthropic.com"
	APIVersion     = "2023-06-01"
)

type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
	model      string
}

type Config struct {
	APIKey  string
	BaseURL string
	Model   string
	Timeout time.Duration
}

func NewClient(config Config) *Client {
	if config.BaseURL == "" {
		config.BaseURL = DefaultBaseURL
	}
	if config.Timeout == 0 {
		config.Timeout = 120 * time.Second
	}

	return &Client{
		apiKey:  config.APIKey,
		baseURL: config.BaseURL,
		model:   config.Model,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

type CreateMessageRequest struct {
	Model      string          `json:"model"`
	Messages   []Message       `json:"messages"`
	MaxTokens  int             `json:"max_tokens"`
	System     string          `json:"system,omitempty"`
	Tools      []ToolDef       `json:"tools,omitempty"`
	ToolChoice *ToolChoice     `json:"tool_choice,omitempty"`
	Thinking   *ThinkingConfig `json:"thinking,omitempty"`
}

type ToolChoice struct {
	Type string `json:"type"`
	Name string `json:"name,omitempty"`
}

type ThinkingConfig struct {
	Type      string `json:"type"`
	MaxTokens int    `json:"max_tokens"`
}

type Message struct {
	Role    string         `json:"role"`
	Content []ContentBlock `json:"content"`
}

type ContentBlock interface{}

type TextContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type ToolUseContent struct {
	Type  string          `json:"type"`
	ID    string          `json:"id"`
	Name  string          `json:"name"`
	Input json.RawMessage `json:"input"`
}

type ToolResultContent struct {
	Type      string `json:"type"`
	ToolUseID string `json:"tool_use_id"`
	Content   string `json:"content"`
	IsError   bool   `json:"is_error,omitempty"`
}

type ToolDef struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"input_schema"`
}

type CreateMessageResponse struct {
	ID           string `json:"id"`
	Type         string `json:"type"`
	Role         string `json:"role"`
	Content      []any  `json:"content"`
	Model        string `json:"model"`
	StopReason   string `json:"stop_reason"`
	StopSequence any    `json:"stop_sequence,omitempty"`
	Usage        Usage  `json:"usage"`
}

type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

func (c *Client) CreateMessage(ctx context.Context, req *CreateMessageRequest) (*CreateMessageResponse, error) {
	if req.Model == "" {
		req.Model = c.model
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/v1/messages", c.baseURL)

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", c.apiKey)
	httpReq.Header.Set("anthropic-version", APIVersion)

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer httpResp.Body.Close()

	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (%d): %s", httpResp.StatusCode, respBody)
	}

	var resp CreateMessageResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	return &resp, nil
}

func (c *Client) CreateMessageStream(ctx context.Context, req *CreateMessageRequest, handler func(resp *StreamEvent) error) error {
	if req.Model == "" {
		req.Model = c.model
	}

	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/v1/messages", c.baseURL)

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", c.apiKey)
	httpReq.Header.Set("anthropic-version", APIVersion)
	httpReq.Header.Set("Accept", "text/event-stream")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	return parseSSEResponse(ctx, resp.Body, handler)
}

type StreamEvent struct {
	Type         string   `json:"type"`
	Index        int      `json:"index,omitempty"`
	ContentBlock []any    `json:"content_block,omitempty"`
	Delta        *Delta   `json:"delta,omitempty"`
	Message      *Message `json:"message,omitempty"`
	Error        *Error   `json:"error,omitempty"`
}

type Delta struct {
	Type        string `json:"type"`
	Text        string `json:"text,omitempty"`
	PartialJSON string `json:"partial_json,omitempty"`
}

type Error struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

func parseSSEResponse(ctx context.Context, reader io.Reader, handler func(resp *StreamEvent) error) error {
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

				var event StreamEvent
				if err := json.Unmarshal([]byte(data), &event); err != nil {
					continue
				}

				if err := handler(&event); err != nil {
					return err
				}
			}
		}
	}
}

func NewTextContent(text string) *TextContent {
	return &TextContent{Type: "text", Text: text}
}

func NewToolUseContent(id, name string, input json.RawMessage) *ToolUseContent {
	return &ToolUseContent{Type: "tool_use", ID: id, Name: name, Input: input}
}

func NewToolResultContent(toolUseID, content string, isError bool) *ToolResultContent {
	return &ToolResultContent{Type: "tool_result", ToolUseID: toolUseID, Content: content, IsError: isError}
}

func MessagesToBlocks(messages []Message) []ContentBlock {
	blocks := make([]ContentBlock, 0)
	for _, msg := range messages {
		for _, block := range msg.Content {
			blocks = append(blocks, block)
		}
	}
	return blocks
}
