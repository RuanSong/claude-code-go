package services

import (
	"time"
)

// MockRateLimitScenario 模拟速率限制场景
type MockRateLimitScenario string

const (
	MockScenarioNormal                  MockRateLimitScenario = "normal"
	MockScenarioSessionLimitReached     MockRateLimitScenario = "session-limit-reached"
	MockScenarioApproachingWeeklyLimit  MockRateLimitScenario = "approaching-weekly-limit"
	MockScenarioWeeklyLimitReached      MockRateLimitScenario = "weekly-limit-reached"
	MockScenarioOverageActive           MockRateLimitScenario = "overage-active"
	MockScenarioOverageWarning          MockRateLimitScenario = "overage-warning"
	MockScenarioOverageExhausted        MockRateLimitScenario = "overage-exhausted"
	MockScenarioOutOfCredits            MockRateLimitScenario = "out-of-credits"
	MockScenarioOrgZeroCreditLimit      MockRateLimitScenario = "org-zero-credit-limit"
	MockScenarioOrgSpendCapHit          MockRateLimitScenario = "org-spend-cap-hit"
	MockScenarioMemberZeroCreditLimit   MockRateLimitScenario = "member-zero-credit-limit"
	MockScenarioSeatTierZeroCreditLimit MockRateLimitScenario = "seat-tier-zero-credit-limit"
	MockScenarioOpusLimit               MockRateLimitScenario = "opus-limit"
	MockScenarioOpusWarning             MockRateLimitScenario = "opus-warning"
	MockScenarioSonnetLimit             MockRateLimitScenario = "sonnet-limit"
	MockScenarioSonnetWarning           MockRateLimitScenario = "sonnet-warning"
	MockScenarioFastModeLimit           MockRateLimitScenario = "fast-mode-limit"
	MockScenarioFastModeShortLimit      MockRateLimitScenario = "fast-mode-short-limit"
	MockScenarioExtraUsageRequired      MockRateLimitScenario = "extra-usage-required"
	MockScenarioClear                   MockRateLimitScenario = "clear"
)

// MockHeaderKey 模拟响应头键名
type MockHeaderKey string

const (
	MockHeaderStatus                MockHeaderKey = "status"
	MockHeaderReset                 MockHeaderKey = "reset"
	MockHeaderClaim                 MockHeaderKey = "claim"
	MockHeaderOverageStatus         MockHeaderKey = "overage-status"
	MockHeaderOverageReset          MockHeaderKey = "overage-reset"
	MockHeaderOverageDisabledReason MockHeaderKey = "overage-disabled-reason"
	MockHeaderFallback              MockHeaderKey = "fallback"
	MockHeaderFallbackPercentage    MockHeaderKey = "fallback-percentage"
	MockHeaderRetryAfter            MockHeaderKey = "retry-after"
	MockHeader5hUtilization         MockHeaderKey = "5h-utilization"
	MockHeader5hReset               MockHeaderKey = "5h-reset"
	MockHeader5hSurpassedThreshold  MockHeaderKey = "5h-surpassed-threshold"
	MockHeader7dUtilization         MockHeaderKey = "7d-utilization"
	MockHeader7dReset               MockHeaderKey = "7d-reset"
	MockHeader7dSurpassedThreshold  MockHeaderKey = "7d-surpassed-threshold"
)

// MockHeaders 模拟速率限制响应头
type MockHeaders map[string]string

// ExceededLimit 超额限制类型
type ExceededLimitType string

const (
	ExceededLimitFiveHour       ExceededLimitType = "five_hour"
	ExceededLimitSevenDay       ExceededLimitType = "seven_day"
	ExceededLimitSevenDayOpus   ExceededLimitType = "seven_day_opus"
	ExceededLimitSevenDaySonnet ExceededLimitType = "seven_day_sonnet"
)

// ExceededLimit 超额限制记录
type ExceededLimit struct {
	Type     ExceededLimitType
	ResetsAt int64 // Unix时间戳
}

// MockRateLimiter 速率限制模拟器
type MockRateLimiter struct {
	headers                     MockHeaders
	enabled                     bool
	headerless429Message        string
	exceededLimits              []ExceededLimit
	fastModeRateLimitDurationMs int64
	fastModeRateLimitExpiresAt  int64
}

// 全局模拟速率限制器实例
var mockRateLimiter = &MockRateLimiter{
	headers:        make(MockHeaders),
	exceededLimits: make([]ExceededLimit, 0),
}

// isMockEnabled 检查是否启用了模拟
func (m *MockRateLimiter) isMockEnabled() bool {
	return m.enabled
}

// SetMockHeader 设置模拟响应头
func SetMockHeader(key MockHeaderKey, value string) {
	mockRateLimiter.enabled = true

	// 处理清除
	if value == "" || value == "clear" {
		delete(mockRateLimiter.headers, string(key))
		if key == MockHeaderClaim {
			mockRateLimiter.exceededLimits = nil
		}
		return
	}

	// 处理重置时间
	if key == MockHeaderReset || key == MockHeaderOverageReset {
		if hours := parseHours(value); hours > 0 {
			resetTime := time.Now().Unix() + int64(hours*3600)
			mockRateLimiter.headers[string(key)] = formatTimestamp(resetTime)
			return
		}
	}

	// 处理claims
	if key == MockHeaderClaim {
		mockRateLimiter.handleClaim(value)
		return
	}

	mockRateLimiter.headers[string(key)] = value
}

// handleClaim 处理限制类型声明
func (m *MockRateLimiter) handleClaim(claim string) {
	validClaims := []string{
		string(ExceededLimitFiveHour),
		string(ExceededLimitSevenDay),
		string(ExceededLimitSevenDayOpus),
		string(ExceededLimitSevenDaySonnet),
	}

	if !contains(validClaims, claim) {
		return
	}

	var resetsAt int64
	switch claim {
	case string(ExceededLimitFiveHour):
		resetsAt = time.Now().Unix() + 5*3600
	case string(ExceededLimitSevenDay), string(ExceededLimitSevenDayOpus), string(ExceededLimitSevenDaySonnet):
		resetsAt = time.Now().Unix() + 7*24*3600
	default:
		resetsAt = time.Now().Unix() + 3600
	}

	// 移除已存在的同类限制
	newLimits := make([]ExceededLimit, 0)
	for _, l := range m.exceededLimits {
		if l.Type != ExceededLimitType(claim) {
			newLimits = append(newLimits, l)
		}
	}
	m.exceededLimits = append(newLimits, ExceededLimit{
		Type:     ExceededLimitType(claim),
		ResetsAt: resetsAt,
	})

	// 更新代表性声明
	m.updateRepresentativeClaim()
}

// updateRepresentativeClaim 更新代表性声明
func (m *MockRateLimiter) updateRepresentativeClaim() {
	if len(m.exceededLimits) == 0 {
		delete(m.headers, "anthropic-ratelimit-unified-representative-claim")
		delete(m.headers, "anthropic-ratelimit-unified-reset")
		delete(m.headers, "retry-after")
		return
	}

	// 找到最远的重置时间
	furthest := m.exceededLimits[0]
	for _, limit := range m.exceededLimits {
		if limit.ResetsAt > furthest.ResetsAt {
			furthest = limit
		}
	}

	m.headers["anthropic-ratelimit-unified-representative-claim"] = string(furthest.Type)
	m.headers["anthropic-ratelimit-unified-reset"] = formatTimestamp(furthest.ResetsAt)

	// 如果状态为拒绝且没有overage可用，添加retry-after
	if m.headers["anthropic-ratelimit-unified-status"] == "rejected" {
		overageStatus := m.headers["anthropic-ratelimit-unified-overage-status"]
		if overageStatus == "" || overageStatus == "rejected" {
			remaining := furthest.ResetsAt - time.Now().Unix()
			if remaining < 0 {
				remaining = 0
			}
			m.headers["retry-after"] = formatInt(remaining)
		}
	}
}

// SetMockRateLimitScenario 设置模拟速率限制场景
func SetMockRateLimitScenario(scenario MockRateLimitScenario) {
	if scenario == MockScenarioClear {
		mockRateLimiter.headers = make(MockHeaders)
		mockRateLimiter.headerless429Message = ""
		mockRateLimiter.enabled = false
		mockRateLimiter.exceededLimits = nil
		mockRateLimiter.fastModeRateLimitDurationMs = 0
		mockRateLimiter.fastModeRateLimitExpiresAt = 0
		return
	}

	mockRateLimiter.enabled = true

	fiveHoursFromNow := time.Now().Unix() + 5*3600
	sevenDaysFromNow := time.Now().Unix() + 7*24*3600

	// 保留overage场景的exceeded limits
	preserveExceeded := scenario == MockScenarioOverageActive ||
		scenario == MockScenarioOverageWarning ||
		scenario == MockScenarioOverageExhausted

	if !preserveExceeded {
		mockRateLimiter.exceededLimits = nil
	}

	mockRateLimiter.headers = make(MockHeaders)

	switch scenario {
	case MockScenarioNormal:
		mockRateLimiter.headers["anthropic-ratelimit-unified-status"] = "allowed"
		mockRateLimiter.headers["anthropic-ratelimit-unified-reset"] = formatTimestamp(fiveHoursFromNow)

	case MockScenarioSessionLimitReached:
		mockRateLimiter.exceededLimits = append(mockRateLimiter.exceededLimits, ExceededLimit{
			Type:     ExceededLimitFiveHour,
			ResetsAt: fiveHoursFromNow,
		})
		mockRateLimiter.updateRepresentativeClaim()
		mockRateLimiter.headers["anthropic-ratelimit-unified-status"] = "rejected"

	case MockScenarioApproachingWeeklyLimit:
		mockRateLimiter.headers["anthropic-ratelimit-unified-status"] = "allowed_warning"
		mockRateLimiter.headers["anthropic-ratelimit-unified-reset"] = formatTimestamp(sevenDaysFromNow)
		mockRateLimiter.headers["anthropic-ratelimit-unified-representative-claim"] = "seven_day"

	case MockScenarioWeeklyLimitReached:
		mockRateLimiter.exceededLimits = append(mockRateLimiter.exceededLimits, ExceededLimit{
			Type:     ExceededLimitSevenDay,
			ResetsAt: sevenDaysFromNow,
		})
		mockRateLimiter.updateRepresentativeClaim()
		mockRateLimiter.headers["anthropic-ratelimit-unified-status"] = "rejected"

	case MockScenarioOverageActive:
		if len(mockRateLimiter.exceededLimits) == 0 {
			mockRateLimiter.exceededLimits = append(mockRateLimiter.exceededLimits, ExceededLimit{
				Type:     ExceededLimitFiveHour,
				ResetsAt: fiveHoursFromNow,
			})
		}
		mockRateLimiter.updateRepresentativeClaim()
		mockRateLimiter.headers["anthropic-ratelimit-unified-status"] = "rejected"
		mockRateLimiter.headers["anthropic-ratelimit-unified-overage-status"] = "allowed"
		// 月末重置
		endOfMonth := getEndOfMonthTime()
		mockRateLimiter.headers["anthropic-ratelimit-unified-overage-reset"] = formatTimestamp(endOfMonth)

	case MockScenarioOverageWarning:
		if len(mockRateLimiter.exceededLimits) == 0 {
			mockRateLimiter.exceededLimits = append(mockRateLimiter.exceededLimits, ExceededLimit{
				Type:     ExceededLimitFiveHour,
				ResetsAt: fiveHoursFromNow,
			})
		}
		mockRateLimiter.updateRepresentativeClaim()
		mockRateLimiter.headers["anthropic-ratelimit-unified-status"] = "rejected"
		mockRateLimiter.headers["anthropic-ratelimit-unified-overage-status"] = "allowed_warning"
		endOfMonth := getEndOfMonthTime()
		mockRateLimiter.headers["anthropic-ratelimit-unified-overage-reset"] = formatTimestamp(endOfMonth)

	case MockScenarioOverageExhausted:
		if len(mockRateLimiter.exceededLimits) == 0 {
			mockRateLimiter.exceededLimits = append(mockRateLimiter.exceededLimits, ExceededLimit{
				Type:     ExceededLimitFiveHour,
				ResetsAt: fiveHoursFromNow,
			})
		}
		mockRateLimiter.updateRepresentativeClaim()
		mockRateLimiter.headers["anthropic-ratelimit-unified-status"] = "rejected"
		mockRateLimiter.headers["anthropic-ratelimit-unified-overage-status"] = "rejected"
		endOfMonth := getEndOfMonthTime()
		mockRateLimiter.headers["anthropic-ratelimit-unified-overage-reset"] = formatTimestamp(endOfMonth)

	case MockScenarioOutOfCredits:
		if len(mockRateLimiter.exceededLimits) == 0 {
			mockRateLimiter.exceededLimits = append(mockRateLimiter.exceededLimits, ExceededLimit{
				Type:     ExceededLimitFiveHour,
				ResetsAt: fiveHoursFromNow,
			})
		}
		mockRateLimiter.updateRepresentativeClaim()
		mockRateLimiter.headers["anthropic-ratelimit-unified-status"] = "rejected"
		mockRateLimiter.headers["anthropic-ratelimit-unified-overage-status"] = "rejected"
		mockRateLimiter.headers["anthropic-ratelimit-unified-overage-disabled-reason"] = "out_of_credits"
		endOfMonth := getEndOfMonthTime()
		mockRateLimiter.headers["anthropic-ratelimit-unified-overage-reset"] = formatTimestamp(endOfMonth)

	case MockScenarioOrgZeroCreditLimit:
		if len(mockRateLimiter.exceededLimits) == 0 {
			mockRateLimiter.exceededLimits = append(mockRateLimiter.exceededLimits, ExceededLimit{
				Type:     ExceededLimitFiveHour,
				ResetsAt: fiveHoursFromNow,
			})
		}
		mockRateLimiter.updateRepresentativeClaim()
		mockRateLimiter.headers["anthropic-ratelimit-unified-status"] = "rejected"
		mockRateLimiter.headers["anthropic-ratelimit-unified-overage-status"] = "rejected"
		mockRateLimiter.headers["anthropic-ratelimit-unified-overage-disabled-reason"] = "org_service_zero_credit_limit"
		endOfMonth := getEndOfMonthTime()
		mockRateLimiter.headers["anthropic-ratelimit-unified-overage-reset"] = formatTimestamp(endOfMonth)

	case MockScenarioOrgSpendCapHit:
		if len(mockRateLimiter.exceededLimits) == 0 {
			mockRateLimiter.exceededLimits = append(mockRateLimiter.exceededLimits, ExceededLimit{
				Type:     ExceededLimitFiveHour,
				ResetsAt: fiveHoursFromNow,
			})
		}
		mockRateLimiter.updateRepresentativeClaim()
		mockRateLimiter.headers["anthropic-ratelimit-unified-status"] = "rejected"
		mockRateLimiter.headers["anthropic-ratelimit-unified-overage-status"] = "rejected"
		mockRateLimiter.headers["anthropic-ratelimit-unified-overage-disabled-reason"] = "org_level_disabled_until"
		endOfMonth := getEndOfMonthTime()
		mockRateLimiter.headers["anthropic-ratelimit-unified-overage-reset"] = formatTimestamp(endOfMonth)

	case MockScenarioMemberZeroCreditLimit:
		if len(mockRateLimiter.exceededLimits) == 0 {
			mockRateLimiter.exceededLimits = append(mockRateLimiter.exceededLimits, ExceededLimit{
				Type:     ExceededLimitFiveHour,
				ResetsAt: fiveHoursFromNow,
			})
		}
		mockRateLimiter.updateRepresentativeClaim()
		mockRateLimiter.headers["anthropic-ratelimit-unified-status"] = "rejected"
		mockRateLimiter.headers["anthropic-ratelimit-unified-overage-status"] = "rejected"
		mockRateLimiter.headers["anthropic-ratelimit-unified-overage-disabled-reason"] = "member_zero_credit_limit"
		endOfMonth := getEndOfMonthTime()
		mockRateLimiter.headers["anthropic-ratelimit-unified-overage-reset"] = formatTimestamp(endOfMonth)

	case MockScenarioSeatTierZeroCreditLimit:
		if len(mockRateLimiter.exceededLimits) == 0 {
			mockRateLimiter.exceededLimits = append(mockRateLimiter.exceededLimits, ExceededLimit{
				Type:     ExceededLimitFiveHour,
				ResetsAt: fiveHoursFromNow,
			})
		}
		mockRateLimiter.updateRepresentativeClaim()
		mockRateLimiter.headers["anthropic-ratelimit-unified-status"] = "rejected"
		mockRateLimiter.headers["anthropic-ratelimit-unified-overage-status"] = "rejected"
		mockRateLimiter.headers["anthropic-ratelimit-unified-overage-disabled-reason"] = "seat_tier_zero_credit_limit"
		endOfMonth := getEndOfMonthTime()
		mockRateLimiter.headers["anthropic-ratelimit-unified-overage-reset"] = formatTimestamp(endOfMonth)

	case MockScenarioOpusLimit:
		mockRateLimiter.exceededLimits = append(mockRateLimiter.exceededLimits, ExceededLimit{
			Type:     ExceededLimitSevenDayOpus,
			ResetsAt: sevenDaysFromNow,
		})
		mockRateLimiter.updateRepresentativeClaim()
		mockRateLimiter.headers["anthropic-ratelimit-unified-status"] = "rejected"

	case MockScenarioOpusWarning:
		mockRateLimiter.headers["anthropic-ratelimit-unified-status"] = "allowed_warning"
		mockRateLimiter.headers["anthropic-ratelimit-unified-reset"] = formatTimestamp(sevenDaysFromNow)
		mockRateLimiter.headers["anthropic-ratelimit-unified-representative-claim"] = "seven_day_opus"

	case MockScenarioSonnetLimit:
		mockRateLimiter.exceededLimits = append(mockRateLimiter.exceededLimits, ExceededLimit{
			Type:     ExceededLimitSevenDaySonnet,
			ResetsAt: sevenDaysFromNow,
		})
		mockRateLimiter.updateRepresentativeClaim()
		mockRateLimiter.headers["anthropic-ratelimit-unified-status"] = "rejected"

	case MockScenarioSonnetWarning:
		mockRateLimiter.headers["anthropic-ratelimit-unified-status"] = "allowed_warning"
		mockRateLimiter.headers["anthropic-ratelimit-unified-reset"] = formatTimestamp(sevenDaysFromNow)
		mockRateLimiter.headers["anthropic-ratelimit-unified-representative-claim"] = "seven_day_sonnet"

	case MockScenarioFastModeLimit:
		mockRateLimiter.updateRepresentativeClaim()
		mockRateLimiter.headers["anthropic-ratelimit-unified-status"] = "rejected"
		mockRateLimiter.fastModeRateLimitDurationMs = 10 * 60 * 1000 // 10分钟

	case MockScenarioFastModeShortLimit:
		mockRateLimiter.updateRepresentativeClaim()
		mockRateLimiter.headers["anthropic-ratelimit-unified-status"] = "rejected"
		mockRateLimiter.fastModeRateLimitDurationMs = 10 * 1000 // 10秒

	case MockScenarioExtraUsageRequired:
		mockRateLimiter.headerless429Message = "Extra usage is required for long context requests."
	}
}

// GetMockHeaders 获取模拟响应头
func GetMockHeaders() MockHeaders {
	if !mockRateLimiter.enabled || len(mockRateLimiter.headers) == 0 {
		return nil
	}
	return mockRateLimiter.headers
}

// GetMockHeaderless429Message 获取无头429错误消息
func GetMockHeaderless429Message() string {
	return mockRateLimiter.headerless429Message
}

// GetCurrentMockScenario 获取当前模拟场景
func GetCurrentMockScenario() MockRateLimitScenario {
	if !mockRateLimiter.enabled {
		return ""
	}

	status := mockRateLimiter.headers["anthropic-ratelimit-unified-status"]
	overage := mockRateLimiter.headers["anthropic-ratelimit-unified-overage-status"]
	claim := mockRateLimiter.headers["anthropic-ratelimit-unified-representative-claim"]

	if claim == "seven_day_opus" {
		if status == "rejected" {
			return MockScenarioOpusLimit
		}
		return MockScenarioOpusWarning
	}

	if claim == "seven_day_sonnet" {
		if status == "rejected" {
			return MockScenarioSonnetLimit
		}
		return MockScenarioSonnetWarning
	}

	if overage == "rejected" {
		return MockScenarioOverageExhausted
	}
	if overage == "allowed_warning" {
		return MockScenarioOverageWarning
	}
	if overage == "allowed" {
		return MockScenarioOverageActive
	}

	if status == "rejected" {
		if claim == "five_hour" {
			return MockScenarioSessionLimitReached
		}
		if claim == "seven_day" {
			return MockScenarioWeeklyLimitReached
		}
	}

	if status == "allowed_warning" && claim == "seven_day" {
		return MockScenarioApproachingWeeklyLimit
	}

	if status == "allowed" {
		return MockScenarioNormal
	}

	return ""
}

// ClearMockHeaders 清除模拟头
func ClearMockHeaders() {
	mockRateLimiter.headers = make(MockHeaders)
	mockRateLimiter.exceededLimits = nil
	mockRateLimiter.headerless429Message = ""
	mockRateLimiter.fastModeRateLimitDurationMs = 0
	mockRateLimiter.fastModeRateLimitExpiresAt = 0
	mockRateLimiter.enabled = false
}

// ShouldProcessMockLimits 检查是否应该处理模拟限制
func ShouldProcessMockLimits() bool {
	return mockRateLimiter.enabled
}

// 辅助函数

func parseHours(s string) int {
	var hours int
	for _, c := range s {
		if c >= '0' && c <= '9' {
			hours = hours*10 + int(c-'0')
		} else {
			return 0
		}
	}
	return hours
}

func formatTimestamp(t int64) string {
	return formatInt(t)
}

func formatInt(i int64) string {
	if i == 0 {
		return "0"
	}
	result := ""
	negative := false
	if i < 0 {
		negative = true
		i = -i
	}
	for i > 0 {
		result = string(rune('0'+i%10)) + result
		i /= 10
	}
	if result == "" {
		result = "0"
	}
	if negative {
		result = "-" + result
	}
	return result
}

func getEndOfMonthTime() int64 {
	now := time.Now()
	nextMonth := time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, now.Location())
	return nextMonth.Unix()
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// IsMockFastModeRateLimitScenario 检查是否是快速模式速率限制场景
func IsMockFastModeRateLimitScenario() bool {
	return mockRateLimiter.fastModeRateLimitDurationMs > 0
}
