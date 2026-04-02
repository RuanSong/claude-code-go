# Claude Code (Go Implementation)

> A Go implementation of Claude Code, Anthropic's AI programming assistant

**[中文版](README_CN.md) | English**

## Overview

Claude Code Go is an open-source re-implementation of [Claude Code](https://github.com/anthropics/claude-code) in Go. It provides the same core functionality:

- **Interactive REPL Mode** - Chat with Claude in your terminal
- **Terminal UI (TUI)** - Rich Bubble Tea-based interface
- **Tool System** - Execute code, read/write files, search, and more
- **Slash Commands** - `/commit`, `/review`, `/help`, and 40+ other commands
- **MCP Protocol Support** - Connect to Model Context Protocol servers
- **Plugin & Skill System** - Extend functionality with plugins and skills

## Architecture

```
claude-code-go/
├── cmd/claude/           # CLI entry point
├── internal/
│   ├── cli/             # CLI commands (Cobra-based)
│   ├── engine/          # Core query engine
│   │   ├── query_engine.go   # Main query loop
│   │   ├── tool.go           # Tool interface & registry
│   │   ├── command.go        # Command interface & registry
│   │   ├── context.go        # Context management
│   │   └── permission.go     # Permission system
│   ├── tools/           # Tool implementations
│   │   ├── builtins.go       # Bash, Read, Write, Glob, Grep
│   │   ├── web.go           # WebFetch, WebSearch, Todo, Edit
│   │   ├── task.go          # Task tools
│   │   ├── agent.go         # Agent tools
│   │   ├── mcp.go           # MCP tools
│   │   └── additional.go     # Additional tools (Plan, Worktree, etc.)
│   ├── commands/        # Slash commands
│   ├── services/        # API clients & services
│   │   ├── analytics/      # Analytics & feature flags
│   │   ├── compact/        # Context compaction
│   │   ├── lsp/           # Language Server Protocol
│   │   ├── mcp/           # MCP protocol client
│   │   ├── oauth/         # OAuth 2.0 authentication
│   │   ├── plugins/       # Plugin system
│   │   ├── settings/       # Settings management
│   │   ├── skills/        # Skill management
│   │   ├── team/          # Team memory sync
│   │   ├── token/         # Token estimation
│   │   └── voice/         # Voice input/output
│   ├── ui/              # Bubble Tea TUI
│   ├── bridge/          # IDE bridge (VS Code, JetBrains)
│   ├── protocol/        # JSON-RPC protocol
│   └── storage/         # Session & config storage
└── pkg/
    └── anthropic/       # Anthropic API client
```

## Tech Stack

| Component | Technology |
|-----------|------------|
| Language | Go 1.21+ |
| CLI Framework | [Cobra](https://github.com/spf13/cobra) |
| TUI Framework | [Bubble Tea](https://github.com/charmbracelet/bubbletea) + Lipgloss |
| HTTP Client | net/http (standard library) |
| Config | [Viper](https://github.com/spf13/viper) |
| Logging | zap |

## Getting Started

### Prerequisites

- Go 1.21 or later
- Anthropic API key

### Installation

```bash
# Clone the repository
git clone https://github.com/yourusername/claude-code-go.git
cd claude-code-go

# Download dependencies
go mod download

# Build the CLI
go build -o claude ./cmd/claude
```

### Configuration

```bash
# Set API key via environment variable
export ANTHROPIC_API_KEY="your-api-key"

# Or use the login command
./claude login
```

## Usage

### Interactive Modes

```bash
# Start interactive REPL mode
./claude repl

# Start terminal UI mode
./claude tui
```

### Single Commands

```bash
# Run a single task
./claude "Write a hello world program in Go"
```

### Slash Commands

Slash commands are prefixed with `/` and provide specialized functionality:

#### Core Commands
| Command | Description |
|---------|-------------|
| `/help` | Show help information |
| `/model [name]` | Get or set the AI model |
| `/config get/set/list` | Manage configuration |
| `/cost` | Show session usage and cost |
| `/exit` | Exit Claude Code |

#### Git Commands
| Command | Description |
|---------|-------------|
| `/commit` | Create a git commit |
| `/diff` | Show changes |
| `/branch` | Manage branches |
| `/logs` | View git logs |

#### Code Commands
| Command | Description |
|---------|-------------|
| `/review` | Review code changes |
| `/ultrareview` | Comprehensive code review |
| `/compact` | Compact context to save tokens |

#### Development Commands
| Command | Description |
|---------|-------------|
| `/init` | Initialize a project |
| `/mcp` | Manage MCP servers |
| `/skills` | Manage skills |
| `/doctor` | Run diagnostics |

### CLI Flags

```bash
-v, --verbose         # Verbose output
--json              # Output JSON format
--config <path>     # Config file path
```

## Tools

Claude Code provides these tools for interacting with the system:

### Built-in Tools
| Tool | Permission | Description |
|------|------------|-------------|
| `Bash` | Elevated | Execute shell commands |
| `Read` | Readonly | Read file contents |
| `Write` | Elevated | Write content to files |
| `Glob` | Readonly | Find files by pattern |
| `Grep` | Readonly | Search file contents |

### Web Tools
| Tool | Permission | Description |
|------|------------|-------------|
| `WebFetch` | Readonly | Fetch URL content |
| `WebSearch` | Readonly | Search the web |
| `TodoWrite` | Normal | Manage todo list |
| `Edit` | Elevated | Edit files (replace text) |

### Task Tools
| Tool | Permission | Description |
|------|------------|-------------|
| `TaskCreate` | Normal | Create a task |
| `TaskList` | Readonly | List all tasks |
| `TaskUpdate` | Normal | Update task status |
| `TaskGet` | Readonly | Get task details |
| `TaskStop` | Normal | Stop a task |

### Additional Tools
| Tool | Permission | Description |
|------|------------|-------------|
| `AskUserQuestion` | Normal | Ask user a question |
| `EnterPlanMode` | Normal | Enter plan mode |
| `ExitPlanMode` | Normal | Exit plan mode |
| `EnterWorktree` | Elevated | Enter git worktree |
| `ExitWorktree` | Elevated | Exit git worktree |
| `NotebookEdit` | Elevated | Edit Jupyter notebooks |
| `Skill` | Normal | Execute a skill |
| `LSP` | Readonly | LSP operations |
| `McpAuth` | Elevated | MCP authentication |
| `PowerShell` | Elevated | Run PowerShell commands |

### Agent Tools
| Tool | Permission | Description |
|------|------------|-------------|
| `Agent` | Normal | Create sub-agent |
| `AgentResult` | Normal | Get agent result |
| `SendMessage` | Normal | Send message |
| `TeamCreate` | Normal | Create agent team |
| `TeamDelete` | Normal | Delete team |

## Services

The project includes these services:

| Service | Description |
|---------|-------------|
| `analytics` | Event tracking and feature flags |
| `compact` | Context compaction for long conversations |
| `lsp` | Language Server Protocol integration |
| `mcp` | Model Context Protocol client |
| `oauth` | OAuth 2.0 authentication |
| `plugins` | Plugin system for extensions |
| `settings` | Configuration management |
| `skills` | Skill management system |
| `team` | Team memory synchronization |
| `token` | Token estimation and cost calculation |
| `voice` | Voice input/output support |

## Testing

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test -v ./internal/services/token/...
```

## Project Status

### Implemented
- [x] CLI with Cobra framework
- [x] Tool interface and registry
- [x] Command interface and registry
- [x] Anthropic API client
- [x] 40+ built-in tools
- [x] Query engine with tool execution loop
- [x] Permission system
- [x] Context management
- [x] Bubble Tea TUI
- [x] 11 service modules with tests
- [x] Comprehensive test coverage

### In Development
- [ ] Full slash command suite
- [ ] Complete MCP protocol support
- [ ] Context compaction implementation
- [ ] Plugin system
- [ ] Skill system
- [ ] IDE bridge integration

## Contributing

Contributions are welcome! Please:

1. Read the architecture documentation
2. Follow the existing code style
3. Add tests for new functionality
4. Run `go vet` and `gofmt` before submitting

## License

This project is for educational purposes. The original Claude Code is copyrighted by Anthropic.

## References

- [Claude Code Documentation](https://docs.anthropic.com/claude-code)
- [Anthropic API](https://docs.anthropic.com)
- [Bubble Tea](https://github.com/charmbracelet/bubbletea)
- [Cobra](https://github.com/spf13/cobra)
