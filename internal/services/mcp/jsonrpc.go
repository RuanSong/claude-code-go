package mcp

import "encoding/json"

type JSONRPCMessage struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	ID      interface{}     `json:"id,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *JSONRPCError   `json:"error,omitempty"`
}

type JSONRPCError struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
	ID      interface{}     `json:"id"`
}

type JSONRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *JSONRPCError   `json:"error,omitempty"`
	ID      interface{}     `json:"id"`
}

type JSONRPCNotification struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

const (
	JSONRPCVersion = "2.0"
)

func NewJSONRPCRequest(method string, params interface{}, id interface{}) (*JSONRPCRequest, error) {
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	return &JSONRPCRequest{
		JSONRPC: JSONRPCVersion,
		Method:  method,
		Params:  paramsJSON,
		ID:      id,
	}, nil
}

func NewJSONRPCNotification(method string, params interface{}) (*JSONRPCNotification, error) {
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	return &JSONRPCNotification{
		JSONRPC: JSONRPCVersion,
		Method:  method,
		Params:  paramsJSON,
	}, nil
}

func (m *JSONRPCMessage) IsRequest() bool {
	return m.Method != "" && m.ID != nil
}

func (m *JSONRPCMessage) IsNotification() bool {
	return m.Method != "" && m.ID == nil
}

func (m *JSONRPCMessage) IsResponse() bool {
	return m.Result != nil || m.Error != nil
}
