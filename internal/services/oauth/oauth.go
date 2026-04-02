package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// Config OAuth 2.0 配置
// 对应 TypeScript: src/services/oauth/ 配置结构
// 存储OAuth认证所需的客户端信息和端点
type Config struct {
	ClientID     string   // OAuth应用客户端ID
	ClientSecret string   // OAuth应用客户端密钥
	AuthURL      string   // 授权端点URL
	TokenURL     string   // 令牌端点URL
	RedirectURI  string   // 回调URI
	Scopes       []string // 请求的权限范围
}

// Token OAuth访问令牌
// 对应 TypeScript: OAuth令牌响应格式
type Token struct {
	AccessToken  string    `json:"access_token"`            // 访问令牌
	TokenType    string    `json:"token_type"`              // 令牌类型（通常为Bearer）
	ExpiresIn    int       `json:"expires_in"`              // 有效期（秒）
	RefreshToken string    `json:"refresh_token,omitempty"` // 刷新令牌
	Scope        string    `json:"scope,omitempty"`         // 授予的权限范围
	CreatedAt    time.Time `json:"created_at"`              // 令牌创建时间
}

// OAuth OAuth 2.0 客户端
// 对应 TypeScript: OAuthService
// 提供OAuth授权码流程和令牌管理功能
type OAuth struct {
	config     *Config
	token      *Token
	mu         sync.RWMutex
	httpClient *http.Client
}

// NewOAuth 创建新的OAuth客户端
// 对应 TypeScript: OAuthService构造函数
// 使用提供的配置初始化客户端
func NewOAuth(config *Config) *OAuth {
	return &OAuth{
		config: config,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// BuildAuthURL 构建OAuth授权URL
// 对应 TypeScript: buildAuthUrl()
// 生成带有PKCE参数的授权请求URL
func (o *OAuth) BuildAuthURL(state string) (string, error) {
	params := url.Values{}
	params.Set("client_id", o.config.ClientID)
	params.Set("redirect_uri", o.config.RedirectURI)
	params.Set("response_type", "code")
	params.Set("state", state)
	if len(o.config.Scopes) > 0 {
		params.Set("scope", strings.Join(o.config.Scopes, " "))
	}

	authURL, err := url.Parse(o.config.AuthURL)
	if err != nil {
		return "", fmt.Errorf("parse auth URL: %w", err)
	}
	authURL.RawQuery = params.Encode()

	return authURL.String(), nil
}

// ExchangeCode 交换授权码为令牌
// 对应 TypeScript: exchangeCodeForTokens()
// 使用授权码向令牌端点请求访问令牌
func (o *OAuth) ExchangeCode(ctx context.Context, code string) (*Token, error) {
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("client_id", o.config.ClientID)
	data.Set("client_secret", o.config.ClientSecret)
	data.Set("code", code)
	data.Set("redirect_uri", o.config.RedirectURI)

	req, err := http.NewRequestWithContext(ctx, "POST", o.config.TokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := o.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("exchange code: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token exchange failed: %d", resp.StatusCode)
	}

	var token Token
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return nil, fmt.Errorf("decode token: %w", err)
	}

	token.CreatedAt = time.Now()
	o.mu.Lock()
	o.token = &token
	o.mu.Unlock()

	return &token, nil
}

// RefreshToken 刷新访问令牌
// 对应 TypeScript: refreshOAuthToken()
// 使用刷新令牌获取新的访问令牌
func (o *OAuth) RefreshToken(ctx context.Context, refreshToken string) (*Token, error) {
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("client_id", o.config.ClientID)
	data.Set("client_secret", o.config.ClientSecret)
	data.Set("refresh_token", refreshToken)

	req, err := http.NewRequestWithContext(ctx, "POST", o.config.TokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := o.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("refresh token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("refresh failed: %d", resp.StatusCode)
	}

	var token Token
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return nil, fmt.Errorf("decode token: %w", err)
	}

	token.CreatedAt = time.Now()
	o.mu.Lock()
	o.token = &token
	o.mu.Unlock()

	return &token, nil
}

// GetToken 获取当前存储的访问令牌
// 对应 TypeScript: 获取缓存的令牌
func (o *OAuth) GetToken() *Token {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return o.token
}

// IsTokenExpired 检查令牌是否已过期
// 对应 TypeScript: isOAuthTokenExpired()
// 比较当前时间与令牌过期时间
func (o *OAuth) IsTokenExpired() bool {
	o.mu.RLock()
	defer o.mu.RUnlock()

	if o.token == nil {
		return true
	}

	expiry := o.token.CreatedAt.Add(time.Duration(o.token.ExpiresIn) * time.Second)
	return time.Now().After(expiry)
}

// SetToken 设置访问令牌
// 对应 TypeScript: 存储令牌到缓存
func (o *OAuth) SetToken(token *Token) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.token = token
}

// ClearToken 清除存储的令牌
// 对应 TypeScript: 登出时清除令牌
func (o *OAuth) ClearToken() {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.token = nil
}

// TokenManager 多提供商令牌管理器
// 对应 TypeScript: 跨多个OAuth提供商管理令牌
// 存储和检索不同提供商的令牌
type TokenManager struct {
	mu     sync.RWMutex
	tokens map[string]*Token
}

// NewTokenManager 创建新的令牌管理器
// 初始化令牌存储映射
func NewTokenManager() *TokenManager {
	return &TokenManager{
		tokens: make(map[string]*Token),
	}
}

// Store 存储指定提供商的令牌
// 对应 TypeScript: storeOAuthAccountInfo()
func (m *TokenManager) Store(provider string, token *Token) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tokens[provider] = token
}

// Get 获取指定提供商的令牌
// 对应 TypeScript: 获取令牌
func (m *TokenManager) Get(provider string) (*Token, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	token, ok := m.tokens[provider]
	return token, ok
}

// Remove 移除指定提供商的令牌
// 对应 TypeScript: 清除提供商令牌
func (m *TokenManager) Remove(provider string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.tokens, provider)
}

// ListProviders 列出所有已存储令牌的提供商
// 对应 TypeScript: 获取所有OAuth账户信息
func (m *TokenManager) ListProviders() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	providers := make([]string, 0, len(m.tokens))
	for p := range m.tokens {
		providers = append(providers, p)
	}
	return providers
}
