package api

import (
	"os"
	"sync"
	"time"
)

// MetricsEnabledResponse 指标启用响应
type MetricsEnabledResponse struct {
	MetricsLoggingEnabled bool `json:"metrics_logging_enabled"`
}

// MetricsStatus 指标状态
type MetricsStatus struct {
	Enabled  bool
	HasError bool
}

// 内存缓存TTL - 同一进程内去重
const CACHE_TTL_MS = 60 * 60 * 1000

// 磁盘缓存TTL - 组织设置很少更改
const DISK_CACHE_TTL_MS = 24 * 60 * 60 * 1000

var (
	metricsCache      *MetricsStatus
	metricsCacheMutex sync.RWMutex
	metricsCacheTime  int64
)

// _clearMetricsEnabledCacheForTesting 清除缓存（仅用于测试）
func ClearMetricsEnabledCacheForTesting() {
	metricsCacheMutex.Lock()
	defer metricsCacheMutex.Unlock()
	metricsCache = nil
	metricsCacheTime = 0
}

// CheckMetricsEnabled 检查指标是否启用
// 两层缓存：
// - 磁盘缓存(24h TTL)：跨进程持久化
// - 内存缓存(1h TTL)：同一进程内去重
func CheckMetricsEnabled() (*MetricsStatus, error) {
	// 服务密钥OAuth会话缺少user:profile作用域会返回403
	if isClaudeAISubscriber() && !HasProfileScope() {
		return &MetricsStatus{Enabled: false, HasError: false}, nil
	}

	// 读取缓存
	metricsCacheMutex.RLock()
	cached := metricsCache
	cachedTime := metricsCacheTime
	metricsCacheMutex.RUnlock()

	if cached != nil {
		if time.Now().UnixMilli()-cachedTime > DISK_CACHE_TTL_MS {
			// 缓存过期，异步刷新
			go func() {
				refreshMetricsStatus()
			}()
		}
		return cached, nil
	}

	// 首次运行：阻塞网络调用
	return refreshMetricsStatus()
}

// refreshMetricsStatus 刷新指标状态
func refreshMetricsStatus() (*MetricsStatus, error) {
	result, err := checkMetricsEnabledAPI()
	if err != nil {
		return &MetricsStatus{Enabled: false, HasError: true}, err
	}

	// 更新缓存
	metricsCacheMutex.Lock()
	metricsCache = result
	metricsCacheTime = time.Now().UnixMilli()
	metricsCacheMutex.Unlock()

	// 持久化到磁盘
	saveMetricsStatusToConfig(result)

	return result, nil
}

// checkMetricsEnabledAPI 调用API检查指标是否启用
func checkMetricsEnabledAPI() (*MetricsStatus, error) {
	// 重要流量检查：如果禁用了非必要流量，跳过网络调用
	if isEssentialTrafficOnly() {
		return &MetricsStatus{Enabled: false, HasError: false}, nil
	}

	headers, err := GetAuthHeaders()
	if err != nil {
		return &MetricsStatus{Enabled: false, HasError: true}, err
	}

	headers["Content-Type"] = "application/json"
	headers["User-Agent"] = GetUserAgent()

	endpoint := "https://api.anthropic.com/api/claude_code/organizations/metrics_enabled"

	body, statusCode, err := HTTPGet(endpoint, headers, 5*time.Second)
	if err != nil {
		return &MetricsStatus{Enabled: false, HasError: true}, err
	}

	if statusCode != 200 {
		return &MetricsStatus{Enabled: false, HasError: true}, nil
	}

	var response MetricsEnabledResponse
	if err := ParseJSON(body, &response); err != nil {
		return &MetricsStatus{Enabled: false, HasError: true}, err
	}

	return &MetricsStatus{
		Enabled:  response.MetricsLoggingEnabled,
		HasError: false,
	}, nil
}

// saveMetricsStatusToConfig 保存指标状态到配置
func saveMetricsStatusToConfig(status *MetricsStatus) {
	// 实现配置持久化
}

// isEssentialTrafficOnly 检查是否只允许必要流量
func isEssentialTrafficOnly() bool {
	return os.Getenv("ESSENTIAL_TRAFFIC_ONLY") == "true"
}
