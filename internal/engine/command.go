package engine

import (
	"context"
	"fmt"
	"strings"
)

// 命令类型
// 对应 TypeScript: slash命令类型
type CommandType int

const (
	CommandTypePrompt   CommandType = iota // 提示词类型 - 生成提示词内容
	CommandTypeCustom                      // 自定义类型 - 自定义实现
	CommandTypeLocal                       // 本地类型 - 本地命令
	CommandTypeLocalJSX                    // 本地JSX类型 - React组件
)

func (c CommandType) String() string {
	switch c {
	case CommandTypePrompt:
		return "prompt"
	case CommandTypeCustom:
		return "custom"
	case CommandTypeLocal:
		return "local"
	case CommandTypeLocalJSX:
		return "local-jsx"
	default:
		return "unknown"
	}
}

// CommandContext 命令执行上下文
// 对应 TypeScript: 命令执行环境
// 提供命令执行所需的信息
type CommandContext struct {
	Cwd       string        // 当前工作目录
	Config    *Config       // 配置信息
	LLMClient interface{}   // LLM客户端
	ToolSys   *ToolRegistry // 工具系统
	UI        interface{}   // UI接口
}

// GetWorkingDirectory 获取当前工作目录
func (c *CommandContext) GetWorkingDirectory() string {
	return c.Cwd
}

// PromptCommand 提示词命令接口
// 对应 TypeScript: prompt命令
// 提供额外方法用于获取允许的工具和提示词模板
type PromptCommand interface {
	Command
	GetAllowedTools() []string
	GetPromptTemplate() string
}

// Command 命令接口
// 对应 TypeScript: slash命令
// 所有slash命令必须实现的接口
type Command interface {
	Type() CommandType                                                        // 命令类型
	Name() string                                                             // 命令名称
	Description() string                                                      // 命令描述
	Execute(ctx context.Context, args []string, execCtx CommandContext) error // 执行命令
}

// BaseCommand 命令基础实现
// 提供命令的默认实现
type BaseCommand struct {
	cmdType     CommandType // 命令类型
	name        string      // 命令名称
	description string      // 命令描述
}

func NewBaseCommand(cmdType CommandType, name, description string) *BaseCommand {
	return &BaseCommand{
		cmdType:     cmdType,
		name:        name,
		description: description,
	}
}

func (b *BaseCommand) Type() CommandType   { return b.cmdType }
func (b *BaseCommand) Name() string        { return b.name }
func (b *BaseCommand) Description() string { return b.description }

// CommandRegistry 命令注册表
// 对应 TypeScript: 命令管理器
// 负责命令的注册和查找
type CommandRegistry struct {
	commands map[string]Command // 命令名称 -> 命令实例
}

func NewCommandRegistry() *CommandRegistry {
	return &CommandRegistry{
		commands: make(map[string]Command),
	}
}

// Register 注册命令
// 对应 TypeScript: 注册slash命令
func (r *CommandRegistry) Register(cmd Command) error {
	name := cmd.Name()
	if name == "" {
		return fmt.Errorf("command name cannot be empty")
	}
	if _, exists := r.commands[name]; exists {
		return fmt.Errorf("command already registered: %s", name)
	}
	r.commands[name] = cmd
	return nil
}

// Get 获取命令
// 对应 TypeScript: 获取命令
func (r *CommandRegistry) Get(name string) (Command, bool) {
	cmd, ok := r.commands[name]
	return cmd, ok
}

// List 列出所有命令
func (r *CommandRegistry) List() []Command {
	cmds := make([]Command, 0, len(r.commands))
	for _, cmd := range r.commands {
		cmds = append(cmds, cmd)
	}
	return cmds
}

// GetByPrefix 按前缀查找命令
// 对应 TypeScript: 命令前缀匹配
func (r *CommandRegistry) GetByPrefix(prefix string) (Command, bool) {
	prefix = strings.TrimPrefix(prefix, "/")
	for name, cmd := range r.commands {
		if strings.HasPrefix(name, prefix) {
			return cmd, true
		}
	}
	return nil, false
}

// CommitCommand git提交命令
// 对应 TypeScript: /commit 命令
type CommitCommand struct {
	BaseCommand
}

func NewCommitCommand() *CommitCommand {
	return &CommitCommand{
		BaseCommand: *NewBaseCommand(CommandTypePrompt, "commit", "Create a git commit"),
	}
}

func (c *CommitCommand) Execute(ctx context.Context, args []string, execCtx CommandContext) error {
	fmt.Println("Commit command - implementation pending")
	return nil
}

// ReviewCommand 代码审查命令
// 对应 TypeScript: /review 命令
type ReviewCommand struct {
	BaseCommand
}

func NewReviewCommand() *ReviewCommand {
	return &ReviewCommand{
		BaseCommand: *NewBaseCommand(CommandTypePrompt, "review", "Review code changes"),
	}
}

func (c *ReviewCommand) Execute(ctx context.Context, args []string, execCtx CommandContext) error {
	fmt.Println("Review command - implementation pending")
	return nil
}

// ConfigCommand 配置管理命令
// 对应 TypeScript: /config 命令
type ConfigCommand struct {
	BaseCommand
}

func NewConfigCommand() *ConfigCommand {
	return &ConfigCommand{
		BaseCommand: *NewBaseCommand(CommandTypeCustom, "config", "Manage configuration settings"),
	}
}

func (c *ConfigCommand) Execute(ctx context.Context, args []string, execCtx CommandContext) error {
	fmt.Println("Config command - implementation pending")
	return nil
}
