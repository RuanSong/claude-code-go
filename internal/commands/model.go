package commands

import (
	"context"
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/claude-code-go/claude/internal/engine"
)

var modelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("cyan"))

var availableModels = []string{
	"claude-opus-4-20250514",
	"claude-sonnet-4-20250514",
	"claude-3-5-sonnet-20241022",
	"claude-3-opus-20240229",
	"claude-3-sonnet-20240229",
	"claude-3-haiku-20240307",
}

var currentModelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("green"))
var availableModelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("blue"))
var selectedModelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("yellow")).Bold(true)

// ModelCommand gets or sets the model
type ModelCommand struct {
	BaseCommand
}

func NewModelCommand() *ModelCommand {
	return &ModelCommand{
		BaseCommand: *newCommand("model", "Get or set the model to use"),
	}
}

func (c *ModelCommand) Execute(ctx context.Context, args []string, execCtx engine.CommandContext) error {
	if len(args) == 0 {
		return c.showCurrentModel(execCtx)
	}

	if len(args) == 1 && (args[0] == "list" || args[0] == "ls") {
		return c.listModels()
	}

	model := args[0]
	return c.setModel(model, execCtx)
}

func (c *ModelCommand) showCurrentModel(execCtx engine.CommandContext) error {
	model := execCtx.Config.Model
	if model == "" {
		model = os.Getenv("ANTHROPIC_MODEL")
		if model == "" {
			model = "claude-sonnet-4-20250514"
		}
	}

	fmt.Println(currentModelStyle.Render("Current model: ") + selectedModelStyle.Render(model))
	fmt.Println()
	fmt.Println("Use " + availableModelStyle.Render("/model list") + " to see available models.")
	fmt.Println("Use " + availableModelStyle.Render("/model <name>") + " to switch models.")
	return nil
}

func (c *ModelCommand) listModels() error {
	fmt.Println(modelStyle.Render("Available Models:"))
	fmt.Println(stringsRepeat("─", 40))
	fmt.Println()

	currentModel := os.Getenv("ANTHROPIC_MODEL")
	if currentModel == "" {
		currentModel = "claude-sonnet-4-20250514"
	}

	for _, model := range availableModels {
		if model == currentModel {
			fmt.Printf("  %s %s (current)\n", selectedModelStyle.Render("●"), model)
		} else {
			fmt.Printf("    %s\n", model)
		}
	}

	fmt.Println()
	fmt.Println("Use " + availableModelStyle.Render("/model <name>") + " to switch models.")
	return nil
}

func (c *ModelCommand) setModel(model string, execCtx engine.CommandContext) error {
	validModel := false
	for _, m := range availableModels {
		if m == model {
			validModel = true
			break
		}
	}

	if !validModel {
		fmt.Printf("Unknown model: %s\n", model)
		fmt.Println("Use " + availableModelStyle.Render("/model list") + " to see available models.")
		return nil
	}

	if err := os.Setenv("ANTHROPIC_MODEL", model); err != nil {
		return fmt.Errorf("failed to set model: %w", err)
	}

	fmt.Printf("Model set to: %s\n", selectedModelStyle.Render(model))
	return nil
}

func stringsRepeat(s string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}
