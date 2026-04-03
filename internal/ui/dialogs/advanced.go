package dialogs

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/claude-code-go/claude/internal/ui/styles"
)

// TeammateInfo 队友信息
type TeammateInfo struct {
	Name      string
	Status    string // "idle", "running", "error"
	Mode      string // "auto", "manual"
	Model     string
	Path      string
	StartedAt time.Time
}

// TeamsDialog 团队对话框
// 参考 TypeScript: src/components/teams/TeamsDialog.tsx
type TeamsDialog struct {
	*BaseDialog
	Teammates   []TeammateInfo
	Cursor      int
	SelectedIdx int
	ViewMode    string // "list", "detail"
	RefreshRate time.Duration
	lastRefresh time.Time
}

// NewTeamsDialog 创建团队对话框
func NewTeamsDialog() *TeamsDialog {
	return &TeamsDialog{
		BaseDialog:  NewBaseDialog("Team"),
		Teammates:   make([]TeammateInfo, 0),
		Cursor:      0,
		SelectedIdx: -1,
		ViewMode:    "list",
		RefreshRate: time.Second,
	}
}

// AddTeammate 添加队友
func (d *TeamsDialog) AddTeammate(info TeammateInfo) {
	d.Teammates = append(d.Teammates, info)
}

// RemoveTeammate 移除队友
func (d *TeamsDialog) RemoveTeammate(index int) {
	if index >= 0 && index < len(d.Teammates) {
		d.Teammates = append(d.Teammates[:index], d.Teammates[index+1:]...)
	}
}

// MoveUp 上移
func (d *TeamsDialog) MoveUp() {
	if d.Cursor > 0 {
		d.Cursor--
	}
}

// MoveDown 下移
func (d *TeamsDialog) MoveDown() {
	if d.Cursor < len(d.Teammates)-1 {
		d.Cursor++
	}
}

// Select 选择当前队友
func (d *TeamsDialog) Select() {
	if d.Cursor >= 0 && d.Cursor < len(d.Teammates) {
		d.SelectedIdx = d.Cursor
		d.ViewMode = "detail"
	}
}

// Back 返回列表
func (d *TeamsDialog) Back() {
	d.ViewMode = "list"
	d.SelectedIdx = -1
}

// ShouldRefresh 检查是否需要刷新
func (d *TeamsDialog) ShouldRefresh() bool {
	if time.Since(d.lastRefresh) >= d.RefreshRate {
		d.lastRefresh = time.Now()
		return true
	}
	return false
}

// Render 渲染
func (d *TeamsDialog) Render() string {
	if d.ViewMode == "detail" {
		return d.renderDetail()
	}
	return d.renderList()
}

// renderList 渲染列表视图
func (d *TeamsDialog) renderList() string {
	var lines []string

	// 标题
	lines = append(lines, styles.TitleStyle.Render("👥 Team"))
	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("  %d teammate(s)", len(d.Teammates)))
	lines = append(lines, "")

	if len(d.Teammates) == 0 {
		lines = append(lines, styles.Dim.Render("  No teammates yet"))
		lines = append(lines, "")
		lines = append(lines, styles.HelpStyle.Render("  k: kill | s: shutdown | p: prune idle | h/H: hide/show"))
		return d.wrapContent(strings.Join(lines, "\n"))
	}

	// 队友列表
	cursorStyle := lipgloss.NewStyle().Foreground(styles.PrimaryColor)

	for i, tm := range d.Teammates {
		cursor := "  "
		if i == d.Cursor {
			cursor = cursorStyle.Render("▶ ")
		}

		// 状态颜色
		statusStr := tm.Status
		switch tm.Status {
		case "running":
			statusStr = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Render("● " + tm.Status)
		case "idle":
			statusStr = lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render("○ " + tm.Status)
		case "error":
			statusStr = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render("✗ " + tm.Status)
		}

		nameStyle := lipgloss.NewStyle()
		if i == d.Cursor {
			nameStyle = styles.Bold
		}

		line := fmt.Sprintf("%s%s %s %s", cursor, statusStr, nameStyle.Render(tm.Name), styles.Dim.Render("("+tm.Mode+")"))
		lines = append(lines, line)
	}

	lines = append(lines, "")

	// 帮助
	helpItems := []string{
		"↑↓: navigate",
		"Enter: view details",
		"k: kill",
		"s: shutdown",
		"p: prune idle",
		"h/H: hide/show",
		"Esc: back",
	}
	lines = append(lines, styles.HelpStyle.Render("  "+strings.Join(helpItems, " | ")))

	return d.wrapContent(strings.Join(lines, "\n"))
}

// renderDetail 渲染详情视图
func (d *TeamsDialog) renderDetail() string {
	if d.SelectedIdx < 0 || d.SelectedIdx >= len(d.Teammates) {
		return d.renderList()
	}

	tm := d.Teammates[d.SelectedIdx]
	var lines []string

	// 标题
	lines = append(lines, styles.TitleStyle.Render(fmt.Sprintf("👥 %s", tm.Name)))
	lines = append(lines, "")

	// 详情
	lines = append(lines, fmt.Sprintf("  %s: %s", styles.Dim.Render("Status"), tm.Status))
	lines = append(lines, fmt.Sprintf("  %s: %s", styles.Dim.Render("Mode"), tm.Mode))
	lines = append(lines, fmt.Sprintf("  %s: %s", styles.Dim.Render("Model"), tm.Model))
	lines = append(lines, fmt.Sprintf("  %s: %s", styles.Dim.Render("Path"), tm.Path))
	lines = append(lines, fmt.Sprintf("  %s: %s", styles.Dim.Render("Started"), tm.StartedAt.Format("15:04:05")))

	elapsed := time.Since(tm.StartedAt)
	lines = append(lines, fmt.Sprintf("  %s: %s", styles.Dim.Render("Duration"), formatDuration(elapsed)))

	lines = append(lines, "")

	// 操作
	lines = append(lines, "  Actions:")
	lines = append(lines, fmt.Sprintf("    %s. %s", lipgloss.NewStyle().Foreground(styles.PrimaryColor).Render("k"), "Kill this teammate"))
	lines = append(lines, fmt.Sprintf("    %s. %s", lipgloss.NewStyle().Foreground(styles.WarningColor).Render("s"), "Shutdown this teammate"))

	lines = append(lines, "")
	lines = append(lines, styles.HelpStyle.Render("  Esc: back to list"))

	return d.wrapContent(strings.Join(lines, "\n"))
}

// wrapContent 包装内容
func (d *TeamsDialog) wrapContent(content string) string {
	dialogStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.BorderColor).
		Padding(1, 2).
		Width(d.Width)

	return dialogStyle.Render(content)
}

// formatDuration 格式化时长
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm %ds", int(d.Minutes()), int(d.Seconds())%60)
	}
	return fmt.Sprintf("%dh %dm", int(d.Hours()), int(d.Minutes())%60)
}

// OnboardingDialog 初始化向导对话框
// 参考 TypeScript: src/components/Onboarding.tsx
type OnboardingDialog struct {
	*BaseDialog
	Steps       []string
	CurrentStep int
	Theme       string
	APIKey      string
}

// NewOnboardingDialog 创建初始化向导
func NewOnboardingDialog() *OnboardingDialog {
	steps := []string{
		"preflight",
		"theme",
		"api-key",
		"oauth",
		"security",
		"terminal",
	}
	return &OnboardingDialog{
		BaseDialog:  NewBaseDialog("Welcome to Claude Code"),
		Steps:       steps,
		CurrentStep: 0,
		Theme:       "dark",
		APIKey:      "",
	}
}

// NextStep 下一步
func (d *OnboardingDialog) NextStep() {
	if d.CurrentStep < len(d.Steps)-1 {
		d.CurrentStep++
	}
}

// PrevStep 上一步
func (d *OnboardingDialog) PrevStep() {
	if d.CurrentStep > 0 {
		d.CurrentStep--
	}
}

// SetTheme 设置主题
func (d *OnboardingDialog) SetTheme(theme string) {
	d.Theme = theme
}

// SetAPIKey 设置API密钥
func (d *OnboardingDialog) SetAPIKey(key string) {
	d.APIKey = key
}

// IsComplete 检查是否完成
func (d *OnboardingDialog) IsComplete() bool {
	return d.CurrentStep >= len(d.Steps)-1
}

// Render 渲染
func (d *OnboardingDialog) Render() string {
	step := d.Steps[d.CurrentStep]

	var content string
	switch step {
	case "preflight":
		content = d.renderPreflight()
	case "theme":
		content = d.renderTheme()
	case "api-key":
		content = d.renderAPIKey()
	case "security":
		content = d.renderSecurity()
	case "terminal":
		content = d.renderTerminal()
	default:
		content = d.renderComplete()
	}

	return d.wrapContent(content)
}

// renderPreflight 渲染预检步骤
func (d *OnboardingDialog) renderPreflight() string {
	var lines []string
	lines = append(lines, styles.TitleStyle.Render("🚀 Getting Started"))
	lines = append(lines, "")
	lines = append(lines, "  Let's set up Claude Code for you!")
	lines = append(lines, "")
	lines = append(lines, "  This wizard will help you configure:")
	lines = append(lines, "  • Theme preferences")
	lines = append(lines, "  • API key configuration")
	lines = append(lines, "  • Terminal settings")
	lines = append(lines, "")
	lines = append(lines, styles.HelpStyle.Render("  Press Enter to continue..."))
	return strings.Join(lines, "\n")
}

// renderTheme 渲染主题选择步骤
func (d *OnboardingDialog) renderTheme() string {
	var lines []string
	lines = append(lines, styles.TitleStyle.Render("🎨 Theme"))
	lines = append(lines, "")
	lines = append(lines, "  Choose your preferred theme:")
	lines = append(lines, "")

	themes := []string{"dark", "light", "system"}
	cursorStyle := lipgloss.NewStyle().Foreground(styles.PrimaryColor)

	for _, theme := range themes {
		cursor := "  "
		if theme == d.Theme {
			cursor = cursorStyle.Render("▶ ")
		}
		lines = append(lines, fmt.Sprintf("%s%s %s", cursor, theme, styles.Dim.Render("(recommended)")))
	}

	lines = append(lines, "")
	lines = append(lines, styles.HelpStyle.Render("  ↑↓: select | Enter: confirm"))
	return strings.Join(lines, "\n")
}

// renderAPIKey 渲染API密钥步骤
func (d *OnboardingDialog) renderAPIKey() string {
	var lines []string
	lines = append(lines, styles.TitleStyle.Render("🔑 API Key"))
	lines = append(lines, "")
	lines = append(lines, "  Enter your Anthropic API key:")
	lines = append(lines, "")

	if d.APIKey != "" {
		masked := strings.Repeat("*", len(d.APIKey)-4) + d.APIKey[len(d.APIKey)-4:]
		lines = append(lines, fmt.Sprintf("  Current: %s", masked))
	} else {
		lines = append(lines, "  No API key set")
	}

	lines = append(lines, "")
	lines = append(lines, styles.HelpStyle.Render("  Enter API key or press Esc to skip"))
	return strings.Join(lines, "\n")
}

// renderSecurity 渲染安全步骤
func (d *OnboardingDialog) renderSecurity() string {
	var lines []string
	lines = append(lines, styles.TitleStyle.Render("🔒 Security Notes"))
	lines = append(lines, "")
	lines = append(lines, "  Important security reminders:")
	lines = append(lines, "")
	lines = append(lines, "  1. Never share your API key with others")
	lines = append(lines, "  2. Review /health before using in production")
	lines = append(lines, "  3. Claude Code may execute code on your behalf")
	lines = append(lines, "  4. Keep your system updated")
	lines = append(lines, "")
	lines = append(lines, styles.HelpStyle.Render("  Press Enter to continue..."))
	return strings.Join(lines, "\n")
}

// renderTerminal 渲染终端设置步骤
func (d *OnboardingDialog) renderTerminal() string {
	var lines []string
	lines = append(lines, styles.TitleStyle.Render("⌨️  Terminal Setup"))
	lines = append(lines, "")
	lines = append(lines, "  Configure terminal integration:")
	lines = append(lines, "")
	lines = append(lines, "  • Enable mouse support for better UX")
	lines = append(lines, "  • Configure shell integration")
	lines = append(lines, "")
	lines = append(lines, styles.Dim.Render("  These features are optional but recommended"))
	lines = append(lines, "")
	lines = append(lines, styles.HelpStyle.Render("  Enter: install | Esc: skip"))
	return strings.Join(lines, "\n")
}

// renderComplete 渲染完成
func (d *OnboardingDialog) renderComplete() string {
	var lines []string
	lines = append(lines, styles.TitleStyle.Render("✅ Setup Complete!"))
	lines = append(lines, "")
	lines = append(lines, "  You're ready to use Claude Code!")
	lines = append(lines, "")
	lines = append(lines, "  Quick tips:")
	lines = append(lines, "  • Type /help for available commands")
	lines = append(lines, "  • Use /doctor to check your setup")
	lines = append(lines, "  • Configure with /config")
	lines = append(lines, "")
	lines = append(lines, styles.HelpStyle.Render("  Press Enter to start..."))
	return strings.Join(lines, "\n")
}

// wrapContent 包装内容
func (d *OnboardingDialog) wrapContent(content string) string {
	dialogStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.PrimaryColor).
		Padding(1, 2).
		Width(d.Width)

	return dialogStyle.Render(content)
}
