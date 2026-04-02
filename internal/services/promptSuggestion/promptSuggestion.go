package promptSuggestion

import (
	"strings"
	"sync"
	"unicode"
)

// PromptVariant 提示变体类型
type PromptVariant string

const (
	PromptVariantUserIntent   PromptVariant = "user_intent"   // 用户意图
	PromptVariantStatedIntent PromptVariant = "stated_intent" // 声明的意图
)

// SUGGESTION_PROMPT 建议提示词模板
const SUGGESTION_PROMPT = `[SUGGESTION MODE: Suggest what the user might naturally type next into Claude Code.]

FIRST: Look at the user's recent messages and original request.

Your job is to predict what THEY would type - not what you think they should do.

THE TEST: Would they think "I was just about to type that"?

EXAMPLES:
User asked "fix the bug and run tests", bug is fixed → "run the tests"
After code written → "try it out"
Claude offers options → suggest the one the user would likely pick, based on conversation
Claude asks to continue → "yes" or "go ahead"
Task complete, obvious follow-up → "commit this" or "push it"
After error or misunderstanding → silence (let them assess/correct)

Be specific: "run the tests" beats "continue".

NEVER SUGGEST:
- Evaluative ("looks good", "thanks")
- Questions ("what about...?")
- Claude-voice ("Let me...", "I'll...", "Here's...")
- New ideas they didn't ask about
- Multiple sentences

Stay silent if the next step isn't obvious from what the user said.

Format: 2-12 words, match the user's style. Or nothing.

Reply with ONLY the suggestion, no quotes or explanation.`

// PromptSuggestionService 提示建议服务
type PromptSuggestionService struct {
	mu               sync.RWMutex
	enabled          bool
	currentAbortCtrl *AbortController
}

var (
	promptSuggestionInstance *PromptSuggestionService
	promptSuggestionOnce     sync.Once
)

// AbortController 中止控制器
type AbortController struct {
	ch chan struct{}
}

func NewAbortController() *AbortController {
	return &AbortController{
		ch: make(chan struct{}),
	}
}

func (ac *AbortController) Abort() {
	close(ac.ch)
}

func (ac *AbortController) IsAborted() bool {
	select {
	case <-ac.ch:
		return true
	default:
		return false
	}
}

func (ac *AbortController) Done() <-chan struct{} {
	return ac.ch
}

// GetInstance 获取单例实例
func GetInstance() *PromptSuggestionService {
	promptSuggestionOnce.Do(func() {
		promptSuggestionInstance = &PromptSuggestionService{
			enabled: true,
		}
	})
	return promptSuggestionInstance
}

// ShouldEnablePromptSuggestion 检查是否应启用提示建议
func (s *PromptSuggestionService) ShouldEnablePromptSuggestion() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.enabled
}

// SetEnabled 设置是否启用
func (s *PromptSuggestionService) SetEnabled(enabled bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.enabled = enabled
}

// AbortPromptSuggestion 中止提示建议
func (s *PromptSuggestionService) AbortPromptSuggestion() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.currentAbortCtrl != nil {
		s.currentAbortCtrl.Abort()
		s.currentAbortCtrl = nil
	}
}

// GetSuggestionSuppressReason 获取建议抑制原因
// 如果返回nil表示允许生成
func (s *PromptSuggestionService) GetSuggestionSuppressReason(state *AppState) string {
	if state == nil {
		return "no_state"
	}

	if !s.enabled {
		return "disabled"
	}

	if state.PendingWorkerRequest || state.PendingSandboxRequest {
		return "pending_permission"
	}

	if state.ElicitationQueueLength > 0 {
		return "elicitation_active"
	}

	if state.ToolPermissionMode == "plan" {
		return "plan_mode"
	}

	return ""
}

// TryGenerateSuggestion 尝试生成建议
// 返回建议和元数据,或nil如果被抑制/过滤
func (s *PromptSuggestionService) TryGenerateSuggestion(
	abortCtrl *AbortController,
	messages []SuggestionMessage,
	state *AppState,
	cacheSafeParams *CacheSafeParams,
) *SuggestionResult {
	// 检查中止
	if abortCtrl.IsAborted() {
		return nil
	}

	// 检查对话是否太短
	assistantTurnCount := countAssistantTurns(messages)
	if assistantTurnCount < 2 {
		return nil
	}

	// 检查最后一条助手消息是否是错误
	lastAssistantMsg := getLastAssistantMessage(messages)
	if lastAssistantMsg != nil && lastAssistantMsg.IsApiError {
		return nil
	}

	// 检查父缓存抑制原因
	if getParentCacheSuppressReason(lastAssistantMsg) != "" {
		return nil
	}

	// 检查应用状态抑制原因
	suppressReason := s.GetSuggestionSuppressReason(state)
	if suppressReason != "" {
		return nil
	}

	// 生成建议
	suggestion := s.GenerateSuggestion(abortCtrl, PromptVariantUserIntent, cacheSafeParams)

	// 再次检查中止
	if abortCtrl.IsAborted() {
		return nil
	}

	if suggestion == "" {
		return nil
	}

	// 过滤建议
	if ShouldFilterSuggestion(suggestion, PromptVariantUserIntent) {
		return nil
	}

	return &SuggestionResult{
		Suggestion: suggestion,
		PromptId:   PromptVariantUserIntent,
	}
}

// GenerateSuggestion 生成建议
func (s *PromptSuggestionService) GenerateSuggestion(
	abortCtrl *AbortController,
	promptId PromptVariant,
	cacheSafeParams *CacheSafeParams,
) string {
	// 在实际实现中,这会调用forked agent
	// 这里返回简化版本
	return ""
}

// ShouldFilterSuggestion 检查是否应过滤建议
func ShouldFilterSuggestion(suggestion string, promptId PromptVariant) bool {
	if suggestion == "" {
		return true
	}

	lower := strings.ToLower(suggestion)
	wordCount := countWords(suggestion)

	// 过滤器列表
	filters := []struct {
		name  string
		check func() bool
	}{
		// 完全匹配 "done"
		{"done", func() bool { return lower == "done" }},

		// 元文本
		{"meta_text", func() bool {
			return lower == "nothing found" ||
				lower == "nothing found." ||
				strings.HasPrefix(lower, "nothing to suggest") ||
				strings.HasPrefix(lower, "no suggestion") ||
				strings.Contains(lower, "silence is") ||
				strings.Contains(lower, "stay silent") ||
				strings.HasPrefix(lower, "silence")
		}},

		// 元包装
		{"meta_wrapped", func() bool {
			return (strings.HasPrefix(lower, "(") && strings.HasSuffix(lower, ")")) ||
				(strings.HasPrefix(lower, "[") && strings.HasSuffix(lower, "]"))
		}},

		// 错误消息
		{"error_message", func() bool {
			return strings.HasPrefix(lower, "api error:") ||
				strings.HasPrefix(lower, "prompt is too long") ||
				strings.HasPrefix(lower, "request timed out") ||
				strings.HasPrefix(lower, "invalid api key") ||
				strings.HasPrefix(lower, "image was too large")
		}},

		// 前缀标签
		{"prefixed_label", func() bool {
			for i, c := range suggestion {
				if c == ':' && i > 0 && i < len(suggestion)-1 && suggestion[i+1] == ' ' {
					return true
				}
			}
			return false
		}},

		// 太少单词
		{"too_few_words", func() bool {
			if wordCount >= 2 {
				return false
			}
			// 允许斜杠命令
			if strings.HasPrefix(suggestion, "/") {
				return false
			}
			// 允许常见单词
			allowedWords := map[string]bool{
				"yes": true, "yeah": true, "yep": true, "yea": true, "yup": true,
				"sure": true, "ok": true, "okay": true,
				"push": true, "commit": true, "deploy": true, "stop": true,
				"continue": true, "check": true, "exit": true, "quit": true,
				"no": true,
			}
			return !allowedWords[lower]
		}},

		// 太多单词
		{"too_many_words", func() bool { return wordCount > 12 }},

		// 太长
		{"too_long", func() bool { return len(suggestion) >= 100 }},

		// 多句
		{"multiple_sentences", func() bool {
			for i := 0; i < len(suggestion)-1; i++ {
				if suggestion[i] == '.' || suggestion[i] == '!' || suggestion[i] == '?' {
					nextChar := rune(suggestion[i+1])
					if unicode.IsUpper(nextChar) {
						return true
					}
				}
			}
			return false
		}},

		// 有格式
		{"has_formatting", func() bool {
			return strings.Contains(suggestion, "\n") ||
				strings.Contains(suggestion, "*") ||
				strings.Contains(suggestion, "**")
		}},

		// 评估性文本
		{"evaluative", func() bool {
			evalWords := []string{"thanks", "thank you", "looks good", "sounds good",
				"that works", "that worked", "that's all", "nice", "great",
				"perfect", "makes sense", "awesome", "excellent"}
			for _, word := range evalWords {
				if strings.Contains(lower, word) {
					return true
				}
			}
			return false
		}},

		// Claude声音
		{"claude_voice", func() bool {
			claudePrefixes := []string{"let me", "i'll", "i've", "i'm", "i can", "i would",
				"i think", "i notice", "here's", "here is", "here are", "that's",
				"this is", "this will", "you can", "you should", "you could",
				"sure,", "of course", "certainly"}
			for _, prefix := range claudePrefixes {
				if strings.HasPrefix(lower, prefix) {
					return true
				}
			}
			return false
		}},
	}

	for _, filter := range filters {
		if filter.check() {
			return true
		}
	}

	return false
}

// Helper functions

func countAssistantTurns(messages []SuggestionMessage) int {
	count := 0
	for _, msg := range messages {
		if msg.Type == "assistant" {
			count++
		}
	}
	return count
}

func getLastAssistantMessage(messages []SuggestionMessage) *SuggestionMessage {
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Type == "assistant" {
			return &messages[i]
		}
	}
	return nil
}

func getParentCacheSuppressReason(lastMsg *SuggestionMessage) string {
	if lastMsg == nil {
		return ""
	}

	// 检查使用量
	usage := lastMsg.Usage
	if usage == nil {
		return ""
	}

	inputTokens := usage.InputTokens
	cacheWriteTokens := usage.CacheCreationInputTokens
	outputTokens := usage.OutputTokens

	maxTokens := 10000
	if int64(inputTokens)+int64(cacheWriteTokens)+int64(outputTokens) > int64(maxTokens) {
		return "cache_cold"
	}

	return ""
}

func countWords(s string) int {
	count := 0
	inWord := false
	for _, r := range s {
		if unicode.IsSpace(r) {
			inWord = false
		} else if !inWord {
			inWord = true
			count++
		}
	}
	return count
}

// SuggestionMessage 建议消息
type SuggestionMessage struct {
	Type       string // "user", "assistant", "system"
	Content    string
	IsApiError bool
	Usage      *TokenUsage
}

// TokenUsage Token使用量
type TokenUsage struct {
	InputTokens              int64
	OutputTokens             int64
	CacheCreationInputTokens int64
}

// AppState 应用状态
type AppState struct {
	PendingWorkerRequest   bool
	PendingSandboxRequest  bool
	ElicitationQueueLength int
	ToolPermissionMode     string
}

// CacheSafeParams 缓存安全参数
type CacheSafeParams struct {
	SystemPrompt  string
	UserContext   string
	SystemContext string
}

// SuggestionResult 建议结果
type SuggestionResult struct {
	Suggestion          string
	PromptId            PromptVariant
	GenerationRequestId string
}
