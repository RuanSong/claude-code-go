package compact

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/claude-code-go/claude/internal/engine"
)

const (
	AutoCompactBufferTokens      = 13000
	WarningThresholdBufferTokens = 20000
	MaxConsecutiveFailures       = 3
	CompactMaxOutputTokens       = 4096
	PostCompactMaxTokensPerFile  = 5000
	PostCompactMaxTokensPerSkill = 5000
)

type CompactionStrategy string

const (
	StrategyFull          CompactionStrategy = "full"
	StrategyPartial       CompactionStrategy = "partial"
	StrategyMicroCompact  CompactionStrategy = "micro"
	StrategySessionMemory CompactionStrategy = "session-memory"
)

type CompactionResult struct {
	BoundaryMarker    string             `json:"boundaryMarker"`
	SummaryMessages   []engine.Message   `json:"summaryMessages"`
	Attachments       []string           `json:"attachments"`
	MessagesRemoved   int                `json:"messagesRemoved"`
	PreCompactTokens  int                `json:"preCompactTokens"`
	PostCompactTokens int                `json:"postCompactTokens"`
	StrategyUsed      CompactionStrategy `json:"strategyUsed"`
	Success           bool               `json:"success"`
	Error             string             `json:"error,omitempty"`
}

type CompactOptions struct {
	Strategy           CompactionStrategy
	CustomInstructions string
	PivotIndex         int
	Direction          string
	SuppressFollowUp   bool
	IsAutoCompact      bool
}

type CompactionStats struct {
	TotalCompactions    int
	ConsecutiveFailures int
	LastCompactTime     time.Time
	LastCompactTokens   int
}

type CompactService struct {
	engine *engine.QueryEngine
	config *CompactConfig
	stats  *CompactionStats
	mu     sync.RWMutex
}

type CompactConfig struct {
	Enabled              bool
	AutoCompactEnabled   bool
	MicroCompactEnabled  bool
	SessionMemoryEnabled bool
	MaxTokens            int
	WarningThreshold     int
	GapThresholdMinutes  int
	KeepRecentCount      int
}

func NewCompactConfig() *CompactConfig {
	return &CompactConfig{
		Enabled:              true,
		AutoCompactEnabled:   true,
		MicroCompactEnabled:  true,
		SessionMemoryEnabled: false,
		MaxTokens:            150000,
		WarningThreshold:     130000,
		GapThresholdMinutes:  60,
		KeepRecentCount:      5,
	}
}

func NewCompactService(qe *engine.QueryEngine) *CompactService {
	return &CompactService{
		engine: qe,
		config: NewCompactConfig(),
		stats:  &CompactionStats{},
	}
}

func NewCompactServiceWithConfig(qe *engine.QueryEngine, config *CompactConfig) *CompactService {
	return &CompactService{
		engine: qe,
		config: config,
		stats: &CompactionStats{
			TotalCompactions:    0,
			ConsecutiveFailures: 0,
		},
	}
}

func (s *CompactService) Compact(ctx context.Context, opts *CompactOptions) (*CompactionResult, error) {
	if s.engine == nil || s.engine.GetContext() == nil {
		return nil, fmt.Errorf("no context available")
	}

	if opts == nil {
		opts = &CompactOptions{Strategy: StrategyFull}
	}

	s.mu.Lock()
	if s.stats.ConsecutiveFailures >= MaxConsecutiveFailures {
		s.mu.Unlock()
		return nil, fmt.Errorf("max consecutive compaction failures reached")
	}
	s.mu.Unlock()

	var result *CompactionResult
	var err error

	switch opts.Strategy {
	case StrategyFull:
		result, err = s.compactFull(ctx, opts)
	case StrategyPartial:
		result, err = s.compactPartial(ctx, opts)
	case StrategyMicroCompact:
		result, err = s.compactMicro(ctx)
	case StrategySessionMemory:
		result, err = s.compactSessionMemory(ctx, opts)
	default:
		result, err = s.compactFull(ctx, opts)
	}

	if err != nil {
		s.mu.Lock()
		s.stats.ConsecutiveFailures++
		s.mu.Unlock()
		return result, err
	}

	s.mu.Lock()
	s.stats.ConsecutiveFailures = 0
	s.stats.TotalCompactions++
	s.stats.LastCompactTime = time.Now()
	if result != nil {
		s.stats.LastCompactTokens = result.PostCompactTokens
	}
	s.mu.Unlock()

	return result, nil
}

func (s *CompactService) compactFull(ctx context.Context, opts *CompactOptions) (*CompactionResult, error) {
	context := s.engine.GetContext()
	messages := context.GetMessages()

	if len(messages) < 4 {
		return &CompactionResult{
			StrategyUsed: StrategyFull,
			Success:      false,
			Error:        "not enough messages to compact",
		}, nil
	}

	preTokens := context.CountTokens()

	summary, err := s.generateSummary(ctx, messages, opts)
	if err != nil {
		return &CompactionResult{
			StrategyUsed:     StrategyFull,
			PreCompactTokens: preTokens,
			Success:          false,
			Error:            err.Error(),
		}, err
	}

	keptMessages := s.getMessagesToKeep(messages, opts.PivotIndex)

	context.Clear()
	for _, msg := range keptMessages {
		context.AddMessage(msg)
	}

	boundaryMsg := &engine.Message{
		Role: "system",
		Content: []engine.ContentBlock{&engine.TextBlock{
			Type: "text",
			Text: "--- Conversation Compacted ---",
		}},
	}
	context.AddMessage(boundaryMsg)
	context.AddMessage(summary)

	postTokens := context.CountTokens()

	return &CompactionResult{
		BoundaryMarker:    "--- Conversation Compacted ---",
		SummaryMessages:   []engine.Message{*summary},
		MessagesRemoved:   len(messages) - len(keptMessages) - 2,
		PreCompactTokens:  preTokens,
		PostCompactTokens: postTokens,
		StrategyUsed:      StrategyFull,
		Success:           true,
	}, nil
}

func (s *CompactService) compactPartial(ctx context.Context, opts *CompactOptions) (*CompactionResult, error) {
	context := s.engine.GetContext()
	messages := context.GetMessages()

	if opts.PivotIndex <= 0 || opts.PivotIndex >= len(messages) {
		return s.compactFull(ctx, opts)
	}

	preTokens := context.CountTokens()

	var prefixMsgs, suffixMsgs []*engine.Message
	if opts.Direction == "from" {
		prefixMsgs = messages[:opts.PivotIndex]
		suffixMsgs = messages[opts.PivotIndex:]
	} else {
		prefixMsgs = messages[:opts.PivotIndex+1]
		suffixMsgs = messages[opts.PivotIndex+1:]
	}

	summary, err := s.generateSummary(ctx, prefixMsgs, opts)
	if err != nil {
		return nil, err
	}

	context.Clear()

	boundaryMsg := &engine.Message{
		Role: "system",
		Content: []engine.ContentBlock{&engine.TextBlock{
			Type: "text",
			Text: "--- Context Compacted (Partial) ---",
		}},
	}
	context.AddMessage(boundaryMsg)
	context.AddMessage(summary)

	for _, msg := range suffixMsgs {
		context.AddMessage(msg)
	}

	postTokens := context.CountTokens()

	return &CompactionResult{
		BoundaryMarker:    "--- Context Compacted (Partial) ---",
		SummaryMessages:   []engine.Message{*summary},
		MessagesRemoved:   len(prefixMsgs),
		PreCompactTokens:  preTokens,
		PostCompactTokens: postTokens,
		StrategyUsed:      StrategyPartial,
		Success:           true,
	}, nil
}

func (s *CompactService) compactMicro(ctx context.Context) (*CompactionResult, error) {
	context := s.engine.GetContext()
	messages := context.GetMessages()

	if len(messages) < 2 {
		return &CompactionResult{
			StrategyUsed: StrategyMicroCompact,
			Success:      true,
		}, nil
	}

	preTokens := context.CountTokens()
	messagesRemoved := 0

	keepCount := s.config.KeepRecentCount
	if keepCount < 0 {
		keepCount = 5
	}

	if len(messages) > keepCount {
		keptMessages := messages[len(messages)-keepCount:]

		context.Clear()
		for _, msg := range keptMessages {
			context.AddMessage(msg)
		}
		messagesRemoved = len(messages) - keepCount
	}

	postTokens := context.CountTokens()

	return &CompactionResult{
		MessagesRemoved:   messagesRemoved,
		PreCompactTokens:  preTokens,
		PostCompactTokens: postTokens,
		StrategyUsed:      StrategyMicroCompact,
		Success:           true,
	}, nil
}

func (s *CompactService) compactSessionMemory(ctx context.Context, opts *CompactOptions) (*CompactionResult, error) {
	return s.compactFull(ctx, opts)
}

func (s *CompactService) generateSummary(ctx context.Context, messages []*engine.Message, opts *CompactOptions) (*engine.Message, error) {
	var sb strings.Builder

	sb.WriteString("Conversation summary:\n\n")

	userCount := 0
	assistantCount := 0
	userMessages := []string{}
	assistantMessages := []string{}

	for _, msg := range messages {
		switch msg.Role {
		case "user":
			userCount++
			if userCount <= 5 {
				if text := s.extractText(msg); text != "" {
					userMessages = append(userMessages, text)
				}
			}
		case "assistant":
			assistantCount++
			if assistantCount <= 5 {
				if text := s.extractText(msg); text != "" {
					assistantMessages = append(assistantMessages, text)
				}
			}
		}
	}

	if len(userMessages) > 0 {
		sb.WriteString("User messages:\n")
		for _, m := range userMessages {
			sb.WriteString(fmt.Sprintf("  - %s\n", m))
		}
		sb.WriteString("\n")
	}

	if len(assistantMessages) > 0 {
		sb.WriteString("Assistant responses:\n")
		for _, m := range assistantMessages {
			sb.WriteString(fmt.Sprintf("  - %s\n", m))
		}
		sb.WriteString("\n")
	}

	sb.WriteString(fmt.Sprintf("(Total: %d user messages, %d assistant messages)\n", userCount, assistantCount))

	if opts != nil && opts.CustomInstructions != "" {
		sb.WriteString(fmt.Sprintf("\nCustom instructions: %s\n", opts.CustomInstructions))
	}

	summaryText := sb.String()

	return &engine.Message{
		Role: "user",
		Content: []engine.ContentBlock{&engine.TextBlock{
			Type: "text",
			Text: summaryText,
		}},
	}, nil
}

func (s *CompactService) extractText(msg *engine.Message) string {
	for _, block := range msg.Content {
		if tb, ok := block.(*engine.TextBlock); ok {
			return tb.Text
		}
	}
	return ""
}

func (s *CompactService) getMessagesToKeep(messages []*engine.Message, pivotIndex int) []*engine.Message {
	keepCount := 2
	if pivotIndex > 0 && pivotIndex < len(messages) {
		return messages[pivotIndex:]
	}

	if len(messages) <= keepCount {
		return messages
	}

	return messages[len(messages)-keepCount:]
}

func (s *CompactService) ShouldCompact(ctx context.Context) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.config.Enabled || !s.config.AutoCompactEnabled {
		return false
	}

	context := s.engine.GetContext()
	if context == nil {
		return false
	}

	tokens := context.CountTokens()
	return tokens > s.config.WarningThreshold
}

func (s *CompactService) ShouldAutoCompact(ctx context.Context) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.config.AutoCompactEnabled {
		return false
	}

	context := s.engine.GetContext()
	if context == nil {
		return false
	}

	tokens := context.CountTokens()
	threshold := s.config.MaxTokens - AutoCompactBufferTokens

	return tokens >= threshold
}

func (s *CompactService) GetStats() CompactionStats {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return *s.stats
}

func (s *CompactService) ResetStats() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.stats = &CompactionStats{}
}

func (s *CompactService) GetConfig() *CompactConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config
}

func (s *CompactService) SetConfig(config *CompactConfig) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.config = config
}

func (s *CompactService) CalculateTokenWarningState(tokens int) (warning bool, error bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if tokens >= s.config.MaxTokens {
		return true, true
	}

	warningThreshold := s.config.MaxTokens - AutoCompactBufferTokens
	errorThreshold := s.config.MaxTokens - WarningThresholdBufferTokens

	if tokens >= errorThreshold {
		return true, true
	}

	if tokens >= warningThreshold {
		return true, false
	}

	return false, false
}

func (s *CompactService) StripImagesFromMessages(messages []engine.Message) []engine.Message {
	stripped := make([]engine.Message, 0, len(messages))
	for _, msg := range messages {
		if msg.Role == "system" && strings.Contains(msg.Content[0].(*engine.TextBlock).Text, "image") {
			continue
		}
		stripped = append(stripped, msg)
	}
	return stripped
}

func (s *CompactService) EvaluateTimeBasedTrigger(lastAssistantTime time.Time) bool {
	s.mu.RLock()
	gapMinutes := s.config.GapThresholdMinutes
	s.mu.RUnlock()

	if gapMinutes <= 0 {
		return false
	}

	gap := time.Since(lastAssistantTime)
	return gap.Minutes() >= float64(gapMinutes)
}

type MicroCompactResult struct {
	MessagesRemoved int
	TokensFreed     int
	ShouldProceed   bool
}

func (s *CompactService) MicroCompactMessages(messages []engine.Message, lastAssistantTime time.Time) *MicroCompactResult {
	if !s.evaluateTimeBasedTrigger(lastAssistantTime) {
		return &MicroCompactResult{ShouldProceed: false}
	}

	s.mu.RLock()
	keepRecent := s.config.KeepRecentCount
	s.mu.RUnlock()

	if len(messages) <= keepRecent {
		return &MicroCompactResult{ShouldProceed: false}
	}

	return &MicroCompactResult{
		MessagesRemoved: len(messages) - keepRecent,
		TokensFreed:     0,
		ShouldProceed:   true,
	}
}

func (s *CompactService) evaluateTimeBasedTrigger(lastAssistantTime time.Time) bool {
	return s.EvaluateTimeBasedTrigger(lastAssistantTime)
}

func (s *CompactService) GroupMessagesByApiRound(messages []engine.Message) [][]engine.Message {
	if len(messages) == 0 {
		return nil
	}

	groups := [][]engine.Message{}
	currentGroup := []engine.Message{messages[0]}

	for i := 1; i < len(messages); i++ {
		msg := messages[i]
		if msg.Role == "assistant" && len(msg.Content) > 0 {
			if _, ok := msg.Content[0].(*engine.ToolUseBlock); ok {
				currentGroup = append(currentGroup, msg)
				groups = append(groups, currentGroup)
				currentGroup = []engine.Message{}
				continue
			}
		}
		currentGroup = append(currentGroup, msg)
	}

	if len(currentGroup) > 0 {
		groups = append(groups, currentGroup)
	}

	return groups
}

func (s *CompactService) AdjustIndexToPreserveApiInvariants(messages []engine.Message, startIndex int) int {
	if startIndex >= len(messages) {
		return startIndex
	}

	for i := startIndex; i < len(messages); i++ {
		msg := messages[i]
		if msg.Role == "user" {
			continue
		}
		if msg.Role == "assistant" {
			hasToolUse := false
			for _, block := range msg.Content {
				if _, ok := block.(*engine.ToolUseBlock); ok {
					hasToolUse = true
					break
				}
			}
			if hasToolUse {
				return i
			}
		}
	}

	return startIndex
}
