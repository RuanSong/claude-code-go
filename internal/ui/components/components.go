package components

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	spinnerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("yellow"))

	progressStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("cyan"))

	tableStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("white"))

	tableHeaderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("green")).
				Bold(true)

	dialogStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("white")).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("green"))

	dialogTitleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("cyan")).
				Bold(true)

	statusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("white")).
			Background(lipgloss.Color("blue"))

	brightStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("white"))
)

type SpinnerModel struct {
	frames []string
	index  int
	width  int
	done   bool
}

func NewSpinner() *SpinnerModel {
	return &SpinnerModel{
		frames: []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
		index:  0,
		width:  10,
	}
}

func (s *SpinnerModel) Init() tea.Cmd {
	return func() tea.Msg {
		time.Sleep(80 * time.Millisecond)
		return spinnerTickMsg{}
	}
}

type spinnerTickMsg struct{}

func (s *SpinnerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if _, ok := msg.(spinnerTickMsg); ok && !s.done {
		s.index++
		return s, s.Init()
	}
	return s, nil
}

func (s *SpinnerModel) View() string {
	if s.done {
		return ""
	}
	frame := s.frames[s.index%len(s.frames)]
	return spinnerStyle.Render(frame)
}

func (s *SpinnerModel) Stop() {
	s.done = true
}

type ProgressModel struct {
	total       int
	current     int
	prefix      string
	showPercent bool
}

func NewProgress(total int, prefix string) *ProgressModel {
	return &ProgressModel{
		total:       total,
		current:     0,
		prefix:      prefix,
		showPercent: true,
	}
}

func (p *ProgressModel) SetProgress(current int) {
	p.current = current
}

func (p *ProgressModel) Increment() {
	if p.current < p.total {
		p.current++
	}
}

func (p *ProgressModel) View() string {
	if p.total == 0 {
		return progressStyle.Render(fmt.Sprintf("%s ...", p.prefix))
	}

	percent := float64(p.current) / float64(p.total) * 100
	barWidth := 20
	filled := int(float64(barWidth) * float64(p.current) / float64(p.total))

	bar := "["
	for i := 0; i < filled; i++ {
		bar += "="
	}
	if filled < barWidth {
		bar += ">"
	}
	for i := filled + 1; i < barWidth; i++ {
		bar += " "
	}
	bar += "]"

	percentStr := ""
	if p.showPercent {
		percentStr = fmt.Sprintf(" %.0f%%", percent)
	}

	return progressStyle.Render(fmt.Sprintf("%s %s %d/%d%s", p.prefix, bar, p.current, p.total, percentStr))
}

type TableModel struct {
	headers []string
	rows    [][]string
	widths  []int
}

func NewTable(headers []string, widths []int) *TableModel {
	return &TableModel{
		headers: headers,
		rows:    make([][]string, 0),
		widths:  widths,
	}
}

func (t *TableModel) AddRow(row []string) {
	if len(row) == len(t.headers) {
		t.rows = append(t.rows, row)
	}
}

func (t *TableModel) View() string {
	if len(t.headers) == 0 {
		return ""
	}

	result := ""

	headerLine := "┌"
	for i, h := range t.headers {
		width := t.widths[i]
		padded := padString(h, width)
		headerLine += fmt.Sprintf("%s┬", tableHeaderStyle.Render(padded))
	}
	headerLine = headerLine[:len(headerLine)-1] + "┐"
	result += headerLine + "\n"

	for _, row := range t.rows {
		rowLine := "│"
		for i, cell := range row {
			width := t.widths[i]
			if width == 0 {
				width = len(cell)
			}
			padded := padString(cell, width)
			rowLine += fmt.Sprintf("%s│", tableStyle.Render(padded))
		}
		result += rowLine + "\n"
	}

	footerLine := "└"
	for i := range t.headers {
		width := t.widths[i]
		footerLine += fmt.Sprintf("%s┴", stringsRepeat("─", width))
	}
	footerLine = footerLine[:len(footerLine)-1] + "┘"
	result += footerLine

	return result
}

func padString(s string, width int) string {
	if len(s) >= width {
		return s[:width]
	}
	return s + stringsRepeat(" ", width-len(s))
}

func stringsRepeat(s string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}

type DialogModel struct {
	title    string
	message  string
	choices  []string
	selected int
	done     bool
	result   string
}

func NewDialog(title, message string, choices []string) *DialogModel {
	return &DialogModel{
		title:    title,
		message:  message,
		choices:  choices,
		selected: 0,
		done:     false,
	}
}

func (d *DialogModel) Init() tea.Cmd {
	return nil
}

func (d *DialogModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if d.done {
		return d, tea.Quit
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyUp, tea.KeyLeft:
			if d.selected > 0 {
				d.selected--
			}
		case tea.KeyDown, tea.KeyRight:
			if d.selected < len(d.choices)-1 {
				d.selected++
			}
		case tea.KeyEnter:
			d.done = true
			d.result = d.choices[d.selected]
			return d, tea.Quit
		}
	}
	return d, nil
}

func (d *DialogModel) View() string {
	width := 50
	padding := 2

	border := stringsRepeat("─", width-2)
	content := fmt.Sprintf("┌%s┐\n", border)
	content += fmt.Sprintf("│%s%s%s│\n", stringsRepeat(" ", padding), dialogTitleStyle.Render(centerString(d.title, width-4)), stringsRepeat(" ", padding))
	content += fmt.Sprintf("│%s┘\n", border[:width-2])
	content += fmt.Sprintf("│\n")
	content += fmt.Sprintf("│  %s\n", d.message)
	content += fmt.Sprintf("│\n")

	for i, choice := range d.choices {
		prefix := "  "
		if i == d.selected {
			prefix = " ●"
		}
		content += fmt.Sprintf("│%s %s%s│\n", prefix, choice, stringsRepeat(" ", width-len(choice)-4))
	}

	content += fmt.Sprintf("│\n")
	content += fmt.Sprintf("│%s┘\n", stringsRepeat("─", width-2))

	return dialogStyle.Render(content)
}

func centerString(s string, width int) string {
	padding := (width - len(s)) / 2
	return stringsRepeat(" ", padding) + s + stringsRepeat(" ", width-len(s)-padding)
}

func (d *DialogModel) Result() string {
	return d.result
}

type StatusBarModel struct {
	left  string
	right string
}

func NewStatusBar(left, right string) *StatusBarModel {
	return &StatusBarModel{
		left:  left,
		right: right,
	}
}

func (s *StatusBarModel) View() string {
	width := 80
	sep := " │ "

	totalLen := len(s.left) + len(s.right) + len(sep) + 4
	if totalLen < width {
		spaces := width - totalLen + len(s.right)
		return fmt.Sprintf(" %s %s%s%s ", statusBarStyle.Render(""), s.left, sep, stringsRepeat(" ", spaces-len(s.right))+s.right)
	}

	return fmt.Sprintf(" %s %s %s ", s.left, sep, s.right)
}

type ConfirmDialogModel struct {
	title   string
	message string
	confirm bool
	done    bool
}

func NewConfirmDialog(title, message string) *ConfirmDialogModel {
	return &ConfirmDialogModel{
		title:   title,
		message: message,
		confirm: false,
		done:    false,
	}
}

func (c *ConfirmDialogModel) Init() tea.Cmd {
	return nil
}

func (c *ConfirmDialogModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if c.done {
		return c, tea.Quit
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyLeft, tea.KeyRight:
			c.confirm = !c.confirm
		case tea.KeyEnter:
			c.done = true
			return c, tea.Quit
		case tea.KeyEscape:
			c.done = true
			c.confirm = false
			return c, tea.Quit
		}
	}
	return c, nil
}

func (c *ConfirmDialogModel) View() string {
	yesSelected := "○"
	noSelected := "○"
	if c.confirm {
		yesSelected = "●"
	} else {
		noSelected = "●"
	}

	return fmt.Sprintf(
		"\n  %s\n\n  %s\n\n    %s Yes    %s No\n\n  Press Enter to confirm, Esc to cancel\n",
		dialogTitleStyle.Render(c.title),
		c.message,
		yesSelected,
		noSelected,
	)
}

func (c *ConfirmDialogModel) Confirmed() bool {
	return c.confirm
}

type InputDialogModel struct {
	title       string
	message     string
	value       string
	placeholder string
	done        bool
	result      string
}

func NewInputDialog(title, message, placeholder string) *InputDialogModel {
	return &InputDialogModel{
		title:       title,
		message:     message,
		value:       "",
		placeholder: placeholder,
		done:        false,
		result:      "",
	}
}

func (i *InputDialogModel) Init() tea.Cmd {
	return nil
}

func (i *InputDialogModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if i.done {
		return i, tea.Quit
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			i.done = true
			i.result = i.value
			return i, tea.Quit
		case tea.KeyBackspace:
			if len(i.value) > 0 {
				i.value = i.value[:len(i.value)-1]
			}
		case tea.KeyEscape:
			i.done = true
			i.result = ""
			return i, tea.Quit
		default:
			i.value += msg.String()
		}
	}
	return i, nil
}

func (i *InputDialogModel) View() string {
	placeholder := i.placeholder
	if i.value != "" {
		placeholder = i.value
	}

	return fmt.Sprintf(
		"\n  %s\n\n  %s\n\n  > %s\n\n  Press Enter to confirm, Esc to cancel\n",
		dialogTitleStyle.Render(i.title),
		i.message,
		lipgloss.NewStyle().Foreground(lipgloss.Color("cyan")).Render(placeholder),
	)
}

func (i *InputDialogModel) Result() string {
	return i.result
}
