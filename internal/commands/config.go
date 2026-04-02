package commands

import (
	"context"
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/claude-code-go/claude/internal/engine"
)

var (
	configKeyStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("cyan"))
	configValueStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("green"))
	configErrorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("red"))
)

type ConfigOption struct {
	Key         string
	Value       string
	Description string
}

var configOptions = []ConfigOption{
	{"model", "claude-sonnet-4-20250514", "Model to use for conversations"},
	{"max-tokens", "8192", "Maximum tokens in response"},
	{"temperature", "1.0", "Temperature for generation (0.0-1.0)"},
	{"api-key", "", "Anthropic API key"},
	{"enable-tools", "true", "Enable tool usage"},
	{"stream-output", "true", "Stream API responses"},
}

var envVarMap = map[string]string{
	"model":         "ANTHROPIC_MODEL",
	"max-tokens":    "ANTHROPIC_MAX_TOKENS",
	"temperature":   "ANTHROPIC_TEMPERATURE",
	"api-key":       "ANTHROPIC_API_KEY",
	"enable-tools":  "CLAUDE_ENABLE_TOOLS",
	"stream-output": "CLAUDE_STREAM_OUTPUT",
}

// ConfigCommand manages configuration
type ConfigCommand struct {
	BaseCommand
}

func NewConfigCommand() *ConfigCommand {
	return &ConfigCommand{
		BaseCommand: *newCommand("config", "Manage configuration settings"),
	}
}

func (c *ConfigCommand) Execute(ctx context.Context, args []string, execCtx engine.CommandContext) error {
	if len(args) == 0 {
		return c.listConfig()
	}

	subcommand := args[0]

	switch subcommand {
	case "list", "ls":
		return c.listConfig()
	case "get":
		if len(args) < 2 {
			return fmt.Errorf("Usage: /config get <key>")
		}
		return c.getConfig(args[1])
	case "set":
		if len(args) < 3 {
			return fmt.Errorf("Usage: /config set <key> <value>")
		}
		return c.setConfig(args[1], args[2])
	case "unset":
		if len(args) < 2 {
			return fmt.Errorf("Usage: /config unset <key>")
		}
		return c.unsetConfig(args[1])
	default:
		return fmt.Errorf("Unknown subcommand: %s. Use 'list', 'get', 'set', or 'unset'", subcommand)
	}
}

func (c *ConfigCommand) listConfig() error {
	fmt.Println(configKeyStyle.Render("Configuration Settings:"))
	fmt.Println(stringsRepeat("─", 50))
	fmt.Println()

	for _, opt := range configOptions {
		envVar := envVarMap[opt.Key]
		value := os.Getenv(envVar)
		if value == "" {
			value = opt.Value
		}

		// Hide sensitive values
		if opt.Key == "api-key" && value != "" {
			value = "***" + value[len(value)-4:]
		}

		fmt.Printf("%s %s\n", configKeyStyle.Render(opt.Key), configValueStyle.Render(value))
		fmt.Printf("   %s\n", opt.Description)
		fmt.Printf("   Env: %s\n\n", envVar)
	}

	return nil
}

func (c *ConfigCommand) getConfig(key string) error {
	envVar, exists := envVarMap[key]
	if !exists {
		return fmt.Errorf("Unknown config key: %s", key)
	}

	value := os.Getenv(envVar)
	if value == "" {
		for _, opt := range configOptions {
			if opt.Key == key {
				value = opt.Value
				break
			}
		}
	}

	if value == "" {
		fmt.Printf("%s is not set\n", key)
	} else {
		if key == "api-key" && value != "" {
			value = "***" + value[len(value)-4:]
		}
		fmt.Printf("%s = %s\n", key, configValueStyle.Render(value))
	}

	return nil
}

func (c *ConfigCommand) setConfig(key, value string) error {
	envVar, exists := envVarMap[key]
	if !exists {
		return fmt.Errorf("Unknown config key: %s", key)
	}

	if err := os.Setenv(envVar, value); err != nil {
		return fmt.Errorf("failed to set config: %w", err)
	}

	fmt.Printf("Set %s = %s\n", key, configValueStyle.Render(value))
	return nil
}

func (c *ConfigCommand) unsetConfig(key string) error {
	envVar, exists := envVarMap[key]
	if !exists {
		return fmt.Errorf("Unknown config key: %s", key)
	}

	if err := os.Unsetenv(envVar); err != nil {
		return fmt.Errorf("failed to unset config: %w", err)
	}

	fmt.Printf("Unset %s\n", key)
	return nil
}
