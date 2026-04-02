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

var skillsStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("cyan"))
var skillsNameStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("green"))
var skillsDescStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("white"))

// SkillsCommand lists and manages skills
type SkillsCommand struct {
	BaseCommand
}

func NewSkillsCommand() *SkillsCommand {
	return &SkillsCommand{
		BaseCommand: *newCommand("skills", "List available skills"),
	}
}

func (c *SkillsCommand) Execute(ctx context.Context, args []string, execCtx engine.CommandContext) error {
	if len(args) == 0 {
		return c.listSkills(ctx, execCtx)
	}

	subcommand := args[0]

	switch subcommand {
	case "list", "ls":
		return c.listSkills(ctx, execCtx)
	case "add":
		return c.addSkill(ctx, args[1:])
	case "remove", "rm":
		if len(args) < 2 {
			return fmt.Errorf("Usage: /skills remove <skill-name>")
		}
		return c.removeSkill(ctx, args[1])
	default:
		return fmt.Errorf("Unknown subcommand: %s. Use 'list' or 'remove'", subcommand)
	}
}

type Skill struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Path        string `json:"path"`
}

func (c *SkillsCommand) listSkills(ctx context.Context, execCtx engine.CommandContext) error {
	fmt.Println(skillsStyle.Render("Available Skills"))
	fmt.Println(strings.Repeat("─", 40))
	fmt.Println()

	// Get skills from multiple locations
	skills := c.findSkills(execCtx.Cwd)

	if len(skills) == 0 {
		fmt.Println("No skills found.")
		fmt.Println()
		fmt.Println("Skills are stored in .claude/skills/ directories.")
		fmt.Println("Create a skill by adding a SKILL.md file to .claude/skills/<skill-name>/")
		return nil
	}

	// Group skills by source
	builtinSkills := []Skill{}
	projectSkills := []Skill{}
	userSkills := []Skill{}

	for _, skill := range skills {
		if strings.Contains(skill.Path, ".claude/skills") {
			if strings.Contains(skill.Path, execCtx.Cwd) {
				projectSkills = append(projectSkills, skill)
			} else {
				userSkills = append(userSkills, skill)
			}
		} else {
			builtinSkills = append(builtinSkills, skill)
		}
	}

	if len(builtinSkills) > 0 {
		fmt.Println(skillsStyle.Render("Bundled Skills:"))
		for _, skill := range builtinSkills {
			fmt.Printf("  %s - %s\n", skillsNameStyle.Render(skill.Name), skill.Description)
		}
		fmt.Println()
	}

	if len(projectSkills) > 0 {
		fmt.Println(skillsStyle.Render("Project Skills:"))
		for _, skill := range projectSkills {
			fmt.Printf("  %s - %s\n", skillsNameStyle.Render(skill.Name), skill.Description)
		}
		fmt.Println()
	}

	if len(userSkills) > 0 {
		fmt.Println(skillsStyle.Render("User Skills:"))
		for _, skill := range userSkills {
			fmt.Printf("  %s - %s\n", skillsNameStyle.Render(skill.Name), skill.Description)
		}
		fmt.Println()
	}

	return nil
}

func (c *SkillsCommand) findSkills(cwd string) []Skill {
	skills := []Skill{}

	// Search locations
	searchPaths := []string{
		filepath.Join(cwd, ".claude", "skills"),
		filepath.Join(cwd, "skills"),
		getUserSkillsDir(),
	}

	for _, path := range searchPaths {
		found := c.scanSkillsDir(path)
		skills = append(skills, found...)
	}

	return skills
}

func (c *SkillsCommand) scanSkillsDir(dir string) []Skill {
	skills := []Skill{}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return skills
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		skillPath := filepath.Join(dir, entry.Name())
		skillFile := filepath.Join(skillPath, "SKILL.md")

		if _, err := os.Stat(skillFile); err == nil {
			desc := c.getSkillDescription(skillFile)
			skills = append(skills, Skill{
				Name:        entry.Name(),
				Description: desc,
				Path:        skillPath,
			})
		}
	}

	return skills
}

func (c *SkillsCommand) getSkillDescription(skillFile string) string {
	content, err := os.ReadFile(skillFile)
	if err != nil {
		return "No description"
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "description:") {
			return strings.TrimPrefix(line, "description:")
		}
		if strings.HasPrefix(line, "---") && strings.Contains(line, "description") {
			continue
		}
		if strings.HasPrefix(line, "#") {
			// Skip markdown headers
			continue
		}
		if len(line) > 0 && !strings.HasPrefix(line, "---") {
			return line
		}
	}

	return "No description"
}

func (c *SkillsCommand) addSkill(ctx context.Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("Usage: /skills add <skill-name>")
	}

	skillName := args[0]
	cwd := os.Getenv("PWD")
	if cwd == "" {
		cwd = "."
	}

	skillsDir := filepath.Join(cwd, ".claude", "skills", skillName)
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		return fmt.Errorf("failed to create skill directory: %w", err)
	}

	skillFile := filepath.Join(skillsDir, "SKILL.md")
	if _, err := os.Stat(skillFile); err == nil {
		return fmt.Errorf("skill '%s' already exists", skillName)
	}

	content := `---
name: ` + skillName + `
description: Describe what this skill does and when to use it
---

# ` + skillName + `

Instructions for Claude on how to use this skill.
`

	if err := os.WriteFile(skillFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to create skill file: %w", err)
	}

	fmt.Printf("%s Created skill: %s\n", skillsNameStyle.Render("✓"), skillName)
	fmt.Printf("  Edit %s to customize the skill.\n", skillFile)

	return nil
}

func (c *SkillsCommand) removeSkill(ctx context.Context, name string) error {
	cwd := os.Getenv("PWD")
	if cwd == "" {
		cwd = "."
	}

	skillPath := filepath.Join(cwd, ".claude", "skills", name)
	skillFile := filepath.Join(skillPath, "SKILL.md")

	if _, err := os.Stat(skillFile); os.IsNotExist(err) {
		return fmt.Errorf("skill '%s' not found", name)
	}

	if err := os.RemoveAll(skillPath); err != nil {
		return fmt.Errorf("failed to remove skill: %w", err)
	}

	fmt.Printf("%s Removed skill: %s\n", skillsNameStyle.Render("✓"), name)
	return nil
}

func getUserSkillsDir() string {
	home, _ := os.UserHomeDir()
	if home == "" {
		home = os.Getenv("HOME")
	}
	return filepath.Join(home, ".claude", "skills")
}
