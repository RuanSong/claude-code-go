package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Transport interface {
	Start(ctx context.Context) error
	Send(msg *JSONRPCMessage) error
	Receive() (*JSONRPCMessage, error)
	Close() error
}

type StdioTransport struct {
	cmd    interface{}
	args   []string
	env    map[string]string
	stdin  io.WriteCloser
	stdout io.ReadCloser
	stderr io.ReadCloser
	mu     sync.Mutex
	closed bool
}

func NewStdioTransport(command string, args []string, env map[string]string) *StdioTransport {
	return &StdioTransport{
		args: args,
		env:  env,
	}
}

type SSETransport struct {
	url         string
	headers     map[string]string
	client      *http.Client
	resp        *http.Response
	eventSource io.ReadCloser
	mu          sync.Mutex
	closed      bool
	lastID      int
	readCh      chan *JSONRPCMessage
}

func NewSSETransport(url string, headers map[string]string) *SSETransport {
	return &SSETransport{
		url:     url,
		headers: headers,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		readCh: make(chan *JSONRPCMessage, 100),
	}
}

type HTTPTransport struct {
	url          string
	headers      map[string]string
	sessionID    string
	client       *http.Client
	mu           sync.Mutex
	closed       bool
	authProvider *AuthProvider
}

func NewHTTPTransport(url string, headers map[string]string) *HTTPTransport {
	return &HTTPTransport{
		url:     url,
		headers: headers,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

type WebSocketTransport struct {
	url     string
	headers map[string]string
	conn    *websocket.Conn
	mu      sync.Mutex
	closed  bool
	readCh  chan *JSONRPCMessage
	writeCh chan *JSONRPCMessage
}

func NewWebSocketTransport(url string, headers map[string]string) *WebSocketTransport {
	return &WebSocketTransport{
		url:     url,
		headers: headers,
		readCh:  make(chan *JSONRPCMessage, 100),
		writeCh: make(chan *JSONRPCMessage, 100),
	}
}

type AuthProvider struct {
	serverName string
	config     *ServerConfig
	tokens     *OAuthToken
	mu         sync.RWMutex
}

func NewAuthProvider(serverName string, config *ServerConfig) *AuthProvider {
	return &AuthProvider{
		serverName: serverName,
		config:     config,
	}
}

func (p *AuthProvider) Tokens() (*OAuthToken, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.tokens, nil
}

func (p *AuthProvider) SetTokens(token *OAuthToken) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.tokens = token
}

func (t *StdioTransport) Start(ctx context.Context) error {
	return nil
}

func (t *StdioTransport) Send(msg *JSONRPCMessage) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.closed {
		return fmt.Errorf("transport closed")
	}
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	_, err = t.stdin.Write(data)
	return err
}

func (t *StdioTransport) Receive() (*JSONRPCMessage, error) {
	if t.closed {
		return nil, fmt.Errorf("transport closed")
	}
	decoder := json.NewDecoder(t.stdout)
	var msg JSONRPCMessage
	if err := decoder.Decode(&msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

func (t *StdioTransport) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.closed = true
	if t.stdin != nil {
		t.stdin.Close()
	}
	return nil
}

func (t *SSETransport) Start(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", t.url, nil)
	if err != nil {
		return err
	}
	for k, v := range t.headers {
		req.Header.Set(k, v)
	}
	resp, err := t.client.Do(req)
	if err != nil {
		return err
	}
	t.resp = resp
	go t.readSSEEvents()
	return nil
}

func (t *SSETransport) readSSEEvents() {
	defer t.resp.Body.Close()
	reader := t.resp.Body
	buf := make([]byte, 4096)
	for {
		n, err := reader.Read(buf)
		if err != nil {
			close(t.readCh)
			return
		}
		t.parseSSEData(buf[:n])
	}
}

func (t *SSETransport) parseSSEData(data []byte) {
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "data: ") {
			jsonStr := strings.TrimPrefix(line, "data: ")
			var msg JSONRPCMessage
			if err := json.Unmarshal([]byte(jsonStr), &msg); err == nil {
				select {
				case t.readCh <- &msg:
				default:
				}
			}
		}
	}
}

func (t *SSETransport) Send(msg *JSONRPCMessage) error {
	return fmt.Errorf("SSE transport does not support sending")
}

func (t *SSETransport) Receive() (*JSONRPCMessage, error) {
	return <-t.readCh, nil
}

func (t *SSETransport) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.closed = true
	if t.resp != nil {
		t.resp.Body.Close()
	}
	return nil
}

func (t *HTTPTransport) Start(ctx context.Context) error {
	return nil
}

func (t *HTTPTransport) Send(msg *JSONRPCMessage) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.closed {
		return fmt.Errorf("transport closed")
	}
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", t.url, strings.NewReader(string(data)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if t.sessionID != "" {
		req.Header.Set("MCP-Session-ID", t.sessionID)
	}
	for k, v := range t.headers {
		req.Header.Set(k, v)
	}
	resp, err := t.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.Header.Get("MCP-Session-ID") != "" {
		t.sessionID = resp.Header.Get("MCP-Session-ID")
	}
	return nil
}

func (t *HTTPTransport) Receive() (*JSONRPCMessage, error) {
	return nil, fmt.Errorf("HTTP transport does not support receiving")
}

func (t *HTTPTransport) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.closed = true
	return nil
}

func (t *WebSocketTransport) Start(ctx context.Context) error {
	header := http.Header{}
	for k, v := range t.headers {
		header.Set(k, v)
	}
	conn, _, err := websocket.DefaultDialer.DialContext(ctx, t.url, header)
	if err != nil {
		return err
	}
	t.conn = conn
	go t.readMessages()
	return nil
}

func (t *WebSocketTransport) readMessages() {
	defer t.conn.Close()
	for {
		_, msg, err := t.conn.ReadMessage()
		if err != nil {
			close(t.readCh)
			return
		}
		var rpcMsg JSONRPCMessage
		if err := json.Unmarshal(msg, &rpcMsg); err == nil {
			select {
			case t.readCh <- &rpcMsg:
			default:
			}
		}
	}
}

func (t *WebSocketTransport) Send(msg *JSONRPCMessage) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.closed {
		return fmt.Errorf("transport closed")
	}
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return t.conn.WriteMessage(websocket.TextMessage, data)
}

func (t *WebSocketTransport) Receive() (*JSONRPCMessage, error) {
	return <-t.readCh, nil
}

func (t *WebSocketTransport) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.closed = true
	if t.conn != nil {
		return t.conn.Close()
	}
	return nil
}
