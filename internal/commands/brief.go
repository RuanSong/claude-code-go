package commands

import (
	"context"
	"fmt"

	"github.com/claude-code-go/claude/internal/engine"
)

// BriefCommand 简洁模式切换命令 - 切换仅Brief工具输出模式
type BriefCommand struct {
	BaseCommand
}

func NewBriefCommand() *BriefCommand {
	return &BriefCommand{
		BaseCommand: *newCommand("brief", "Toggle brief-only mode"),
	}
}

// Execute 切换简洁模式
// 简洁模式只显示Brief工具的输出，隐藏普通文本回复
func (c *BriefCommand) Execute(ctx context.Context, args []string, execCtx engine.CommandContext) error {
	fmt.Println("Brief mode toggle")
	fmt.Println()
	fmt.Println("Brief-only mode controls whether only the Brief tool's output is shown.")
	fmt.Println("When enabled, plain text outside of Brief tool calls is hidden from view.")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  /brief - Toggle brief mode on/off")
	fmt.Println()
	fmt.Println("Note: This feature requires Kairos access and may not be available for all accounts.")

	return nil
}
