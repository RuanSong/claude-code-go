package ui

// UI包 - Claude Code Go 版本的终端UI组件
// 使用 Bubble Tea 框架实现交互式终端界面
// 参考 TypeScript: src/components/, src/ink.tsx

import (
	"github.com/claude-code-go/claude/internal/ui/components"
	"github.com/claude-code-go/claude/internal/ui/dialogs"
)

// UI组件包
// 包含所有可复用的UI组件

// Components 组件列表
var (
	// Spinner 加载动画
	Spinner = components.SpinnerWithVerb{}

	// Select 选择器
	Select = struct {
		NewModel      func(options []*components.SelectOption) *components.SelectModel
		NewModelMulti func(options []*components.SelectOption) *components.SelectModel
		NewOption     func(label, value string) *components.SelectOption
	}{
		NewModel:      components.NewSelectModel,
		NewModelMulti: components.NewSelectModelMulti,
		NewOption:     components.NewSelectOption,
	}
)

// Dialogs 对话框列表
var (
	// NewConfirmation 创建确认对话框
	NewConfirmation = func(title, description string) *dialogs.ConfirmationDialog {
		return dialogs.NewConfirmationDialog(title, description)
	}

	// NewMCPServerApproval 创建MCP服务器审批对话框
	NewMCPServerApproval = func(serverName string) *dialogs.MCPServerApprovalDialog {
		return dialogs.NewMCPServerApprovalDialog(serverName)
	}

	// NewMCPServerMultiselect 创建MCP多选对话框
	NewMCPServerMultiselect = func(servers []string) *dialogs.MCPServerMultiselectDialog {
		return dialogs.NewMCPServerMultiselectDialog(servers)
	}

	// NewTextInput 创建文本输入对话框
	NewTextInput = func(title, placeholder string) *dialogs.TextInputDialog {
		return dialogs.NewTextInputDialog(title, placeholder)
	}
)

// InitUI 初始化UI
// 在应用启动时调用
func InitUI() {
	// 样式系统在包加载时自动初始化
}
