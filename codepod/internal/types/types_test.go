package types

import (
	"testing"
	"time"
)

func TestWorkspaceState(t *testing.T) {
	if WorkspaceStateCreated != "created" {
		t.Errorf("Expected 'created', got %s", WorkspaceStateCreated)
	}
	if WorkspaceStateRunning != "running" {
		t.Errorf("Expected 'running', got %s", WorkspaceStateRunning)
	}
	if WorkspaceStateStopped != "stopped" {
		t.Errorf("Expected 'stopped', got %s", WorkspaceStateStopped)
	}
	if WorkspaceStateError != "error" {
		t.Errorf("Expected 'error', got %s", WorkspaceStateError)
	}
}

func TestIDEType(t *testing.T) {
	if IDETypeVSCode != "vscode" {
		t.Errorf("Expected 'vscode', got %s", IDETypeVSCode)
	}
	if IDETypeJetBrains != "jetbrains" {
		t.Errorf("Expected 'jetbrains', got %s", IDETypeJetBrains)
	}
}

func TestWorkspace(t *testing.T) {
	ws := Workspace{
		Name:    "test-workspace",
		UUID:    "test-uuid-123",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		State:    WorkspaceStateRunning,
		Repository: Repository{
			URL:      "https://github.com/test/repo",
			Branch:   "main",
			LocalPath: "/tmp/test",
		},
		IDE: IDE{
			Type: IDETypeVSCode,
			Settings: map[string]string{
				"theme": "dark",
			},
		},
		Container: Container{
			Image: "ubuntu:22.04",
			Name:  "test-container",
		},
		SSH: SSH{
			ConfigPath: "/home/user/.ssh/config",
			KeyPath:    "/home/user/.ssh/id_rsa",
		},
		Agent: Agent{
			Port:   22022,
			Status: "running",
		},
		Port:        22000,
		StoragePath: "/home/user/.codepod/workspaces/test",
		CodePath:    "/home/user/.codepod/workspaces/test/code",
	}

	if ws.Name != "test-workspace" {
		t.Errorf("Expected 'test-workspace', got %s", ws.Name)
	}
	if ws.State != WorkspaceStateRunning {
		t.Errorf("Expected 'running', got %s", ws.State)
	}
	if ws.Port != 22000 {
		t.Errorf("Expected 22000, got %d", ws.Port)
	}
}

func TestConfig(t *testing.T) {
	cfg := Config{
		Version: "1.0.0",
		WSL: WSLConfig{
			Distribution: "Ubuntu-22.04",
			DockerHost:  "unix:///var/run/docker.sock",
		},
		General: GeneralConfig{
			DefaultIDE: "vscode",
			SSHPort:    22,
		},
		PortPool: PortPool{
			Start: 22000,
			End:   22999,
			Used:  []int{22000, 22001},
		},
		DataDir: "/home/user/.codepod",
	}

	if cfg.Version != "1.0.0" {
		t.Errorf("Expected '1.0.0', got %s", cfg.Version)
	}
	if cfg.WSL.Distribution != "Ubuntu-22.04" {
		t.Errorf("Expected 'Ubuntu-22.04', got %s", cfg.WSL.Distribution)
	}
	if cfg.PortPool.Start != 22000 {
		t.Errorf("Expected 22000, got %d", cfg.PortPool.Start)
	}
	if len(cfg.PortPool.Used) != 2 {
		t.Errorf("Expected 2, got %d", len(cfg.PortPool.Used))
	}
}
