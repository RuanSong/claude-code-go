package ui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

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

	toolStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("yellow"))

	userMessageStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("cyan"))

	assistantMessageStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("green"))
)

type Message struct {
	Role       string
	Content    string
	Timestamp  time.Time
	IsToolCall bool
	ToolName   string
}

type Model struct {
	messages       []Message
	input          string
	quitting       bool
	sessionStart   time.Time
	commandHistory []string
	historyIndex   int
}

// NewModel creates a new TUI model
func NewModel() *Model {
	return &Model{
		messages:       make([]Message, 0),
		input:          "",
		sessionStart:   time.Now(),
		commandHistory: make([]string, 0),
		historyIndex:   -1,
	}
}

func (m *Model) Init() tea.Cmd {
	m.messages = append(m.messages, Message{
		Role:      "system",
		Content:   "Welcome to Claude Code (Go)! Type /help for available commands.",
		Timestamp: time.Now(),
	})
	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
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

				if m.input == "/quit" || m.input == "/exit" {
					m.quitting = true
					return m, tea.Quit
				}

				m.input = ""
			}
		case tea.KeyBackspace:
			if len(m.input) > 0 {
				m.input = m.input[:len(m.input)-1]
			}
		case tea.KeyUp:
			if m.historyIndex > 0 {
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
		default:
			m.input += msg.String()
		}
	}
	return m, nil
}

func completeCommand(input string) string {
	commands := []string{
		"/help", "/model", "/cost", "/tasks", "/compact",
		"/review", "/commit", "/diff", "/config", "/doctor",
		"/memory", "/resume", "/share", "/theme", "/keybindings",
		"/vim", "/exit", "/quit",
	}

	for _, cmd := range commands {
		if len(input) > 0 && cmd[:len(input)] == input {
			return cmd + " "
		}
	}
	return input
}

func (m *Model) View() string {
	s := headerStyle.Render("┌─ Claude Code (Go)") + "\n"
	s += headerStyle.Render("│ Session: ") + messageStyle.Render(formatDuration(time.Since(m.sessionStart))) + "\n"
	s += headerStyle.Render("└─") + "\n\n"

	for _, msg := range m.messages {
		switch msg.Role {
		case "system":
			s += toolStyle.Render(msg.Content) + "\n\n"
		case "user":
			s += userMessageStyle.Render("You: ") + msg.Content + "\n\n"
		case "assistant":
			s += assistantMessageStyle.Render("Claude: ") + msg.Content + "\n\n"
		case "tool":
			toolName := msg.ToolName
			if toolName == "" {
				toolName = "tool"
			}
			s += toolStyle.Render(fmt.Sprintf("[%s] ", toolName)) + messageStyle.Render(msg.Content) + "\n\n"
		default:
			s += messageStyle.Render(msg.Content) + "\n\n"
		}
	}

	s += promptStyle.Render("> ") + messageStyle.Render(m.input) + "\n"

	if !m.quitting {
		s += "\n" + messageStyle.Render("Ctrl+C quit  ↑↓ history  Tab complete")
	} else {
		s += "\n" + errorStyle.Render("Goodbye!")
	}

	return s
}

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

func (m *Model) AddMessage(role, content string) {
	m.messages = append(m.messages, Message{
		Role:      role,
		Content:   content,
		Timestamp: time.Now(),
	})
}

func (m *Model) AddAssistantMessage(content string) {
	m.AddMessage("assistant", content)
}

func (m *Model) AddUserMessage(content string) {
	m.AddMessage("user", content)
}

func (m *Model) AddToolMessage(toolName, content string) {
	m.messages = append(m.messages, Message{
		Role:       "tool",
		ToolName:   toolName,
		Content:    content,
		Timestamp:  time.Now(),
		IsToolCall: true,
	})
}

func (m *Model) GetMessages() []Message {
	return m.messages
}

func (m *Model) ClearMessages() {
	m.messages = make([]Message, 0)
}

func Run() error {
	model := NewModel()
	program := tea.NewProgram(model)
	_, err := program.Run()
	return err
}
