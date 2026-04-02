package bridge

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"
)

type BridgeType string

const (
	BridgeTypeVSCode    BridgeType = "vscode"
	BridgeTypeJetBrains BridgeType = "jetbrains"
)

type Message struct {
	Type    string      `json:"type"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
	ID      string      `json:"id,omitempty"`
	Session string      `json:"session,omitempty"`
}

type Bridge struct {
	mu       sync.RWMutex
	ctype    BridgeType
	conn     net.Conn
	handlers map[string]Handler
	enabled  bool
}

type Handler func(msg *Message) (*Message, error)

func NewBridge(ctype BridgeType) *Bridge {
	return &Bridge{
		ctype:    ctype,
		handlers: make(map[string]Handler),
		enabled:  false,
	}
}

func (b *Bridge) Connect(addr string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		return fmt.Errorf("dial: %w", err)
	}

	b.conn = conn
	b.enabled = true
	return nil
}

func (b *Bridge) Disconnect() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.conn != nil {
		err := b.conn.Close()
		b.conn = nil
		b.enabled = false
		return err
	}
	return nil
}

func (b *Bridge) IsConnected() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.enabled && b.conn != nil
}

func (b *Bridge) Send(msg *Message) error {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.conn == nil {
		return fmt.Errorf("not connected")
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	_, err = b.conn.Write(append(data, '\n'))
	return err
}

func (b *Bridge) Receive() (*Message, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.conn == nil {
		return nil, fmt.Errorf("not connected")
	}

	buffer := make([]byte, 4096)
	n, err := b.conn.Read(buffer)
	if err != nil {
		return nil, fmt.Errorf("read: %w", err)
	}

	var msg Message
	if err := json.Unmarshal(buffer[:n], &msg); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}

	return &msg, nil
}

func (b *Bridge) RegisterHandler(method string, handler Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[method] = handler
}

func (b *Bridge) UnregisterHandler(method string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.handlers, method)
}

func (b *Bridge) HandleMessage(ctx context.Context, msg *Message) (*Message, error) {
	b.mu.RLock()
	handler, exists := b.handlers[msg.Method]
	b.mu.RUnlock()

	if !exists {
		return &Message{
			Type:   "error",
			Method: msg.Method,
			ID:     msg.ID,
			Params: map[string]interface{}{
				"code":    -32601,
				"message": fmt.Sprintf("method not found: %s", msg.Method),
			},
		}, nil
	}

	msg, err := handler(msg)
	return msg, err
}

type VSCodeBridge struct {
	*Bridge
}

func NewVSCodeBridge() *VSCodeBridge {
	return &VSCodeBridge{
		Bridge: NewBridge(BridgeTypeVSCode),
	}
}

func (b *VSCodeBridge) RegisterBuiltinHandlers() {
	b.RegisterHandler("getFileContents", func(msg *Message) (*Message, error) {
		return &Message{
			Type:   "response",
			ID:     msg.ID,
			Params: map[string]interface{}{"contents": ""},
		}, nil
	})

	b.RegisterHandler("applyEdit", func(msg *Message) (*Message, error) {
		return &Message{
			Type:   "response",
			ID:     msg.ID,
			Params: map[string]interface{}{"applied": true},
		}, nil
	})

	b.RegisterHandler("showNotification", func(msg *Message) (*Message, error) {
		return &Message{
			Type:   "response",
			ID:     msg.ID,
			Params: map[string]interface{}{"shown": true},
		}, nil
	})
}

type JetBrainsBridge struct {
	*Bridge
}

func NewJetBrainsBridge() *JetBrainsBridge {
	return &JetBrainsBridge{
		Bridge: NewBridge(BridgeTypeJetBrains),
	}
}

func (b *JetBrainsBridge) RegisterBuiltinHandlers() {
	b.RegisterHandler("getFileContents", func(msg *Message) (*Message, error) {
		return &Message{
			Type:   "response",
			ID:     msg.ID,
			Params: map[string]interface{}{"contents": ""},
		}, nil
	})

	b.RegisterHandler("applyEdit", func(msg *Message) (*Message, error) {
		return &Message{
			Type:   "response",
			ID:     msg.ID,
			Params: map[string]interface{}{"applied": true},
		}, nil
	})
}

type BridgeManager struct {
	mu      sync.RWMutex
	bridges map[BridgeType]*Bridge
}

func NewBridgeManager() *BridgeManager {
	return &BridgeManager{
		bridges: make(map[BridgeType]*Bridge),
	}
}

func (m *BridgeManager) Register(bridge *Bridge) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.bridges[bridge.ctype] = bridge
}

func (m *BridgeManager) Get(ctype BridgeType) (*Bridge, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	bridge, ok := m.bridges[ctype]
	return bridge, ok
}

func (m *BridgeManager) Unregister(ctype BridgeType) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.bridges, ctype)
}

func (m *BridgeManager) List() []BridgeType {
	m.mu.RLock()
	defer m.mu.RUnlock()
	types := make([]BridgeType, 0, len(m.bridges))
	for ctype := range m.bridges {
		types = append(types, ctype)
	}
	return types
}

func RunBridgeServer(addr string, manager *BridgeManager) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}

		go handleConnection(conn, manager)
	}
}

func handleConnection(conn net.Conn, manager *BridgeManager) {
	defer conn.Close()

	decoder := json.NewDecoder(conn)
	encoder := json.NewEncoder(conn)

	for {
		var msg Message
		if err := decoder.Decode(&msg); err != nil {
			return
		}

		response := &Message{
			Type:   "response",
			ID:     msg.ID,
			Params: map[string]interface{}{"handled": true},
		}

		encoder.Encode(response)
	}
}
