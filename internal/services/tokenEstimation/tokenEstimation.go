package tokenEstimation

import (
	"math"
	"strings"
	"sync"
)

const (
	// TOKEN_COUNT_THINKING_BUDGET 启用思考时的最小token预算
	TOKEN_COUNT_THINKING_BUDGET = 1024
	// TOKEN_COUNT_MAX_TOKENS 启用思考时的最大token数
	TOKEN_COUNT_MAX_TOKENS = 2048
	// DEFAULT_BYTES_PER_TOKEN 默认每token字节数
	DEFAULT_BYTES_PER_TOKEN = 4
	// JSON_BYTES_PER_TOKEN JSON文件每token字节数(更密集)
	JSON_BYTES_PER_TOKEN = 2
	// IMAGE_TOKEN_ESTIMATE 图片token估计值
	IMAGE_TOKEN_ESTIMATE = 2000
)

// TokenEstimationService Token计数服务
type TokenEstimationService struct{}

var (
	tokenInstance *TokenEstimationService
	tokenOnce     sync.Once
)

// GetInstance 获取单例实例
func GetInstance() *TokenEstimationService {
	tokenOnce.Do(func() {
		tokenInstance = &TokenEstimationService{}
	})
	return tokenInstance
}

// RoughTokenCountEstimation 粗略token计数估计
// content: 待计数内容
// bytesPerToken: 每token字节数,默认4
func (s *TokenEstimationService) RoughTokenCountEstimation(content string, bytesPerToken int) int {
	if bytesPerToken <= 0 {
		bytesPerToken = DEFAULT_BYTES_PER_TOKEN
	}
	return int(math.Round(float64(len(content)) / float64(bytesPerToken)))
}

// BytesPerTokenForFileType 根据文件扩展名返回每token字节数
// 密集JSON有更多单字符token({, }, :, ,, "),使比率接近2而非默认的4
func (s *TokenEstimationService) BytesPerTokenForFileType(fileExtension string) int {
	switch strings.ToLower(fileExtension) {
	case "json", "jsonl", "jsonc":
		return JSON_BYTES_PER_TOKEN
	default:
		return DEFAULT_BYTES_PER_TOKEN
	}
}

// RoughTokenCountEstimationForFileType 根据文件类型使用更准确的比率估计token数
func (s *TokenEstimationService) RoughTokenCountEstimationForFileType(content string, fileExtension string) int {
	return s.RoughTokenCountEstimation(content, s.BytesPerTokenForFileType(fileExtension))
}

// RoughTokenCountEstimationForMessage 估计消息的token数
func (s *TokenEstimationService) RoughTokenCountEstimationForMessage(message MessageContent) int {
	content := message.GetContent()
	if content == nil {
		return 0
	}

	switch v := content.(type) {
	case string:
		return s.RoughTokenCountEstimation(v, DEFAULT_BYTES_PER_TOKEN)
	case []ContentBlock:
		return s.RoughTokenCountEstimationForBlocks(v)
	default:
		return 0
	}
}

// RoughTokenCountEstimationForMessages 估计多个消息的token总数
func (s *TokenEstimationService) RoughTokenCountEstimationForMessages(messages []MessageContent) int {
	total := 0
	for _, msg := range messages {
		total += s.RoughTokenCountEstimationForMessage(msg)
	}
	return total
}

// RoughTokenCountEstimationForBlocks 估计内容块的token数
func (s *TokenEstimationService) RoughTokenCountEstimationForBlocks(blocks []ContentBlock) int {
	total := 0
	for _, block := range blocks {
		total += s.RoughTokenCountEstimationForBlock(block)
	}
	return total
}

// RoughTokenCountEstimationForBlock 估计单个内容块的token数
func (s *TokenEstimationService) RoughTokenCountEstimationForBlock(block ContentBlock) int {
	switch block.Type {
	case "text":
		if text, ok := block.Text.(string); ok {
			return s.RoughTokenCountEstimation(text, DEFAULT_BYTES_PER_TOKEN)
		}
	case "image", "document":
		return IMAGE_TOKEN_ESTIMATE
	case "tool_result":
		if content, ok := block.Content.([]ContentBlock); ok {
			return s.RoughTokenCountEstimationForBlocks(content)
		}
	case "tool_use":
		inputStr := jsonStringify(block.Input)
		nameStr := block.Name
		return s.RoughTokenCountEstimation(nameStr+inputStr, DEFAULT_BYTES_PER_TOKEN)
	case "thinking":
		if thinking, ok := block.Thinking.(string); ok {
			return s.RoughTokenCountEstimation(thinking, DEFAULT_BYTES_PER_TOKEN)
		}
	case "redacted_thinking":
		if data, ok := block.Data.(string); ok {
			return s.RoughTokenCountEstimation(data, DEFAULT_BYTES_PER_TOKEN)
		}
	default:
		return s.RoughTokenCountEstimation(jsonStringify(block), DEFAULT_BYTES_PER_TOKEN)
	}
	return 0
}

// jsonStringify JSON序列化(简化版本,用于token估计)
func jsonStringify(v interface{}) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	case map[string]interface{}:
		if len(val) == 0 {
			return "{}"
		}
		var sb strings.Builder
		sb.WriteString("{")
		first := true
		for k, v := range val {
			if !first {
				sb.WriteString(",")
			}
			first = false
			sb.WriteString(`"` + k + `":`)
			switch vv := v.(type) {
			case string:
				sb.WriteString(`"` + vv + `"`)
			default:
				sb.WriteString(jsonStringify(vv))
			}
		}
		sb.WriteString("}")
		return sb.String()
	default:
		return ""
	}
}

// MessageContent 消息内容接口
type MessageContent interface {
	GetContent() interface{}
}

type messageContent struct {
	content interface{}
}

func (m *messageContent) GetContent() interface{} {
	return m.content
}

// NewMessageContent 创建消息内容
func NewMessageContent(content interface{}) MessageContent {
	return &messageContent{content: content}
}

// ContentBlock 内容块
type ContentBlock struct {
	Type     string      `json:"type"`
	Text     interface{} `json:"text,omitempty"`
	Thinking interface{} `json:"thinking,omitempty"`
	Data     interface{} `json:"data,omitempty"`
	Input    interface{} `json:"input,omitempty"`
	Name     string      `json:"name,omitempty"`
	Content  interface{} `json:"content,omitempty"`
}

// TokenUsage API返回的token使用量
type TokenUsage struct {
	InputTokens              int64 `json:"input_tokens"`
	OutputTokens             int64 `json:"output_tokens"`
	CacheReadInputTokens     int64 `json:"cache_read_input_tokens,omitempty"`
	CacheCreationInputTokens int64 `json:"cache_creation_input_tokens,omitempty"`
}
