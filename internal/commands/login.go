package commands

import (
	"context"
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/claude-code-go/claude/internal/engine"
)

var (
	loginStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("cyan"))
	successLStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("green"))
	errorStyleL   = lipgloss.NewStyle().Foreground(lipgloss.Color("red"))
)

// LoginCommand handles authentication
type LoginCommand struct {
	BaseCommand
}

func NewLoginCommand() *LoginCommand {
	return &LoginCommand{
		BaseCommand: *newCommand("login", "Sign in with your Anthropic account"),
	}
}

func (c *LoginCommand) Execute(ctx context.Context, args []string, execCtx engine.CommandContext) error {
	fmt.Println(loginStyle.Render("Claude Code Login"))
	fmt.Println()

	// Check if ANTHROPIC_API_KEY is already set
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey != "" {
		fmt.Println("You are already logged in with an API key.")
		fmt.Println("To switch accounts, first run /logout, then /login again.")
		return nil
	}

	// Check for existing auth
	if hasAuth() {
		fmt.Println("You are already authenticated.")
		return nil
	}

	fmt.Println("To login, you need an Anthropic API key.")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("1. Set ANTHROPIC_API_KEY environment variable")
	fmt.Println("2. Use /config set api-key <your-key>")
	fmt.Println()
	fmt.Println("Get your API key at: https://console.anthropic.com/settings/keys")
	fmt.Println()

	// In interactive mode, could prompt for API key
	fmt.Println("Run `/config set api-key YOUR_KEY` to configure your API key.")

	return nil
}

func hasAuth() bool {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	return apiKey != "" && len(apiKey) > 10
}

// LogoutCommand handles sign out
type LogoutCommand struct {
	BaseCommand
}

func NewLogoutCommand() *LogoutCommand {
	return &LogoutCommand{
		BaseCommand: *newCommand("logout", "Sign out and clear stored credentials"),
	}
}

func (c *LogoutCommand) Execute(ctx context.Context, args []string, execCtx engine.CommandContext) error {
	fmt.Println(loginStyle.Render("Claude Code Logout"))
	fmt.Println()

	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		fmt.Println("You are not currently logged in.")
		return nil
	}

	// Clear the API key
	if err := os.Unsetenv("ANTHROPIC_API_KEY"); err != nil {
		return fmt.Errorf("failed to clear API key: %w", err)
	}

	// Also clear any cached config
	configPath := getConfigPath()
	if configPath != "" {
		os.RemoveAll(configPath)
	}

	fmt.Println(successLStyle.Render("✓ Logged out successfully"))
	fmt.Println()
	fmt.Println("Your API key has been cleared.")
	fmt.Println("Run /login to sign in again.")

	return nil
}

func getConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return home + "/.claude"
}
