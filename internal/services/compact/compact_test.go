package compact

import (
	"testing"

	"github.com/claude-code-go/claude/internal/engine"
)

func TestNewCompactService(t *testing.T) {
	engine := &engine.QueryEngine{}
	service := NewCompactService(engine)

	if service == nil {
		t.Error("NewCompactService() returned nil")
	}

	if service.engine != engine {
		t.Error("NewCompactService() did not set engine correctly")
	}
}

func TestNewCompactConfig(t *testing.T) {
	config := NewCompactConfig()

	if config == nil {
		t.Error("NewCompactConfig() returned nil")
	}

	if !config.Enabled {
		t.Error("NewCompactConfig() should enable compaction by default")
	}

	if !config.AutoCompactEnabled {
		t.Error("NewCompactConfig() should enable auto-compaction by default")
	}
}

func TestCompactConfig_SetThreshold(t *testing.T) {
	config := NewCompactConfig()

	config.WarningThreshold = 100000
	if config.WarningThreshold != 100000 {
		t.Errorf("WarningThreshold = %d, want 100000", config.WarningThreshold)
	}

	config.MaxTokens = 200000
	if config.MaxTokens != 200000 {
		t.Errorf("MaxTokens = %d, want 200000", config.MaxTokens)
	}
}
