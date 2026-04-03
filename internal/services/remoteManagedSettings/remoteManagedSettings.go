package services

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

// RemoteManagedSettings 远程托管设置服务
// 为企业客户提供远程托管设置的获取、缓存和验证
// 使用校验和验证来最小化网络流量，并在失败时优雅降级

const (
	SettingsTimeoutMs     = 10000   // 设置获取超时10秒
	DefaultMaxRetries     = 5       // 默认最大重试次数
	PollingIntervalMs     = 3600000 // 1小时轮询间隔
	LoadingPromiseTimeout = 30000   // 30秒加载超时
)

// RemoteManagedSettingsResponse 远程托管设置响应
type RemoteManagedSettingsResponse struct {
	UUID     string                 `json:"uuid"`
	Checksum string                 `json:"checksum"`
	Settings map[string]interface{} `json:"settings"`
}

// RemoteManagedSettingsFetchResult 获取结果
type RemoteManagedSettingsFetchResult struct {
	Success   bool
	Settings  map[string]interface{}
	Checksum  string
	Error     string
	SkipRetry bool
}

// 远程托管设置服务
type RemoteManagedSettingsService struct {
	mu                sync.RWMutex
	pollingIntervalId *time.Timer
	loadingComplete   chan struct{}
	sessionCache      map[string]interface{}
	eligible          bool
	cachedEligible    *bool
}

// 全局实例
var remoteSettingsService = &RemoteManagedSettingsService{
	loadingComplete: make(chan struct{}),
	sessionCache:    nil,
	eligible:        false,
}

// GetRemoteManagedSettingsEndpoint 获取远程设置API端点
func GetRemoteManagedSettingsEndpoint() string {
	// 使用OAuth配置的基础API URL
	return "https://auth.anthropic.com/api/claude_code/settings"
}

// sortKeysDeep 递归排序对象的所有键
func sortKeysDeep(obj interface{}) interface{} {
	if arr, ok := obj.([]interface{}); ok {
		result := make([]interface{}, len(arr))
		for i, v := range arr {
			result[i] = sortKeysDeep(v)
		}
		return result
	}
	if m, ok := obj.(map[string]interface{}); ok {
		keys := make([]string, 0, len(m))
		for k := range m {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		result := make(map[string]interface{})
		for _, k := range keys {
			result[k] = sortKeysDeep(m[k])
		}
		return result
	}
	return obj
}

// ComputeChecksumFromSettings 计算设置的校验和
func ComputeChecksumFromSettings(settings map[string]interface{}) string {
	sorted := sortKeysDeep(settings)
	data, _ := json.Marshal(sorted)
	hash := sha256.Sum256(data)
	return fmt.Sprintf("sha256:%x", hash)
}

// IsEligibleForRemoteManagedSettings 检查用户是否有资格使用远程托管设置
func IsEligibleForRemoteManagedSettings() bool {
	return isRemoteManagedSettingsEligible()
}

// isRemoteManagedSettingsEligible 内部资格检查
func isRemoteManagedSettingsEligible() bool {
	if remoteSettingsService.cachedEligible != nil {
		return *remoteSettingsService.cachedEligible
	}

	// 检查是否使用第三方提供商
	if !isFirstPartyProvider() {
		eligible := false
		remoteSettingsService.cachedEligible = &eligible
		return false
	}

	// 检查是否使用自定义基础URL
	if !isDefaultBaseUrl() {
		eligible := false
		remoteSettingsService.cachedEligible = &eligible
		return false
	}

	// TODO: 检查OAuth令牌和订阅类型
	// 企业版和团队版用户有资格

	eligible := true
	remoteSettingsService.cachedEligible = &eligible
	return true
}

func isFirstPartyProvider() bool {
	// 检查API提供商
	// 在Go版本中，这需要检查配置
	return true
}

func isDefaultBaseUrl() bool {
	// 检查是否使用默认基础URL
	baseUrl := os.Getenv("ANTHROPIC_BASE_URL")
	return baseUrl == "" || baseUrl == "https://api.anthropic.com"
}

// WaitForRemoteManagedSettingsToLoad 等待初始远程设置加载完成
func WaitForRemoteManagedSettingsToLoad() {
	select {
	case <-remoteSettingsService.loadingComplete:
		return
	case <-time.After(time.Duration(LoadingPromiseTimeout) * time.Millisecond):
		return
	}
}

// getRemoteSettingsAuthHeaders 获取远程设置的认证头
func getRemoteSettingsAuthHeaders() (map[string]string, error) {
	// 尝试API密钥（适用于控制台用户）
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey != "" {
		return map[string]string{
			"x-api-key": apiKey,
		}, nil
	}

	// 备用OAuth令牌（适用于Claude.ai用户）
	// TODO: 实现OAuth令牌获取

	return nil, fmt.Errorf("no authentication available")
}

// fetchRemoteManagedSettings 获取远程托管设置
func fetchRemoteManagedSettings(cachedChecksum string) RemoteManagedSettingsFetchResult {
	// 确保OAuth令牌是最新的
	checkAndRefreshOAuthTokenIfNeeded()

	authHeaders, err := getRemoteSettingsAuthHeaders()
	if err != nil {
		return RemoteManagedSettingsFetchResult{
			Success:   false,
			Error:     fmt.Sprintf("Authentication required for remote settings: %v", err),
			SkipRetry: true,
		}
	}

	endpoint := GetRemoteManagedSettingsEndpoint()
	headers := map[string]string{
		"User-Agent": getClaudeCodeUserAgent(),
	}
	for k, v := range authHeaders {
		headers[k] = v
	}

	// 添加缓存验证
	if cachedChecksum != "" {
		headers["If-None-Match"] = fmt.Sprintf(`"%s"`, cachedChecksum)
	}

	// 创建HTTP请求
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return RemoteManagedSettingsFetchResult{
			Success: false,
			Error:   err.Error(),
		}
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	client := &http.Client{
		Timeout: time.Duration(SettingsTimeoutMs) * time.Millisecond,
	}

	resp, err := client.Do(req)
	if err != nil {
		return RemoteManagedSettingsFetchResult{
			Success: false,
			Error:   fmt.Sprintf("Cannot connect to server: %v", err),
		}
	}
	defer resp.Body.Close()

	// 处理304 Not Modified
	if resp.StatusCode == 304 {
		return RemoteManagedSettingsFetchResult{
			Success:  true,
			Settings: nil, // 表示缓存仍然有效
			Checksum: cachedChecksum,
		}
	}

	// 处理204/404 - 没有设置
	if resp.StatusCode == 204 || resp.StatusCode == 404 {
		return RemoteManagedSettingsFetchResult{
			Success:  true,
			Settings: make(map[string]interface{}),
			Checksum: "",
		}
	}

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return RemoteManagedSettingsFetchResult{
			Success: false,
			Error:   fmt.Sprintf("Failed to read response: %v", err),
		}
	}

	// 解析响应
	var result RemoteManagedSettingsResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return RemoteManagedSettingsFetchResult{
			Success: false,
			Error:   fmt.Sprintf("Invalid response format: %v", err),
		}
	}

	return RemoteManagedSettingsFetchResult{
		Success:  true,
		Settings: result.Settings,
		Checksum: result.Checksum,
	}
}

// fetchWithRetry 带重试逻辑获取设置
func fetchWithRetry(cachedChecksum string) RemoteManagedSettingsFetchResult {
	var lastResult RemoteManagedSettingsFetchResult

	for attempt := 1; attempt <= DefaultMaxRetries+1; attempt++ {
		lastResult = fetchRemoteManagedSettings(cachedChecksum)

		if lastResult.Success {
			return lastResult
		}

		if lastResult.SkipRetry {
			return lastResult
		}

		if attempt > DefaultMaxRetries {
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
	return delay
}

// getSettingsPath 获取设置文件路径
func getSettingsPath() string {
	configHome := os.Getenv("CLAUDE_CONFIG_HOME")
	if configHome == "" {
		home, _ := os.UserHomeDir()
		configHome = filepath.Join(home, ".config", "claude")
	}
	return filepath.Join(configHome, "remote-settings.json")
}

// saveSettings 保存设置到文件
func saveSettings(settings map[string]interface{}) error {
	path := getSettingsPath()

	// 确保目录存在
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}

// loadSettingsFromFile 从文件加载设置
func loadSettingsFromFile() map[string]interface{} {
	path := getSettingsPath()

	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	var settings map[string]interface{}
	if err := json.Unmarshal(data, &settings); err != nil {
		return nil
	}

	return settings
}

// ClearRemoteManagedSettingsCache 清除所有远程托管设置缓存
func ClearRemoteManagedSettingsCache() {
	remoteSettingsService.mu.Lock()
	defer remoteSettingsService.mu.Unlock()

	// 停止后台轮询
	stopBackgroundPollingInternal()

	// 清除会话缓存
	remoteSettingsService.sessionCache = nil

	// 重置资格缓存
	remoteSettingsService.cachedEligible = nil

	// 删除设置文件
	path := getSettingsPath()
	os.Remove(path)
}

// LoadRemoteManagedSettings 加载远程托管设置
func LoadRemoteManagedSettings() error {
	remoteSettingsService.mu.Lock()
	defer remoteSettingsService.mu.Unlock()

	// 检查资格
	if !isRemoteManagedSettingsEligible() {
		close(remoteSettingsService.loadingComplete)
		return nil
	}

	// 从文件加载缓存
	cachedSettings := loadSettingsFromFile()

	// 计算本地校验和
	var cachedChecksum string
	if cachedSettings != nil {
		remoteSettingsService.sessionCache = cachedSettings
		cachedChecksum = ComputeChecksumFromSettings(cachedSettings)
	}

	// 获取设置
	result := fetchWithRetry(cachedChecksum)

	if !result.Success {
		// 失败时使用缓存
		if cachedSettings != nil {
			remoteSettingsService.sessionCache = cachedSettings
		}
		close(remoteSettingsService.loadingComplete)
		return nil
	}

	// 处理304 Not Modified
	if result.Settings == nil && cachedSettings != nil {
		close(remoteSettingsService.loadingComplete)
		return nil
	}

	// 处理新设置
	newSettings := result.Settings
	if newSettings == nil {
		newSettings = make(map[string]interface{})
	}

	// 检查是否有内容
	hasContent := len(newSettings) > 0

	if hasContent {
		// 安全检查
		if !checkManagedSettingsSecurity(cachedSettings, newSettings) {
			// 用户拒绝，使用缓存
			if cachedSettings != nil {
				remoteSettingsService.sessionCache = cachedSettings
			}
			close(remoteSettingsService.loadingComplete)
			return nil
		}

		remoteSettingsService.sessionCache = newSettings
		saveSettings(newSettings)
	} else {
		// 空设置 - 删除缓存文件
		remoteSettingsService.sessionCache = newSettings
		os.Remove(getSettingsPath())
	}

	// 启动后台轮询
	startBackgroundPollingInternal()

	close(remoteSettingsService.loadingComplete)
	return nil
}

// RefreshRemoteManagedSettings 刷新远程托管设置
func RefreshRemoteManagedSettings() error {
	ClearRemoteManagedSettingsCache()

	if !isRemoteManagedSettingsEligible() {
		return nil
	}

	return LoadRemoteManagedSettings()
}

// GetRemoteManagedSettingsFromCache 从缓存获取设置
func GetRemoteManagedSettingsFromCache() map[string]interface{} {
	remoteSettingsService.mu.RLock()
	defer remoteSettingsService.mu.RUnlock()
	return remoteSettingsService.sessionCache
}

// checkManagedSettingsSecurity 检查托管设置的安全性
func checkManagedSettingsSecurity(cached, new map[string]interface{}) bool {
	// TODO: 实现安全检查逻辑
	// 检查危险的设置更改
	return true
}

// startBackgroundPollingInternal 启动后台轮询
func startBackgroundPollingInternal() {
	if remoteSettingsService.pollingIntervalId != nil {
		return
	}

	if !isRemoteManagedSettingsEligible() {
		return
	}

	remoteSettingsService.pollingIntervalId = time.NewTimer(time.Duration(PollingIntervalMs) * time.Millisecond)
	go func() {
		<-remoteSettingsService.pollingIntervalId.C
		pollRemoteSettings()
	}()
}

// stopBackgroundPollingInternal 停止后台轮询
func stopBackgroundPollingInternal() {
	if remoteSettingsService.pollingIntervalId != nil {
		remoteSettingsService.pollingIntervalId.Stop()
		remoteSettingsService.pollingIntervalId = nil
	}
}

// pollRemoteSettings 轮询远程设置
func pollRemoteSettings() {
	if !isRemoteManagedSettingsEligible() {
		return
	}

	// 获取当前缓存以进行比较
	remoteSettingsService.mu.RLock()
	prevCache := remoteSettingsService.sessionCache
	remoteSettingsService.mu.RUnlock()

	var cachedChecksum string
	if prevCache != nil {
		cachedChecksum = ComputeChecksumFromSettings(prevCache)
	}

	// 获取新设置
	result := fetchWithRetry(cachedChecksum)

	remoteSettingsService.mu.Lock()
	defer remoteSettingsService.mu.Unlock()

	if result.Success && result.Settings != nil {
		// 检查设置是否实际更改
		newChecksum := result.Checksum
		if newChecksum != cachedChecksum {
			remoteSettingsService.sessionCache = result.Settings
			saveSettings(result.Settings)
		}
	}

	// 重新启动轮询
	startBackgroundPollingInternal()
}

// 辅助函数

func checkAndRefreshOAuthTokenIfNeeded() {
	// TODO: 实现OAuth令牌刷新
}

func getClaudeCodeUserAgent() string {
	return "Claude-Code-Go/0.1.0"
}
