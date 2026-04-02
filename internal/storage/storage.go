package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Store struct {
	mu       sync.RWMutex
	data     map[string]interface{}
	filePath string
}

func NewStore(filePath string) *Store {
	return &Store{
		data:     make(map[string]interface{}),
		filePath: filePath,
	}
}

func (s *Store) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.filePath == "" {
		return nil
	}

	data, err := os.ReadFile(s.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read file: %w", err)
	}

	if len(data) == 0 {
		return nil
	}

	if err := json.Unmarshal(data, &s.data); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

	return nil
}

func (s *Store) Save() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.filePath == "" {
		return nil
	}

	dir := filepath.Dir(s.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create dir: %w", err)
	}

	data, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	if err := os.WriteFile(s.filePath, data, 0644); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	return nil
}

func (s *Store) Get(key string) (interface{}, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	val, ok := s.data[key]
	return val, ok
}

func (s *Store) Set(key string, value interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
}

func (s *Store) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, key)
}

func (s *Store) Keys() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	keys := make([]string, 0, len(s.data))
	for k := range s.data {
		keys = append(keys, k)
	}
	return keys
}

func (s *Store) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data = make(map[string]interface{})
}

type SessionStore struct {
	mu       sync.RWMutex
	sessions map[string]*Session
}

type Session struct {
	ID        string                 `json:"id"`
	Messages  []SessionMessage       `json:"messages"`
	Metadata  map[string]interface{} `json:"metadata"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

type SessionMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func NewSessionStore() *SessionStore {
	return &SessionStore{
		sessions: make(map[string]*Session),
	}
}

func (s *SessionStore) Create(id string) *Session {
	s.mu.Lock()
	defer s.mu.Unlock()

	session := &Session{
		ID:        id,
		Messages:  make([]SessionMessage, 0),
		Metadata:  make(map[string]interface{}),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	s.sessions[id] = session
	return session
}

func (s *SessionStore) Get(id string) (*Session, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	session, ok := s.sessions[id]
	return session, ok
}

func (s *SessionStore) Delete(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, id)
}

func (s *SessionStore) AddMessage(sessionID string, role, content string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, ok := s.sessions[sessionID]
	if !ok {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	session.Messages = append(session.Messages, SessionMessage{
		Role:    role,
		Content: content,
	})
	session.UpdatedAt = time.Now()
	return nil
}

func (s *SessionStore) List() []*Session {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sessions := make([]*Session, 0, len(s.sessions))
	for _, session := range s.sessions {
		sessions = append(sessions, session)
	}
	return sessions
}

func (s *SessionStore) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.sessions)
}

type ConfigStore struct {
	*Store
}

func NewConfigStore() *ConfigStore {
	homeDir, _ := os.UserHomeDir()
	configPath := filepath.Join(homeDir, ".config", "claude-code-go", "config.json")

	store := NewStore(configPath)
	if err := store.Load(); err != nil {
		fmt.Printf("Warning: failed to load config: %v\n", err)
	}

	return &ConfigStore{Store: store}
}

func (c *ConfigStore) GetString(key, defaultValue string) string {
	val, ok := c.Get(key)
	if ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return defaultValue
}

func (c *ConfigStore) GetInt(key string, defaultValue int) int {
	val, ok := c.Get(key)
	if ok {
		if f, ok := val.(float64); ok {
			return int(f)
		}
	}
	return defaultValue
}

func (c *ConfigStore) GetBool(key string, defaultValue bool) bool {
	val, ok := c.Get(key)
	if ok {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return defaultValue
}

func (c *ConfigStore) SetDefaults(defaults map[string]interface{}) {
	for key, value := range defaults {
		if _, exists := c.Get(key); !exists {
			c.Set(key, value)
		}
	}
}
