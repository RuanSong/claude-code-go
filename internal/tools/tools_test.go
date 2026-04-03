package tools

import (
	"testing"
)

func TestBashTool(t *testing.T) {
	tool := &BashTool{}

	if tool.Name() != "Bash" {
		t.Errorf("Expected name 'Bash', got %s", tool.Name())
	}

	if tool.Description() == "" {
		t.Error("Description should not be empty")
	}
}

func TestReadTool(t *testing.T) {
	tool := &ReadTool{}

	if tool.Name() != "Read" {
		t.Errorf("Expected name 'Read', got %s", tool.Name())
	}

	if tool.Description() == "" {
		t.Error("Description should not be empty")
	}
}

func TestWriteTool(t *testing.T) {
	tool := &WriteTool{}

	if tool.Name() != "Write" {
		t.Errorf("Expected name 'Write', got %s", tool.Name())
	}
}

func TestGlobTool(t *testing.T) {
	tool := &GlobTool{}

	if tool.Name() != "Glob" {
		t.Errorf("Expected name 'Glob', got %s", tool.Name())
	}
}

func TestGrepTool(t *testing.T) {
	tool := &GrepTool{}

	if tool.Name() != "Grep" {
		t.Errorf("Expected name 'Grep', got %s", tool.Name())
	}
}

func TestFileEditTool(t *testing.T) {
	tool := &FileEditTool{}

	if tool.Name() != "Edit" {
		t.Errorf("Expected name 'Edit', got %s", tool.Name())
	}
}

func TestWebFetchTool(t *testing.T) {
	tool := &WebFetchTool{}

	if tool.Name() != "WebFetch" {
		t.Errorf("Expected name 'WebFetch', got %s", tool.Name())
	}
}

func TestWebSearchTool(t *testing.T) {
	tool := &WebSearchTool{}

	if tool.Name() != "WebSearch" {
		t.Errorf("Expected name 'WebSearch', got %s", tool.Name())
	}
}

func TestTodoWriteTool(t *testing.T) {
	tool := &TodoWriteTool{}

	if tool.Name() != "TodoWrite" {
		t.Errorf("Expected name 'TodoWrite', got %s", tool.Name())
	}
}

func TestTaskCreateTool(t *testing.T) {
	tool := &TaskCreateTool{}

	if tool.Name() != "TaskCreate" {
		t.Errorf("Expected name 'TaskCreate', got %s", tool.Name())
	}
}

func TestTaskListTool(t *testing.T) {
	tool := &TaskListTool{}

	if tool.Name() != "TaskList" {
		t.Errorf("Expected name 'TaskList', got %s", tool.Name())
	}
}

func TestTaskUpdateTool(t *testing.T) {
	tool := &TaskUpdateTool{}

	if tool.Name() != "TaskUpdate" {
		t.Errorf("Expected name 'TaskUpdate', got %s", tool.Name())
	}
}

func TestTaskGetTool(t *testing.T) {
	tool := &TaskGetTool{}

	if tool.Name() != "TaskGet" {
		t.Errorf("Expected name 'TaskGet', got %s", tool.Name())
	}
}

func TestTaskStopTool(t *testing.T) {
	tool := &TaskStopTool{}

	if tool.Name() != "TaskStop" {
		t.Errorf("Expected name 'TaskStop', got %s", tool.Name())
	}
}

func TestAgentTool(t *testing.T) {
	tool := &AgentTool{}

	if tool.Name() != "Agent" {
		t.Errorf("Expected name 'Agent', got %s", tool.Name())
	}
}

func TestAgentResultTool(t *testing.T) {
	tool := &AgentResultTool{}

	if tool.Name() != "AgentResult" {
		t.Errorf("Expected name 'AgentResult', got %s", tool.Name())
	}
}

func TestSendMessageTool(t *testing.T) {
	tool := &SendMessageTool{}

	if tool.Name() != "SendMessage" {
		t.Errorf("Expected name 'SendMessage', got %s", tool.Name())
	}
}

func TestSleepTool(t *testing.T) {
	tool := &SleepTool{}

	if tool.Name() != "Sleep" {
		t.Errorf("Expected name 'Sleep', got %s", tool.Name())
	}
}

func TestBriefTool(t *testing.T) {
	tool := &BriefTool{}

	if tool.Name() != "Brief" {
		t.Errorf("Expected name 'Brief', got %s", tool.Name())
	}
}

func TestConfigTool(t *testing.T) {
	tool := &ConfigTool{}

	if tool.Name() != "Config" {
		t.Errorf("Expected name 'Config', got %s", tool.Name())
	}
}

func TestTeamCreateTool(t *testing.T) {
	tool := &TeamCreateTool{}

	if tool.Name() != "TeamCreate" {
		t.Errorf("Expected name 'TeamCreate', got %s", tool.Name())
	}
}

func TestTeamDeleteTool(t *testing.T) {
	tool := &TeamDeleteTool{}

	if tool.Name() != "TeamDelete" {
		t.Errorf("Expected name 'TeamDelete', got %s", tool.Name())
	}
}

func TestToolSearchTool(t *testing.T) {
	tool := &ToolSearchTool{}

	if tool.Name() != "ToolSearch" {
		t.Errorf("Expected name 'ToolSearch', got %s", tool.Name())
	}
}

func TestSyntheticOutputTool(t *testing.T) {
	tool := &SyntheticOutputTool{}

	if tool.Name() != "SyntheticOutput" {
		t.Errorf("Expected name 'SyntheticOutput', got %s", tool.Name())
	}
}

func TestRemoteTriggerTool(t *testing.T) {
	tool := &RemoteTriggerTool{}

	if tool.Name() != "RemoteTrigger" {
		t.Errorf("Expected name 'RemoteTrigger', got %s", tool.Name())
	}
}

func TestMCPTool(t *testing.T) {
	tool := &MCPTool{}

	if tool.Name() != "MCP" {
		t.Errorf("Expected name 'MCP', got %s", tool.Name())
	}
}

func TestListMcpResourcesTool(t *testing.T) {
	tool := &ListMcpResourcesTool{}

	if tool.Name() != "ListMcpResources" {
		t.Errorf("Expected name 'ListMcpResources', got %s", tool.Name())
	}
}

func TestReadMcpResourceTool(t *testing.T) {
	tool := &ReadMcpResourceTool{}

	if tool.Name() != "ReadMcpResource" {
		t.Errorf("Expected name 'ReadMcpResource', got %s", tool.Name())
	}
}

func TestAskUserQuestionTool(t *testing.T) {
	tool := &AskUserQuestionTool{}

	if tool.Name() != "AskUserQuestion" {
		t.Errorf("Expected name 'AskUserQuestion', got %s", tool.Name())
	}
}

func TestEnterPlanModeTool(t *testing.T) {
	tool := &EnterPlanModeTool{}

	if tool.Name() != "EnterPlanMode" {
		t.Errorf("Expected name 'EnterPlanMode', got %s", tool.Name())
	}
}

func TestExitPlanModeTool(t *testing.T) {
	tool := &ExitPlanModeTool{}

	if tool.Name() != "ExitPlanMode" {
		t.Errorf("Expected name 'ExitPlanMode', got %s", tool.Name())
	}
}

func TestEnterWorktreeTool(t *testing.T) {
	tool := &EnterWorktreeTool{}

	if tool.Name() != "EnterWorktree" {
		t.Errorf("Expected name 'EnterWorktree', got %s", tool.Name())
	}
}

func TestExitWorktreeTool(t *testing.T) {
	tool := &ExitWorktreeTool{}

	if tool.Name() != "ExitWorktree" {
		t.Errorf("Expected name 'ExitWorktree', got %s", tool.Name())
	}
}

func TestSkillTool(t *testing.T) {
	tool := &SkillTool{}

	if tool.Name() != "Skill" {
		t.Errorf("Expected name 'Skill', got %s", tool.Name())
	}
}

func TestNotebookEditTool(t *testing.T) {
	tool := &NotebookEditTool{}

	if tool.Name() != "NotebookEdit" {
		t.Errorf("Expected name 'NotebookEdit', got %s", tool.Name())
	}
}

func TestPowerShellTool(t *testing.T) {
	tool := &PowerShellTool{}

	if tool.Name() != "PowerShell" {
		t.Errorf("Expected name 'PowerShell', got %s", tool.Name())
	}
}

func TestREPLTool(t *testing.T) {
	tool := &REPLTool{}

	if tool.Name() != "REPL" {
		t.Errorf("Expected name 'REPL', got %s", tool.Name())
	}
}

func TestScheduleCronTool(t *testing.T) {
	tool := &ScheduleCronTool{}

	if tool.Name() != "ScheduleCron" {
		t.Errorf("Expected name 'ScheduleCron', got %s", tool.Name())
	}
}

func TestMcpAuthTool(t *testing.T) {
	tool := &McpAuthTool{}

	if tool.Name() != "McpAuth" {
		t.Errorf("Expected name 'McpAuth', got %s", tool.Name())
	}
}

func TestTaskOutputTool(t *testing.T) {
	tool := &TaskOutputTool{}

	if tool.Name() != "TaskOutput" {
		t.Errorf("Expected name 'TaskOutput', got %s", tool.Name())
	}
}

func TestLSPTool(t *testing.T) {
	tool := &LSPTool{}

	if tool.Name() != "LSP" {
		t.Errorf("Expected name 'LSP', got %s", tool.Name())
	}
}

func TestGetExtendedTools(t *testing.T) {
	tools := GetExtendedTools()

	if len(tools) == 0 {
		t.Error("GetExtendedTools should return at least one tool")
	}

	// Check that all tools have names
	for _, tool := range tools {
		if tool.Name() == "" {
			t.Error("Tool should have a name")
		}
	}
}

func TestGetAllTools(t *testing.T) {
	tools := GetAllTools()

	if len(tools) == 0 {
		t.Error("GetAllTools should return at least one tool")
	}

	// All tools should have unique names
	names := make(map[string]bool)
	for _, tool := range tools {
		if names[tool.Name()] {
			t.Errorf("Duplicate tool name: %s", tool.Name())
		}
		names[tool.Name()] = true
	}
}

func TestToolPermission(t *testing.T) {
	tools := GetAllTools()

	for _, tool := range tools {
		perm := tool.Permission()
		if perm < 0 || perm > 3 {
			t.Errorf("Tool %s has invalid permission: %d", tool.Name(), perm)
		}
	}
}
