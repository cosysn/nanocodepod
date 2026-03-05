package bootstrapper

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestNewBootstrapper(t *testing.T) {
	b := NewBootstrapper("/path/to/binary")
	if b.binaryPath != "/path/to/binary" {
		t.Fatalf("expected /path/to/binary, got %s", b.binaryPath)
	}
}

func TestBootstrap_EmptyTarget(t *testing.T) {
	b := NewBootstrapper("/path/to/binary")

	config := &BootstrapConfig{
		TargetType: TargetTypeDocker,
		Target:     "",
		AgentType:  AgentTypeLocal,
		BinaryPath: "/path/to/binary",
	}

	err := b.Bootstrap(context.Background(), config)
	if err == nil {
		t.Fatal("expected error for empty target, got nil")
	}
}

func TestBootstrap_BinaryNotFound(t *testing.T) {
	b := NewBootstrapper("/nonexistent/binary")

	config := &BootstrapConfig{
		TargetType: TargetTypeDocker,
		Target:     "container1",
		AgentType:  AgentTypeLocal,
		BinaryPath: "/nonexistent/binary",
	}

	err := b.Bootstrap(context.Background(), config)
	if err == nil {
		t.Fatal("expected error for missing binary, got nil")
	}
}

func TestBootstrap_UnknownTargetType(t *testing.T) {
	// Create a temp binary file
	tmpDir := t.TempDir()
	binaryPath := filepath.Join(tmpDir, "agent")
	err := os.WriteFile(binaryPath, []byte("#!/bin/bash\necho hi"), 0755)
	if err != nil {
		t.Fatalf("failed to create test binary: %v", err)
	}

	b := NewBootstrapper(binaryPath)

	config := &BootstrapConfig{
		TargetType: "unknown",
		Target:     "container1",
		AgentType:  AgentTypeLocal,
		BinaryPath: binaryPath,
	}

	err = b.Bootstrap(context.Background(), config)
	if err == nil {
		t.Fatal("expected error for unknown target type, got nil")
	}
}

func TestGetAgentFlag(t *testing.T) {
	b := NewBootstrapper("/path/to/binary")

	tests := []struct {
		agentType AgentType
		expected  string
	}{
		{AgentTypeLocal, "--local"},
		{AgentTypeWorkspace, "--workspace"},
		{AgentTypeContainer, "--container"},
		{"unknown", ""},
	}

	for _, tt := range tests {
		result := b.getAgentFlag(tt.agentType)
		if result != tt.expected {
			t.Errorf("getAgentFlag(%s) = %s; want %s", tt.agentType, result, tt.expected)
		}
	}
}

func TestTargetTypeConstants(t *testing.T) {
	if TargetTypeWSL != "wsl" {
		t.Errorf("TargetTypeWSL = %s; want wsl", TargetTypeWSL)
	}
	if TargetTypeDocker != "docker" {
		t.Errorf("TargetTypeDocker = %s; want docker", TargetTypeDocker)
	}
	if TargetTypeSSH != "ssh" {
		t.Errorf("TargetTypeSSH = %s; want ssh", TargetTypeSSH)
	}
	if TargetTypeUDS != "uds" {
		t.Errorf("TargetTypeUDS = %s; want uds", TargetTypeUDS)
	}
}

func TestAgentTypeConstants(t *testing.T) {
	if AgentTypeLocal != "local" {
		t.Errorf("AgentTypeLocal = %s; want local", AgentTypeLocal)
	}
	if AgentTypeWorkspace != "workspace" {
		t.Errorf("AgentTypeWorkspace = %s; want workspace", AgentTypeWorkspace)
	}
	if AgentTypeContainer != "container" {
		t.Errorf("AgentTypeContainer = %s; want container", AgentTypeContainer)
	}
}

func TestBootstrapConfig(t *testing.T) {
	config := &BootstrapConfig{
		TargetType: TargetTypeDocker,
		Target:     "my-container",
		AgentType:  AgentTypeContainer,
		BinaryPath: "/usr/local/bin/agent",
	}

	if config.TargetType != TargetTypeDocker {
		t.Errorf("TargetType = %s; want docker", config.TargetType)
	}
	if config.Target != "my-container" {
		t.Errorf("Target = %s; want my-container", config.Target)
	}
	if config.AgentType != AgentTypeContainer {
		t.Errorf("AgentType = %s; want container", config.AgentType)
	}
	if config.BinaryPath != "/usr/local/bin/agent" {
		t.Errorf("BinaryPath = %s; want /usr/local/bin/agent", config.BinaryPath)
	}
}
