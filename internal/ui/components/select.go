package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/claude-code-go/claude/internal/ui/styles"
)

// SelectOption 选择项
// 参考 TypeScript: SelectOption
type SelectOption struct {
	Label       string // 显示标签
	Description string // 描述文本
	Value       string // 值
	Disabled    bool   // 是否禁用
	Key         string // 快捷键
}

// NewSelectOption 创建选择项
func NewSelectOption(label, value string) *SelectOption {
	return &SelectOption{
		Label: label,
		Value: value,
	}
}

// WithDescription 设置描述
func (o *SelectOption) WithDescription(desc string) *SelectOption {
	o.Description = desc
	return o
}

// WithKey 设置快捷键
func (o *SelectOption) WithKey(key string) *SelectOption {
	o.Key = key
	return o
}

// Disable 禁用选项
func (o *SelectOption) Disable() *SelectOption {
	o.Disabled = true
	return o
}

// SelectModel Select选择器模型
// 使用 Bubble Tea 架构
type SelectModel struct {
	Options     []*SelectOption // 选项列表
	Cursor      int             // 当前光标位置
	Selected    int             // 已选择的索引
	FilterText  string          // 过滤文本
	ShowHelp    bool            // 是否显示帮助
	Title       string          // 标题
	Description string          // 描述
	MultiSelect bool            // 是否多选
	Checked     map[int]bool    // 多选时的选中状态
}

// NewSelectModel 创建Select模型
func NewSelectModel(options []*SelectOption) *SelectModel {
	return &SelectModel{
		Options:     options,
		Cursor:      0,
		Selected:    -1,
		ShowHelp:    true,
		MultiSelect: false,
		Checked:     make(map[int]bool),
	}
}

// NewSelectModelMulti 创建多选模型
func NewSelectModelMulti(options []*SelectOption) *SelectModel {
	model := NewSelectModel(options)
	model.MultiSelect = true
	return model
}

// SetTitle 设置标题
func (m *SelectModel) SetTitle(title string) *SelectModel {
	m.Title = title
	return m
}

// SetDescription 设置描述
func (m *SelectModel) SetDescription(desc string) *SelectModel {
	m.Description = desc
	return m
}

// SetShowHelp 设置是否显示帮助
func (m *SelectModel) SetShowHelp(show bool) *SelectModel {
	m.ShowHelp = show
	return m
}

// FilterOptions 过滤选项
func (m *SelectModel) FilterOptions() []*SelectOption {
	if m.FilterText == "" {
		return m.Options
	}

	filtered := make([]*SelectOption, 0)
	for _, opt := range m.Options {
		if strings.Contains(strings.ToLower(opt.Label), strings.ToLower(m.FilterText)) {
			filtered = append(filtered, opt)
		}
	}
	return filtered
}

// SelectedOption 获取已选择的选项
func (m *SelectModel) SelectedOption() *SelectOption {
	if m.Selected < 0 || m.Selected >= len(m.Options) {
		return nil
	}
	return m.Options[m.Selected]
}

// SelectedValue 获取已选择的值
func (m *SelectModel) SelectedValue() string {
	opt := m.SelectedOption()
	if opt == nil {
		return ""
	}
	return opt.Value
}

// SelectedValues 获取所有已选择的值 (多选模式)
func (m *SelectModel) SelectedValues() []string {
	values := make([]string, 0)
	for i, checked := range m.Checked {
		if checked && i < len(m.Options) {
			values = append(values, m.Options[i].Value)
		}
	}
	return values
}

// MoveUp 上移光标
func (m *SelectModel) MoveUp() {
	if m.Cursor > 0 {
		m.Cursor--
	}
}

// MoveDown 下移光标
func (m *SelectModel) MoveDown() {
	filtered := m.FilterOptions()
	if m.Cursor < len(filtered)-1 {
		m.Cursor++
	}
}

// Select 选择当前项
func (m *SelectModel) Select() {
	filtered := m.FilterOptions()
	if m.Cursor < len(filtered) {
		// 找到原始索引
		filteredOpt := filtered[m.Cursor]
		for i, opt := range m.Options {
			if opt == filteredOpt {
				m.Selected = i
				break
			}
		}
	}
}

// Toggle 多选模式切换选中状态
func (m *SelectModel) Toggle() {
	if !m.MultiSelect {
		return
	}
	filtered := m.FilterOptions()
	if m.Cursor < len(filtered) {
		filteredOpt := filtered[m.Cursor]
		for i, opt := range m.Options {
			if opt == filteredOpt {
				m.Checked[i] = !m.Checked[i]
				break
			}
		}
	}
}

// CheckAll 全选
func (m *SelectModel) CheckAll() {
	for i := range m.Options {
		m.Checked[i] = true
	}
}

// UncheckAll 取消全选
func (m *SelectModel) UncheckAll() {
	m.Checked = make(map[int]bool)
}

// IsChecked 检查是否选中
func (m *SelectModel) IsChecked(index int) bool {
	return m.Checked[index]
}

// Render 渲染Select
func (m *SelectModel) Render() string {
	var lines []string
	filtered := m.FilterOptions()

	// 标题
	if m.Title != "" {
		lines = append(lines, styles.TitleStyle.Render(m.Title))
		lines = append(lines, "")
	}

	// 描述
	if m.Description != "" {
		lines = append(lines, styles.SubtitleStyle.Render(m.Description))
		lines = append(lines, "")
	}

	// 选项列表
	for i, opt := range filtered {
		// 计算原始索引
		origIndex := -1
		for j, o := range m.Options {
			if o == opt {
				origIndex = j
				break
			}
		}

		// 光标指示器
		cursorStyle := lipgloss.NewStyle().Foreground(styles.PrimaryColor)
		cursorStr := "  "
		if i == m.Cursor {
			cursorStr = cursorStyle.Render("▶ ")
		}

		// 选项样式
		successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
		prefix := "[ ]"
		if m.MultiSelect {
			if m.IsChecked(origIndex) {
				prefix = successStyle.Render("[✓]")
			}
		} else if origIndex == m.Selected {
			prefix = cursorStyle.Render("[●]")
		}

		// 选项文本
		optionText := opt.Label
		if opt.Disabled {
			optionText = styles.Dim.Render(optionText + " (disabled)")
		} else if i == m.Cursor {
			optionText = styles.Bold.Render(optionText)
		}

		// 描述
		if opt.Description != "" {
			descStyle := styles.Dim
			if i == m.Cursor && !opt.Disabled {
				descStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("7"))
			}
			optionText += "\n    " + descStyle.Render(opt.Description)
		}

		// 快捷键
		if opt.Key != "" {
			optionText += " " + styles.KeyStyle.Render(opt.Key)
		}

		line := fmt.Sprintf("%s%s%s %s", cursorStr, prefix, strings.Repeat(" ", 2), optionText)
		lines = append(lines, line)
	}

	// 帮助文本
	if m.ShowHelp {
		lines = append(lines, "")
		if m.MultiSelect {
			lines = append(lines, styles.HelpStyle.Render("Space: 选择 | Enter: 确认 | Esc: 取消"))
		} else {
			lines = append(lines, styles.HelpStyle.Render("↑↓: 导航 | Enter: 确认 | Esc: 取消"))
		}
	}

	return strings.Join(lines, "\n")
}

// SelectView Select视图 (用于 Bubble Tea)
type SelectView struct {
	model *SelectModel
}

// NewSelectView 创建Select视图
func NewSelectView(model *SelectModel) *SelectView {
	return &SelectView{model: model}
}

// Update 更新模型
func (v *SelectView) Update(msg string) {
	switch msg {
	case "up":
		v.model.MoveUp()
	case "down":
		v.model.MoveDown()
	case "select":
		if v.model.MultiSelect {
			v.model.Toggle()
		} else {
			v.model.Select()
		}
	case "enter":
		if v.model.MultiSelect {
			// 确认多选
		} else {
			v.model.Select()
		}
	case " ":
		if v.model.MultiSelect {
			v.model.Toggle()
		}
	}
}

// View 返回视图字符串
func (v *SelectView) View() string {
	return v.model.Render()
}
