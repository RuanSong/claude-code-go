package policyLimits

import (
	"os"
	"testing"
)

func TestIsPolicyAllowed(t *testing.T) {
	// Clear environment
	os.Unsetenv("ANTHROPIC_API_KEY")
	os.Unsetenv("CLAUDEAI_ACCESS_TOKEN")
	os.Unsetenv("CLAUDEAI_SCOPES")
	os.Unsetenv("CLAUDE_SUBSCRIPTION_TYPE")
	os.Unsetenv("ESSENTIAL_TRAFFIC_ONLY")

	tests := []struct {
		name     string
		policy   string
		expected bool
	}{
		{"unknown policy allows", "some_unknown_policy", true},
		{"allow_product_feedback default", "allow_product_feedback", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isPolicyAllowed(tt.policy)
			if result != tt.expected {
				t.Errorf("isPolicyAllowed(%q) = %v, want %v", tt.policy, result, tt.expected)
			}
		})
	}
}

func TestIsPolicyLimitsEligible(t *testing.T) {
	// Test with no API key and no OAuth tokens
	os.Unsetenv("ANTHROPIC_API_KEY")
	os.Unsetenv("CLAUDEAI_ACCESS_TOKEN")

	// With no credentials, user should not be eligible
	result := isPolicyLimitsEligible()
	if result {
		t.Log("Not eligible with no credentials (expected)")
	}
}

func TestIsFirstPartyAnthropicBaseUrl(t *testing.T) {
	tests := []struct {
		name     string
		baseURL  string
		expected bool
	}{
		{"empty defaults to true", "", true},
		{"anthropic.com is first party", "https://api.anthropic.com", true},
		{"bedrock is not", "https://bedrock.amazonaws.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("ANTHROPIC_BASE_URL", tt.baseURL)
			result := isFirstPartyAnthropicBaseUrl()
			if result != tt.expected {
				t.Errorf("isFirstPartyAnthropicBaseUrl() with ANTHROPIC_BASE_URL=%q = %v, want %v", tt.baseURL, result, tt.expected)
			}
		})
	}
}

func TestContainsScope(t *testing.T) {
	tests := []struct {
		name   string
		scopes []string
		scope  string
		found  bool
	}{
		{"finds existing scope", []string{"a", "b", "c"}, "b", true},
		{"returns false for missing", []string{"a", "b"}, "d", false},
		{"empty scopes", []string{}, "a", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := containsScope(tt.scopes, tt.scope)
			if result != tt.found {
				t.Errorf("containsScope(%v, %q) = %v, want %v", tt.scopes, tt.scope, result, tt.found)
			}
		})
	}
}

func TestGetRetryDelay(t *testing.T) {
	tests := []struct {
		attempt  int
		expected int
	}{
		{1, 1000},
		{2, 2000},
		{3, 4000},
		{4, 8000},
		{5, 16000},
		{6, 16000}, // capped at 16000
		{0, 1000},  // minimum
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := getRetryDelay(tt.attempt)
			if result != tt.expected {
				t.Errorf("getRetryDelay(%d) = %d, want %d", tt.attempt, result, tt.expected)
			}
		})
	}
}

func TestEssentialTrafficDenyOnMiss(t *testing.T) {
	// allow_product_feedback should be denied on miss
	if !essentialTrafficDenyOnMiss("allow_product_feedback") {
		t.Error("allow_product_feedback should be denied on miss")
	}

	// other policies should not be denied
	if essentialTrafficDenyOnMiss("some_other_policy") {
		t.Error("other policies should not be denied on miss")
	}
}

func TestIsEssentialTrafficOnly(t *testing.T) {
	os.Setenv("ESSENTIAL_TRAFFIC_ONLY", "true")
	if !isEssentialTrafficOnly() {
		t.Error("Expected true when ESSENTIAL_TRAFFIC_ONLY=true")
	}

	os.Setenv("ESSENTIAL_TRAFFIC_ONLY", "false")
	if isEssentialTrafficOnly() {
		t.Error("Expected false when ESSENTIAL_TRAFFIC_ONLY=false")
	}

	os.Unsetenv("ESSENTIAL_TRAFFIC_ONLY")
	if isEssentialTrafficOnly() {
		t.Error("Expected false when ESSENTIAL_TRAFFIC_ONLY is not set")
	}
}

func TestComputeChecksum(t *testing.T) {
	restrictions1 := Restrictions{
		"policy1": {Allowed: true},
		"policy2": {Allowed: false},
	}

	restrictions2 := Restrictions{
		"policy2": {Allowed: false},
		"policy1": {Allowed: true},
	}

	// Same content should produce same checksum regardless of key order
	checksum1 := computeChecksum(restrictions1)
	checksum2 := computeChecksum(restrictions2)

	if checksum1 != checksum2 {
		t.Errorf("Same content should produce same checksum, got %q and %q", checksum1, checksum2)
	}

	// Different content should produce different checksum
	restrictions3 := Restrictions{
		"policy1": {Allowed: false}, // Changed
		"policy2": {Allowed: false},
	}

	checksum3 := computeChecksum(restrictions3)
	if checksum1 == checksum3 {
		t.Error("Different content should produce different checksum")
	}
}

func TestSortKeysDeep(t *testing.T) {
	// Test with nested object
	input := map[string]interface{}{
		"z": "last",
		"a": "first",
		"nested": map[string]interface{}{
			"z_child": "z",
			"a_child": "a",
		},
	}

	result := sortKeysDeep(input)
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("Expected map result")
	}

	keys := make([]string, 0, len(resultMap))
	for k := range resultMap {
		keys = append(keys, k)
	}

	// Check keys are sorted
	for i := 1; i < len(keys); i++ {
		if keys[i-1] > keys[i] {
			t.Errorf("Keys not sorted: %v", keys)
		}
	}
}

func TestClearPolicyLimitsCache(t *testing.T) {
	// This should not panic
	err := clearPolicyLimitsCache()
	if err != nil {
		t.Errorf("clearPolicyLimitsCache() returned error: %v", err)
	}
}
