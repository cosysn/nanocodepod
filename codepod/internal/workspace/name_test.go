package workspace

import (
	"testing"
)

func TestSanitizeContainerName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"test-workspace", "test-workspace"},
		{"test_workspace", "test_workspace"},
		{"test.workspace", "test.workspace"},
		{"test-workspace!", "test-workspace_"},
		{"TestWorkspace", "testworkspace"},
		{"test@workspace#", "test_workspace_"},
		{"", "workspace"},
		{"123", "123"},
		{"a", "a"},
		{"test workspace", "test_workspace"},
		{"test/workspace", "test_workspace"},
	}

	for _, tt := range tests {
		result := sanitizeContainerName(tt.input)
		if result != tt.expected {
			t.Errorf("sanitizeContainerName(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestGetContainerName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"myproject", "codepod-myproject"},
		{"test", "codepod-test"},
		{"test-1", "codepod-test-1"},
	}

	for _, tt := range tests {
		result := GetContainerName(tt.input)
		if result != tt.expected {
			t.Errorf("GetContainerName(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}
