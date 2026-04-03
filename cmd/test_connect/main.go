package main

import (
	"context"
	"fmt"
	"os"

	"github.com/claude-code-go/claude/internal/commands"
	"github.com/claude-code-go/claude/internal/engine"
)

func main() {
	registry := commands.DefaultRegistry()
	cmd, exists := registry.Get("connect")
	if !exists {
		fmt.Println("ERROR: /connect command not found!")
		os.Exit(1)
	}
	fmt.Printf("Found command: %s\n", cmd.Name())
	fmt.Printf("Description: %s\n", cmd.Description())
	fmt.Println()

	// Try executing with "list" argument
	ctx := context.Background()
	execCtx := engine.CommandContext{}

	fmt.Println("Running /connect list:")
	err := cmd.Execute(ctx, []string{"list"}, execCtx)
	if err != nil {
		fmt.Printf("Error executing: %v\n", err)
	}

	fmt.Println()
	fmt.Println("Running /connect env:")
	err = cmd.Execute(ctx, []string{"env"}, execCtx)
	if err != nil {
		fmt.Printf("Error executing: %v\n", err)
	}
}
