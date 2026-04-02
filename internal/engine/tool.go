package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/claude-code-go/claude/pkg/schema"
)

// 权限模式
// 对应 TypeScript: 工具权限级别
// 定义工具执行所需的权限级别
type PermissionMode int

const (
	PermissionNormal   PermissionMode = iota // 普通权限 - 正常工具调用
	PermissionElevated                       //  elevated权限 - 需要提升权限
	PermissionReadonly                       // 只读权限 - 只能读取，不能修改
)

func (p PermissionMode) String() string {
	switch p {
	case PermissionNormal:
		return "normal"
	case PermissionElevated:
		return "elevated"
	case PermissionReadonly:
		return "readonly"
	default:
		return "unknown"
	}
}

// ToolResult 工具执行结果
// 对应 TypeScript: 工具返回结果
// 包含执行状态、输出内容和错误信息
type ToolResult struct {
	Content []ContentBlock // 结果内容块列表
	Error   error          // 执行错误
	IsError bool           // 是否为错误结果
}

// AddText 添加文本内容到结果
func (r *ToolResult) AddText(text string) {
	r.Content = append(r.Content, &TextBlock{Text: text})
}

// AddError 添加错误内容到结果
func (r *ToolResult) AddError(text string) {
	r.Content = append(r.Content, &TextBlock{Text: text})
	r.IsError = true
	r.Error = fmt.Errorf("%s", text)
}

// ToolExecContext 工具执行上下文
// 对应 TypeScript: 工具执行环境
// 提供工具执行所需的信息和环境
type ToolExecContext struct {
	Cwd         string                // 当前工作目录
	WorkingDir  string                // 工作目录
	Env         map[string]string     // 环境变量
	Tools       map[string]Tool       // 可用工具映射
	AbortSignal <-chan struct{}       // 中止信号通道
	OnProgress  func(progress string) // 进度回调函数
	Todos       []TodoItem            // 待办事项列表
}

// TodoItem 待办事项
// 对应 TypeScript: todo item
type TodoItem struct {
	Content    string `json:"content"`              // 内容描述
	Status     string `json:"status"`               // 状态 (in_progress/completed/pending)
	ActiveForm string `json:"activeForm,omitempty"` // 进行中的描述
}

func (ctx *ToolExecContext) GetTodos() []TodoItem {
	return ctx.Todos
}

func (ctx *ToolExecContext) SetTodos(todos []TodoItem) {
	ctx.Todos = todos
}

// Tool 工具接口
// 对应 TypeScript: 工具定义
// 所有工具必须实现的接口
type Tool interface {
	Name() string                                                                                     // 工具名称
	Description() string                                                                              // 工具描述
	InputSchema() schema.Schema                                                                       // 输入参数模式
	Permission() PermissionMode                                                                       // 所需权限级别
	Execute(ctx context.Context, input json.RawMessage, execCtx ToolExecContext) (*ToolResult, error) // 执行工具
}

// ToolRegistry 工具注册表
// 对应 TypeScript: 工具管理器
// 负责工具的注册、查找和列表
type ToolRegistry struct {
	mu    sync.RWMutex
	tools map[string]Tool // 工具名称 -> 工具实例
}

func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{
		tools: make(map[string]Tool),
	}
}

// Register 注册工具
// 对应 TypeScript: 注册工具
func (r *ToolRegistry) Register(tool Tool) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := tool.Name()
	if name == "" {
		return fmt.Errorf("tool name cannot be empty")
	}
	if _, exists := r.tools[name]; exists {
		return fmt.Errorf("tool already registered: %s", name)
	}
	r.tools[name] = tool
	return nil
}

// Get 获取工具
// 对应 TypeScript: 获取工具
func (r *ToolRegistry) Get(name string) (Tool, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	tool, ok := r.tools[name]
	return tool, ok
}

// List 列出所有工具
// 对应 TypeScript: 列出工具
func (r *ToolRegistry) List() []Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	tools := make([]Tool, 0, len(r.tools))
	for _, tool := range r.tools {
		tools = append(tools, tool)
	}
	return tools
}

// Names 获取所有工具名称
func (r *ToolRegistry) Names() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.tools))
	for name := range r.tools {
		names = append(names, name)
	}
	return names
}

// ContentBlock 内容块接口
// 对应 TypeScript: 消息内容块
// 定义消息中不同类型内容的接口
type ContentBlock interface {
	blockType() string
}

type TextBlock struct {
	Type string `json:"type"` // "text"
	Text string `json:"text"` // 文本内容
}

func (t *TextBlock) blockType() string { return "text" }

type ToolUseBlock struct {
	Type  string          `json:"type"`  // "tool_use"
	ID    string          `json:"id"`    // 工具调用ID
	Name  string          `json:"name"`  // 工具名称
	Input json.RawMessage `json:"input"` // 工具输入参数
}

func (t *ToolUseBlock) blockType() string { return "tool_use" }

type ToolResultBlock struct {
	Type      string `json:"type"`               // "tool_result"
	ToolUseID string `json:"tool_use_id"`        // 关联的工具调用ID
	Content   any    `json:"content"`            // 结果内容
	IsError   bool   `json:"is_error,omitempty"` // 是否为错误
}

func (t *ToolResultBlock) blockType() string { return "tool_result" }

type ImageBlock struct {
	Type      string `json:"type"`                 // "image"
	Source    string `json:"source"`               // 图片源（base64或URL）
	MediaType string `json:"media_type,omitempty"` // 媒体类型
}

func (i *ImageBlock) blockType() string { return "image" }
