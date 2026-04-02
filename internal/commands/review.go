package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/claude-code-go/claude/internal/engine"
)

const reviewTermsURL = "https://code.claude.com/docs/en/claude-code-on-the-web"

// ReviewCommand reviews pull requests
type ReviewCommand struct {
	BaseCommand
}

func NewReviewCommand() *ReviewCommand {
	return &ReviewCommand{
		BaseCommand: *newPromptCommand("review", "Review a pull request"),
	}
}

func (c *ReviewCommand) GetAllowedTools() []string {
	return []string{
		"Bash(gh pr list:*)",
		"Bash(gh pr view:*)",
		"Bash(gh pr diff:*)",
		"Bash(gh api:*)",
	}
}

func (c *ReviewCommand) GetPromptTemplate() string {
	return `You are an expert code reviewer. Follow these steps:

1. If no PR number is provided in the args, run ` + "`gh pr list`" + ` to show open PRs
2. If a PR number is provided, run ` + "`gh pr view <number>`" + ` to get PR details
3. Run ` + "`gh pr diff <number>`" + ` to get the diff
4. Analyze the changes and provide a thorough code review that includes:
   - Overview of what the PR does
   - Analysis of code quality and style
   - Specific suggestions for improvements
   - Any potential issues or risks

Keep your review concise but thorough. Focus on:
- Code correctness
- Following project conventions
- Performance implications
- Test coverage
- Security considerations

Format your review with clear sections and bullet points.

Be thorough but respectful. Highlight both strengths and areas for improvement.
`
}

type ReviewResult struct {
	PRNumber   int    `json:"pr_number"`
	Title      string `json:"title"`
	Author     string `json:"author"`
	ReviewText string `json:"review_text"`
}

func (c *ReviewCommand) Execute(ctx context.Context, args []string, execCtx engine.CommandContext) error {
	// Check if gh CLI is installed
	if err := checkGHInstalled(); err != nil {
		return err
	}

	// Check if we're in a git repo
	if err := checkGitRepo(); err != nil {
		return err
	}

	// Determine PR number from args
	prNumber := 0
	if len(args) > 0 {
		// Try to parse PR number
		fmt.Sscanf(args[0], "%d", &prNumber)
	}

	// List open PRs if no number provided
	if prNumber == 0 {
		return c.listPullRequests(ctx)
	}

	// Get PR details and diff
	return c.reviewPullRequest(ctx, prNumber)
}

func checkGHInstalled() error {
	cmd := exec.Command("gh", "--version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("GitHub CLI (gh) is not installed. Install from: https://cli.github.com")
	}
	return nil
}

func checkGitRepo() error {
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Not in a git repository")
	}
	return nil
}

func (c *ReviewCommand) listPullRequests(ctx context.Context) error {
	fmt.Println("Fetching open pull requests...")

	cmd := exec.CommandContext(ctx, "gh", "pr", "list", "--state", "open", "--limit", "20", "--json", "number,title,author,url")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to list PRs: %w", err)
	}

	var prs []struct {
		Number int    `json:"number"`
		Title  string `json:"title"`
		Author struct {
			Login string `json:"login"`
		} `json:"author"`
		URL string `json:"url"`
	}

	if err := json.Unmarshal(output, &prs); err != nil {
		return fmt.Errorf("failed to parse PR list: %w", err)
	}

	if len(prs) == 0 {
		fmt.Println("No open pull requests found.")
		return nil
	}

	fmt.Println("\nOpen Pull Requests:")
	fmt.Println("-------------------")
	for _, pr := range prs {
		fmt.Printf("#%d: %s (@%s)\n", pr.Number, pr.Title, pr.Author.Login)
		fmt.Printf("   %s\n\n", pr.URL)
	}

	fmt.Println("Run `/review <number>` to review a specific PR.")
	return nil
}

func (c *ReviewCommand) reviewPullRequest(ctx context.Context, prNumber int) error {
	fmt.Printf("Reviewing PR #%d...\n\n", prNumber)

	// Get PR details
	prCmd := exec.CommandContext(ctx, "gh", "pr", "view", fmt.Sprintf("%d", prNumber), "--json", "title,body,author,additions,deletions,changedFiles")
	prOutput, err := prCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get PR details: %w", err)
	}

	var prInfo struct {
		Title  string `json:"title"`
		Body   string `json:"body"`
		Author struct {
			Login string `json:"login"`
		} `json:"author"`
		Additions    int `json:"additions"`
		Deletions    int `json:"deletions"`
		ChangedFiles int `json:"changedFiles"`
	}

	if err := json.Unmarshal(prOutput, &prInfo); err != nil {
		return fmt.Errorf("failed to parse PR info: %w", err)
	}

	// Get diff
	diffCmd := exec.CommandContext(ctx, "gh", "pr", "diff", fmt.Sprintf("%d", prNumber))
	diffOutput, err := diffCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get PR diff: %w", err)
	}

	// Display PR summary
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("PR #%d: %s\n", prNumber, prInfo.Title)
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("Author: @%s\n", prInfo.Author.Login)
	fmt.Printf("Changes: +%d additions, -%d deletions, %d files changed\n\n",
		prInfo.Additions, prInfo.Deletions, prInfo.ChangedFiles)

	if prInfo.Body != "" {
		fmt.Println("Description:")
		fmt.Println(prInfo.Body)
		fmt.Println()
	}

	// Get file changes summary
	_ = string(diffOutput) // diff is available for LLM to analyze

	// Show diff summary by file
	files := extractChangedFiles(string(diffOutput))
	fmt.Println("Changed Files:")
	for i, file := range files {
		if i >= 20 {
			fmt.Printf("  ... and %d more files\n", len(files)-20)
			break
		}
		fmt.Printf("  %s\n", file)
	}

	fmt.Println()
	fmt.Println("To get a detailed review, run:")
	fmt.Printf("  claude /review %d\n", prNumber)
	fmt.Println("This will analyze the full diff and provide a comprehensive review.")

	// TODO: Call LLM to do actual review
	fmt.Println("\n[Full review analysis pending LLM integration]")

	return nil
}

func extractChangedFiles(diff string) []string {
	files := make([]string, 0)
	lines := strings.Split(diff, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "diff --git") {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				// Extract file path from "a/path b/path" format
				path := parts[2]
				path = strings.TrimPrefix(path, "a/")
				path = strings.TrimPrefix(path, "b/")
				files = append(files, path)
			}
		}
	}
	return files
}
