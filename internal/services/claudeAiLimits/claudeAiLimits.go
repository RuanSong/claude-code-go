package claudeAiLimits

import (
	"sync"
	"time"
)

// QuotaStatus 配额状态
type QuotaStatus string

const (
	QuotaStatusAllowed        QuotaStatus = "allowed"         // 允许
	QuotaStatusAllowedWarning QuotaStatus = "allowed_warning" // 允许但警告
	QuotaStatusRejected       QuotaStatus = "rejected"        // 拒绝
)

// RateLimitType 速率限制类型
type RateLimitType string

const (
	RateLimitTypeFiveHour       RateLimitType = "five_hour"        // 5小时限制
	RateLimitTypeSevenDay       RateLimitType = "seven_day"        // 7天限制
	RateLimitTypeSevenDayOpus   RateLimitType = "seven_day_opus"   // Opus 7天限制
	RateLimitTypeSevenDaySonnet RateLimitType = "seven_day_sonnet" // Sonnet 7天限制
	RateLimitTypeOverage        RateLimitType = "overage"          // 超额限制
)

// OverageDisabledReason 超额禁用原因
type OverageDisabledReason string

const (
	OverageNotProvisioned     OverageDisabledReason = "overage_not_provisioned"       // 超额未配置
	OrgLevelDisabled          OverageDisabledReason = "org_level_disabled"            // 组织级别禁用
	OrgLevelDisabledUntil     OverageDisabledReason = "org_level_disabled_until"      // 组织级别暂时禁用
	OutOfCredits              OverageDisabledReason = "out_of_credits"                // 信用额度不足
	SeatTierLevelDisabled     OverageDisabledReason = "seat_tier_level_disabled"      // 座位层级别禁用
	MemberLevelDisabled       OverageDisabledReason = "member_level_disabled"         // 成员级别禁用
	SeatTierZeroCreditLimit   OverageDisabledReason = "seat_tier_zero_credit_limit"   // 座位层零信用限制
	GroupZeroCreditLimit      OverageDisabledReason = "group_zero_credit_limit"       // 组零信用限制
	MemberZeroCreditLimit     OverageDisabledReason = "member_zero_credit_limit"      // 成员零信用限制
	OrgServiceLevelDisabled   OverageDisabledReason = "org_service_level_disabled"    // 组织服务级别禁用
	OrgServiceZeroCreditLimit OverageDisabledReason = "org_service_zero_credit_limit" // 组织服务零信用限制
	NoLimitsConfigured        OverageDisabledReason = "no_limits_configured"          // 未配置限制
	Unknown                   OverageDisabledReason = "unknown"                       // 未知原因
)

// EarlyWarningThreshold 早期警告阈值
type EarlyWarningThreshold struct {
	Utilization float64 // 0-1范围: 当使用量 >= 此值时触发警告
	TimePct     float64 // 0-1范围: 当时间已过 <= 此比例时触发警告
}

// EarlyWarningConfig 早期警告配置
type EarlyWarningConfig struct {
	RateLimitType RateLimitType
	ClaimAbbrev   string
	WindowSeconds int
	Thresholds    []EarlyWarningThreshold
}

// EARLY_WARNING_CONFIGS 早期警告配置(按优先级排序)
var EARLY_WARNING_CONFIGS = []EarlyWarningConfig{
	{
		RateLimitType: RateLimitTypeFiveHour,
		ClaimAbbrev:   "5h",
		WindowSeconds: 5 * 60 * 60, // 5小时
		Thresholds: []EarlyWarningThreshold{
			{Utilization: 0.9, TimePct: 0.72},
		},
	},
	{
		RateLimitType: RateLimitTypeSevenDay,
		ClaimAbbrev:   "7d",
		WindowSeconds: 7 * 24 * 60 * 60, // 7天
		Thresholds: []EarlyWarningThreshold{
			{Utilization: 0.75, TimePct: 0.6},
			{Utilization: 0.5, TimePct: 0.35},
			{Utilization: 0.25, TimePct: 0.15},
		},
	},
}

// RATE_LIMIT_DISPLAY_NAMES 速率限制显示名称
var RATE_LIMIT_DISPLAY_NAMES = map[RateLimitType]string{
	RateLimitTypeFiveHour:       "session limit",
	RateLimitTypeSevenDay:       "weekly limit",
	RateLimitTypeSevenDayOpus:   "Opus limit",
	RateLimitTypeSevenDaySonnet: "Sonnet limit",
	RateLimitTypeOverage:        "extra usage limit",
}

// GetRateLimitDisplayName 获取速率限制显示名称
func GetRateLimitDisplayName(rateLimitType RateLimitType) string {
	if name, ok := RATE_LIMIT_DISPLAY_NAMES[rateLimitType]; ok {
		return name
	}
	return string(rateLimitType)
}

// RawWindowUtilization 原始窗口使用量
type RawWindowUtilization struct {
	Utilization float64 // 0-1 比例
	ResetsAt    int64   // unix时间戳(秒)
}

// RawUtilization 原始使用量
type RawUtilization struct {
	FiveHour *RawWindowUtilization
	SevenDay *RawWindowUtilization
}

// ClaudeAILimits Claude AI限制状态
type ClaudeAILimits struct {
	Status                            QuotaStatus           // 当前状态
	UnifiedRateLimitFallbackAvailable bool                  // 统一速率限制回退是否可用
	ResetsAt                          int64                 // 重置时间戳
	RateLimitType                     RateLimitType         // 速率限制类型
	Utilization                       float64               // 使用量(0-1)
	OverageStatus                     QuotaStatus           // 超额状态
	OverageResetsAt                   int64                 // 超额重置时间戳
	OverageDisabledReason             OverageDisabledReason // 超额禁用原因
	IsUsingOverage                    bool                  // 是否使用超额模式
	SurpassedThreshold                float64               // 超过的阈值
}

// StatusChangeListener 状态变更监听器
type StatusChangeListener func(limits *ClaudeAILimits)

// ClaudeAiLimitsService Claude AI限制服务
// 追踪Claude AI API的速率限制和使用量状态
type ClaudeAiLimitsService struct {
	mu             sync.RWMutex
	currentLimits  *ClaudeAILimits
	rawUtilization *RawUtilization
	listeners      []StatusChangeListener
}

var (
	instance *ClaudeAiLimitsService
	once     sync.Once
)

// GetInstance 获取单例实例
func GetInstance() *ClaudeAiLimitsService {
	once.Do(func() {
		instance = &ClaudeAiLimitsService{
			currentLimits: &ClaudeAILimits{
				Status:                            QuotaStatusAllowed,
				UnifiedRateLimitFallbackAvailable: false,
				IsUsingOverage:                    false,
			},
			rawUtilization: &RawUtilization{},
			listeners:      []StatusChangeListener{},
		}
	})
	return instance
}

// GetCurrentLimits 获取当前限制状态
func (s *ClaudeAiLimitsService) GetCurrentLimits() *ClaudeAILimits {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.currentLimits
}

// GetRawUtilization 获取原始使用量
func (s *ClaudeAiLimitsService) GetRawUtilization() *RawUtilization {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.rawUtilization
}

// AddListener 添加状态变更监听器
func (s *ClaudeAiLimitsService) AddListener(listener StatusChangeListener) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.listeners = append(s.listeners, listener)
}

// RemoveListener 移除状态变更监听器
func (s *ClaudeAiLimitsService) RemoveListener(listener StatusChangeListener) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, l := range s.listeners {
		if &l == &listener {
			s.listeners = append(s.listeners[:i], s.listeners[i+1:]...)
			break
		}
	}
}

// emitStatusChange 触发状态变更
func (s *ClaudeAiLimitsService) emitStatusChange(limits *ClaudeAILimits) {
	s.mu.Lock()
	s.currentLimits = limits
	listeners := make([]StatusChangeListener, len(s.listeners))
	copy(listeners, s.listeners)
	s.mu.Unlock()

	for _, listener := range listeners {
		listener(limits)
	}
}

// ExtractQuotaStatusFromHeaders 从响应头提取配额状态
func (s *ClaudeAiLimitsService) ExtractQuotaStatusFromHeaders(headers map[string]string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 处理headers (可能包含mock数据)
	headersToUse := processHeaders(headers)

	// 提取原始使用量
	s.extractRawUtilizationLocked(headersToUse)

	// 计算新限制
	newLimits := s.computeNewLimitsFromHeadersLocked(headersToUse)

	// 检查是否需要触发更新
	if !s.isLimitsEqual(s.currentLimits, newLimits) {
		s.emitStatusChange(newLimits)
	}
}

// processHeaders 处理headers(支持mock)
func processHeaders(headers map[string]string) map[string]string {
	// 在真实实现中,这会调用rateLimitMocking处理
	return headers
}

// extractRawUtilizationLocked 提取原始使用量(需要持有锁)
func (s *ClaudeAiLimitsService) extractRawUtilizationLocked(headers map[string]string) {
	rawUtil := &RawUtilization{}

	if util, ok := headers["anthropic-ratelimit-unified-5h-utilization"]; ok {
		if reset, ok := headers["anthropic-ratelimit-unified-5h-reset"]; ok {
			rawUtil.FiveHour = &RawWindowUtilization{
				Utilization: parseFloat(util),
				ResetsAt:    parseInt64(reset),
			}
		}
	}

	if util, ok := headers["anthropic-ratelimit-unified-7d-utilization"]; ok {
		if reset, ok := headers["anthropic-ratelimit-unified-7d-reset"]; ok {
			rawUtil.SevenDay = &RawWindowUtilization{
				Utilization: parseFloat(util),
				ResetsAt:    parseInt64(reset),
			}
		}
	}

	s.rawUtilization = rawUtil
}

// computeNewLimitsFromHeadersLocked 从headers计算新限制(需要持有锁)
func (s *ClaudeAiLimitsService) computeNewLimitsFromHeadersLocked(headers map[string]string) *ClaudeAILimits {
	limits := &ClaudeAILimits{}

	// 解析状态
	if status, ok := headers["anthropic-ratelimit-unified-status"]; ok {
		limits.Status = QuotaStatus(status)
	} else {
		limits.Status = QuotaStatusAllowed
	}

	// 解析重置时间
	if reset, ok := headers["anthropic-ratelimit-unified-reset"]; ok {
		limits.ResetsAt = parseInt64(reset)
	}

	// 解析统一速率限制回退可用性
	limits.UnifiedRateLimitFallbackAvailable = headers["anthropic-ratelimit-unified-fallback"] == "available"

	// 解析速率限制类型
	if rateLimitType, ok := headers["anthropic-ratelimit-unified-representative-claim"]; ok {
		limits.RateLimitType = RateLimitType(rateLimitType)
	}

	// 解析超额状态
	if overageStatus, ok := headers["anthropic-ratelimit-unified-overage-status"]; ok {
		limits.OverageStatus = QuotaStatus(overageStatus)
	}

	// 解析超额重置时间
	if overageReset, ok := headers["anthropic-ratelimit-unified-overage-reset"]; ok {
		limits.OverageResetsAt = parseInt64(overageReset)
	}

	// 解析超额禁用原因
	if reason, ok := headers["anthropic-ratelimit-unified-overage-disabled-reason"]; ok {
		limits.OverageDisabledReason = OverageDisabledReason(reason)
	}

	// 确定是否使用超额模式
	limits.IsUsingOverage = limits.Status == QuotaStatusRejected &&
		(limits.OverageStatus == QuotaStatusAllowed || limits.OverageStatus == QuotaStatusAllowedWarning)

	// 检查早期警告
	if limits.Status == QuotaStatusAllowed || limits.Status == QuotaStatusAllowedWarning {
		if earlyWarning := s.getEarlyWarningFromHeadersLocked(headers, limits.UnifiedRateLimitFallbackAvailable); earlyWarning != nil {
			return earlyWarning
		}
		limits.Status = QuotaStatusAllowed
	}

	return limits
}

// getHeaderBasedEarlyWarning 获取基于header的早期警告
func (s *ClaudeAiLimitsService) getHeaderBasedEarlyWarning(
	headers map[string]string,
	unifiedRateLimitFallbackAvailable bool,
) *ClaudeAILimits {
	claimMap := map[string]RateLimitType{
		"5h":      RateLimitTypeFiveHour,
		"7d":      RateLimitTypeSevenDay,
		"overage": RateLimitTypeOverage,
	}

	for abbrev, rateLimitType := range claimMap {
		surpassedKey := "anthropic-ratelimit-unified-" + abbrev + "-surpassed-threshold"
		if surpassed, ok := headers[surpassedKey]; ok && surpassed != "" {
			utilizationKey := "anthropic-ratelimit-unified-" + abbrev + "-utilization"
			resetKey := "anthropic-ratelimit-unified-" + abbrev + "-reset"

			var utilization float64
			var resetsAt int64

			if util, ok := headers[utilizationKey]; ok {
				utilization = parseFloat(util)
			}
			if reset, ok := headers[resetKey]; ok {
				resetsAt = parseInt64(reset)
			}

			return &ClaudeAILimits{
				Status:                            QuotaStatusAllowedWarning,
				ResetsAt:                          resetsAt,
				RateLimitType:                     rateLimitType,
				Utilization:                       utilization,
				UnifiedRateLimitFallbackAvailable: unifiedRateLimitFallbackAvailable,
				IsUsingOverage:                    false,
				SurpassedThreshold:                parseFloat(surpassed),
			}
		}
	}

	return nil
}

// getTimeRelativeEarlyWarning 获取时间相关的早期警告
func (s *ClaudeAiLimitsService) getTimeRelativeEarlyWarning(
	headers map[string]string,
	config EarlyWarningConfig,
	unifiedRateLimitFallbackAvailable bool,
) *ClaudeAILimits {
	utilizationKey := "anthropic-ratelimit-unified-" + config.ClaimAbbrev + "-utilization"
	resetKey := "anthropic-ratelimit-unified-" + config.ClaimAbbrev + "-reset"

	utilStr, hasUtil := headers[utilizationKey]
	resetStr, hasReset := headers[resetKey]

	if !hasUtil || !hasReset {
		return nil
	}

	utilization := parseFloat(utilStr)
	resetsAt := parseInt64(resetStr)
	timeProgress := computeTimeProgress(resetsAt, config.WindowSeconds)

	// 检查是否超过阈值
	for _, threshold := range config.Thresholds {
		if utilization >= threshold.Utilization && timeProgress <= threshold.TimePct {
			return &ClaudeAILimits{
				Status:                            QuotaStatusAllowedWarning,
				ResetsAt:                          resetsAt,
				RateLimitType:                     config.RateLimitType,
				Utilization:                       utilization,
				UnifiedRateLimitFallbackAvailable: unifiedRateLimitFallbackAvailable,
				IsUsingOverage:                    false,
			}
		}
	}

	return nil
}

// getEarlyWarningFromHeadersLocked 获取早期警告(需要持有锁)
func (s *ClaudeAiLimitsService) getEarlyWarningFromHeadersLocked(
	headers map[string]string,
	unifiedRateLimitFallbackAvailable bool,
) *ClaudeAILimits {
	// 首先尝试基于header的检测
	if warning := s.getHeaderBasedEarlyWarning(headers, unifiedRateLimitFallbackAvailable); warning != nil {
		return warning
	}

	// 回退到时间相对阈值
	for _, config := range EARLY_WARNING_CONFIGS {
		if warning := s.getTimeRelativeEarlyWarning(headers, config, unifiedRateLimitFallbackAvailable); warning != nil {
			return warning
		}
	}

	return nil
}

// computeTimeProgress 计算时间进度
func computeTimeProgress(resetsAt int64, windowSeconds int) float64 {
	nowSeconds := float64(time.Now().Unix())
	windowStart := float64(resetsAt) - float64(windowSeconds)
	elapsed := nowSeconds - windowStart
	progress := elapsed / float64(windowSeconds)
	if progress < 0 {
		return 0
	}
	if progress > 1 {
		return 1
	}
	return progress
}

// isLimitsEqual 比较两个限制是否相等
func (s *ClaudeAiLimitsService) isLimitsEqual(a, b *ClaudeAILimits) bool {
	if a == nil || b == nil {
		return a == b
	}
	return a.Status == b.Status &&
		a.UnifiedRateLimitFallbackAvailable == b.UnifiedRateLimitFallbackAvailable &&
		a.ResetsAt == b.ResetsAt &&
		a.RateLimitType == b.RateLimitType &&
		a.Utilization == b.Utilization &&
		a.OverageStatus == b.OverageStatus &&
		a.OverageResetsAt == b.OverageResetsAt &&
		a.OverageDisabledReason == b.OverageDisabledReason &&
		a.IsUsingOverage == b.IsUsingOverage
}

// Reset 重置限制状态
func (s *ClaudeAiLimitsService) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.currentLimits = &ClaudeAILimits{
		Status:                            QuotaStatusAllowed,
		UnifiedRateLimitFallbackAvailable: false,
		IsUsingOverage:                    false,
	}
	s.rawUtilization = &RawUtilization{}
}

// Helper functions

func parseFloat(s string) float64 {
	var f float64
	for _, c := range s {
		if c >= '0' && c <= '9' {
			f = f*10 + float64(c-'0')
		} else if c == '.' {
			break
		}
	}
	return f
}

func parseInt64(s string) int64 {
	var v int64
	negative := false
	for _, c := range s {
		if c == '-' {
			negative = true
		} else if c >= '0' && c <= '9' {
			v = v*10 + int64(c-'0')
		}
	}
	if negative {
		v = -v
	}
	return v
}
