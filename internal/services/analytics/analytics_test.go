package analytics

import (
	"testing"
	"time"
)

func TestAnalyticsCreation(t *testing.T) {
	analytics := NewAnalytics()
	if analytics == nil {
		t.Fatal("NewAnalytics should not return nil")
	}
	if analytics.IsEnabled() != true {
		t.Error("Analytics should be enabled by default")
	}
}

func TestAnalyticsSetEnabled(t *testing.T) {
	analytics := NewAnalytics()

	analytics.SetEnabled(false)
	if analytics.IsEnabled() != false {
		t.Error("Analytics should be disabled after SetEnabled(false)")
	}

	analytics.SetEnabled(true)
	if analytics.IsEnabled() != true {
		t.Error("Analytics should be enabled after SetEnabled(true)")
	}
}

func TestAnalyticsUserID(t *testing.T) {
	analytics := NewAnalytics()

	analytics.SetUserID("user123")
	if analytics.GetUserID() != "user123" {
		t.Errorf("Expected userID 'user123', got '%s'", analytics.GetUserID())
	}
}

func TestAnalyticsSessionID(t *testing.T) {
	analytics := NewAnalytics()
	sessionID := analytics.GetSessionID()
	if sessionID == "" {
		t.Error("SessionID should not be empty")
	}
}

func TestAnalyticsTrack(t *testing.T) {
	analytics := NewAnalytics()
	analytics.SetEnabled(true)

	analytics.Track("test_event", map[string]interface{}{
		"key": "value",
	})

	events := analytics.GetEvents()
	if len(events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(events))
	}

	if events[0].Name != "test_event" {
		t.Errorf("Expected event name 'test_event', got '%s'", events[0].Name)
	}
}

func TestAnalyticsTrackDisabled(t *testing.T) {
	analytics := NewAnalytics()
	analytics.SetEnabled(false)

	analytics.Track("test_event", nil)

	events := analytics.GetEvents()
	if len(events) != 0 {
		t.Error("Should not track events when disabled")
	}
}

func TestAnalyticsClearEvents(t *testing.T) {
	analytics := NewAnalytics()
	analytics.SetEnabled(true)

	analytics.Track("event1", nil)
	analytics.Track("event2", nil)

	analytics.ClearEvents()

	events := analytics.GetEvents()
	if len(events) != 0 {
		t.Errorf("Expected 0 events after clear, got %d", len(events))
	}
}

func TestEventNames(t *testing.T) {
	en := &EventNames{}

	if en.ToolInvoked() != "tool_invoked" {
		t.Error("ToolInvoked should return 'tool_invoked'")
	}
	if en.CommandExecuted() != "command_executed" {
		t.Error("CommandExecuted should return 'command_executed'")
	}
	if en.SessionStarted() != "session_started" {
		t.Error("SessionStarted should return 'session_started'")
	}
	if en.SessionEnded() != "session_ended" {
		t.Error("SessionEnded should return 'session_ended'")
	}
	if en.ErrorOccurred() != "error_occurred" {
		t.Error("ErrorOccurred should return 'error_occurred'")
	}
	if en.ModelChanged() != "model_changed" {
		t.Error("ModelChanged should return 'model_changed'")
	}
	if en.CompactTriggered() != "compact_triggered" {
		t.Error("CompactTriggered should return 'compact_triggered'")
	}
}

func TestFeatureFlags(t *testing.T) {
	flags := NewFeatureFlags()

	if flags.IsEnabled("EXPERIMENTAL_SKILL_SEARCH") != false {
		t.Error("EXPERIMENTAL_SKILL_SEARCH should be disabled by default")
	}

	flags.SetEnabled("EXPERIMENTAL_SKILL_SEARCH", true)
	if flags.IsEnabled("EXPERIMENTAL_SKILL_SEARCH") != true {
		t.Error("EXPERIMENTAL_SKILL_SEARCH should be enabled after SetEnabled")
	}
}

func TestFeatureFlagsList(t *testing.T) {
	flags := NewFeatureFlags()

	list := flags.List()
	if len(list) != 5 {
		t.Errorf("Expected 5 flags, got %d", len(list))
	}

	if _, ok := list["EXPERIMENTAL_SKILL_SEARCH"]; !ok {
		t.Error("EXPERIMENTAL_SKILL_SEARCH should be in list")
	}
	if _, ok := list["VOICE_MODE"]; !ok {
		t.Error("VOICE_MODE should be in list")
	}
}

func TestFeatureFlagsMarshalJSON(t *testing.T) {
	flags := NewFeatureFlags()
	flags.SetEnabled("TEST_FLAG", true)

	data, err := flags.MarshalJSON()
	if err != nil {
		t.Errorf("MarshalJSON failed: %v", err)
	}

	if data == nil {
		t.Error("MarshalJSON should return data")
	}
}

func TestAnalyticsGetEvents(t *testing.T) {
	analytics := NewAnalytics()
	analytics.SetEnabled(true)

	analytics.Track("event1", map[string]interface{}{"prop": "value1"})
	analytics.Track("event2", map[string]interface{}{"prop": "value2"})

	events := analytics.GetEvents()
	if len(events) != 2 {
		t.Errorf("Expected 2 events, got %d", len(events))
	}

	if events[0].Properties["prop"] != "value1" {
		t.Error("First event should have prop=value1")
	}
	if events[1].Properties["prop"] != "value2" {
		t.Error("Second event should have prop=value2")
	}
}

func TestAnalyticsEventTimestamp(t *testing.T) {
	analytics := NewAnalytics()
	analytics.SetEnabled(true)

	before := time.Now()
	analytics.Track("test_event", nil)
	after := time.Now()

	events := analytics.GetEvents()
	if len(events) != 1 {
		t.Fatal("Should have 1 event")
	}

	eventTime := events[0].Timestamp
	if eventTime.Before(before) || eventTime.After(after) {
		t.Error("Event timestamp should be between before and after")
	}
}
