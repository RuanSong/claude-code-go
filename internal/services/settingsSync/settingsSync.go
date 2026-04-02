package settingsSync

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Setting struct {
	Key       string      `json:"key"`
	Value     interface{} `json:"value"`
	UpdatedAt time.Time   `json:"updated_at"`
}

type SettingsSyncService struct {
	mu        sync.RWMutex
	settings  map[string]*Setting
	changes   chan *Setting
	stopChan  chan struct{}
	isRunning bool
}

var (
	instance     *SettingsSyncService
	instanceOnce sync.Once
)

func GetSettingsSyncService() *SettingsSyncService {
	instanceOnce.Do(func() {
		instance = &SettingsSyncService{
			settings: make(map[string]*Setting),
			changes:  make(chan *Setting, 100),
			stopChan: make(chan struct{}),
		}
	})
	return instance
}

func (s *SettingsSyncService) Set(key string, value interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.settings[key] = &Setting{
		Key:       key,
		Value:     value,
		UpdatedAt: time.Now(),
	}

	select {
	case s.changes <- s.settings[key]:
	default:
	}
}

func (s *SettingsSyncService) Get(key string) (interface{}, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	setting, exists := s.settings[key]
	if !exists {
		return nil, false
	}
	return setting.Value, true
}

func (s *SettingsSyncService) GetString(key string) (string, bool) {
	value, exists := s.Get(key)
	if !exists {
		return "", false
	}
	str, ok := value.(string)
	return str, ok
}

func (s *SettingsSyncService) GetInt(key string) (int, bool) {
	value, exists := s.Get(key)
	if !exists {
		return 0, false
	}
	switch v := value.(type) {
	case int:
		return v, true
	case float64:
		return int(v), true
	}
	return 0, false
}

func (s *SettingsSyncService) GetBool(key string) (bool, bool) {
	value, exists := s.Get(key)
	if !exists {
		return false, false
	}
	b, ok := value.(bool)
	return b, ok
}

func (s *SettingsSyncService) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.settings, key)
}

func (s *SettingsSyncService) GetAllSettings() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string]interface{})
	for k, v := range s.settings {
		result[k] = v.Value
	}
	return result
}

func (s *SettingsSyncService) SaveToFile(filePath string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data := make(map[string]interface{})
	for k, v := range s.settings {
		data[k] = map[string]interface{}{
			"value":      v.Value,
			"updated_at": v.UpdatedAt.Format(time.RFC3339),
		}
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(filePath, jsonData, 0644)
}

func (s *SettingsSyncService) LoadFromFile(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	var rawData map[string]json.RawMessage
	if err := json.Unmarshal(data, &rawData); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for k, v := range rawData {
		var setting map[string]interface{}
		if err := json.Unmarshal(v, &setting); err != nil {
			continue
		}

		value, ok := setting["value"]
		if !ok {
			continue
		}

		updatedAt := time.Now()
		if updatedAtStr, ok := setting["updated_at"].(string); ok {
			if t, err := time.Parse(time.RFC3339, updatedAtStr); err == nil {
				updatedAt = t
			}
		}

		s.settings[k] = &Setting{
			Key:       k,
			Value:     value,
			UpdatedAt: updatedAt,
		}
	}

	return nil
}

func (s *SettingsSyncService) StartAutoSync(interval time.Duration) {
	s.mu.Lock()
	if s.isRunning {
		s.mu.Unlock()
		return
	}
	s.isRunning = true
	s.mu.Unlock()

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-s.stopChan:
				return
			case <-ticker.C:
				s.syncNow()
			}
		}
	}()
}

func (s *SettingsSyncService) StopAutoSync() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.isRunning {
		return
	}
	s.isRunning = false
	close(s.stopChan)
	s.stopChan = make(chan struct{})
}

func (s *SettingsSyncService) syncNow() {
}

func (s *SettingsSyncService) SubscribeChanges() <-chan *Setting {
	return s.changes
}

func (s *SettingsSyncService) GetStats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return map[string]interface{}{
		"total_settings": len(s.settings),
		"is_syncing":     s.isRunning,
	}
}
