package services

import (
	"net/http"
	"time"

	mocklimits "github.com/claude-code-go/claude/internal/services/mockRateLimits"
)

// RateLimitMocking 速率限制模拟门面
// 将模拟逻辑与生产代码隔离

// ProcessRateLimitHeaders 处理响应头，应用模拟数据（如果/mocks-limits命令处于激活状态）
func ProcessRateLimitHeaders(headers http.Header) http.Header {
	if mocklimits.ShouldProcessMockLimits() {
		return applyMockHeadersInternal(headers)
	}
	return headers
}

// ShouldProcessRateLimits 检查是否应该处理速率限制（真实订阅者或/mocks-limits命令）
func ShouldProcessRateLimits(isSubscriber bool) bool {
	return isSubscriber || mocklimits.ShouldProcessMockLimits()
}

// applyMockHeadersInternal 将模拟头应用到现有头
func applyMockHeadersInternal(headers http.Header) http.Header {
	mockHeaders := mocklimits.GetMockHeaders()
	if mockHeaders == nil {
		return headers
	}

	// 创建新的Headers副本
	result := make(http.Header)
	for k, v := range headers {
		result[k] = v
	}

	// 应用模拟头
	for k, v := range mockHeaders {
		result.Set(k, v)
	}

	return result
}

// CheckMockRateLimitError 检查模拟速率限制是否应该抛出429错误
// 返回要抛出的错误，如果不应抛出则返回nil
// currentModel: 当前请求使用的模型
// isFastModeActive: 快速模式是否处于激活状态
func CheckMockRateLimitError(currentModel string, isFastModeActive bool) *RateLimitError {
	if !mocklimits.ShouldProcessMockLimits() {
		return nil
	}

	headerlessMessage := mocklimits.GetMockHeaderless429Message()
	if headerlessMessage != "" {
		return NewRateLimitError(429, headerlessMessage)
	}

	mockHeaders := mocklimits.GetMockHeaders()
	if mockHeaders == nil {
		return nil
	}

	// 检查是否应该抛出429错误
	status := mockHeaders["anthropic-ratelimit-unified-status"]
	overageStatus := mockHeaders["anthropic-ratelimit-unified-overage-status"]
	rateLimitType := mockHeaders["anthropic-ratelimit-unified-representative-claim"]

	// 检查是否是Opus特定限制
	isOpusLimit := rateLimitType == "seven_day_opus"

	// 检查当前模型是否是Opus模型
	isUsingOpus := containsOpus(currentModel)

	// 对于Opus限制，仅在实际使用Opus时才抛出429
	if isOpusLimit && !isUsingOpus {
		return nil
	}

	// 检查快速模式限制
	if mocklimits.IsMockFastModeRateLimitScenario() {
		headers := checkMockFastModeRateLimitInternal(isFastModeActive)
		if headers == nil {
			return nil
		}
		return NewRateLimitErrorWithHeaders(429, "Rate limit exceeded", headers)
	}

	shouldThrow429 := status == "rejected" && (overageStatus == "" || overageStatus == "rejected")

	if shouldThrow429 {
		return NewRateLimitErrorWithHeaders(429, "Rate limit exceeded", mockHeaders)
	}

	return nil
}

// IsMockRateLimitError 检查是否是模拟429错误（不应重试）
func IsMockRateLimitError(statusCode int) bool {
	return mocklimits.ShouldProcessMockLimits() && statusCode == 429
}

// RateLimitError 速率限制错误
type RateLimitError struct {
	StatusCode int
	Message    string
	Headers    map[string]string
}

// NewRateLimitError 创建速率限制错误
func NewRateLimitError(statusCode int, message string) *RateLimitError {
	return &RateLimitError{
		StatusCode: statusCode,
		Message:    message,
		Headers:    make(map[string]string),
	}
}

// NewRateLimitErrorWithHeaders 创建带响应头的速率限制错误
func NewRateLimitErrorWithHeaders(statusCode int, message string, headers map[string]string) *RateLimitError {
	return &RateLimitError{
		StatusCode: statusCode,
		Message:    message,
		Headers:    headers,
	}
}

// checkMockFastModeRateLimitInternal 检查快速模式速率限制
func checkMockFastModeRateLimitInternal(isFastModeActive bool) map[string]string {
	if !mocklimits.IsMockFastModeRateLimitScenario() {
		return nil
	}

	// 仅在快速模式激活时抛出
	if !isFastModeActive {
		return nil
	}

	// 获取mock headers进行进一步处理
	mockHeaders := mocklimits.GetMockHeaders()
	if mockHeaders == nil {
		return nil
	}

	// 这里可以添加更多逻辑，但现在只返回mock headers
	return mockHeaders
}

// 辅助函数

func containsOpus(model string) bool {
	return containsIgnoreCase(model, "opus")
}

func containsIgnoreCase(s, substr string) bool {
	sLower := toLower(s)
	substrLower := toLower(substr)
	return len(sLower) >= len(substrLower) && findSubstring(sLower, substrLower)
}

func toLower(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		result[i] = c
	}
	return string(result)
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// timeNow 获取当前时间（可mock）
var timeNow = func() time.Time { return time.Now() }
