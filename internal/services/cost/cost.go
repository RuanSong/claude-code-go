package cost

import (
	"fmt"
	"sync"
	"time"
)

// 模型成本配置
type ModelCosts struct {
	InputTokens            float64 // 输入令牌每百万成本
	OutputTokens           float64 // 输出令牌每百万成本
	PromptCacheWriteTokens float64 // 提示缓存写入每百万成本
	PromptCacheReadTokens  float64 // 提示缓存读取每百万成本
	WebSearchRequests      float64 // 网页搜索请求成本
}

// 标准定价层: Sonnet 模型 $3 输入 / $15 输出 每百万令牌
var COST_TIER_3_15 = ModelCosts{
	InputTokens:            3.0,
	OutputTokens:           15.0,
	PromptCacheWriteTokens: 3.75,
	PromptCacheReadTokens:  0.3,
	WebSearchRequests:      0.01,
}

// Opus 4/4.1 定价层: $15 输入 / $75 输出 每百万令牌
var COST_TIER_15_75 = ModelCosts{
	InputTokens:            15.0,
	OutputTokens:           75.0,
	PromptCacheWriteTokens: 18.75,
	PromptCacheReadTokens:  1.5,
	WebSearchRequests:      0.01,
}

// Opus 4.5 定价层: $5 输入 / $25 输出 每百万令牌
var COST_TIER_5_25 = ModelCosts{
	InputTokens:            5.0,
	OutputTokens:           25.0,
	PromptCacheWriteTokens: 6.25,
	PromptCacheReadTokens:  0.5,
	WebSearchRequests:      0.01,
}

// Haiku 3.5 定价: $0.80 输入 / $4 输出 每百万令牌
var COST_HAIKU_35 = ModelCosts{
	InputTokens:            0.8,
	OutputTokens:           4.0,
	PromptCacheWriteTokens: 1.0,
	PromptCacheReadTokens:  0.08,
	WebSearchRequests:      0.01,
}

// Haiku 4.5 定价: $1 输入 / $5 输出 每百万令牌
var COST_HAIKU_45 = ModelCosts{
	InputTokens:            1.0,
	OutputTokens:           5.0,
	PromptCacheWriteTokens: 1.25,
	PromptCacheReadTokens:  0.1,
	WebSearchRequests:      0.01,
}

// 默认未知模型定价
var DEFAULT_UNKNOWN_MODEL_COST = COST_TIER_5_25

// 模型定价映射表
var MODEL_COSTS = map[string]ModelCosts{
	"claude-3-5-haiku-20241022":  COST_HAIKU_35,
	"claude-3-haiku-20241022":    COST_HAIKU_45,
	"claude-3-5-sonnet-20241022": COST_TIER_3_15,
	"claude-3-7-sonnet-20241022": COST_TIER_3_15,
	"claude-sonnet-4-20250514":   COST_TIER_3_15,
	"claude-sonnet-4-20241120":   COST_TIER_3_15,
	"claude-sonnet-4-5-20250620": COST_TIER_3_15,
	"claude-opus-4-20250514":     COST_TIER_15_75,
	"claude-opus-4-20241120":     COST_TIER_15_75,
	"claude-opus-4-5-20251114":   COST_TIER_5_25,
	"claude-opus-4-6-20251114":   COST_TIER_5_25,
}

// TokenUsage Token使用量统计
type TokenUsage struct {
	InputTokens              int64 // 输入令牌数
	OutputTokens             int64 // 输出令牌数
	CacheReadInputTokens     int64 // 缓存读取的输入令牌
	CacheCreationInputTokens int64 // 缓存创建的输入令牌
	WebSearchRequests        int64 // 网页搜索请求数
}

// ModelUsage 模型使用量统计
type ModelUsage struct {
	Usage      TokenUsage
	CostUSD    float64
	DurationMs int64
}

// CostTracker 成本追踪器 - 追踪API使用成本
type CostTracker struct {
	mu sync.RWMutex

	// 累计成本统计
	totalCostUSD                   float64
	totalAPIDuration               int64 // API调用总时长(毫秒)
	totalAPIDurationWithoutRetries int64 // 不含重试的API调用总时长
	totalToolDuration              int64 // 工具执行总时长(毫秒)
	totalLinesAdded                int64 // 新增代码行数
	totalLinesRemoved              int64 // 删除代码行数
	totalInputTokens               int64
	totalOutputTokens              int64
	totalCacheReadInputTokens      int64
	totalCacheCreationInputTokens  int64
	totalWebSearchRequests         int64

	// 每个模型的使用量
	modelUsage map[string]*ModelUsage

	// 未知模型标记
	hasUnknownModelCost bool

	// 会话信息
	sessionID string
}

// NewCostTracker 创建新的成本追踪器
func NewCostTracker(sessionID string) *CostTracker {
	return &CostTracker{
		modelUsage: make(map[string]*ModelUsage),
		sessionID:  sessionID,
	}
}

// AddUsage 添加模型使用量
func (c *CostTracker) AddUsage(model string, usage *TokenUsage, durationMs int64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 计算成本
	cost := c.calculateCost(model, usage)

	// 更新累计统计
	c.totalInputTokens += usage.InputTokens
	c.totalOutputTokens += usage.OutputTokens
	c.totalCacheReadInputTokens += usage.CacheReadInputTokens
	c.totalCacheCreationInputTokens += usage.CacheCreationInputTokens
	c.totalWebSearchRequests += usage.WebSearchRequests
	c.totalAPIDuration += durationMs

	// 更新模型特定统计
	if _, ok := c.modelUsage[model]; !ok {
		c.modelUsage[model] = &ModelUsage{}
	}
	c.modelUsage[model].Usage.InputTokens += usage.InputTokens
	c.modelUsage[model].Usage.OutputTokens += usage.OutputTokens
	c.modelUsage[model].Usage.CacheReadInputTokens += usage.CacheReadInputTokens
	c.modelUsage[model].Usage.CacheCreationInputTokens += usage.CacheCreationInputTokens
	c.modelUsage[model].Usage.WebSearchRequests += usage.WebSearchRequests
	c.modelUsage[model].CostUSD += cost
	c.modelUsage[model].DurationMs += durationMs
}

// calculateCost 计算令牌使用的成本
func (c *CostTracker) calculateCost(model string, usage *TokenUsage) float64 {
	costs, ok := MODEL_COSTS[model]
	if !ok {
		costs = DEFAULT_UNKNOWN_MODEL_COST
		c.hasUnknownModelCost = true
	}

	// 计算成本: (令牌数/1_000_000) * 每百万成本
	cost := (float64(usage.InputTokens) / 1_000_000) * costs.InputTokens
	cost += (float64(usage.OutputTokens) / 1_000_000) * costs.OutputTokens
	cost += (float64(usage.CacheReadInputTokens) / 1_000_000) * costs.PromptCacheReadTokens
	cost += (float64(usage.CacheCreationInputTokens) / 1_000_000) * costs.PromptCacheWriteTokens
	cost += float64(usage.WebSearchRequests) * costs.WebSearchRequests

	return cost
}

// AddLinesChanged 添加代码行数变化
func (c *CostTracker) AddLinesChanged(linesAdded, linesRemoved int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.totalLinesAdded += linesAdded
	c.totalLinesRemoved += linesRemoved
}

// AddToolDuration 添加工具执行时长
func (c *CostTracker) AddToolDuration(durationMs int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.totalToolDuration += durationMs
}

// AddAPIDuration 添加API调用时长
func (c *CostTracker) AddAPIDuration(durationMs int64, withoutRetry bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.totalAPIDuration += durationMs
	if withoutRetry {
		c.totalAPIDurationWithoutRetries += durationMs
	}
}

// GetTotalCost 获取总成本(美元)
func (c *CostTracker) GetTotalCost() float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.totalCostUSD
}

// GetTotalDuration 获取总执行时长
func (c *CostTracker) GetTotalDuration() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.totalAPIDuration
}

// GetTotalAPIDuration 获取API调用总时长
func (c *CostTracker) GetTotalAPIDuration() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.totalAPIDuration
}

// GetTotalAPIDurationWithoutRetries 获取不含重试的API调用总时长
func (c *CostTracker) GetTotalAPIDurationWithoutRetries() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.totalAPIDurationWithoutRetries
}

// GetTotalLinesAdded 获取新增代码行数
func (c *CostTracker) GetTotalLinesAdded() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.totalLinesAdded
}

// GetTotalLinesRemoved 获取删除代码行数
func (c *CostTracker) GetTotalLinesRemoved() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.totalLinesRemoved
}

// GetTotalInputTokens 获取输入令牌总数
func (c *CostTracker) GetTotalInputTokens() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.totalInputTokens
}

// GetTotalOutputTokens 获取输出令牌总数
func (c *CostTracker) GetTotalOutputTokens() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.totalOutputTokens
}

// GetTotalCacheReadInputTokens 获取缓存读取输入令牌总数
func (c *CostTracker) GetTotalCacheReadInputTokens() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.totalCacheReadInputTokens
}

// GetTotalCacheCreationInputTokens 获取缓存创建输入令牌总数
func (c *CostTracker) GetTotalCacheCreationInputTokens() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.totalCacheCreationInputTokens
}

// GetTotalWebSearchRequests 获取网页搜索请求总数
func (c *CostTracker) GetTotalWebSearchRequests() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.totalWebSearchRequests
}

// GetModelUsage 获取指定模型的使用量
func (c *CostTracker) GetModelUsage(model string) *ModelUsage {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.modelUsage[model]
}

// GetAllModelUsage 获取所有模型的使用量
func (c *CostTracker) GetAllModelUsage() map[string]*ModelUsage {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make(map[string]*ModelUsage)
	for k, v := range c.modelUsage {
		result[k] = v
	}
	return result
}

// HasUnknownModelCost 是否有未知模型成本
func (c *CostTracker) HasUnknownModelCost() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.hasUnknownModelCost
}

// FormatCost 格式化成本显示
func (c *CostTracker) FormatCost() string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	cost := c.totalCostUSD
	if cost == 0 {
		return "$0"
	}

	// 整数无小数位，其他保留2位小数
	if cost == float64(int64(cost)) {
		return fmt.Sprintf("$%.0f", cost)
	}
	return fmt.Sprintf("$%.2f", cost)
}

// FormatDuration 格式化时长显示
func FormatDuration(durationMs int64) string {
	if durationMs < 1000 {
		return fmt.Sprintf("%dms", durationMs)
	}
	seconds := durationMs / 1000
	if seconds < 60 {
		return fmt.Sprintf("%.1fs", float64(durationMs)/1000.0)
	}
	minutes := seconds / 60
	remainingSeconds := seconds % 60
	if minutes < 60 {
		return fmt.Sprintf("%dm %ds", minutes, remainingSeconds)
	}
	hours := minutes / 60
	remainingMinutes := minutes % 60
	return fmt.Sprintf("%dh %dm", hours, remainingMinutes)
}

// GetCostSummary 获取成本摘要
func (c *CostTracker) GetCostSummary() string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	summary := fmt.Sprintf(`
成本摘要:
  总成本: %s
  API调用时长: %s
  工具执行时长: %s
  代码变更: +%d / -%d 行
  令牌使用:
    输入: %d
    输出: %d
    缓存读取: %d
    缓存创建: %d
    网页搜索: %d`,
		c.FormatCostLocked(),
		FormatDuration(c.totalAPIDuration),
		FormatDuration(c.totalToolDuration),
		c.totalLinesAdded,
		c.totalLinesRemoved,
		c.totalInputTokens,
		c.totalOutputTokens,
		c.totalCacheReadInputTokens,
		c.totalCacheCreationInputTokens,
		c.totalWebSearchRequests,
	)

	// 添加每个模型的详细使用量
	if len(c.modelUsage) > 0 {
		summary += "\n\n模型使用详情:"
		for model, usage := range c.modelUsage {
			summary += fmt.Sprintf("\n  %s:", model)
			summary += fmt.Sprintf("\n    成本: $%.4f", usage.CostUSD)
			summary += fmt.Sprintf("\n    输入令牌: %d", usage.Usage.InputTokens)
			summary += fmt.Sprintf("\n    输出令牌: %d", usage.Usage.OutputTokens)
		}
	}

	return summary
}

// FormatCostLocked 格式化成本显示(内部锁定版本)
func (c *CostTracker) FormatCostLocked() string {
	cost := c.totalCostUSD
	if cost == 0 {
		return "$0"
	}
	if cost == float64(int64(cost)) {
		return fmt.Sprintf("$%.0f", cost)
	}
	return fmt.Sprintf("$%.2f", cost)
}

// Reset 重置所有统计
func (c *CostTracker) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.totalCostUSD = 0
	c.totalAPIDuration = 0
	c.totalAPIDurationWithoutRetries = 0
	c.totalToolDuration = 0
	c.totalLinesAdded = 0
	c.totalLinesRemoved = 0
	c.totalInputTokens = 0
	c.totalOutputTokens = 0
	c.totalCacheReadInputTokens = 0
	c.totalCacheCreationInputTokens = 0
	c.totalWebSearchRequests = 0
	c.modelUsage = make(map[string]*ModelUsage)
	c.hasUnknownModelCost = false
}

// CostState 成本状态快照 - 用于保存和恢复
type CostState struct {
	TotalCostUSD                   float64
	TotalAPIDuration               int64
	TotalAPIDurationWithoutRetries int64
	TotalToolDuration              int64
	TotalLinesAdded                int64
	TotalLinesRemoved              int64
	TotalInputTokens               int64
	TotalOutputTokens              int64
	TotalCacheReadInputTokens      int64
	TotalCacheCreationInputTokens  int64
	TotalWebSearchRequests         int64
	ModelUsage                     map[string]*ModelUsage
	LastUpdated                    time.Time
}

// GetState 获取当前状态快照
func (c *CostTracker) GetState() *CostState {
	c.mu.RLock()
	defer c.mu.RUnlock()

	modelUsage := make(map[string]*ModelUsage)
	for k, v := range c.modelUsage {
		modelUsage[k] = v
	}

	return &CostState{
		TotalCostUSD:                   c.totalCostUSD,
		TotalAPIDuration:               c.totalAPIDuration,
		TotalAPIDurationWithoutRetries: c.totalAPIDurationWithoutRetries,
		TotalToolDuration:              c.totalToolDuration,
		TotalLinesAdded:                c.totalLinesAdded,
		TotalLinesRemoved:              c.totalLinesRemoved,
		TotalInputTokens:               c.totalInputTokens,
		TotalOutputTokens:              c.totalOutputTokens,
		TotalCacheReadInputTokens:      c.totalCacheReadInputTokens,
		TotalCacheCreationInputTokens:  c.totalCacheCreationInputTokens,
		TotalWebSearchRequests:         c.totalWebSearchRequests,
		ModelUsage:                     modelUsage,
		LastUpdated:                    time.Now(),
	}
}

// RestoreFromState 从状态快照恢复
func (c *CostTracker) RestoreFromState(state *CostState) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.totalCostUSD = state.TotalCostUSD
	c.totalAPIDuration = state.TotalAPIDuration
	c.totalAPIDurationWithoutRetries = state.TotalAPIDurationWithoutRetries
	c.totalToolDuration = state.TotalToolDuration
	c.totalLinesAdded = state.TotalLinesAdded
	c.totalLinesRemoved = state.TotalLinesRemoved
	c.totalInputTokens = state.TotalInputTokens
	c.totalOutputTokens = state.TotalOutputTokens
	c.totalCacheReadInputTokens = state.TotalCacheReadInputTokens
	c.totalCacheCreationInputTokens = state.TotalCacheCreationInputTokens
	c.totalWebSearchRequests = state.TotalWebSearchRequests
	c.modelUsage = state.ModelUsage
}
