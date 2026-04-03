package components

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/claude-code-go/claude/internal/ui/styles"
)

// Markdown Markdown渲染组件
// 参考 TypeScript: src/components/Markdown.tsx
// 支持 GitHub Flavored Markdown (GFM)

// MarkdownRenderer Markdown渲染器
type MarkdownRenderer struct {
	CodeHighlighter func(code string, lang string) string
}

// NewMarkdownRenderer 创建Markdown渲染器
func NewMarkdownRenderer() *MarkdownRenderer {
	return &MarkdownRenderer{
		CodeHighlighter: defaultCodeHighlighter,
	}
}

// defaultCodeHighlighter 默认代码高亮
func defaultCodeHighlighter(code, lang string) string {
	// 简单的语法高亮模拟
	if lang == "" {
		lang = "text"
	}
	header := lipgloss.NewStyle().
		Foreground(lipgloss.Color("12")).
		Render(fmt.Sprintf("```%s", lang))
	return fmt.Sprintf("%s\n%s\n%s", header, code, "```")
}

// Render 渲染Markdown文本
func (r *MarkdownRenderer) Render(md string) string {
	if md == "" {
		return ""
	}

	// 检测是否为纯文本（无markdown语法）
	if !containsMarkdownSyntax(md) {
		return md
	}

	var lines []string
	var inCodeBlock bool
	var codeBlockLang string
	var codeBlockContent []string

	for _, line := range strings.Split(md, "\n") {
		// 代码块处理
		if strings.HasPrefix(line, "```") {
			if inCodeBlock {
				// 代码块结束
				code := strings.Join(codeBlockContent, "\n")
				lines = append(lines, r.CodeHighlighter(code, codeBlockLang))
				lines = append(lines, "")
				inCodeBlock = false
				codeBlockLang = ""
				codeBlockContent = nil
			} else {
				// 代码块开始
				inCodeBlock = true
				codeBlockLang = strings.TrimPrefix(line, "```")
				codeBlockContent = make([]string, 0)
			}
			continue
		}

		if inCodeBlock {
			codeBlockContent = append(codeBlockContent, line)
			continue
		}

		// 行内处理
		line = r.renderInline(line)
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

// containsMarkdownSyntax 检测是否包含Markdown语法
func containsMarkdownSyntax(md string) bool {
	patterns := []string{
		`\*\*[^*]+\*\*`,  // 粗体
		`\*[^*]+\*`,      // 斜体
		`__[^_]+__`,      // 粗体
		`_[^_]+_`,        // 斜体
		`\[.+?\]\(.+?\)`, // 链接
		`#+\s`,           // 标题
		`^\s*[-*+]\s`,    // 无序列表
		`^\s*\d+\.\s`,    // 有序列表
		`^\s*>`,          // 引用
		"```",            // 代码块
		"`[^`]+`",        // 行内代码
	}

	for _, pattern := range patterns {
		matched, _ := regexp.MatchString(pattern, md)
		if matched {
			return true
		}
	}
	return false
}

// renderInline 渲染行内元素
func (r *MarkdownRenderer) renderInline(line string) string {
	// 标题
	if strings.HasPrefix(line, "# ") {
		return styles.TitleStyle.Render(line[2:])
	}
	if strings.HasPrefix(line, "## ") {
		return lipgloss.NewStyle().Bold(true).Foreground(styles.PrimaryColor).Render(line[3:])
	}
	if strings.HasPrefix(line, "### ") {
		return lipgloss.NewStyle().Bold(true).Foreground(styles.TextColor).Render(line[4:])
	}

	// 引用
	if strings.HasPrefix(line, "> ") {
		return styles.Dim.Render("│ " + line[2:])
	}

	// 无序列表
	if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ") {
		return "  • " + r.renderBoldItalic(line[2:])
	}

	// 有序列表
	if matched, _ := regexp.MatchString(`^\d+\.\s`, line); matched {
		parts := strings.SplitN(line, ". ", 2)
		if len(parts) == 2 {
			return fmt.Sprintf("  %s. %s", lipgloss.NewStyle().Foreground(styles.PrimaryColor).Render(parts[0]), r.renderBoldItalic(parts[1]))
		}
	}

	// 水平线
	if strings.Contains(line, "---") || strings.Contains(line, "***") || strings.Contains(line, "___") {
		return styles.Dim.Render("────────────────────────────────")
	}

	// 行内代码
	line = r.renderInlineCode(line)

	// 粗体和斜体
	line = r.renderBoldItalic(line)

	// 链接 (简化处理)
	line = r.renderLinks(line)

	return line
}

// renderInlineCode 渲染行内代码
func (r *MarkdownRenderer) renderInlineCode(line string) string {
	re := regexp.MustCompile("`([^`]+)`")
	return re.ReplaceAllStringFunc(line, func(match string) string {
		code := match[1 : len(match)-1]
		return lipgloss.NewStyle().
			Background(lipgloss.Color("8")).
			Foreground(lipgloss.Color("15")).
			Render(code)
	})
}

// renderBoldItalic 渲染粗体和斜体
func (r *MarkdownRenderer) renderBoldItalic(line string) string {
	// 粗体 **text**
	re := regexp.MustCompile(`\*\*([^*]+)\*\*`)
	line = re.ReplaceAllStringFunc(line, func(match string) string {
		text := match[2 : len(match)-2]
		return lipgloss.NewStyle().Bold(true).Render(text)
	})

	// 斜体 *text*
	re = regexp.MustCompile(`\*([^*]+)\*`)
	line = re.ReplaceAllStringFunc(line, func(match string) string {
		text := match[1 : len(match)-1]
		return lipgloss.NewStyle().Italic(true).Render(text)
	})

	// 粗体 __text__
	re = regexp.MustCompile(`__([^_]+)__`)
	line = re.ReplaceAllStringFunc(line, func(match string) string {
		text := match[2 : len(match)-2]
		return lipgloss.NewStyle().Bold(true).Render(text)
	})

	// 斜体 _text_
	re = regexp.MustCompile(`_([^_]+)_`)
	line = re.ReplaceAllStringFunc(line, func(match string) string {
		text := match[1 : len(match)-1]
		return lipgloss.NewStyle().Italic(true).Render(text)
	})

	return line
}

// renderLinks 渲染链接
func (r *MarkdownRenderer) renderLinks(line string) string {
	re := regexp.MustCompile(`\[([^\]]+)\]\(([^\)]+)\)`)
	return re.ReplaceAllStringFunc(line, func(match string) string {
		parts := re.FindStringSubmatch(match)
		if len(parts) == 3 {
			text := parts[1]
			url := parts[2]
			return lipgloss.NewStyle().
				Foreground(styles.AccentColor).
				Underline(true).
				Render(text) + lipgloss.NewStyle().Foreground(styles.MutedColor).Render(" ("+url+")")
		}
		return match
	})
}

// MessageResponse 消息响应组件
// 参考 TypeScript: src/components/MessageResponse.tsx
type MessageResponse struct {
	Content    string
	IsNested   bool
	ShowCursor bool
	Width      int
}

// NewMessageResponse 创建消息响应
func NewMessageResponse(content string) *MessageResponse {
	return &MessageResponse{
		Content:    content,
		IsNested:   false,
		ShowCursor: true,
		Width:      80,
	}
}

// SetNested 设置嵌套模式
func (m *MessageResponse) SetNested(nested bool) *MessageResponse {
	m.IsNested = nested
	return m
}

// SetShowCursor 设置是否显示光标
func (m *MessageResponse) SetShowCursor(show bool) *MessageResponse {
	m.ShowCursor = show
	return m
}

// SetWidth 设置宽度
func (m *MessageResponse) SetWidth(width int) *MessageResponse {
	m.Width = width
	return m
}

// Render 渲染消息响应
func (m *MessageResponse) Render() string {
	renderer := NewMarkdownRenderer()
	content := renderer.Render(m.Content)

	if !m.IsNested {
		// 非嵌套模式，添加树形连接符
		treeChar := styles.TreeConnector.Render(styles.TreeConnectorChar)
		lines := strings.Split(content, "\n")
		for i, line := range lines {
			if i == 0 {
				lines[i] = treeChar + line
			} else {
				// 后续行缩进
				lines[i] = strings.Repeat(" ", lipgloss.Width(treeChar)) + line
			}
		}
		content = strings.Join(lines, "\n")
	}

	return content
}

// StreamingMarkdown 流式Markdown渲染
// 用于在流式输出时逐步显示内容
type StreamingMarkdown struct {
	*MarkdownRenderer
	Buffer    string
	Completed string
}

// NewStreamingMarkdown 创建流式Markdown渲染器
func NewStreamingMarkdown() *StreamingMarkdown {
	return &StreamingMarkdown{
		MarkdownRenderer: NewMarkdownRenderer(),
		Buffer:           "",
		Completed:        "",
	}
}

// Append 添加新内容
func (s *StreamingMarkdown) Append(text string) {
	s.Buffer += text
}

// Complete 完成当前缓冲
func (s *StreamingMarkdown) Complete() {
	s.Completed += s.Buffer
	s.Buffer = ""
}

// RenderBuffer 渲染当前缓冲
func (s *StreamingMarkdown) RenderBuffer() string {
	if s.Buffer == "" {
		return ""
	}
	return s.Render(s.Buffer)
}

// RenderAll 渲染所有内容
func (s *StreamingMarkdown) RenderAll() string {
	return s.Render(s.Completed + s.Buffer)
}

// FullContent 获取完整内容
func (s *StreamingMarkdown) FullContent() string {
	return s.Completed + s.Buffer
}
