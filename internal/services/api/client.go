package api

import (
	"encoding/json"
	"net/http"
	"os"
	"time"
)

// UltrareviewQuotaResponse UltraReview配额响应
type UltrareviewQuotaResponse struct {
	ReviewsUsed      int  `json:"reviews_used"`
	ReviewsLimit     int  `json:"reviews_limit"`
	ReviewsRemaining int  `json:"reviews_remaining"`
	IsOverage        bool `json:"is_overage"`
}

// fetchUltrareviewQuota 获取UltraReview配额
// 用于显示和决定是否提示用户。配额消耗在服务端进行。
func FetchUltrareviewQuota(accessToken, orgUUID string) (*UltrareviewQuotaResponse, error) {
	// 检查是否为订阅用户
	if !isClaudeAISubscriber() {
		return nil, nil
	}

	// 注意：这里需要OAuth配置和API调用
	// 简化实现，仅返回nil
	return nil, nil
}

// isClaudeAISubscriber 检查是否为Claude AI订阅用户
func isClaudeAISubscriber() bool {
	// 从环境变量或配置获取订阅状态
	return os.Getenv("CLAUDEAI_SUBSCRIBER") == "true"
}

// GetOAuthConfig 获取OAuth配置
func GetOAuthConfig() OAuthConfig {
	return OAuthConfig{
		BaseAPIURL: "https://api.anthropic.com",
	}
}

// OAuthConfig OAuth配置
type OAuthConfig struct {
	BaseAPIURL string
}

// APIResponse API通用响应
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// ErrorResponse 错误响应
type ErrorResponse struct {
	Error struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error"`
}

// GetAuthHeaders 获取认证头
func GetAuthHeaders() (map[string]string, error) {
	headers := make(map[string]string)

	// API Key认证
	if apiKey := os.Getenv("ANTHROPIC_API_KEY"); apiKey != "" {
		headers["x-api-key"] = apiKey
		return headers, nil
	}

	return nil, nil
}

// IsResponseSuccess 检查响应是否成功
func IsResponseSuccess(statusCode int) bool {
	return statusCode >= 200 && statusCode < 300
}

// ParseJSON 解析JSON响应
func ParseJSON(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

// MarshalJSON 序列化为JSON
func MarshalJSON(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// HTTPGet 执行GET请求
func HTTPGet(url string, headers map[string]string, timeout time.Duration) ([]byte, int, error) {
	client := &http.Client{Timeout: timeout}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, 0, err
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	body := make([]byte, 0, resp.ContentLength)
	buf := make([]byte, 1024)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			body = append(body, buf[:n]...)
		}
		if err != nil {
			break
		}
	}

	return body, resp.StatusCode, nil
}
