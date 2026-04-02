package commands

import (
	"context"
	"testing"

	"github.com/claude-code-go/claude/internal/engine"
)

func TestIssueCommand(t *testing.T) {
	cmd := NewIssueCommand()

	if cmd.Name() != "issue" {
		t.Errorf("Expected name 'issue', got %q", cmd.Name())
	}

	if cmd.Description() == "" {
		t.Error("Description should not be empty")
	}

	execCtx := engine.CommandContext{
		Cwd: t.TempDir(),
	}
	err := cmd.Execute(context.Background(), []string{}, execCtx)
	if err != nil {
		t.Errorf("Execute failed: %v", err)
	}
}

func TestPRCommentsCommand(t *testing.T) {
	cmd := NewPRCommentsCommand()

	if cmd.Name() != "pr-comments" {
		t.Errorf("Expected name 'pr-comments', got %q", cmd.Name())
	}

	execCtx := engine.CommandContext{
		Cwd: t.TempDir(),
	}
	err := cmd.Execute(context.Background(), []string{}, execCtx)
	if err != nil {
		t.Errorf("Execute failed: %v", err)
	}
}

func TestBugHunterCommand(t *testing.T) {
	cmd := NewBugHunterCommand()

	if cmd.Name() != "bughunter" {
		t.Errorf("Expected name 'bughunter', got %q", cmd.Name())
	}

	execCtx := engine.CommandContext{
		Cwd: t.TempDir(),
	}
	err := cmd.Execute(context.Background(), []string{}, execCtx)
	if err != nil {
		t.Errorf("Execute failed: %v", err)
	}
}

func TestAntTraceCommand(t *testing.T) {
	cmd := NewAntTraceCommand()

	if cmd.Name() != "ant-trace" {
		t.Errorf("Expected name 'ant-trace', got %q", cmd.Name())
	}

	execCtx := engine.CommandContext{
		Cwd: t.TempDir(),
	}
	err := cmd.Execute(context.Background(), []string{}, execCtx)
	if err != nil {
		t.Errorf("Execute failed: %v", err)
	}
}

func TestBreakCacheCommand(t *testing.T) {
	cmd := NewBreakCacheCommand()

	if cmd.Name() != "break-cache" {
		t.Errorf("Expected name 'break-cache', got %q", cmd.Name())
	}

	execCtx := engine.CommandContext{
		Cwd: t.TempDir(),
	}
	err := cmd.Execute(context.Background(), []string{}, execCtx)
	if err != nil {
		t.Errorf("Execute failed: %v", err)
	}
}

func TestGoodClaudeCommand(t *testing.T) {
	cmd := NewGoodClaudeCommand()

	if cmd.Name() != "good-claude" {
		t.Errorf("Expected name 'good-claude', got %q", cmd.Name())
	}

	execCtx := engine.CommandContext{
		Cwd: t.TempDir(),
	}
	err := cmd.Execute(context.Background(), []string{}, execCtx)
	if err != nil {
		t.Errorf("Execute failed: %v", err)
	}
}

func TestResetLimitsCommand(t *testing.T) {
	cmd := NewResetLimitsCommand()

	if cmd.Name() != "reset-limits" {
		t.Errorf("Expected name 'reset-limits', got %q", cmd.Name())
	}

	execCtx := engine.CommandContext{
		Cwd: t.TempDir(),
	}
	err := cmd.Execute(context.Background(), []string{}, execCtx)
	if err != nil {
		t.Errorf("Execute failed: %v", err)
	}
}

func TestDebugToolCallCommand(t *testing.T) {
	cmd := NewDebugToolCallCommand()

	if cmd.Name() != "debug-tool-call" {
		t.Errorf("Expected name 'debug-tool-call', got %q", cmd.Name())
	}

	execCtx := engine.CommandContext{
		Cwd: t.TempDir(),
	}
	err := cmd.Execute(context.Background(), []string{}, execCtx)
	if err != nil {
		t.Errorf("Execute failed: %v", err)
	}
}

func TestPerfIssueCommand(t *testing.T) {
	cmd := NewPerfIssueCommand()

	if cmd.Name() != "perf-issue" {
		t.Errorf("Expected name 'perf-issue', got %q", cmd.Name())
	}

	execCtx := engine.CommandContext{
		Cwd: t.TempDir(),
	}
	err := cmd.Execute(context.Background(), []string{}, execCtx)
	if err != nil {
		t.Errorf("Execute failed: %v", err)
	}
}

func TestOAuthRefreshCommand(t *testing.T) {
	cmd := NewOAuthRefreshCommand()

	if cmd.Name() != "oauth-refresh" {
		t.Errorf("Expected name 'oauth-refresh', got %q", cmd.Name())
	}

	execCtx := engine.CommandContext{
		Cwd: t.TempDir(),
	}
	err := cmd.Execute(context.Background(), []string{}, execCtx)
	if err != nil {
		t.Errorf("Execute failed: %v", err)
	}
}
