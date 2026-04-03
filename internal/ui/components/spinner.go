package components

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/claude-code-go/claude/internal/ui/styles"
)

// Spinner 加载动画组件
// 参考 TypeScript: src/components/Spinner.tsx
// 支持多种动画帧和状态显示

// 动画帧定义
var (
	// 标准加载动画帧
	SpinnerFrames = []string{
		"⠋",
		"⠙",
		"⠹",
		"⠸",
		"⠼",
		"⠴",
		"⠦",
		"⠧",
		"⠇",
		"⠏",
	}

	// 简洁加载动画帧
	BriefSpinnerFrames = []string{
		"◐",
		"◓",
		"◑",
		"◒",
	}
)

// SpinnerState Spinner状态
type SpinnerState struct {
	Frame       int          // 当前帧索引
	Verb        string       // 当前动作 (e.g., "Thinking...")
	StartTime   time.Time    // 开始时间
	IsActive    bool         // 是否活跃
	IsPaused    bool         // 是否暂停
	SubText     string       // 副文本
	NextTask    string       // 下一个任务提示
	TokenBudget *TokenBudget // Token预算显示
}

// TokenBudget Token预算信息
type TokenBudget struct {
	Used      int
	Remaining int
	Total     int
}

// NewSpinnerState 创建新的Spinner状态
func NewSpinnerState() *SpinnerState {
	return &SpinnerState{
		Frame:     0,
		Verb:      "Loading...",
		StartTime: time.Now(),
		IsActive:  true,
	}
}

// SetVerb 设置当前动作
func (s *SpinnerState) SetVerb(verb string) {
	s.Verb = verb
}

// SetSubText 设置副文本
func (s *SpinnerState) SetSubText(text string) {
	s.SubText = text
}

// SetNextTask 设置下一个任务
func (s *SpinnerState) SetNextTask(task string) {
	s.NextTask = task
}

// SetTokenBudget 设置Token预算
func (s *SpinnerState) SetTokenBudget(used, remaining, total int) {
	s.TokenBudget = &TokenBudget{
		Used:      used,
		Remaining: remaining,
		Total:     total,
	}
}

// Pause 暂停动画
func (s *SpinnerState) Pause() {
	s.IsPaused = true
}

// Resume 恢复动画
func (s *SpinnerState) Resume() {
	s.IsPaused = false
}

// Stop 停止动画
func (s *SpinnerState) Stop() {
	s.IsActive = false
}

// NextFrame 下一帧
func (s *SpinnerState) NextFrame() {
	if !s.IsPaused {
		s.Frame = (s.Frame + 1) % len(SpinnerFrames)
	}
}

// GetFrame 获取当前帧
func (s *SpinnerState) GetFrame() string {
	return SpinnerFrames[s.Frame]
}

// GetElapsed 获取已用时间
func (s *SpinnerState) GetElapsed() time.Duration {
	return time.Since(s.StartTime)
}

// FormatElapsed 格式化已用时间
func (s *SpinnerState) FormatElapsed() string {
	elapsed := s.GetElapsed()
	if elapsed < time.Minute {
		return fmt.Sprintf("%ds", int(elapsed.Seconds()))
	}
	minutes := int(elapsed.Minutes())
	seconds := int(elapsed.Seconds()) % 60
	return fmt.Sprintf("%dm %ds", minutes, seconds)
}

// Render 渲染Spinner
func (s *SpinnerState) Render() string {
	var lines []string

	// 动画帧 + 动作文本
	frame := s.GetFrame()
	elapsed := s.FormatElapsed()

	mainText := fmt.Sprintf("%s %s %s",
		lipgloss.NewStyle().Foreground(styles.PrimaryColor).Render(frame),
		styles.Bold.Render(s.Verb),
		styles.Dim.Render(elapsed),
	)
	lines = append(lines, mainText)

	// 副文本
	if s.SubText != "" {
		lines = append(lines, "  "+styles.Dim.Render(s.SubText))
	}

	// Token预算
	if s.TokenBudget != nil {
		budgetText := fmt.Sprintf("  Tokens: %d / %d (remaining: %d)",
			s.TokenBudget.Used,
			s.TokenBudget.Total,
			s.TokenBudget.Remaining,
		)
		lines = append(lines, styles.Dim.Render(budgetText))
	}

	// 下一个任务提示
	if s.NextTask != "" {
		lines = append(lines, "")
		lines = append(lines, styles.Dim.Render(fmt.Sprintf("  Next: %s", s.NextTask)))
	}

	return strings.Join(lines, "\n")
}

// SpinnerWithVerb 带动作描述的Spinner
// 参考 TypeScript: SpinnerWithVerb
type SpinnerWithVerb struct {
	state *SpinnerState
}

// NewSpinnerWithVerb 创建新的SpinnerWithVerb
func NewSpinnerWithVerb(verb string) *SpinnerWithVerb {
	state := NewSpinnerState()
	state.SetVerb(verb)
	return &SpinnerWithVerb{state: state}
}

// SetVerb 设置动作描述
func (s *SpinnerWithVerb) SetVerb(verb string) {
	s.state.SetVerb(verb)
}

// SetSubText 设置副文本
func (s *SpinnerWithVerb) SetSubText(text string) {
	s.state.SetSubText(text)
}

// SetNextTask 设置下一个任务
func (s *SpinnerWithVerb) SetNextTask(task string) {
	s.state.SetNextTask(task)
}

// SetTokenBudget 设置Token预算
func (s *SpinnerWithVerb) SetTokenBudget(used, remaining, total int) {
	s.state.SetTokenBudget(used, remaining, total)
}

// Tick 更新动画帧
func (s *SpinnerWithVerb) Tick() {
	s.state.NextFrame()
}

// Render 渲染
func (s *SpinnerWithVerb) Render() string {
	return s.state.Render()
}

// IsActive 检查是否活跃
func (s *SpinnerWithVerb) IsActive() bool {
	return s.state.IsActive
}

// Stop 停止
func (s *SpinnerWithVerb) Stop() {
	s.state.Stop()
}

// BriefSpinner 简洁Spinner
// 参考 TypeScript: BriefSpinner
type BriefSpinner struct {
	frame  int
	active bool
}

// NewBriefSpinner 创建新的BriefSpinner
func NewBriefSpinner() *BriefSpinner {
	return &BriefSpinner{
		frame:  0,
		active: true,
	}
}

// Tick 更新动画帧
func (s *BriefSpinner) Tick() {
	if s.active {
		s.frame = (s.frame + 1) % len(BriefSpinnerFrames)
	}
}

// Render 渲染
func (s *BriefSpinner) Render() string {
	if !s.active {
		return ""
	}
	return lipgloss.NewStyle().
		Foreground(styles.PrimaryColor).
		Render(BriefSpinnerFrames[s.frame])
}

// Stop 停止
func (s *BriefSpinner) Stop() {
	s.active = false
}

// IsActive 检查是否活跃
func (s *BriefSpinner) IsActive() bool {
	return s.active
}
