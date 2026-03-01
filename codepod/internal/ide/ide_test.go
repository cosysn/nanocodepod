package ide

import (
	"testing"

	"github.com/codepod-io/codepod/internal/types"
)

func TestParseIDEType(t *testing.T) {
	tests := []struct {
		input    string
		expected types.IDEType
	}{
		{"vscode", types.IDETypeVSCode},
		{"code", types.IDETypeVSCode},
		{"VSCODE", types.IDETypeVSCode},
		{"jetbrains", types.IDETypeJetBrains},
		{"idea", types.IDETypeJetBrains},
		{"goland", types.IDETypeJetBrains},
		{"unknown", types.IDETypeVSCode}, // defaults to vscode
		{"", types.IDETypeVSCode},        // defaults to vscode
	}

	for _, tt := range tests {
		result := ParseIDEType(tt.input)
		if result != tt.expected {
			t.Errorf("ParseIDEType(%q) = %v, want %v", tt.input, result, tt.expected)
		}
	}
}

func TestLauncher_GetSupportedIDEs(t *testing.T) {
	launcher := New()
	ides := launcher.GetSupportedIDEs()

	if len(ides) == 0 {
		t.Error("should return at least one supported IDE")
	}

	// Check for expected IDEs
	found := false
	for _, ide := range ides {
		if ide == "vscode" {
			found = true
		}
	}
	if !found {
		t.Error("should include vscode in supported IDEs")
	}
}

func TestLauncher_IsIDEAvailable(t *testing.T) {
	launcher := New()

	// Test with vscode - may or may not be available depending on system
	result := launcher.IsIDEAvailable("vscode")
	_ = result // Just ensure it doesn't panic
}

func TestGetVSCodeRemoteArgs(t *testing.T) {
	args := GetVSCodeRemoteArgs("localhost", 22000, "root", "/workspace")

	if len(args) != 2 {
		t.Errorf("expected 2 args, got %d", len(args))
	}

	if args[0] != "--remote" {
		t.Errorf("expected --remote, got %s", args[0])
	}
}

func TestGetVSCodeWebURL(t *testing.T) {
	workspace := &types.Workspace{
		Name: "test",
		Port: 22000,
	}

	url := GetVSCodeWebURL(workspace)

	if url == "" {
		t.Error("url should not be empty")
	}
}
