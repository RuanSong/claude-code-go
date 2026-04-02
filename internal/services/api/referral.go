package api

import (
	"os"
	"sync"
	"time"
)

// ReferralCampaign 推荐活动类型
type ReferralCampaign string

const (
	CClaudeCodeGuestPass ReferralCampaign = "claude_code_guest_pass"
)

// ReferralEligibilityResponse 推荐资格响应
type ReferralEligibilityResponse struct {
	Eligible           bool   `json:"eligible"`
	Reason             string `json:"reason,omitempty"`
	MaxRedemptions     int    `json:"max_redemptions,omitempty"`
	CurrentRedemptions int    `json:"current_redemptions,omitempty"`
}

// ReferralRedemptionsResponse 推荐兑换响应
type ReferralRedemptionsResponse struct {
	Redemptions []ReferralRedemption `json:"redemptions"`
	TotalCount  int                  `json:"total_count"`
}

// ReferralRedemption 推荐兑换记录
type ReferralRedemption struct {
	ID         string    `json:"id"`
	Email      string    `json:"email"`
	RedeemedAt time.Time `json:"redeemed_at"`
	Campaign   string    `json:"campaign"`
	Status     string    `json:"status"`
}

// ReferrerRewardInfo 推荐人奖励信息
type ReferrerRewardInfo struct {
	CreditAmount   int64  `json:"credit_amount"`
	Currency       string `json:"currency"`
	ExpirationDate string `json:"expiration_date"`
}

// 缓存过期时间: 24小时
const CACHE_EXPIRATION_MS = 24 * 60 * 60 * 1000

var (
	referralCache      *ReferralEligibilityResponse
	referralCacheTime  int64
	referralCacheMutex sync.RWMutex
	fetchInProgress    *ReferralEligibilityResponse
	fetchMutex         sync.Mutex
)

// FetchReferralEligibility 获取推荐资格
func FetchReferralEligibility(campaign ReferralCampaign) (*ReferralEligibilityResponse, error) {
	// 检查缓存
	referralCacheMutex.RLock()
	if referralCache != nil && time.Now().UnixMilli()-referralCacheTime < CACHE_EXPIRATION_MS {
		referralCacheMutex.RUnlock()
		return referralCache, nil
	}
	referralCacheMutex.RUnlock()

	// 检查进行中的请求
	fetchMutex.Lock()
	if fetchInProgress != nil {
		fetchMutex.Unlock()
		return fetchInProgress, nil
	}

	// 获取OAuth信息
	orgID := getOAuthAccountInfoOrgID()
	if orgID == "" {
		return &ReferralEligibilityResponse{Eligible: false, Reason: "not_authenticated"}, nil
	}

	if !isClaudeAISubscriber() {
		return &ReferralEligibilityResponse{Eligible: false, Reason: "not_subscriber"}, nil
	}

	subscriptionType := getSubscriptionType()
	if subscriptionType != "max" {
		return &ReferralEligibilityResponse{Eligible: false, Reason: "not_max_subscription"}, nil
	}

	// 实际实现需要调用API
	// 简化版本返回未实现
	result := &ReferralEligibilityResponse{Eligible: false, Reason: "not_implemented"}
	fetchInProgress = result
	fetchMutex.Unlock()

	// 更新缓存
	referralCacheMutex.Lock()
	referralCache = result
	referralCacheTime = time.Now().UnixMilli()
	referralCacheMutex.Unlock()

	return result, nil
}

// FetchReferralRedemptions 获取推荐兑换记录
func FetchReferralRedemptions(campaign string) (*ReferralRedemptionsResponse, error) {
	orgID := getOAuthAccountInfoOrgID()
	if orgID == "" {
		return nil, nil
	}

	// 简化实现
	return &ReferralRedemptionsResponse{
		Redemptions: []ReferralRedemption{},
		TotalCount:  0,
	}, nil
}

// ShouldCheckForPasses 检查是否应该检查通行证
func ShouldCheckForPasses() bool {
	orgID := getOAuthAccountInfoOrgID()
	if orgID == "" {
		return false
	}
	if !isClaudeAISubscriber() {
		return false
	}
	return getSubscriptionType() == "max"
}

// IsReferralEligible 检查推荐资格
func IsReferralEligible() bool {
	result, err := FetchReferralEligibility(CClaudeCodeGuestPass)
	if err != nil || result == nil {
		return false
	}
	return result.Eligible
}

// GetReferralURL 获取推荐链接
func GetReferralURL() string {
	return "https://claude.ai/referral"
}

// getSubscriptionType 获取订阅类型
func getSubscriptionType() string {
	return os.Getenv("CLAUDE_SUBSCRIPTION_TYPE")
}

// ClearReferralCache 清除推荐缓存
func ClearReferralCache() {
	referralCacheMutex.Lock()
	defer referralCacheMutex.Unlock()
	referralCache = nil
	referralCacheTime = 0
}
