package styles

import (
	"github.com/charmbracelet/lipgloss"
)

// UI样式定义 - Claude Code Go 版本的样式系统
// 参考 TypeScript 版本的 design-system

// 颜色定义
var (
	// 主色调
	PrimaryColor   = lipgloss.Color("12") // 蓝色
	SecondaryColor = lipgloss.Color("14") // 绿色
	AccentColor    = lipgloss.Color("13") // 洋红
	WarningColor   = lipgloss.Color("11") // 黄色
	ErrorColor     = lipgloss.Color("9")  // 红色
	SuccessColor   = lipgloss.Color("10") // 绿色

	// 文本颜色
	TextColor  = lipgloss.Color("15") // 白色
	MutedColor = lipgloss.Color("8")  // 灰色
	DimColor   = lipgloss.Color("7")  // 暗灰色

	// 边框颜色
	BorderColor      = lipgloss.Color("8")
	FocusBorderColor = lipgloss.Color("12")

	// 背景色
	BackgroundColor = lipgloss.Color("0")
	SurfaceColor    = lipgloss.Color("0")
)

// 通用样式
var (
	// 粗体文本
	Bold = lipgloss.NewStyle().Bold(true)

	// 斜体文本
	Italic = lipgloss.NewStyle().Italic(true)

	// 高亮文本
	Highlight = lipgloss.NewStyle().Foreground(PrimaryColor)

	// 成功样式
	Success = lipgloss.NewStyle().Foreground(SuccessColor)

	// 警告样式
	Warning = lipgloss.NewStyle().Foreground(WarningColor)

	// 错误样式
	Error = lipgloss.NewStyle().Foreground(ErrorColor)

	// 暗淡文本
	Dim = lipgloss.NewStyle().Foreground(DimColor)

	// 超链接样式
	Link = lipgloss.NewStyle().Foreground(AccentColor).Underline(true)
)

// PaneStyle 面板样式
var PaneStyle = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(BorderColor).
	Padding(1, 2)

// DialogStyle 对话框样式
var DialogStyle = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(FocusBorderColor).
	Background(SurfaceColor).
	Padding(1, 2).
	Width(60).
	Align(lipgloss.Center)

// SelectOptionStyle 选择项样式
var SelectOptionStyle = lipgloss.NewStyle().
	Foreground(TextColor).
	Padding(0, 1)

var SelectOptionSelectedStyle = lipgloss.NewStyle().
	Foreground(PrimaryColor).
	Bold(true).
	Background(lipgloss.Color("8"))

// SpinnerStyle 加载动画样式
var SpinnerStyle = lipgloss.NewStyle().
	Foreground(PrimaryColor)

// 树形连接符
var TreeConnector = lipgloss.NewStyle().
	Foreground(DimColor).
	Width(2)

var TreeConnectorChar = "⎿ "

// 帮助文本样式
var HelpStyle = lipgloss.NewStyle().
	Foreground(MutedColor).
	Italic(true)

// 按键样式
var KeyStyle = lipgloss.NewStyle().
	Foreground(TextColor).
	Background(lipgloss.Color("8")).
	Padding(0, 1).
	MarginRight(1)

// 标题样式
var TitleStyle = lipgloss.NewStyle().
	Foreground(TextColor).
	Bold(true)

// 副标题样式
var SubtitleStyle = lipgloss.NewStyle().
	Foreground(MutedColor)

// GetWidth 获取终端宽度
func GetWidth() int {
	return 80
}

// GetHeight 获取终端高度
func GetHeight() int {
	return 24
}

// CenterText 居中文本
func CenterText(text string, width int) string {
	textWidth := lipgloss.Width(text)
	if textWidth >= width {
		return text
	}
	padding := (width - textWidth) / 2
	return lipgloss.NewStyle().MarginLeft(padding).Render(text)
}
