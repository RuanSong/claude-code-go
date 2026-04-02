package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/claude-code-go/claude/internal/engine"
	"github.com/claude-code-go/claude/pkg/schema"
)

var taskCounter int
var taskMu sync.Mutex

type Task struct {
	ID          string `json:"id"`
	Subject     string `json:"subject"`
	Description string `json:"description,omitempty"`
	Status      string `json:"status"`
	ActiveForm  string `json:"activeForm,omitempty"`
	CreatedAt   int64  `json:"createdAt"`
}

type TaskManager struct {
	mu    sync.RWMutex
	tasks map[string]Task
}

var taskManager = &TaskManager{
	tasks: make(map[string]Task),
}

func nextTaskID() string {
	taskMu.Lock()
	defer taskMu.Unlock()
	taskCounter++
	return fmt.Sprintf("%d", taskCounter)
}

type TaskCreateTool struct{}

func (t *TaskCreateTool) Name() string { return "TaskCreate" }

func (t *TaskCreateTool) Description() string {
	return "Create a new task in the task list"
}

func (t *TaskCreateTool) InputSchema() schema.Schema {
	return schema.Object{
		Properties: map[string]schema.Schema{
			"subject":     schema.String{},
			"description": schema.String{},
			"activeForm":  schema.String{},
		},
		Required: []string{"subject"},
	}
}

func (t *TaskCreateTool) Permission() engine.PermissionMode {
	return engine.PermissionNormal
}

type TaskCreateInput struct {
	Subject     string `json:"subject"`
	Description string `json:"description,omitempty"`
	ActiveForm  string `json:"activeForm,omitempty"`
}

type TaskCreateOutput struct {
	Task struct {
		ID      string `json:"id"`
		Subject string `json:"subject"`
	} `json:"task"`
}

func (t *TaskCreateTool) Execute(ctx context.Context, input json.RawMessage, execCtx *engine.ToolExecContext) (*engine.ToolResult, error) {
	var req TaskCreateInput
	if err := json.Unmarshal(input, &req); err != nil {
		return nil, fmt.Errorf("parse input: %w", err)
	}

	taskID := nextTaskID()
	task := Task{
		ID:          taskID,
		Subject:     req.Subject,
		Description: req.Description,
		Status:      "pending",
		ActiveForm:  req.ActiveForm,
		CreatedAt:   time.Now().Unix(),
	}

	taskManager.mu.Lock()
	taskManager.tasks[taskID] = task
	taskManager.mu.Unlock()

	output := TaskCreateOutput{}
	output.Task.ID = taskID
	output.Task.Subject = req.Subject

	resultJSON, _ := json.Marshal(output)

	return &engine.ToolResult{
		Content: []engine.ContentBlock{&engine.TextBlock{
			Text: fmt.Sprintf("Task #%s created: %s", taskID, req.Subject),
		}, &engine.TextBlock{Text: string(resultJSON)}},
	}, nil
}

type TaskListTool struct{}

func (t *TaskListTool) Name() string { return "TaskList" }

func (t *TaskListTool) Description() string {
	return "List all tasks in the task list"
}

func (t *TaskListTool) InputSchema() schema.Schema {
	return schema.Object{
		Properties: map[string]schema.Schema{},
	}
}

func (t *TaskListTool) Permission() engine.PermissionMode {
	return engine.PermissionReadonly
}

type TaskListOutput struct {
	Tasks []Task `json:"tasks"`
}

func (t *TaskListTool) Execute(ctx context.Context, input json.RawMessage, execCtx *engine.ToolExecContext) (*engine.ToolResult, error) {
	taskManager.mu.RLock()
	tasks := make([]Task, 0, len(taskManager.tasks))
	for _, task := range taskManager.tasks {
		tasks = append(tasks, task)
	}
	taskManager.mu.RUnlock()

	output := TaskListOutput{Tasks: tasks}
	resultJSON, _ := json.Marshal(output)

	var text string
	if len(tasks) == 0 {
		text = "No tasks in the list."
	} else {
		text = fmt.Sprintf("Tasks (%d):\n", len(tasks))
		for _, task := range tasks {
			text += fmt.Sprintf("  #%s [%s] %s\n", task.ID, task.Status, task.Subject)
		}
	}

	return &engine.ToolResult{
		Content: []engine.ContentBlock{&engine.TextBlock{Text: text}, &engine.TextBlock{Text: string(resultJSON)}},
	}, nil
}

type TaskUpdateTool struct{}

func (t *TaskUpdateTool) Name() string { return "TaskUpdate" }

func (t *TaskUpdateTool) Description() string {
	return "Update an existing task's status or details"
}

func (t *TaskUpdateTool) InputSchema() schema.Schema {
	return schema.Object{
		Properties: map[string]schema.Schema{
			"taskId":      schema.String{},
			"status":      schema.String{},
			"subject":     schema.String{},
			"description": schema.String{},
		},
		Required: []string{"taskId"},
	}
}

func (t *TaskUpdateTool) Permission() engine.PermissionMode {
	return engine.PermissionNormal
}

type TaskUpdateInput struct {
	TaskID      string `json:"taskId"`
	Status      string `json:"status,omitempty"`
	Subject     string `json:"subject,omitempty"`
	Description string `json:"description,omitempty"`
}

func (t *TaskUpdateTool) Execute(ctx context.Context, input json.RawMessage, execCtx *engine.ToolExecContext) (*engine.ToolResult, error) {
	var req TaskUpdateInput
	if err := json.Unmarshal(input, &req); err != nil {
		return nil, fmt.Errorf("parse input: %w", err)
	}

	taskManager.mu.Lock()
	defer taskManager.mu.Unlock()

	task, exists := taskManager.tasks[req.TaskID]
	if !exists {
		return &engine.ToolResult{
			Content: []engine.ContentBlock{&engine.TextBlock{Text: fmt.Sprintf("Task #%s not found", req.TaskID)}},
			IsError: true,
		}, nil
	}

	if req.Status != "" {
		task.Status = req.Status
	}
	if req.Subject != "" {
		task.Subject = req.Subject
	}
	if req.Description != "" {
		task.Description = req.Description
	}

	taskManager.tasks[req.TaskID] = task

	return &engine.ToolResult{
		Content: []engine.ContentBlock{&engine.TextBlock{
			Text: fmt.Sprintf("Task #%s updated: [%s] %s", req.TaskID, task.Status, task.Subject),
		}},
	}, nil
}

type TaskGetTool struct{}

func (t *TaskGetTool) Name() string { return "TaskGet" }

func (t *TaskGetTool) Description() string {
	return "Get details of a specific task"
}

func (t *TaskGetTool) InputSchema() schema.Schema {
	return schema.Object{
		Properties: map[string]schema.Schema{
			"taskId": schema.String{},
		},
		Required: []string{"taskId"},
	}
}

func (t *TaskGetTool) Permission() engine.PermissionMode {
	return engine.PermissionReadonly
}

type TaskGetInput struct {
	TaskID string `json:"taskId"`
}

func (t *TaskGetTool) Execute(ctx context.Context, input json.RawMessage, execCtx *engine.ToolExecContext) (*engine.ToolResult, error) {
	var req TaskGetInput
	if err := json.Unmarshal(input, &req); err != nil {
		return nil, fmt.Errorf("parse input: %w", err)
	}

	taskManager.mu.RLock()
	task, exists := taskManager.tasks[req.TaskID]
	taskManager.mu.RUnlock()

	if !exists {
		return &engine.ToolResult{
			Content: []engine.ContentBlock{&engine.TextBlock{Text: fmt.Sprintf("Task #%s not found", req.TaskID)}},
			IsError: true,
		}, nil
	}

	resultJSON, _ := json.Marshal(task)

	return &engine.ToolResult{
		Content: []engine.ContentBlock{&engine.TextBlock{Text: string(resultJSON)}},
	}, nil
}

type TaskStopTool struct{}

func (t *TaskStopTool) Name() string { return "TaskStop" }

func (t *TaskStopTool) Description() string {
	return "Stop or cancel a running task"
}

func (t *TaskStopTool) InputSchema() schema.Schema {
	return schema.Object{
		Properties: map[string]schema.Schema{
			"taskId": schema.String{},
			"reason": schema.String{},
		},
		Required: []string{"taskId"},
	}
}

func (t *TaskStopTool) Permission() engine.PermissionMode {
	return engine.PermissionNormal
}

type TaskStopInput struct {
	TaskID string `json:"taskId"`
	Reason string `json:"reason,omitempty"`
}

func (t *TaskStopTool) Execute(ctx context.Context, input json.RawMessage, execCtx *engine.ToolExecContext) (*engine.ToolResult, error) {
	var req TaskStopInput
	if err := json.Unmarshal(input, &req); err != nil {
		return nil, fmt.Errorf("parse input: %w", err)
	}

	taskManager.mu.Lock()
	defer taskManager.mu.Unlock()

	task, exists := taskManager.tasks[req.TaskID]
	if !exists {
		return &engine.ToolResult{
			Content: []engine.ContentBlock{&engine.TextBlock{Text: fmt.Sprintf("Task #%s not found", req.TaskID)}},
			IsError: true,
		}, nil
	}

	task.Status = "cancelled"
	taskManager.tasks[req.TaskID] = task

	return &engine.ToolResult{
		Content: []engine.ContentBlock{&engine.TextBlock{
			Text: fmt.Sprintf("Task #%s stopped", req.TaskID),
		}},
	}, nil
}

func GetTaskTools() []engine.Tool {
	return []engine.Tool{
		&TaskCreateTool{},
		&TaskListTool{},
		&TaskUpdateTool{},
		&TaskGetTool{},
		&TaskStopTool{},
	}
}
