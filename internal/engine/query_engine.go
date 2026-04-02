package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/claude-code-go/claude/pkg/anthropic"
)

// QueryEngine 查询引擎
// 对应 TypeScript: QueryEngine 或类似的查询处理核心
// 是处理用户查询、执行工具调用、管理对话上下文的核心引擎
type QueryEngine struct {
	config     Config             // 引擎配置
	context    *ContextManager    // 对话上下文管理器
	tools      *ToolRegistry      // 工具注册表
	llmClient  *anthropic.Client  // Anthropic API客户端
	permission *PermissionManager // 权限管理器
	turnCount  int                // 当前对话轮次计数
	mu         sync.Mutex         // 互斥锁，保护共享状态
}

// NewQueryEngine 创建新的查询引擎
// 对应 TypeScript: QueryEngine 构造函数
// 初始化上下文管理器、权限管理器等组件
func NewQueryEngine(config Config, client *anthropic.Client) *QueryEngine {
	return &QueryEngine{
		config:     config,
		context:    NewContextManager(config.SystemPrompt, config.MaxTokens),
		tools:      config.Tools,
		llmClient:  client,
		permission: NewPermissionManager(),
	}
}

// SubmitMessage 处理用户消息并返回响应
// 对应 TypeScript: 消息提交入口
// 1. 检查是否超出最大轮次限制
// 2. 处理slash命令（以/开头的命令）
// 3. 将用户消息添加到上下文
// 4. 检查是否需要上下文压缩
// 5. 进入主查询循环
func (e *QueryEngine) SubmitMessage(ctx context.Context, userInput string) error {
	e.mu.Lock()
	e.turnCount++
	if e.turnCount > e.config.MaxTurns && e.config.MaxTurns > 0 {
		e.mu.Unlock()
		return fmt.Errorf("max turns exceeded: %d", e.config.MaxTurns)
	}
	e.mu.Unlock()

	// 检查是否为slash命令
	if strings.HasPrefix(userInput, "/") {
		return e.handleSlashCommand(ctx, userInput)
	}

	// 添加用户消息到上下文
	userMsg := NewUserMessage(userInput)
	e.context.AddMessage(userMsg)

	// 检查是否需要上下文压缩
	if e.context.NeedsCompaction() {
		if err := e.context.Compact(ctx); err != nil {
			return fmt.Errorf("compact context: %w", err)
		}
	}

	// 主查询循环
	return e.queryLoop(ctx)
}

// handleSlashCommand 处理slash命令
// 对应 TypeScript: slash命令处理
// 解析命令名称和参数，调用注册的命令处理器
func (e *QueryEngine) handleSlashCommand(ctx context.Context, input string) error {
	parts := strings.SplitN(input, " ", 2)
	cmdName := strings.TrimPrefix(parts[0], "/")
	var args []string
	if len(parts) > 1 {
		args = strings.Split(parts[1], " ")
	}

	cmd, exists := e.config.Commands.Get(cmdName)
	if !exists {
		return fmt.Errorf("unknown command: /%s", cmdName)
	}

	execCtx := CommandContext{
		Cwd: e.config.Cwd,
	}

	return cmd.Execute(ctx, args, execCtx)
}

// queryLoop 主查询循环
// 对应 TypeScript: 查询循环
// 1. 构建API请求
// 2. 发送请求到LLM
// 3. 更新令牌使用量
// 4. 解析响应并添加到上下文
// 5. 检查是否有工具调用
// 6. 执行工具调用并返回结果
// 7. 检查预算限制
func (e *QueryEngine) queryLoop(ctx context.Context) error {
	for {
		// 构建API请求
		req := e.buildRequest()

		// 发送请求到LLM
		resp, err := e.llmClient.CreateMessage(ctx, req)
		if err != nil {
			return fmt.Errorf("LLM query failed: %w", err)
		}

		// 更新令牌使用量
		e.context.UpdateUsage(TokenUsage{
			InputTokens:  resp.Usage.InputTokens,
			OutputTokens: resp.Usage.OutputTokens,
			TotalTokens:  resp.Usage.InputTokens + resp.Usage.OutputTokens,
		})

		// 解析响应并添加到上下文
		assistantMsg := e.parseResponse(resp)
		e.context.AddMessage(assistantMsg)

		// 检查是否有工具调用
		if !assistantMsg.HasToolCalls() {
			return nil // 无工具调用，完成
		}

		// 执行工具调用
		for _, toolCall := range assistantMsg.GetToolCalls() {
			result, err := e.executeTool(ctx, toolCall)
			if err != nil {
				result = &ToolResult{
					Content: []ContentBlock{&TextBlock{Text: fmt.Sprintf("Error: %v", err)}},
					IsError: true,
				}
			}

			// 添加工具结果到上下文
			var resultContent any
			if len(result.Content) > 0 {
				if tb, ok := result.Content[0].(*TextBlock); ok {
					resultContent = tb.Text
				}
			}
			toolResultMsg := NewToolResultMessage(toolCall.ID, resultContent, result.IsError)
			e.context.AddMessage(toolResultMsg)
		}

		// 检查预算
		if e.config.MaxBudgetUSD > 0 {
			cost := e.estimateCost()
			if cost > e.config.MaxBudgetUSD {
				return fmt.Errorf("budget exceeded: $%.2f > $%.2f", cost, e.config.MaxBudgetUSD)
			}
		}
	}
}

// buildRequest 从当前上下文构建API请求
// 对应 TypeScript: 构建API请求
// 将内部消息格式转换为Anthropic API格式
func (e *QueryEngine) buildRequest() *anthropic.CreateMessageRequest {
	messages := e.context.GetMessages()

	// 转换为API消息格式
	apiMessages := make([]anthropic.Message, 0, len(messages))
	for _, msg := range messages {
		blocks := make([]anthropic.ContentBlock, 0, len(msg.Content))
		for _, block := range msg.Content {
			switch b := block.(type) {
			case *TextBlock:
				blocks = append(blocks, anthropic.NewTextContent(b.Text))
			case *ToolUseBlock:
				blocks = append(blocks, anthropic.NewToolUseContent(b.ID, b.Name, b.Input))
			case *ToolResultBlock:
				content := ""
				if c, ok := b.Content.(string); ok {
					content = c
				}
				blocks = append(blocks, anthropic.NewToolResultContent(b.ToolUseID, content, b.IsError))
			}
		}
		apiMessages = append(apiMessages, anthropic.Message{
			Role:    msg.Role,
			Content: blocks,
		})
	}

	// 转换工具为API格式
	tools := make([]anthropic.ToolDef, 0)
	for _, tool := range e.tools.List() {
		inputSchema, _ := json.Marshal(tool.InputSchema())
		tools = append(tools, anthropic.ToolDef{
			Name:        tool.Name(),
			Description: tool.Description(),
			InputSchema: inputSchema,
		})
	}

	return &anthropic.CreateMessageRequest{
		Model:     e.config.Model,
		Messages:  apiMessages,
		MaxTokens: 4096,
		System:    e.context.GetSystemPrompt(),
		Tools:     tools,
	}
}

// parseResponse 解析API响应
// 对应 TypeScript: 响应解析
// 将API返回的content blocks转换为内部消息格式
func (e *QueryEngine) parseResponse(resp *anthropic.CreateMessageResponse) *Message {
	content := make([]ContentBlock, 0, len(resp.Content))
	for _, block := range resp.Content {
		switch b := block.(type) {
		case map[string]any:
			if t, ok := b["type"].(string); ok {
				switch t {
				case "text":
					if text, ok := b["text"].(string); ok {
						content = append(content, &TextBlock{Type: "text", Text: text})
					}
				case "tool_use":
					if id, ok := b["id"].(string); ok {
						if name, ok := b["name"].(string); ok {
							if input, ok := b["input"].(map[string]any); ok {
								inputJSON, _ := json.Marshal(input)
								content = append(content, &ToolUseBlock{
									Type:  "tool_use",
									ID:    id,
									Name:  name,
									Input: inputJSON,
								})
							}
						}
					}
				}
			}
		}
	}

	return &Message{
		Type:    MessageTypeAssistant,
		Role:    RoleAssistant,
		Content: content,
	}
}

// executeTool 执行工具调用
// 对应 TypeScript: 工具执行
// 1. 检查权限
// 2. 获取工具
// 3. 执行工具并返回结果
func (e *QueryEngine) executeTool(ctx context.Context, toolCall *ToolUseBlock) (*ToolResult, error) {
	// 检查权限
	allowed, err := e.permission.Check(ctx, toolCall.Name)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return &ToolResult{
			Content: []ContentBlock{&TextBlock{Text: "Permission denied"}},
			IsError: true,
		}, nil
	}

	// 获取工具
	tool, exists := e.tools.Get(toolCall.Name)
	if !exists {
		return nil, fmt.Errorf("tool not found: %s", toolCall.Name)
	}

	// 执行工具
	cwd := ""
	if tools := e.tools.List(); len(tools) > 0 {
		cwd = tools[0].Name()
	}
	execCtx := ToolExecContext{
		Cwd: cwd,
	}

	return tool.Execute(ctx, toolCall.Input, execCtx)
}

// estimateCost 估算当前会话的成本
// 对应 TypeScript: 成本估算
// 基于令牌使用量和模型定价计算美元成本
func (e *QueryEngine) estimateCost() float64 {
	usage := e.context.GetUsage()
	// 近似定价: $3/MTok 输入, $15/MTok 输出
	inputCost := float64(usage.InputTokens) / 1_000_000 * 3
	outputCost := float64(usage.OutputTokens) / 1_000_000 * 15
	return inputCost + outputCost
}

// GetContext 获取当前上下文管理器
func (e *QueryEngine) GetContext() *ContextManager {
	return e.context
}

// Reset 重置引擎状态
// 对应 TypeScript: 重置会话
// 清空上下文和轮次计数
func (e *QueryEngine) Reset() {
	e.context.Clear()
	e.turnCount = 0
}
