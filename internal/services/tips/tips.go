package tips

import (
	"math/rand"
	"sync"
)

type Tip struct {
	ID       string
	Title    string
	Content  string
	Category string
}

type TipsService struct {
	mu   sync.RWMutex
	tips []*Tip
}

var (
	instance     *TipsService
	instanceOnce sync.Once
)

func GetTipsService() *TipsService {
	instanceOnce.Do(func() {
		instance = &TipsService{
			tips: getDefaultTips(),
		}
	})
	return instance
}

func (s *TipsService) GetAllTips() []*Tip {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*Tip, len(s.tips))
	copy(result, s.tips)
	return result
}

func (s *TipsService) GetRandomTip() *Tip {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.tips) == 0 {
		return nil
	}
	return s.tips[rand.Intn(len(s.tips))]
}

func (s *TipsService) GetTipsByCategory(category string) []*Tip {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*Tip
	for _, tip := range s.tips {
		if tip.Category == category {
			result = append(result, tip)
		}
	}
	return result
}

func (s *TipsService) AddTip(title, content, category string) string {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := generateID()
	s.tips = append(s.tips, &Tip{
		ID:       id,
		Title:    title,
		Content:  content,
		Category: category,
	})
	return id
}

func (s *TipsService) RemoveTip(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, tip := range s.tips {
		if tip.ID == id {
			s.tips = append(s.tips[:i], s.tips[i+1:]...)
			return true
		}
	}
	return false
}

func (s *TipsService) GetCategories() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	categorySet := make(map[string]bool)
	for _, tip := range s.tips {
		categorySet[tip.Category] = true
	}

	categories := make([]string, 0, len(categorySet))
	for category := range categorySet {
		categories = append(categories, category)
	}
	return categories
}

func generateID() string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	id := make([]byte, 8)
	for i := range id {
		id[i] = chars[rand.Intn(len(chars))]
	}
	return string(id)
}

func getDefaultTips() []*Tip {
	return []*Tip{
		{
			ID:       "1",
			Title:    "Use /compact to save context",
			Content:  "Running /compact periodically helps maintain conversation context by summarizing earlier messages.",
			Category: "productivity",
		},
		{
			ID:       "2",
			Title:    "Tab completion works in chat",
			Content:  "Press Tab to autocomplete commands, file paths, and tool names.",
			Category: "navigation",
		},
		{
			ID:       "3",
			Title:    "Use /resume to continue work",
			Content:  "If you restart Claude Code, use /resume to continue your previous session.",
			Category: "workflow",
		},
		{
			ID:       "4",
			Title:    "MCP tools extend capabilities",
			Content:  "Configure MCP servers in ~/.claude/settings.json to add custom tools and integrations.",
			Category: "extensibility",
		},
		{
			ID:       "5",
			Title:    "Review changes before committing",
			Content:  "Use /diff to review uncommitted changes before creating a commit.",
			Category: "git",
		},
		{
			ID:       "6",
			Title:    "Debug with /bughunter",
			Content:  "The /bughunter command helps investigate and debug issues systematically.",
			Category: "debugging",
		},
		{
			ID:       "7",
			Title:    "Use @mentions for context",
			Content:  "Use @ to reference files, documents, or specific code sections in your prompt.",
			Category: "productivity",
		},
		{
			ID:       "8",
			Title:    "Memory persists across sessions",
			Content:  "Claude Code remembers important information across sessions. Use /memory to manage what it remembers.",
			Category: "memory",
		},
		{
			ID:       "9",
			Title:    "Rate limits can be managed",
			Content:  "Use /rate-limit-options to configure how Claude Code handles API rate limits.",
			Category: "settings",
		},
		{
			ID:       "10",
			Title:    "Voice mode available",
			Content:  "Use /voice to enable voice interaction with Claude Code.",
			Category: "interaction",
		},
	}
}
