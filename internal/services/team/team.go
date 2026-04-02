package team

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

type Memory struct {
	ID        string                 `json:"id"`
	Content   string                 `json:"content"`
	Type      string                 `json:"type"`
	CreatedBy string                 `json:"createdBy"`
	CreatedAt time.Time              `json:"createdAt"`
	Tags      []string               `json:"tags,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

type TeamMemorySync struct {
	mu        sync.RWMutex
	memories  map[string]*Memory
	serverURL string
	connected bool
}

func NewTeamMemorySync(serverURL string) *TeamMemorySync {
	return &TeamMemorySync{
		memories:  make(map[string]*Memory),
		serverURL: serverURL,
		connected: false,
	}
}

func (t *TeamMemorySync) Connect() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.connected = true
	return nil
}

func (t *TeamMemorySync) Disconnect() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.connected = false
}

func (t *TeamMemorySync) IsConnected() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.connected
}

func (t *TeamMemorySync) AddMemory(memory *Memory) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.connected {
		return fmt.Errorf("not connected to team server")
	}

	memory.ID = fmt.Sprintf("%d", time.Now().UnixNano())
	memory.CreatedAt = time.Now()
	t.memories[memory.ID] = memory
	return nil
}

func (t *TeamMemorySync) GetMemory(id string) (*Memory, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	memory, ok := t.memories[id]
	return memory, ok
}

func (t *TeamMemorySync) SearchMemories(query string) []*Memory {
	t.mu.RLock()
	defer t.mu.RUnlock()

	results := make([]*Memory, 0)
	for _, memory := range t.memories {
		if contains(memory.Content, query) {
			results = append(results, memory)
			continue
		}
		for _, tag := range memory.Tags {
			if contains(tag, query) {
				results = append(results, memory)
				break
			}
		}
	}
	return results
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsRecursive(s, substr))
}

func containsRecursive(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func (t *TeamMemorySync) ListMemories() []*Memory {
	t.mu.RLock()
	defer t.mu.RUnlock()

	memories := make([]*Memory, 0, len(t.memories))
	for _, memory := range t.memories {
		memories = append(memories, memory)
	}
	return memories
}

func (t *TeamMemorySync) DeleteMemory(id string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if _, exists := t.memories[id]; !exists {
		return fmt.Errorf("memory not found: %s", id)
	}

	delete(t.memories, id)
	return nil
}

func (t *TeamMemorySync) SyncToServer() error {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if !t.connected {
		return fmt.Errorf("not connected")
	}

	data, err := json.MarshalIndent(t.memories, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	_ = data
	return nil
}

func (t *TeamMemorySync) ToJSON() ([]byte, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return json.MarshalIndent(t.memories, "", "  ")
}

type ExtractMemoriesService struct {
	mu    sync.RWMutex
	rules []MemoryExtractionRule
}

type MemoryExtractionRule struct {
	Pattern     string   `json:"pattern"`
	Type        string   `json:"type"`
	Tags        []string `json:"tags,omitempty"`
	Description string   `json:"description,omitempty"`
}

func NewExtractMemoriesService() *ExtractMemoriesService {
	return &ExtractMemoriesService{
		rules: defaultExtractionRules(),
	}
}

func defaultExtractionRules() []MemoryExtractionRule {
	return []MemoryExtractionRule{
		{Pattern: "TODO:", Type: "task", Tags: []string{"task"}},
		{Pattern: "FIXME:", Type: "bug", Tags: []string{"bug", "fixme"}},
		{Pattern: "NOTE:", Type: "note", Tags: []string{"note"}},
		{Pattern: "IMPORTANT:", Type: "important", Tags: []string{"important"}},
	}
}

func (s *ExtractMemoriesService) ExtractMemories(content string) []*Memory {
	s.mu.RLock()
	defer s.mu.RUnlock()

	memories := make([]*Memory, 0)
	for _, rule := range s.rules {
		if contains(content, rule.Pattern) {
			memory := &Memory{
				ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
				Content:   rule.Description,
				Type:      rule.Type,
				Tags:      rule.Tags,
				CreatedAt: time.Now(),
			}
			memories = append(memories, memory)
		}
	}
	return memories
}

func (s *ExtractMemoriesService) AddRule(rule MemoryExtractionRule) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.rules = append(s.rules, rule)
}

func (s *ExtractMemoriesService) ListRules() []MemoryExtractionRule {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rules := make([]MemoryExtractionRule, len(s.rules))
	copy(rules, s.rules)
	return rules
}
