package engine

import (
	"context"
	"fmt"
)

// PermissionModeType defines different permission modes
type PermissionModeType int

const (
	PermissionModeDefault PermissionModeType = iota
	PermissionModePlan                       // plan mode - requires confirmation
	PermissionModeBypass                     // skip all permission checks
	PermissionModeAuto                       // auto decide based on risk
	PermissionModeYolo                       // allow everything
)

func (p PermissionModeType) String() string {
	switch p {
	case PermissionModeDefault:
		return "default"
	case PermissionModePlan:
		return "plan"
	case PermissionModeBypass:
		return "bypass"
	case PermissionModeAuto:
		return "auto"
	case PermissionModeYolo:
		return "yolo"
	default:
		return "unknown"
	}
}

// RiskLevel defines the risk level of an operation
type RiskLevel int

const (
	RiskLow RiskLevel = iota
	RiskMedium
	RiskHigh
	RiskCritical
)

func (r RiskLevel) String() string {
	switch r {
	case RiskLow:
		return "low"
	case RiskMedium:
		return "medium"
	case RiskHigh:
		return "high"
	case RiskCritical:
		return "critical"
	default:
		return "unknown"
	}
}

// PermissionRequest represents a request to use a tool
type PermissionRequest struct {
	Tool   string
	Input  []byte
	Reason string
	Risk   RiskLevel
}

// PermissionRules defines permission rules
type PermissionRules struct {
	AlwaysAllow []string // tools that are always allowed
	AlwaysDeny  []string // tools that are always denied
	AlwaysAsk   []string // tools that always ask
}

// PermissionManager handles permission checking
type PermissionManager struct {
	mode  PermissionModeType
	rules PermissionRules
}

var (
	ErrPermissionRequired = fmt.Errorf("permission required")
	ErrPermissionDenied   = fmt.Errorf("permission denied")
)

// NewPermissionManager creates a new permission manager
func NewPermissionManager() *PermissionManager {
	return &PermissionManager{
		mode: PermissionModeDefault,
		rules: PermissionRules{
			AlwaysAllow: []string{
				"Read",
				"Glob",
				"Grep",
				"LSP",
			},
			AlwaysDeny: []string{
				// dangerous commands that should never be auto-allowed
			},
			AlwaysAsk: []string{
				"Bash",
				"Write",
				"Edit",
				"NotebookEdit",
			},
		},
	}
}

// Check checks if a tool can be executed
func (m *PermissionManager) Check(ctx context.Context, toolName string) (bool, error) {
	// Check always allow
	for _, allowed := range m.rules.AlwaysAllow {
		if allowed == toolName {
			return true, nil
		}
	}

	// Check always deny
	for _, denied := range m.rules.AlwaysDeny {
		if denied == toolName {
			return false, ErrPermissionDenied
		}
	}

	// Check mode
	switch m.mode {
	case PermissionModeBypass, PermissionModeYolo:
		return true, nil

	case PermissionModePlan:
		return false, ErrPermissionRequired

	case PermissionModeAuto:
		return m.autoDecide(toolName)

	case PermissionModeDefault:
		// Check always ask list
		for _, ask := range m.rules.AlwaysAsk {
			if ask == toolName {
				return false, ErrPermissionRequired
			}
		}
		return true, nil

	default:
		return false, ErrPermissionRequired
	}
}

// autoDecide makes an automatic decision based on tool name patterns
func (m *PermissionManager) autoDecide(toolName string) (bool, error) {
	// High risk tools always require permission
	highRiskPatterns := []string{
		"Bash",
		"Write",
		"Edit",
		"NotebookEdit",
		"Task",
		"Agent",
		"WebSearch",
		"WebFetch",
	}

	for _, pattern := range highRiskPatterns {
		if toolName == pattern {
			return false, ErrPermissionRequired
		}
	}

	return true, nil
}

// SetMode sets the permission mode
func (m *PermissionManager) SetMode(mode PermissionModeType) {
	m.mode = mode
}

// GetMode returns the current permission mode
func (m *PermissionManager) GetMode() PermissionModeType {
	return m.mode
}

// SetRules sets custom permission rules
func (m *PermissionManager) SetRules(rules PermissionRules) {
	m.rules = rules
}

// RiskAssessor assesses the risk of a tool call
type RiskAssessor struct{}

func (a *RiskAssessor) Assess(toolName string, input []byte) RiskLevel {
	switch toolName {
	case "Bash":
		return RiskHigh
	case "Write", "Edit", "NotebookEdit":
		return RiskMedium
	case "WebSearch", "WebFetch":
		return RiskMedium
	case "Read", "Glob", "Grep":
		return RiskLow
	default:
		return RiskMedium
	}
}
