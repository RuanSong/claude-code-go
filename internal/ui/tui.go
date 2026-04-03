package ui

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/claude-code-go/claude/internal/ui/components"
)

// Message 消息结构
type Message struct {
	Role       string
	Content    string
	Timestamp  time.Time
	IsToolCall bool
	ToolName   string
}

// Model TUI模型 - 使用Bubble Tea架构
type Model struct {
	messages       []Message
	input          string
	quitting       bool
	sessionStart   time.Time
	commandHistory []string
	historyIndex   int

	// 新增: 模式支持
	mode           string // "input", "select", "spinner"
	selectModel    *components.SelectModel
	spinner        *components.SpinnerWithVerb
	pendingCommand string
}

// NewModel 创建新的TUI模型
func NewModel() *Model {
	return &Model{
		messages:       make([]Message, 0),
		input:          "",
		sessionStart:   time.Now(),
		commandHistory: make([]string, 0),
		historyIndex:   -1,
		mode:           "input",
		selectModel:    nil,
		spinner:        nil,
	}
}

// Init 初始化模型
func (m *Model) Init() tea.Cmd {
	m.messages = append(m.messages, Message{
		Role:      "system",
		Content:   "Welcome to Claude Code (Go)! Type /help for available commands.",
		Timestamp: time.Now(),
	})
	return nil
}

// Update 更新模型
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)

	case tea.WindowSizeMsg:
		// 窗口大小变化 - 暂时忽略
		return m, nil

	default:
		return m, nil
	}
}

// handleKey 处理按键
func (m *Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.mode {
	case "spinner":
		return m.handleSpinnerKey(msg)
	case "select":
		return m.handleSelectKey(msg)
	default:
		return m.handleInputKey(msg)
	}
}

// handleInputKey 处理输入模式按键
func (m *Model) handleInputKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlC, tea.KeyCtrlD, tea.KeyEsc:
		m.quitting = true
		return m, tea.Quit

	case tea.KeyEnter:
		if m.input != "" {
			m.commandHistory = append(m.commandHistory, m.input)
			m.historyIndex = len(m.commandHistory)

			m.messages = append(m.messages, Message{
				Role:      "user",
				Content:   m.input,
				Timestamp: time.Now(),
			})

			// 检查特殊命令
			if m.input == "/quit" || m.input == "/exit" {
				m.quitting = true
				return m, tea.Quit
			}

			if m.input == "/help" {
				m.messages = append(m.messages, Message{
					Role:      "assistant",
					Content:   getHelpText(),
					Timestamp: time.Now(),
				})
				m.input = ""
				return m, nil
			}

			// 检查是否需要显示选择器
			if m.input == "/model" {
				m.showModelSelect()
				return m, nil
			}

			// 模拟处理
			m.startSpinner("Thinking...")

			// 模拟异步响应
			go func() {
				time.Sleep(2 * time.Second)
				m.stopSpinner()
				m.messages = append(m.messages, Message{
					Role:      "assistant",
					Content:   fmt.Sprintf("Echo: %s", m.input),
					Timestamp: time.Now(),
				})
			}()

			m.input = ""
		}

	case tea.KeyBackspace:
		if len(m.input) > 0 {
			m.input = m.input[:len(m.input)-1]
		}

	case tea.KeyUp:
		if len(m.commandHistory) > 0 && m.historyIndex > 0 {
			m.historyIndex--
			m.input = m.commandHistory[m.historyIndex]
		}

	case tea.KeyDown:
		if m.historyIndex < len(m.commandHistory)-1 {
			m.historyIndex++
			m.input = m.commandHistory[m.historyIndex]
		} else {
			m.historyIndex = len(m.commandHistory)
			m.input = ""
		}

	case tea.KeyTab:
		m.input = completeCommand(m.input)

	case tea.KeyRunes:
		m.input += string(msg.Runes)

	default:
		// 其他按键
	}

	return m, nil
}

// handleSpinnerKey 处理加载动画模式按键
func (m *Model) handleSpinnerKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlC:
		m.stopSpinner()
	}
	return m, nil
}

// handleSelectKey 处理选择器模式按键
func (m *Model) handleSelectKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.selectModel == nil {
		m.mode = "input"
		return m, nil
	}

	switch msg.Type {
	case tea.KeyUp:
		m.selectModel.MoveUp()

	case tea.KeyDown:
		m.selectModel.MoveDown()

	case tea.KeyEnter:
		m.selectModel.Select()
		if m.selectModel.SelectedValue() != "" {
			m.messages = append(m.messages, Message{
				Role:      "assistant",
				Content:   fmt.Sprintf("Selected: %s", m.selectModel.SelectedValue()),
				Timestamp: time.Now(),
			})
		}
		m.mode = "input"
		m.selectModel = nil

	case tea.KeyEsc:
		m.mode = "input"
		m.selectModel = nil
	}

	return m, nil
}

// showModelSelect 显示模型选择器
func (m *Model) showModelSelect() {
	models := []*components.SelectOption{
		components.NewSelectOption("Claude Sonnet 4", "claude-sonnet-4-20250514"),
		components.NewSelectOption("Claude Opus 3", "claude-opus-3-5-20250514"),
		components.NewSelectOption("Claude Haiku", "claude-haiku-3-5-20250514"),
	}
	m.selectModel = components.NewSelectModel(models)
	m.selectModel.SetTitle("Select Model")
	m.mode = "select"
}

// startSpinner 开始加载动画
func (m *Model) startSpinner(verb string) {
	m.mode = "spinner"
	m.spinner = components.NewSpinnerWithVerb(verb)
}

// stopSpinner 停止加载动画
func (m *Model) stopSpinner() {
	m.mode = "input"
	m.spinner = nil
}

// View 返回视图
func (m *Model) View() string {
	switch m.mode {
	case "spinner":
		return m.renderSpinnerView()
	case "select":
		return m.renderSelectView()
	default:
		return m.renderInputView()
	}
}

// renderSpinnerView 渲染加载动画视图
func (m *Model) renderSpinnerView() string {
	var s string

	// 头部
	s = headerStyle.Render("┌─ Claude Code (Go)") + "\n"
	s += headerStyle.Render("│ Session: ") + messageStyle.Render(formatDuration(time.Since(m.sessionStart))) + "\n"
	s += headerStyle.Render("└─") + "\n\n"

	// 消息历史
	for _, msg := range m.messages {
		s += renderMessage(msg) + "\n"
	}

	// 加载动画
	if m.spinner != nil {
		s += "\n" + m.spinner.Render() + "\n"
		s += "\n" + dimStyle.Render("Press Ctrl+C to cancel")
	}

	return s
}

// renderSelectView 渲染选择器视图
func (m *Model) renderSelectView() string {
	var s string

	// 头部
	s = headerStyle.Render("┌─ Claude Code (Go)") + "\n"
	s += headerStyle.Render("│ Session: ") + messageStyle.Render(formatDuration(time.Since(m.sessionStart))) + "\n"
	s += headerStyle.Render("└─") + "\n\n"

	// 消息历史
	for _, msg := range m.messages {
		s += renderMessage(msg) + "\n"
	}

	// 选择器
	if m.selectModel != nil {
		s += "\n" + m.selectModel.Render() + "\n"
	}

	return s
}

// renderInputView 渲染输入视图
func (m *Model) renderInputView() string {
	var s string

	// 头部
	s = headerStyle.Render("┌─ Claude Code (Go)") + "\n"
	s += headerStyle.Render("│ Session: ") + messageStyle.Render(formatDuration(time.Since(m.sessionStart))) + "\n"
	s += headerStyle.Render("└─") + "\n\n"

	// 消息历史
	for _, msg := range m.messages {
		s += renderMessage(msg) + "\n"
	}

	// 输入框
	s += promptStyle.Render("> ") + inputStyle.Render(m.input) + "\n"

	// 帮助
	if !m.quitting {
		s += "\n" + dimStyle.Render("Ctrl+C quit  ↑↓ history  Tab complete")
	} else {
		s += "\n" + errorStyle.Render("Goodbye!")
	}

	return s
}

// renderMessage 渲染单条消息
func renderMessage(msg Message) string {
	switch msg.Role {
	case "system":
		return toolStyle.Render(msg.Content)
	case "user":
		return userMessageStyle.Render("You: ") + msg.Content
	case "assistant":
		return assistantMessageStyle.Render("Claude: ") + msg.Content
	case "tool":
		toolName := msg.ToolName
		if toolName == "" {
			toolName = "tool"
		}
		return toolStyle.Render(fmt.Sprintf("[%s] ", toolName)) + messageStyle.Render(msg.Content)
	default:
		return messageStyle.Render(msg.Content)
	}
}

// formatDuration 格式化时长
func formatDuration(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
	}
	if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	return fmt.Sprintf("%ds", seconds)
}

// completeCommand 命令补全
func completeCommand(input string) string {
	commands := []string{
		"/help", "/model", "/cost", "/tasks", "/compact",
		"/review", "/commit", "/diff", "/config", "/doctor",
		"/memory", "/resume", "/share", "/theme", "/keybindings",
		"/vim", "/exit", "/quit",
	}

	for _, cmd := range commands {
		if len(input) > 0 && strings.HasPrefix(cmd, input) {
			return cmd + " "
		}
	}
	return input
}

// getHelpText 获取帮助文本
func getHelpText() string {
	return `Available commands:
  /help     - Show this help
  /model    - Select AI model
  /cost     - Show cost statistics
  /tasks    - Manage tasks
  /compact  - Compact context
  /review   - Review code
  /commit   - Commit changes
  /diff     - Show changes
  /config   - Manage configuration
  /doctor   - Run diagnostics
  /exit     - Exit REPL
  /quit     - Exit REPL

Type any message to chat with Claude!`
}

// AddMessage 添加消息
func (m *Model) AddMessage(role, content string) {
	m.messages = append(m.messages, Message{
		Role:      role,
		Content:   content,
		Timestamp: time.Now(),
	})
}

// AddAssistantMessage 添加助手消息
func (m *Model) AddAssistantMessage(content string) {
	m.AddMessage("assistant", content)
}

// AddUserMessage 添加用户消息
func (m *Model) AddUserMessage(content string) {
	m.AddMessage("user", content)
}

// AddToolMessage添加工具消息
func (m *Model) AddToolMessage(toolName, content string) {
	m.messages = append(m.messages, Message{
		Role:       "tool",
		ToolName:   toolName,
		Content:    content,
		Timestamp:  time.Now(),
		IsToolCall: true,
	})
}

// GetMessages 获取所有消息
func (m *Model) GetMessages() []Message {
	return m.messages
}

// ClearMessages 清除消息
func (m *Model) ClearMessages() {
	m.messages = make([]Message, 0)
}

// Run 启动TUI
func Run() error {
	model := NewModel()
	program := tea.NewProgram(model, tea.WithAltScreen())
	_, err := program.Run()
	return err
}

// 样式定义
var (
	headerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("green")).
			Bold(true)

	promptStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("cyan"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("red"))

	messageStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("white"))

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8"))

	toolStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("yellow"))

	userMessageStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("cyan"))

	assistantMessageStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("green"))

	inputStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("white"))
)

// CheckTerminal 检查终端是否支持
func CheckTerminal() error {
	if os.Getenv("TERM") == "dumb" {
		return fmt.Errorf("dumb terminal detected, TUI not supported")
	}
	return nil
}
