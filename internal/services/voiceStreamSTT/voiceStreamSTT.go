package voiceStreamSTT

import (
	"os"
	"sync"
	"time"
)

const (
	VOICE_STREAM_PATH     = "/api/ws/speech_to_text/voice_stream"
	KEEPALIVE_INTERVAL_MS = 8000
	FINALIZE_TIMEOUT_MS   = 5000
	NO_DATA_TIMEOUT_MS    = 1500
)

// VoiceStreamCallbacks 语音流回调接口
type VoiceStreamCallbacks interface {
	OnTranscript(text string, isFinal bool)
	OnError(error string, opts ...ErrorOpts)
	OnClose()
	OnReady(connection *VoiceStreamConnection)
}

// ErrorOpts 错误选项
type ErrorOpts struct {
	Fatal bool
}

// VoiceStreamConnection 语音流连接
type VoiceStreamConnection struct {
	mu         sync.RWMutex
	connected  bool
	finalized  bool
	finalizing bool
	ws         WebSocketClient
}

// WebSocketClient WebSocket客户端接口
type WebSocketClient interface {
	Send(data []byte) error
	Close() error
	IsConnected() bool
}

// FinalizeSource 结束源
type FinalizeSource string

const (
	FinalizePostCloseStream FinalizeSource = "post_closestream_endpoint"
	FinalizeNoDataTimeout   FinalizeSource = "no_data_timeout"
	FinalizeSafetyTimeout   FinalizeSource = "safety_timeout"
	FinalizeWSClose         FinalizeSource = "ws_close"
	FinalizeWSAlreadyClosed FinalizeSource = "ws_already_closed"
)

// VoiceStreamMessage 语音流消息
type VoiceStreamMessage struct {
	Type        string `json:"type"`
	Data        string `json:"data,omitempty"`
	ErrorCode   string `json:"error_code,omitempty"`
	Description string `json:"description,omitempty"`
	Message     string `json:"message,omitempty"`
}

// IsVoiceStreamAvailable 检查语音流是否可用
func IsVoiceStreamAvailable() bool {
	if !IsAnthropicAuthEnabled() {
		return false
	}

	tokens := GetClaudeAIOAuthTokens()
	return tokens != nil && tokens.AccessToken != ""
}

// IsAnthropicAuthEnabled 检查是否启用了Anthropic认证
func IsAnthropicAuthEnabled() bool {
	return os.Getenv("ANTHROPIC_AUTH_ENABLED") == "true"
}

// GetClaudeAIOAuthTokens 获取Claude AI OAuth令牌
func GetClaudeAIOAuthTokens() *OAUTHTokens {
	accessToken := os.Getenv("CLAUDEAI_ACCESS_TOKEN")
	if accessToken == "" {
		return nil
	}
	return &OAUTHTokens{
		AccessToken: accessToken,
		ExpiresAt:   time.Now().Add(time.Hour),
	}
}

// OAUTHTokens OAuth令牌
type OAUTHTokens struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
}

// ConnectVoiceStream 连接语音流
func ConnectVoiceStream(callbacks VoiceStreamCallbacks, options *VoiceStreamOptions) (*VoiceStreamConnection, error) {
	// 确保OAuth令牌是新鲜的
	if err := CheckAndRefreshOAuthTokenIfNeeded(); err != nil {
		return nil, err
	}

	tokens := GetClaudeAIOAuthTokens()
	if tokens == nil || tokens.AccessToken == "" {
		return nil, nil
	}

	// 构建WebSocket URL
	baseURL := GetVoiceStreamBaseURL()
	params := BuildVoiceStreamParams(options)
	url := baseURL + VOICE_STREAM_PATH + "?" + params

	// 创建连接
	conn := &VoiceStreamConnection{
		connected: false,
	}

	// 启动WebSocket连接
	if err := conn.connect(url, tokens.AccessToken, callbacks); err != nil {
		return nil, err
	}

	return conn, nil
}

// VoiceStreamOptions 语音流选项
type VoiceStreamOptions struct {
	Language string
	Keyterms []string
}

// GetVoiceStreamBaseURL 获取语音流基础URL
func GetVoiceStreamBaseURL() string {
	if override := os.Getenv("VOICE_STREAM_BASE_URL"); override != "" {
		return override
	}

	// 从OAuth配置替换协议
	baseAPIURL := GetOAuthConfig().BaseAPIURL
	// 将 https:// 替换为 wss://
	if len(baseAPIURL) >= 5 && baseAPIURL[:5] == "https" {
		return "wss://" + baseAPIURL[8:]
	}
	if len(baseAPIURL) >= 4 && baseAPIURL[:4] == "http" {
		return "ws://" + baseAPIURL[7:]
	}
	return baseAPIURL
}

// GetOAuthConfig 获取OAuth配置
func GetOAuthConfig() OAuthConfig {
	return OAuthConfig{
		BaseAPIURL: os.Getenv("OAUTH_BASE_API_URL"),
	}
}

// OAuthConfig OAuth配置
type OAuthConfig struct {
	BaseAPIURL string
}

// BuildVoiceStreamParams 构建语音流参数
func BuildVoiceStreamParams(options *VoiceStreamOptions) string {
	params := "encoding=linear16&sample_rate=16000&channels=1&endpointing_ms=300&utterance_end_ms=1000&language=en"

	if options != nil && options.Language != "" {
		params += "&language=" + options.Language
	}

	// 添加关键词
	if options != nil && len(options.Keyterms) > 0 {
		for _, term := range options.Keyterms {
			params += "&keyterms=" + term
		}
	}

	return params
}

// CheckAndRefreshOAuthTokenIfNeeded 检查并刷新OAuth令牌
func CheckAndRefreshOAuthTokenIfNeeded() error {
	// 简化实现
	return nil
}

func (c *VoiceStreamConnection) connect(url, token string, callbacks VoiceStreamCallbacks) error {
	c.mu.Lock()
	c.connected = true
	c.mu.Unlock()

	// 通知已连接
	go callbacks.OnReady(c)

	return nil
}

// Send 发送音频数据
func (c *VoiceStreamConnection) Send(audioChunk []byte) error {
	c.mu.RLock()
	if !c.connected || c.finalized {
		c.mu.RUnlock()
		return nil
	}
	c.mu.RUnlock()

	// 实际通过WebSocket发送
	return nil
}

// Finalize 结束语音流
func (c *VoiceStreamConnection) Finalize() (FinalizeSource, error) {
	c.mu.Lock()
	if c.finalizing || c.finalized {
		c.mu.Unlock()
		return FinalizeWSAlreadyClosed, nil
	}
	c.finalizing = true
	c.finalized = true
	c.mu.Unlock()

	// 发送CloseStream消息
	// 实现结束逻辑

	return FinalizeWSClose, nil
}

// Close 关闭连接
func (c *VoiceStreamConnection) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.connected {
		c.connected = false
		c.finalized = true
	}
}

// IsConnected 检查是否已连接
func (c *VoiceStreamConnection) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connected
}
