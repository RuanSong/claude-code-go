package commands

import (
	"context"
	"fmt"

	"github.com/claude-code-go/claude/internal/engine"
)

// StatuslineCommand 状态栏设置命令 - 配置Claude Code的状态栏UI
type StatuslineCommand struct {
	BaseCommand
}

func NewStatuslineCommand() *StatuslineCommand {
	return &StatuslineCommand{
		BaseCommand: *newCommand("statusline", "Set up Claude Code's status line UI"),
	}
}

func (c *StatuslineCommand) Execute(ctx context.Context, args []string, execCtx engine.CommandContext) error {
	// 状态栏设置提示
	// 用户可以提供自定义提示来配置状态栏

	message := "Configure my statusLine from my shell PS1 configuration"
	if len(args) > 0 {
		message = args[0]
	}

	fmt.Println("Setting up statusLine...")
	fmt.Printf("Prompt: %s\n", message)
	fmt.Println()
	fmt.Println("This command creates an Agent with subagent_type 'statusline-setup'")
	fmt.Println("to help configure your shell's status line display.")
	fmt.Println()
	fmt.Println("Status line features:")
	fmt.Println("  - Display Claude Code status in your terminal prompt")
	fmt.Println("  - Show current model and session info")
	fmt.Println("  - Integration with popular shells (bash, zsh, fish)")

	return nil
}
