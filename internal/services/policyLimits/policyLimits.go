package policyLimits

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// PolicyLimitsResponse 策略限制响应
type PolicyLimitsResponse struct {
	Restrictions Restrictions `json:"restrictions"`
}

// Restrictions 限制映射
type Restrictions map[string]Restriction

// Restriction 单个限制
type Restriction struct {
	Allowed bool `json:"allowed"`
}

// PolicyLimitsFetchResult 策略限制获取结果
type PolicyLimitsFetchResult struct {
	Success      bool
	Restrictions Restrictions
	ETag         string
	Error        string
	SkipRetry    bool
}

const (
	// CACHE_FILENAME 缓存文件名
	CACHE_FILENAME = "policy-limits.json"
	// FETCH_TIMEOUT_MS 获取超时时间
	FETCH_TIMEOUT_MS = 10000
	// DEFAULT_MAX_RETRIES 最大重试次数
	DEFAULT_MAX_RETRIES = 5
	// POLLING_INTERVAL_MS 轮询间隔
	POLLING_INTERVAL_MS = 60 * 60 * 1000
	// LOADING_PROMISE_TIMEOUT_MS 加载超时
	LOADING_PROMISE_TIMEOUT_MS = 30000
)

var (
	// sessionCache 会话级缓存
	sessionCache Restrictions
	// cacheMutex 缓存互斥锁
	cacheMutex sync.RWMutex
	// loadingCompletePromise 加载完成promise
	loadingCompletePromise *Promise
	// loadingCompleteResolve 加载完成resolve函数
	loadingCompleteResolve func()
	// pollingIntervalId 轮询interval ID
	pollingIntervalId *time.Timer
	// cleanupRegistered 清理是否已注册
	cleanupRegistered bool
)

// Promise Go风格的Promise简化实现
type Promise struct {
	done   chan struct{}
	result error
}

func NewPromise() *Promise {
	return &Promise{done: make(chan struct{})}
}

func (p *Promise) Resolve() {
	close(p.done)
}

func (p *Promise) Wait() error {
	<-p.done
	return p.result
}

// getCachePath 获取缓存文件路径
func getCachePath() string {
	configHome := getClaudeConfigHomeDir()
	return filepath.Join(configHome, CACHE_FILENAME)
}

// getClaudeConfigHomeDir 获取Claude配置目录
func getClaudeConfigHomeDir() string {
	if home := os.Getenv("CLAUDE_CONFIG_DIR"); home != "" {
		return home
	}
	if home := os.Getenv("HOME"); home != "" {
		return filepath.Join(home, ".config", "claude")
	}
	return ""
}

// getPolicyLimitsEndpoint 获取策略限制API端点
func getPolicyLimitsEndpoint() string {
	baseURL := getOAuthBaseAPIURL()
	return fmt.Sprintf("%s/api/claude_code/policy_limits", baseURL)
}

// getOAuthBaseAPIURL 获取OAuth基础API URL
func getOAuthBaseAPIURL() string {
	if url := os.Getenv("OAUTH_BASE_API_URL"); url != "" {
		return url
	}
	return "https://api.anthropic.com"
}

// isFirstPartyAnthropicBaseUrl 检查是否使用第一方Anthropic URL
func isFirstPartyAnthropicBaseUrl() bool {
	baseURL := os.Getenv("ANTHROPIC_BASE_URL")
	if baseURL == "" {
		return true // 默认认为是第一方
	}
	return strings.Contains(baseURL, "anthropic.com") && !strings.Contains(baseURL, "bedrock")
}

// computeChecksum 计算校验和
func computeChecksum(restrictions Restrictions) string {
	sorted := sortKeysDeep(restrictions)
	normalized, _ := json.Marshal(sorted)
	hash := sha256.Sum256(normalized)
	return fmt.Sprintf("sha256:%x", hash)
}

// sortKeysDeep 递归排序对象键
func sortKeysDeep(obj interface{}) interface{} {
	switch v := obj.(type) {
	case map[string]interface{}:
		keys := make([]string, 0, len(v))
		for k := range v {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		sorted := make(map[string]interface{})
		for _, k := range keys {
			sorted[k] = sortKeysDeep(v[k])
		}
		return sorted
	case []interface{}:
		result := make([]interface{}, len(v))
		for i, item := range v {
			result[i] = sortKeysDeep(item)
		}
		return result
	default:
		return v
	}
}

// isPolicyLimitsEligible 检查用户是否有资格获取策略限制
func isPolicyLimitsEligible() bool {
	// 第三方提供商用户不应该访问策略限制端点
	if !isFirstPartyAnthropicBaseUrl() {
		return false
	}

	// Console用户(API key)
	if apiKey := os.Getenv("ANTHROPIC_API_KEY"); apiKey != "" {
		return true
	}

	// OAuth用户检查
	tokens := getClaudeAIOAuthTokens()
	if tokens == nil || tokens.AccessToken == "" {
		return false
	}

	// 必须有Claude.ai推理范围
	if tokens.Scopes == nil || !containsScope(tokens.Scopes, "claude-ai-inference") {
		return false
	}

	// 只有Team和Enterprise OAuth用户有资格
	if tokens.SubscriptionType != "enterprise" && tokens.SubscriptionType != "team" {
		return false
	}

	return true
}

// containsScope 检查范围列表是否包含指定范围
func containsScope(scopes []string, scope string) bool {
	for _, s := range scopes {
		if s == scope {
			return true
		}
	}
	return false
}

// OAUTHTokens OAuth令牌
type OAUTHTokens struct {
	AccessToken      string
	RefreshToken     string
	ExpiresAt        time.Time
	Scopes           []string
	SubscriptionType string
}

// getClaudeAIOAuthTokens 获取Claude AI OAuth令牌
func getClaudeAIOAuthTokens() *OAUTHTokens {
	accessToken := os.Getenv("CLAUDEAI_ACCESS_TOKEN")
	if accessToken == "" {
		return nil
	}
	return &OAUTHTokens{
		AccessToken:      accessToken,
		Scopes:           strings.Split(os.Getenv("CLAUDEAI_SCOPES"), ","),
		SubscriptionType: os.Getenv("CLAUDE_SUBSCRIPTION_TYPE"),
	}
}

// initializePolicyLimitsLoadingPromise 初始化策略限制加载promise
func initializePolicyLimitsLoadingPromise() {
	if loadingCompletePromise != nil {
		return
	}

	if isPolicyLimitsEligible() {
		loadingCompletePromise = NewPromise()
		go func() {
			time.Sleep(LOADING_PROMISE_TIMEOUT_MS * time.Millisecond)
			if loadingCompleteResolve != nil {
				loadingCompleteResolve()
				loadingCompleteResolve = nil
			}
		}()
	}
}

// waitForPolicyLimitsToLoad 等待策略限制加载完成
func waitForPolicyLimitsToLoad() error {
	if loadingCompletePromise != nil {
		return loadingCompletePromise.Wait()
	}
	return nil
}

// loadCachedRestrictions 从缓存加载限制
func loadCachedRestrictions() Restrictions {
	cacheMutex.RLock()
	defer cacheMutex.RUnlock()

	path := getCachePath()
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	var response PolicyLimitsResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil
	}

	return response.Restrictions
}

// saveCachedRestrictions 保存限制到缓存
func saveCachedRestrictions(restrictions Restrictions) error {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	path := getCachePath()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data := PolicyLimitsResponse{Restrictions: restrictions}
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, jsonData, 0600)
}

// isPolicyAllowed 检查特定策略是否允许
func isPolicyAllowed(policy string) bool {
	restrictions := getRestrictionsFromCache()
	if restrictions == nil {
		// essential-traffic-only模式下某些策略默认拒绝
		if isEssentialTrafficOnly() && essentialTrafficDenyOnMiss(policy) {
			return false
		}
		return true // 默认开放
	}

	restriction, ok := restrictions[policy]
	if !ok {
		return true // 未知策略 = 允许
	}
	return restriction.Allowed
}

// essentialTrafficDenyOnMiss 检查是否在essential-traffic-only模式下默认拒绝
func essentialTrafficDenyOnMiss(policy string) bool {
	denyList := []string{"allow_product_feedback"}
	for _, p := range denyList {
		if p == policy {
			return true
		}
	}
	return false
}

// isEssentialTrafficOnly 检查是否只允许必要流量
func isEssentialTrafficOnly() bool {
	return os.Getenv("ESSENTIAL_TRAFFIC_ONLY") == "true"
}

// getRestrictionsFromCache 从缓存获取限制
func getRestrictionsFromCache() Restrictions {
	if !isPolicyLimitsEligible() {
		return nil
	}

	cacheMutex.RLock()
	defer cacheMutex.RUnlock()

	if sessionCache != nil {
		return sessionCache
	}

	cached := loadCachedRestrictions()
	if cached != nil {
		sessionCache = cached
	}
	return cached
}

// fetchPolicyLimits 获取策略限制(单次尝试)
func fetchPolicyLimits(cachedChecksum string) PolicyLimitsFetchResult {
	// 注意：这里需要实际的HTTP请求
	// 简化实现返回失败
	return PolicyLimitsFetchResult{
		Success:   false,
		Error:     "fetch not implemented",
		SkipRetry: true,
	}
}

// fetchWithRetry 带重试获取策略限制
func fetchWithRetry(cachedChecksum string) PolicyLimitsFetchResult {
	var lastResult PolicyLimitsFetchResult

	for attempt := 1; attempt <= DEFAULT_MAX_RETRIES+1; attempt++ {
		lastResult = fetchPolicyLimits(cachedChecksum)

		if lastResult.Success {
			return lastResult
		}

		if lastResult.SkipRetry {
			return lastResult
		}

		if attempt > DEFAULT_MAX_RETRIES {
			return lastResult
		}

		delayMs := getRetryDelay(attempt)
		time.Sleep(time.Duration(delayMs) * time.Millisecond)
	}

	return lastResult
}

// getRetryDelay 获取重试延迟
func getRetryDelay(attempt int) int {
	// 指数退避: 1s, 2s, 4s, 8s, 16s
	if attempt <= 0 {
		return 1000
	}
	delay := 1000
	for i := 1; i < attempt; i++ {
		delay *= 2
	}
	if delay > 16000 {
		delay = 16000
	}
	return delay
}

// loadPolicyLimits 加载策略限制
func loadPolicyLimits() error {
	if !isPolicyLimitsEligible() {
		return nil
	}

	// 如果已有缓存，直接使用
	cached := loadCachedRestrictions()
	if cached != nil {
		cacheMutex.Lock()
		sessionCache = cached
		cacheMutex.Unlock()
	}

	// 注意：实际实现需要异步获取
	return nil
}

// refreshPolicyLimits 刷新策略限制
func refreshPolicyLimits() error {
	if err := clearPolicyLimitsCache(); err != nil {
		return err
	}
	return loadPolicyLimits()
}

// clearPolicyLimitsCache 清除策略限制缓存
func clearPolicyLimitsCache() error {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	sessionCache = nil
	loadingCompletePromise = nil
	loadingCompleteResolve = nil

	// 删除缓存文件
	path := getCachePath()
	os.Remove(path)

	return nil
}

// getPolicyAllowed 检查策略是否允许的导出函数
func GetPolicyAllowed(policy string) bool {
	return isPolicyAllowed(policy)
}

// IsPolicyLimitsEligible 是否符合策略限制资格
func IsPolicyLimitsEligible() bool {
	return isPolicyLimitsEligible()
}
