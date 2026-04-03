package llm

import (
	"context"
	"encoding/json"
)

// ProviderCapabilities 提供者能力
type ProviderCapabilities struct {
	SupportsTools    bool // 是否支持工具调用
	SupportsVision   bool // 是否支持视觉理解
	SupportsStream   bool // 是否支持流式输出
	SupportsThinking bool // 是否支持思考模式 (R1/R2系列)
}

// ProviderConfig 提供者配置
type ProviderConfig struct {
	Name         string               // 提供者名称
	BaseURL      string               // API 基础URL
	APIKey       string               // API密钥 (如果为空，使用 APIKeyEnv)
	APIKeyEnv    string               // API密钥环境变量名
	DefaultModel string               // 默认模型
	CodingModel  string               // Coding专用模型
	Capabilities ProviderCapabilities // 提供者能力
	Extra        map[string]any       // 额外配置
}

// LLMProvider LLM提供者接口
// 所有支持的 LLM Provider 都必须实现此接口
type LLMProvider interface {
	// Name 返回提供者名称
	Name() string

	// Config 返回提供者配置
	Config() ProviderConfig

	// SupportsTools 是否支持工具调用
	SupportsTools() bool

	// SupportsVision 是否支持视觉理解
	SupportsVision() bool

	// CreateMessage 创建消息 (非流式)
	CreateMessage(ctx context.Context, req *MessageRequest) (*MessageResponse, error)

	// CreateMessageStream 创建消息 (流式)
	CreateMessageStream(ctx context.Context, req *MessageRequest, handler func(*StreamEvent) error) error
}

// MessageRequest 消息请求
type MessageRequest struct {
	Model        string          // 模型名称
	Messages     []Message       // 消息列表
	SystemPrompt string          // 系统提示
	MaxTokens    int             // 最大token数
	Temperature  float64         // 温度
	Tools        []ToolDef       // 工具定义
	ToolChoice   string          // 工具选择
	Stream       bool            // 是否流式
	Thinking     *ThinkingConfig // 思考配置 (可选)
}

// Message 消息
type Message struct {
	Role    string // role: system, user, assistant, tool
	Content any    // content: string 或 []ContentBlock
}

// ContentBlock 内容块
type ContentBlock interface{}

// TextContent 文本内容块
type TextContent struct {
	Type string `json:"type"` // text
	Text string `json:"text"`
}

// ImageContent 图片内容块
type ImageContent struct {
	Type     string   `json:"type"` // image_url
	ImageURL ImageURL `json:"image_url"`
}

// ImageURL 图片URL
type ImageURL struct {
	URL string `json:"url"`
}

// ToolUseContent 工具使用内容块
type ToolUseContent struct {
	Type  string          `json:"type"` // tool_use
	ID    string          `json:"id"`
	Name  string          `json:"name"`
	Input json.RawMessage `json:"input"`
}

// ToolResultContent 工具结果内容块
type ToolResultContent struct {
	Type      string `json:"type"` // tool_result
	ToolUseID string `json:"tool_use_id"`
	Content   string `json:"content"`
	IsError   bool   `json:"is_error,omitempty"`
}

// ToolDef 工具定义
type ToolDef struct {
	Type        string          `json:"type"` // function
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  json.RawMessage `json:"parameters"`
}

// ThinkingConfig 思考配置
type ThinkingConfig struct {
	Type         string `json:"type"` // enabled
	BudgetTokens int    `json:"budget_tokens"`
}

// MessageResponse 消息响应
type MessageResponse struct {
	ID           string         `json:"id"`
	Type         string         `json:"type"`
	Role         string         `json:"role"`
	Content      []any          `json:"content"`
	Model        string         `json:"model"`
	StopReason   string         `json:"stop_reason"`
	StopSequence any            `json:"stop_sequence,omitempty"`
	Usage        Usage          `json:"usage"`
	Error        *ResponseError `json:"error,omitempty"`
}

// Usage 使用量
type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
	TotalTokens  int `json:"total_tokens"`
}

// ResponseError 响应错误
type ResponseError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

// StreamEvent 流式事件
type StreamEvent struct {
	Type         string         `json:"type"`
	Index        int            `json:"index,omitempty"`
	ContentBlock ContentBlock   `json:"content_block,omitempty"`
	Delta        *StreamDelta   `json:"delta,omitempty"`
	Message      *Message       `json:"message,omitempty"`
	Error        *ResponseError `json:"error,omitempty"`
}

// StreamDelta 流式增量
type StreamDelta struct {
	Type        string `json:"type"` // content_block_delta, input_json_delta, message_delta, message_stop
	Text        string `json:"text,omitempty"`
	PartialJSON string `json:"partial_json,omitempty"`
	Usage       *Usage `json:"usage,omitempty"`
}

// NewTextContent 创建文本内容块
func NewTextContent(text string) *TextContent {
	return &TextContent{Type: "text", Text: text}
}

// NewImageContent 创建图片内容块
func NewImageContent(url string) *ImageContent {
	return &ImageContent{Type: "image_url", ImageURL: ImageURL{URL: url}}
}

// NewToolUseContent 创建工具使用内容块
func NewToolUseContent(id, name string, input json.RawMessage) *ToolUseContent {
	return &ToolUseContent{Type: "tool_use", ID: id, Name: name, Input: input}
}

// NewToolResultContent 创建工具结果内容块
func NewToolResultContent(toolUseID, content string, isError bool) *ToolResultContent {
	return &ToolResultContent{Type: "tool_result", ToolUseID: toolUseID, Content: content, IsError: isError}
}

// ProviderRegistry 提供者注册表
type ProviderRegistry struct {
	providers       map[string]LLMProvider
	defaultProvider string
}

// NewProviderRegistry 创建提供者注册表
func NewProviderRegistry() *ProviderRegistry {
	return &ProviderRegistry{
		providers: make(map[string]LLMProvider),
	}
}

// Register 注册提供者
func (r *ProviderRegistry) Register(name string, provider LLMProvider) {
	r.providers[name] = provider
}

// Get 获取提供者
func (r *ProviderRegistry) Get(name string) (LLMProvider, bool) {
	p, ok := r.providers[name]
	return p, ok
}

// List 列出所有提供者
func (r *ProviderRegistry) List() []string {
	names := make([]string, 0, len(r.providers))
	for name := range r.providers {
		names = append(names, name)
	}
	return names
}

// SetDefault 设置默认提供者
func (r *ProviderRegistry) SetDefault(name string) error {
	if _, ok := r.providers[name]; !ok {
		return &ProviderNotFoundError{Name: name}
	}
	r.defaultProvider = name
	return nil
}

// GetDefault 获取默认提供者
func (r *ProviderRegistry) GetDefault() (LLMProvider, error) {
	if r.defaultProvider == "" {
		return nil, &NoDefaultProviderError{}
	}
	p, ok := r.providers[r.defaultProvider]
	if !ok {
		return nil, &ProviderNotFoundError{Name: r.defaultProvider}
	}
	return p, nil
}

// ProviderNotFoundError 提供者未找到错误
type ProviderNotFoundError struct {
	Name string
}

func (e *ProviderNotFoundError) Error() string {
	return "provider not found: " + e.Name
}

// NoDefaultProviderError 没有默认提供者错误
type NoDefaultProviderError struct{}

func (e *NoDefaultProviderError) Error() string {
	return "no default provider set"
}
