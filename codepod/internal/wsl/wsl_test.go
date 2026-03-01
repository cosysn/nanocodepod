package wsl

import (
	"testing"
)

func TestDetectPlatform(t *testing.T) {
	platform := DetectPlatform()

	// Should return one of the platform types
	switch platform {
	case PlatformLinux, PlatformWSL, PlatformWindows:
		// OK
	default:
		t.Errorf("unexpected platform: %s", platform)
	}
}

func TestIsDockerAvailable(t *testing.T) {
	// This test depends on Docker being installed
	// Just ensure it doesn't panic
	_ = IsDockerAvailable()
}

func TestGetDockerHost(t *testing.T) {
	tests := []struct {
		platform PlatformType
		expected string
	}{
		{PlatformLinux, "unix:///var/run/docker.sock"},
		{PlatformWSL, "tcp://localhost:2375"},
		{PlatformWindows, "tcp://localhost:2375"},
	}

	for _, tt := range tests {
		result := GetDockerHost(tt.platform)
		if result != tt.expected {
			t.Errorf("GetDockerHost(%s) = %s, want %s", tt.platform, result, tt.expected)
		}
	}
}

func TestPlatform_GetHostname(t *testing.T) {
	platform, err := NewPlatform()
	if err != nil {
		t.Fatalf("failed to create platform: %v", err)
	}

	hostname, err := platform.GetHostname()
	if err != nil {
		t.Fatalf("failed to get hostname: %v", err)
	}

	if hostname == "" {
		t.Error("hostname should not be empty")
	}
}

func TestPlatform_RunCommand(t *testing.T) {
	platform, err := NewPlatform()
	if err != nil {
		t.Fatalf("failed to create platform: %v", err)
	}

	// Test simple command
	output, err := platform.RunCommand("echo hello")
	if err != nil {
		t.Fatalf("failed to run command: %v", err)
	}

	if output != "hello" {
		t.Errorf("expected 'hello', got '%s'", output)
	}
}

func TestPlatform_FileExists(t *testing.T) {
	platform, err := NewPlatform()
	if err != nil {
		t.Fatalf("failed to create platform: %v", err)
	}

	// Test with existing file
	exists := platform.FileExists("/etc/passwd")
	if !exists {
		t.Error("/etc/passwd should exist")
	}

	// Test with non-existing file
	exists = platform.FileExists("/nonexistent/path/file")
	if exists {
		t.Error("should not exist")
	}
}

func TestNewPlatform(t *testing.T) {
	platform, err := NewPlatform()
	if err != nil {
		t.Fatalf("failed to create platform: %v", err)
	}

	if platform.Type == "" {
		t.Error("platform type should not be empty")
	}
}
