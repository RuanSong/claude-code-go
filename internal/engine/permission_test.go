package engine

import (
	"context"
	"errors"
	"testing"
)

func TestNewPermissionManager(t *testing.T) {
	pm := NewPermissionManager()

	if pm == nil {
		t.Fatal("NewPermissionManager should not return nil")
	}

	if pm.mode != PermissionModeDefault {
		t.Errorf("Expected default mode, got %v", pm.mode)
	}

	// Check default rules
	if len(pm.rules.AlwaysAllow) == 0 {
		t.Error("Expected AlwaysAllow to have default tools")
	}
	if len(pm.rules.AlwaysAsk) == 0 {
		t.Error("Expected AlwaysAsk to have default tools")
	}
}

func TestPermissionManagerCheckAlwaysAllow(t *testing.T) {
	pm := NewPermissionManager()

	// Read, Glob, Grep, LSP should always be allowed
	allowedTools := []string{"Read", "Glob", "Grep", "LSP"}

	for _, tool := range allowedTools {
		allowed, err := pm.Check(context.Background(), tool)
		if err != nil {
			t.Errorf("Check(%s) failed: %v", tool, err)
		}
		if !allowed {
			t.Errorf("Expected %s to be allowed", tool)
		}
	}
}

func TestPermissionManagerCheckAlwaysDeny(t *testing.T) {
	pm := NewPermissionManager()

	// No tools are in AlwaysDeny by default
	// This test verifies the structure is correct
	deniedTools := pm.rules.AlwaysDeny
	if len(deniedTools) != 0 {
		t.Errorf("Expected no always denied tools, got %d", len(deniedTools))
	}
}

func TestPermissionManagerCheckBypass(t *testing.T) {
	pm := NewPermissionManager()
	pm.SetMode(PermissionModeBypass)

	// In bypass mode, everything should be allowed
	allowed, err := pm.Check(context.Background(), "Bash")
	if err != nil {
		t.Errorf("Check failed: %v", err)
	}
	if !allowed {
		t.Error("Expected allowed in bypass mode")
	}
}

func TestPermissionManagerCheckYolo(t *testing.T) {
	pm := NewPermissionManager()
	pm.SetMode(PermissionModeYolo)

	// In yolo mode, everything should be allowed
	allowed, err := pm.Check(context.Background(), "Write")
	if err != nil {
		t.Errorf("Check failed: %v", err)
	}
	if !allowed {
		t.Error("Expected allowed in yolo mode")
	}
}

func TestPermissionManagerCheckPlan(t *testing.T) {
	pm := NewPermissionManager()
	pm.SetMode(PermissionModePlan)

	// In plan mode, high-risk tools should require permission
	// Use "Bash" which is not in AlwaysAllow list
	allowed, err := pm.Check(context.Background(), "Bash")
	if !errors.Is(err, ErrPermissionRequired) {
		t.Errorf("Expected ErrPermissionRequired, got %v", err)
	}
	if allowed {
		t.Error("Expected not allowed in plan mode")
	}
}

func TestPermissionManagerCheckAuto(t *testing.T) {
	pm := NewPermissionManager()
	pm.SetMode(PermissionModeAuto)

	// High risk tools should require permission
	highRiskTools := []string{"Bash", "Write", "Edit", "NotebookEdit"}

	for _, tool := range highRiskTools {
		allowed, err := pm.Check(context.Background(), tool)
		if !errors.Is(err, ErrPermissionRequired) {
			t.Errorf("Expected ErrPermissionRequired for %s, got %v", tool, err)
		}
		if allowed {
			t.Errorf("Expected %s to require permission in auto mode", tool)
		}
	}
}

func TestPermissionManagerSetMode(t *testing.T) {
	pm := NewPermissionManager()

	tests := []struct {
		mode PermissionModeType
		name string
	}{
		{PermissionModeDefault, "default"},
		{PermissionModePlan, "plan"},
		{PermissionModeBypass, "bypass"},
		{PermissionModeAuto, "auto"},
		{PermissionModeYolo, "yolo"},
	}

	for _, tt := range tests {
		pm.SetMode(tt.mode)
		if pm.GetMode() != tt.mode {
			t.Errorf("Expected mode %s, got %v", tt.name, pm.GetMode())
		}
	}
}

func TestPermissionManagerSetRules(t *testing.T) {
	pm := NewPermissionManager()

	newRules := PermissionRules{
		AlwaysAllow: []string{"Read", "Glob"},
		AlwaysDeny:  []string{"Bash"},
		AlwaysAsk:   []string{"Write"},
	}

	pm.SetRules(newRules)

	if len(pm.rules.AlwaysAllow) != 2 {
		t.Errorf("Expected 2 always allow rules, got %d", len(pm.rules.AlwaysAllow))
	}
	if len(pm.rules.AlwaysDeny) != 1 {
		t.Errorf("Expected 1 always deny rule, got %d", len(pm.rules.AlwaysDeny))
	}
	if len(pm.rules.AlwaysAsk) != 1 {
		t.Errorf("Expected 1 always ask rule, got %d", len(pm.rules.AlwaysAsk))
	}
}

func TestRiskAssessorAssess(t *testing.T) {
	ra := &RiskAssessor{}

	tests := []struct {
		tool     string
		expected RiskLevel
	}{
		{"Bash", RiskHigh},
		{"Write", RiskMedium},
		{"Edit", RiskMedium},
		{"NotebookEdit", RiskMedium},
		{"WebSearch", RiskMedium},
		{"WebFetch", RiskMedium},
		{"Read", RiskLow},
		{"Glob", RiskLow},
		{"Grep", RiskLow},
		{"UnknownTool", RiskMedium},
	}

	for _, tt := range tests {
		risk := ra.Assess(tt.tool, nil)
		if risk != tt.expected {
			t.Errorf("Expected %s to be %v, got %v", tt.tool, tt.expected, risk)
		}
	}
}

func TestPermissionModeString(t *testing.T) {
	tests := []struct {
		mode     PermissionModeType
		expected string
	}{
		{PermissionModeDefault, "default"},
		{PermissionModePlan, "plan"},
		{PermissionModeBypass, "bypass"},
		{PermissionModeAuto, "auto"},
		{PermissionModeYolo, "yolo"},
	}

	for _, tt := range tests {
		if tt.mode.String() != tt.expected {
			t.Errorf("Expected %v.String() to be '%s', got '%s'", tt.mode, tt.expected, tt.mode.String())
		}
	}
}

func TestRiskLevelString(t *testing.T) {
	tests := []struct {
		level    RiskLevel
		expected string
	}{
		{RiskLow, "low"},
		{RiskMedium, "medium"},
		{RiskHigh, "high"},
		{RiskCritical, "critical"},
	}

	for _, tt := range tests {
		if tt.level.String() != tt.expected {
			t.Errorf("Expected %v.String() to be '%s', got '%s'", tt.level, tt.expected, tt.level.String())
		}
	}
}

func TestPermissionRequest(t *testing.T) {
	pr := PermissionRequest{
		Tool:   "Bash",
		Input:  []byte(`{"command": "ls"}`),
		Reason: "List directory contents",
		Risk:   RiskHigh,
	}

	if pr.Tool != "Bash" {
		t.Errorf("Expected Tool 'Bash', got '%s'", pr.Tool)
	}
	if pr.Risk != RiskHigh {
		t.Errorf("Expected Risk High, got %v", pr.Risk)
	}
}
