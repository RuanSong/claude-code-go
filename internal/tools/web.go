package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/claude-code-go/claude/internal/engine"
	"github.com/claude-code-go/claude/pkg/schema"
)

// WebFetchTool fetches content from URLs
type WebFetchTool struct{}

func (w *WebFetchTool) Name() string { return "WebFetch" }

func (w *WebFetchTool) Description() string {
	return "Fetch and extract content from a URL"
}

func (w *WebFetchTool) InputSchema() schema.Schema {
	return schema.Object{
		Properties: map[string]schema.Schema{
			"url":    schema.String{},
			"prompt": schema.String{},
		},
		Required: []string{"url"},
	}
}

func (w *WebFetchTool) Permission() engine.PermissionMode {
	return engine.PermissionReadonly
}

type WebFetchResult struct {
	Bytes      int    `json:"bytes"`
	Code       int    `json:"code"`
	CodeText   string `json:"codeText"`
	Result     string `json:"result"`
	DurationMs int64  `json:"durationMs"`
	URL        string `json:"url"`
}

func (w *WebFetchTool) Execute(ctx context.Context, input json.RawMessage, execCtx engine.ToolExecContext) (*engine.ToolResult, error) {
	var req struct {
		URL    string `json:"url"`
		Prompt string `json:"prompt"`
	}

	if err := json.Unmarshal(input, &req); err != nil {
		return nil, fmt.Errorf("parse input: %w", err)
	}

	// Validate URL
	if _, err := url.ParseRequestURI(req.URL); err != nil {
		return &engine.ToolResult{
			Content: []engine.ContentBlock{&engine.TextBlock{Text: fmt.Sprintf("Error: Invalid URL \"%s\"", req.URL)}},
			IsError: true,
		}, nil
	}

	start := time.Now()

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "GET", req.URL, nil)
	if err != nil {
		return &engine.ToolResult{
			Content: []engine.ContentBlock{&engine.TextBlock{Text: fmt.Sprintf("Error creating request: %v", err)}},
			IsError: true,
		}, nil
	}

	httpReq.Header.Set("User-Agent", "Claude Code/Go")
	httpReq.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	client := &http.Client{
		Timeout: 30 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		return &engine.ToolResult{
			Content: []engine.ContentBlock{&engine.TextBlock{Text: fmt.Sprintf("Error fetching URL: %v", err)}},
			IsError: true,
		}, nil
	}
	defer resp.Body.Close()

	duration := time.Since(start).Milliseconds()

	// Handle redirects
	if resp.StatusCode >= 300 && resp.StatusCode < 400 {
		redirectURL := resp.Header.Get("Location")
		if redirectURL == "" {
			redirectURL = "unknown"
		}
		return &engine.ToolResult{
			Content: []engine.ContentBlock{&engine.TextBlock{
				Text: fmt.Sprintf("REDIRECT DETECTED: The URL redirects to %s\n\nOriginal URL: %s\nRedirect URL: %s\n\nTo complete your request, I need to fetch content from the redirected URL. Please use WebFetch again with these parameters:\n- url: \"%s\"\n- prompt: \"%s\"",
					resp.Status, req.URL, redirectURL, redirectURL, req.Prompt),
			}},
			IsError: false,
		}, nil
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &engine.ToolResult{
			Content: []engine.ContentBlock{&engine.TextBlock{Text: fmt.Sprintf("Error reading response: %v", err)}},
			IsError: true,
		}, nil
	}

	content := string(body)

	// Apply prompt to content if provided
	result := content
	if req.Prompt != "" {
		result = applyPrompt(req.Prompt, content)
	}

	// Truncate if too long (100K chars limit)
	if len(result) > 100000 {
		result = result[:100000] + "\n\n[Output truncated]"
	}

	webResult := WebFetchResult{
		Bytes:      len(body),
		Code:       resp.StatusCode,
		CodeText:   resp.Status,
		Result:     result,
		DurationMs: duration,
		URL:        req.URL,
	}

	resultJSON, _ := json.Marshal(webResult)

	return &engine.ToolResult{
		Content: []engine.ContentBlock{&engine.TextBlock{Text: string(resultJSON)}},
	}, nil
}

func applyPrompt(prompt string, content string) string {
	// Simple prompt application - in production this would use an LLM
	// For now, just return the content with a note about the prompt
	if prompt == "" {
		return content
	}
	return fmt.Sprintf("[Applied prompt: %s]\n\n%s", prompt, content)
}

// WebSearchTool searches the web
type WebSearchTool struct{}

func (w *WebSearchTool) Name() string { return "WebSearch" }

func (w *WebSearchTool) Description() string {
	return "Search the web for current information"
}

func (w *WebSearchTool) InputSchema() schema.Schema {
	return schema.Object{
		Properties: map[string]schema.Schema{
			"query":           schema.String{},
			"allowed_domains": schema.Array{Items: schema.String{}},
			"blocked_domains": schema.Array{Items: schema.String{}},
		},
		Required: []string{"query"},
	}
}

func (w *WebSearchTool) Permission() engine.PermissionMode {
	return engine.PermissionReadonly
}

type SearchResult struct {
	Title string `json:"title"`
	URL   string `json:"url"`
}

type WebSearchOutput struct {
	Query           string         `json:"query"`
	Results         []SearchResult `json:"results"`
	DurationSeconds float64        `json:"durationSeconds"`
}

func (w *WebSearchTool) Execute(ctx context.Context, input json.RawMessage, execCtx engine.ToolExecContext) (*engine.ToolResult, error) {
	var req struct {
		Query          string   `json:"query"`
		AllowedDomains []string `json:"allowed_domains"`
		BlockedDomains []string `json:"blocked_domains"`
	}

	if err := json.Unmarshal(input, &req); err != nil {
		return nil, fmt.Errorf("parse input: %w", err)
	}

	if len(req.Query) < 2 {
		return &engine.ToolResult{
			Content: []engine.ContentBlock{&engine.TextBlock{Text: "Error: Missing query"}},
			IsError: true,
		}, nil
	}

	start := time.Now()

	// Use DuckDuckGo as a simple search API
	searchURL := fmt.Sprintf("https://duckduckgo.com/?q=%s&format=json", url.QueryEscape(req.Query))

	httpReq, _ := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
	httpReq.Header.Set("User-Agent", "Mozilla/5.0 (compatible; Claude Code Bot)")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return &engine.ToolResult{
			Content: []engine.ContentBlock{&engine.TextBlock{Text: fmt.Sprintf("Web search error: %v", err)}},
			IsError: true,
		}, nil
	}
	defer resp.Body.Close()

	duration := time.Since(start).Seconds()

	// Parse response
	body, _ := io.ReadAll(resp.Body)
	content := string(body)

	// Extract URLs from DuckDuckGo HTML response
	results := extractSearchResults(content, req.Query)

	output := WebSearchOutput{
		Query:           req.Query,
		Results:         results,
		DurationSeconds: duration,
	}

	outputJSON, _ := json.Marshal(output)

	resultText := fmt.Sprintf("Web search results for query: \"%s\"\n\n", req.Query)
	for _, r := range results {
		resultText += fmt.Sprintf("Links: [{\"title\": \"%s\", \"url\": \"%s\"}]\n\n", r.Title, r.URL)
	}
	resultText += "\nREMINDER: You MUST include the sources above in your response to the user using markdown hyperlinks."

	return &engine.ToolResult{
		Content: []engine.ContentBlock{&engine.TextBlock{Text: resultText}, &engine.TextBlock{Text: string(outputJSON)}},
	}, nil
}

func extractSearchResults(html string, query string) []SearchResult {
	results := []SearchResult{}

	// Simple extraction - look for URL patterns
	lines := strings.Split(html, "\n")
	for _, line := range lines {
		if strings.Contains(line, "uddg=") || strings.Contains(line, "href=\"http") {
			// Extract URL
			start := strings.Index(line, "href=\"")
			if start != -1 {
				start += 6
				end := strings.Index(line[start:], "\"")
				if end != -1 {
					urlStr := line[start : start+end]
					if strings.HasPrefix(urlStr, "http") {
						results = append(results, SearchResult{
							Title: urlStr,
							URL:   urlStr,
						})
					}
				}
			}
		}
		if len(results) >= 10 {
			break
		}
	}

	return results
}

// TodoWriteTool manages todo list
type TodoWriteTool struct{}

func (t *TodoWriteTool) Name() string { return "TodoWrite" }

func (t *TodoWriteTool) Description() string {
	return "Manage the session task checklist"
}

func (t *TodoWriteTool) InputSchema() schema.Schema {
	return schema.Object{
		Properties: map[string]schema.Schema{
			"todos": schema.Array{
				Items: schema.Object{
					Properties: map[string]schema.Schema{
						"content":    schema.String{},
						"status":     schema.String{},
						"activeForm": schema.String{},
					},
					Required: []string{"content", "status"},
				},
			},
		},
		Required: []string{"todos"},
	}
}

func (t *TodoWriteTool) Permission() engine.PermissionMode {
	return engine.PermissionNormal
}

type TodoOutput struct {
	OldTodos                []engine.TodoItem `json:"oldTodos"`
	NewTodos                []engine.TodoItem `json:"newTodos"`
	VerificationNudgeNeeded bool              `json:"verificationNudgeNeeded,omitempty"`
}

func (t *TodoWriteTool) Execute(ctx context.Context, input json.RawMessage, execCtx engine.ToolExecContext) (*engine.ToolResult, error) {
	var req struct {
		Todos []engine.TodoItem `json:"todos"`
	}

	if err := json.Unmarshal(input, &req); err != nil {
		return nil, fmt.Errorf("parse input: %w", err)
	}

	// Get old todos from context
	oldTodos := execCtx.GetTodos()

	// Check if all are done
	allDone := true
	for _, todo := range req.Todos {
		if todo.Status != "completed" {
			allDone = false
			break
		}
	}

	newTodos := req.Todos
	if allDone {
		newTodos = []engine.TodoItem{} // Clear todos when all are done
	}

	// Update todos in context
	execCtx.SetTodos(newTodos)

	// Check for verification nudge (3+ items completed without verification)
	verificationNudge := false
	if allDone && len(req.Todos) >= 3 {
		hasVerification := false
		for _, todo := range req.Todos {
			if strings.Contains(strings.ToLower(todo.Content), "verif") {
				hasVerification = true
				break
			}
		}
		verificationNudge = !hasVerification
	}

	output := TodoOutput{
		OldTodos:                oldTodos,
		NewTodos:                newTodos,
		VerificationNudgeNeeded: verificationNudge,
	}

	outputJSON, _ := json.Marshal(output)

	baseText := "Todos have been modified successfully. Ensure that you continue to the todo list to track your current tasks if applicable"
	if verificationNudge {
		baseText += "\n\nNOTE: You just closed out 3+ tasks and none of them was a verification step."
	}

	return &engine.ToolResult{
		Content: []engine.ContentBlock{&engine.TextBlock{Text: baseText}, &engine.TextBlock{Text: string(outputJSON)}},
	}, nil
}

// FileEditTool edits files using a simple replacement approach
type FileEditTool struct{}

func (f *FileEditTool) Name() string { return "Edit" }

func (f *FileEditTool) Description() string {
	return "Make edits to a file by replacing specific content"
}

func (f *FileEditTool) InputSchema() schema.Schema {
	return schema.Object{
		Properties: map[string]schema.Schema{
			"file_path":  schema.String{},
			"old_string": schema.String{},
			"new_string": schema.String{},
		},
		Required: []string{"file_path", "old_string", "new_string"},
	}
}

func (f *FileEditTool) Permission() engine.PermissionMode {
	return engine.PermissionElevated
}

func (f *FileEditTool) Execute(ctx context.Context, input json.RawMessage, execCtx engine.ToolExecContext) (*engine.ToolResult, error) {
	var req struct {
		FilePath  string `json:"file_path"`
		OldString string `json:"old_string"`
		NewString string `json:"new_string"`
	}

	if err := json.Unmarshal(input, &req); err != nil {
		return &engine.ToolResult{
			Content: []engine.ContentBlock{&engine.TextBlock{Text: fmt.Sprintf("Error parsing input: %v", err)}},
			IsError: true,
		}, nil
	}

	if req.FilePath == "" || req.OldString == "" {
		return &engine.ToolResult{
			Content: []engine.ContentBlock{&engine.TextBlock{Text: "Error: file_path and old_string are required"}},
			IsError: true,
		}, nil
	}

	// Read the file
	content, err := os.ReadFile(req.FilePath)
	if err != nil {
		return &engine.ToolResult{
			Content: []engine.ContentBlock{&engine.TextBlock{Text: fmt.Sprintf("Error reading file: %v", err)}},
			IsError: true,
		}, nil
	}

	oldContent := string(content)

	// Check if old_string exists
	if !strings.Contains(oldContent, req.OldString) {
		return &engine.ToolResult{
			Content: []engine.ContentBlock{&engine.TextBlock{Text: fmt.Sprintf("Error: old_string not found in file. Make sure to match the exact text including any whitespace.")}},
			IsError: true,
		}, nil
	}

	// Replace
	newContent := strings.Replace(oldContent, req.OldString, req.NewString, 1)

	// Write back
	if err := os.WriteFile(req.FilePath, []byte(newContent), 0644); err != nil {
		return &engine.ToolResult{
			Content: []engine.ContentBlock{&engine.TextBlock{Text: fmt.Sprintf("Error writing file: %v", err)}},
			IsError: true,
		}, nil
	}

	return &engine.ToolResult{
		Content: []engine.ContentBlock{&engine.TextBlock{Text: fmt.Sprintf("Applied edit to %s", req.FilePath)}},
	}, nil
}

func GetAllTools() []engine.Tool {
	tools := []engine.Tool{
		&BashTool{},
		&ReadTool{},
		&GlobTool{},
		&GrepTool{},
		&WriteTool{},
		&WebFetchTool{},
		&WebSearchTool{},
		&TodoWriteTool{},
		&FileEditTool{},
		&TaskCreateTool{},
		&TaskListTool{},
		&TaskUpdateTool{},
		&TaskGetTool{},
		&TaskStopTool{},
		&AgentTool{},
		&AgentResultTool{},
		&SendMessageTool{},
		&SleepTool{},
		&BriefTool{},
		&ConfigTool{},
		&TeamCreateTool{},
		&TeamDeleteTool{},
		&ToolSearchTool{},
		&SyntheticOutputTool{},
		&RemoteTriggerTool{},
		&MCPTool{},
		&ListMcpResourcesTool{},
		&ReadMcpResourceTool{},
		&AskUserQuestionTool{},
		&EnterPlanModeTool{},
		&ExitPlanModeTool{},
		&EnterWorktreeTool{},
		&ExitWorktreeTool{},
		&SkillTool{},
		&NotebookEditTool{},
		&PowerShellTool{},
		&REPLTool{},
		&ScheduleCronTool{},
		&McpAuthTool{},
		&TaskOutputTool{},
		&LSPTool{},
	}
	return tools
}

func GetExtendedTools() []engine.Tool {
	return GetAllTools()
}
