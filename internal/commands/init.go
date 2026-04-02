package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/claude-code-go/claude/internal/engine"
)

var initStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("cyan")).Bold(true)
var successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("green"))
var warningStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("yellow"))

const CLAUDE_MD_HEADER = `# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.
`

const OLD_INIT_PROMPT = `Please analyze this codebase and create a CLAUDE.md file, which will be given to future instances of Claude Code to operate in this repository.

What to add:
1. Commands that will be commonly used, such as how to build, lint, and run tests. Include the necessary commands to develop in this codebase, such as how to run a single test.
2. High-level code architecture and structure so that future instances can be productive more quickly. Focus on the "big picture" architecture that requires reading multiple files to understand.

Usage notes:
- If there's already a CLAUDE.md, suggest improvements to it.
- When you make the initial CLAUDE.md, do not repeat yourself and do not include obvious instructions.
- Avoid listing every component or file structure that can be easily discovered.
- Don't include generic development practices.
- If there are Cursor rules (in .cursor/rules/ or .cursorrules) or Copilot rules (in .github/copilot-instructions.md), make sure to include the important parts.
- If there is a README.md, make sure to include the important parts.
- Do not make up information such as "Common Development Tasks", "Tips for Development", "Support and Documentation" unless this is expressly included in other files that you read.
`

// InitCommand initializes a new project with CLAUDE.md
type InitCommand struct {
	BaseCommand
}

func NewInitCommand() *InitCommand {
	return &InitCommand{
		BaseCommand: *newPromptCommand("init", "Initialize a new CLAUDE.md file with codebase documentation"),
	}
}

func (c *InitCommand) GetAllowedTools() []string {
	return []string{
		"Glob(*)",
		"Grep(*)",
		"Read(*)",
		"Bash(*)",
	}
}

func (c *InitCommand) GetPromptTemplate() string {
	return OLD_INIT_PROMPT
}

func (c *InitCommand) Execute(ctx context.Context, args []string, execCtx engine.CommandContext) error {
	fmt.Println(initStyle.Render("Claude Code Project Initialization"))
	fmt.Println()

	// Check if CLAUDE.md already exists
	claudeMdPath := filepath.Join(execCtx.Cwd, "CLAUDE.md")

	if _, err := os.Stat(claudeMdPath); err == nil {
		fmt.Println(warningStyle.Render("CLAUDE.md already exists. Offering to improve it."))
		return c.improveExistingCLAUDEMD(ctx, claudeMdPath, execCtx)
	}

	fmt.Println("Creating new CLAUDE.md...")
	return c.createNewCLAUDEMD(ctx, claudeMdPath, execCtx)
}

func (c *InitCommand) createNewCLAUDEMD(ctx context.Context, path string, execCtx engine.CommandContext) error {
	// Detect project type
	projectInfo := detectProjectInfo(execCtx.Cwd)

	fmt.Printf("Detected: %s project\n", projectInfo.Language)
	if projectInfo.BuildCommand != "" {
		fmt.Printf("Build: %s\n", projectInfo.BuildCommand)
	}
	if projectInfo.TestCommand != "" {
		fmt.Printf("Test: %s\n", projectInfo.TestCommand)
	}

	// Create basic CLAUDE.md
	content := CLAUDE_MD_HEADER + "\n"
	content += "## Project Overview\n\n"
	content += fmt.Sprintf("- **Language**: %s\n", projectInfo.Language)
	if projectInfo.PackageManager != "" {
		content += fmt.Sprintf("- **Package Manager**: %s\n", projectInfo.PackageManager)
	}
	content += "\n## Commands\n\n"
	if projectInfo.BuildCommand != "" {
		content += fmt.Sprintf("- **Build**: `%s`\n", projectInfo.BuildCommand)
	}
	if projectInfo.TestCommand != "" {
		content += fmt.Sprintf("- **Test**: `%s`\n", projectInfo.TestCommand)
	}
	if projectInfo.LintCommand != "" {
		content += fmt.Sprintf("- **Lint**: `%s`\n", projectInfo.LintCommand)
	}
	content += "\n## Project Structure\n\n"
	content += "<!-- Add project structure overview here -->\n"

	// Check for existing docs to include
	readmePath := filepath.Join(execCtx.Cwd, "README.md")
	if _, err := os.Stat(readmePath); err == nil {
		content += "\n## From README.md\n\n"
		content += "<!-- Consider including key information from README.md -->\n"
	}

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write CLAUDE.md: %w", err)
	}

	fmt.Println(successStyle.Render("✓ Created CLAUDE.md"))
	fmt.Println()
	fmt.Println("You can now run Claude Code and it will use this file for context.")

	return nil
}

func (c *InitCommand) improveExistingCLAUDEMD(ctx context.Context, path string, execCtx engine.CommandContext) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read CLAUDE.md: %w", err)
	}

	fmt.Println("Current CLAUDE.md content:")
	fmt.Println("---")
	fmt.Println(string(content))
	fmt.Println("---")
	fmt.Println()
	fmt.Println("The LLM will analyze this file and suggest improvements when you run claude with a task.")
	fmt.Println("Consider asking: 'improve the CLAUDE.md file'")

	return nil
}

type ProjectInfo struct {
	Language       string
	PackageManager string
	BuildCommand   string
	TestCommand    string
	LintCommand    string
	OtherCommands  []string
}

func detectProjectInfo(dir string) ProjectInfo {
	info := ProjectInfo{}

	// Detect language/package manager
	if _, err := os.Stat(filepath.Join(dir, "package.json")); err == nil {
		info.Language = "JavaScript/TypeScript"
		info.PackageManager = "npm"
		info.BuildCommand = "npm run build"
		info.TestCommand = "npm test"
		info.LintCommand = "npm run lint"
	} else if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
		info.Language = "Go"
		info.PackageManager = "go"
		info.BuildCommand = "go build ./..."
		info.TestCommand = "go test ./..."
		info.LintCommand = "go vet ./..."
	} else if _, err := os.Stat(filepath.Join(dir, "Cargo.toml")); err == nil {
		info.Language = "Rust"
		info.PackageManager = "cargo"
		info.BuildCommand = "cargo build"
		info.TestCommand = "cargo test"
		info.LintCommand = "cargo clippy"
	} else if _, err := os.Stat(filepath.Join(dir, "pyproject.toml")); err == nil {
		info.Language = "Python"
		info.PackageManager = "pip/poetry"
		info.BuildCommand = "python -m build"
		info.TestCommand = "pytest"
		info.LintCommand = "ruff check ."
	} else if _, err := os.Stat(filepath.Join(dir, "pom.xml")); err == nil {
		info.Language = "Java/Kotlin"
		info.PackageManager = "Maven"
		info.BuildCommand = "mvn compile"
		info.TestCommand = "mvn test"
	} else if _, err := os.Stat(filepath.Join(dir, "build.gradle")); err == nil {
		info.Language = "Java/Groovy"
		info.PackageManager = "Gradle"
		info.BuildCommand = "gradle build"
		info.TestCommand = "gradle test"
	} else {
		// Try to detect from files
		if hasFiles(dir, ".go") {
			info.Language = "Go"
		} else if hasFiles(dir, ".ts", ".tsx", ".js", ".jsx") {
			info.Language = "JavaScript/TypeScript"
		} else if hasFiles(dir, ".rs") {
			info.Language = "Rust"
		} else if hasFiles(dir, ".py") {
			info.Language = "Python"
		} else {
			info.Language = "Unknown"
		}
	}

	return info
}

func hasFiles(dir string, extensions ...string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		for _, ext := range extensions {
			if strings.HasSuffix(name, ext) {
				return true
			}
		}
	}
	return false
}
