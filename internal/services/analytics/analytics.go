package analytics

import (
	"encoding/json"
	"sync"
	"time"
)

type Event struct {
	Name       string                 `json:"name"`
	Properties map[string]interface{} `json:"properties,omitempty"`
	Timestamp  time.Time              `json:"timestamp"`
}

type Analytics struct {
	enabled   bool
	events    []Event
	mu        sync.RWMutex
	userID    string
	sessionID string
}

func NewAnalytics() *Analytics {
	return &Analytics{
		enabled:   true,
		events:    make([]Event, 0),
		sessionID: generateSessionID(),
	}
}

func generateSessionID() string {
	return time.Now().Format("20060102150405")
}

func (a *Analytics) SetEnabled(enabled bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.enabled = enabled
}

func (a *Analytics) IsEnabled() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.enabled
}

func (a *Analytics) SetUserID(userID string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.userID = userID
}

func (a *Analytics) GetUserID() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.userID
}

func (a *Analytics) GetSessionID() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.sessionID
}

func (a *Analytics) Track(eventName string, properties map[string]interface{}) {
	if !a.IsEnabled() {
		return
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	event := Event{
		Name:       eventName,
		Properties: properties,
		Timestamp:  time.Now(),
	}

	a.events = append(a.events, event)
}

func (a *Analytics) GetEvents() []Event {
	a.mu.RLock()
	defer a.mu.RUnlock()

	events := make([]Event, len(a.events))
	copy(events, a.events)
	return events
}

func (a *Analytics) ClearEvents() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.events = make([]Event, 0)
}

type EventNames struct{}

var EventName = EventNames{}

func (e *EventNames) ToolInvoked() string      { return "tool_invoked" }
func (e *EventNames) CommandExecuted() string  { return "command_executed" }
func (e *EventNames) SessionStarted() string   { return "session_started" }
func (e *EventNames) SessionEnded() string     { return "session_ended" }
func (e *EventNames) ErrorOccurred() string    { return "error_occurred" }
func (e *EventNames) ModelChanged() string     { return "model_changed" }
func (e *EventNames) CompactTriggered() string { return "compact_triggered" }

type FeatureFlags struct {
	flags map[string]bool
	mu    sync.RWMutex
}

func NewFeatureFlags() *FeatureFlags {
	return &FeatureFlags{
		flags: map[string]bool{
			"EXPERIMENTAL_SKILL_SEARCH": false,
			"VOICE_MODE":                false,
			"BRIDGE_MODE":               false,
			"DAEMON":                    false,
			"PROACTIVE":                 false,
		},
	}
}

func (f *FeatureFlags) IsEnabled(name string) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.flags[name]
}

func (f *FeatureFlags) SetEnabled(name string, enabled bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.flags[name] = enabled
}

func (f *FeatureFlags) List() map[string]bool {
	f.mu.RLock()
	defer f.mu.RUnlock()

	result := make(map[string]bool)
	for k, v := range f.flags {
		result[k] = v
	}
	return result
}

func (f *FeatureFlags) MarshalJSON() ([]byte, error) {
	type Alias FeatureFlags
	return json.Marshal(&struct {
		Type  string          `json:"type"`
		Flags map[string]bool `json:"flags"`
		*Alias
	}{
		Type:  "FeatureFlags",
		Flags: f.List(),
		Alias: (*Alias)(f),
	})
}
