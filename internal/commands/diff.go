package commands

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/claude-code-go/claude/internal/engine"
)

// DiffCommand shows git diff
type DiffCommand struct {
	BaseCommand
}

func NewDiffCommand() *DiffCommand {
	return &DiffCommand{
		BaseCommand: *newPromptCommand("diff", "Show changes in the working directory"),
	}
}

func (c *DiffCommand) Execute(ctx context.Context, args []string, execCtx engine.CommandContext) error {
	diffArgs := []string{"diff"}
	if len(args) > 0 && args[0] == "--staged" {
		diffArgs = []string{"diff", "--cached"}
	}

	cmd := exec.CommandContext(ctx, "git", diffArgs...)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("git diff failed: %w", err)
	}

	fmt.Println(string(output))
	return nil
}
