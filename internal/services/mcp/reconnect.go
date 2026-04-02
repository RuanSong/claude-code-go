package mcp

import (
	"context"
	"math"
	"sync"
	"time"
)

const (
	DefaultMaxRetries        = 3
	DefaultInitialDelayMs    = 1000
	DefaultMaxDelayMs        = 30000
	DefaultBackoffMultiplier = 2.0
)

type ReconnectConfig struct {
	MaxRetries        int
	InitialDelayMs    int64
	MaxDelayMs        int64
	BackoffMultiplier float64
}

func DefaultReconnectConfig() *ReconnectConfig {
	return &ReconnectConfig{
		MaxRetries:        DefaultMaxRetries,
		InitialDelayMs:    DefaultInitialDelayMs,
		MaxDelayMs:        DefaultMaxDelayMs,
		BackoffMultiplier: DefaultBackoffMultiplier,
	}
}

type ReconnectState struct {
	mu                  sync.Mutex
	attempt             int
	lastAttempt         time.Time
	consecutiveFailures int
	disabled            bool
}

func NewReconnectState() *ReconnectState {
	return &ReconnectState{
		attempt:             0,
		consecutiveFailures: 0,
		disabled:            false,
	}
}

func (s *ReconnectState) NextDelay(config *ReconnectConfig) time.Duration {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.attempt >= config.MaxRetries {
		return 0
	}

	delay := float64(config.InitialDelayMs) * math.Pow(config.BackoffMultiplier, float64(s.attempt))
	if delay > float64(config.MaxDelayMs) {
		delay = float64(config.MaxDelayMs)
	}

	s.attempt++
	s.lastAttempt = time.Now()

	return time.Duration(delay) * time.Millisecond
}

func (s *ReconnectState) RecordSuccess() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.attempt = 0
	s.consecutiveFailures = 0
}

func (s *ReconnectState) RecordFailure() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.consecutiveFailures++
}

func (s *ReconnectState) ShouldRetry(config *ReconnectConfig) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.attempt < config.MaxRetries && !s.disabled
}

func (s *ReconnectState) Disable() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.disabled = true
}

func (s *ReconnectState) Enable() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.disabled = false
}

func (s *ReconnectState) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.attempt = 0
	s.consecutiveFailures = 0
	s.disabled = false
}

type ConnectionManager struct {
	mu        sync.RWMutex
	clients   map[string]*Client
	configs   map[string]*ServerConfig
	transport map[string]Transport
	states    map[string]*ReconnectState
	config    *ReconnectConfig
	protocol  *MCPProtocol
}

func NewConnectionManager(protocol *MCPProtocol) *ConnectionManager {
	return &ConnectionManager{
		clients:   make(map[string]*Client),
		configs:   make(map[string]*ServerConfig),
		transport: make(map[string]Transport),
		states:    make(map[string]*ReconnectState),
		config:    DefaultReconnectConfig(),
		protocol:  protocol,
	}
}

func (m *ConnectionManager) AddServer(name string, config *ServerConfig) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.configs[name] = config
	m.states[name] = NewReconnectState()
}

func (m *ConnectionManager) RemoveServer(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.clients, name)
	delete(m.configs, name)
	delete(m.transport, name)
	delete(m.states, name)
}

func (m *ConnectionManager) GetClient(name string) (*Client, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	client, ok := m.clients[name]
	return client, ok
}

func (m *ConnectionManager) Connect(ctx context.Context, name string) error {
	m.mu.RLock()
	config, ok := m.configs[name]
	state := m.states[name]
	m.mu.RUnlock()

	if !ok {
		return &McpToolCallError{Code: -32000, Message: "server not found: " + name}
	}

	for state.ShouldRetry(m.config) {
		client := NewClient(&ClientConfig{
			Name:      name,
			Transport: config.Transport,
			Command:   config.Command,
			Args:      config.Args,
			Env:       config.Env,
			URL:       config.URL,
			Headers:   config.Headers,
		})

		var transport Transport
		switch config.Transport {
		case TransportStdio:
			transport = NewStdioTransport(config.Command, config.Args, config.Env)
		case TransportSSE:
			transport = NewSSETransport(config.URL, config.Headers)
		case TransportHTTP:
			transport = NewHTTPTransport(config.URL, config.Headers)
		case TransportWS:
			transport = NewWebSocketTransport(config.URL, config.Headers)
		default:
			transport = NewStdioTransport(config.Command, config.Args, config.Env)
		}

		if err := transport.Start(ctx); err != nil {
			state.RecordFailure()
			delay := state.NextDelay(m.config)
			if delay > 0 {
				time.Sleep(delay)
			}
			continue
		}

		m.mu.Lock()
		m.clients[name] = client
		m.transport[name] = transport
		m.mu.Unlock()

		state.RecordSuccess()
		return nil
	}

	state.Disable()
	return &McpToolCallError{Code: -32000, Message: "failed to connect after " + string(rune('0'+m.config.MaxRetries)) + " attempts"}
}

func (m *ConnectionManager) Disconnect(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	transport, ok := m.transport[name]
	if !ok {
		return nil
	}

	if err := transport.Close(); err != nil {
		return err
	}

	delete(m.transport, name)
	delete(m.clients, name)

	state, ok := m.states[name]
	if ok {
		state.Reset()
	}

	return nil
}

func (m *ConnectionManager) Reconnect(ctx context.Context, name string) error {
	if err := m.Disconnect(name); err != nil {
	}
	return m.Connect(ctx, name)
}

type McpToolCallError struct {
	Code    int
	Message string
}

func (e *McpToolCallError) Error() string {
	return e.Message
}
