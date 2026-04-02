package voiceKeyterms

import (
	"testing"
)

func TestVoiceKeytermsServiceInstance(t *testing.T) {
	instance1 := GetInstance()
	instance2 := GetInstance()

	if instance1 != instance2 {
		t.Error("GetInstance should return the same instance (singleton)")
	}
}

func TestSplitIdentifier(t *testing.T) {
	service := &VoiceKeytermsService{}

	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "camelCase",
			input:    "myVariableName",
			expected: []string{"my", "Variable", "Name"},
		},
		{
			name:     "PascalCase",
			input:    "MyClassName",
			expected: []string{"My", "Class", "Name"},
		},
		{
			name:     "kebab-case",
			input:    "my-class-name",
			expected: []string{"my", "class", "name"},
		},
		{
			name:     "snake_case",
			input:    "my_class_name",
			expected: []string{"my", "class", "name"},
		},
		{
			name:     "single short word",
			input:    "ab",
			expected: []string{},
		},
		{
			name:     "mixed with numbers",
			input:    "test123File",
			expected: []string{"test", "File"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.splitIdentifier(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("splitIdentifier(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestFileNameWords(t *testing.T) {
	service := &VoiceKeytermsService{}

	tests := []struct {
		name     string
		input    string
		expected int // minimum expected count
	}{
		{
			name:     "Go file",
			input:    "/path/to/main.go",
			expected: 1,
		},
		{
			name:     "TypeScript file",
			input:    "/path/to/index.ts",
			expected: 1,
		},
		{
			name:     "nested path",
			input:    "/path/to/myComponent.tsx",
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.fileNameWords(tt.input)
			if len(result) < tt.expected {
				t.Errorf("fileNameWords(%q) returned %d words, expected at least %d", tt.input, len(result), tt.expected)
			}
		})
	}
}

func TestGetGlobalKeyterms(t *testing.T) {
	service := &VoiceKeytermsService{}

	keyterms := service.GetGlobalKeyterms()

	if len(keyterms) == 0 {
		t.Error("GetGlobalKeyterms should return non-empty slice")
	}

	// Check for expected keyterms
	expectedTerms := []string{"MCP", "symlink", "grep", "regex"}
	for _, term := range expectedTerms {
		found := false
		for _, k := range keyterms {
			if k == term {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected keyterm %q not found", term)
		}
	}
}

func TestGetVoiceKeyterms(t *testing.T) {
	service := &VoiceKeytermsService{}

	// Test with no recent files
	keyterms := service.GetVoiceKeyterms(nil)

	if len(keyterms) == 0 {
		t.Error("GetVoiceKeyterms should return non-empty slice")
	}

	// Check that global keyterms are included
	for _, term := range GLOBAL_KEYTERMS {
		found := false
		for _, k := range keyterms {
			if k == term {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Global keyterm %q should be in result", term)
		}
	}
}

func TestGetVoiceKeytermsWithFiles(t *testing.T) {
	service := &VoiceKeytermsService{}

	recentFiles := []string{
		"/path/to/main.go",
		"/path/to/utils.go",
	}

	keyterms := service.GetVoiceKeyterms(recentFiles)

	if len(keyterms) == 0 {
		t.Error("GetVoiceKeyterms should return non-empty slice")
	}

	// Should contain words from file names
	// Note: The exact words depend on the implementation
	if len(keyterms) < len(GLOBAL_KEYTERMS) {
		t.Errorf("Expected at least %d keyterms (global), got %d", len(GLOBAL_KEYTERMS), len(keyterms))
	}
}

func TestGetVoiceKeytermsMaxLimit(t *testing.T) {
	service := &VoiceKeytermsService{}

	// Create a list of files that would produce more than MAX_KEYTERMS
	recentFiles := make([]string, 100)
	for i := 0; i < 100; i++ {
		recentFiles[i] = "/path/to/file" + string(rune('a'+i%26)) + "_component" + string(rune('0'+i%10)) + ".ts"
	}

	keyterms := service.GetVoiceKeyterms(recentFiles)

	if len(keyterms) > MAX_KEYTERMS {
		t.Errorf("Expected at most %d keyterms, got %d", MAX_KEYTERMS, len(keyterms))
	}
}
