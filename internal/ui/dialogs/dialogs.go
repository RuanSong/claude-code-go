package dialogs

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/claude-code-go/claude/internal/ui/components"
	"github.com/claude-code-go/claude/internal/ui/styles"
)

// BaseDialog 基础对话框
// 参考 TypeScript: src/components/design-system/Dialog.tsx
type BaseDialog struct {
	Title       string
	Description string
	Width       int
	BorderColor lipgloss.TerminalColor
}

// NewBaseDialog 创建基础对话框
func NewBaseDialog(title string) *BaseDialog {
	return &BaseDialog{
		Title:       title,
		Width:       60,
		BorderColor: styles.BorderColor,
	}
}

// SetDescription 设置描述
func (d *BaseDialog) SetDescription(desc string) *BaseDialog {
	d.Description = desc
	return d
}

// SetWidth 设置宽度
func (d *BaseDialog) SetWidth(width int) *BaseDialog {
	d.Width = width
	return d
}

// SetBorderColor 设置边框颜色
func (d *BaseDialog) SetBorderColor(color lipgloss.TerminalColor) *BaseDialog {
	d.BorderColor = color
	return d
}

// Render 渲染对话框
func (d *BaseDialog) Render() string {
	dialogStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(d.BorderColor).
		Padding(1, 2).
		Width(d.Width)

	var lines []string

	// 标题
	if d.Title != "" {
		lines = append(lines, styles.TitleStyle.Render(d.Title))
	}

	// 描述
	if d.Description != "" {
		if d.Title != "" {
			lines = append(lines, "")
		}
		lines = append(lines, styles.SubtitleStyle.Render(d.Description))
	}

	content := strings.Join(lines, "\n")
	return dialogStyle.Render(content)
}

// ConfirmationDialog 确认对话框
// 参考 TypeScript: src/components/ConfirmDialog.tsx
type ConfirmationDialog struct {
	*BaseDialog
	Options     []*components.SelectOption
	CancelText  string
	ConfirmText string
}

// NewConfirmationDialog 创建确认对话框
func NewConfirmationDialog(title, description string) *ConfirmationDialog {
	return &ConfirmationDialog{
		BaseDialog:  NewBaseDialog(title),
		Options:     make([]*components.SelectOption, 0),
		CancelText:  "取消",
		ConfirmText: "确认",
	}
}

// AddOption 添加选项
func (d *ConfirmationDialog) AddOption(label, value string) *ConfirmationDialog {
	d.Options = append(d.Options, components.NewSelectOption(label, value))
	return d
}

// SetCancelText 设置取消文本
func (d *ConfirmationDialog) SetCancelText(text string) *ConfirmationDialog {
	d.CancelText = text
	return d
}

// SetConfirmText 设置确认文本
func (d *ConfirmationDialog) SetConfirmText(text string) *ConfirmationDialog {
	d.ConfirmText = text
	return d
}

// Render 渲染确认对话框
func (d *ConfirmationDialog) Render() string {
	var lines []string

	// 标题
	if d.Title != "" {
		lines = append(lines, styles.TitleStyle.Render(d.Title))
		lines = append(lines, "")
	}

	// 描述
	if d.Description != "" {
		lines = append(lines, d.Description)
		lines = append(lines, "")
	}

	// 选项
	for i, opt := range d.Options {
		key := ""
		if i == 0 {
			key = "Enter"
		}
		opt := opt.WithKey(key)
		lines = append(lines, fmt.Sprintf("  %s. %s", lipgloss.NewStyle().Foreground(styles.PrimaryColor).Render(fmt.Sprintf("%d", i+1)), opt.Label))
	}

	// 取消选项
	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("  %s. %s", lipgloss.NewStyle().Foreground(styles.MutedColor).Render("Esc"), d.CancelText))

	// 帮助文本
	lines = append(lines, "")
	lines = append(lines, styles.HelpStyle.Render("Enter: 确认 | Esc: 取消"))

	content := strings.Join(lines, "\n")

	// 使用对话框样式包装
	dialogStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(d.BorderColor).
		Padding(1, 2).
		Width(d.Width)

	return dialogStyle.Render(content)
}

// MCPServerApprovalDialog MCP服务器审批对话框
// 参考 TypeScript: src/components/MCPServerApprovalDialog.tsx
type MCPServerApprovalDialog struct {
	*BaseDialog
	ServerName string
	IsCustom   bool
}

// NewMCPServerApprovalDialog 创建MCP服务器审批对话框
func NewMCPServerApprovalDialog(serverName string) *MCPServerApprovalDialog {
	return &MCPServerApprovalDialog{
		BaseDialog: NewBaseDialog("MCP Server Approval"),
		ServerName: serverName,
		IsCustom:   false,
	}
}

// SetCustom 设置是否为自定义服务器
func (d *MCPServerApprovalDialog) SetCustom(isCustom bool) *MCPServerApprovalDialog {
	d.IsCustom = isCustom
	return d
}

// Render 渲染MCP审批对话框
func (d *MCPServerApprovalDialog) Render() string {
	var lines []string

	// 标题
	lines = append(lines, styles.TitleStyle.Render("⚠️  MCP Server Approval Required"))
	lines = append(lines, "")
	lines = append(lines, "A new MCP server has been configured:")
	lines = append(lines, "")
	lines = append(lines, styles.Bold.Render("  Server: ")+d.ServerName)
	if d.IsCustom {
		lines = append(lines, styles.Warning.Render("  ⚠️  Custom server - verify the URL is correct"))
	}
	lines = append(lines, "")

	// 选项
	lines = append(lines, "How would you like to proceed?")
	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("  %s. %s", lipgloss.NewStyle().Foreground(styles.PrimaryColor).Render("1"), "Approve and use this server"))
	lines = append(lines, fmt.Sprintf("  %s. %s", lipgloss.NewStyle().Foreground(styles.PrimaryColor).Render("2"), "Approve and allow all future MCP servers"))
	lines = append(lines, fmt.Sprintf("  %s. %s", lipgloss.NewStyle().Foreground(styles.MutedColor).Render("3"), "Skip this server"))
	lines = append(lines, "")

	// 帮助
	lines = append(lines, styles.HelpStyle.Render("1/2/3: 选择 | Enter: 确认 | Esc: 取消"))

	content := strings.Join(lines, "\n")

	dialogStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("11")). // 警告黄色
		Padding(1, 2).
		Width(d.Width)

	return dialogStyle.Render(content)
}

// MCPServerMultiselectDialog MCP服务器多选对话框
// 参考 TypeScript: src/components/MCPServerMultiselectDialog.tsx
type MCPServerMultiselectDialog struct {
	*BaseDialog
	Servers  []string
	Approved []string
	Rejected []string
	Checked  map[int]bool
	Cursor   int
}

// NewMCPServerMultiselectDialog 创建MCP多选对话框
func NewMCPServerMultiselectDialog(servers []string) *MCPServerMultiselectDialog {
	checked := make(map[int]bool)
	for i := range servers {
		checked[i] = true // 默认全部选中
	}
	return &MCPServerMultiselectDialog{
		BaseDialog: NewBaseDialog("MCP Server Approval"),
		Servers:    servers,
		Approved:   make([]string, 0),
		Rejected:   make([]string, 0),
		Checked:    checked,
		Cursor:     0,
	}
}

// MoveUp 上移
func (d *MCPServerMultiselectDialog) MoveUp() {
	if d.Cursor > 0 {
		d.Cursor--
	}
}

// MoveDown 下移
func (d *MCPServerMultiselectDialog) MoveDown() {
	if d.Cursor < len(d.Servers)-1 {
		d.Cursor++
	}
}

// Toggle 切换选中状态
func (d *MCPServerMultiselectDialog) Toggle() {
	d.Checked[d.Cursor] = !d.Checked[d.Cursor]
}

// SelectAll 全选
func (d *MCPServerMultiselectDialog) SelectAll() {
	for i := range d.Servers {
		d.Checked[i] = true
	}
}

// SelectNone 取消全选
func (d *MCPServerMultiselectDialog) SelectNone() {
	for i := range d.Servers {
		d.Checked[i] = false
	}
}

// Confirm 确认选择
func (d *MCPServerMultiselectDialog) Confirm() {
	d.Approved = make([]string, 0)
	d.Rejected = make([]string, 0)
	for i, server := range d.Servers {
		if d.Checked[i] {
			d.Approved = append(d.Approved, server)
		} else {
			d.Rejected = append(d.Rejected, server)
		}
	}
}

// Render 渲染多选对话框
func (d *MCPServerMultiselectDialog) Render() string {
	var lines []string

	// 标题
	lines = append(lines, styles.TitleStyle.Render("MCP Server Approval"))
	lines = append(lines, "")
	lines = append(lines, "Select which MCP servers to approve:")
	lines = append(lines, "")

	// 服务器列表
	successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	cursorStyle := lipgloss.NewStyle().Foreground(styles.PrimaryColor)

	for i, server := range d.Servers {
		cursor := "  "
		if i == d.Cursor {
			cursor = cursorStyle.Render("▶ ")
		}

		checked := d.Checked[i]
		checkbox := "[ ]"
		if checked {
			checkbox = successStyle.Render("[✓]")
		}

		serverText := server
		if i == d.Cursor {
			serverText = styles.Bold.Render(server)
		}

		lines = append(lines, fmt.Sprintf("%s%s %s", cursor, checkbox, serverText))
	}

	lines = append(lines, "")

	// 帮助
	lines = append(lines, styles.HelpStyle.Render("Space: 选择 | a: 全选 | n: 取消全选 | Enter: 确认 | Esc: 取消"))

	content := strings.Join(lines, "\n")

	dialogStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.BorderColor).
		Padding(1, 2).
		Width(d.Width)

	return dialogStyle.Render(content)
}

// TextInputDialog 文本输入对话框
// 参考 TypeScript: src/components/TextInput.tsx
type TextInputDialog struct {
	*BaseDialog
	Placeholder  string
	DefaultValue string
	Value        string
	CursorPos    int
	MaxLength    int
}

// NewTextInputDialog 创建文本输入对话框
func NewTextInputDialog(title, placeholder string) *TextInputDialog {
	return &TextInputDialog{
		BaseDialog:  NewBaseDialog(title),
		Placeholder: placeholder,
		Value:       "",
		CursorPos:   0,
		MaxLength:   0,
	}
}

// SetDefaultValue 设置默认值
func (d *TextInputDialog) SetDefaultValue(value string) *TextInputDialog {
	d.DefaultValue = value
	d.Value = value
	d.CursorPos = len(value)
	return d
}

// SetMaxLength 设置最大长度
func (d *TextInputDialog) SetMaxLength(max int) *TextInputDialog {
	d.MaxLength = max
	return d
}

// Append 添加字符
func (d *TextInputDialog) Append(ch string) {
	if d.MaxLength > 0 && len(d.Value) >= d.MaxLength {
		return
	}
	before := d.Value[:d.CursorPos]
	after := d.Value[d.CursorPos:]
	d.Value = before + ch + after
	d.CursorPos += len(ch)
}

// Delete 删除字符
func (d *TextInputDialog) Delete() {
	if d.CursorPos == 0 {
		return
	}
	before := d.Value[:d.CursorPos-1]
	after := d.Value[d.CursorPos:]
	d.Value = before + after
	d.CursorPos--
}

// Backspace 退格
func (d *TextInputDialog) Backspace() {
	d.Delete()
}

// MoveLeft 左移光标
func (d *TextInputDialog) MoveLeft() {
	if d.CursorPos > 0 {
		d.CursorPos--
	}
}

// MoveRight 右移光标
func (d *TextInputDialog) MoveRight() {
	if d.CursorPos < len(d.Value) {
		d.CursorPos++
	}
}

// MoveToStart 移到开始
func (d *TextInputDialog) MoveToStart() {
	d.CursorPos = 0
}

// MoveToEnd 移到结尾
func (d *TextInputDialog) MoveToEnd() {
	d.CursorPos = len(d.Value)
}

// Clear 清除
func (d *TextInputDialog) Clear() {
	d.Value = ""
	d.CursorPos = 0
}

// Render 渲染文本输入对话框
func (d *TextInputDialog) Render() string {
	var lines []string

	// 标题
	if d.Title != "" {
		lines = append(lines, styles.TitleStyle.Render(d.Title))
		lines = append(lines, "")
	}

	// 输入框
	inputStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.FocusBorderColor).
		Padding(0, 1)

	// 显示文本 + 光标
	displayText := d.Value
	if displayText == "" {
		displayText = d.Placeholder
	}

	// 构建带光标的文本
	before := displayText[:d.CursorPos]
	after := displayText[d.CursorPos:]
	cursor := lipgloss.NewStyle().
		Background(styles.PrimaryColor).
		Foreground(lipgloss.Color("0")).
		Render(" ")

	if d.Value == "" {
		before = styles.Dim.Render(d.Placeholder)
	}

	inputLine := before + cursor + after
	lines = append(lines, inputStyle.Render(inputLine))
	lines = append(lines, "")

	// 帮助
	helpText := "←→: 移动 | Ctrl+A: 开始 | Ctrl+E: 结尾 | Enter: 确认 | Esc: 取消"
	lines = append(lines, styles.HelpStyle.Render(helpText))

	return strings.Join(lines, "\n")
}
