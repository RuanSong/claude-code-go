package cli

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/claude-code-go/claude/internal/commands"
	"github.com/claude-code-go/claude/internal/engine"
	"github.com/claude-code-go/claude/internal/ui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var ErrNoAPIKey = errors.New("ANTHROPIC_API_KEY not set")

var (
	verbose    bool
	jsonOutput bool
	configPath string
)

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "claude",
		Short: "Claude Code - AI programming assistant",
		Long: `Claude Code is an AI programming assistant that helps you write, review,
and understand code. It can execute commands, edit files, and coordinate
complex software engineering tasks.`,
		Version: "0.1.0",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if err := viper.BindPFlags(cmd.PersistentFlags()); err != nil {
				return err
			}
			return nil
		},
	}

	cmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	cmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "output JSON")
	cmd.PersistentFlags().StringVar(&configPath, "config", "", "config file path")

	return cmd
}

// Execute runs the CLI
func Execute() error {
	rootCmd := newRootCmd()

	// Add subcommands
	rootCmd.AddCommand(newReplCmd())
	rootCmd.AddCommand(newLoginCmd())
	rootCmd.AddCommand(newLogoutCmd())
	rootCmd.AddCommand(newConfigCmd())
	rootCmd.AddCommand(newDoctorCmd())
	rootCmd.AddCommand(newModelCmd())
	rootCmd.AddCommand(newVersionCmd())
	rootCmd.AddCommand(newHelpCmd())
	rootCmd.AddCommand(newTuiCmd())

	return rootCmd.Execute()
}

func newReplCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "repl",
		Short: "Start interactive REPL mode",
		Long:  "Start an interactive conversation with Claude",
		RunE: func(cmd *cobra.Command, args []string) error {
			apiKey := viper.GetString("api_key")
			if apiKey == "" {
				apiKey = os.Getenv("ANTHROPIC_API_KEY")
			}
			if apiKey == "" {
				return ErrNoAPIKey
			}

			model := viper.GetString("model")
			if model == "" {
				model = "claude-sonnet-4-20250514"
			}

			repl := NewREPL(apiKey, model)
			return repl.Run()
		},
	}
}

func newLoginCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "login",
		Short: "Login to Claude",
		Long:  "Authenticate with your Anthropic API key",
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: implement interactive login
			fmt.Println("Login command - implementation pending")
			return nil
		},
	}
}

func newLogoutCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Logout from Claude",
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: implement logout
			fmt.Println("Logout command - implementation pending")
			return nil
		},
	}
}

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage configuration",
		Long:  "View and modify Claude Code configuration",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "get [key]",
		Short: "Get configuration value",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			val := viper.Get(args[0])
			fmt.Println(val)
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "set [key] [value]",
		Short: "Set configuration value",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			viper.Set(args[0], args[1])
			return viper.WriteConfig()
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List all configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			for _, key := range viper.AllKeys() {
				fmt.Printf("%s: %v\n", key, viper.Get(key))
			}
			return nil
		},
	})

	return cmd
}

func newDoctorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Run diagnostics",
		Long:  "Check environment and configuration for issues",
		RunE: func(cmd *cobra.Command, args []string) error {
			doctor := commands.NewDoctorCommand()
			return doctor.Execute(context.Background(), args, engine.CommandContext{})
		},
	}
}

func newModelCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "model [model-name]",
		Short: "Get or set the model",
		Args:  cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				model := viper.GetString("model")
				if model == "" {
					model = "claude-sonnet-4-20250514"
				}
				fmt.Println(model)
			} else {
				viper.Set("model", args[0])
				if err := viper.WriteConfig(); err != nil {
					return err
				}
				fmt.Printf("Model set to: %s\n", args[0])
			}
			return nil
		},
	}
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		RunE: func(cmd *cobra.Command, args []string) error {
			version := commands.NewVersionCommand()
			return version.Execute(context.Background(), args, engine.CommandContext{})
		},
	}
}

func newHelpCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "help",
		Short: "Show help information",
		RunE: func(cmd *cobra.Command, args []string) error {
			help := commands.NewHelpCommand()
			return help.Execute(context.Background(), args, engine.CommandContext{})
		},
	}
}

func newTuiCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "tui",
		Short: "Start the terminal UI",
		RunE: func(cmd *cobra.Command, args []string) error {
			return ui.Run()
		},
	}
}
