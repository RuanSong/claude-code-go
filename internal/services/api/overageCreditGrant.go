package api

import (
	"os"
	"sync"
	"time"
)

// OverageCreditGrantInfo 超额信用授予信息
type OverageCreditGrantInfo struct {
	Available        bool   `json:"available"`
	Eligible         bool   `json:"eligible"`
	Granted          bool   `json:"granted"`
	AmountMinorUnits *int64 `json:"amount_minor_units,omitempty"`
	Currency         string `json:"currency,omitempty"`
}

// CachedGrantEntry 缓存条目
type CachedGrantEntry struct {
	Info      OverageCreditGrantInfo `json:"info"`
	Timestamp int64                  `json:"timestamp"`
}

// 缓存TTL - 1小时
const OVERAGE_CACHE_TTL_MS = 60 * 60 * 1000

var (
	overageCache      map[string]*CachedGrantEntry
	overageCacheMutex sync.RWMutex
)

// fetchOverageCreditGrant 获取超额信用授予信息
func fetchOverageCreditGrant(accessToken, orgUUID string) (*OverageCreditGrantInfo, error) {
	// 实际实现需要调用API
	// 简化版本返回nil
	return nil, nil
}

// GetCachedOverageCreditGrant 获取缓存的超额信用授予信息
// 如果没有缓存或缓存过期返回nil
func GetCachedOverageCreditGrant() *OverageCreditGrantInfo {
	orgID := getOAuthAccountInfoOrgID()
	if orgID == "" {
		return nil
	}

	overageCacheMutex.RLock()
	defer overageCacheMutex.RUnlock()

	if overageCache == nil {
		return nil
	}

	cached, ok := overageCache[orgID]
	if !ok {
		return nil
	}

	if time.Now().UnixMilli()-cached.Timestamp > OVERAGE_CACHE_TTL_MS {
		return nil
	}

	return &cached.Info
}

// InvalidateOverageCreditGrantCache 使缓存失效
func InvalidateOverageCreditGrantCache() {
	orgID := getOAuthAccountInfoOrgID()
	if orgID == "" {
		return
	}

	overageCacheMutex.Lock()
	defer overageCacheMutex.Unlock()

	if overageCache != nil {
		delete(overageCache, orgID)
	}
}

// RefreshOverageCreditGrantCache 刷新超额信用授予缓存
// 异步调用，在需要显示界面时触发
func RefreshOverageCreditGrantCache() error {
	if isEssentialTrafficOnly() {
		return nil
	}

	orgID := getOAuthAccountInfoOrgID()
	if orgID == "" {
		return nil
	}

	info, err := fetchOverageCreditGrant("", orgID)
	if err != nil || info == nil {
		return err
	}

	// 检查数据是否变化
	overageCacheMutex.Lock()
	defer overageCacheMutex.Unlock()

	var existing OverageCreditGrantInfo
	prevCached := overageCache[orgID]
	if prevCached != nil {
		existing = prevCached.Info
	}

	dataUnchanged := existing.Available == info.Available &&
		existing.Eligible == info.Eligible &&
		existing.Granted == info.Granted &&
		existing.AmountMinorUnits == info.AmountMinorUnits &&
		existing.Currency == info.Currency

	if dataUnchanged && prevCached != nil && time.Now().UnixMilli()-prevCached.Timestamp <= OVERAGE_CACHE_TTL_MS {
		return nil
	}

	if overageCache == nil {
		overageCache = make(map[string]*CachedGrantEntry)
	}

	entry := &CachedGrantEntry{
		Info:      *info,
		Timestamp: time.Now().UnixMilli(),
	}

	overageCache[orgID] = entry
	return nil
}

// FormatGrantAmount 格式化授予金额显示
// 如果金额不可用返回nil
func FormatGrantAmount(info *OverageCreditGrantInfo) string {
	if info.AmountMinorUnits == nil || info.Currency == "" {
		return ""
	}

	// 目前只支持USD
	if info.Currency == "USD" || info.Currency == "usd" {
		dollars := float64(*info.AmountMinorUnits) / 100.0
		if dollars == float64(int64(dollars)) {
			return "$" + formatInt64(int64(dollars))
		}
		return formatFloat(dollars)
	}

	return ""
}

func formatInt64(n int64) string {
	if n < 0 {
		return "-" + formatInt64(-n)
	}
	if n == 0 {
		return "0"
	}
	digits := ""
	for n > 0 {
		digits = string(rune('0'+n%10)) + digits
		n /= 10
	}
	return digits
}

func formatFloat(f float64) string {
	intPart := int64(f)
	decPart := int64((f - float64(intPart)) * 100)
	return "$" + formatInt64(intPart) + "." + formatInt64(decPart)
}

// getOAuthAccountInfoOrgID 获取OAuth账户信息的组织ID
func getOAuthAccountInfoOrgID() string {
	return os.Getenv("CLAUDE_ORG_UUID")
}
