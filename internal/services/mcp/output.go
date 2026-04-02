package mcp

import (
	"encoding/json"
	"fmt"
	"strings"
)

const (
	DefaultMaxOutputChars   = 25000
	MaxMcpDescriptionLength = 2048
)

type OutputStorage struct {
	outputs map[string]string
}

func NewOutputStorage() *OutputStorage {
	return &OutputStorage{
		outputs: make(map[string]string),
	}
}

func (s *OutputStorage) Store(id string, content string) {
	s.outputs[id] = content
}

func (s *OutputStorage) Get(id string) (string, bool) {
	content, ok := s.outputs[id]
	return content, ok
}

func (s *OutputStorage) Delete(id string) {
	delete(s.outputs, id)
}

func GetContentSizeEstimate(content string) int {
	return len(content)
}

func MCPContentNeedsTruncation(content string, maxChars int) bool {
	return GetContentSizeEstimate(content) > maxChars
}

func TruncateMCPContentIfNeeded(content string, maxChars int) string {
	if !MCPContentNeedsTruncation(content, maxChars) {
		return content
	}
	return content[:maxChars] + "... [truncated]"
}

func ContentBlockToString(block ContentBlock) string {
	switch block.Type {
	case "text":
		return block.Text
	case "image":
		return fmt.Sprintf("[Image: %s]", block.Data)
	case "resource":
		return fmt.Sprintf("[Resource: %s]", block.URI)
	default:
		data, _ := json.Marshal(block)
		return string(data)
	}
}

func FormatMCPToolResult(result *MCPToolResult) string {
	if result == nil {
		return ""
	}
	var sb strings.Builder
	for i, block := range result.Content {
		if i > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(ContentBlockToString(block))
	}
	return sb.String()
}

func PersistToolResult(content string) string {
	id := fmt.Sprintf("mcp_result_%d", len(content))
	return id
}

func GetLargeOutputInstructions() string {
	return "Output exceeds maximum size. Full result stored but truncated in display."
}

func GetFormatDescription(format string) string {
	switch format {
	case "text":
		return "Text output"
	case "json":
		return "JSON data"
	case "image":
		return "Image data"
	default:
		return "Unknown format"
	}
}
