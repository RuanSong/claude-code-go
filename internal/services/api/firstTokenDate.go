package api

import (
	"os"
	"time"
)

// FirstTokenDate 首次使用令牌日期
type FirstTokenDate struct {
	FirstTokenDate string `json:"first_token_date"`
}

// FetchAndStoreClaudeCodeFirstTokenDate 获取并存储用户首次使用Claude Code的日期
// 在登录成功后调用，用于缓存用户开始使用Claude Code的时间
func FetchAndStoreClaudeCodeFirstTokenDate() error {
	config := GetGlobalConfig()

	// 如果已缓存，直接返回
	if config.ClaudeCodeFirstTokenDate != "" {
		return nil
	}

	authHeaders, err := GetAuthHeaders()
	if err != nil {
		return err
	}

	oauthConfig := GetOAuthConfig()
	url := oauthConfig.BaseAPIURL + "/api/organization/claude_code_first_token_date"

	headers := authHeaders
	headers["User-Agent"] = GetUserAgent()

	body, _, err := HTTPGet(url, headers, 10*time.Second)
	if err != nil {
		return err
	}

	var response FirstTokenDate
	if err := ParseJSON(body, &response); err != nil {
		return err
	}

	firstTokenDate := response.FirstTokenDate

	// 验证日期有效性
	if firstTokenDate != "" {
		dateTime, err := time.Parse(time.RFC3339, firstTokenDate)
		if err != nil {
			return err
		}
		if dateTime.IsZero() {
			// 无效日期，不保存
			return nil
		}
	}

	// 保存到配置
	SaveGlobalConfig(func(current *GlobalConfig) *GlobalConfig {
		current.ClaudeCodeFirstTokenDate = firstTokenDate
		return current
	})

	return nil
}

// GlobalConfig 全局配置
type GlobalConfig struct {
	ClaudeCodeFirstTokenDate string              `json:"claudeCodeFirstTokenDate,omitempty"`
	MetricsStatusCache       *MetricsStatusCache `json:"metricsStatusCache,omitempty"`
}

// MetricsStatusCache 指标状态缓存
type MetricsStatusCache struct {
	Enabled   bool  `json:"enabled"`
	Timestamp int64 `json:"timestamp"`
}

// GetGlobalConfig 获取全局配置
func GetGlobalConfig() *GlobalConfig {
	// 从环境变量或配置文件读取
	config := &GlobalConfig{}

	if claudeCodeFirstTokenDate := os.Getenv("CLAUDE_CODE_FIRST_TOKEN_DATE"); claudeCodeFirstTokenDate != "" {
		config.ClaudeCodeFirstTokenDate = claudeCodeFirstTokenDate
	}

	return config
}

// SaveGlobalConfig 保存全局配置
func SaveGlobalConfig(update func(*GlobalConfig) *GlobalConfig) {
	// 实现配置保存
	config := GetGlobalConfig()
	config = update(config)

	// 保存到环境变量或文件
	_ = config
}
