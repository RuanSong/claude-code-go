package api

import (
	"sync"
)

// GroveConfig Grove通知功能配置
type GroveConfig struct {
	GroveEnabled            bool `json:"grove_enabled"`
	DomainExcluded          bool `json:"domain_excluded"`
	NoticeIsGracePeriod     bool `json:"notice_is_grace_period"`
	NoticeReminderFrequency *int `json:"notice_reminder_frequency,omitempty"`
}

// AccountSettings 账户设置
type AccountSettings struct {
	GroveEnabled        bool   `json:"grove_enabled"`
	GroveNoticeViewedAt string `json:"grove_notice_viewed_at,omitempty"`
}

// GroveCacheExpirationMS Grove缓存过期时间: 24小时
const GroveCacheExpirationMS = 24 * 60 * 60 * 1000

var (
	groveSettings     *AccountSettings
	groveSettingsErr  error
	groveSettingsOnce sync.Once
)

// ApiResult API结果类型
type ApiResult struct {
	Success bool
	Data    interface{}
	Error   string
}

// GetGroveSettings 获取Grove设置
// 使用memoize避免重复请求
func GetGroveSettings() (*AccountSettings, error) {
	// essential-traffic-only模式下跳过
	if isEssentialTrafficOnly() {
		return nil, nil
	}

	groveSettingsOnce.Do(func() {
		// 实际实现需要调用API
		// 简化版本返回空设置
		groveSettings = &AccountSettings{
			GroveEnabled:        false,
			GroveNoticeViewedAt: "",
		}
		groveSettingsErr = nil
	})

	return groveSettings, groveSettingsErr
}

// IsGroveEnabled 检查Grove是否启用
func IsGroveEnabled() bool {
	settings, err := GetGroveSettings()
	if err != nil || settings == nil {
		return false
	}
	return settings.GroveEnabled
}

// UpdateGroveSettings 更新Grove设置
func UpdateGroveSettings(settings *AccountSettings) error {
	// 更新缓存
	groveSettingsOnce.Do(func() {
		groveSettings = settings
		groveSettingsErr = nil
	})

	return nil
}

// GetGroveConfig 获取完整Grove配置
func GetGroveConfig() *GroveConfig {
	enabled := IsGroveEnabled()
	return &GroveConfig{
		GroveEnabled:            enabled,
		DomainExcluded:          false,
		NoticeIsGracePeriod:     false,
		NoticeReminderFrequency: nil,
	}
}

// ClearGroveCache 清除Grove缓存
func ClearGroveCache() {
	groveSettingsOnce = sync.Once{}
	groveSettings = nil
	groveSettingsErr = nil
}
