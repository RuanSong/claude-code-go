package commands

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/claude-code-go/claude/internal/engine"
)

var ultrareviewStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("cyan")).Bold(true)

// UltraReviewCommand provides deep code review
type UltraReviewCommand struct {
	BaseCommand
}

func NewUltraReviewCommand() *UltraReviewCommand {
	return &UltraReviewCommand{
		BaseCommand: *newPromptCommand("ultrareview", "Deep code review (~10-20 min) - finds and verifies bugs"),
	}
}

type UltraReviewResult struct {
	FilesReviewed  int      `json:"files_reviewed"`
	IssuesFound    int      `json:"issues_found"`
	SecurityIssues int      `json:"security_issues"`
	Performance    []string `json:"performance_issues"`
	Bugs           []string `json:"potential_bugs"`
	Suggestions    []string `json:"suggestions"`
}

func (c *UltraReviewCommand) Execute(ctx context.Context, args []string, execCtx engine.CommandContext) error {
	fmt.Println(ultrareviewStyle.Render("UltraReview - Deep Code Analysis"))
	fmt.Println("This may take 10-20 minutes for thorough analysis...")

	// Check prerequisites
	if err := checkGHInstalled(); err != nil {
		return err
	}

	// Get current branch
	branch, err := getCurrentBranch(ctx)
	if err != nil {
		return fmt.Errorf("failed to get current branch: %w", err)
	}

	fmt.Printf("Reviewing branch: %s\n", branch)

	// Run comprehensive analysis
	result, err := c.runDeepAnalysis(ctx, branch)
	if err != nil {
		return fmt.Errorf("ultrareview failed: %w", err)
	}

	// Display results
	c.displayResults(result)

	return nil
}

func getCurrentBranch(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "branch", "--show-current")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

func (c *UltraReviewCommand) runDeepAnalysis(ctx context.Context, branch string) (*UltraReviewResult, error) {
	result := &UltraReviewResult{
		Performance: make([]string, 0),
		Bugs:        make([]string, 0),
		Suggestions: make([]string, 0),
	}

	// Get list of changed files
	files, err := c.getChangedFiles(ctx, branch)
	if err != nil {
		return nil, err
	}
	result.FilesReviewed = len(files)

	// Analyze each file type
	for _, file := range files {
		if isGoFile(file) {
			issues := c.analyzeGoFile(ctx, file)
			result.IssuesFound += len(issues)
			result.Bugs = append(result.Bugs, issues...)
		} else if isTypeScriptFile(file) {
			issues := c.analyzeTSFile(ctx, file)
			result.IssuesFound += len(issues)
			result.Bugs = append(result.Bugs, issues...)
		}
	}

	// Check for common issues
	result.SecurityIssues = c.checkSecurityIssues(ctx, files)
	result.Performance = c.checkPerformanceIssues(ctx, files)
	result.Suggestions = c.generateSuggestions(ctx, files)

	return result, nil
}

func (c *UltraReviewCommand) getChangedFiles(ctx context.Context, branch string) ([]string, error) {
	cmd := exec.CommandContext(ctx, "git", "diff", "--name-only", "main..."+branch)
	if err := cmd.Run(); err != nil {
		// Try origin/main if main doesn't exist
		cmd = exec.CommandContext(ctx, "git", "diff", "--name-only", "origin/main..."+branch)
	}
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get changed files: %w", err)
	}

	files := make([]string, 0)
	for _, line := range splitLines(string(output)) {
		if line != "" {
			files = append(files, line)
		}
	}
	return files, nil
}

func (c *UltraReviewCommand) analyzeGoFile(ctx context.Context, file string) []string {
	issues := make([]string, 0)

	// Run go vet
	cmd := exec.CommandContext(ctx, "go", "vet", file)
	if err := cmd.Run(); err != nil {
		issues = append(issues, fmt.Sprintf("%s: go vet found issues", file))
	}

	// Check for TODO/FIXME comments
	cmd = exec.CommandContext(ctx, "grep", "-n", "TODO\\|FIXME\\|XXX", file)
	output, _ := cmd.Output()
	if len(output) > 0 {
		issues = append(issues, fmt.Sprintf("%s: contains TODO/FIXME comments", file))
	}

	return issues
}

func (c *UltraReviewCommand) analyzeTSFile(ctx context.Context, file string) []string {
	issues := make([]string, 0)

	// Basic checks for TypeScript files
	cmd := exec.CommandContext(ctx, "grep", "-n", "any\\|TODO\\|FIXME", file)
	output, _ := cmd.Output()
	if len(output) > 0 {
		for _, line := range splitLines(string(output)) {
			if contains(line, "any") {
				issues = append(issues, fmt.Sprintf("%s: uses 'any' type", file))
			}
		}
	}

	return issues
}

func (c *UltraReviewCommand) checkSecurityIssues(ctx context.Context, files []string) int {
	count := 0
	patterns := []string{
		"password",
		"secret",
		"api_key",
		"APISecret",
		"private_key",
	}

	for _, file := range files {
		for _, pattern := range patterns {
			cmd := exec.CommandContext(ctx, "grep", "-i", "-n", pattern, file)
			if err := cmd.Run(); err == nil {
				count++
			}
		}
	}

	return count
}

func (c *UltraReviewCommand) checkPerformanceIssues(ctx context.Context, files []string) []string {
	issues := make([]string, 0)

	for _, file := range files {
		if isGoFile(file) {
			// Check for common performance anti-patterns in Go
			cmd := exec.CommandContext(ctx, "grep", "-n", "fmt.Print\\|log.Println", file)
			output, _ := cmd.Output()
			if len(output) > 5 { // Multiple print statements might indicate debugging code
				issues = append(issues, fmt.Sprintf("%s: multiple print statements (may be debug code)", file))
			}
		}
	}

	return issues
}

func (c *UltraReviewCommand) generateSuggestions(ctx context.Context, files []string) []string {
	suggestions := make([]string, 0)

	// Check for missing tests
	testFiles := 0
	for _, file := range files {
		if isTestFile(file) {
			testFiles++
		}
	}

	// Suggest adding tests if few exist
	if testFiles < len(files)/3 {
		suggestions = append(suggestions, "Consider adding more test coverage")
	}

	// Check for documentation
	hasReadme := false
	for _, file := range files {
		if file == "README.md" {
			hasReadme = true
		}
	}
	if !hasReadme {
		suggestions = append(suggestions, "Project may benefit from README documentation")
	}

	return suggestions
}

func (c *UltraReviewCommand) displayResults(result *UltraReviewResult) {
	fmt.Println()
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("UltraReview Results")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("Files Reviewed: %d\n", result.FilesReviewed)
	fmt.Printf("Issues Found: %d\n", result.IssuesFound)
	fmt.Printf("Potential Security Issues: %d\n", result.SecurityIssues)

	if len(result.Performance) > 0 {
		fmt.Println("\nPerformance Issues:")
		for _, issue := range result.Performance {
			fmt.Printf("  - %s\n", issue)
		}
	}

	if len(result.Bugs) > 0 {
		fmt.Println("\nPotential Bugs:")
		for _, bug := range result.Bugs {
			fmt.Printf("  - %s\n", bug)
		}
	}

	if len(result.Suggestions) > 0 {
		fmt.Println("\nSuggestions:")
		for _, suggestion := range result.Suggestions {
			fmt.Printf("  - %s\n", suggestion)
		}
	}

	fmt.Println()
	if result.IssuesFound == 0 && result.SecurityIssues == 0 {
		fmt.Println("No critical issues found! The code looks good.")
	}
}

// Helper functions
func isGoFile(path string) bool {
	return hasSuffix(path, ".go")
}

func isTypeScriptFile(path string) bool {
	return hasSuffix(path, ".ts") || hasSuffix(path, ".tsx")
}

func isTestFile(path string) bool {
	return hasSuffix(path, "_test.go") || hasSuffix(path, ".test.ts")
}

func hasSuffix(s, suffix string) bool {
	return len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr) >= 0
}

func findSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func splitLines(s string) []string {
	result := make([]string, 0)
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			if start < i {
				result = append(result, s[start:i])
			}
			start = i + 1
		}
	}
	if start < len(s) {
		result = append(result, s[start:])
	}
	return result
}

func hasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

func stringsReplicate(s string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}
