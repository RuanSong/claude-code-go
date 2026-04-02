package awaySummary

import (
	"context"
	"strings"
	"sync"
)

const (
	// RECENT_MESSAGE_WINDOW 最近的30条消息用于生成摘要
	RECENT_MESSAGE_WINDOW = 30
)

// AwaySummaryService "您离开期间"会话摘要服务
type AwaySummaryService struct {
	mu      sync.RWMutex
	enabled bool
}

var (
	instance *AwaySummaryService
	once     sync.Once
)

// GetInstance 获取单例实例
func GetInstance() *AwaySummaryService {
	once.Do(func() {
		instance = &AwaySummaryService{
			enabled: true,
		}
	})
	return instance
}

// BuildAwaySummaryPrompt 构建离开摘要提示词
func BuildAwaySummaryPrompt(memory string) string {
	var memoryBlock strings.Builder
	if memory != "" {
		memoryBlock.WriteString("Session memory (broader context):\n")
		memoryBlock.WriteString(memory)
		memoryBlock.WriteString("\n\n")
	}
	memoryBlock.WriteString("The user stepped away and is coming back. Write exactly 1-3 short sentences. Start by stating the high-level task — what they are building or debugging, not implementation details. Next: the concrete next step. Skip status reports and commit recaps.")
	return memoryBlock.String()
}

// GenerateAwaySummary 生成离开摘要
// messages: 会话消息列表
// signal: 中止信号
// memory: 会话记忆内容(可选)
// generateFn: 生成函数,用于调用模型生成摘要
func (s *AwaySummaryService) GenerateAwaySummary(
	ctx context.Context,
	messages []Message,
	signal <-chan struct{},
	memory string,
	generateFn func(ctx context.Context, messages []Message, prompt string) (string, error),
) (string, error) {
	s.mu.Lock()
	enabled := s.enabled
	s.mu.Unlock()

	if !enabled {
		return "", nil
	}

	if len(messages) == 0 {
		return "", nil
	}

	// 限制消息数量
	recentMessages := messages
	if len(messages) > RECENT_MESSAGE_WINDOW {
		recentMessages = messages[len(messages)-RECENT_MESSAGE_WINDOW:]
	}

	// 添加用户提示消息
	prompt := BuildAwaySummaryPrompt(memory)
	userMessage := Message{
		Role:    "user",
		Content: prompt,
	}

	// 创建带中止信号的上下文
	ctxWithCancel, cancel := context.WithCancel(ctx)
	defer cancel()

	// 监听中止信号
	go func() {
		select {
		case <-signal:
			cancel()
		case <-ctxWithCancel.Done():
		}
	}()

	messagesToSend := make([]Message, len(recentMessages)+1)
	copy(messagesToSend, recentMessages)
	messagesToSend[len(recentMessages)] = userMessage

	result, err := generateFn(ctxWithCancel, messagesToSend, prompt)
	if err != nil {
		if ctxWithCancel.Err() == context.Canceled {
			return "", nil
		}
		return "", err
	}

	return strings.TrimSpace(result), nil
}

// SetEnabled 设置是否启用
func (s *AwaySummaryService) SetEnabled(enabled bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.enabled = enabled
}

// IsEnabled 检查是否启用
func (s *AwaySummaryService) IsEnabled() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.enabled
}

// Message 消息结构
type Message struct {
	Role    string
	Content string
}
