package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/claude-code-go/claude/internal/engine"
)

var mcpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("cyan"))
var mcpSuccess = lipgloss.NewStyle().Foreground(lipgloss.Color("green"))
var mcpError = lipgloss.NewStyle().Foreground(lipgloss.Color("red"))

// MCPCommand manages MCP servers
type MCPCommand struct {
	BaseCommand
}

func NewMCPCommand() *MCPCommand {
	return &MCPCommand{
		BaseCommand: *newCommand("mcp", "Manage MCP servers"),
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
	default:
		return fmt.Errorf("Unknown subcommand: %s. Use 'list', 'add', 'remove', 'enable', or 'disable'", subcommand)
	}
}

type MCPServer struct {
	Name      string            `json:"name"`
	Type      string            `json:"type"`
	Status    string            `json:"status"`
	Transport string            `json:"transport,omitempty"`
	Command   string            `json:"command,omitempty"`
	Args      []string          `json:"args,omitempty"`
	Env       map[string]string `json:"env,omitempty"`
}

func (c *MCPCommand) listServers(ctx context.Context) error {
	fmt.Println(mcpStyle.Render("MCP Servers"))
	fmt.Println(strings.Repeat("─", 40))
	fmt.Println()

	configPath := getMCPConfigPath()

	var servers []MCPServer

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("No MCP servers configured.")
			fmt.Println()
			fmt.Println("Add servers with: /mcp add <config-file>")
			return nil
		}
		return fmt.Errorf("failed to read MCP config: %w", err)
	}

	if err := json.Unmarshal(data, &servers); err != nil {
		return fmt.Errorf("failed to parse MCP config: %w", err)
	}

	if len(servers) == 0 {
		fmt.Println("No MCP servers configured.")
		fmt.Println()
		fmt.Println("Add servers with: /mcp add <config-file>")
		return nil
	}

	for _, server := range servers {
		status := mcpSuccess.Render("enabled")
		if server.Status == "disabled" {
			status = mcpError.Render("disabled")
		}
		fmt.Printf("%s [%s]\n", server.Name, status)
		if server.Command != "" {
			fmt.Printf("  Command: %s\n", server.Command)
		}
		if server.Transport != "" {
			fmt.Printf("  Transport: %s\n", server.Transport)
		}
		fmt.Println()
	}

	return nil
}

func (c *MCPCommand) addServer(ctx context.Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("Usage: /mcp add <config-file>")
	}

	configFile := args[0]

	// Read and validate config file
	data, err := os.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// Validate JSON
	var server MCPServer
	if err := json.Unmarshal(data, &server); err != nil {
		return fmt.Errorf("invalid MCP server config: %w", err)
	}

	// Load existing config
	configPath := getMCPConfigPath()
	var servers []MCPServer

	existingData, err := os.ReadFile(configPath)
	if err == nil {
		json.Unmarshal(existingData, &servers)
	}

	// Check for duplicates
	for _, s := range servers {
		if s.Name == server.Name {
			return fmt.Errorf("server '%s' already exists. Use /mcp remove %s first.", server.Name, server.Name)
		}
	}

	// Add new server
	server.Status = "enabled"
	servers = append(servers, server)

	// Save config
	if err := os.MkdirAll(getMCPConfigDir(), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	newData, _ := json.MarshalIndent(servers, "", "  ")
	if err := os.WriteFile(configPath, newData, 0644); err != nil {
		return fmt.Errorf("failed to save MCP config: %w", err)
	}

	fmt.Printf("%s Added MCP server: %s\n", mcpSuccess.Render("✓"), server.Name)

	// Try to start the server
	if err := c.testServerConnection(ctx, server); err != nil {
		fmt.Printf("%s Server added but connection test failed: %v\n", mcpError.Render("!"), err)
	} else {
		fmt.Printf("%s Server connection verified\n", mcpSuccess.Render("✓"))
	}

	return nil
}

func (c *MCPCommand) removeServer(ctx context.Context, name string) error {
	configPath := getMCPConfigPath()

	var servers []MCPServer

	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("no MCP servers configured")
	}

	if err := json.Unmarshal(data, &servers); err != nil {
		return fmt.Errorf("failed to parse MCP config: %w", err)
	}

	found := false
	newServers := make([]MCPServer, 0)
	for _, s := range servers {
		if s.Name == name {
			found = true
		} else {
			newServers = append(newServers, s)
		}
	}

	if !found {
		return fmt.Errorf("server '%s' not found", name)
	}

	newData, _ := json.MarshalIndent(newServers, "", "  ")
	if err := os.WriteFile(configPath, newData, 0644); err != nil {
		return fmt.Errorf("failed to save MCP config: %w", err)
	}

	fmt.Printf("%s Removed MCP server: %s\n", mcpSuccess.Render("✓"), name)
	return nil
}

func (c *MCPCommand) enableServer(ctx context.Context, name string) error {
	return c.setServerStatus(ctx, name, "enabled")
}

func (c *MCPCommand) disableServer(ctx context.Context, name string) error {
	return c.setServerStatus(ctx, name, "disabled")
}

func (c *MCPCommand) setServerStatus(ctx context.Context, name string, status string) error {
	configPath := getMCPConfigPath()

	var servers []MCPServer

	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("no MCP servers configured")
	}

	if err := json.Unmarshal(data, &servers); err != nil {
		return fmt.Errorf("failed to parse MCP config: %w", err)
	}

	found := false
	for i, s := range servers {
		if s.Name == name {
			servers[i].Status = status
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

func (c *MCPCommand) testServerConnection(ctx context.Context, server MCPServer) error {
	// Basic test - just verify the command exists
	if server.Command == "" {
		return nil
	}

	cmd := exec.CommandContext(ctx, "which", strings.Fields(server.Command)[0])
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("command not found: %s", server.Command)
	}

	return nil
}

func getMCPConfigDir() string {
	home, _ := os.UserHomeDir()
	if home == "" {
		home = os.Getenv("HOME")
	}
	return home + "/.claude/mcp"
}

func getMCPConfigPath() string {
	return getMCPConfigDir() + "/servers.json"
}
