package mcp

import (
	"context"
	"encoding/json"
	"fmt"
)

type Prompt struct {
	Name        string           `json:"name"`
	Description string           `json:"description,omitempty"`
	Arguments   []PromptArgument `json:"arguments,omitempty"`
}

type PromptArgument struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Required    bool   `json:"required,omitempty"`
}

type PromptMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ListPromptsResult struct {
	Prompts []Prompt `json:"prompts"`
}

type GetPromptResult struct {
	Messages    []PromptMessage `json:"messages"`
	Description string          `json:"description,omitempty"`
}

type ListRootsResult struct {
	Roots []Root `json:"roots"`
}

type Root struct {
	URI  string `json:"uri"`
	Name string `json:"name,omitempty"`
}

type RootsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

type PromptsHandler interface {
	ListPrompts(ctx context.Context) (*ListPromptsResult, error)
	GetPrompt(ctx context.Context, name string, arguments map[string]string) (*GetPromptResult, error)
}

type RootsHandler interface {
	ListRoots(ctx context.Context) (*ListRootsResult, error)
}

type PromptHandler struct {
	transport Transport
}

func NewPromptHandler(transport Transport) *PromptHandler {
	return &PromptHandler{transport: transport}
}

func (h *PromptHandler) ListPrompts(ctx context.Context) (*ListPromptsResult, error) {
	if h.transport == nil {
		return nil, fmt.Errorf("transport not initialized")
	}

	params := map[string]interface{}{}
	paramsJSON, _ := json.Marshal(params)

	msg := &JSONRPCMessage{
		JSONRPC: JSONRPCVersion,
		Method:  "prompts/list",
		Params:  paramsJSON,
		ID:      1,
	}

	if err := h.transport.Send(msg); err != nil {
		return nil, err
	}

	resp, err := h.transport.Receive()
	if err != nil {
		return nil, err
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("MCP error: %s", resp.Error.Message)
	}

	var result ListPromptsResult
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (h *PromptHandler) GetPrompt(ctx context.Context, name string, arguments map[string]string) (*GetPromptResult, error) {
	if h.transport == nil {
		return nil, fmt.Errorf("transport not initialized")
	}

	params := map[string]interface{}{
		"name": name,
	}
	if len(arguments) > 0 {
		params["arguments"] = arguments
	}
	paramsJSON, _ := json.Marshal(params)

	msg := &JSONRPCMessage{
		JSONRPC: JSONRPCVersion,
		Method:  "prompts/get",
		Params:  paramsJSON,
		ID:      1,
	}

	if err := h.transport.Send(msg); err != nil {
		return nil, err
	}

	resp, err := h.transport.Receive()
	if err != nil {
		return nil, err
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("MCP error: %s", resp.Error.Message)
	}

	var result GetPromptResult
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

type RootsHandlerImpl struct {
	transport Transport
}

func NewRootsHandler(transport Transport) *RootsHandlerImpl {
	return &RootsHandlerImpl{transport: transport}
}

func (h *RootsHandlerImpl) ListRoots(ctx context.Context) (*ListRootsResult, error) {
	if h.transport == nil {
		return nil, fmt.Errorf("transport not initialized")
	}

	params := map[string]interface{}{}
	paramsJSON, _ := json.Marshal(params)

	msg := &JSONRPCMessage{
		JSONRPC: JSONRPCVersion,
		Method:  "roots/list",
		Params:  paramsJSON,
		ID:      1,
	}

	if err := h.transport.Send(msg); err != nil {
		return nil, err
	}

	resp, err := h.transport.Receive()
	if err != nil {
		return nil, err
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("MCP error: %s", resp.Error.Message)
	}

	var result ListRootsResult
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, err
	}

	return &result, nil
}
