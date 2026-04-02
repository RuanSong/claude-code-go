package tools

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/claude-code-go/claude/internal/engine"
	"github.com/claude-code-go/claude/pkg/schema"
)

const (
	MaxFileReadLines = 2000
	MaxTokenEstimate = 120000
	TokensPerChar    = 0.25
)

type FileReadTool struct{}

func (f *FileReadTool) Name() string { return "Read" }

func (f *FileReadTool) Description() string {
	return `Reads a file from the local filesystem. You can access any file directly by using this tool.
Usage:
- The file_path parameter must be an absolute path, not a relative path
- By default, it reads up to 2000 lines starting from the beginning of the file
- You can optionally specify a line offset and limit to read specific portions
- Results are returned using cat -n format, with line numbers starting at 1
- This tool can read images (PNG, JPG, etc) and returns them visually
- This tool can only read files, not directories. To list a directory, use the Bash tool with ls.`
}

func (f *FileReadTool) InputSchema() schema.Schema {
	return schema.Object{
		Properties: map[string]schema.Schema{
			"file_path":  schema.String{},
			"offset":     schema.Integer{},
			"limit":      schema.Integer{},
			"show_lines": schema.Boolean{},
		},
		Required: []string{"file_path"},
	}
}

func (f *FileReadTool) Permission() engine.PermissionMode {
	return engine.PermissionReadonly
}

type FileReadOutput struct {
	Type       string `json:"type"`
	FilePath   string `json:"file_path"`
	Content    string `json:"content,omitempty"`
	NumLines   int    `json:"numLines"`
	StartLine  int    `json:"startLine"`
	TotalLines int    `json:"totalLines"`
}

func (f *FileReadTool) Execute(ctx context.Context, input json.RawMessage, execCtx *engine.ToolExecContext) (*engine.ToolResult, error) {
	var req struct {
		FilePath  string `json:"file_path"`
		Offset    *int   `json:"offset"`
		Limit     *int   `json:"limit"`
		ShowLines bool   `json:"show_lines"`
	}

	if err := json.Unmarshal(input, &req); err != nil {
		return &engine.ToolResult{
			Content: []engine.ContentBlock{&engine.TextBlock{Text: fmt.Sprintf("Error parsing input: %v", err)}},
			IsError: true,
		}, nil
	}

	if req.FilePath == "" {
		return &engine.ToolResult{
			Content: []engine.ContentBlock{&engine.TextBlock{Text: "Error: file_path is required"}},
			IsError: true,
		}, nil
	}

	offset := 0
	limit := MaxFileReadLines
	if req.Offset != nil {
		offset = *req.Offset
	}
	if req.Limit != nil {
		limit = *req.Limit
	}

	file, err := os.Open(req.FilePath)
	if err != nil {
		return &engine.ToolResult{
			Content: []engine.ContentBlock{&engine.TextBlock{Text: fmt.Sprintf("Error opening file: %v", err)}},
			IsError: true,
		}, nil
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return &engine.ToolResult{
			Content: []engine.ContentBlock{&engine.TextBlock{Text: fmt.Sprintf("Error getting file info: %v", err)}},
			IsError: true,
		}, nil
	}

	if info.IsDir() {
		return &engine.ToolResult{
			Content: []engine.ContentBlock{&engine.TextBlock{Text: "Error: cannot read a directory, use ls via Bash tool"}},
			IsError: true,
		}, nil
	}

	ext := strings.ToLower(filepath.Ext(req.FilePath))
	isImage := ext == ".png" || ext == ".jpg" || ext == ".jpeg" || ext == ".gif" || ext == ".webp"

	if isImage {
		data, err := os.ReadFile(req.FilePath)
		if err != nil {
			return &engine.ToolResult{
				Content: []engine.ContentBlock{&engine.TextBlock{Text: fmt.Sprintf("Error reading image: %v", err)}},
				IsError: true,
			}, nil
		}
		base64Data := base64.StdEncoding.EncodeToString(data)
		mimeType := "image/png"
		if ext == ".jpg" || ext == ".jpeg" {
			mimeType = "image/jpeg"
		} else if ext == ".gif" {
			mimeType = "image/gif"
		} else if ext == ".webp" {
			mimeType = "image/webp"
		}

		imageOutput := FileReadOutput{
			Type:     "image",
			FilePath: req.FilePath,
		}
		outputJSON, _ := json.Marshal(imageOutput)

		return &engine.ToolResult{
			Content: []engine.ContentBlock{
				&engine.TextBlock{Text: fmt.Sprintf("Image: %s (%d bytes, MIME: %s)\n```\n%s\n```", req.FilePath, len(data), mimeType, base64Data)},
				&engine.TextBlock{Text: string(outputJSON)},
			},
		}, nil
	}

	totalLines := 0
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		totalLines++
	}
	if err := scanner.Err(); err != nil {
		return &engine.ToolResult{
			Content: []engine.ContentBlock{&engine.TextBlock{Text: fmt.Sprintf("Error counting lines: %v", err)}},
			IsError: true,
		}, nil
	}

	if offset >= totalLines {
		return &engine.ToolResult{
			Content: []engine.ContentBlock{&engine.TextBlock{Text: fmt.Sprintf("Error: offset %d exceeds total lines %d", offset, totalLines)}},
			IsError: true,
		}, nil
	}

	file.Seek(0, 0)
	reader := bufio.NewReader(file)

	currentLine := 0
	var content strings.Builder
	startLine := offset + 1

	for {
		line, err := reader.ReadString('\n')
		if err != nil && err.Error() != "EOF" {
			break
		}
		currentLine++
		if currentLine <= offset {
			if err == nil {
				continue
			}
		}
		if currentLine > offset+limit {
			break
		}

		line = strings.TrimRight(line, "\r\n")
		if req.ShowLines {
			fmt.Fprintf(&content, "%6d\t%s\n", currentLine, line)
		} else {
			fmt.Fprintf(&content, "%s\n", line)
		}

		if err != nil {
			break
		}
	}

	numLines := currentLine - offset
	if numLines > limit {
		numLines = limit
	}

	readOutput := FileReadOutput{
		Type:       "text",
		FilePath:   req.FilePath,
		Content:    content.String(),
		NumLines:   numLines,
		StartLine:  startLine,
		TotalLines: totalLines,
	}
	outputJSON, _ := json.Marshal(readOutput)

	return &engine.ToolResult{
		Content: []engine.ContentBlock{
			&engine.TextBlock{Text: content.String()},
			&engine.TextBlock{Text: "\n" + string(outputJSON)},
		},
	}, nil
}

type FileWriteTool struct{}

func (f *FileWriteTool) Name() string { return "Write" }

func (f *FileWriteTool) Description() string {
	return `Writes content to a file, creating the file if it doesn't exist or overwriting it if it does.
Usage:
- The file_path must be an absolute path
- The content parameter contains the full content to write
- Parent directories will be created if they don't exist
- This will overwrite existing files entirely`
}

func (f *FileWriteTool) InputSchema() schema.Schema {
	return schema.Object{
		Properties: map[string]schema.Schema{
			"file_path": schema.String{},
			"content":   schema.String{},
		},
		Required: []string{"file_path", "content"},
	}
}

func (f *FileWriteTool) Permission() engine.PermissionMode {
	return engine.PermissionElevated
}

type FileWriteOutput struct {
	Type        string `json:"type"`
	FilePath    string `json:"file_path"`
	Bytes       int    `json:"bytes"`
	Lines       int    `json:"lines"`
	Created     bool   `json:"created"`
	Overwritten bool   `json:"overwritten"`
}

func (f *FileWriteTool) Execute(ctx context.Context, input json.RawMessage, execCtx *engine.ToolExecContext) (*engine.ToolResult, error) {
	var req struct {
		FilePath string `json:"file_path"`
		Content  string `json:"content"`
	}

	if err := json.Unmarshal(input, &req); err != nil {
		return &engine.ToolResult{
			Content: []engine.ContentBlock{&engine.TextBlock{Text: fmt.Sprintf("Error parsing input: %v", err)}},
			IsError: true,
		}, nil
	}

	if req.FilePath == "" {
		return &engine.ToolResult{
			Content: []engine.ContentBlock{&engine.TextBlock{Text: "Error: file_path is required"}},
			IsError: true,
		}, nil
	}

	dir := filepath.Dir(req.FilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return &engine.ToolResult{
			Content: []engine.ContentBlock{&engine.TextBlock{Text: fmt.Sprintf("Error creating directory: %v", err)}},
			IsError: true,
		}, nil
	}

	exists := true
	if _, err := os.Stat(req.FilePath); os.IsNotExist(err) {
		exists = false
	}

	if err := os.WriteFile(req.FilePath, []byte(req.Content), 0644); err != nil {
		return &engine.ToolResult{
			Content: []engine.ContentBlock{&engine.TextBlock{Text: fmt.Sprintf("Error writing file: %v", err)}},
			IsError: true,
		}, nil
	}

	lines := strings.Count(req.Content, "\n")
	if !strings.HasSuffix(req.Content, "\n") {
		lines++
	}

	created := !exists
	overwritten := exists

	writeOutput := FileWriteOutput{
		Type:        "text",
		FilePath:    req.FilePath,
		Bytes:       len(req.Content),
		Lines:       lines,
		Created:     created,
		Overwritten: overwritten,
	}
	outputJSON, _ := json.Marshal(writeOutput)

	msg := "File written successfully"
	if created {
		msg = fmt.Sprintf("File created successfully at %s", req.FilePath)
	} else {
		msg = fmt.Sprintf("File overwritten successfully at %s", req.FilePath)
	}

	return &engine.ToolResult{
		Content: []engine.ContentBlock{
			&engine.TextBlock{Text: msg},
			&engine.TextBlock{Text: string(outputJSON)},
		},
	}, nil
}

func GetFileTools() []engine.Tool {
	return []engine.Tool{
		&FileReadTool{},
		&FileWriteTool{},
	}
}

func detectSessionFileType(filePath string) string {
	homeDir, _ := os.UserHomeDir()
	configDir := filepath.Join(homeDir, ".claude")

	if !strings.HasPrefix(filePath, configDir) {
		return ""
	}

	normalizedPath := filepath.ToSlash(filePath)

	if strings.Contains(normalizedPath, "/session-memory/") && strings.HasSuffix(normalizedPath, ".md") {
		return "session_memory"
	}

	if strings.Contains(normalizedPath, "/projects/") && strings.HasSuffix(normalizedPath, ".jsonl") {
		return "session_transcript"
	}

	return ""
}

func countFileTokens(content string) int {
	return int(float64(len(content)) * TokensPerChar)
}

var devicePaths = map[string]bool{
	"/dev/zero":    true,
	"/dev/random":  true,
	"/dev/full":    true,
	"/dev/stdin":   true,
	"/dev/tty":     true,
	"/dev/console": true,
	"/dev/stdout":  true,
	"/dev/stderr":  true,
}

func isBlockedDevicePath(filePath string) bool {
	if devicePaths[filePath] {
		return true
	}

	if strings.HasPrefix(filePath, "/proc/") {
		if strings.HasSuffix(filePath, "/fd/0") ||
			strings.HasSuffix(filePath, "/fd/1") ||
			strings.HasSuffix(filePath, "/fd/2") {
			return true
		}
	}

	return false
}

func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		if homeDir, err := os.UserHomeDir(); err == nil {
			return filepath.Join(homeDir, path[2:])
		}
	}
	return path
}
