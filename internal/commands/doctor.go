package commands

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/charmbracelet/lipgloss"
	"github.com/claude-code-go/claude/internal/engine"
)

var (
	infoStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("green"))
	warnStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("yellow"))
	errorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("red"))
)

// DoctorCommand checks environment health
type DoctorCommand struct {
	BaseCommand
}

func NewDoctorCommand() *DoctorCommand {
	return &DoctorCommand{
		BaseCommand: *newCommand("doctor", "Run diagnostics to check environment health"),
	}
}

func (c *DoctorCommand) Execute(ctx context.Context, args []string, execCtx engine.CommandContext) error {
	fmt.Println(infoStyle.Render("Running diagnostics..."))
	fmt.Println()

	checks := []struct {
		name  string
		check func() error
	}{
		{"Go installation", c.checkGo},
		{"Git installation", c.checkGit},
		{"API key", c.checkAPIKey},
		{"Network connectivity", c.checkNetwork},
	}

	allPassed := true
	for _, check := range checks {
		if err := check.check(); err != nil {
			fmt.Printf("%s %s: %s\n", errorStyle.Render("[FAIL]"), check.name, err)
			allPassed = false
		} else {
			fmt.Printf("%s %s: OK\n", infoStyle.Render("[PASS]"), check.name)
		}
	}

	fmt.Println()
	if allPassed {
		fmt.Println(infoStyle.Render("All checks passed!"))
	} else {
		fmt.Println(warnStyle.Render("Some checks failed. Please review the issues above."))
	}

	return nil
}

func (c *DoctorCommand) checkGo() error {
	cmd := exec.Command("go", "version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go not found")
	}
	return nil
}

func (c *DoctorCommand) checkGit() error {
	cmd := exec.Command("git", "--version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git not found")
	}
	return nil
}

func (c *DoctorCommand) checkAPIKey() error {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("ANTHROPIC_API_KEY not set")
	}
	if len(apiKey) < 10 {
		return fmt.Errorf("ANTHROPIC_API_KEY appears to be invalid")
	}
	return nil
}

func (c *DoctorCommand) checkNetwork() error {
	cmd := exec.Command("curl", "-s", "-o", "/dev/null", "-w", "%{http_code}", "https://api.anthropic.com")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("cannot reach api.anthropic.com")
	}
	if string(output) != "200" {
		return fmt.Errorf("api.anthropic.com returned status: %s", string(output))
	}
	return nil
}

// VersionCommand shows version info
type VersionCommand struct {
	BaseCommand
}

func NewVersionCommand() *VersionCommand {
	return &VersionCommand{
		BaseCommand: *newCommand("version", "Show version information"),
	}
}

func (c *VersionCommand) Execute(ctx context.Context, args []string, execCtx engine.CommandContext) error {
	fmt.Println("Claude Code (Go port)")
	fmt.Printf("Version: 0.1.0\n")
	fmt.Printf("Go version: %s\n", runtime.Version())
	return nil
}

// HelpCommand shows help information
type HelpCommand struct {
	BaseCommand
}

func NewHelpCommand() *HelpCommand {
	return &HelpCommand{
		BaseCommand: *newCommand("help", "Show help information"),
	}
}

func (c *HelpCommand) Execute(ctx context.Context, args []string, execCtx engine.CommandContext) error {
	fmt.Println("Claude Code - AI Programming Assistant")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  claude [command] [args]")
	fmt.Println()
	fmt.Println("Available Commands:")
	fmt.Println("  /commit    Create a git commit")
	fmt.Println("  /diff      Show changes in working directory")
	fmt.Println("  /compact   Compress conversation context")
	fmt.Println("  /doctor    Run diagnostics")
	fmt.Println("  /help      Show this help message")
	fmt.Println("  /version   Show version information")
	fmt.Println("  /model     Get or set the model")
	fmt.Println("  /config    Manage configuration")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  claude \"Write a hello world program\"")
	fmt.Println("  claude /commit")
	fmt.Println("  claude /model claude-sonnet-4-20250514")
	return nil
}
