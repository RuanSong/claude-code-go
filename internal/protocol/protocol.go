package protocol

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

type MessageType string

const (
	MessageTypeRequest  MessageType = "request"
	MessageTypeResponse MessageType = "response"
	MessageTypeNotify   MessageType = "notify"
)

type Message struct {
	ID        string      `json:"id"`
	Type      MessageType `json:"type"`
	Method    string      `json:"method"`
	Params    interface{} `json:"params,omitempty"`
	Result    interface{} `json:"result,omitempty"`
	Error     *Error      `json:"error,omitempty"`
	Timestamp int64       `json:"timestamp"`
}

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type Protocol struct {
	mu       sync.RWMutex
	handlers map[string]Handler
	pending  map[string]chan *Message
}

type Handler func(ctx context.Context, msg *Message) (*Message, error)

func NewProtocol() *Protocol {
	return &Protocol{
		handlers: make(map[string]Handler),
		pending:  make(map[string]chan *Message),
	}
}

func (p *Protocol) RegisterHandler(method string, handler Handler) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.handlers[method] = handler
}

func (p *Protocol) UnregisterHandler(method string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.handlers, method)
}

func (p *Protocol) HandleMessage(ctx context.Context, msg *Message) (*Message, error) {
	p.mu.RLock()
	handler, exists := p.handlers[msg.Method]
	p.mu.RUnlock()

	if !exists {
		return &Message{
			ID:    msg.ID,
			Type:  MessageTypeResponse,
			Error: &Error{Code: -32601, Message: fmt.Sprintf("method not found: %s", msg.Method)},
		}, nil
	}

	return handler(ctx, msg)
}

func (p *Protocol) SendRequest(ctx context.Context, method string, params interface{}) (*Message, error) {
	msg := &Message{
		ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
		Type:      MessageTypeRequest,
		Method:    method,
		Params:    params,
		Timestamp: time.Now().Unix(),
	}

	p.mu.Lock()
	responseCh := make(chan *Message, 1)
	p.pending[msg.ID] = responseCh
	p.mu.Unlock()

	defer func() {
		p.mu.Lock()
		delete(p.pending, msg.ID)
		p.mu.Unlock()
	}()

	response, err := p.HandleMessage(ctx, msg)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (p *Protocol) SendNotification(method string, params interface{}) error {
	msg := &Message{
		ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
		Type:      MessageTypeNotify,
		Method:    method,
		Params:    params,
		Timestamp: time.Now().Unix(),
	}

	ctx := context.Background()
	_, err := p.HandleMessage(ctx, msg)
	return err
}

type JSONRPCProtocol struct {
	*Protocol
}

func NewJSONRPCProtocol() *JSONRPCProtocol {
	return &JSONRPCProtocol{
		Protocol: NewProtocol(),
	}
}

func (p *JSONRPCProtocol) HandleJSONRequest(ctx context.Context, jsonBody []byte) ([]byte, error) {
	var msg Message
	if err := json.Unmarshal(jsonBody, &msg); err != nil {
		return nil, fmt.Errorf("parse request: %w", err)
	}

	response, err := p.HandleMessage(ctx, &msg)
	if err != nil {
		response = &Message{
			ID:    msg.ID,
			Type:  MessageTypeResponse,
			Error: &Error{Code: -32603, Message: err.Error()},
		}
	}

	return json.Marshal(response)
}

func (p *JSONRPCProtocol) RegisterBuiltinHandlers() {
	p.RegisterHandler("ping", func(ctx context.Context, msg *Message) (*Message, error) {
		return &Message{
			ID:     msg.ID,
			Type:   MessageTypeResponse,
			Result: "pong",
		}, nil
	})

	p.RegisterHandler("echo", func(ctx context.Context, msg *Message) (*Message, error) {
		return &Message{
			ID:     msg.ID,
			Type:   MessageTypeResponse,
			Result: msg.Params,
		}, nil
	})

	p.RegisterHandler("status", func(ctx context.Context, msg *Message) (*Message, error) {
		return &Message{
			ID:   msg.ID,
			Type: MessageTypeResponse,
			Result: map[string]interface{}{
				"status":    "ok",
				"timestamp": time.Now().Unix(),
			},
		}, nil
	})
}
