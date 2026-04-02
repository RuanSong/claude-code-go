package api

import (
	"os"
	"time"
)

// RateLimit 速率限制信息
type RateLimit struct {
	Utilization *float64 // 使用率 0-100
	ResetsAt    *string  // 重置时间 ISO 8601格式
}

// ExtraUsage 额外使用量
type ExtraUsage struct {
	IsEnabled    bool     `json:"is_enabled"`
	MonthlyLimit *int64   `json:"monthly_limit,omitempty"`
	UsedCredits  *int64   `json:"used_credits,omitempty"`
	Utilization  *float64 `json:"utilization,omitempty"`
}

// Utilization 使用量统计
type Utilization struct {
	FiveHour          *RateLimit  `json:"five_hour,omitempty"`
	SevenDay          *RateLimit  `json:"seven_day,omitempty"`
	SevenDayOauthApps *RateLimit  `json:"seven_day_oauth_apps,omitempty"`
	SevenDayOpus      *RateLimit  `json:"seven_day_opus,omitempty"`
	SevenDaySonnet    *RateLimit  `json:"seven_day_sonnet,omitempty"`
	ExtraUsage        *ExtraUsage `json:"extra_usage,omitempty"`
}

// FetchUtilization 获取使用量统计
// 从API获取用户的Claude AI使用量信息
func FetchUtilization() (*Utilization, error) {
	// 非订阅用户返回空
	if !isClaudeAISubscriber() {
		return &Utilization{}, nil
	}

	// OAuth令牌过期检查
	tokens := GetClaudeAIOAuthTokens()
	if tokens != nil && isOAuthTokenExpired(tokens.ExpiresAt) {
		return nil, nil
	}

	headers, err := GetAuthHeaders()
	if err != nil {
		return nil, err
	}

	headers["Content-Type"] = "application/json"
	headers["User-Agent"] = GetUserAgent()

	url := GetOAuthConfig().BaseAPIURL + "/api/oauth/usage"

	body, statusCode, err := HTTPGet(url, headers, 5*time.Second)
	if err != nil {
		return nil, err
	}

	if !IsResponseSuccess(statusCode) {
		return nil, nil
	}

	var utilization Utilization
	if err := ParseJSON(body, &utilization); err != nil {
		return nil, err
	}

	return &utilization, nil
}

// GetUserAgent 获取User-Agent字符串
func GetUserAgent() string {
	return "claude-code-go/1.0"
}

// OAUTHTokens OAuth令牌
type OAUTHTokens struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
}

// GetClaudeAIOAuthTokens 获取Claude AI OAuth令牌
func GetClaudeAIOAuthTokens() *OAUTHTokens {
	accessToken := os.Getenv("CLAUDEAI_ACCESS_TOKEN")
	if accessToken == "" {
		return nil
	}

	return &OAUTHTokens{
		AccessToken: accessToken,
		ExpiresAt:   time.Now().Add(time.Hour), // 默认1小时后过期
	}
}

// isOAuthTokenExpired 检查OAuth令牌是否过期
func isOAuthTokenExpired(expiresAt time.Time) bool {
	return time.Now().After(expiresAt)
}

// HasProfileScope 检查是否有profile作用域
func HasProfileScope() bool {
	// 简化实现
	return isClaudeAISubscriber()
}
