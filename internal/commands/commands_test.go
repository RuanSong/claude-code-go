package commands

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/claude-code-go/claude/internal/engine"
)

func TestInitCommand(t *testing.T) {
	// Create temp directory for testing
	tmpDir := t.TempDir()

	execCtx := engine.CommandContext{
		Cwd: tmpDir,
	}

	cmd := NewInitCommand()
	if cmd.Name() != "init" {
		t.Errorf("Expected name 'init', got '%s'", cmd.Name())
	}
	if cmd.Description() == "" {
		t.Error("Description should not be empty")
	}

	// Test execute
	err := cmd.Execute(context.Background(), []string{}, execCtx)
	if err != nil {
		t.Errorf("Execute failed: %v", err)
	}

	// Check if CLAUDE.md was created
	claudeMdPath := filepath.Join(tmpDir, "CLAUDE.md")
	if _, err := os.Stat(claudeMdPath); os.IsNotExist(err) {
		t.Error("CLAUDE.md should have been created")
	}
}

func TestInitCommandExistingFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create existing CLAUDE.md
	existingPath := filepath.Join(tmpDir, "CLAUDE.md")
	os.WriteFile(existingPath, []byte("# Existing CLAUDE.md"), 0644)

	execCtx := engine.CommandContext{
		Cwd: tmpDir,
	}

	cmd := NewInitCommand()
	err := cmd.Execute(context.Background(), []string{}, execCtx)
	if err != nil {
		t.Errorf("Execute failed: %v", err)
	}
}

func TestDetectProjectInfo(t *testing.T) {
	tmpDir := t.TempDir()

	// Test Go project detection
	goModPath := filepath.Join(tmpDir, "go.mod")
	os.WriteFile(goModPath, []byte("module test"), 0644)

	info := detectProjectInfo(tmpDir)
	if info.Language != "Go" {
		t.Errorf("Expected Go language, got %s", info.Language)
	}

	// Clean up and test JS detection
	os.Remove(goModPath)
	packageJsonPath := filepath.Join(tmpDir, "package.json")
	os.WriteFile(packageJsonPath, []byte(`{"name": "test"}`), 0644)

	info = detectProjectInfo(tmpDir)
	if info.Language != "JavaScript/TypeScript" {
		t.Errorf("Expected JavaScript/TypeScript, got %s", info.Language)
	}
}

func TestHasFiles(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte("package main"), 0644)

	if !hasFiles(tmpDir, ".go") {
		t.Error("Should detect .go files")
	}
	if hasFiles(tmpDir, ".py") {
		t.Error("Should not detect .py files")
	}
}

func TestLoginCommand(t *testing.T) {
	cmd := NewLoginCommand()
	if cmd.Name() != "login" {
		t.Errorf("Expected name 'login', got '%s'", cmd.Name())
	}

	execCtx := engine.CommandContext{
		Cwd: t.TempDir(),
	}

	// Test login with no API key
	err := cmd.Execute(context.Background(), []string{}, execCtx)
	if err != nil {
		t.Errorf("Execute failed: %v", err)
	}
}

func TestLogoutCommand(t *testing.T) {
	cmd := NewLogoutCommand()
	if cmd.Name() != "logout" {
		t.Errorf("Expected name 'logout', got '%s'", cmd.Name())
	}

	execCtx := engine.CommandContext{
		Cwd: t.TempDir(),
	}

	// Test logout
	err := cmd.Execute(context.Background(), []string{}, execCtx)
	if err != nil {
		t.Errorf("Execute failed: %v", err)
	}
}

func TestHasAuth(t *testing.T) {
	// Clear the env first
	os.Unsetenv("ANTHROPIC_API_KEY")

	if hasAuth() {
		t.Error("Should return false when API key is not set")
	}

	// Set a valid-looking API key
	os.Setenv("ANTHROPIC_API_KEY", "sk-ant-example1234567890")
	defer os.Unsetenv("ANTHROPIC_API_KEY")

	if !hasAuth() {
		t.Error("Should return true when API key is set")
	}
}

func TestGetConfigPath(t *testing.T) {
	path := getConfigPath()
	if path == "" {
		t.Error("Config path should not be empty on Unix systems")
	}
	if path != "" && filepath.Base(path) != ".claude" {
		t.Errorf("Expected path ending in .claude, got %s", path)
	}
}

func TestModelCommand(t *testing.T) {
	cmd := NewModelCommand()
	if cmd.Name() != "model" {
		t.Errorf("Expected name 'model', got '%s'", cmd.Name())
	}

	execCtx := engine.CommandContext{
		Cwd: t.TempDir(),
	}

	// Test list models
	err := cmd.Execute(context.Background(), []string{"list"}, execCtx)
	if err != nil {
		t.Errorf("Execute list failed: %v", err)
	}
}

func TestConfigCommand(t *testing.T) {
	cmd := NewConfigCommand()
	if cmd.Name() != "config" {
		t.Errorf("Expected name 'config', got '%s'", cmd.Name())
	}

	execCtx := engine.CommandContext{
		Cwd: t.TempDir(),
	}

	// Test list config
	err := cmd.Execute(context.Background(), []string{"list"}, execCtx)
	if err != nil {
		t.Errorf("Execute list failed: %v", err)
	}
}

func TestDoctorCommand(t *testing.T) {
	cmd := NewDoctorCommand()
	if cmd.Name() != "doctor" {
		t.Errorf("Expected name 'doctor', got '%s'", cmd.Name())
	}

	execCtx := engine.CommandContext{
		Cwd: t.TempDir(),
	}

	err := cmd.Execute(context.Background(), []string{}, execCtx)
	if err != nil {
		t.Errorf("Execute failed: %v", err)
	}
}

func TestVersionCommand(t *testing.T) {
	cmd := NewVersionCommand()
	if cmd.Name() != "version" {
		t.Errorf("Expected name 'version', got '%s'", cmd.Name())
	}

	execCtx := engine.CommandContext{
		Cwd: t.TempDir(),
	}

	err := cmd.Execute(context.Background(), []string{}, execCtx)
	if err != nil {
		t.Errorf("Execute failed: %v", err)
	}
}

func TestHelpCommand(t *testing.T) {
	cmd := NewHelpCommand()
	if cmd.Name() != "help" {
		t.Errorf("Expected name 'help', got '%s'", cmd.Name())
	}

	execCtx := engine.CommandContext{
		Cwd: t.TempDir(),
	}

	err := cmd.Execute(context.Background(), []string{}, execCtx)
	if err != nil {
		t.Errorf("Execute failed: %v", err)
	}
}

func TestMCPCommand(t *testing.T) {
	cmd := NewMCPCommand()
	if cmd.Name() != "mcp" {
		t.Errorf("Expected name 'mcp', got '%s'", cmd.Name())
	}

	execCtx := engine.CommandContext{
		Cwd: t.TempDir(),
	}

	// Test list servers
	err := cmd.Execute(context.Background(), []string{"list"}, execCtx)
	if err != nil {
		t.Errorf("Execute list failed: %v", err)
	}
}

func TestSkillsCommand(t *testing.T) {
	cmd := NewSkillsCommand()
	if cmd.Name() != "skills" {
		t.Errorf("Expected name 'skills', got '%s'", cmd.Name())
	}

	execCtx := engine.CommandContext{
		Cwd: t.TempDir(),
	}

	// Test list skills
	err := cmd.Execute(context.Background(), []string{"list"}, execCtx)
	if err != nil {
		t.Errorf("Execute list failed: %v", err)
	}
}

func TestReviewCommand(t *testing.T) {
	cmd := NewReviewCommand()
	if cmd.Name() != "review" {
		t.Errorf("Expected name 'review', got '%s'", cmd.Name())
	}

	execCtx := engine.CommandContext{
		Cwd: t.TempDir(),
	}

	// Test review - should fail since not in git repo
	_ = cmd.Execute(context.Background(), []string{}, execCtx)
	// We don't check for error since we're not in a git repo
}

func TestUltraReviewCommand(t *testing.T) {
	cmd := NewUltraReviewCommand()
	if cmd.Name() != "ultrareview" {
		t.Errorf("Expected name 'ultrareview', got '%s'", cmd.Name())
	}

	execCtx := engine.CommandContext{
		Cwd: t.TempDir(),
	}

	// Test ultrareview
	_ = cmd.Execute(context.Background(), []string{}, execCtx)
	// We don't check for error since we're not in a git repo
}

func TestCompactCommand(t *testing.T) {
	cmd := NewCompactCommand()
	if cmd.Name() != "compact" {
		t.Errorf("Expected name 'compact', got '%s'", cmd.Name())
	}

	execCtx := engine.CommandContext{
		Cwd: t.TempDir(),
	}

	err := cmd.Execute(context.Background(), []string{}, execCtx)
	if err != nil {
		t.Errorf("Execute failed: %v", err)
	}
}

func TestDiffCommand(t *testing.T) {
	cmd := NewDiffCommand()
	if cmd.Name() != "diff" {
		t.Errorf("Expected name 'diff', got '%s'", cmd.Name())
	}

	execCtx := engine.CommandContext{
		Cwd: t.TempDir(),
	}

	// Test diff - may fail since not in git repo
	_ = cmd.Execute(context.Background(), []string{}, execCtx)
	// Don't check error since not in git repo
}

func TestCommitCommand(t *testing.T) {
	cmd := NewCommitCommand()
	if cmd.Name() != "commit" {
		t.Errorf("Expected name 'commit', got '%s'", cmd.Name())
	}

	execCtx := engine.CommandContext{
		Cwd: t.TempDir(),
	}

	// Test commit - may fail since not in git repo
	_ = cmd.Execute(context.Background(), []string{}, execCtx)
	// Don't check error since not in git repo
}

func TestDefaultRegistry(t *testing.T) {
	registry := DefaultRegistry()

	// Check that all expected commands are registered
	expectedCommands := []string{
		"commit", "diff", "compact", "doctor", "review",
		"ultrareview", "model", "config", "help", "version",
		"init", "login", "logout", "mcp", "skills",
	}

	for _, name := range expectedCommands {
		cmd, ok := registry.Get(name)
		if !ok {
			t.Errorf("Command %s not found in registry", name)
		}
		if cmd.Name() != name {
			t.Errorf("Command name mismatch: expected %s, got %s", name, cmd.Name())
		}
	}
}

func TestCommandRegistry(t *testing.T) {
	registry := NewRegistry()

	// Test registering commands
	cmd := NewCommitCommand()
	err := registry.Register(cmd)
	if err != nil {
		t.Errorf("Register failed: %v", err)
	}

	// Test duplicate registration
	err = registry.Register(cmd)
	if err == nil {
		t.Error("Expected error on duplicate registration")
	}

	// Test getting command
	gotCmd, ok := registry.Get("commit")
	if !ok {
		t.Error("Get failed")
	}
	if gotCmd.Name() != "commit" {
		t.Errorf("Got wrong command: %s", gotCmd.Name())
	}

	// Test listing commands
	cmds := registry.List()
	if len(cmds) != 1 {
		t.Errorf("Expected 1 command, got %d", len(cmds))
	}

	// Test names
	names := registry.Names()
	if len(names) != 1 {
		t.Errorf("Expected 1 name, got %d", len(names))
	}
}

func TestCommandTypes(t *testing.T) {
	// Test CommandTypePrompt
	cmd := NewCommitCommand()
	if cmd.Type() != engine.CommandTypePrompt {
		t.Errorf("Expected CommandTypePrompt, got %v", cmd.Type())
	}

	// Test CommandTypeCustom
	customCmd := NewDoctorCommand()
	if customCmd.Type() != engine.CommandTypeCustom {
		t.Errorf("Expected CommandTypeCustom, got %v", customCmd.Type())
	}
}
