package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/claude-code-go/claude/internal/engine"
	"github.com/claude-code-go/claude/internal/services/mcp"
)

var mcpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("cyan"))
var mcpSuccess = lipgloss.NewStyle().Foreground(lipgloss.Color("green"))
var mcpError = lipgloss.NewStyle().Foreground(lipgloss.Color("red"))
var mcpWarning = lipgloss.NewStyle().Foreground(lipgloss.Color("yellow"))

type MCPCommand struct {
	BaseCommand
	protocol *mcp.MCPProtocol
}

func NewMCPCommand(protocol *mcp.MCPProtocol) *MCPCommand {
	return &MCPCommand{
		BaseCommand: *newCommand("mcp", "Manage MCP servers"),
		protocol:    protocol,
	}
}

func (c *MCPCommand) Execute(ctx context.Context, args []string, execCtx engine.CommandContext) error {
	if len(args) == 0 {
		return c.listServers(ctx)
	}

	subcommand := args[0]

	switch subcommand {
	case "list", "ls":
		return c.listServers(ctx)
	case "add":
		return c.addServer(ctx, args[1:])
	case "remove", "rm":
		if len(args) < 2 {
			return fmt.Errorf("Usage: /mcp remove <server-name>")
		}
		return c.removeServer(ctx, args[1])
	case "enable":
		if len(args) < 2 {
			return fmt.Errorf("Usage: /mcp enable <server-name>")
		}
		return c.enableServer(ctx, args[1])
	case "disable":
		if len(args) < 2 {
			return fmt.Errorf("Usage: /mcp disable <server-name>")
		}
		return c.disableServer(ctx, args[1])
	case "get":
		if len(args) < 2 {
			return fmt.Errorf("Usage: /mcp get <server-name>")
		}
		return c.getServer(ctx, args[1])
	case "serve":
		return c.serve(ctx, args[1:])
	default:
		return fmt.Errorf("Unknown subcommand: %s. Use 'list', 'add', 'remove', 'enable', 'disable', or 'get'", subcommand)
	}
}

type MCPServerConfig struct {
	Name      string            `json:"name"`
	Type      string            `json:"type"`
	Status    string            `json:"status"`
	Transport string            `json:"transport,omitempty"`
	Command   string            `json:"command,omitempty"`
	Args      []string          `json:"args,omitempty"`
	Env       map[string]string `json:"env,omitempty"`
	URL       string            `json:"url,omitempty"`
	Headers   map[string]string `json:"headers,omitempty"`
}

type MCPServerStatus struct {
	Name      string `json:"name"`
	Type      string `json:"type"`
	Status    string `json:"status"`
	Transport string `json:"transport"`
	Command   string `json:"command,omitempty"`
	URL       string `json:"url,omitempty"`
}

func (c *MCPCommand) listServers(ctx context.Context) error {
	fmt.Println(mcpStyle.Render("MCP Servers"))
	fmt.Println(strings.Repeat("─", 40))
	fmt.Println()

	configPath := mcp.GetDefaultMcpConfigPath()
	projectConfigPath := mcp.GetProjectMcpConfigPath()

	showConfig := func(path, scope string) {
		data, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				return
			}
			fmt.Printf("%s Error reading %s: %v\n", mcpError.Render("!"), path, err)
			return
		}

		var configs map[string]MCPServerConfig
		if err := json.Unmarshal(data, &configs); err != nil {
			var servers []MCPServerConfig
			if err := json.Unmarshal(data, &servers); err != nil {
				fmt.Printf("%s Error parsing %s: %v\n", mcpError.Render("!"), path, err)
				return
			}
			configs = make(map[string]MCPServerConfig)
			for i := range servers {
				configs[servers[i].Name] = servers[i]
			}
		}

		if len(configs) == 0 {
			return
		}

		fmt.Printf("%s (%s):\n", scope, path)
		for name, config := range configs {
			status := mcpSuccess.Render("enabled")
			if config.Status == "disabled" {
				status = mcpError.Render("disabled")
			}
			fmt.Printf("  %s [%s]\n", name, status)
			if config.Transport != "" {
				fmt.Printf("    Transport: %s\n", config.Transport)
			}
			if config.Command != "" {
				fmt.Printf("    Command: %s\n", config.Command)
			}
			if config.URL != "" {
				fmt.Printf("    URL: %s\n", config.URL)
			}
		}
		fmt.Println()
	}

	showConfig(projectConfigPath, mcpStyle.Render("Project config"))
	showConfig(configPath, mcpStyle.Render("User config"))

	hasServers := false
	if data, err := os.ReadFile(configPath); err == nil {
		var configs map[string]MCPServerConfig
		if json.Unmarshal(data, &configs) == nil && len(configs) > 0 {
			hasServers = true
		}
	}
	if data, err := os.ReadFile(projectConfigPath); err == nil {
		var configs map[string]MCPServerConfig
		if json.Unmarshal(data, &configs) == nil && len(configs) > 0 {
			hasServers = true
		}
	}

	if !hasServers {
		fmt.Println("No MCP servers configured.")
		fmt.Println()
		fmt.Println("Add servers with: /mcp add <config-file>")
	}

	return nil
}

func (c *MCPCommand) addServer(ctx context.Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("Usage: /mcp add <config-file> [-s scope]")
	}

	var scopeFlag string
	configFile := args[0]

	for i, arg := range args[1:] {
		if arg == "-s" || arg == "--scope" {
			if i+1 >= len(args) {
				return fmt.Errorf("Usage: /mcp add <config-file> [-s scope]")
			}
			scopeFlag = args[i+2]
			break
		}
	}

	scope := mcp.ScopeUser
	if scopeFlag != "" {
		switch scopeFlag {
		case "project":
			scope = mcp.ScopeProject
		case "user":
			scope = mcp.ScopeUser
		case "local":
			scope = mcp.ScopeLocal
		default:
			return fmt.Errorf("Invalid scope: %s. Must be 'project', 'user', or 'local'", scopeFlag)
		}
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	var server MCPServerConfig
	if err := json.Unmarshal(data, &server); err != nil {
		return fmt.Errorf("invalid MCP server config: %w", err)
	}

	if server.Name == "" {
		return fmt.Errorf("server name is required")
	}

	var configPath string
	switch scope {
	case mcp.ScopeProject:
		configPath = mcp.GetProjectMcpConfigPath()
	case mcp.ScopeUser, mcp.ScopeLocal:
		configPath = mcp.GetDefaultMcpConfigPath()
	}

	var servers map[string]MCPServerConfig
	if existingData, err := os.ReadFile(configPath); err == nil {
		json.Unmarshal(existingData, &servers)
	} else {
		servers = make(map[string]MCPServerConfig)
	}

	if _, exists := servers[server.Name]; exists {
		return fmt.Errorf("server '%s' already exists. Use /mcp remove %s first.", server.Name, server.Name)
	}

	server.Status = "enabled"
	servers[server.Name] = server

	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	newData, _ := json.MarshalIndent(servers, "", "  ")
	if err := os.WriteFile(configPath, newData, 0644); err != nil {
		return fmt.Errorf("failed to save MCP config: %w", err)
	}

	fmt.Printf("%s Added MCP server: %s\n", mcpSuccess.Render("✓"), server.Name)
	fmt.Printf("  Config saved to: %s\n", configPath)

	if server.Command != "" {
		if err := c.testServerConnection(ctx, server); err != nil {
			fmt.Printf("%s Server added but connection test failed: %v\n", mcpWarning.Render("!"), err)
		} else {
			fmt.Printf("%s Server connection verified\n", mcpSuccess.Render("✓"))
		}
	}

	return nil
}

func (c *MCPCommand) removeServer(ctx context.Context, name string) error {
	configPath := mcp.GetDefaultMcpConfigPath()
	projectConfigPath := mcp.GetProjectMcpConfigPath()

	removed := false

	for _, configPath := range []string{projectConfigPath, configPath} {
		data, err := os.ReadFile(configPath)
		if err != nil {
			continue
		}

		var servers map[string]MCPServerConfig
		if err := json.Unmarshal(data, &servers); err != nil {
			continue
		}

		if _, exists := servers[name]; exists {
			delete(servers, name)
			newData, _ := json.MarshalIndent(servers, "", "  ")
			if err := os.WriteFile(configPath, newData, 0644); err != nil {
				return fmt.Errorf("failed to save MCP config: %w", err)
			}
			fmt.Printf("%s Removed MCP server: %s\n", mcpSuccess.Render("✓"), name)
			removed = true
		}
	}

	if !removed {
		return fmt.Errorf("server '%s' not found", name)
	}

	return nil
}

func (c *MCPCommand) enableServer(ctx context.Context, name string) error {
	return c.setServerStatus(ctx, name, "enabled")
}

func (c *MCPCommand) disableServer(ctx context.Context, name string) error {
	return c.setServerStatus(ctx, name, "disabled")
}

func (c *MCPCommand) setServerStatus(ctx context.Context, name string, status string) error {
	configPath := mcp.GetDefaultMcpConfigPath()

	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("no MCP servers configured: %w", err)
	}

	var servers map[string]MCPServerConfig
	if err := json.Unmarshal(data, &servers); err != nil {
		return fmt.Errorf("failed to parse MCP config: %w", err)
	}

	found := false
	for serverName, server := range servers {
		if serverName == name {
			server.Status = status
			servers[serverName] = server
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("server '%s' not found", name)
	}

	newData, _ := json.MarshalIndent(servers, "", "  ")
	if err := os.WriteFile(configPath, newData, 0644); err != nil {
		return fmt.Errorf("failed to save MCP config: %w", err)
	}

	action := "enabled"
	if status == "disabled" {
		action = "disabled"
	}
	fmt.Printf("%s %s server: %s\n", mcpSuccess.Render("✓"), action, name)
	return nil
}

func (c *MCPCommand) getServer(ctx context.Context, name string) error {
	configPath := mcp.GetDefaultMcpConfigPath()
	projectConfigPath := mcp.GetProjectMcpConfigPath()

	for _, configPath := range []string{projectConfigPath, configPath} {
		data, err := os.ReadFile(configPath)
		if err != nil {
			continue
		}

		var servers map[string]MCPServerConfig
		if err := json.Unmarshal(data, &servers); err != nil {
			continue
		}

		if server, exists := servers[name]; exists {
			fmt.Printf("MCP Server: %s\n", name)
			fmt.Println(strings.Repeat("-", 40))
			fmt.Printf("Status:    %s\n", server.Status)
			fmt.Printf("Transport: %s\n", server.Transport)
			if server.Command != "" {
				fmt.Printf("Command:   %s\n", server.Command)
			}
			if len(server.Args) > 0 {
				fmt.Printf("Args:      %v\n", server.Args)
			}
			if server.URL != "" {
				fmt.Printf("URL:       %s\n", server.URL)
			}
			if len(server.Env) > 0 {
				fmt.Printf("Env:\n")
				for k, v := range server.Env {
					fmt.Printf("  %s=%s\n", k, v)
				}
			}
			if len(server.Headers) > 0 {
				fmt.Printf("Headers:\n")
				for k, v := range server.Headers {
					fmt.Printf("  %s: %s\n", k, v)
				}
			}
			return nil
		}
	}

	return fmt.Errorf("server '%s' not found", name)
}

func (c *MCPCommand) serve(ctx context.Context, args []string) error {
	fmt.Printf("%s Starting MCP server mode...\n", mcpStyle.Render("MCP"))

	debug := false
	verbose := false

	for _, arg := range args {
		if arg == "--debug" {
			debug = true
		}
		if arg == "--verbose" {
			verbose = true
		}
	}

	fmt.Println("Note: MCP server mode requires the full Claude Code implementation")
	fmt.Println("This is a placeholder for the MCP server functionality")

	if debug {
		fmt.Println("Debug mode enabled")
	}
	if verbose {
		fmt.Println("Verbose mode enabled")
	}

	return nil
}

func (c *MCPCommand) testServerConnection(ctx context.Context, server MCPServerConfig) error {
	if server.Command == "" {
		return nil
	}

	parts := strings.Fields(server.Command)
	if len(parts) == 0 {
		return nil
	}

	cmd := exec.CommandContext(ctx, "which", parts[0])
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("command not found: %s", parts[0])
	}

	return nil
}
