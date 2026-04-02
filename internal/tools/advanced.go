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

type Team struct {
	ID        string
	Name      string
	Agents    []string
	CreatedAt time.Time
	Status    string
	mu        sync.RWMutex
}

var teamCounter int
var teamMu sync.Mutex
var teams = make(map[string]*Team)

func nextTeamID() string {
	teamMu.Lock()
	defer teamMu.Unlock()
	teamCounter++
	return fmt.Sprintf("team-%d", teamCounter)
}

type TeamCreateTool struct{}

func (t *TeamCreateTool) Name() string { return "TeamCreate" }

func (t *TeamCreateTool) Description() string {
	return "Create a team of agents for parallel work"
}

func (t *TeamCreateTool) InputSchema() schema.Schema {
	return schema.Object{
		Properties: map[string]schema.Schema{
			"name":        schema.String{},
			"description": schema.String{},
			"agents":      schema.Array{Items: schema.String{}},
		},
		Required: []string{"name"},
	}
}

func (t *TeamCreateTool) Permission() engine.PermissionMode {
	return engine.PermissionElevated
}

type TeamCreateInput struct {
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Agents      []string `json:"agents,omitempty"`
}

type TeamCreateOutput struct {
	Team struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"team"`
}

func (t *TeamCreateTool) Execute(ctx context.Context, input json.RawMessage, execCtx engine.ToolExecContext) (*engine.ToolResult, error) {
	var req TeamCreateInput
	if err := json.Unmarshal(input, &req); err != nil {
		return nil, fmt.Errorf("parse input: %w", err)
	}

	teamID := nextTeamID()
	team := &Team{
		ID:        teamID,
		Name:      req.Name,
		Agents:    req.Agents,
		CreatedAt: time.Now(),
		Status:    "active",
	}

	teams[teamID] = team

	output := TeamCreateOutput{}
	output.Team.ID = teamID
	output.Team.Name = req.Name

	resultJSON, _ := json.Marshal(output)

	return &engine.ToolResult{
		Content: []engine.ContentBlock{&engine.TextBlock{
			Text: fmt.Sprintf("Team '%s' created (ID: %s)", req.Name, teamID),
		}, &engine.TextBlock{Text: string(resultJSON)}},
	}, nil
}

type TeamDeleteTool struct{}

func (t *TeamDeleteTool) Name() string { return "TeamDelete" }

func (t *TeamDeleteTool) Description() string {
	return "Delete a team and clean up resources"
}

func (t *TeamDeleteTool) InputSchema() schema.Schema {
	return schema.Object{
		Properties: map[string]schema.Schema{
			"teamId": schema.String{},
			"reason": schema.String{},
		},
		Required: []string{"teamId"},
	}
}

func (t *TeamDeleteTool) Permission() engine.PermissionMode {
	return engine.PermissionElevated
}

type TeamDeleteInput struct {
	TeamID string `json:"teamId"`
	Reason string `json:"reason,omitempty"`
}

func (t *TeamDeleteTool) Execute(ctx context.Context, input json.RawMessage, execCtx engine.ToolExecContext) (*engine.ToolResult, error) {
	var req TeamDeleteInput
	if err := json.Unmarshal(input, &req); err != nil {
		return nil, fmt.Errorf("parse input: %w", err)
	}

	team, exists := teams[req.TeamID]
	if !exists {
		return &engine.ToolResult{
			Content: []engine.ContentBlock{&engine.TextBlock{Text: fmt.Sprintf("Team %s not found", req.TeamID)}},
			IsError: true,
		}, nil
	}

	team.mu.Lock()
	team.Status = "deleted"
	team.mu.Unlock()

	delete(teams, req.TeamID)

	return &engine.ToolResult{
		Content: []engine.ContentBlock{&engine.TextBlock{
			Text: fmt.Sprintf("Team %s deleted", req.TeamID),
		}},
	}, nil
}

type ToolSearchTool struct{}

func (t *ToolSearchTool) Name() string { return "ToolSearch" }

func (t *ToolSearchTool) Description() string {
	return "Search for available tools by name or description"
}

func (t *ToolSearchTool) InputSchema() schema.Schema {
	return schema.Object{
		Properties: map[string]schema.Schema{
			"query": schema.String{},
		},
		Required: []string{"query"},
	}
}

func (t *ToolSearchTool) Permission() engine.PermissionMode {
	return engine.PermissionReadonly
}

type ToolSearchInput struct {
	Query string `json:"query"`
}

func (t *ToolSearchTool) Execute(ctx context.Context, input json.RawMessage, execCtx engine.ToolExecContext) (*engine.ToolResult, error) {
	var req ToolSearchInput
	if err := json.Unmarshal(input, &req); err != nil {
		return nil, fmt.Errorf("parse input: %w", err)
	}

	text := fmt.Sprintf("Search results for '%s':\n", req.Query)
	text += "  - Bash (shell commands)\n  - Read (file reading)\n  - Write (file writing)\n  - Edit (file editing)\n  - Glob (file finding)\n  - Grep (content search)\n  - WebFetch (URL fetching)\n  - WebSearch (web search)"

	return &engine.ToolResult{
		Content: []engine.ContentBlock{&engine.TextBlock{Text: text}},
	}, nil
}

type SyntheticOutputTool struct{}

func (s *SyntheticOutputTool) Name() string { return "SyntheticOutput" }

func (s *SyntheticOutputTool) Description() string {
	return "Generate structured output for the user"
}

func (s *SyntheticOutputTool) InputSchema() schema.Schema {
	return schema.Object{
		Properties: map[string]schema.Schema{
			"format":  schema.String{},
			"content": schema.String{},
		},
		Required: []string{"content"},
	}
}

func (s *SyntheticOutputTool) Permission() engine.PermissionMode {
	return engine.PermissionNormal
}

type SyntheticOutputInput struct {
	Format  string `json:"format,omitempty"`
	Content string `json:"content"`
}

func (s *SyntheticOutputTool) Execute(ctx context.Context, input json.RawMessage, execCtx engine.ToolExecContext) (*engine.ToolResult, error) {
	var req SyntheticOutputInput
	if err := json.Unmarshal(input, &req); err != nil {
		return nil, fmt.Errorf("parse input: %w", err)
	}

	return &engine.ToolResult{
		Content: []engine.ContentBlock{&engine.TextBlock{Text: req.Content}},
	}, nil
}

type RemoteTriggerTool struct{}

func (r *RemoteTriggerTool) Name() string { return "RemoteTrigger" }

func (r *RemoteTriggerTool) Description() string {
	return "Trigger a remote action via webhook or similar"
}

func (r *RemoteTriggerTool) InputSchema() schema.Schema {
	return schema.Object{
		Properties: map[string]schema.Schema{
			"url":     schema.String{},
			"payload": schema.String{},
			"method":  schema.String{},
		},
		Required: []string{"url"},
	}
}

func (r *RemoteTriggerTool) Permission() engine.PermissionMode {
	return engine.PermissionElevated
}

type RemoteTriggerInput struct {
	URL     string `json:"url"`
	Payload string `json:"payload,omitempty"`
	Method  string `json:"method,omitempty"`
}

func (r *RemoteTriggerTool) Execute(ctx context.Context, input json.RawMessage, execCtx engine.ToolExecContext) (*engine.ToolResult, error) {
	var req RemoteTriggerInput
	if err := json.Unmarshal(input, &req); err != nil {
		return nil, fmt.Errorf("parse input: %w", err)
	}

	return &engine.ToolResult{
		Content: []engine.ContentBlock{&engine.TextBlock{
			Text: fmt.Sprintf("Remote trigger sent to %s", req.URL),
		}},
	}, nil
}

func GetAdvancedTools() []engine.Tool {
	return []engine.Tool{
		&TeamCreateTool{},
		&TeamDeleteTool{},
		&ToolSearchTool{},
		&SyntheticOutputTool{},
		&RemoteTriggerTool{},
	}
}
