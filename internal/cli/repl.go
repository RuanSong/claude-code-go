package cli

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/claude-code-go/claude/internal/commands"
	"github.com/claude-code-go/claude/internal/engine"
	"github.com/claude-code-go/claude/internal/tools"
	"github.com/claude-code-go/claude/pkg/anthropic"
	"github.com/claude-code-go/claude/pkg/llm"
)

var (
	headerStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("green")).Bold(true)
	promptStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("cyan"))
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("red"))
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("green"))
)

// REPL represents the interactive REPL
type REPL struct {
	engine   *engine.QueryEngine
	commands *commands.Registry
	apiKey   string
	model    string
}

// NewREPL creates a new REPL instance
func NewREPL(apiKey, model string) *REPL {
	return &REPL{
		apiKey:   apiKey,
		model:    model,
		commands: commands.DefaultRegistry(),
	}
}

// Run starts the REPL
func (r *REPL) Run() error {
	r.printWelcome()

	// Initialize LLM provider using the adapter for backward compatibility
	apiClient := llm.NewAnthropicProvider(
		r.apiKey,
		anthropic.DefaultBaseURL,
		r.model,
	)

	toolRegistry := engine.NewToolRegistry()
	for _, tool := range GetBuiltInTools() {
		if err := toolRegistry.Register(tool); err != nil {
			return fmt.Errorf("register tool: %w", err)
		}
	}

	qe := engine.NewQueryEngine(engine.Config{
		Model:     r.model,
		Tools:     toolRegistry,
		MaxTurns:  100,
		MaxTokens: 200000,
	}, apiClient)
	r.engine = qe

	// Main loop
	for {
		input, err := r.readInput()
		if err != nil {
			if err.Error() == "EOF" {
				fmt.Println("\nGoodbye!")
				return nil
			}
			fmt.Fprintf(os.Stderr, "%s %v\n", errorStyle.Render("Error:"), err)
			continue
		}

		if input == "" {
			continue
		}

		if err := r.processInput(input); err != nil {
			fmt.Fprintf(os.Stderr, "%s %v\n", errorStyle.Render("Error:"), err)
		}
	}
}

func (r *REPL) printWelcome() {
	fmt.Println(headerStyle.Render("Claude Code REPL"))
	fmt.Println("Type '/help' for available commands or enter a message to chat.")
	fmt.Println()
}

func (r *REPL) readInput() (string, error) {
	fmt.Print(promptStyle.Render("> "))
	var input string
	_, err := fmt.Scanln(&input)
	return strings.TrimSpace(input), err
}

func (r *REPL) processInput(input string) error {
	// Check for slash command
	if strings.HasPrefix(input, "/") {
		return r.handleCommand(input)
	}

	// Process as chat message
	ctx := context.Background()
	return r.engine.SubmitMessage(ctx, input)
}

func (r *REPL) handleCommand(input string) error {
	parts := strings.SplitN(input, " ", 2)
	cmdName := strings.TrimPrefix(parts[0], "/")
	var args []string
	if len(parts) > 1 {
		args = strings.Fields(parts[1])
	}

	cmd, exists := r.commands.Get(cmdName)
	if !exists {
		return fmt.Errorf("unknown command: /%s", cmdName)
	}

	execCtx := engine.CommandContext{
		Cwd: r.commands.List()[0].Name(),
	}

	ctx := context.Background()
	if err := cmd.Execute(ctx, args, execCtx); err != nil {
		return err
	}

	fmt.Println(successStyle.Render("Done"))
	return nil
}

// GetBuiltInTools returns the built-in tools for REPL
func GetBuiltInTools() []engine.Tool {
	return tools.GetExtendedTools()
}
