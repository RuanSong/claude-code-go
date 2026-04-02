package token

import (
	"encoding/json"
	"strings"
)

// TokenEstimator 令牌计数和成本估算服务
// 对应 TypeScript: src/services/tokenEstimation.ts
// 提供令牌数量估算和API调用成本计算功能
type TokenEstimator struct{}

// NewTokenEstimator 创建一个新的令牌估算器实例
// 用于计算文本的令牌数量和估算API调用成本
func NewTokenEstimator() *TokenEstimator {
	return &TokenEstimator{}
}

// TokenUsage 令牌使用量统计
// 追踪一次请求或会话中的输入/输出令牌使用情况
type TokenUsage struct {
	InputTokens  int `json:"inputTokens"`  // 输入令牌数（发送的提示词）
	OutputTokens int `json:"outputTokens"` // 输出令牌数（AI生成的回复）
	TotalTokens  int `json:"totalTokens"`  // 总令牌数
}

// ModelPricing 模型定价信息
// 存储不同AI模型的每百万令牌调用价格
type ModelPricing struct {
	InputCostPerMillion  float64 `json:"inputCostPerMillion"`  // 输入每百万令牌价格（美元）
	OutputCostPerMillion float64 `json:"outputCostPerMillion"` // 输出每百万令牌价格（美元）
}

// modelPricing 支持的模型定价表
// 对应 TypeScript: 模型定价常量
var modelPricing = map[string]ModelPricing{
	"claude-opus-4-20250514":     {InputCostPerMillion: 15.0, OutputCostPerMillion: 75.0},
	"claude-sonnet-4-20250514":   {InputCostPerMillion: 3.0, OutputCostPerMillion: 15.0},
	"claude-3-5-sonnet-20241022": {InputCostPerMillion: 3.0, OutputCostPerMillion: 15.0},
	"claude-3-opus-20240229":     {InputCostPerMillion: 15.0, OutputCostPerMillion: 75.0},
	"claude-3-sonnet-20240229":   {InputCostPerMillion: 3.0, OutputCostPerMillion: 15.0},
	"claude-3-haiku-20240307":    {InputCostPerMillion: 0.25, OutputCostPerMillion: 1.25},
}

// CountTokens 计算文本的令牌数量（粗略估算）
// 对应 TypeScript: roughTokenCountEstimation()
// 使用词数 * 1.3 的经验公式估算（英文）
func (e *TokenEstimator) CountTokens(text string) int {
	words := strings.Fields(text)
	return int(float64(len(words)) * 1.3)
}

// CountMessagesTokens 计算多条消息的总令牌数
// 对应 TypeScript: roughTokenCountEstimationForMessages()
// 累加每条消息的令牌数
func (e *TokenEstimator) CountMessagesTokens(messages []string) int {
	total := 0
	for _, msg := range messages {
		total += e.CountTokens(msg)
	}
	return total
}

// EstimateCost 根据令牌使用量和模型计算API调用成本
// 对应 TypeScript: 计算美元成本的逻辑
// 公式: (输入令牌数/1M) * 输入价格 + (输出令牌数/1M) * 输出价格
func (e *TokenEstimator) EstimateCost(usage TokenUsage, model string) float64 {
	pricing, ok := modelPricing[model]
	if !ok {
		// 未知模型默认使用 claude-sonnet 价格
		pricing = ModelPricing{InputCostPerMillion: 3.0, OutputCostPerMillion: 15.0}
	}

	inputCost := float64(usage.InputTokens) / 1_000_000 * pricing.InputCostPerMillion
	outputCost := float64(usage.OutputTokens) / 1_000_000 * pricing.OutputCostPerMillion

	return inputCost + outputCost
}

// GetModelPricing 获取指定模型的定价信息
// 对应 TypeScript: 模型定价查询
// 如果模型不在定价表中，返回默认值（claude-sonnet价格）
func (e *TokenEstimator) GetModelPricing(model string) ModelPricing {
	if pricing, ok := modelPricing[model]; ok {
		return pricing
	}
	return ModelPricing{InputCostPerMillion: 3.0, OutputCostPerMillion: 15.0}
}

// Message 消息结构（用于令牌估算）
// 包含角色和内容的基本消息格式
type Message struct {
	Role    string `json:"role"`    // 角色: user, assistant, system
	Content string `json:"content"` // 消息内容
}

// TokenEstimateResult 令牌估算结果
// 返回估算的总令牌数和预计成本
type TokenEstimateResult struct {
	Messages    []Message `json:"messages"`    // 输入的消息列表
	TotalTokens int       `json:"totalTokens"` // 估算的总令牌数
	EstimateUSD float64   `json:"estimateUSD"` // 估算的美元成本
}

// EstimateMessages 估算多条消息的总令牌数和成本
// 对应 TypeScript: roughTokenCountEstimationForMessages()
// 累加所有消息的令牌数并计算总成本
func (e *TokenEstimator) EstimateMessages(messages []Message, model string) *TokenEstimateResult {
	total := 0
	for _, msg := range messages {
		total += e.CountTokens(msg.Content)
	}

	usage := TokenUsage{
		TotalTokens: total,
	}
	cost := e.EstimateCost(usage, model)

	return &TokenEstimateResult{
		Messages:    messages,
		TotalTokens: total,
		EstimateUSD: cost,
	}
}

// MarshalJSON 将TokenEstimator序列化为JSON
// 添加类型标识便于调试和日志记录
func (e *TokenEstimator) MarshalJSON() ([]byte, error) {
	type Alias TokenEstimator
	return json.Marshal(&struct {
		Type string `json:"type"`
		*Alias
	}{
		Type:  "TokenEstimator",
		Alias: (*Alias)(e),
	})
}
