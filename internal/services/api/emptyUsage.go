package api

// EMPTY_USAGE 空使用量对象 - 用于未使用API的情况
// 从logging.ts提取出来，避免循环依赖
var EMPTY_USAGE = Usage{
	InputTokens:              0,
	CacheCreationInputTokens: 0,
	CacheReadInputTokens:     0,
	OutputTokens:             0,
	ServerToolUse: ServerToolUse{
		WebSearchRequests: 0,
		WebFetchRequests:  0,
	},
	ServiceTier: "standard",
	CacheCreation: CacheCreation{
		Ephemeral1hInputTokens: 0,
		Ephemeral5mInputTokens: 0,
	},
	InferenceGeo: "",
	Iterations:   []interface{}{},
	Speed:        "standard",
}

// Usage API使用量结构
type Usage struct {
	InputTokens              int64         `json:"input_tokens"`
	CacheCreationInputTokens int64         `json:"cache_creation_input_tokens"`
	CacheReadInputTokens     int64         `json:"cache_read_input_tokens"`
	OutputTokens             int64         `json:"output_tokens"`
	ServerToolUse            ServerToolUse `json:"server_tool_use"`
	ServiceTier              string        `json:"service_tier"`
	CacheCreation            CacheCreation `json:"cache_creation"`
	InferenceGeo             string        `json:"inference_geo"`
	Iterations               []interface{} `json:"iterations"`
	Speed                    string        `json:"speed"`
}

// ServerToolUse 服务器工具使用统计
type ServerToolUse struct {
	WebSearchRequests int `json:"web_search_requests"`
	WebFetchRequests  int `json:"web_fetch_requests"`
}

// CacheCreation 缓存创建统计
type CacheCreation struct {
	Ephemeral1hInputTokens int `json:"ephemeral_1h_input_tokens"`
	Ephemeral5mInputTokens int `json:"ephemeral_5m_input_tokens"`
}

// GetEmptyUsage 获取空使用量对象
func GetEmptyUsage() Usage {
	return EMPTY_USAGE
}
