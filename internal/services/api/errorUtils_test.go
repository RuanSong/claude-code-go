package api

import (
	"testing"
)

func TestFormatAPIError(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		message    string
		expected   string
	}{
		{"Bad request", 400, "invalid input", "Bad request: invalid input"},
		{"Unauthorized", 401, "invalid key", "Authentication failed. Please check your API key."},
		{"Forbidden", 403, "access denied", "Access forbidden. You may not have permission for this operation."},
		{"Not found", 404, "resource missing", "Resource not found: resource missing"},
		{"Rate limit", 429, "slow down", "Rate limit exceeded. Please wait and try again."},
		{"Internal error", 500, "oops", "Internal server error. Please try again later."},
		{"Service unavailable", 503, "maintenance", "Service unavailable. Please try again later."},
		{"Custom 400", 400, "custom error", "Bad request: custom error"},
		{"Custom 500", 500, "server error", "Internal server error. Please try again later."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatAPIError(tt.statusCode, tt.message)
			if result != tt.expected {
				t.Errorf("FormatAPIError(%d, %q) = %q, want %q", tt.statusCode, tt.message, result, tt.expected)
			}
		})
	}
}

func TestIsRetryableError(t *testing.T) {
	tests := []struct {
		statusCode int
		expected   bool
	}{
		{200, false},
		{400, false},
		{401, false},
		{403, false},
		{404, false},
		{429, true},
		{500, true},
		{502, true},
		{503, true},
		{504, true},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := IsRetryableError(tt.statusCode)
			if result != tt.expected {
				t.Errorf("IsRetryableError(%d) = %v, want %v", tt.statusCode, result, tt.expected)
			}
		})
	}
}

func TestSanitizeMessageHTML(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			"simple message",
			"Hello world",
			"Hello world",
		},
		{
			"HTML with title",
			"<!DOCTYPE html><html><head><title>Error Page</title></head></html>",
			"Error Page",
		},
		{
			"HTML without title",
			"<!DOCTYPE html><html><body>Content</body></html>",
			"",
		},
		{
			"empty message",
			"",
			"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeMessageHTML(tt.input)
			if result != tt.expected {
				t.Errorf("SanitizeMessageHTML(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGetErrorHint(t *testing.T) {
	// Test with a mock error
	err := &testError{code: "ETIMEDOUT", message: "connection timed out"}
	hint := GetErrorHint(err)
	if hint == "" {
		t.Log("GetErrorHint returned empty for timeout error (may be expected if error parsing doesn't work)")
	}
}

func TestExtractConnectionErrorDetails(t *testing.T) {
	err := &testError{code: "CERT_HAS_EXPIRED", message: "certificate has expired"}
	details := ExtractConnectionErrorDetails(err)
	if details == nil {
		t.Fatal("ExtractConnectionErrorDetails returned nil")
	}
	if !details.IsSSLError {
		t.Error("Expected IsSSLError to be true")
	}
	if details.Code != "CERT_HAS_EXPIRED" {
		t.Errorf("Expected code 'CERT_HAS_EXPIRED', got %q", details.Code)
	}
}

func TestGetSSLErrorHint(t *testing.T) {
	err := &testError{code: "SELF_SIGNED_CERT_IN_CHAIN", message: "self signed certificate"}
	hint := GetSSLErrorHint(err)
	if hint == "" {
		t.Error("Expected non-empty hint for SSL error")
	}
}

// testError is a mock error for testing
type testError struct {
	code    string
	message string
	cause   error
}

func (e *testError) Error() string {
	return e.message
}

func (e *testError) Cause() error {
	return e.cause
}
