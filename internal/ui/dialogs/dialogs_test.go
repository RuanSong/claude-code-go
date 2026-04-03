package dialogs

import (
	"testing"
)

func TestBaseDialog(t *testing.T) {
	dialog := NewBaseDialog("Test Title")

	if dialog.Title != "Test Title" {
		t.Errorf("Expected title 'Test Title', got %s", dialog.Title)
	}

	dialog.SetDescription("Test Description")
	if dialog.Description != "Test Description" {
		t.Error("SetDescription should work")
	}

	dialog.SetWidth(80)
	if dialog.Width != 80 {
		t.Error("SetWidth should work")
	}
}

func TestConfirmationDialog(t *testing.T) {
	dialog := NewConfirmationDialog("Confirm", "Are you sure?")

	dialog.AddOption("Yes", "yes")
	dialog.AddOption("No", "no")

	if len(dialog.Options) != 2 {
		t.Errorf("Expected 2 options, got %d", len(dialog.Options))
	}

	dialog.SetCancelText("Cancel")
	if dialog.CancelText != "Cancel" {
		t.Error("SetCancelText should work")
	}

	dialog.SetConfirmText("OK")
	if dialog.ConfirmText != "OK" {
		t.Error("SetConfirmText should work")
	}

	output := dialog.Render()
	if output == "" {
		t.Error("Render should return non-empty string")
	}
}

func TestMCPServerApprovalDialog(t *testing.T) {
	dialog := NewMCPServerApprovalDialog("example.com")

	if dialog.ServerName != "example.com" {
		t.Errorf("Expected server name 'example.com', got %s", dialog.ServerName)
	}

	dialog.SetCustom(true)
	if !dialog.IsCustom {
		t.Error("SetCustom should work")
	}

	output := dialog.Render()
	if output == "" {
		t.Error("Render should return non-empty string")
	}
}

func TestMCPServerMultiselectDialog(t *testing.T) {
	servers := []string{"server1.com", "server2.com", "server3.com"}
	dialog := NewMCPServerMultiselectDialog(servers)

	if len(dialog.Servers) != 3 {
		t.Errorf("Expected 3 servers, got %d", len(dialog.Servers))
	}

	if dialog.Cursor != 0 {
		t.Error("Initial cursor should be 0")
	}

	dialog.MoveDown()
	if dialog.Cursor != 1 {
		t.Error("MoveDown should increment cursor")
	}

	dialog.MoveUp()
	if dialog.Cursor != 0 {
		t.Error("MoveUp should decrement cursor")
	}

	dialog.Toggle()
	if dialog.Checked[0] {
		t.Error("Toggle should uncheck first item")
	}

	dialog.SelectAll()
	for i := 0; i < len(servers); i++ {
		if !dialog.Checked[i] {
			t.Error("SelectAll should check all items")
		}
	}

	dialog.SelectNone()
	for i := 0; i < len(servers); i++ {
		if dialog.Checked[i] {
			t.Error("SelectNone should uncheck all items")
		}
	}

	dialog.Confirm()
	if len(dialog.Approved) != 0 {
		t.Error("With none selected, Approved should be empty")
	}
}

func TestTextInputDialog(t *testing.T) {
	dialog := NewTextInputDialog("Enter Name", "Your name...")

	dialog.SetDefaultValue("John")
	if dialog.DefaultValue != "John" {
		t.Error("SetDefaultValue should work")
	}

	if dialog.Value != "John" {
		t.Errorf("Value should be initialized to DefaultValue, got %s", dialog.Value)
	}

	dialog.SetMaxLength(10)

	dialog.Clear()
	dialog.Append("Hello")
	if dialog.Value != "Hello" {
		t.Errorf("Expected 'Hello', got %s", dialog.Value)
	}

	dialog.MoveLeft()
	if dialog.CursorPos != 4 {
		t.Error("MoveLeft should move cursor")
	}

	dialog.MoveRight()
	if dialog.CursorPos != 5 {
		t.Error("MoveRight should move cursor")
	}

	dialog.MoveToStart()
	if dialog.CursorPos != 0 {
		t.Error("MoveToStart should move cursor to 0")
	}

	dialog.MoveToEnd()
	if dialog.CursorPos != 5 {
		t.Error("MoveToEnd should move cursor to end")
	}

	dialog.Delete()
	if dialog.Value != "Hell" {
		t.Error("Delete should remove character")
	}

	dialog.Clear()
	if dialog.Value != "" {
		t.Error("Clear should empty value")
	}

	output := dialog.Render()
	if output == "" {
		t.Error("Render should return non-empty string")
	}
}

func TestTeamsDialog(t *testing.T) {
	dialog := NewTeamsDialog()

	info := TeammateInfo{
		Name:   "test-teammate",
		Status: "idle",
		Mode:   "auto",
		Model:  "claude-sonnet-4",
		Path:   "/tmp",
	}
	dialog.AddTeammate(info)

	if len(dialog.Teammates) != 1 {
		t.Error("AddTeammate should add teammate")
	}

	if dialog.Cursor != 0 {
		t.Error("Initial cursor should be 0")
	}

	dialog.MoveUp()
	if dialog.Cursor != 0 {
		t.Error("MoveUp at index 0 should stay at 0")
	}

	dialog.Select()
	if dialog.ViewMode != "detail" {
		t.Error("Select should switch to detail view")
	}

	dialog.Back()
	if dialog.ViewMode != "list" {
		t.Error("Back should switch back to list view")
	}

	output := dialog.Render()
	if output == "" {
		t.Error("Render should return non-empty string")
	}
}

func TestOnboardingDialog(t *testing.T) {
	dialog := NewOnboardingDialog()

	if len(dialog.Steps) == 0 {
		t.Error("Onboarding should have steps")
	}

	if dialog.CurrentStep != 0 {
		t.Error("Initial step should be 0")
	}

	dialog.NextStep()
	if dialog.CurrentStep != 1 {
		t.Error("NextStep should increment")
	}

	dialog.PrevStep()
	if dialog.CurrentStep != 0 {
		t.Error("PrevStep should decrement")
	}

	dialog.SetTheme("light")
	if dialog.Theme != "light" {
		t.Error("SetTheme should work")
	}

	dialog.SetAPIKey("sk-xxx")
	if dialog.APIKey != "sk-xxx" {
		t.Error("SetAPIKey should work")
	}

	if dialog.IsComplete() {
		t.Error("Should not be complete yet")
	}

	// Go to last step
	for i := 0; i < len(dialog.Steps); i++ {
		dialog.NextStep()
	}
	if !dialog.IsComplete() {
		t.Error("Should be complete at last step")
	}

	output := dialog.Render()
	if output == "" {
		t.Error("Render should return non-empty string")
	}
}
