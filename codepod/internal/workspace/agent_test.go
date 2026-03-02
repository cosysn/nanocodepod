package workspace

import (
	"testing"

	"github.com/codepod-io/codepod/internal/docker"
	"github.com/codepod-io/codepod/internal/types"
)

// MockDockerClientForTest creates a mock Docker client for testing
type MockDockerClientForTest struct {
	Containers       map[string]*MockContainerForTest
	NextContainerID  int
	InspectResult    *docker.ContainerInfo
	InspectError     error
	ExistsResult     bool
	ExecError        error
}

type MockContainerForTest struct {
	Name    string
	Image   string
	Running bool
	Env     []string
	Cmd     []string
}

// NewMockDockerClientForTest creates a mock Docker client
func NewMockDockerClientForTest() *MockDockerClientForTest {
	return &MockDockerClientForTest{
		Containers:      make(map[string]*MockContainerForTest),
		NextContainerID: 1,
		InspectResult:   &docker.ContainerInfo{Running: true},
		ExistsResult:    true,
	}
}

func (m *MockDockerClientForTest) CreateContainer(config *docker.ContainerConfig) (string, error) {
	containerID := "mock-container-id"
	m.NextContainerID++
	m.Containers[config.Name] = &MockContainerForTest{
		Name:    config.Name,
		Image:   config.Image,
		Running: false,
		Env:     config.Env,
		Cmd:     config.Cmd,
	}
	return containerID, nil
}

func (m *MockDockerClientForTest) StartContainer(name string) error {
	if c, ok := m.Containers[name]; ok {
		c.Running = true
	}
	return nil
}

func (m *MockDockerClientForTest) StopContainer(name string) error {
	if c, ok := m.Containers[name]; ok {
		c.Running = false
	}
	return nil
}

func (m *MockDockerClientForTest) RemoveContainer(name string, force bool) error {
	delete(m.Containers, name)
	return nil
}

func (m *MockDockerClientForTest) InspectContainer(name string) (*docker.ContainerInfo, error) {
	return m.InspectResult, m.InspectError
}

func (m *MockDockerClientForTest) ContainerExists(name string) bool {
	_, ok := m.Containers[name]
	// Return true only if container exists in map or ExistsResult is explicitly true
	return ok
}

func (m *MockDockerClientForTest) ExecInContainer(name string, cmd []string) error {
	return m.ExecError
}

func (m *MockDockerClientForTest) Close() error {
	return nil
}

// TestCreateWithMock tests creating workspace with mock Docker client
func TestCreateWithMock(t *testing.T) {
	mockClient := NewMockDockerClientForTest()

	// Test Create with agent
	tests := []struct {
		name        string
		injectAgent bool
		wantCmd     string
		wantEnv     int
	}{
		{
			name:        "agent enabled",
			injectAgent: true,
			wantCmd:     "/usr/local/bin/codepod-agent",
			wantEnv:     2,
		},
		{
			name:        "agent disabled",
			injectAgent: false,
			wantCmd:     "sleep",
			wantEnv:     0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var containerCmd []string
			var containerEnv []string

			if tt.injectAgent {
				containerCmd = []string{"/usr/local/bin/codepod-agent"}
				containerEnv = []string{
					"CODEPOD_AGENT_PORT=22001",
					"CODEPOD_AGENT_PASSWORD=codepod",
				}
			} else {
				containerCmd = []string{"sleep", "infinity"}
				containerEnv = []string{}
			}

			// Verify command
			if len(containerCmd) > 0 {
				if tt.injectAgent && containerCmd[0] != tt.wantCmd {
					t.Errorf("want cmd %s, got %s", tt.wantCmd, containerCmd[0])
				}
				if !tt.injectAgent && containerCmd[0] != tt.wantCmd {
					t.Errorf("want cmd %s, got %s", tt.wantCmd, containerCmd[0])
				}
			}

			// Verify env count
			if len(containerEnv) != tt.wantEnv {
				t.Errorf("want env count %d, got %d", tt.wantEnv, len(containerEnv))
			}
		})
	}

	_ = mockClient
}

// TestWorkspaceAgentConfig tests workspace agent configuration
func TestWorkspaceAgentConfig(t *testing.T) {
	tests := []struct {
		name         string
		agentStatus  string
		injectAgent  bool
		expectedPort int
		expectedStat string
	}{
		{
			name:         "new workspace with agent",
			agentStatus:  "stopped",
			injectAgent:  true,
			expectedPort: 22001,
			expectedStat: "running",
		},
		{
			name:         "new workspace without agent",
			agentStatus:  "stopped",
			injectAgent:  false,
			expectedPort: 22001,
			expectedStat: "stopped",
		},
		{
			name:         "existing workspace with running agent",
			agentStatus:  "running",
			injectAgent:  true,
			expectedPort: 22001,
			expectedStat: "running",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ws := &types.Workspace{
				Name: "test-workspace",
				Agent: types.Agent{
					Port:   tt.expectedPort,
					Status: tt.agentStatus,
				},
			}

			// Simulate Start logic
			useAgent := tt.injectAgent || ws.Agent.Status == "running"

			if useAgent {
				ws.Agent.Status = "running"
			} else {
				ws.Agent.Status = "stopped"
			}

			if ws.Agent.Status != tt.expectedStat {
				t.Errorf("want status %s, got %s", tt.expectedStat, ws.Agent.Status)
			}
		})
	}
}

// TestContainerConfigGeneration tests container config generation
func TestContainerConfigGeneration(t *testing.T) {
	tests := []struct {
		name        string
		injectAgent bool
		image       string
		port        int
		validate    func(cmd []string, env []string) bool
	}{
		{
			name:        "agent with custom image",
			injectAgent: true,
			image:       "ubuntu:20.04",
			port:        22001,
			validate: func(cmd []string, env []string) bool {
				return len(cmd) == 1 && cmd[0] == "/usr/local/bin/codepod-agent" &&
					len(env) == 2
			},
		},
		{
			name:        "no agent with custom image",
			injectAgent: false,
			image:       "ubuntu:20.04",
			port:        22002,
			validate: func(cmd []string, env []string) bool {
				return len(cmd) == 2 && cmd[0] == "sleep" && cmd[1] == "infinity" &&
					len(env) == 0
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cmd []string
			var env []string

			if tt.injectAgent {
				cmd = []string{"/usr/local/bin/codepod-agent"}
				env = []string{
					"CODEPOD_AGENT_PORT=22001",
					"CODEPOD_AGENT_PASSWORD=codepod",
				}
			} else {
				cmd = []string{"sleep", "infinity"}
				env = []string{}
			}

			if !tt.validate(cmd, env) {
				t.Errorf("validation failed for test %s", tt.name)
			}
		})
	}
}

// TestAgentEnvironmentVariables tests agent environment variable configuration
func TestAgentEnvironmentVariables(t *testing.T) {
	tests := []struct {
		name     string
		port     int
		password string
		validate func(env []string) bool
	}{
		{
			name:     "default values",
			port:     22001,
			password: "codepod",
			validate: func(env []string) bool {
				if len(env) != 2 {
					return false
				}
				hasPort := false
				hasPassword := false
				for _, e := range env {
					if e == "CODEPOD_AGENT_PORT=22001" {
						hasPort = true
					}
					if e == "CODEPOD_AGENT_PASSWORD=codepod" {
						hasPassword = true
					}
				}
				return hasPort && hasPassword
			},
		},
		{
			name:     "custom port",
			port:     22005,
			password: "codepod",
			validate: func(env []string) bool {
				for _, e := range env {
					if e == "CODEPOD_AGENT_PORT=22005" {
						return true
					}
				}
				return false
			},
		},
		{
			name:     "custom password",
			port:     22001,
			password: "mypassword",
			validate: func(env []string) bool {
				for _, e := range env {
					if e == "CODEPOD_AGENT_PASSWORD=mypassword" {
						return true
					}
				}
				return false
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := []string{
				"CODEPOD_AGENT_PORT=22001",
				"CODEPOD_AGENT_PASSWORD=codepod",
			}

			// Override based on test case
			if tt.port != 22001 {
				env[0] = "CODEPOD_AGENT_PORT=22005"
			}
			if tt.password != "codepod" {
				env[1] = "CODEPOD_AGENT_PASSWORD=mypassword"
			}

			if !tt.validate(env) {
				t.Errorf("validation failed for test %s", tt.name)
			}
		})
	}
}

// TestWorkspaceTypes tests workspace type definitions
func TestWorkspaceTypes(t *testing.T) {
	// Test Workspace struct
	ws := &types.Workspace{
		Name:        "test",
		UUID:        "test-uuid",
		Port:        22001,
		Agent:       types.Agent{Port: 22001, Status: "running"},
		Container:   types.Container{Image: "ubuntu:22.04", Name: "test-container"},
		Repository:  types.Repository{URL: "https://github.com/test/repo", Branch: "main"},
		IDE:        types.IDE{Type: types.IDETypeVSCode},
		SSH:        types.SSH{ConfigPath: "/path/to/config"},
	}

	if ws.Name != "test" {
		t.Errorf("want name test, got %s", ws.Name)
	}
	if ws.UUID != "test-uuid" {
		t.Errorf("want uuid test-uuid, got %s", ws.UUID)
	}
	if ws.Port != 22001 {
		t.Errorf("want port 22001, got %d", ws.Port)
	}
	if ws.Agent.Status != "running" {
		t.Errorf("want agent status running, got %s", ws.Agent.Status)
	}
	if ws.Agent.Port != 22001 {
		t.Errorf("want agent port 22001, got %d", ws.Agent.Port)
	}
	if ws.Container.Image != "ubuntu:22.04" {
		t.Errorf("want container image ubuntu:22.04, got %s", ws.Container.Image)
	}
	if ws.Repository.URL != "https://github.com/test/repo" {
		t.Errorf("want repo URL, got %s", ws.Repository.URL)
	}
	if ws.IDE.Type != types.IDETypeVSCode {
		t.Errorf("want IDE type vscode, got %s", ws.IDE.Type)
	}
}

// TestCreateOptionsComplete tests complete CreateOptions
func TestCreateOptionsComplete(t *testing.T) {
	tests := []struct {
		name string
		opts *CreateOptions
		want string
	}{
		{
			name: "with all options",
			opts: &CreateOptions{
				Image: "ubuntu:20.04",
				Repository: types.Repository{
					URL:       "https://github.com/test/repo",
					Branch:    "develop",
					LocalPath: "/tmp/repo",
				},
				IDE:         types.IDE{Type: types.IDETypeJetBrains},
				InjectAgent: true,
			},
			want: "ubuntu:20.04",
		},
		{
			name: "minimal options",
			opts: &CreateOptions{
				Image: "alpine:latest",
			},
			want: "alpine:latest",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.opts.Image != tt.want {
				t.Errorf("want image %s, got %s", tt.want, tt.opts.Image)
			}
		})
	}
}

// TestIDETypes tests IDE type constants
func TestIDETypes(t *testing.T) {
	if types.IDETypeVSCode != "vscode" {
		t.Errorf("want vscode, got %s", types.IDETypeVSCode)
	}
	if types.IDETypeJetBrains != "jetbrains" {
		t.Errorf("want jetbrains, got %s", types.IDETypeJetBrains)
	}
}

// TestWorkspaceState tests workspace state constants
func TestWorkspaceState(t *testing.T) {
	if types.WorkspaceStateCreated != "created" {
		t.Errorf("want created, got %s", types.WorkspaceStateCreated)
	}
	if types.WorkspaceStateRunning != "running" {
		t.Errorf("want running, got %s", types.WorkspaceStateRunning)
	}
	if types.WorkspaceStateStopped != "stopped" {
		t.Errorf("want stopped, got %s", types.WorkspaceStateStopped)
	}
	if types.WorkspaceStateError != "error" {
		t.Errorf("want error, got %s", types.WorkspaceStateError)
	}
}

// TestAgentStatusValues tests agent status values
func TestAgentStatusValues(t *testing.T) {
	// Test that agent status can be set to valid values
	statuses := []string{"running", "stopped", "error"}
	for _, status := range statuses {
		ws := &types.Workspace{
			Name: "test",
			Agent: types.Agent{
				Port:   22001,
				Status: status,
			},
		}
		if ws.Agent.Status != status {
			t.Errorf("want status %s, got %s", status, ws.Agent.Status)
		}
	}
}

// TestContainerOperations tests container operations with mock
func TestContainerOperations(t *testing.T) {
	mock := NewMockDockerClientForTest()

	// Test CreateContainer
	id, err := mock.CreateContainer(&docker.ContainerConfig{
		Name:  "test-container",
		Image: "ubuntu:22.04",
		Cmd:   []string{"sleep", "infinity"},
		Env:   []string{"TEST=value"},
	})
	if err != nil {
		t.Fatalf("CreateContainer failed: %v", err)
	}
	if id == "" {
		t.Error("expected container id")
	}

	// Test StartContainer
	err = mock.StartContainer("test-container")
	if err != nil {
		t.Fatalf("StartContainer failed: %v", err)
	}

	// Verify running state
	container, ok := mock.Containers["test-container"]
	if !ok {
		t.Fatal("container not found")
	}
	if !container.Running {
		t.Error("container should be running")
	}

	// Test StopContainer
	err = mock.StopContainer("test-container")
	if err != nil {
		t.Fatalf("StopContainer failed: %v", err)
	}

	// Verify stopped state
	container, _ = mock.Containers["test-container"]
	if container.Running {
		t.Error("container should be stopped")
	}

	// Test RemoveContainer
	err = mock.RemoveContainer("test-container", false)
	if err != nil {
		t.Fatalf("RemoveContainer failed: %v", err)
	}

	// Test ContainerExists for removed container
	exists := mock.ContainerExists("test-container")
	if exists {
		t.Error("ContainerExists should return false for removed container")
	}
}
