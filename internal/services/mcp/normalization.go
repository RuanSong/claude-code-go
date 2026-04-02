package mcp

import (
	"regexp"
	"strings"
	"unicode"
)

var nonAlphanumericRegex = regexp.MustCompile(`[^a-zA-Z0-9_]`)

func NormalizeNameForMCP(name string) string {
	result := nonAlphanumericRegex.ReplaceAllString(name, "_")
	result = strings.TrimLeft(result, "_")
	var sb strings.Builder
	for i, r := range result {
		if i == 0 || result[i-1] == '_' {
			sb.WriteRune(unicode.ToLower(r))
		} else {
			sb.WriteRune(r)
		}
	}
	return sb.String()
}

func BuildMcpToolName(serverName, toolName string) string {
	normalizedServer := NormalizeNameForMCP(serverName)
	return "mcp__" + normalizedServer + "__" + toolName
}

func ParseMcpToolName(toolName string) (serverName string, mcpToolName string, ok bool) {
	if !strings.HasPrefix(toolName, "mcp__") {
		return "", "", false
	}
	remainder := strings.TrimPrefix(toolName, "mcp__")
	parts := strings.SplitN(remainder, "__", 2)
	if len(parts) != 2 {
		return "", "", false
	}
	return parts[0], parts[1], true
}

func IsMcpToolName(toolName string) bool {
	return strings.HasPrefix(toolName, "mcp__")
}

func IsMcpCommandName(commandName string) bool {
	return strings.HasPrefix(commandName, "mcp__")
}

func FilterToolsByServer(tools []MCPTool, serverName string) []MCPTool {
	prefix := "mcp__" + NormalizeNameForMCP(serverName) + "__"
	var result []MCPTool
	for _, tool := range tools {
		if strings.HasPrefix(tool.Name, prefix) {
			result = append(result, tool)
		}
	}
	return result
}

func FilterCommandsByServer(commands []string, serverName string) []string {
	prefix := "mcp__" + NormalizeNameForMCP(serverName) + "__"
	var result []string
	for _, cmd := range commands {
		if strings.HasPrefix(cmd, prefix) {
			result = append(result, cmd)
		}
	}
	return result
}

func ExcludeToolsByServer(tools []MCPTool, serverName string) []MCPTool {
	prefix := "mcp__" + NormalizeNameForMCP(serverName) + "__"
	var result []MCPTool
	for _, tool := range tools {
		if !strings.HasPrefix(tool.Name, prefix) {
			result = append(result, tool)
		}
	}
	return result
}

func ExcludeCommandsByServer(commands []string, serverName string) []string {
	prefix := "mcp__" + NormalizeNameForMCP(serverName) + "__"
	var result []string
	for _, cmd := range commands {
		if !strings.HasPrefix(cmd, prefix) {
			result = append(result, cmd)
		}
	}
	return result
}

type MCPInfo struct {
	ServerName   string
	ToolName     string
	OriginalName string
}

func MCPInfoFromString(name string) *MCPInfo {
	if !strings.HasPrefix(name, "mcp__") {
		return nil
	}
	remainder := strings.TrimPrefix(name, "mcp__")
	parts := strings.SplitN(remainder, "__", 2)
	if len(parts) != 2 {
		return nil
	}
	return &MCPInfo{
		ServerName:   parts[0],
		ToolName:     parts[1],
		OriginalName: name,
	}
}
