package types

import (
	"testing"
	"time"
)

func TestWorkspaceState(t *testing.T) {
	tests := []struct {
		name  string
		state WorkspaceState
		want  string
	}{
		{"created", WorkspaceStateCreated, "created"},
		{"running", WorkspaceStateRunning, "running"},
		{"stopped", WorkspaceStateStopped, "stopped"},
		{"error", WorkspaceStateError, "error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.state) != tt.want {
				t.Errorf("want %s, got %s", tt.want, tt.state)
			}
		})
	}
}

func TestIDEType(t *testing.T) {
	if IDETypeVSCode != "vscode" {
		t.Errorf("want vscode, got %s", IDETypeVSCode)
	}
	if IDETypeJetBrains != "jetbrains" {
		t.Errorf("want jetbrains, got %s", IDETypeJetBrains)
	}
}

func TestWorkspace(t *testing.T) {
	ws := Workspace{
		Name:        "test-workspace",
		UUID:        "test-uuid-123",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		State:       WorkspaceStateRunning,
		Domain:      "test.local",
		Port:        22001,
		StoragePath: "/root/.codepod/workspaces/test",
		CodePath:    "/workspace",
		Repository: Repository{
			URL:       "https://github.com/test/repo",
			Branch:    "main",
			LocalPath: "",
		},
		IDE: IDE{
			Type: IDETypeVSCode,
			Settings: map[string]string{
				"fontSize": "14",
			},
		},
		Container: Container{
			Image: "ubuntu:22.04",
			Name:  "codepod-test-workspace",
		},
		SSH: SSH{
			ConfigPath: "/root/.ssh/config",
			KeyPath:    "/root/.ssh/id_rsa",
		},
		Agent: Agent{
			Port:   22001,
			Status: "running",
		},
	}

	if ws.Name != "test-workspace" {
		t.Errorf("want test-workspace, got %s", ws.Name)
	}
	if ws.UUID != "test-uuid-123" {
		t.Errorf("want test-uuid-123, got %s", ws.UUID)
	}
	if ws.State != WorkspaceStateRunning {
		t.Errorf("want running, got %s", ws.State)
	}
	if ws.Port != 22001 {
		t.Errorf("want 22001, got %d", ws.Port)
	}
	if ws.Repository.URL != "https://github.com/test/repo" {
		t.Errorf("want URL, got %s", ws.Repository.URL)
	}
	if ws.IDE.Type != IDETypeVSCode {
		t.Errorf("want vscode, got %s", ws.IDE.Type)
	}
	if ws.Container.Image != "ubuntu:22.04" {
		t.Errorf("want ubuntu:22.04, got %s", ws.Container.Image)
	}
	if ws.Agent.Status != "running" {
		t.Errorf("want running, got %s", ws.Agent.Status)
	}
}

func TestRepository(t *testing.T) {
	repo := Repository{
		URL:       "https://github.com/test/repo",
		Branch:    "develop",
		LocalPath: "/tmp/repo",
	}

	if repo.URL != "https://github.com/test/repo" {
		t.Errorf("want URL, got %s", repo.URL)
	}
	if repo.Branch != "develop" {
		t.Errorf("want develop, got %s", repo.Branch)
	}
	if repo.LocalPath != "/tmp/repo" {
		t.Errorf("want /tmp/repo, got %s", repo.LocalPath)
	}
}

func TestIDE(t *testing.T) {
	ide := IDE{
		Type: IDETypeJetBrains,
		Settings: map[string]string{
			"theme":       "dark",
			"fontSize":    "16",
			"tabSize":     "4",
		},
	}

	if ide.Type != IDETypeJetBrains {
		t.Errorf("want jetbrains, got %s", ide.Type)
	}
	if ide.Settings["theme"] != "dark" {
		t.Errorf("want dark, got %s", ide.Settings["theme"])
	}
}

func TestContainer(t *testing.T) {
	container := Container{
		Image: "alpine:latest",
		Name:  "test-container",
	}

	if container.Image != "alpine:latest" {
		t.Errorf("want alpine:latest, got %s", container.Image)
	}
	if container.Name != "test-container" {
		t.Errorf("want test-container, got %s", container.Name)
	}
}

func TestSSH(t *testing.T) {
	ssh := SSH{
		ConfigPath: "/home/user/.ssh/config",
		KeyPath:    "/home/user/.ssh/id_ed25519",
	}

	if ssh.ConfigPath != "/home/user/.ssh/config" {
		t.Errorf("want config path, got %s", ssh.ConfigPath)
	}
	if ssh.KeyPath != "/home/user/.ssh/id_ed25519" {
		t.Errorf("want key path, got %s", ssh.KeyPath)
	}
}

func TestAgent(t *testing.T) {
	agent := Agent{
		Port:   22002,
		Status: "stopped",
	}

	if agent.Port != 22002 {
		t.Errorf("want 22002, got %d", agent.Port)
	}
	if agent.Status != "stopped" {
		t.Errorf("want stopped, got %s", agent.Status)
	}
}

func TestAgentEmpty(t *testing.T) {
	agent := Agent{}

	if agent.Port != 0 {
		t.Errorf("want 0, got %d", agent.Port)
	}
	if agent.Status != "" {
		t.Errorf("want empty, got %s", agent.Status)
	}
}

func TestConfig(t *testing.T) {
	cfg := Config{
		Version: "1.0.0",
		WSL: WSLConfig{
			Distribution: "Ubuntu-22.04",
			DockerHost:   "tcp://localhost:2375",
		},
		General: GeneralConfig{
			DefaultIDE: "vscode",
			SSHPort:    22,
		},
		PortPool: PortPool{
			Start: 22000,
			End:   22010,
			Used:  []int{22001, 22002},
		},
	}

	if cfg.Version != "1.0.0" {
		t.Errorf("want 1.0.0, got %s", cfg.Version)
	}
	if cfg.WSL.Distribution != "Ubuntu-22.04" {
		t.Errorf("want Ubuntu-22.04, got %s", cfg.WSL.Distribution)
	}
	if cfg.General.DefaultIDE != "vscode" {
		t.Errorf("want vscode, got %s", cfg.General.DefaultIDE)
	}
	if cfg.PortPool.Start != 22000 {
		t.Errorf("want 22000, got %d", cfg.PortPool.Start)
	}
	if len(cfg.PortPool.Used) != 2 {
		t.Errorf("want 2, got %d", len(cfg.PortPool.Used))
	}
}

func TestPortPool(t *testing.T) {
	pool := PortPool{
		Start: 22000,
		End:   22010,
		Used:  []int{22001, 22002, 22003},
	}

	if pool.Start != 22000 {
		t.Errorf("want 22000, got %d", pool.Start)
	}
	if pool.End != 22010 {
		t.Errorf("want 22010, got %d", pool.End)
	}
	if len(pool.Used) != 3 {
		t.Errorf("want 3, got %d", len(pool.Used))
	}
}

func TestWorkspaceEmpty(t *testing.T) {
	ws := Workspace{}

	if ws.Name != "" {
		t.Errorf("want empty, got %s", ws.Name)
	}
	if ws.UUID != "" {
		t.Errorf("want empty, got %s", ws.UUID)
	}
	if ws.Port != 0 {
		t.Errorf("want 0, got %d", ws.Port)
	}
}
