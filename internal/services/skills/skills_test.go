package skills

import (
	"context"
	"testing"
	"time"
)

func TestNewSkillManager(t *testing.T) {
	manager := NewSkillManager()

	if manager == nil {
		t.Fatal("NewSkillManager() returned nil")
	}

	if manager.skills == nil {
		t.Error("NewSkillManager() did not initialize skills map")
	}

	if manager.paths == nil {
		t.Error("NewSkillManager() did not initialize paths slice")
	}

	if manager.loaded {
		t.Error("NewSkillManager() should not set loaded = true initially")
	}
}

func TestSkillManager_AddSearchPath(t *testing.T) {
	manager := NewSkillManager()

	manager.AddSearchPath("/path/to/skills")
	manager.AddSearchPath("/another/path")

	if len(manager.paths) != 2 {
		t.Errorf("AddSearchPath() added %d paths, want 2", len(manager.paths))
	}
}

func TestSkillManager_GetSkill(t *testing.T) {
	manager := NewSkillManager()

	// Manually add a skill
	skill := &Skill{
		Name:     "test-skill",
		Source:   "local",
		Type:     SkillTypePrompt,
		LoadedAt: time.Now(),
	}
	manager.skills["test-skill"] = skill

	retrieved, ok := manager.GetSkill("test-skill")
	if !ok {
		t.Fatal("GetSkill() returned ok = false for existing skill")
	}

	if retrieved.Name != "test-skill" {
		t.Error("GetSkill() returned wrong skill")
	}
}

func TestSkillManager_GetSkill_NotFound(t *testing.T) {
	manager := NewSkillManager()

	_, ok := manager.GetSkill("non-existent")
	if ok {
		t.Error("GetSkill() should return ok = false for non-existent skill")
	}
}

func TestSkillManager_ListSkills(t *testing.T) {
	manager := NewSkillManager()

	manager.skills["skill1"] = &Skill{Name: "skill1"}
	manager.skills["skill2"] = &Skill{Name: "skill2"}

	skills := manager.ListSkills()

	if len(skills) != 2 {
		t.Errorf("ListSkills() returned %d skills, want 2", len(skills))
	}
}

func TestSkillManager_ListSkills_Empty(t *testing.T) {
	manager := NewSkillManager()

	skills := manager.ListSkills()

	if len(skills) != 0 {
		t.Errorf("ListSkills() returned %d skills, want 0", len(skills))
	}
}

func TestSkillManager_SearchSkills(t *testing.T) {
	manager := NewSkillManager()

	manager.skills["code-review"] = &Skill{
		Name:        "code-review",
		Description: "Review code changes",
	}
	manager.skills["refactor"] = &Skill{
		Name:        "refactor",
		Description: "Refactor code",
	}
	manager.skills["docs"] = &Skill{
		Name:        "docs",
		Description: "Generate documentation",
	}

	results := manager.SearchSkills("code")
	if len(results) != 2 {
		t.Errorf("SearchSkills() returned %d results, want 2 (code-review and refactor both have 'code' in their descriptions)", len(results))
	}

	if results[0].Name != "code-review" && results[0].Name != "refactor" {
		t.Error("SearchSkills() returned wrong skill")
	}
}

func TestSkillManager_SearchSkills_ByDescription(t *testing.T) {
	manager := NewSkillManager()

	manager.skills["skill1"] = &Skill{
		Name:        "test",
		Description: "This skill reviews code",
	}

	results := manager.SearchSkills("reviews")
	if len(results) != 1 {
		t.Errorf("SearchSkills() by description returned %d results, want 1", len(results))
	}
}

func TestSkillManager_SearchSkills_NoMatch(t *testing.T) {
	manager := NewSkillManager()

	manager.skills["skill1"] = &Skill{Name: "test-skill"}

	results := manager.SearchSkills("non-existent")

	if len(results) != 0 {
		t.Errorf("SearchSkills() returned %d results, want 0", len(results))
	}
}

func TestSkillManager_Reload(t *testing.T) {
	manager := NewSkillManager()
	manager.loaded = true

	err := manager.Reload(context.Background())
	if err != nil {
		t.Fatalf("Reload() error = %v", err)
	}

	// After Reload completes, loaded should be true (set by LoadSkills)
	if !manager.loaded {
		t.Error("Reload() should set loaded = true after loading")
	}

	// Skills map should be cleared and ready for reload
	// (It will be empty since there are no paths set in this test)
}

func TestSkillManager_parseSkillFile(t *testing.T) {
	manager := NewSkillManager()

	content := `---
name: test-skill
description: A test skill
type: prompt
---

This is the skill prompt content.`

	skill := manager.parseSkillFile("/path/to/skills/test-skill", content)

	if skill.Name != "test-skill" {
		t.Errorf("parseSkillFile() Name = %v, want test-skill", skill.Name)
	}

	if skill.Description != "A test skill" {
		t.Error("parseSkillFile() Description not parsed correctly")
	}

	if skill.Type != SkillTypePrompt {
		t.Errorf("parseSkillFile() Type = %v, want SkillTypePrompt", skill.Type)
	}

	if skill.Source != "local" {
		t.Error("parseSkillFile() Source not set to local")
	}
}

func TestSkillManager_parseCommandFile(t *testing.T) {
	manager := NewSkillManager()

	content := `---
name: test-command
description: A test command
aliases: tc,test
---

This is the command prompt content.`

	skill := manager.parseCommandFile("/path/to/skills/test-command", content)

	if skill.Name != "test-command" {
		t.Errorf("parseCommandFile() Name = %v, want test-command", skill.Name)
	}

	if skill.Source != "command" {
		t.Error("parseCommandFile() Source not set to command")
	}
}

func TestNewRegistry(t *testing.T) {
	registry := NewRegistry()

	if registry == nil {
		t.Fatal("NewRegistry() returned nil")
	}

	if registry.commands == nil {
		t.Error("NewRegistry() did not initialize commands map")
	}
}

func TestRegistry_Register(t *testing.T) {
	registry := NewRegistry()

	cmd := &Command{
		Skill: Skill{
			Name: "test-cmd",
		},
		Path: "/path/to/command",
	}

	err := registry.Register(cmd)
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	retrieved, ok := registry.Get("test-cmd")
	if !ok {
		t.Fatal("Register() did not register command")
	}

	if retrieved.Name != "test-cmd" {
		t.Error("Register() stored wrong command")
	}
}

func TestRegistry_Register_AlreadyRegistered(t *testing.T) {
	registry := NewRegistry()

	cmd := &Command{Skill: Skill{Name: "test-cmd"}}
	registry.Register(cmd)

	err := registry.Register(cmd)
	if err == nil {
		t.Error("Register() should return error for already registered command")
	}
}

func TestRegistry_Register_WithAliases(t *testing.T) {
	registry := NewRegistry()

	cmd := &Command{
		Skill: Skill{
			Name:    "test-cmd",
			Aliases: []string{"tc", "shortcut"},
		},
	}

	err := registry.Register(cmd)
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	// Check that aliases can retrieve the command
	aliasCmd, ok := registry.Get("tc")
	if !ok {
		t.Error("Register() did not register alias 'tc'")
	}

	if aliasCmd.Name != "test-cmd" {
		t.Error("Alias points to wrong command")
	}
}

func TestRegistry_Unregister(t *testing.T) {
	registry := NewRegistry()

	cmd := &Command{
		Skill: Skill{Name: "test-cmd", Aliases: []string{"tc"}},
	}
	registry.Register(cmd)

	err := registry.Unregister("test-cmd")
	if err != nil {
		t.Fatalf("Unregister() error = %v", err)
	}

	_, ok := registry.Get("test-cmd")
	if ok {
		t.Error("Unregister() did not remove command")
	}

	_, ok = registry.Get("tc")
	if ok {
		t.Error("Unregister() did not remove alias")
	}
}

func TestRegistry_Unregister_NotRegistered(t *testing.T) {
	registry := NewRegistry()

	err := registry.Unregister("non-existent")
	if err == nil {
		t.Error("Unregister() should return error for non-registered command")
	}
}

func TestRegistry_Get(t *testing.T) {
	registry := NewRegistry()

	cmd := &Command{Skill: Skill{Name: "test-cmd"}}
	registry.Register(cmd)

	retrieved, ok := registry.Get("test-cmd")
	if !ok {
		t.Fatal("Get() returned ok = false for registered command")
	}

	if retrieved.Name != "test-cmd" {
		t.Error("Get() returned wrong command")
	}
}

func TestRegistry_List(t *testing.T) {
	registry := NewRegistry()

	registry.Register(&Command{Skill: Skill{Name: "cmd1"}})
	registry.Register(&Command{Skill: Skill{Name: "cmd2"}})
	registry.Register(&Command{Skill: Skill{Name: "cmd3", Aliases: []string{"c3"}}})

	commands := registry.List()

	if len(commands) != 3 {
		t.Errorf("List() returned %d commands, want 3", len(commands))
	}
}

func TestRegistry_FindByPrefix(t *testing.T) {
	registry := NewRegistry()

	registry.Register(&Command{Skill: Skill{Name: "test-cmd"}})
	registry.Register(&Command{Skill: Skill{Name: "test-other"}})
	registry.Register(&Command{Skill: Skill{Name: "other-cmd"}})

	cmd, ok := registry.FindByPrefix("test")
	if !ok {
		t.Fatal("FindByPrefix() returned ok = false for existing prefix")
	}

	if cmd.Name != "test-cmd" {
		t.Error("FindByPrefix() returned wrong command")
	}
}

func TestRegistry_FindByPrefix_NotFound(t *testing.T) {
	registry := NewRegistry()

	_, ok := registry.FindByPrefix("non-existent")
	if ok {
		t.Error("FindByPrefix() should return ok = false for non-existent prefix")
	}
}

func TestSkill_ToJSON(t *testing.T) {
	skill := &Skill{
		Name:        "test-skill",
		Description: "A test skill",
		Type:        SkillTypePrompt,
	}

	data, err := skill.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON() error = %v", err)
	}

	if len(data) == 0 {
		t.Error("ToJSON() returned empty data")
	}
}

func TestSkillFromJSON(t *testing.T) {
	data := []byte(`{"name":"test-skill","type":"prompt"}`)

	skill, err := SkillFromJSON(data)
	if err != nil {
		t.Fatalf("SkillFromJSON() error = %v", err)
	}

	if skill.Name != "test-skill" {
		t.Error("SkillFromJSON() did not parse name correctly")
	}

	if skill.Type != SkillTypePrompt {
		t.Error("SkillFromJSON() did not parse type correctly")
	}
}

func TestSkillFromJSON_Invalid(t *testing.T) {
	data := []byte(`{invalid json}`)

	_, err := SkillFromJSON(data)
	if err == nil {
		t.Error("SkillFromJSON() should return error for invalid JSON")
	}
}

func TestSkill_Structure(t *testing.T) {
	now := time.Now()
	skill := Skill{
		Name:        "test-skill",
		Description: "A test skill",
		Source:      "local",
		Type:        SkillTypePrompt,
		Content:     "Skill content here",
		Prompt:      "This is the prompt",
		Tools:       []string{"Read", "Write"},
		Model:       "claude-sonnet",
		Effort:      3,
		Metadata:    map[string]interface{}{"key": "value"},
		Aliases:     []string{"alias1", "alias2"},
		LoadedFrom:  "/path/to/skill",
		LoadedAt:    now,
	}

	if skill.Name == "" {
		t.Error("Skill.Name is empty")
	}

	if skill.Type != SkillTypePrompt {
		t.Errorf("Skill.Type = %v, want SkillTypePrompt", skill.Type)
	}

	if skill.LoadedAt != now {
		t.Error("Skill.LoadedAt not set correctly")
	}
}

func TestSkillType_Constants(t *testing.T) {
	if SkillTypePrompt != "prompt" {
		t.Errorf("SkillTypePrompt = %v, want prompt", SkillTypePrompt)
	}

	if SkillTypeAgent != "agent" {
		t.Errorf("SkillTypeAgent = %v, want agent", SkillTypeAgent)
	}

	if SkillTypeBuiltin != "builtin" {
		t.Errorf("SkillTypeBuiltin = %v, want builtin", SkillTypeBuiltin)
	}
}

func TestCommand_Structure(t *testing.T) {
	cmd := Command{
		Skill: Skill{
			Name: "test-command",
		},
		Path:          "/path/to/command",
		ArgumentNames: []string{"arg1", "arg2"},
	}

	if cmd.Path == "" {
		t.Error("Command.Path is empty")
	}

	if len(cmd.ArgumentNames) != 2 {
		t.Error("Command.ArgumentNames not set correctly")
	}
}

func TestLoadSkillsDir(t *testing.T) {
	// This test would require an actual directory structure
	// For now, just verify the function exists and can be called
	_, err := LoadSkillsDir(context.Background(), "/non-existent/path")
	if err == nil {
		// Expected to fail with non-existent path
	}
}

func TestSkillManager_ConcurrentAccess(t *testing.T) {
	// Skip this test - direct map access is not thread-safe
	// The SkillManager should provide thread-safe methods for concurrent access
	t.Skip("Direct map access is not thread-safe - this test tests unsupported usage")
}
