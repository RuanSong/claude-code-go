package tui

// TUI包 - 使用 Bubble Tea 框架实现交互式终端用户界面
// 参考 TypeScript: src/ink.tsx, src/components/

import (
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/claude-code-go/claude/internal/ui/components"
	"github.com/claude-code-go/claude/internal/ui/dialogs"
	"github.com/claude-code-go/claude/internal/ui/styles"
)

// TUI模式 - 决定当前显示哪个视图
type TUIMode int

const (
	ModeInput      TUIMode = iota // 默认输入模式
	ModeSpinner                   // 加载动画模式
	ModeSelect                    // 选择器模式
	ModeDialog                    // 对话框模式
	ModeTeams                     // 团队管理模式
	ModeOnboarding                // 初始化向导模式
)

// MainModel 主TUI模型
// 使用 Bubble Tea 架构实现交互式界面
type MainModel struct {
	Mode        TUIMode
	InputBuffer string
	CursorPos   int
	History     []string
	HistoryIdx  int

	// 组件
	Spinner     *components.SpinnerWithVerb
	SelectModel *components.SelectModel
	Dialog      interface{}

	// 消息历史
	Messages []MessageItem
	Width    int
	Height   int

	// 状态
	IsThinking bool
	ShowHelp   bool
}

// MessageItem 消息项
type MessageItem struct {
	Role    string // "user", "assistant", "system"
	Content string
}

// NewMainModel 创建主模型
func NewMainModel() *MainModel {
	return &MainModel{
		Mode:        ModeInput,
		InputBuffer: "",
		CursorPos:   0,
		History:     make([]string, 0),
		HistoryIdx:  -1,
		Messages:    make([]MessageItem, 0),
		Spinner:     components.NewSpinnerWithVerb("Thinking..."),
		Width:       80,
		Height:      24,
		IsThinking:  false,
		ShowHelp:    true,
	}
}

// Init 初始化
func (m *MainModel) Init() tea.Cmd {
	return nil
}

// Update 更新模型
func (m *MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		return m, nil

	default:
		return m, nil
	}
}

// handleKey 处理按键
func (m *MainModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.Mode {
	case ModeInput:
		return m.handleInputKey(msg)

	case ModeSpinner:
		if msg.Type == tea.KeyCtrlC {
			m.IsThinking = false
			m.Mode = ModeInput
		}
		return m, nil

	case ModeSelect:
		return m.handleSelectKey(msg)

	case ModeDialog:
		return m.handleDialogKey(msg)

	case ModeOnboarding:
		return m.handleOnboardingKey(msg)

	default:
		return m, nil
	}
}

// handleInputKey 处理输入模式按键
func (m *MainModel) handleInputKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEnter:
		if m.InputBuffer != "" {
			// 添加到历史
			m.History = append(m.History, m.InputBuffer)
			m.HistoryIdx = len(m.History)
			return m, nil
		}

	case tea.KeyRunes:
		m.InputBuffer = m.InputBuffer[:m.CursorPos] + string(msg.Runes) + m.InputBuffer[m.CursorPos:]
		m.CursorPos++

	case tea.KeyBackspace:
		if m.CursorPos > 0 {
			m.InputBuffer = m.InputBuffer[:m.CursorPos-1] + m.InputBuffer[m.CursorPos:]
			m.CursorPos--
		}

	case tea.KeyDelete:
		if m.CursorPos < len(m.InputBuffer) {
			m.InputBuffer = m.InputBuffer[:m.CursorPos] + m.InputBuffer[m.CursorPos+1:]
		}

	case tea.KeyLeft:
		if m.CursorPos > 0 {
			m.CursorPos--
		}

	case tea.KeyRight:
		if m.CursorPos < len(m.InputBuffer) {
			m.CursorPos++
		}

	case tea.KeyHome:
		m.CursorPos = 0

	case tea.KeyEnd:
		m.CursorPos = len(m.InputBuffer)

	case tea.KeyUp:
		if len(m.History) > 0 && m.HistoryIdx > 0 {
			m.HistoryIdx--
			m.InputBuffer = m.History[m.HistoryIdx]
			m.CursorPos = len(m.InputBuffer)
		}

	case tea.KeyDown:
		if m.HistoryIdx < len(m.History)-1 {
			m.HistoryIdx++
			m.InputBuffer = m.History[m.HistoryIdx]
			m.CursorPos = len(m.InputBuffer)
		} else {
			m.HistoryIdx = len(m.History)
			m.InputBuffer = ""
			m.CursorPos = 0
		}

	case tea.KeyCtrlC:
		return m, tea.Quit

	case tea.KeyCtrlL:
		m.ShowHelp = !m.ShowHelp

	default:
		if msg.String() == "/" {
			// 切换到命令模式
			m.InputBuffer = "/"
			m.CursorPos = 1
		}
	}

	return m, nil
}

// handleSelectKey 处理选择器按键
func (m *MainModel) handleSelectKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.SelectModel == nil {
		m.Mode = ModeInput
		return m, nil
	}

	switch msg.Type {
	case tea.KeyUp:
		m.SelectModel.MoveUp()

	case tea.KeyDown:
		m.SelectModel.MoveDown()

	case tea.KeyEnter:
		m.SelectModel.Select()
		m.Mode = ModeInput

	case tea.KeyEscape:
		m.SelectModel = nil
		m.Mode = ModeInput
	}

	return m, nil
}

// handleDialogKey 处理对话框按键
func (m *MainModel) handleDialogKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEscape:
		m.Dialog = nil
		m.Mode = ModeInput

	case tea.KeyEnter:
		// 确认对话框
		m.Dialog = nil
		m.Mode = ModeInput
	}

	return m, nil
}

// handleOnboardingKey 处理初始化向导按键
func (m *MainModel) handleOnboardingKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.Dialog == nil {
		m.Mode = ModeInput
		return m, nil
	}

	dialog, ok := m.Dialog.(*dialogs.OnboardingDialog)
	if !ok {
		m.Mode = ModeInput
		return m, nil
	}

	switch msg.Type {
	case tea.KeyEnter:
		dialog.NextStep()
		if dialog.IsComplete() {
			m.Mode = ModeInput
			m.Dialog = nil
		}

	case tea.KeyEscape:
		m.Mode = ModeInput
		m.Dialog = nil

	case tea.KeyUp, tea.KeyDown:
		// 主题选择等
	}

	return m, nil
}

// View 返回视图
func (m *MainModel) View() string {
	switch m.Mode {
	case ModeSpinner:
		return m.renderSpinnerView()

	case ModeSelect:
		return m.renderSelectView()

	case ModeOnboarding:
		return m.renderOnboardingView()

	default:
		return m.renderInputView()
	}
}

// renderSpinnerView 渲染加载动画视图
func (m *MainModel) renderSpinnerView() string {
	var lines []string

	// 渲染消息历史
	for _, msg := range m.Messages {
		roleStyle := lipgloss.NewStyle().Foreground(styles.MutedColor)
		if msg.Role == "user" {
			roleStyle = lipgloss.NewStyle().Foreground(styles.PrimaryColor)
		} else if msg.Role == "assistant" {
			roleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
		}
		lines = append(lines, roleStyle.Render(msg.Role+":"))
		lines = append(lines, msg.Content)
		lines = append(lines, "")
	}

	// 渲染加载动画
	if m.Spinner != nil && m.IsThinking {
		lines = append(lines, m.Spinner.Render())
	}

	// 帮助文本
	lines = append(lines, "")
	lines = append(lines, styles.HelpStyle.Render("Ctrl+C: 取消"))

	return joinLines(lines)
}

// renderSelectView 渲染选择器视图
func (m *MainModel) renderSelectView() string {
	if m.SelectModel == nil {
		return m.renderInputView()
	}

	var lines []string

	// 渲染消息历史
	for _, msg := range m.Messages {
		lines = append(lines, msg.Role+": "+msg.Content)
	}
	lines = append(lines, "")

	// 渲染选择器
	lines = append(lines, m.SelectModel.Render())

	return joinLines(lines)
}

// renderOnboardingView 渲染初始化向导视图
func (m *MainModel) renderOnboardingView() string {
	if m.Dialog == nil {
		return m.renderInputView()
	}

	dialog, ok := m.Dialog.(*dialogs.OnboardingDialog)
	if !ok {
		return m.renderInputView()
	}

	var lines []string

	// 进度指示
	lines = append(lines, dialog.Render())

	// 帮助文本
	lines = append(lines, "")
	lines = append(lines, styles.HelpStyle.Render("Enter: 继续 | Esc: 退出"))

	return joinLines(lines)
}

// renderInputView 渲染输入视图
func (m *MainModel) renderInputView() string {
	var lines []string

	// 渲染消息历史
	for _, msg := range m.Messages {
		roleStyle := lipgloss.NewStyle().Foreground(styles.MutedColor)
		if msg.Role == "user" {
			roleStyle = lipgloss.NewStyle().Foreground(styles.PrimaryColor)
		} else if msg.Role == "assistant" {
			roleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
		}
		lines = append(lines, roleStyle.Render(msg.Role+":"))
		lines = append(lines, msg.Content)
		lines = append(lines, "")
	}

	// 渲染输入框
	inputStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.FocusBorderColor).
		Padding(0, 1)

	// 构建带光标的输入行
	before := m.InputBuffer[:m.CursorPos]
	after := m.InputBuffer[m.CursorPos:]
	cursorStyle := lipgloss.NewStyle().
		Background(styles.PrimaryColor).
		Foreground(lipgloss.Color("0"))
	cursor := cursorStyle.Render(" ")

	inputLine := before + cursor + after
	if m.InputBuffer == "" {
		inputLine = styles.Dim.Render("输入消息或命令 (输入 / 打开命令列表)...")
	}

	lines = append(lines, inputStyle.Render(inputLine))

	// 帮助文本
	if m.ShowHelp {
		lines = append(lines, "")
		lines = append(lines, styles.HelpStyle.Render("↑↓: 历史 | Tab: 自动补全 | Ctrl+C: 退出"))
	}

	return joinLines(lines)
}

// joinLines 连接行
func joinLines(lines []string) string {
	result := ""
	for i, line := range lines {
		if i > 0 {
			result += "\n"
		}
		result += line
	}
	return result
}

// StartThinking 开始思考
func (m *MainModel) StartThinking(verb string) {
	m.IsThinking = true
	m.Spinner = components.NewSpinnerWithVerb(verb)
	m.Mode = ModeSpinner
}

// StopThinking 停止思考
func (m *MainModel) StopThinking() {
	m.IsThinking = false
	m.Mode = ModeInput
}

// AddMessage 添加消息
func (m *MainModel) AddMessage(role, content string) {
	m.Messages = append(m.Messages, MessageItem{
		Role:    role,
		Content: content,
	})
}

// StartSelect 开始选择
func (m *MainModel) StartSelect(options []*components.SelectOption) {
	m.SelectModel = components.NewSelectModel(options)
	m.Mode = ModeSelect
}

// StartOnboarding 开始初始化向导
func (m *MainModel) StartOnboarding() {
	m.Dialog = dialogs.NewOnboardingDialog()
	m.Mode = ModeOnboarding
}
