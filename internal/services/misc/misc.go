package services

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

type AgentSummary struct {
	AgentID   string                 `json:"agentId"`
	Summary   string                 `json:"summary"`
	ToolsUsed []string               `json:"toolsUsed"`
	Messages  int                    `json:"messages"`
	StartTime time.Time              `json:"startTime"`
	EndTime   time.Time              `json:"endTime,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

type AgentSummaryService struct {
	mu        sync.RWMutex
	summaries map[string]*AgentSummary
}

func NewAgentSummaryService() *AgentSummaryService {
	return &AgentSummaryService{
		summaries: make(map[string]*AgentSummary),
	}
}

func (s *AgentSummaryService) CreateSummary(agentID string, summary string) *AgentSummary {
	s.mu.Lock()
	defer s.mu.Unlock()

	agentSummary := &AgentSummary{
		AgentID:   agentID,
		Summary:   summary,
		ToolsUsed: make([]string, 0),
		StartTime: time.Now(),
		Metadata:  make(map[string]interface{}),
	}
	s.summaries[agentID] = agentSummary
	return agentSummary
}

func (s *AgentSummaryService) AddToolUsed(agentID, toolName string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if summary, exists := s.summaries[agentID]; exists {
		summary.ToolsUsed = append(summary.ToolsUsed, toolName)
	}
}

func (s *AgentSummaryService) CompleteSummary(agentID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if summary, exists := s.summaries[agentID]; exists {
		summary.EndTime = time.Now()
	}
}

func (s *AgentSummaryService) GetSummary(agentID string) (*AgentSummary, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	summary, ok := s.summaries[agentID]
	return summary, ok
}

func (s *AgentSummaryService) ListSummaries() []*AgentSummary {
	s.mu.RLock()
	defer s.mu.RUnlock()

	summaries := make([]*AgentSummary, 0, len(s.summaries))
	for _, summary := range s.summaries {
		summaries = append(summaries, summary)
	}
	return summaries
}

func (s *AgentSummaryService) ToJSON() ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return json.MarshalIndent(s.summaries, "", "  ")
}

type SessionMemory struct {
	SessionID string          `json:"sessionId"`
	Messages  []MemoryMessage `json:"messages"`
	KeyFacts  []string        `json:"keyFacts"`
	CreatedAt time.Time       `json:"createdAt"`
	UpdatedAt time.Time       `json:"updatedAt"`
}

type MemoryMessage struct {
	Role    string    `json:"role"`
	Content string    `json:"content"`
	Time    time.Time `json:"time"`
}

type SessionMemoryService struct {
	mu       sync.RWMutex
	memories map[string]*SessionMemory
}

func NewSessionMemoryService() *SessionMemoryService {
	return &SessionMemoryService{
		memories: make(map[string]*SessionMemory),
	}
}

func (s *SessionMemoryService) CreateSession(sessionID string) *SessionMemory {
	s.mu.Lock()
	defer s.mu.Unlock()

	memory := &SessionMemory{
		SessionID: sessionID,
		Messages:  make([]MemoryMessage, 0),
		KeyFacts:  make([]string, 0),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	s.memories[sessionID] = memory
	return memory
}

func (s *SessionMemoryService) AddMessage(sessionID, role, content string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if memory, exists := s.memories[sessionID]; exists {
		memory.Messages = append(memory.Messages, MemoryMessage{
			Role:    role,
			Content: content,
			Time:    time.Now(),
		})
		memory.UpdatedAt = time.Now()
	}
}

func (s *SessionMemoryService) AddKeyFact(sessionID, fact string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if memory, exists := s.memories[sessionID]; exists {
		memory.KeyFacts = append(memory.KeyFacts, fact)
		memory.UpdatedAt = time.Now()
	}
}

func (s *SessionMemoryService) GetSession(sessionID string) (*SessionMemory, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	memory, ok := s.memories[sessionID]
	return memory, ok
}

func (s *SessionMemoryService) GetOrCreateSession(sessionID string) *SessionMemory {
	s.mu.Lock()
	defer s.mu.Unlock()

	if memory, exists := s.memories[sessionID]; exists {
		return memory
	}

	memory := &SessionMemory{
		SessionID: sessionID,
		Messages:  make([]MemoryMessage, 0),
		KeyFacts:  make([]string, 0),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	s.memories[sessionID] = memory
	return memory
}

type TipsService struct {
	mu    sync.RWMutex
	tips  []Tip
	shown map[string]bool
}

type Tip struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Content  string `json:"content"`
	Category string `json:"category"`
}

func NewTipsService() *TipsService {
	return &TipsService{
		tips:  defaultTips(),
		shown: make(map[string]bool),
	}
}

func defaultTips() []Tip {
	return []Tip{
		{ID: "1", Title: "Tab Completion", Content: "Press Tab for auto-completion", Category: "navigation"},
		{ID: "2", Title: "Command History", Content: "Use ↑/↓ to navigate history", Category: "navigation"},
		{ID: "3", Title: "Slash Commands", Content: "Type / to see available commands", Category: "commands"},
	}
}

func (s *TipsService) GetRandomTip() *Tip {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, tip := range s.tips {
		if !s.shown[tip.ID] {
			return &tip
		}
	}
	return &s.tips[0]
}

func (s *TipsService) MarkTipShown(tipID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.shown[tipID] = true
}

func (s *TipsService) ListTips() []Tip {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.tips
}

type NotifierService struct {
	mu            sync.RWMutex
	notifications []Notification
}

type Notification struct {
	ID      string    `json:"id"`
	Type    string    `json:"type"`
	Title   string    `json:"title"`
	Message string    `json:"message"`
	Time    time.Time `json:"time"`
	Read    bool      `json:"read"`
}

func NewNotifierService() *NotifierService {
	return &NotifierService{
		notifications: make([]Notification, 0),
	}
}

func (s *NotifierService) Notify(notifType, title, message string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	notif := Notification{
		ID:      fmt.Sprintf("%d", time.Now().UnixNano()),
		Type:    notifType,
		Title:   title,
		Message: message,
		Time:    time.Now(),
		Read:    false,
	}
	s.notifications = append(s.notifications, notif)
}

func (s *NotifierService) GetNotifications() []Notification {
	s.mu.RLock()
	defer s.mu.RUnlock()

	notifications := make([]Notification, len(s.notifications))
	copy(notifications, s.notifications)
	return notifications
}

func (s *NotifierService) MarkRead(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.notifications {
		if s.notifications[i].ID == id {
			s.notifications[i].Read = true
			return
		}
	}
}

func (s *NotifierService) ClearAll() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.notifications = make([]Notification, 0)
}

type DiagnosticTracking struct {
	mu          sync.RWMutex
	diagnostics []Diagnostic
}

type Diagnostic struct {
	Severity string    `json:"severity"`
	Source   string    `json:"source"`
	Message  string    `json:"message"`
	Location string    `json:"location"`
	Time     time.Time `json:"time"`
}

func NewDiagnosticTracking() *DiagnosticTracking {
	return &DiagnosticTracking{
		diagnostics: make([]Diagnostic, 0),
	}
}

func (t *DiagnosticTracking) AddDiagnostic(severity, source, message, location string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.diagnostics = append(t.diagnostics, Diagnostic{
		Severity: severity,
		Source:   source,
		Message:  message,
		Location: location,
		Time:     time.Now(),
	})
}

func (t *DiagnosticTracking) GetDiagnostics() []Diagnostic {
	t.mu.RLock()
	defer t.mu.RUnlock()

	diagnostics := make([]Diagnostic, len(t.diagnostics))
	copy(diagnostics, t.diagnostics)
	return diagnostics
}

func (t *DiagnosticTracking) Clear() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.diagnostics = make([]Diagnostic, 0)
}
