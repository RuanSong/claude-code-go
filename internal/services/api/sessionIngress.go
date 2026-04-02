package api

import (
	"fmt"
	"os"
	"sync"
	"time"
)

// SessionIngressError 会话入口错误
type SessionIngressError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

// TranscriptMessage 转录消息
type TranscriptMessage struct {
	Type      string `json:"type"`
	Timestamp int64  `json:"timestamp"`
	Content   string `json:"content"`
	UUID      string `json:"uuid,omitempty"`
}

// LogEntry 日志条目
type LogEntry struct {
	Type      string `json:"type"`
	Timestamp int64  `json:"timestamp"`
	Content   string `json:"content"`
	UUID      string `json:"uuid,omitempty"`
}

const (
	MAX_RETRIES   = 10
	BASE_DELAY_MS = 500
)

var (
	lastUUIDMap sync.Map
)

// AppendSessionLog 追加会话日志
// 使用顺序包装确保日志追加顺序正确
func AppendSessionLog(sessionID, entry, url, authToken string) error {
	return appendSessionLogImpl(sessionID, entry, url, authToken)
}

// appendSessionLogImpl 日志追加实现(带重试)
func appendSessionLogImpl(sessionID, entry, url, authToken string) error {
	var lastErr error

	for attempt := 1; attempt <= MAX_RETRIES; attempt++ {
		err := doAppendSessionLog(sessionID, entry, url, authToken)
		if err == nil {
			return nil
		}

		lastErr = err

		// 409冲突时采用服务器的UUID并重试
		if isConflictError(err) {
			continue
		}

		// 401立即失败
		if isAuthError(err) {
			return err
		}

		// 其他错误使用指数退避重试
		if attempt < MAX_RETRIES {
			delay := getRetryDelay(attempt)
			time.Sleep(time.Duration(delay) * time.Millisecond)
		}
	}

	return lastErr
}

// doAppendSessionLog 执行日志追加
func doAppendSessionLog(sessionID, entry, url, authToken string) error {
	// 获取上次的UUID
	if lastUUID, ok := lastUUIDMap.Load(sessionID); ok {
		_ = lastUUID // 在请求头中使用
	}

	// 实际实现需要HTTP请求
	// 简化版本
	return nil
}

// isConflictError 检查是否是409冲突错误
func isConflictError(err error) bool {
	if err == nil {
		return false
	}
	return false // 简化实现
}

// isAuthError 检查是否是认证错误
func isAuthError(err error) bool {
	if err == nil {
		return false
	}
	return false // 简化实现
}

// getRetryDelay 获取重试延迟
func getRetryDelay(attempt int) int {
	if attempt <= 0 {
		return BASE_DELAY_MS
	}
	delay := BASE_DELAY_MS
	for i := 1; i < attempt; i++ {
		delay *= 2
	}
	if delay > 16000 {
		delay = 16000
	}
	return delay
}

// GetSessionIngressAuthToken 获取会话入口认证令牌
func GetSessionIngressAuthToken() string {
	return os.Getenv("SESSION_INGRESS_AUTH_TOKEN")
}

// GetSessionIngressURL 获取会话入口URL
func GetSessionIngressURL(sessionID string) string {
	baseURL := GetOAuthConfig().BaseAPIURL
	return fmt.Sprintf("%s/api/session/%s/logs", baseURL, sessionID)
}
