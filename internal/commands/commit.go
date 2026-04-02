package commands

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/claude-code-go/claude/internal/engine"
)

const commitAllowedTools = "Bash(git add:*),Bash(git status:*),Bash(git commit:*)"

// CommitCommand creates a git commit
type CommitCommand struct {
	BaseCommand
}

func NewCommitCommand() *CommitCommand {
	return &CommitCommand{
		BaseCommand: *newPromptCommand("commit", "Create a git commit"),
	}
}

func (c *CommitCommand) GetAllowedTools() []string {
	return []string{
		"Bash(git add:*)",
		"Bash(git status:*)",
		"Bash(git commit:*)",
	}
}

func (c *CommitCommand) GetPromptTemplate() string {
	return `## Context

- Current git status: !` + "`git status`" + `
- Current git diff (staged and unstaged changes): !` + "`git diff HEAD`" + `
- Current branch: !` + "`git branch --show-current`" + `
- Recent commits: !` + "`git log --oneline -10`" + `

## Git Safety Protocol

- NEVER update the git config
- NEVER skip hooks (--no-verify, --no-gpg-sign, etc) unless the user explicitly requests it
- CRITICAL: ALWAYS create NEW commits. NEVER use git commit --amend, unless the user explicitly requests it
- Do not commit files that likely contain secrets (.env, credentials.json, etc). Warn the user if they specifically request to commit those files
- If there are no changes to commit (i.e., no untracked files and no modifications), do not create an empty commit
- Never use git commands with the -i flag (like git rebase -i or git add -i) since they require interactive input which is not supported

## Your task

Based on the above changes, create a single git commit:

1. Analyze all staged changes and draft a commit message:
   - Look at the recent commits above to follow this Repository's commit message style
   - Summarize the nature of the changes (new feature, enhancement, bug fix, refactoring, test, docs, etc.)
   - Ensure the message accurately reflects the changes and their purpose (i.e. "add" means a wholly new feature, "update" means an enhancement to an existing feature, "fix" means a bug fix, etc.)
   - Draft a concise (1-2 sentences) commit message that focuses on the "why" rather than the "what"

2. Stage relevant files and create the commit using HEREDOC syntax:
` + "```" + `
git commit -m "$(cat <<'EOF'
Commit message here.
EOF
)"
` + "```" + `

You have the capability to call multiple tools in a single response. Stage and create the commit using a single message. Do not use any other tools or do anything else. Do not send any other text or messages besides these tool calls.`
}

func (c *CommitCommand) Execute(ctx context.Context, args []string, execCtx engine.CommandContext) error {
	// First check git status
	statusCmd := exec.CommandContext(ctx, "git", "status")
	statusOutput, err := statusCmd.Output()
	if err != nil {
		return fmt.Errorf("git status failed: %w", err)
	}

	// Check if there are changes
	statusStr := string(statusOutput)
	if strings.Contains(statusStr, "nothing to commit") {
		fmt.Println("No changes to commit")
		return nil
	}

	// Get staged diff
	diffCmd := exec.CommandContext(ctx, "git", "diff", "HEAD")
	diffOutput, err := diffCmd.Output()
	if err != nil {
		return fmt.Errorf("git diff failed: %w", err)
	}

	// Get current branch
	branchCmd := exec.CommandContext(ctx, "git", "branch", "--show-current")
	branchOutput, err := branchCmd.Output()
	if err != nil {
		return fmt.Errorf("git branch failed: %w", err)
	}

	// Get recent commits
	logCmd := exec.CommandContext(ctx, "git", "log", "--oneline", "-10")
	logOutput, err := logCmd.Output()
	if err != nil {
		return fmt.Errorf("git log failed: %w", err)
	}

	fmt.Println("=== Git Status ===")
	fmt.Println(statusStr)
	fmt.Println("=== Git Diff ===")
	fmt.Println(string(diffOutput))
	fmt.Println("=== Current Branch ===")
	fmt.Println(strings.TrimSpace(string(branchOutput)))
	fmt.Println("=== Recent Commits ===")
	fmt.Println(string(logOutput))
	fmt.Println("\nPrompt for LLM to create commit...")

	// TODO: Execute LLM with prompt to create commit
	return nil
}
