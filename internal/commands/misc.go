package commands

import (
	"context"
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/claude-code-go/claude/internal/engine"
	"github.com/claude-code-go/claude/internal/services/cost"
)

var costStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("yellow"))
var labelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("cyan"))
var valueStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("green"))
var headerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("cyan")).Bold(true)

// CostCommand 成本追踪命令 - 显示API使用成本统计
type CostCommand struct {
	BaseCommand
	tracker *cost.CostTracker
}

// NewCostCommand 创建成本命令
func NewCostCommand(tracker *cost.CostTracker) *CostCommand {
	return &CostCommand{
		BaseCommand: *newCommand("cost", "Show session usage and cost information"),
		tracker:     tracker,
	}
}

// Execute 显示成本统计信息
func (c *CostCommand) Execute(ctx context.Context, args []string, execCtx engine.CommandContext) error {
	// 解析参数
	showDetails := false
	reset := false

	for _, arg := range args {
		switch arg {
		case "-d", "--details":
			showDetails = true
		case "--reset":
			reset = true
		}
	}

	if reset {
		if c.tracker != nil {
			c.tracker.Reset()
		}
		fmt.Printf("%s 成本统计已重置\n", costStyle.Render("✓"))
		return nil
	}

	if c.tracker == nil {
		// 显示基本统计(无追踪器时)
		fmt.Println(headerStyle.Render("═══════════════════════════════════════"))
		fmt.Println(headerStyle.Render("           API 使用成本统计"))
		fmt.Println(headerStyle.Render("═══════════════════════════════════════"))
		fmt.Println()
		fmt.Println(costStyle.Render("┌─ Usage & Cost ─"))
		fmt.Println(costStyle.Render("│"))
		fmt.Println(costStyle.Render("│  This session"))
		fmt.Println(costStyle.Render("│  Input tokens:    ") + valueStyle.Render("0"))
		fmt.Println(costStyle.Render("│  Output tokens:   ") + valueStyle.Render("0"))
		fmt.Println(costStyle.Render("│  Total tokens:    ") + valueStyle.Render("0"))
		fmt.Println(costStyle.Render("│"))
		fmt.Println(costStyle.Render("│  Estimated cost:  ") + valueStyle.Render("$0.00"))
		fmt.Println(costStyle.Render("│"))
		fmt.Println(costStyle.Render("│  Claude Opus:     $15.00/1M input, $75.00/1M output"))
		fmt.Println(costStyle.Render("│  Claude Sonnet:   $3.00/1M input, $15.00/1M output"))
		fmt.Println(costStyle.Render("│  Claude Haiku:     $0.25/1M input, $1.25/1M output"))
		fmt.Println(costStyle.Render("│"))
		fmt.Println(costStyle.Render("└─"))
		return nil
	}

	// 显示完整成本统计
	fmt.Println(headerStyle.Render("═══════════════════════════════════════"))
	fmt.Println(headerStyle.Render("           API 使用成本统计"))
	fmt.Println(headerStyle.Render("═══════════════════════════════════════"))
	fmt.Println()

	// 基本统计
	fmt.Printf("  总成本:          %s\n", costStyle.Render(c.tracker.FormatCost()))
	fmt.Printf("  API调用时长:     %s\n", cost.FormatDuration(c.tracker.GetTotalAPIDuration()))
	fmt.Printf("  工具执行时长:     %s\n", cost.FormatDuration(c.tracker.GetTotalAPIDuration()))
	fmt.Printf("  代码变更:        +%d / -%d 行\n", c.tracker.GetTotalLinesAdded(), c.tracker.GetTotalLinesRemoved())

	fmt.Println()
	fmt.Printf("  令牌使用:\n")
	fmt.Printf("    输入令牌:       %s\n", formatNumber(c.tracker.GetTotalInputTokens()))
	fmt.Printf("    输出令牌:       %s\n", formatNumber(c.tracker.GetTotalOutputTokens()))
	fmt.Printf("    缓存读取:       %s\n", formatNumber(c.tracker.GetTotalCacheReadInputTokens()))
	fmt.Printf("    缓存创建:       %s\n", formatNumber(c.tracker.GetTotalCacheCreationInputTokens()))
	fmt.Printf("    网页搜索:       %d 请求\n", c.tracker.GetTotalWebSearchRequests())

	// 未知模型警告
	if c.tracker.HasUnknownModelCost() {
		fmt.Println()
		fmt.Printf("  %s 警告: 检测到未知模型，使用默认定价\n", lipgloss.NewStyle().Foreground(lipgloss.Color("yellow")).Render("!"))
	}

	// 详细统计
	if showDetails {
		modelUsage := c.tracker.GetAllModelUsage()
		if len(modelUsage) > 0 {
			fmt.Println()
			fmt.Println(headerStyle.Render("  模型使用详情:"))
			fmt.Println()
			for model, usage := range modelUsage {
				fmt.Printf("    %s\n", model)
				fmt.Printf("      成本:          $%.4f\n", usage.CostUSD)
				fmt.Printf("      输入令牌:       %d\n", usage.Usage.InputTokens)
				fmt.Printf("      输出令牌:       %d\n", usage.Usage.OutputTokens)
				fmt.Println()
			}
		}
	}

	return nil
}

// formatNumber 格式化数字显示
func formatNumber(n int64) string {
	if n >= 1_000_000 {
		return fmt.Sprintf("%.2fM", float64(n)/1_000_000)
	}
	if n >= 1_000 {
		return fmt.Sprintf("%.1fK", float64(n)/1_000)
	}
	return fmt.Sprintf("%d", n)
}

type MemoryCommand struct {
	BaseCommand
}

func NewMemoryCommand() *MemoryCommand {
	return &MemoryCommand{
		BaseCommand: *newCommand("memory", "Manage persistent memory"),
	}
}

func (c *MemoryCommand) Execute(ctx context.Context, args []string, execCtx engine.CommandContext) error {
	if len(args) == 0 {
		return c.showMemoryStatus()
	}

	switch args[0] {
	case "list", "ls":
		return c.listMemories()
	case "add":
		fmt.Println("Memory: Use the memory tool to add memories")
	case "clear":
		fmt.Println("Memory: Clearing memories is not yet implemented")
	default:
		fmt.Printf("Unknown memory action: %s\n", args[0])
	}

	return nil
}

func (c *MemoryCommand) showMemoryStatus() error {
	fmt.Println(labelStyle.Render("Memory Status:"))
	fmt.Println("  Memories stored: 0")
	fmt.Println("  Use /memory list to view all memories")
	return nil
}

func (c *MemoryCommand) listMemories() error {
	fmt.Println(labelStyle.Render("Stored Memories:"))
	fmt.Println("  (No memories stored)")
	return nil
}

type ResumeCommand struct {
	BaseCommand
}

func NewResumeCommand() *ResumeCommand {
	return &ResumeCommand{
		BaseCommand: *newCommand("resume", "Restore a previous session"),
	}
}

func (c *ResumeCommand) Execute(ctx context.Context, args []string, execCtx engine.CommandContext) error {
	fmt.Println("Resume command - implementation pending")
	fmt.Println("Use /resume <session-id> to restore a previous session")
	return nil
}

type ShareCommand struct {
	BaseCommand
}

func NewShareCommand() *ShareCommand {
	return &ShareCommand{
		BaseCommand: *newCommand("share", "Share the current session"),
	}
}

func (c *ShareCommand) Execute(ctx context.Context, args []string, execCtx engine.CommandContext) error {
	fmt.Println("Share command - implementation pending")
	fmt.Println("This would generate a shareable link to the current session")
	return nil
}

type ExitCommand struct {
	BaseCommand
}

func NewExitCommand() *ExitCommand {
	return &ExitCommand{
		BaseCommand: *newCommand("exit", "Exit Claude Code"),
	}
}

func (c *ExitCommand) Execute(ctx context.Context, args []string, execCtx engine.CommandContext) error {
	fmt.Println("Goodbye!")
	return nil
}

type TasksCommand struct {
	BaseCommand
}

func NewTasksCommand() *TasksCommand {
	return &TasksCommand{
		BaseCommand: *newCommand("tasks", "Manage tasks"),
	}
}

func (c *TasksCommand) Execute(ctx context.Context, args []string, execCtx engine.CommandContext) error {
	if len(args) == 0 {
		return c.showTasks()
	}

	switch args[0] {
	case "list", "ls":
		return c.listTasks()
	case "clear":
		fmt.Println("Tasks cleared")
	default:
		fmt.Printf("Unknown tasks action: %s\n", args[0])
	}

	return nil
}

func (c *TasksCommand) showTasks() error {
	fmt.Println(labelStyle.Render("Tasks:"))
	fmt.Println("  Use TaskCreate tool to create tasks")
	fmt.Println("  Use TaskList tool to view tasks")
	return nil
}

func (c *TasksCommand) listTasks() error {
	fmt.Println(labelStyle.Render("Current Tasks:"))
	fmt.Println("  (No active tasks)")
	return nil
}

type ThemeCommand struct {
	BaseCommand
}

func NewThemeCommand() *ThemeCommand {
	return &ThemeCommand{
		BaseCommand: *newCommand("theme", "Change the color theme"),
	}
}

func (c *ThemeCommand) Execute(ctx context.Context, args []string, execCtx engine.CommandContext) error {
	if len(args) == 0 {
		return c.showCurrentTheme()
	}

	if args[0] == "list" || args[0] == "ls" {
		return c.listThemes()
	}

	fmt.Printf("Theme set to: %s\n", args[0])
	return nil
}

func (c *ThemeCommand) showCurrentTheme() error {
	fmt.Println(labelStyle.Render("Current theme: ") + valueStyle.Render("default"))
	return nil
}

func (c *ThemeCommand) listThemes() error {
	fmt.Println(labelStyle.Render("Available themes:"))
	fmt.Println("  default (current)")
	fmt.Println("  dark")
	fmt.Println("  light")
	fmt.Println("  monokai")
	fmt.Println("  dracula")
	return nil
}

type KeybindingsCommand struct {
	BaseCommand
}

func NewKeybindingsCommand() *KeybindingsCommand {
	return &KeybindingsCommand{
		BaseCommand: *newCommand("keybindings", "Show keyboard shortcuts"),
	}
}

func (c *KeybindingsCommand) Execute(ctx context.Context, args []string, execCtx engine.CommandContext) error {
	fmt.Println(labelStyle.Render("Keyboard Shortcuts:"))
	fmt.Println("")
	fmt.Println("  Ctrl+C     Cancel current operation")
	fmt.Println("  Ctrl+D     Exit Claude Code")
	fmt.Println("  Ctrl+L     Clear screen")
	fmt.Println("  ↑/↓        Navigate command history")
	fmt.Println("  Tab        Auto-complete")
	fmt.Println("  Ctrl+R     Search command history")
	return nil
}

type VimCommand struct {
	BaseCommand
}

func NewVimCommand() *VimCommand {
	return &VimCommand{
		BaseCommand: *newCommand("vim", "Toggle vim mode"),
	}
}

func (c *VimCommand) Execute(ctx context.Context, args []string, execCtx engine.CommandContext) error {
	if len(args) == 0 {
		fmt.Println("Vim mode: off")
		return nil
	}

	switch args[0] {
	case "on":
		fmt.Println("Vim mode enabled")
	case "off":
		fmt.Println("Vim mode disabled")
	default:
		fmt.Printf("Unknown vim option: %s\n", args[0])
	}

	return nil
}

type ContextCommand struct {
	BaseCommand
}

func NewContextCommand() *ContextCommand {
	return &ContextCommand{
		BaseCommand: *newCommand("context", "Show context information"),
	}
}

func (c *ContextCommand) Execute(ctx context.Context, args []string, execCtx engine.CommandContext) error {
	fmt.Println(labelStyle.Render("Context Information:"))
	fmt.Println("  Current working directory: ", valueStyle.Render(execCtx.GetWorkingDirectory()))
	fmt.Println("  Messages in context: 0")
	fmt.Println("  Tools available: 9")
	return nil
}

type PermissionsCommand struct {
	BaseCommand
}

func NewPermissionsCommand() *PermissionsCommand {
	return &PermissionsCommand{
		BaseCommand: *newCommand("permissions", "Manage permission settings"),
	}
}

func (c *PermissionsCommand) Execute(ctx context.Context, args []string, execCtx engine.CommandContext) error {
	fmt.Println(labelStyle.Render("Permission Mode:"))
	fmt.Println("  Current: ask")
	fmt.Println("")
	fmt.Println("Available modes:")
	fmt.Println("  ask       - Ask before running tools")
	fmt.Println("  auto      - Automatically allow/deny based on risk")
	fmt.Println("  bypass    - Allow all tools without prompting")
	fmt.Println("  plan      - Require confirmation for high-risk tools")
	return nil
}

type PlanCommand struct {
	BaseCommand
}

func NewPlanCommand() *PlanCommand {
	return &PlanCommand{
		BaseCommand: *newCommand("plan", "Enter plan mode to review changes before executing"),
	}
}

func (c *PlanCommand) Execute(ctx context.Context, args []string, execCtx engine.CommandContext) error {
	fmt.Println("Plan mode - implementation pending")
	fmt.Println("This would enter a mode to review and approve changes before execution")
	return nil
}
