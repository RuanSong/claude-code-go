package lsp

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"
)

// Client LSP客户端
// 对应 TypeScript: src/services/lsp/LSPClient.ts
// 与LSP服务器通信的客户端实现
type Client struct {
	conn         net.Conn
	name         string
	mu           sync.RWMutex
	capabilities ServerCapabilities
	ready        bool
}

// ServerCapabilities 服务器能力
// 对应 TypeScript: LSP Server Capabilities
// 声明服务器支持的功能
type ServerCapabilities struct {
	TextDocumentSync    int                 `json:"textDocumentSync"`              // 文本同步方式
	HoverProvider       bool                `json:"hoverProvider"`                 // 是否支持悬停提示
	CompletionProvider  *CompletionOptions  `json:"completionProvider,omitempty"`  // 自动完成选项
	DefinitionProvider  bool                `json:"definitionProvider"`            // 是否支持跳转定义
	ReferencesProvider  bool                `json:"referencesProvider"`            // 是否支持查找引用
	DiagnosticsProvider *DiagnosticsOptions `json:"diagnosticsProvider,omitempty"` // 诊断选项
}

// CompletionOptions 自动完成选项
type CompletionOptions struct {
	TriggerCharacters []string `json:"triggerCharacters,omitempty"` // 触发字符
	ResolveProvider   bool     `json:"resolveProvider,omitempty"`   // 是否支持resolve
}

// DiagnosticsOptions 诊断选项
type DiagnosticsOptions struct {
	Identifier string `json:"identifier,omitempty"` // 诊断标识符
}

// ClientCapabilities 客户端能力
// 对应 TypeScript: LSP Client Capabilities
// 声明客户端支持的功能
type ClientCapabilities struct {
	TextDocumentSync   int                `json:"textDocumentSync,omitempty"`   // 文本同步方式
	HoverProvider      bool               `json:"hoverProvider,omitempty"`      // 是否支持悬停提示
	CompletionProvider *CompletionOptions `json:"completionProvider,omitempty"` // 自动完成选项
	DefinitionProvider bool               `json:"definitionProvider,omitempty"` // 是否支持跳转定义
	ReferencesProvider bool               `json:"referencesProvider,omitempty"` // 是否支持查找引用
}

// LSPMessage LSP协议消息
// 对应 TypeScript: JSON-RPC格式的LSP消息
type LSPMessage struct {
	ID      interface{} `json:"id"`               // 消息ID，请求/响应配对
	Method  string      `json:"method"`           // 方法名
	Params  interface{} `json:"params,omitempty"` // 参数
	Result  interface{} `json:"result,omitempty"` // 结果
	Error   *LSPError   `json:"error,omitempty"`  // 错误
	JSONRPC string      `json:"jsonrpc"`          // JSON-RPC版本，固定为"2.0"
}

// LSPError LSP错误
// 对应 TypeScript: JSON-RPC错误格式
type LSPError struct {
	Code    int    `json:"code"`    // 错误代码
	Message string `json:"message"` // 错误消息
}

// TextDocumentItem 文本文档项
// 对应 TypeScript: textDocument/item
// 表示打开的文档
type TextDocumentItem struct {
	URI        string `json:"uri"`        // 文档URI
	LanguageID string `json:"languageId"` // 语言ID (如 "go", "typescript")
	Version    int    `json:"version"`    // 文档版本号
	Text       string `json:"text"`       // 文档内容
}

// Position 位置
// 对应 TypeScript: LSP Position
// 行号(从0开始)和列号(从0开始)
type Position struct {
	Line      int `json:"line"`      // 行号 (从0开始)
	Character int `json:"character"` // 列号 (从0开始)
}

// Range 范围
// 对应 TypeScript: LSP Range
// 表示一个文本区域
type Range struct {
	Start Position `json:"start"` // 起始位置（包含）
	End   Position `json:"end"`   // 结束位置（不包含）
}

// TextDocumentIdentifier 文本文档标识符
// 对应 TypeScript: textDocument/identifier
type TextDocumentIdentifier struct {
	URI string `json:"uri"` // 文档URI
}

// TextDocumentPositionParams 文本文档位置参数
// 对应 TypeScript: textDocument/positionParams
// 用于请求时的文档和位置信息
type TextDocumentPositionParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
}

// CompletionParams 自动完成参数
// 对应 TypeScript: completionParams
type CompletionParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
	Context      CompletionContext      `json:"context,omitempty"`
}

// CompletionContext 自动完成上下文
// 触发自动完成的上下文信息
type CompletionContext struct {
	TriggerKind      int    `json:"triggerKind"`                // 触发类型
	TriggerCharacter string `json:"triggerCharacter,omitempty"` // 触发字符
}

// CompletionList 自动完成列表
// 对应 TypeScript: completionList
type CompletionList struct {
	IsIncomplete bool             `json:"isIncomplete"` // 列表是否不完整
	Items        []CompletionItem `json:"items"`        // 自动完成项列表
}

// CompletionItem 自动完成项
// 对应 TypeScript: completionItem
type CompletionItem struct {
	Label         string    `json:"label"`                   // 显示标签
	Kind          int       `json:"kind,omitempty"`          // 项类型（如函数、变量）
	Detail        string    `json:"detail,omitempty"`        // 详细信息
	Documentation string    `json:"documentation,omitempty"` // 文档
	InsertText    string    `json:"insertText,omitempty"`    // 插入文本
	TextEdit      *TextEdit `json:"textEdit,omitempty"`      // 文本编辑
	Range         *Range    `json:"range,omitempty"`         // 范围
}

// TextEdit 文本编辑
// 对应 TypeScript: textEdit
type TextEdit struct {
	Range   Range  `json:"range"`   // 要替换的范围
	NewText string `json:"newText"` // 新文本
}

// Hover 悬停信息
// 对应 TypeScript: hover
type Hover struct {
	Contents interface{} `json:"contents"`        // 悬停内容
	Range    *Range      `json:"range,omitempty"` // 范围
}

// Location 位置信息
// 对应 TypeScript: location
type Location struct {
	URI   string `json:"uri"`   // 文档URI
	Range Range  `json:"range"` // 范围
}

// Diagnostic 诊断信息
// 对应 TypeScript: diagnostic
// 表示代码问题或警告
type Diagnostic struct {
	Range    Range  `json:"range"`            // 问题范围
	Severity int    `json:"severity"`         // 严重程度
	Source   string `json:"source,omitempty"` // 来源
	Message  string `json:"message"`          // 问题消息
}

// PublishDiagnosticsParams 发布诊断参数
// 对应 TypeScript: textDocument/publishDiagnostics
type PublishDiagnosticsParams struct {
	URI         string       `json:"uri"`         // 文档URI
	Diagnostics []Diagnostic `json:"diagnostics"` // 诊断列表
}

// NewClient 创建新的LSP客户端
// 对应 TypeScript: createLSPClient()
func NewClient(name string, conn net.Conn) *Client {
	return &Client{
		conn:  conn,
		name:  name,
		ready: false,
	}
}

// Send 发送LSP消息
// 对应 TypeScript: sendRequest()
func (c *Client) Send(msg *LSPMessage) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	_, err = c.conn.Write(append(data, '\n'))
	return err
}

// Receive 接收LSP消息
// 对应 TypeScript: 接收服务器响应
func (c *Client) Receive() (*LSPMessage, error) {
	buffer := make([]byte, 4096)
	n, err := c.conn.Read(buffer)
	if err != nil {
		return nil, fmt.Errorf("read: %w", err)
	}

	var msg LSPMessage
	if err := json.Unmarshal(buffer[:n], &msg); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}

	return &msg, nil
}

// Initialize 初始化LSP会话
// 对应 TypeScript: initialize request
// 发送初始化请求并获取服务器能力
func (c *Client) Initialize(rootURI string) (*ServerCapabilities, error) {
	msg := &LSPMessage{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params: map[string]interface{}{
			"processId": nil,
			"rootUri":   rootURI,
			"capabilities": ClientCapabilities{
				TextDocumentSync: 1,
				HoverProvider:    true,
			},
		},
	}

	if err := c.Send(msg); err != nil {
		return nil, err
	}

	resp, err := c.Receive()
	if err != nil {
		return nil, err
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("initialize error: %s", resp.Error.Message)
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if caps, ok := resp.Result.(map[string]interface{}); ok {
		if capsData, err := json.Marshal(caps["capabilities"]); err == nil {
			json.Unmarshal(capsData, &c.capabilities)
		}
	}

	c.ready = true
	return &c.capabilities, nil
}

// Initialized 发送初始化完成通知
// 对应 TypeScript: initialized notification
// 通知服务器客户端已初始化完成
func (c *Client) Initialized() error {
	msg := &LSPMessage{
		JSONRPC: "2.0",
		Method:  "initialized",
		Params:  map[string]interface{}{},
	}
	return c.Send(msg)
}

// Shutdown 请求服务器关闭
// 对应 TypeScript: shutdown request
// 请求服务器优雅关闭
func (c *Client) Shutdown() error {
	msg := &LSPMessage{
		JSONRPC: "2.0",
		ID:      2,
		Method:  "shutdown",
	}
	return c.Send(msg)
}

// Exit 退出LSP会话
// 对应 TypeScript: exit notification
func (c *Client) Exit() error {
	msg := &LSPMessage{
		JSONRPC: "2.0",
		Method:  "exit",
	}
	return c.Send(msg)
}

// TextDocumentDidOpen 文档打开通知
// 对应 TypeScript: textDocument/didOpen
func (c *Client) TextDocumentDidOpen(params TextDocumentItem) error {
	msg := &LSPMessage{
		JSONRPC: "2.0",
		Method:  "textDocument/didOpen",
		Params:  params,
	}
	return c.Send(msg)
}

// TextDocumentDidChange 文档变更通知
// 对应 TypeScript: textDocument/didChange
func (c *Client) TextDocumentDidChange(uri string, version int, text string) error {
	msg := &LSPMessage{
		JSONRPC: "2.0",
		Method:  "textDocument/didChange",
		Params: map[string]interface{}{
			"textDocument": map[string]interface{}{
				"uri":     uri,
				"version": version,
			},
			"contentChanges": []map[string]interface{}{
				{"text": text},
			},
		},
	}
	return c.Send(msg)
}

// TextDocumentDidClose 文档关闭通知
// 对应 TypeScript: textDocument/didClose
func (c *Client) TextDocumentDidClose(uri string) error {
	msg := &LSPMessage{
		JSONRPC: "2.0",
		Method:  "textDocument/didClose",
		Params: map[string]interface{}{
			"textDocument": map[string]interface{}{
				"uri": uri,
			},
		},
	}
	return c.Send(msg)
}

// Completion 请求自动完成
// 对应 TypeScript: textDocument/completion
func (c *Client) Completion(params CompletionParams) (*CompletionList, error) {
	msg := &LSPMessage{
		JSONRPC: "2.0",
		ID:      3,
		Method:  "textDocument/completion",
		Params:  params,
	}

	if err := c.Send(msg); err != nil {
		return nil, err
	}

	resp, err := c.Receive()
	if err != nil {
		return nil, err
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("completion error: %s", resp.Error.Message)
	}

	var list CompletionList
	if resp.Result != nil {
		if data, err := json.Marshal(resp.Result); err == nil {
			json.Unmarshal(data, &list)
		}
	}

	return &list, nil
}

// Hover 请求悬停信息
// 对应 TypeScript: textDocument/hover
func (c *Client) Hover(params TextDocumentPositionParams) (*Hover, error) {
	msg := &LSPMessage{
		JSONRPC: "2.0",
		ID:      4,
		Method:  "textDocument/hover",
		Params:  params,
	}

	if err := c.Send(msg); err != nil {
		return nil, err
	}

	resp, err := c.Receive()
	if err != nil {
		return nil, err
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("hover error: %s", resp.Error.Message)
	}

	var hover Hover
	if resp.Result != nil {
		if data, err := json.Marshal(resp.Result); err == nil {
			json.Unmarshal(data, &hover)
		}
	}

	return &hover, nil
}

// Definition 请求跳转定义
// 对应 TypeScript: textDocument/definition
func (c *Client) Definition(params TextDocumentPositionParams) ([]Location, error) {
	msg := &LSPMessage{
		JSONRPC: "2.0",
		ID:      5,
		Method:  "textDocument/definition",
		Params:  params,
	}

	if err := c.Send(msg); err != nil {
		return nil, err
	}

	resp, err := c.Receive()
	if err != nil {
		return nil, err
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("definition error: %s", resp.Error.Message)
	}

	var locations []Location
	if resp.Result != nil {
		if data, err := json.Marshal(resp.Result); err == nil {
			json.Unmarshal(data, &locations)
		}
	}

	return locations, nil
}

// IsReady 检查客户端是否已初始化
func (c *Client) IsReady() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.ready
}

// Close 关闭客户端连接
func (c *Client) Close() error {
	return c.conn.Close()
}

// Server LSP服务器
// 对应 TypeScript: LSPServerManager 内部服务器
// 管理多个客户端连接
type Server struct {
	mu      sync.RWMutex
	clients map[string]*Client
}

// NewServer 创建新的LSP服务器
func NewServer() *Server {
	return &Server{
		clients: make(map[string]*Client),
	}
}

// AddClient 添加客户端连接
// 对应 TypeScript: 添加LSP客户端
func (s *Server) AddClient(name string, conn net.Conn) *Client {
	s.mu.Lock()
	defer s.mu.Unlock()

	client := NewClient(name, conn)
	s.clients[name] = client
	return client
}

// RemoveClient 移除客户端连接
func (s *Server) RemoveClient(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.clients, name)
}

// GetClient 获取客户端
func (s *Server) GetClient(name string) (*Client, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	client, ok := s.clients[name]
	return client, ok
}

// Broadcast 广播消息到所有客户端
// 对应 TypeScript: 广播通知
func (s *Server) Broadcast(msg *LSPMessage) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, client := range s.clients {
		if err := client.Send(msg); err != nil {
			continue
		}
	}
	return nil
}

// Manager LSP服务器管理器
// 对应 TypeScript: LSPServerManager
// 管理多个LSP服务器实例
type Manager struct {
	mu      sync.RWMutex
	servers map[string]*Server
}

// NewManager 创建新的管理器
func NewManager() *Manager {
	return &Manager{
		servers: make(map[string]*Server),
	}
}

// CreateServer 创建新服务器
// 对应 TypeScript: createLSPServerManager()
func (m *Manager) CreateServer(name string) *Server {
	m.mu.Lock()
	defer m.mu.Unlock()

	server := NewServer()
	m.servers[name] = server
	return server
}

// GetServer 获取服务器
func (m *Manager) GetServer(name string) (*Server, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	server, ok := m.servers[name]
	return server, ok
}

// RemoveServer 移除服务器
func (m *Manager) RemoveServer(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.servers, name)
}

// ListServers 列出所有服务器
func (m *Manager) ListServers() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	servers := make([]string, 0, len(m.servers))
	for name := range m.servers {
		servers = append(servers, name)
	}
	return servers
}

// LanguageServer 语言服务器配置
// 对应 TypeScript: vscode-languageserver
// 存储要启动的LSP服务器进程配置
type LanguageServer struct {
	name    string   // 服务器名称
	command []string // 启动命令
	env     []string // 环境变量
	dir     string   // 工作目录
}

// NewLanguageServer 创建语言服务器配置
// 对应 TypeScript: 配置LSP服务器
func NewLanguageServer(name string, command []string) *LanguageServer {
	return &LanguageServer{
		name:    name,
		command: command,
		env:     make([]string, 0),
		dir:     "",
	}
}

// WithEnv 设置环境变量
func (ls *LanguageServer) WithEnv(env []string) *LanguageServer {
	ls.env = env
	return ls
}

// WithWorkingDir 设置工作目录
func (ls *LanguageServer) WithWorkingDir(dir string) *LanguageServer {
	ls.dir = dir
	return ls
}

// Start 启动语言服务器
// 对应 TypeScript: 启动LSP进程
func (ls *LanguageServer) Start(ctx context.Context) (net.Conn, error) {
	return nil, fmt.Errorf("not implemented: use ConnectToServer instead")
}

// ConnectToServer 连接到已运行的LSP服务器
// 对应 TypeScript: 连接到stdio服务器
// 通过TCP连接建立到服务器的连接
func ConnectToServer(ctx context.Context, addr string) (*Client, error) {
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("dial: %w", err)
	}

	client := NewClient(addr, conn)

	capabilities, err := client.Initialize("")
	if err != nil {
		conn.Close()
		return nil, err
	}

	if capabilities == nil {
		conn.Close()
		return nil, fmt.Errorf("failed to get server capabilities")
	}

	if err := client.Initialized(); err != nil {
		conn.Close()
		return nil, err
	}

	return client, nil
}
