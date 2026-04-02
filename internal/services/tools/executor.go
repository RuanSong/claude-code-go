package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// ToolUseBlock 工具调用块 - 表示一次工具调用
type ToolUseBlock struct {
	ID        string                 `json:"id"`        // 工具调用ID
	Name      string                 `json:"name"`      // 工具名称
	Input     map[string]interface{} `json:"input"`     // 工具输入参数
	Timestamp time.Time              `json:"timestamp"` // 调用时间戳
}

// ToolUseResult 工具执行结果
type ToolUseResult struct {
	ID        string        `json:"id"`              // 工具调用ID
	Name      string        `json:"name"`            // 工具名称
	Success   bool          `json:"success"`         // 是否成功
	Result    string        `json:"result"`          // 结果内容
	Error     string        `json:"error,omitempty"` // 错误信息
	Duration  time.Duration `json:"duration"`        // 执行时长
	Timestamp time.Time     `json:"timestamp"`       // 完成时间戳
}

// ToolExecutor 工具执行器 - 负责工具的编排和执行
type ToolExecutor struct {
	// 工具注册表
	tools map[string]Tool

	// 并发控制
	maxConcurrency int

	// 统计信息
	mu              sync.RWMutex
	totalExecutions int64
	totalErrors     int64
}

// Tool 工具接口 - 所有工具必须实现此接口
type Tool interface {
	// Name 返回工具名称
	Name() string

	// Description 返回工具描述
	Description() string

	// InputSchema 返回输入参数模式
	InputSchema() map[string]interface{}

	// Execute 执行工具
	Execute(ctx context.Context, input json.RawMessage, execCtx ToolExecContext) (*ToolResult, error)
}

// ToolExecContext 工具执行上下文
type ToolExecContext struct {
	WorkingDirectory string            // 工作目录
	Environment      map[string]string // 环境变量
	Tools            []Tool            // 可用工具列表
	UserID           string            // 用户ID
	SessionID        string            // 会话ID
}

// ToolResult 工具执行结果
type ToolResult struct {
	Content []ContentBlock `json:"content"`            // 结果内容块
	IsError bool           `json:"is_error,omitempty"` // 是否为错误
}

// ContentBlock 内容块
type ContentBlock struct {
	Type string `json:"type"`           // 内容类型: text, image, resource
	Text string `json:"text,omitempty"` // 文本内容
	Data string `json:"data,omitempty"` // 数据(用于图片等)
	URI  string `json:"uri,omitempty"`  // 资源URI
}

// TextBlock 文本内容块
type TextBlock struct {
	Text string `json:"text"`
}

// NewToolExecutor 创建新的工具执行器
func NewToolExecutor() *ToolExecutor {
	return &ToolExecutor{
		tools:          make(map[string]Tool),
		maxConcurrency: 10,
	}
}

// RegisterTool 注册工具
func (e *ToolExecutor) RegisterTool(tool Tool) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	name := tool.Name()
	if name == "" {
		return fmt.Errorf("工具名称不能为空")
	}

	if _, exists := e.tools[name]; exists {
		return fmt.Errorf("工具已注册: %s", name)
	}

	e.tools[name] = tool
	return nil
}

// UnregisterTool 注销工具
func (e *ToolExecutor) UnregisterTool(name string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if _, exists := e.tools[name]; !exists {
		return fmt.Errorf("工具未注册: %s", name)
	}

	delete(e.tools, name)
	return nil
}

// GetTool 获取工具
func (e *ToolExecutor) GetTool(name string) (Tool, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	tool, ok := e.tools[name]
	return tool, ok
}

// GetAllTools 获取所有已注册的工具
func (e *ToolExecutor) GetAllTools() []Tool {
	e.mu.RLock()
	defer e.mu.RUnlock()

	tools := make([]Tool, 0, len(e.tools))
	for _, tool := range e.tools {
		tools = append(tools, tool)
	}
	return tools
}

// SetMaxConcurrency 设置最大并发数
func (e *ToolExecutor) SetMaxConcurrency(n int) {
	if n > 0 {
		e.maxConcurrency = n
	}
}

// ExecuteTool 执行单个工具
func (e *ToolExecutor) ExecuteTool(ctx context.Context, block *ToolUseBlock, execCtx *ToolExecContext) *ToolUseResult {
	startTime := time.Now()

	// 获取工具
	tool, ok := e.GetTool(block.Name)
	if !ok {
		return &ToolUseResult{
			ID:        block.ID,
			Name:      block.Name,
			Success:   false,
			Error:     fmt.Sprintf("工具未找到: %s", block.Name),
			Duration:  time.Since(startTime),
			Timestamp: time.Now(),
		}
	}

	// 解析输入参数
	inputJSON, err := json.Marshal(block.Input)
	if err != nil {
		return &ToolUseResult{
			ID:        block.ID,
			Name:      block.Name,
			Success:   false,
			Error:     fmt.Sprintf("输入参数序列化失败: %v", err),
			Duration:  time.Since(startTime),
			Timestamp: time.Now(),
		}
	}

	// 执行工具
	result, err := tool.Execute(ctx, inputJSON, *execCtx)
	if err != nil {
		e.mu.Lock()
		e.totalErrors++
		e.mu.Unlock()

		return &ToolUseResult{
			ID:        block.ID,
			Name:      block.Name,
			Success:   false,
			Error:     fmt.Sprintf("执行失败: %v", err),
			Duration:  time.Since(startTime),
			Timestamp: time.Now(),
		}
	}

	// 处理结果
	content := ""
	if result != nil && len(result.Content) > 0 {
		for _, block := range result.Content {
			if block.Type == "text" {
				content += block.Text
			}
		}
	}

	if result != nil && result.IsError {
		e.mu.Lock()
		e.totalErrors++
		e.mu.Unlock()
	}

	e.mu.Lock()
	e.totalExecutions++
	e.mu.Unlock()

	return &ToolUseResult{
		ID:        block.ID,
		Name:      block.Name,
		Success:   result == nil || !result.IsError,
		Result:    content,
		Duration:  time.Since(startTime),
		Timestamp: time.Now(),
	}
}

// ExecuteToolsConcurrently 并发执行多个工具
func (e *ToolExecutor) ExecuteToolsConcurrently(ctx context.Context, blocks []*ToolUseBlock, execCtx *ToolExecContext) []*ToolUseResult {
	if len(blocks) == 0 {
		return nil
	}

	// 使用信号量控制并发数
	sem := make(chan struct{}, e.maxConcurrency)
	var wg sync.WaitGroup
	results := make([]*ToolUseResult, len(blocks))

	for i, block := range blocks {
		wg.Add(1)
		go func(index int, b *ToolUseBlock) {
			defer wg.Done()
			sem <- struct{}{}        // 获取信号量
			defer func() { <-sem }() // 释放信号量

			results[index] = e.ExecuteTool(ctx, b, execCtx)
		}(i, block)
	}

	wg.Wait()
	return results
}

// ExecuteToolsSerially 串行执行多个工具
func (e *ToolExecutor) ExecuteToolsSerially(ctx context.Context, blocks []*ToolUseBlock, execCtx *ToolExecContext) []*ToolUseResult {
	results := make([]*ToolUseResult, len(blocks))

	for i, block := range blocks {
		results[i] = e.ExecuteTool(ctx, block, execCtx)
	}

	return results
}

// IsReadOnlyTool 判断工具是否为只读
func (e *ToolExecutor) IsReadOnlyTool(name string) bool {
	tool, ok := e.GetTool(name)
	if !ok {
		return false
	}

	// 通过工具名称判断
	switch name {
	case "Read", "Glob", "Grep", "WebFetch", "WebSearch",
		"ListMcpResources", "ReadMcpResource", "LSPTool":
		return true
	}

	_ = tool
	return false
}

// PartitionToolCalls 将工具调用分区为可并发和需串行的批次
// 只读工具可以并发执行，非只读工具需串行执行
func (e *ToolExecutor) PartitionToolCalls(blocks []*ToolUseBlock) [][]*ToolUseBlock {
	if len(blocks) == 0 {
		return nil
	}

	var batches [][]*ToolUseBlock
	var currentBatch []*ToolUseBlock

	for _, block := range blocks {
		if e.IsReadOnlyTool(block.Name) {
			// 只读工具添加到当前批次
			currentBatch = append(currentBatch, block)
		} else {
			// 非只读工具：先保存当前批次，然后单独作为新批次
			if len(currentBatch) > 0 {
				batches = append(batches, currentBatch)
				currentBatch = nil
			}
			// 非只读工具单独一个批次
			batches = append(batches, []*ToolUseBlock{block})
		}
	}

	// 处理最后一批
	if len(currentBatch) > 0 {
		batches = append(batches, currentBatch)
	}

	return batches
}

// ExecuteTools 执行工具调用序列
// 自动处理并发和串行的分区
func (e *ToolExecutor) ExecuteTools(ctx context.Context, blocks []*ToolUseBlock, execCtx *ToolExecContext) []*ToolUseResult {
	if len(blocks) == 0 {
		return nil
	}

	partitions := e.PartitionToolCalls(blocks)
	var allResults []*ToolUseResult

	for _, partition := range partitions {
		if len(partition) == 1 && !e.IsReadOnlyTool(partition[0].Name) {
			// 单个非只读工具，串行执行
			results := e.ExecuteToolsSerially(ctx, partition, execCtx)
			allResults = append(allResults, results...)
		} else {
			// 只读工具批次，并发执行
			results := e.ExecuteToolsConcurrently(ctx, partition, execCtx)
			allResults = append(allResults, results...)
		}
	}

	return allResults
}

// GetStats 获取执行统计信息
func (e *ToolExecutor) GetStats() (executions, errors int64) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.totalExecutions, e.totalErrors
}

// ResetStats 重置统计信息
func (e *ToolExecutor) ResetStats() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.totalExecutions = 0
	e.totalErrors = 0
}
