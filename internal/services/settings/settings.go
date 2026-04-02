package settings

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Settings struct {
	mu          sync.RWMutex
	data        map[string]interface{}
	filePath    string
	modified    time.Time
	syncedAt    time.Time
	cloudSource string
}

func NewSettings(filePath string) *Settings {
	return &Settings{
		data:     make(map[string]interface{}),
		filePath: filePath,
	}
}

func (s *Settings) Load() error {
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
		return fmt.Errorf("read settings: %w", err)
	}

	if len(data) == 0 {
		return nil
	}

	if err := json.Unmarshal(data, &s.data); err != nil {
		return fmt.Errorf("unmarshal settings: %w", err)
	}

	return nil
}

func (s *Settings) Save() error {
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
		return fmt.Errorf("marshal settings: %w", err)
	}

	if err := os.WriteFile(s.filePath, data, 0644); err != nil {
		return fmt.Errorf("write settings: %w", err)
	}

	s.modified = time.Now()
	return nil
}

func (s *Settings) Get(key string) (interface{}, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	val, ok := s.data[key]
	return val, ok
}

func (s *Settings) Set(key string, value interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
	s.modified = time.Now()
}

func (s *Settings) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, key)
	s.modified = time.Now()
}

func (s *Settings) GetString(key, defaultValue string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if val, ok := s.data[key].(string); ok {
		return val
	}
	return defaultValue
}

func (s *Settings) GetInt(key string, defaultValue int) int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if val, ok := s.data[key].(float64); ok {
		return int(val)
	}
	return defaultValue
}

func (s *Settings) GetBool(key string, defaultValue bool) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if val, ok := s.data[key].(bool); ok {
		return val
	}
	return defaultValue
}

func (s *Settings) SetDefaults(defaults map[string]interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for key, value := range defaults {
		if _, exists := s.data[key]; !exists {
			s.data[key] = value
		}
	}
}

func (s *Settings) All() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string]interface{})
	for k, v := range s.data {
		result[k] = v
	}
	return result
}

func (s *Settings) IsModified() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return !s.modified.IsZero()
}

func (s *Settings) LastModified() time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.modified
}

func (s *Settings) MarkSynced() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.syncedAt = time.Now()
}

func (s *Settings) LastSynced() time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.syncedAt
}

type SettingsSyncService struct {
	mu       sync.RWMutex
	local    *Settings
	remote   *Settings
	provider string
}

func NewSettingsSyncService(localPath, remotePath, provider string) *SettingsSyncService {
	return &SettingsSyncService{
		local:    NewSettings(localPath),
		remote:   NewSettings(remotePath),
		provider: provider,
	}
}

func (s *SettingsSyncService) Load() error {
	if err := s.local.Load(); err != nil {
		return err
	}
	s.remote.Load()
	return nil
}

func (s *SettingsSyncService) Save() error {
	return s.local.Save()
}

func (s *SettingsSyncService) SyncToCloud() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := json.MarshalIndent(s.local.All(), "", "  ")
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	return os.WriteFile(s.remote.filePath, data, 0644)
}

func (s *SettingsSyncService) SyncFromCloud() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	remoteData, err := os.ReadFile(s.remote.filePath)
	if err != nil {
		return fmt.Errorf("read remote: %w", err)
	}

	var data map[string]interface{}
	if err := json.Unmarshal(remoteData, &data); err != nil {
		return fmt.Errorf("unmarshal remote: %w", err)
	}

	s.local.mu.Lock()
	s.local.data = data
	s.local.mu.Unlock()

	return nil
}

func (s *SettingsSyncService) GetLocal() *Settings {
	return s.local
}

func (s *SettingsSyncService) GetRemote() *Settings {
	return s.remote
}

type RemoteManagedSettings struct {
	mu        sync.RWMutex
	settings  map[string]interface{}
	serverURL string
	enabled   bool
}

func NewRemoteManagedSettings(serverURL string) *RemoteManagedSettings {
	return &RemoteManagedSettings{
		settings:  make(map[string]interface{}),
		serverURL: serverURL,
		enabled:   false,
	}
}

func (r *RemoteManagedSettings) Enable() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.enabled = true
}

func (r *RemoteManagedSettings) Disable() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.enabled = false
}

func (r *RemoteManagedSettings) IsEnabled() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.enabled
}

func (r *RemoteManagedSettings) Get(key string) (interface{}, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	val, ok := r.settings[key]
	return val, ok
}

func (r *RemoteManagedSettings) SetRemoteSettings(settings map[string]interface{}) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.settings = settings
}

func (r *RemoteManagedSettings) MergeRemoteSettings(settings map[string]interface{}) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for k, v := range settings {
		r.settings[k] = v
	}
}

func (r *RemoteManagedSettings) GetAll() map[string]interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[string]interface{})
	for k, v := range r.settings {
		result[k] = v
	}
	return result
}
