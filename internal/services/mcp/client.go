package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"
)

const (
	DefaultMcpToolTimeoutMs = 100000000
	McpRequestTimeoutMs     = 60000
	McpAuthCacheTtlMs       = 15 * 60 * 1000
)

type TransportType string

const (
	TransportStdio         TransportType = "stdio"
	TransportSSE           TransportType = "sse"
	TransportHTTP          TransportType = "http"
	TransportWS            TransportType = "ws"
	TransportSSEIde        TransportType = "sse-ide"
	TransportWsIde         TransportType = "ws-ide"
	TransportSDK           TransportType = "sdk"
	TransportClaudeAiProxy TransportType = "claudeai-proxy"
)

type ConfigScope string

const (
	ScopeLocal      ConfigScope = "local"
	ScopeUser       ConfigScope = "user"
	ScopeProject    ConfigScope = "project"
	ScopeDynamic    ConfigScope = "dynamic"
	ScopeEnterprise ConfigScope = "enterprise"
	ScopeClaudeAI   ConfigScope = "claudeai"
	ScopeManaged    ConfigScope = "managed"
)

type ConnectionState string

const (
	StateConnected ConnectionState = "connected"
	StateFailed    ConnectionState = "failed"
	StateNeedsAuth ConnectionState = "needs-auth"
	StatePending   ConnectionState = "pending"
	StateDisabled  ConnectionState = "disabled"
)

type ClientConfig struct {
	Name       string            `json:"name"`
	Version    string            `json:"version"`
	HTTPClient interface{}       `json:"-"`
	Timeout    time.Duration     `json:"timeout"`
	Transport  TransportType     `json:"transport"`
	Command    string            `json:"command,omitempty"`
	Args       []string          `json:"args,omitempty"`
	Env        map[string]string `json:"env,omitempty"`
	URL        string            `json:"url,omitempty"`
	Headers    map[string]string `json:"headers,omitempty"`
	Scope      ConfigScope       `json:"scope,omitempty"`
}

type ServerConfig struct {
	Name      string            `json:"name"`
	Transport TransportType     `json:"transport"`
	Command   string            `json:"command,omitempty"`
	Args      []string          `json:"args,omitempty"`
	Env       map[string]string `json:"env,omitempty"`
	URL       string            `json:"url,omitempty"`
	Headers   map[string]string `json:"headers,omitempty"`
	Scope     ConfigScope       `json:"scope,omitempty"`
}

type MCPTool struct {
	Name         string                 `json:"name"`
	Description  string                 `json:"description,omitempty"`
	InputSchema  map[string]interface{} `json:"input_schema,omitempty"`
	OutputSchema map[string]interface{} `json:"output_schema,omitempty"`
}

type ListToolsResult struct {
	Tools []MCPTool `json:"tools"`
}

type ToolCallRequest struct {
	Method string                 `json:"method"`
	Params map[string]interface{} `json:"params"`
}

type ToolCallResponse struct {
	Success bool                   `json:"success"`
	Result  map[string]interface{} `json:"result,omitempty"`
	Error   *ErrorResponse         `json:"error,omitempty"`
}

type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type ServerStatus struct {
	Name      string          `json:"name"`
	State     ConnectionState `json:"state"`
	Error     string          `json:"error,omitempty"`
	LastError time.Time       `json:"lastError,omitempty"`
}

type OAuthToken struct {
	AccessToken  string    `json:"access_token"`
	TokenType    string    `json:"token_type"`
	ExpiresIn    int       `json:"expires_in"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	Scope        string    `json:"scope,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

type Resource struct {
	URI         string `json:"uri"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	MimeType    string `json:"mimeType,omitempty"`
}

type ResourceList struct {
	Resources []Resource `json:"resources"`
}

type NotificationHandler func(method string, params map[string]interface{})

type Client struct {
	config    *ClientConfig
	transport Transport
	tools     map[string]MCPTool
	resources map[string]Resource
	mu        sync.RWMutex
	state     ConnectionState
	lastError error
	handlers  map[string][]NotificationHandler
}

func NewClient(config *ClientConfig) *Client {
	if config.HTTPClient == nil {
		config.HTTPClient = &http.Client{
			Timeout: 30 * time.Second,
		}
	}
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	return &Client{
		config:    config,
		tools:     make(map[string]MCPTool),
		resources: make(map[string]Resource),
		state:     StatePending,
		handlers:  make(map[string][]NotificationHandler),
	}
}

func (c *Client) Connect(ctx context.Context, serverURL string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	_, err := url.ParseRequestURI(serverURL)
	if err != nil {
		return fmt.Errorf("invalid server URL: %w", err)
	}

	c.state = StateConnected
	return nil
}

func (c *Client) ListTools(ctx context.Context) (*ListToolsResult, error) {
	c.mu.RLock()
	tools := make([]MCPTool, 0, len(c.tools))
	for _, tool := range c.tools {
		tools = append(tools, tool)
	}
	c.mu.RUnlock()

	return &ListToolsResult{Tools: tools}, nil
}

func (c *Client) CallTool(ctx context.Context, name string, args map[string]interface{}) (*ToolCallResponse, error) {
	c.mu.RLock()
	tool, exists := c.tools[name]
	c.mu.RUnlock()

	if !exists {
		return &ToolCallResponse{
			Success: false,
			Error: &ErrorResponse{
				Code:    -32601,
				Message: fmt.Sprintf("Tool not found: %s", name),
			},
		}, nil
	}

	result := map[string]interface{}{
		"tool":      tool.Name,
		"input":     args,
		"timestamp": time.Now().Unix(),
	}

	return &ToolCallResponse{
		Success: true,
		Result:  result,
	}, nil
}

func (c *Client) RegisterTool(tool MCPTool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.tools[tool.Name] = tool
}

func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.tools = make(map[string]MCPTool)
	c.resources = make(map[string]Resource)
	c.state = StateDisabled
	return nil
}

func (c *Client) GetState() ConnectionState {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.state
}

func (c *Client) SetState(state ConnectionState) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.state = state
}

func (c *Client) OnNotification(method string, handler NotificationHandler) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.handlers[method] = append(c.handlers[method], handler)
}

func (c *Client) HandleNotification(method string, params map[string]interface{}) {
	c.mu.RLock()
	handlers := c.handlers[method]
	c.mu.RUnlock()

	for _, handler := range handlers {
		handler(method, params)
	}
}

type ServerManager struct {
	clients map[string]*Client
	configs map[string]*ServerConfig
	mu      sync.RWMutex
}

func NewServerManager() *ServerManager {
	return &ServerManager{
		clients: make(map[string]*Client),
		configs: make(map[string]*ServerConfig),
	}
}

func (m *ServerManager) AddServer(name string, config *ClientConfig) *Client {
	m.mu.Lock()
	defer m.mu.Unlock()

	client := NewClient(config)
	m.clients[name] = client
	return client
}

func (m *ServerManager) GetServer(name string) (*Client, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	client, ok := m.clients[name]
	return client, ok
}

func (m *ServerManager) RemoveServer(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.clients, name)
	delete(m.configs, name)
}

func (m *ServerManager) ListServers() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	servers := make([]string, 0, len(m.clients))
	for name := range m.clients {
		servers = append(servers, name)
	}
	return servers
}

func (m *ServerManager) GetServerStatus(name string) *ServerStatus {
	m.mu.RLock()
	client, exists := m.clients[name]
	config, configExists := m.configs[name]
	m.mu.RUnlock()

	if !exists {
		return &ServerStatus{
			Name:  name,
			State: StateDisabled,
			Error: "Server not found",
		}
	}

	status := &ServerStatus{
		Name:  name,
		State: client.GetState(),
	}

	if configExists {
		status.Error = config.URL
	}

	return status
}

type MCPProtocol struct {
	manager       *ServerManager
	auth          *AuthHandler
	resourceCache map[string]*ResourceList
	mu            sync.RWMutex
}

func NewMCPProtocol() *MCPProtocol {
	return &MCPProtocol{
		manager:       NewServerManager(),
		auth:          NewAuthHandler(),
		resourceCache: make(map[string]*ResourceList),
	}
}

func (p *MCPProtocol) Initialize(ctx context.Context, servers []map[string]interface{}) error {
	for _, server := range servers {
		name, ok := server["name"].(string)
		if !ok {
			continue
		}

		config := &ClientConfig{
			Name: name,
		}

		if version, ok := server["version"].(string); ok {
			config.Version = version
		}

		p.manager.AddServer(name, config)

		serverConfig := &ServerConfig{
			Name: name,
		}

		if transport, ok := server["transport"].(string); ok {
			serverConfig.Transport = TransportType(transport)
		}

		if cmd, ok := server["command"].(string); ok {
			serverConfig.Command = cmd
		}

		if args, ok := server["args"].([]string); ok {
			serverConfig.Args = args
		}

		if urlStr, ok := server["url"].(string); ok {
			serverConfig.URL = urlStr
		}

		p.manager.configs[name] = serverConfig
	}

	return nil
}

func (p *MCPProtocol) CallTool(ctx context.Context, serverName, toolName string, args map[string]interface{}) (*ToolCallResponse, error) {
	client, ok := p.manager.GetServer(serverName)
	if !ok {
		return &ToolCallResponse{
			Success: false,
			Error: &ErrorResponse{
				Code:    -32000,
				Message: fmt.Sprintf("Server not found: %s", serverName),
			},
		}, nil
	}

	return client.CallTool(ctx, toolName, args)
}

func (p *MCPProtocol) ListTools(ctx context.Context, serverName string) (*ListToolsResult, error) {
	client, ok := p.manager.GetServer(serverName)
	if !ok {
		return nil, fmt.Errorf("server not found: %s", serverName)
	}

	return client.ListTools(ctx)
}

func (p *MCPProtocol) ListResources(ctx context.Context, serverName string) (*ResourceList, error) {
	p.mu.RLock()
	if cached, ok := p.resourceCache[serverName]; ok {
		p.mu.RUnlock()
		return cached, nil
	}
	p.mu.RUnlock()

	_, ok := p.manager.GetServer(serverName)
	if !ok {
		return nil, fmt.Errorf("server not found: %s", serverName)
	}

	p.mu.Lock()
	p.resourceCache[serverName] = &ResourceList{Resources: []Resource{}}
	p.mu.Unlock()

	return &ResourceList{Resources: []Resource{}}, nil
}

func (p *MCPProtocol) ReadResource(ctx context.Context, serverName, uri string) (map[string]interface{}, error) {
	return map[string]interface{}{
		"uri":      uri,
		"mimeType": "text/plain",
		"contents": []map[string]interface{}{{
			"type": "text",
			"text": "Resource content placeholder",
		}},
	}, nil
}

func (p *MCPProtocol) MarshalJSON() ([]byte, error) {
	type Alias MCPProtocol
	return json.Marshal(&struct {
		Type    string   `json:"type"`
		Servers []string `json:"servers"`
		*Alias
	}{
		Type:    "MCPProtocol",
		Servers: p.manager.ListServers(),
		Alias:   (*Alias)(p),
	})
}

type AuthHandler struct {
	mu        sync.RWMutex
	tokens    map[string]*OAuthToken
	authCache map[string]time.Time
}

func NewAuthHandler() *AuthHandler {
	return &AuthHandler{
		tokens:    make(map[string]*OAuthToken),
		authCache: make(map[string]time.Time),
	}
}

func (h *AuthHandler) StoreToken(serverName string, token *OAuthToken) {
	h.mu.Lock()
	defer h.mu.Unlock()
	token.CreatedAt = time.Now()
	h.tokens[serverName] = token
}

func (h *AuthHandler) GetToken(serverName string) (*OAuthToken, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	token, ok := h.tokens[serverName]
	return token, ok
}

func (h *AuthHandler) IsTokenExpired(serverName string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	token, ok := h.tokens[serverName]
	if !ok {
		return true
	}

	expiry := token.CreatedAt.Add(time.Duration(token.ExpiresIn) * time.Second)
	return time.Now().After(expiry)
}

func (h *AuthHandler) ClearToken(serverName string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.tokens, serverName)
}

func (h *AuthHandler) IsAuthCached(serverName string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if cachedAt, ok := h.authCache[serverName]; ok {
		return time.Since(cachedAt) < McpAuthCacheTtlMs
	}
	return false
}

func (h *AuthHandler) SetAuthCached(serverName string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.authCache[serverName] = time.Now()
}

type ToolResponse struct {
	Content []map[string]interface{} `json:"content"`
	IsError bool                     `json:"is_error,omitempty"`
}

type ContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
	Data string `json:"data,omitempty"`
	URI  string `json:"uri,omitempty"`
}

type MCPToolResult struct {
	Content []ContentBlock `json:"content"`
	IsError bool           `json:"is_error,omitempty"`
	Meta    *ResultMeta    `json:"_meta,omitempty"`
}

type ResultMeta struct {
	Elapsed  int64  `json:"Elapsed,omitempty"`
	Provider string `json:"provider,omitempty"`
}

type ProgressCallback func(progress float64, message string)

type ToolCallOptions struct {
	Timeout    time.Duration
	OnProgress ProgressCallback
	Signal     <-chan struct{}
}

func DefaultToolCallOptions() *ToolCallOptions {
	return &ToolCallOptions{
		Timeout: DefaultMcpToolTimeoutMs * time.Millisecond,
	}
}
