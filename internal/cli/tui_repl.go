package cli

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbletea"
	"github.com/claude-code-go/claude/internal/commands"
	"github.com/claude-code-go/claude/internal/engine"
	"github.com/claude-code-go/claude/internal/tui"
	"github.com/claude-code-go/claude/pkg/anthropic"
)

// TUI REPL配置
type TUIREPLConfig struct {
	APIKey string
	Model  string
}

// TUIREPL 基于Bubble Tea的交互式REPL
type TUIREPL struct {
	engine   *engine.QueryEngine
	commands *commands.Registry
	config   TUIREPLConfig
	model    *tui.MainModel
}

// NewTUIREPL 创建新的TUI REPL
func NewTUIREPL(config TUIREPLConfig) *TUIREPL {
	return &TUIREPL{
		commands: commands.DefaultRegistry(),
		config:   config,
		model:    tui.NewMainModel(),
	}
}

// Run 启动TUI REPL
func (r *TUIREPL) Run() error {
	// 初始化API客户端
	apiClient := anthropic.NewClient(anthropic.Config{
		APIKey: r.config.APIKey,
		Model:  r.config.Model,
	})

	// 初始化工具注册表
	toolRegistry := engine.NewToolRegistry()
	for _, tool := range GetBuiltInTools() {
		if err := toolRegistry.Register(tool); err != nil {
			return fmt.Errorf("register tool: %w", err)
		}
	}

	// 初始化查询引擎
	qe := engine.NewQueryEngine(engine.Config{
		Model:     r.config.Model,
		Tools:     toolRegistry,
		MaxTurns:  100,
		MaxTokens: 200000,
	}, apiClient)
	r.engine = qe

	// 添加欢迎消息
	r.model.AddMessage("system", "Claude Code REPL - Type /help for commands")
	r.model.AddMessage("assistant", "Hello! I'm Claude Code. How can I help you today?")

	// 创建Bubble Tea程序
	p := tea.NewProgram(r.model, tea.WithAltScreen())

	// 运行程序
	if err := p.Start(); err != nil {
		return fmt.Errorf("TUI run error: %w", err)
	}

	return nil
}

// HandleMessage 处理用户消息
func (r *TUIREPL) HandleMessage(input string) error {
	// 检查命令
	if strings.HasPrefix(input, "/") {
		return r.handleCommand(input)
	}

	// 添加用户消息
	r.model.AddMessage("user", input)

	// 开始思考动画
	r.model.StartThinking("Thinking...")

	// 在后台处理消息
	go func() {
		ctx := context.Background()
		err := r.engine.SubmitMessage(ctx, input)

		// 停止思考动画
		r.model.StopThinking()

		if err != nil {
			r.model.AddMessage("assistant", fmt.Sprintf("Error: %v", err))
		} else {
			r.model.AddMessage("assistant", "Response from Claude...")
		}
	}()

	return nil
}

// handleCommand 处理命令
func (r *TUIREPL) handleCommand(input string) error {
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
		Cwd: getCurrentDir(),
	}

	ctx := context.Background()
	if err := cmd.Execute(ctx, args, execCtx); err != nil {
		return err
	}

	r.model.AddMessage("assistant", fmt.Sprintf("Command /%s executed", cmdName))
	return nil
}

// getCurrentDir 获取当前目录
func getCurrentDir() string {
	cwd, err := os.Getwd()
	if err != nil {
		return "."
	}
	return cwd
}

// SimpleREPL 简单的非TUI REPL（备用）
type SimpleREPL struct {
	engine   *engine.QueryEngine
	commands *commands.Registry
	config   TUIREPLConfig
}

// NewSimpleREPL 创建简单的REPL
func NewSimpleREPL(config TUIREPLConfig) *SimpleREPL {
	return &SimpleREPL{
		commands: commands.DefaultRegistry(),
		config:   config,
	}
}

// Run 运行简单REPL
func (r *SimpleREPL) Run() error {
	r.printWelcome()

	// 初始化
	apiClient := anthropic.NewClient(anthropic.Config{
		APIKey: r.config.APIKey,
		Model:  r.config.Model,
	})

	toolRegistry := engine.NewToolRegistry()
	for _, tool := range GetBuiltInTools() {
		if err := toolRegistry.Register(tool); err != nil {
			return fmt.Errorf("register tool: %w", err)
		}
	}

	qe := engine.NewQueryEngine(engine.Config{
		Model:     r.config.Model,
		Tools:     toolRegistry,
		MaxTurns:  100,
		MaxTokens: 200000,
	}, apiClient)
	r.engine = qe

	// 主循环
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

func (r *SimpleREPL) printWelcome() {
	fmt.Println(headerStyle.Render("Claude Code REPL"))
	fmt.Println("Type '/help' for available commands or enter a message to chat.")
	fmt.Println()
}

func (r *SimpleREPL) readInput() (string, error) {
	fmt.Print(promptStyle.Render("> "))
	var input string
	_, err := fmt.Scanln(&input)
	return strings.TrimSpace(input), err
}

func (r *SimpleREPL) processInput(input string) error {
	if strings.HasPrefix(input, "/") {
		return r.handleCommand(input)
	}

	ctx := context.Background()
	return r.engine.SubmitMessage(ctx, input)
}

func (r *SimpleREPL) handleCommand(input string) error {
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
		Cwd: getCurrentDir(),
	}

	ctx := context.Background()
	if err := cmd.Execute(ctx, args, execCtx); err != nil {
		return err
	}

	fmt.Println(successStyle.Render("Done"))
	return nil
}
