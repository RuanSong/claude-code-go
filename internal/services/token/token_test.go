package token

import (
	"testing"
)

func TestTokenEstimator_CountTokens(t *testing.T) {
	estimator := NewTokenEstimator()

	tests := []struct {
		name    string
		text    string
		wantMin int
		wantMax int
	}{
		{
			name:    "empty string",
			text:    "",
			wantMin: 0,
			wantMax: 0,
		},
		{
			name:    "single word",
			text:    "hello",
			wantMin: 1,
			wantMax: 2,
		},
		{
			name:    "multiple words",
			text:    "hello world foo bar",
			wantMin: 5,
			wantMax: 6,
		},
		{
			name:    "sentence",
			text:    "The quick brown fox jumps over the lazy dog",
			wantMin: 10,
			wantMax: 12,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := estimator.CountTokens(tt.text)
			if got < tt.wantMin || got > tt.wantMax {
				t.Errorf("CountTokens() = %v, want between %v and %v", got, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestTokenEstimator_CountMessagesTokens(t *testing.T) {
	estimator := NewTokenEstimator()

	messages := []string{
		"Hello world",
		"How are you today?",
		"Goodbye",
	}

	got := estimator.CountMessagesTokens(messages)
	if got <= 0 {
		t.Errorf("CountMessagesTokens() returned non-positive value: %v", got)
	}
}

func TestTokenEstimator_EstimateCost(t *testing.T) {
	estimator := NewTokenEstimator()

	tests := []struct {
		name  string
		usage TokenUsage
		model string
	}{
		{
			name: "claude-sonnet usage",
			usage: TokenUsage{
				InputTokens:  1000,
				OutputTokens: 500,
			},
			model: "claude-sonnet-4-20250514",
		},
		{
			name: "claude-opus usage",
			usage: TokenUsage{
				InputTokens:  1000,
				OutputTokens: 500,
			},
			model: "claude-opus-4-20250514",
		},
		{
			name: "unknown model defaults",
			usage: TokenUsage{
				InputTokens:  1000,
				OutputTokens: 500,
			},
			model: "unknown-model",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := estimator.EstimateCost(tt.usage, tt.model)
			if got < 0 {
				t.Errorf("EstimateCost() returned negative value: %v", got)
			}
		})
	}
}

func TestTokenEstimator_GetModelPricing(t *testing.T) {
	estimator := NewTokenEstimator()

	tests := []struct {
		name      string
		model     string
		wantInput float64
	}{
		{
			name:      "claude-sonnet pricing",
			model:     "claude-sonnet-4-20250514",
			wantInput: 3.0,
		},
		{
			name:      "claude-opus pricing",
			model:     "claude-opus-4-20250514",
			wantInput: 15.0,
		},
		{
			name:      "unknown model defaults to sonnet",
			model:     "unknown",
			wantInput: 3.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := estimator.GetModelPricing(tt.model)
			if got.InputCostPerMillion != tt.wantInput {
				t.Errorf("GetModelPricing() input = %v, want %v", got.InputCostPerMillion, tt.wantInput)
			}
		})
	}
}

func TestTokenEstimator_EstimateMessages(t *testing.T) {
	estimator := NewTokenEstimator()

	messages := []Message{
		{Role: "user", Content: "Hello"},
		{Role: "assistant", Content: "Hi there!"},
		{Role: "user", Content: "How are you?"},
	}

	result := estimator.EstimateMessages(messages, "claude-sonnet-4-20250514")

	if result.TotalTokens <= 0 {
		t.Errorf("EstimateMessages() TotalTokens = %v, want > 0", result.TotalTokens)
	}

	if result.EstimateUSD < 0 {
		t.Errorf("EstimateMessages() EstimateUSD = %v, want >= 0", result.EstimateUSD)
	}

	if len(result.Messages) != len(messages) {
		t.Errorf("EstimateMessages() Messages length = %v, want %v", len(result.Messages), len(messages))
	}
}

func TestTokenEstimator_EstimateMessagesWithStructContent(t *testing.T) {
	estimator := NewTokenEstimator()

	messages := []Message{
		{Role: "user", Content: "Test content"},
	}

	result := estimator.EstimateMessages(messages, "claude-sonnet-4-20250514")

	if result == nil {
		t.Fatalf("EstimateMessages() returned nil")
	}

	if result.TotalTokens == 0 {
		t.Errorf("EstimateMessages() TotalTokens = 0, want > 0 for non-empty message")
	}
}

func TestTokenUsage_TotalTokens(t *testing.T) {
	usage := TokenUsage{
		InputTokens:  1000,
		OutputTokens: 500,
		TotalTokens:  1500,
	}

	if usage.TotalTokens != usage.InputTokens+usage.OutputTokens {
		t.Errorf("TokenUsage TotalTokens = %v, want %v", usage.TotalTokens, usage.InputTokens+usage.OutputTokens)
	}
}

func TestModelPricing_KnownModels(t *testing.T) {
	knownModels := []string{
		"claude-opus-4-20250514",
		"claude-sonnet-4-20250514",
		"claude-3-5-sonnet-20241022",
		"claude-3-opus-20240229",
		"claude-3-sonnet-20240229",
		"claude-3-haiku-20240307",
	}

	estimator := NewTokenEstimator()

	for _, model := range knownModels {
		t.Run(model, func(t *testing.T) {
			pricing := estimator.GetModelPricing(model)
			if pricing.InputCostPerMillion <= 0 {
				t.Errorf("GetModelPricing(%q) InputCostPerMillion = %v, want > 0", model, pricing.InputCostPerMillion)
			}
			if pricing.OutputCostPerMillion <= 0 {
				t.Errorf("GetModelPricing(%q) OutputCostPerMillion = %v, want > 0", model, pricing.OutputCostPerMillion)
			}
		})
	}
}

func TestNewTokenEstimator(t *testing.T) {
	estimator := NewTokenEstimator()
	if estimator == nil {
		t.Error("NewTokenEstimator() returned nil")
	}
}

func TestTokenEstimator_MarshalJSON(t *testing.T) {
	estimator := NewTokenEstimator()

	data, err := estimator.MarshalJSON()
	if err != nil {
		t.Errorf("MarshalJSON() error = %v", err)
		return
	}

	if len(data) == 0 {
		t.Error("MarshalJSON() returned empty data")
	}
}
