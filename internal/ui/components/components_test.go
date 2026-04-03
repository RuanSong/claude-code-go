package components

import (
	"testing"
)

func TestSpinnerState(t *testing.T) {
	state := NewSpinnerState()

	if state.Frame != 0 {
		t.Errorf("Expected Frame 0, got %d", state.Frame)
	}

	if state.Verb != "Loading..." {
		t.Errorf("Expected 'Loading...', got %s", state.Verb)
	}

	if !state.IsActive {
		t.Error("Expected IsActive to be true")
	}
}

func TestSpinnerStateNextFrame(t *testing.T) {
	state := NewSpinnerState()

	initialFrame := state.Frame
	state.NextFrame()

	if state.Frame == initialFrame {
		t.Error("NextFrame should change frame")
	}
}

func TestSpinnerStatePauseResume(t *testing.T) {
	state := NewSpinnerState()

	state.Pause()
	if !state.IsPaused {
		t.Error("Expected IsPaused to be true")
	}

	state.Resume()
	if state.IsPaused {
		t.Error("Expected IsPaused to be false")
	}
}

func TestSpinnerStateSetVerb(t *testing.T) {
	state := NewSpinnerState()
	state.SetVerb("Thinking...")

	if state.Verb != "Thinking..." {
		t.Errorf("Expected 'Thinking...', got %s", state.Verb)
	}
}

func TestSpinnerStateSetSubText(t *testing.T) {
	state := NewSpinnerState()
	state.SetSubText("Processing files...")

	if state.SubText != "Processing files..." {
		t.Errorf("Expected 'Processing files...', got %s", state.SubText)
	}
}

func TestSpinnerStateSetTokenBudget(t *testing.T) {
	state := NewSpinnerState()
	state.SetTokenBudget(1000, 9000, 10000)

	if state.TokenBudget == nil {
		t.Fatal("TokenBudget should not be nil")
	}

	if state.TokenBudget.Used != 1000 {
		t.Errorf("Expected Used=1000, got %d", state.TokenBudget.Used)
	}

	if state.TokenBudget.Remaining != 9000 {
		t.Errorf("Expected Remaining=9000, got %d", state.TokenBudget.Remaining)
	}

	if state.TokenBudget.Total != 10000 {
		t.Errorf("Expected Total=10000, got %d", state.TokenBudget.Total)
	}
}

func TestSpinnerStateRender(t *testing.T) {
	state := NewSpinnerState()
	state.SetVerb("Thinking...")
	state.SetSubText("Processing...")

	output := state.Render()

	if output == "" {
		t.Error("Render should return non-empty string")
	}
}

func TestSpinnerWithVerb(t *testing.T) {
	s := NewSpinnerWithVerb("Thinking...")

	if s == nil {
		t.Fatal("NewSpinnerWithVerb should not return nil")
	}

	if !s.IsActive() {
		t.Error("Expected IsActive to be true")
	}

	s.SetVerb("Reading files...")
	s.SetSubText("file.txt")
	s.SetNextTask("Next: writing...")

	s.Tick()

	output := s.Render()
	if output == "" {
		t.Error("Render should return non-empty string")
	}

	s.Stop()
	if s.IsActive() {
		t.Error("Expected IsActive to be false after Stop")
	}
}

func TestBriefSpinner(t *testing.T) {
	s := NewBriefSpinner()

	if !s.IsActive() {
		t.Error("Expected IsActive to be true")
	}

	s.Tick()

	output := s.Render()
	if output == "" {
		t.Error("Render should return non-empty string")
	}

	s.Stop()
	if s.IsActive() {
		t.Error("Expected IsActive to be false after Stop")
	}
}

func TestSelectOption(t *testing.T) {
	opt := NewSelectOption("Label", "value")

	if opt.Label != "Label" {
		t.Errorf("Expected Label='Label', got %s", opt.Label)
	}

	if opt.Value != "value" {
		t.Errorf("Expected Value='value', got %s", opt.Value)
	}

	opt.WithDescription("Description")
	if opt.Description != "Description" {
		t.Error("WithDescription should set Description")
	}

	opt.WithKey("k")
	if opt.Key != "k" {
		t.Error("WithKey should set Key")
	}

	opt.Disable()
	if !opt.Disabled {
		t.Error("Disable should set Disabled to true")
	}
}

func TestSelectModel(t *testing.T) {
	options := []*SelectOption{
		NewSelectOption("Option 1", "1"),
		NewSelectOption("Option 2", "2"),
		NewSelectOption("Option 3", "3"),
	}

	model := NewSelectModel(options)

	if len(model.Options) != 3 {
		t.Errorf("Expected 3 options, got %d", len(model.Options))
	}

	if model.Cursor != 0 {
		t.Error("Expected Cursor to be 0")
	}

	if model.Selected != -1 {
		t.Error("Expected Selected to be -1")
	}

	model.MoveDown()
	if model.Cursor != 1 {
		t.Error("MoveDown should increment Cursor")
	}

	model.MoveUp()
	if model.Cursor != 0 {
		t.Error("MoveUp should decrement Cursor")
	}

	model.Select()
	if model.Selected != 0 {
		t.Error("Select should set Selected to Cursor")
	}
}

func TestSelectModelMulti(t *testing.T) {
	options := []*SelectOption{
		NewSelectOption("Option 1", "1"),
		NewSelectOption("Option 2", "2"),
	}

	model := NewSelectModelMulti(options)

	if !model.MultiSelect {
		t.Error("MultiSelect should be true")
	}

	model.Toggle()
	if !model.IsChecked(0) {
		t.Error("First option should be checked after Toggle")
	}

	model.Toggle()
	if model.IsChecked(0) {
		t.Error("First option should be unchecked after second Toggle")
	}

	model.CheckAll()
	if !model.IsChecked(0) || !model.IsChecked(1) {
		t.Error("CheckAll should check all options")
	}

	model.UncheckAll()
	if model.IsChecked(0) || model.IsChecked(1) {
		t.Error("UncheckAll should uncheck all options")
	}
}

func TestSelectModelFilter(t *testing.T) {
	options := []*SelectOption{
		NewSelectOption("Apple", "apple"),
		NewSelectOption("Banana", "banana"),
		NewSelectOption("Cherry", "cherry"),
	}

	model := NewSelectModel(options)
	model.FilterText = "ap"

	filtered := model.FilterOptions()
	if len(filtered) != 1 {
		t.Errorf("Expected 1 filtered option, got %d", len(filtered))
	}

	if filtered[0].Label != "Apple" {
		t.Error("Filtered option should be Apple")
	}
}

func TestSelectModelRender(t *testing.T) {
	options := []*SelectOption{
		NewSelectOption("Option 1", "1").WithDescription("First option"),
		NewSelectOption("Option 2", "2"),
	}

	model := NewSelectModel(options)
	model.SetTitle("Test Title")
	model.SetDescription("Test Description")

	output := model.Render()

	if output == "" {
		t.Error("Render should return non-empty string")
	}
}
