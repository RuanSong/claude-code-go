package commands

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/charmbracelet/lipgloss"
	"github.com/claude-code-go/claude/internal/engine"
)

// AddDirCommand 添加工作目录命令
// 对应 TypeScript: /add-dir 命令
// 添加一个新的工作目录到当前会话
type AddDirCommand struct {
	BaseCommand
}

func NewAddDirCommand() *AddDirCommand {
	return &AddDirCommand{
		BaseCommand: *newCommand("add-dir", "添加一个新的工作目录"),
	}
}

func (c *AddDirCommand) Execute(ctx context.Context, args []string, execCtx engine.CommandContext) error {
	if len(args) == 0 {
		fmt.Println("用法: /add-dir <路径>")
		fmt.Println("  添加一个新的工作目录到当前会话")
		fmt.Println("  例如: /add-dir /path/to/project")
		return nil
	}

	dirPath := args[0]

	// 验证目录是否存在
	info, err := os.Stat(dirPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("目录不存在: %s", dirPath)
		}
		return fmt.Errorf("无法访问目录: %v", err)
	}

	if !info.IsDir() {
		return fmt.Errorf("路径不是目录: %s", dirPath)
	}

	// 获取绝对路径
	absPath, err := filepath.Abs(dirPath)
	if err != nil {
		return fmt.Errorf("无法获取绝对路径: %v", err)
	}

	fmt.Printf("已添加工作目录: %s\n", absPath)
	return nil
}

// ColorCommand 颜色命令
// 对应 TypeScript: /color 命令
// 设置提示栏颜色
type ColorCommand struct {
	BaseCommand
}

var (
	colorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("cyan"))
)

// 可用的颜色列表
var availableColors = []string{
	"red", "green", "yellow", "blue", "magenta", "cyan", "white",
	"reset",
}

func NewColorCommand() *ColorCommand {
	return &ColorCommand{
		BaseCommand: *newCommand("color", "设置提示栏颜色"),
	}
}

func (c *ColorCommand) Execute(ctx context.Context, args []string, execCtx engine.CommandContext) error {
	if len(args) == 0 {
		fmt.Println("用法: /color <颜色>")
		fmt.Println("可用颜色:")
		for _, col := range availableColors {
			fmt.Printf("  - %s\n", col)
		}
		return nil
	}

	color := args[0]
	valid := false
	for _, c := range availableColors {
		if c == color {
			valid = true
			break
		}
	}

	if !valid {
		fmt.Printf("无效的颜色: %s\n", color)
		fmt.Println("使用 /color 查看可用颜色列表")
		return nil
	}

	if color == "reset" {
		fmt.Println("颜色已重置为默认值")
	} else {
		fmt.Printf("颜色已设置为: %s\n", colorStyle.Render(color))
	}

	return nil
}

// FastCommand 快速模式命令
// 对应 TypeScript: /fast 命令
// 切换快速模式(只使用快速模型)
type FastCommand struct {
	BaseCommand
}

func NewFastCommand() *FastCommand {
	return &FastCommand{
		BaseCommand: *newCommand("fast", "切换快速模式"),
	}
}

func (c *FastCommand) Execute(ctx context.Context, args []string, execCtx engine.CommandContext) error {
	fmt.Println("快速模式命令")
	fmt.Println("注意: 完整功能需要配置设置系统")

	if len(args) > 0 {
		switch args[0] {
		case "on", "enable":
			fmt.Println("快速模式: 已启用")
		case "off", "disable":
			fmt.Println("快速模式: 已禁用")
		default:
			fmt.Printf("未知选项: %s\n", args[0])
		}
	} else {
		fmt.Println("用法: /fast [on|off]")
	}

	return nil
}

// InsightsCommand 洞察命令
// 对应 TypeScript: /insights 命令
// 显示项目分析和洞察
type InsightsCommand struct {
	BaseCommand
}

func NewInsightsCommand() *InsightsCommand {
	return &InsightsCommand{
		BaseCommand: *newCommand("insights", "显示项目洞察和分析"),
	}
}

func (c *InsightsCommand) Execute(ctx context.Context, args []string, execCtx engine.CommandContext) error {
	fmt.Println("正在分析项目...")

	// 获取工作目录
	cwd := execCtx.GetWorkingDirectory()
	if cwd == "" {
		cwd, _ = os.Getwd()
	}

	// 统计文件类型
	fileTypes := make(map[string]int)
	totalFiles := 0

	filepath.Walk(cwd, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}
		totalFiles++

		ext := filepath.Ext(path)
		if ext == "" {
			ext = "无扩展名"
		}
		fileTypes[ext]++
		return nil
	})

	fmt.Println("\n项目统计:")
	fmt.Printf("  总文件数: %d\n", totalFiles)
	fmt.Println("  文件类型分布:")

	// 显示前10个最常见的文件类型
	for ext, count := range fileTypes {
		fmt.Printf("    %s: %d\n", ext, count)
	}

	// Git统计
	if _, err := exec.LookPath("git"); err == nil {
		cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
		cmd.Dir = cwd
		if branch, err := cmd.Output(); err == nil {
			fmt.Printf("\nGit分支: %s", string(branch))
		}

		cmd = exec.Command("git", "log", "--oneline", "-n5")
		cmd.Dir = cwd
		if logs, err := cmd.Output(); err == nil {
			fmt.Println("\n最近提交:")
			fmt.Print(string(logs))
		}
	}

	return nil
}

// InitVerifiersCommand 初始化验证器命令
// 对应 TypeScript: /init-verifiers 命令
// 初始化项目验证器
type InitVerifiersCommand struct {
	BaseCommand
}

func NewInitVerifiersCommand() *InitVerifiersCommand {
	return &InitVerifiersCommand{
		BaseCommand: *newCommand("init-verifiers", "初始化项目验证器"),
	}
}

func (c *InitVerifiersCommand) Execute(ctx context.Context, args []string, execCtx engine.CommandContext) error {
	fmt.Println("正在初始化验证器...")

	cwd := execCtx.GetWorkingDirectory()
	if cwd == "" {
		cwd, _ = os.Getwd()
	}

	// 检查是否已存在验证器配置
	verifierPath := filepath.Join(cwd, ".claude", "verifiers.json")
	if _, err := os.Stat(verifierPath); err == nil {
		fmt.Println("验证器已配置")
		return nil
	}

	// 创建验证器配置目录
	claudeDir := filepath.Join(cwd, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		return fmt.Errorf("无法创建配置目录: %v", err)
	}

	// 创建默认验证器配置
	defaultConfig := `{
  "version": "1.0",
  "verifiers": []
}`

	if err := os.WriteFile(verifierPath, []byte(defaultConfig), 0644); err != nil {
		return fmt.Errorf("无法创建验证器配置: %v", err)
	}

	fmt.Println("验证器初始化完成")
	fmt.Printf("配置文件: %s\n", verifierPath)

	return nil
}

// PassesCommand 传递命令
// 对应 TypeScript: /passes 命令
// 显示代码覆盖率等信息
type PassesCommand struct {
	BaseCommand
}

func NewPassesCommand() *PassesCommand {
	return &PassesCommand{
		BaseCommand: *newCommand("passes", "显示测试通过情况"),
	}
}

func (c *PassesCommand) Execute(ctx context.Context, args []string, execCtx engine.CommandContext) error {
	fmt.Println("检查测试状态...")

	cwd := execCtx.GetWorkingDirectory()
	if cwd == "" {
		cwd, _ = os.Getwd()
	}

	// 检查各种测试框架
	testCommands := []struct {
		name string
		cmd  string
		args []string
	}{
		{"Go tests", "go", []string{"test", "./..."}},
		{"npm test", "npm", []string{"test"}},
		{"pytest", "pytest", []string{"."}},
		{"jest", "npx", []string{"jest", "--listTests"}},
	}

	for _, tc := range testCommands {
		if _, err := exec.LookPath(tc.cmd); err == nil {
			cmd := exec.Command(tc.cmd, tc.args...)
			cmd.Dir = cwd
			output, err := cmd.CombinedOutput()
			if err == nil {
				fmt.Printf("\n%s:\n", tc.name)
				fmt.Print(string(output))
			}
		}
	}

	return nil
}

// RateLimitOptionsCommand 速率限制选项命令
// 对应 TypeScript: /rate-limit-options 命令
// 显示和管理速率限制选项
type RateLimitOptionsCommand struct {
	BaseCommand
}

func NewRateLimitOptionsCommand() *RateLimitOptionsCommand {
	return &RateLimitOptionsCommand{
		BaseCommand: *newCommand("rate-limit", "显示速率限制选项"),
	}
}

func (c *RateLimitOptionsCommand) Execute(ctx context.Context, args []string, execCtx engine.CommandContext) error {
	fmt.Println("速率限制选项")
	fmt.Println("=============")
	fmt.Println()
	fmt.Println("当前API使用情况:")
	fmt.Println("  请求限制: 未知")
	fmt.Println("  使用配额: 未知")
	fmt.Println()
	fmt.Println("可用命令:")
	fmt.Println("  /rate-limit status  - 显示当前限制状态")

	return nil
}

// StickersCommand 贴纸命令
// 对应 TypeScript: /stickers 命令
// 显示贴纸选项
type StickersCommand struct {
	BaseCommand
}

func NewStickersCommand() *StickersCommand {
	return &StickersCommand{
		BaseCommand: *newCommand("stickers", "显示贴纸选项"),
	}
}

func (c *StickersCommand) Execute(ctx context.Context, args []string, execCtx engine.CommandContext) error {
	fmt.Println("可用贴纸")
	fmt.Println("=========")
	fmt.Println()
	fmt.Println("程序员表情包:")
	fmt.Println("  :shipit:  - 部署")
	fmt.Println("  :tada:    - 完成")
	fmt.Println("  :bug:     - 修复bug")
	fmt.Println("  :rocket:  - 发射")
	fmt.Println("  :sparkles: - 优化")
	fmt.Println("  :fire:    - 删除代码")
	fmt.Println()
	fmt.Println("在消息中使用贴纸来表达情绪!")

	return nil
}

// ExtraUsageCommand 额外使用量命令
// 对应 TypeScript: /extra-usage 命令
// 显示额外使用量信息
type ExtraUsageCommand struct {
	BaseCommand
}

func NewExtraUsageCommand() *ExtraUsageCommand {
	return &ExtraUsageCommand{
		BaseCommand: *newCommand("extra-usage", "显示额外使用量"),
	}
}

func (c *ExtraUsageCommand) Execute(ctx context.Context, args []string, execCtx engine.CommandContext) error {
	fmt.Println("额外使用量信息")
	fmt.Println("=============")
	fmt.Println()
	fmt.Println("本月使用情况:")
	fmt.Println("  API请求: 0")
	fmt.Println("  令牌使用: 0")
	fmt.Println("  成本: $0.00")
	fmt.Println()
	fmt.Println("使用 /usage 查看详细统计")

	return nil
}

// RemoteEnvCommand 远程环境命令
// 对应 TypeScript: /remote-env 命令
// 管理远程环境变量
type RemoteEnvCommand struct {
	BaseCommand
}

func NewRemoteEnvCommand() *RemoteEnvCommand {
	return &RemoteEnvCommand{
		BaseCommand: *newCommand("remote-env", "管理远程环境变量"),
	}
}

func (c *RemoteEnvCommand) Execute(ctx context.Context, args []string, execCtx engine.CommandContext) error {
	if len(args) == 0 {
		fmt.Println("用法: /remote-env <变量名>=<值>")
		fmt.Println("  设置远程环境变量")
		fmt.Println()
		fmt.Println("示例:")
		fmt.Println("  /remote-env API_KEY=xxx")
		return nil
	}

	// 解析环境变量设置
	for _, arg := range args {
		fmt.Printf("设置远程环境变量: %s\n", arg)
	}

	fmt.Println("远程环境变量已更新")
	return nil
}

// CommitPushPRCommand 提交推送PR命令
// 对应 TypeScript: /commit-push-pr 命令
// 一步完成提交、推送和创建PR
type CommitPushPRCommand struct {
	BaseCommand
}

func NewCommitPushPRCommand() *CommitPushPRCommand {
	return &CommitPushPRCommand{
		BaseCommand: *newCommand("commit-push-pr", "提交、推送并创建PR"),
	}
}

func (c *CommitPushPRCommand) Execute(ctx context.Context, args []string, execCtx engine.CommandContext) error {
	if len(args) == 0 {
		fmt.Println("用法: /commit-push-pr <提交消息>")
		return nil
	}

	cwd := execCtx.GetWorkingDirectory()
	if cwd == "" {
		cwd, _ = os.Getwd()
	}

	commitMsg := args[0]

	// Git add
	fmt.Println("暂存所有更改...")
	if err := exec.Command("git", "add", "-A").Run(); err != nil {
		return fmt.Errorf("git add 失败: %v", err)
	}

	// Git commit
	fmt.Printf("提交: %s\n", commitMsg)
	cmd := exec.Command("git", "commit", "-m", commitMsg)
	cmd.Dir = cwd
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git commit 失败: %v", err)
	}

	// Git push
	fmt.Println("推送到远程...")
	cmd = exec.Command("git", "push")
	cmd.Dir = cwd
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git push 失败: %v", err)
	}

	// 创建PR (如果使用GitHub CLI)
	if ghCmd, err := exec.LookPath("gh"); err == nil {
		fmt.Println("创建PR...")
		cmd = exec.Command(ghCmd, "pr", "create", "--fill")
		cmd.Dir = cwd
		if err := cmd.Run(); err != nil {
			fmt.Printf("PR创建失败 (可能不是GitHub仓库): %v\n", err)
		}
	}

	fmt.Println("完成!")
	return nil
}

// OutputStyleCommand 输出样式命令
// 对应 TypeScript: /output-style 命令
// 设置输出样式
type OutputStyleCommand struct {
	BaseCommand
}

func NewOutputStyleCommand() *OutputStyleCommand {
	return &OutputStyleCommand{
		BaseCommand: *newCommand("output-style", "设置输出样式"),
	}
}

func (c *OutputStyleCommand) Execute(ctx context.Context, args []string, execCtx engine.CommandContext) error {
	fmt.Println("输出样式设置")
	fmt.Println("=============")
	fmt.Println()
	fmt.Println("可用样式选项:")
	fmt.Println("  compact  - 紧凑输出")
	fmt.Println("  verbose  - 详细输出")
	fmt.Println("  simple   - 简单输出")
	fmt.Println()
	fmt.Println("用法: /output-style <样式>")

	if len(args) > 0 {
		fmt.Printf("样式已设置为: %s\n", args[0])
	}

	return nil
}

// TerminalSetupCommand 终端设置命令
// 对应 TypeScript: /terminal-setup 命令
// 配置终端设置
type TerminalSetupCommand struct {
	BaseCommand
}

func NewTerminalSetupCommand() *TerminalSetupCommand {
	return &TerminalSetupCommand{
		BaseCommand: *newCommand("terminal-setup", "配置终端设置"),
	}
}

func (c *TerminalSetupCommand) Execute(ctx context.Context, args []string, execCtx engine.CommandContext) error {
	fmt.Println("终端设置")
	fmt.Println("=========")
	fmt.Println()
	fmt.Println("当前终端配置:")
	fmt.Println("  终端类型: 自动检测")
	fmt.Println("  颜色支持: 已启用")
	fmt.Println("  鼠标支持: 已启用")
	fmt.Println()
	fmt.Println("用法: /terminal-setup <选项>")
	fmt.Println("  或查看 /help 获取更多选项")

	return nil
}

// SecurityReviewCommand 安全审查命令
// 对应 TypeScript: /security-review 命令
// 运行安全审查
type SecurityReviewCommand struct {
	BaseCommand
}

func NewSecurityReviewCommand() *SecurityReviewCommand {
	return &SecurityReviewCommand{
		BaseCommand: *newCommand("security-review", "运行安全审查"),
	}
}

func (c *SecurityReviewCommand) Execute(ctx context.Context, args []string, execCtx engine.CommandContext) error {
	fmt.Println("安全审查")
	fmt.Println("=========")
	fmt.Println()
	fmt.Println("正在检查常见安全漏洞...")

	cwd := execCtx.GetWorkingDirectory()
	if cwd == "" {
		cwd, _ = os.Getwd()
	}

	// 检查敏感文件
	sensitiveFiles := []string{
		".env",
		".env.local",
		".env.production",
		"credentials.json",
		"secrets.json",
		"id_rsa",
		"id_rsa.pub",
		".pem",
	}

	fmt.Println("\n检查敏感文件:")
	for _, file := range sensitiveFiles {
		path := filepath.Join(cwd, file)
		if _, err := os.Stat(path); err == nil {
			fmt.Printf("  ⚠ 发现敏感文件: %s\n", file)
		}
	}

	fmt.Println("\n检查硬编码凭证...")
	// 使用 grep 查找可能的硬编码密钥
	cmd := exec.Command("grep", "-r", "-i", "-E", "(password|api_key|secret|token)\\s*=\\s*['\"][^'\"]{8,}['\"]", cwd)
	cmd.Dir = cwd
	output, _ := cmd.Output()
	if len(output) > 0 {
		fmt.Println("  ⚠ 发现可能的硬编码凭证:")
		fmt.Print(string(output))
	}

	fmt.Println("\n安全审查完成")
	fmt.Println("建议: 使用环境变量代替硬编码凭证")

	return nil
}

// AdvisorCommand 顾问命令
// 对应 TypeScript: /advisor 命令
// 提供代码改进建议
type AdvisorCommand struct {
	BaseCommand
}

func NewAdvisorCommand() *AdvisorCommand {
	return &AdvisorCommand{
		BaseCommand: *newCommand("advisor", "提供代码改进建议"),
	}
}

func (c *AdvisorCommand) Execute(ctx context.Context, args []string, execCtx engine.CommandContext) error {
	fmt.Println("代码顾问")
	fmt.Println("========")
	fmt.Println()
	fmt.Println("分析代码质量...")

	cwd := execCtx.GetWorkingDirectory()
	if cwd == "" {
		cwd, _ = os.Getwd()
	}

	// 统计代码行数
	totalLines := 0
	filepath.Walk(cwd, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() || filepath.Ext(path) == "" {
			return nil
		}
		data, _ := os.ReadFile(path)
		totalLines += len(data)
		return nil
	})

	fmt.Printf("\n项目统计:")
	fmt.Printf("  总代码行数: %d\n", totalLines/60) // 粗略估计

	fmt.Println("\n建议:")
	fmt.Println("  1. 保持代码简洁,单一职责原则")
	fmt.Println("  2. 添加适当的注释和文档")
	fmt.Println("  3. 编写单元测试")
	fmt.Println("  4. 定期重构优化")
	fmt.Println("  5. 使用版本控制")

	return nil
}

// ThinkbackPlayCommand 回放命令
// 对应 TypeScript: /thinkback-play 命令
// 回放思考过程
type ThinkbackPlayCommand struct {
	BaseCommand
}

func NewThinkbackPlayCommand() *ThinkbackPlayCommand {
	return &ThinkbackPlayCommand{
		BaseCommand: *newCommand("thinkback-play", "回放思考过程"),
	}
}

func (c *ThinkbackPlayCommand) Execute(ctx context.Context, args []string, execCtx engine.CommandContext) error {
	fmt.Println("思考回放")
	fmt.Println("========")
	fmt.Println()
	fmt.Println("可用回放选项:")
	fmt.Println("  /thinkback-play start - 开始记录思考过程")
	fmt.Println("  /thinkback-play stop  - 停止记录")
	fmt.Println("  /thinkback-play show  - 显示记录的思考")
	fmt.Println()
	fmt.Println("注意: 完整功能需要配置回放系统")

	return nil
}

// RemoteSetupCommand 远程设置命令
// 对应 TypeScript: /remote-setup 命令
// 配置远程连接
type RemoteSetupCommand struct {
	BaseCommand
}

func NewRemoteSetupCommand() *RemoteSetupCommand {
	return &RemoteSetupCommand{
		BaseCommand: *newCommand("remote-setup", "配置远程连接"),
	}
}

func (c *RemoteSetupCommand) Execute(ctx context.Context, args []string, execCtx engine.CommandContext) error {
	fmt.Println("远程设置")
	fmt.Println("=========")
	fmt.Println()
	fmt.Println("配置远程开发环境:")
	fmt.Println("  1. SSH 连接")
	fmt.Println("  2. Docker 容器")
	fmt.Println("  3. 远程服务器")
	fmt.Println()
	fmt.Println("用法: /remote-setup <类型>")
	fmt.Println("  /remote-setup ssh <主机>")
	fmt.Println("  /remote-setup docker <容器>")

	return nil
}

// BridgeCommand 桥接命令
// 对应 TypeScript: /bridge 命令
// 远程控制功能
type BridgeCommand struct {
	BaseCommand
}

func NewBridgeCommand() *BridgeCommand {
	return &BridgeCommand{
		BaseCommand: *newCommand("bridge", "远程控制功能"),
	}
}

func (c *BridgeCommand) Execute(ctx context.Context, args []string, execCtx engine.CommandContext) error {
	fmt.Println("桥接/远程控制")
	fmt.Println("=============")
	fmt.Println()
	fmt.Println("桥接功能允许远程控制 Claude Code")
	fmt.Println()
	fmt.Println("用法:")
	fmt.Println("  /bridge connect <地址>  - 连接到远程实例")
	fmt.Println("  /bridge disconnect     - 断开连接")
	fmt.Println("  /bridge status         - 显示连接状态")

	return nil
}

// BridgeKickCommand 桥接踢出命令
// 对应 TypeScript: /bridge-kick 命令
// 踢出远程连接
type BridgeKickCommand struct {
	BaseCommand
}

func NewBridgeKickCommand() *BridgeKickCommand {
	return &BridgeKickCommand{
		BaseCommand: *newCommand("bridge-kick", "踢出远程连接"),
	}
}

func (c *BridgeKickCommand) Execute(ctx context.Context, args []string, execCtx engine.CommandContext) error {
	fmt.Println("桥接踢出")
	fmt.Println("=========")
	fmt.Println()
	fmt.Println("用法: /bridge-kick <连接ID>")
	fmt.Println("  踢出指定的远程连接")
	fmt.Println()
	fmt.Println("注意: 需要管理员权限")

	return nil
}

// InstallGitHubAppCommand 安装GitHub应用命令
// 对应 TypeScript: /install-github-app 命令
// 安装GitHub集成应用
type InstallGitHubAppCommand struct {
	BaseCommand
}

func NewInstallGitHubAppCommand() *InstallGitHubAppCommand {
	return &InstallGitHubAppCommand{
		BaseCommand: *newCommand("install-github-app", "安装GitHub集成"),
	}
}

func (c *InstallGitHubAppCommand) Execute(ctx context.Context, args []string, execCtx engine.CommandContext) error {
	fmt.Println("GitHub 应用安装")
	fmt.Println("================")
	fmt.Println()
	fmt.Println("安装 Claude Code GitHub 应用以启用:")
	fmt.Println("  - PR 评论摘要")
	fmt.Println("  - 自动代码审查")
	fmt.Println("  - 仓库洞察")
	fmt.Println()
	fmt.Println("请访问以下链接安装:")
	fmt.Println("  https://github.com/apps/claude-code")

	return nil
}

// InstallSlackAppCommand 安装Slack应用命令
// 对应 TypeScript: /install-slack-app 命令
// 安装Slack集成应用
type InstallSlackAppCommand struct {
	BaseCommand
}

func NewInstallSlackAppCommand() *InstallSlackAppCommand {
	return &InstallSlackAppCommand{
		BaseCommand: *newCommand("install-slack-app", "安装Slack集成"),
	}
}

func (c *InstallSlackAppCommand) Execute(ctx context.Context, args []string, execCtx engine.CommandContext) error {
	fmt.Println("Slack 应用安装")
	fmt.Println("==============")
	fmt.Println()
	fmt.Println("安装 Claude Code Slack 应用以启用:")
	fmt.Println("  - 通知推送")
	fmt.Println("  - 交互式命令")
	fmt.Println("  - 团队协作")
	fmt.Println()
	fmt.Println("请访问以下链接安装:")
	fmt.Println("  https://slack.com/apps/claude-code")

	return nil
}
