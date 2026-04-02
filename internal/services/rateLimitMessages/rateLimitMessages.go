package rateLimitMessages

import (
	"fmt"
	"strings"
	"time"

	"github.com/claude-code-go/claude/internal/services/claudeAiLimits"
)

// RATE_LIMIT_ERROR_PREFIXES 速率限制错误前缀
var RATE_LIMIT_ERROR_PREFIXES = []string{
	"You've hit your",
	"You've used",
	"You're now using extra usage",
	"You're close to",
	"You're out of extra usage",
}

// RateLimitMessage 速率限制消息
type RateLimitMessage struct {
	Message  string // 消息内容
	Severity string // 严重程度: "error" 或 "warning"
}

// IsRateLimitErrorMessage 检查是否为速率限制错误消息
func IsRateLimitErrorMessage(text string) bool {
	for _, prefix := range RATE_LIMIT_ERROR_PREFIXES {
		if strings.HasPrefix(text, prefix) {
			return true
		}
	}
	return false
}

// GetRateLimitMessage 根据限制状态获取适当的速率限制消息
// 返回nil如果不需要显示消息
func GetRateLimitMessage(limits *claudeAiLimits.ClaudeAILimits, model string) *RateLimitMessage {
	// 首先检查超额场景
	if limits.IsUsingOverage {
		if limits.OverageStatus == claudeAiLimits.QuotaStatusAllowedWarning {
			return &RateLimitMessage{
				Message:  "You're close to your extra usage spending limit",
				Severity: "warning",
			}
		}
		return nil
	}

	// 错误状态 - 限制被拒绝
	if limits.Status == claudeAiLimits.QuotaStatusRejected {
		return &RateLimitMessage{
			Message:  GetLimitReachedText(limits, model),
			Severity: "error",
		}
	}

	// 警告状态 - 接近限制
	if limits.Status == claudeAiLimits.QuotaStatusAllowedWarning {
		// 仅当使用量超过阈值(70%)时显示警告
		// 这可以防止API在低使用量时发送过时数据后产生误报
		warningThreshold := 0.7
		if limits.Utilization < warningThreshold {
			return nil
		}

		text := GetEarlyWarningText(limits)
		if text != "" {
			return &RateLimitMessage{
				Message:  text,
				Severity: "warning",
			}
		}
	}

	return nil
}

// GetRateLimitErrorMessage 获取API错误的错误消息
// 返回nil如果不显示错误消息
func GetRateLimitErrorMessage(limits *claudeAiLimits.ClaudeAILimits, model string) string {
	msg := GetRateLimitMessage(limits, model)
	if msg != nil && msg.Severity == "error" {
		return msg.Message
	}
	return ""
}

// GetRateLimitWarning 获取UI页脚的警告消息
// 返回nil如果不显示警告
func GetRateLimitWarning(limits *claudeAiLimits.ClaudeAILimits, model string) string {
	msg := GetRateLimitMessage(limits, model)
	if msg != nil && msg.Severity == "warning" {
		return msg.Message
	}
	return ""
}

// GetLimitReachedText 获取限制达到文本
func GetLimitReachedText(limits *claudeAiLimits.ClaudeAILimits, model string) string {
	var resetTimeStr string
	if limits.ResetsAt > 0 {
		resetTimeStr = formatResetTime(limits.ResetsAt, true)
	}

	var resetMessage string
	if resetTimeStr != "" {
		resetMessage = " · resets " + resetTimeStr
	}

	var overageResetTimeStr string
	if limits.OverageResetsAt > 0 {
		overageResetTimeStr = formatResetTime(limits.OverageResetsAt, true)
	}

	// 如果订阅和超额都耗尽
	if limits.OverageStatus == claudeAiLimits.QuotaStatusRejected {
		var overageResetMessage string
		if limits.ResetsAt > 0 && limits.OverageResetsAt > 0 {
			if limits.ResetsAt < limits.OverageResetsAt {
				overageResetMessage = " · resets " + resetTimeStr
			} else {
				overageResetMessage = " · resets " + overageResetTimeStr
			}
		} else if resetTimeStr != "" {
			overageResetMessage = " · resets " + resetTimeStr
		} else if overageResetTimeStr != "" {
			overageResetMessage = " · resets " + overageResetTimeStr
		}

		if limits.OverageDisabledReason == claudeAiLimits.OutOfCredits {
			return fmt.Sprintf("You're out of extra usage%s", overageResetMessage)
		}

		return formatLimitReachedText("limit", overageResetMessage, model)
	}

	if limits.RateLimitType == claudeAiLimits.RateLimitTypeSevenDaySonnet {
		limit := "Sonnet limit"
		return formatLimitReachedText(limit, resetMessage, model)
	}

	if limits.RateLimitType == claudeAiLimits.RateLimitTypeSevenDayOpus {
		return formatLimitReachedText("Opus limit", resetMessage, model)
	}

	if limits.RateLimitType == claudeAiLimits.RateLimitTypeSevenDay {
		return formatLimitReachedText("weekly limit", resetMessage, model)
	}

	if limits.RateLimitType == claudeAiLimits.RateLimitTypeFiveHour {
		return formatLimitReachedText("session limit", resetMessage, model)
	}

	return formatLimitReachedText("usage limit", resetMessage, model)
}

// GetEarlyWarningText 获取早期警告文本
func GetEarlyWarningText(limits *claudeAiLimits.ClaudeAILimits) string {
	var limitName string
	switch limits.RateLimitType {
	case claudeAiLimits.RateLimitTypeSevenDay:
		limitName = "weekly limit"
	case claudeAiLimits.RateLimitTypeFiveHour:
		limitName = "session limit"
	case claudeAiLimits.RateLimitTypeSevenDayOpus:
		limitName = "Opus limit"
	case claudeAiLimits.RateLimitTypeSevenDaySonnet:
		limitName = "Sonnet limit"
	case claudeAiLimits.RateLimitTypeOverage:
		limitName = "extra usage"
	default:
		return ""
	}

	var usedStr string
	if limits.Utilization > 0 {
		used := int(limits.Utilization * 100)
		usedStr = fmt.Sprintf("You've used %d%% of your %s", used, limitName)
	}

	var resetTimeStr string
	if limits.ResetsAt > 0 {
		resetTimeStr = formatResetTime(limits.ResetsAt, true)
	}

	var upsell string
	if getWarningUpsellText(limits.RateLimitType) != "" {
		upsell = " · " + getWarningUpsellText(limits.RateLimitType)
	}

	if usedStr != "" && resetTimeStr != "" {
		return fmt.Sprintf("%s · resets %s%s", usedStr, resetTimeStr, upsell)
	}

	if usedStr != "" {
		return usedStr + upsell
	}

	if limits.RateLimitType == claudeAiLimits.RateLimitTypeOverage {
		limitName += " limit"
	}

	if resetTimeStr != "" {
		base := fmt.Sprintf("Approaching %s · resets %s", limitName, resetTimeStr)
		if upsell != "" {
			return base + upsell
		}
		return base
	}

	base := fmt.Sprintf("Approaching %s", limitName)
	if upsell != "" {
		return base + upsell
	}
	return base
}

// getWarningUpsellText 根据订阅类型和限制类型获取警告upsell文本
func getWarningUpsellText(rateLimitType claudeAiLimits.RateLimitType) string {
	// 注意: 在完整实现中,这会检查订阅类型和超额启用状态
	// 这里返回简化的版本
	switch rateLimitType {
	case claudeAiLimits.RateLimitTypeFiveHour:
		return "/upgrade to keep using Claude Code"
	case claudeAiLimits.RateLimitTypeOverage:
		return "/extra-usage to request more"
	}
	return ""
}

// formatLimitReachedText 格式化限制达到文本
func formatLimitReachedText(limit, resetMessage, model string) string {
	return fmt.Sprintf("You've hit your %s%s", limit, resetMessage)
}

// GetUsingOverageText 获取超额模式转换的通知文本
func GetUsingOverageText(limits *claudeAiLimits.ClaudeAILimits) string {
	var resetTimeStr string
	if limits.ResetsAt > 0 {
		resetTimeStr = formatResetTime(limits.ResetsAt, true)
	}

	var limitName string
	switch limits.RateLimitType {
	case claudeAiLimits.RateLimitTypeFiveHour:
		limitName = "session limit"
	case claudeAiLimits.RateLimitTypeSevenDay:
		limitName = "weekly limit"
	case claudeAiLimits.RateLimitTypeSevenDayOpus:
		limitName = "Opus limit"
	case claudeAiLimits.RateLimitTypeSevenDaySonnet:
		limitName = "Sonnet limit"
	}

	if limitName == "" {
		return "Now using extra usage"
	}

	var resetMessage string
	if resetTimeStr != "" {
		resetMessage = fmt.Sprintf(" · Your %s resets %s", limitName, resetTimeStr)
	}

	return fmt.Sprintf("You're now using extra usage%s", resetMessage)
}

// formatResetTime 格式化重置时间
func formatResetTime(resetsAt int64, relative bool) string {
	if resetsAt <= 0 {
		return ""
	}

	resetTime := time.Unix(resetsAt, 0)
	now := time.Now()
	duration := resetTime.Sub(now)

	if duration < 0 {
		return "now"
	}

	hours := int(duration.Hours())
	minutes := int(duration.Minutes()) % 60

	if relative {
		if hours >= 24 {
			days := hours / 24
			hours = hours % 24
			if hours > 0 {
				return fmt.Sprintf("in %dd %dh", days, hours)
			}
			return fmt.Sprintf("in %dd", days)
		}
		if hours > 0 {
			return fmt.Sprintf("in %dh %dm", hours, minutes)
		}
		if minutes > 0 {
			return fmt.Sprintf("in %dm", minutes)
		}
		return "in less than a minute"
	}

	return resetTime.Format("3:04 PM")
}
