package commands

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/claude-code-go/claude/internal/engine"
	"github.com/claude-code-go/claude/pkg/llm"
)

var (
	connectTitleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("cyan")).
				Bold(true)

	connectValueStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("green"))

	connectErrorStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("red"))

	connectMutedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("gray"))
)

// ProviderManager 提供者管理器
type ProviderManager struct {
	registry *llm.ProviderRegistry
}

// NewProviderManager 创建新的提供者管理器
func NewProviderManager() *ProviderManager {
	return &ProviderManager{
		registry: llm.DefaultProviderRegistry(),
	}
}

// ConnectCommand /connect 命令
type ConnectCommand struct {
	BaseCommand
	manager *ProviderManager
}

// NewConnectCommand 创建 ConnectCommand
func NewConnectCommand() *ConnectCommand {
	return &ConnectCommand{
		BaseCommand: *newCommand("connect", "Connect to an LLM provider"),
		manager:     NewProviderManager(),
	}
}

// Execute 执行 /connect 命令
func (c *ConnectCommand) Execute(ctx context.Context, args []string, execCtx engine.CommandContext) error {
	if len(args) == 0 {
		return c.listProviders()
	}

	subcommand := args[0]

	switch subcommand {
	case "list", "ls":
		return c.listProviders()
	case "current":
		return c.currentProvider()
	case "set":
		if len(args) < 2 {
			return fmt.Errorf("Usage: /connect set <provider>")
		}
		return c.setProvider(args[1])
	case "add":
		if len(args) < 2 {
			return fmt.Errorf("Usage: /connect add <provider> [--base-url <url>] [--api-key <key>]")
		}
		return c.addProvider(args[1], args[2:])
	case "remove", "rm":
		if len(args) < 2 {
			return fmt.Errorf("Usage: /connect remove <provider>")
		}
		return c.removeProvider(args[1])
	case "test":
		return c.testProvider()
	case "env":
		return c.showEnvVars()
	default:
		return fmt.Errorf("Unknown subcommand: %s. Use 'list', 'set', 'add', 'remove', 'test', or 'env'", subcommand)
	}
}

// listProviders 列出所有可用的 Provider
func (c *ConnectCommand) listProviders() error {
	fmt.Println(connectTitleStyle.Render("Available Providers:"))
	fmt.Println(strings.Repeat("─", 50))
	fmt.Println()

	providers := c.manager.registry.List()
	current, _ := c.manager.registry.GetDefault()
	currentName := ""
	if current != nil {
		currentName = current.Name()
	}

	for _, name := range providers {
		provider, _ := c.manager.registry.Get(name)
		config := provider.Config()

		// 检查 API Key 是否配置
		hasKey := config.APIKey != ""
		keyStatus := connectMutedStyle.Render("✗ no API key")
		if hasKey {
			keyStatus = connectValueStyle.Render("✓ configured")
		}

		// 标记当前 Provider
		marker := "  "
		if name == currentName {
			marker = connectValueStyle.Render("►")
		}

		fmt.Printf("%s %s %s\n", marker, connectValueStyle.Render(name), keyStatus)
		fmt.Printf("   %s\n", connectMutedStyle.Render(config.BaseURL))
		if config.DefaultModel != "" {
			fmt.Printf("   Default: %s\n", config.DefaultModel)
		}
		if config.CodingModel != "" {
			fmt.Printf("   Coding:  %s\n", config.CodingModel)
		}
		fmt.Println()
	}

	return nil
}

// currentProvider 显示当前 Provider
func (c *ConnectCommand) currentProvider() error {
	provider, err := c.manager.registry.GetDefault()
	if err != nil {
		return fmt.Errorf("No default provider set. Use /connect set <provider> to set one.")
	}

	config := provider.Config()

	fmt.Println(connectTitleStyle.Render("Current Provider:"))
	fmt.Println(strings.Repeat("─", 50))
	fmt.Println()

	fmt.Printf("Name:     %s\n", connectValueStyle.Render(config.Name))
	fmt.Printf("Base URL: %s\n", config.BaseURL)
	fmt.Printf("Default:  %s\n", config.DefaultModel)
	fmt.Printf("Coding:   %s\n", config.CodingModel)

	// 检查 API Key
	if config.APIKey != "" {
		masked := config.APIKey[:8] + "..." + config.APIKey[len(config.APIKey)-4:]
		fmt.Printf("API Key:  %s\n", connectValueStyle.Render(masked))
	} else if config.APIKeyEnv != "" {
		fmt.Printf("API Key:  %s (env: %s)\n", connectErrorStyle.Render("not set"), config.APIKeyEnv)
	} else {
		fmt.Printf("API Key:  %s\n", connectErrorStyle.Render("not configured"))
	}

	// 能力
	fmt.Println()
	fmt.Println("Capabilities:")
	fmt.Printf("  Tools:  %v\n", config.Capabilities.SupportsTools)
	fmt.Printf("  Vision: %v\n", config.Capabilities.SupportsVision)
	fmt.Printf("  Thinking: %v\n", config.Capabilities.SupportsThinking)

	return nil
}

// setProvider 设置当前 Provider
func (c *ConnectCommand) setProvider(name string) error {
	err := c.manager.registry.SetDefault(name)
	if err != nil {
		return fmt.Errorf("Provider not found: %s. Use /connect list to see available providers.", name)
	}

	provider, _ := c.manager.registry.Get(name)
	config := provider.Config()

	fmt.Printf("Switched to provider: %s\n", connectValueStyle.Render(name))
	fmt.Printf("Base URL: %s\n", config.BaseURL)
	fmt.Printf("Default model: %s\n", config.DefaultModel)

	// 如果 API Key 未配置，给出警告
	if config.APIKey == "" {
		fmt.Println()
		fmt.Printf("%s No API key configured for %s. ",
			connectErrorStyle.Render("Warning:"), name)
		fmt.Printf("Set with: export %s=<your-api-key>\n", config.APIKeyEnv)
	}

	return nil
}

// addProvider 添加自定义 Provider
func (c *ConnectCommand) addProvider(name string, args []string) error {
	// 解析参数
	baseURL := ""
	apiKey := ""
	defaultModel := ""

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--base-url", "-u":
			if i+1 < len(args) {
				baseURL = args[i+1]
				i++
			}
		case "--api-key", "-k":
			if i+1 < len(args) {
				apiKey = args[i+1]
				i++
			}
		case "--model", "-m":
			if i+1 < len(args) {
				defaultModel = args[i+1]
				i++
			}
		}
	}

	if baseURL == "" {
		return fmt.Errorf("Base URL is required. Use: /connect add %s --base-url https://...", name)
	}

	// 创建自定义 Provider 配置
	config := llm.NewProviderConfigBuilder(name).
		WithBaseURL(baseURL).
		WithAPIKey(apiKey).
		WithDefaultModel(defaultModel).
		Build()

	provider := llm.NewOpenAIClient(config)
	c.manager.registry.Register(name, provider)

	fmt.Printf("Added provider: %s\n", connectValueStyle.Render(name))
	fmt.Printf("Base URL: %s\n", baseURL)
	if defaultModel != "" {
		fmt.Printf("Default model: %s\n", defaultModel)
	}

	return nil
}

// removeProvider 移除 Provider
func (c *ConnectCommand) removeProvider(name string) error {
	// 不能移除内置 Provider
	builtinProviders := []string{"openai", "deepseek", "kimi", "bailian", "zhipu", "minimax", "opencode-zen", "opencode-go", "ollama"}
	for _, bp := range builtinProviders {
		if name == bp {
			return fmt.Errorf("Cannot remove built-in provider: %s", name)
		}
	}

	provider, ok := c.manager.registry.Get(name)
	if !ok {
		return fmt.Errorf("Provider not found: %s", name)
	}

	// 如果是当前 Provider，需要先切换
	current, _ := c.manager.registry.GetDefault()
	if current != nil && current.Name() == name {
		fmt.Printf("Warning: %s is currently active. Switching to openai first.\n", name)
		c.manager.registry.SetDefault("openai")
	}

	// 在 Go 中无法真正"移除"，只能标记
	_ = provider
	fmt.Printf("Provider '%s' would be removed. (Removal not fully implemented yet)\n", name)

	return nil
}

// testProvider 测试当前 Provider 连接
func (c *ConnectCommand) testProvider() error {
	provider, err := c.manager.registry.GetDefault()
	if err != nil {
		return fmt.Errorf("No default provider set. Use /connect set <provider> first.")
	}

	config := provider.Config()

	if config.APIKey == "" {
		return fmt.Errorf("No API key configured for %s. Set with: export %s=<your-api-key>",
			config.Name, config.APIKeyEnv)
	}

	fmt.Printf("Testing connection to %s...\n", connectValueStyle.Render(config.Name))
	fmt.Printf("Base URL: %s\n", config.BaseURL)

	// 发送测试请求
	req := &llm.MessageRequest{
		Model: config.DefaultModel,
		Messages: []llm.Message{
			{Role: "user", Content: "Hello! Reply with just 'OK' if you receive this."},
		},
		MaxTokens: 10,
	}

	resp, err := provider.CreateMessage(context.Background(), req)
	if err != nil {
		return fmt.Errorf("Connection failed: %v", err)
	}

	if resp.Error != nil {
		return fmt.Errorf("API error: %s - %s", resp.Error.Type, resp.Error.Message)
	}

	fmt.Println()
	fmt.Printf("%s Connection successful!\n", connectValueStyle.Render("✓"))
	fmt.Printf("Model: %s\n", resp.Model)

	return nil
}

// showEnvVars 显示所有 Provider 相关的环境变量
func (c *ConnectCommand) showEnvVars() error {
	fmt.Println(connectTitleStyle.Render("Provider Environment Variables:"))
	fmt.Println(strings.Repeat("─", 50))
	fmt.Println()

	envVars := []struct {
		Name        string
		Description string
	}{
		{"OPENAI_API_KEY", "OpenAI API Key"},
		{"DEEPSEEK_API_KEY", "DeepSeek API Key"},
		{"MOONSHOT_API_KEY", "Kimi (Moonshot) API Key"},
		{"DASHSCOPE_API_KEY", "百炼 (Bailian) API Key"},
		{"BIGMODEL_API_KEY", "智谱 (BigModel) API Key"},
		{"MINIMAX_API_KEY", "MiniMax API Key"},
		{"OPENCODE_ZEN_API_KEY", "OpenCode Zen API Key"},
		{"OPENCODE_GO_API_KEY", "OpenCode Go API Key"},
		{"OLLAMA_BASE_URL", "Ollama Base URL (default: http://localhost:11434/v1)"},
	}

	for _, ev := range envVars {
		value := os.Getenv(ev.Name)
		status := connectMutedStyle.Render("✗ not set")
		if value != "" {
			masked := value[:min(len(value), 8)] + "..."
			status = connectValueStyle.Render(masked)
		}
		fmt.Printf("%s: %s\n", ev.Name, status)
		fmt.Printf("   %s\n", ev.Description)
	}

	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
