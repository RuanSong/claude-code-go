package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func ConfigCommand() *cobra.Command {
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

	cmd.PersistentFlags().StringVar(&configPath, "config", "", "config file path")
	return cmd
}

func LoginCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "login",
		Short: "Login to Claude",
		Long:  "Authenticate with your Anthropic API key",
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: implement login
			fmt.Println("Login command - implementation pending")
			return nil
		},
	}
}

func LogoutCommand() *cobra.Command {
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

func DoctorCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Run diagnostics",
		Long:  "Check environment and configuration for issues",
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: implement doctor
			fmt.Println("Doctor command - implementation pending")
			return nil
		},
	}
}

func ModelCommand() *cobra.Command {
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
					return fmt.Errorf("save config: %w", err)
				}
				fmt.Printf("Model set to: %s\n", args[0])
			}
			return nil
		},
	}
}
