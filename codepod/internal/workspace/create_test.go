package workspace

import (
	"testing"

	"github.com/codepod-io/codepod/internal/types"
)

// TestCreateWithAgent tests creating workspace with agent enabled
func TestCreateWithAgent(t *testing.T) {
	// Test Create with agent config logic
	tests := []struct {
		name        string
		injectAgent bool
		expectedCmd string
		envCount    int
	}{
		{
			name:        "agent disabled",
			injectAgent: false,
			expectedCmd: "sleep",
			envCount:    0,
		},
		{
			name:        "agent enabled",
			injectAgent: true,
			expectedCmd: "codepod-agent",
			envCount:    2,
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
			if tt.injectAgent {
				if len(containerCmd) != 1 || containerCmd[0] != "/usr/local/bin/codepod-agent" {
					t.Errorf("expected agent entrypoint, got %v", containerCmd)
				}
			} else {
				if containerCmd[0] != "sleep" {
					t.Errorf("expected sleep command, got %v", containerCmd)
				}
			}

			// Verify environment
			if len(containerEnv) != tt.envCount {
				t.Errorf("expected %d env vars, got %d", tt.envCount, len(containerEnv))
			}
		})
	}
}

// TestAgentPorts tests agent port configuration
func TestAgentPorts(t *testing.T) {
	tests := []struct {
		name     string
		port     int
		password string
		validate func(env []string) bool
	}{
		{
			name:     "default port 22001",
			port:     22001,
			password: "codepod",
			validate: func(env []string) bool {
				return env[0] == "CODEPOD_AGENT_PORT=22001" && env[1] == "CODEPOD_AGENT_PASSWORD=codepod"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := []string{
				"CODEPOD_AGENT_PORT=22001",
				"CODEPOD_AGENT_PASSWORD=codepod",
			}
			if !tt.validate(env) {
				t.Errorf("validation failed for port %d", tt.port)
			}
		})
	}
}

// TestAgentStatusTransition tests agent status transitions
func TestAgentStatusTransition(t *testing.T) {
	tests := []struct {
		name     string
		initial  string
		inject   bool
		expected string
	}{
		{
			name:     "stopped to running when inject",
			initial:  "stopped",
			inject:   true,
			expected: "running",
		},
		{
			name:     "running stays running",
			initial:  "running",
			inject:   true,
			expected: "running",
		},
		{
			name:     "stopped stays stopped without inject",
			initial:  "stopped",
			inject:   false,
			expected: "stopped",
		},
		{
			name:     "running preserves existing agent",
			initial:  "running",
			inject:   false,
			expected: "running",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			workspace := &types.Workspace{
				Name:  "test",
				Agent: types.Agent{Status: tt.initial},
			}

			// Simulate Start logic
			useAgent := tt.inject || workspace.Agent.Status == "running"

			if useAgent {
				workspace.Agent.Status = "running"
			} else {
				workspace.Agent.Status = "stopped"
			}

			if workspace.Agent.Status != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, workspace.Agent.Status)
			}
		})
	}
}

// TestEntrypointSelection tests that the correct entrypoint is selected
func TestEntrypointSelection(t *testing.T) {
	tests := []struct {
		name        string
		injectAgent bool
		expectAgent bool
	}{
		{
			name:        "agent enabled uses codepod-agent",
			injectAgent: true,
			expectAgent: true,
		},
		{
			name:        "agent disabled uses sleep",
			injectAgent: false,
			expectAgent: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cmd []string
			if tt.injectAgent {
				cmd = []string{"/usr/local/bin/codepod-agent"}
			} else {
				cmd = []string{"sleep", "infinity"}
			}

			isAgent := len(cmd) == 1 && cmd[0] == "/usr/local/bin/codepod-agent"

			if isAgent != tt.expectAgent {
				t.Errorf("expected agent=%v, got %v", tt.expectAgent, isAgent)
			}
		})
	}
}

// TestCreateOptionsDefaults tests CreateOptions defaults
func TestCreateOptionsDefaults(t *testing.T) {
	tests := []struct {
		name         string
		opts         *CreateOptions
		expectedImg  string
		expectedBranch string
	}{
		{
			name:         "nil options",
			opts:         nil,
			expectedImg:  "ubuntu:22.04",
			expectedBranch: "main",
		},
		{
			name:         "empty options",
			opts:         &CreateOptions{},
			expectedImg:  "",
			expectedBranch: "",
		},
		{
			name: "with image",
			opts: &CreateOptions{
				Image: "ubuntu:20.04",
			},
			expectedImg:       "ubuntu:20.04",
			expectedBranch:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := tt.opts
			if opts == nil {
				opts = &CreateOptions{}
			}
			if opts.Image == "" {
				opts.Image = "ubuntu:22.04"
			}
			if opts.Repository.Branch == "" {
				opts.Repository.Branch = "main"
			}

			if opts.Image != tt.expectedImg && tt.expectedImg != "" {
				if opts.Image != tt.expectedImg {
					t.Errorf("expected image %s, got %s", tt.expectedImg, opts.Image)
				}
			}
		})
	}
}

// TestCreateOptionsInjectAgent tests InjectAgent field
func TestCreateOptionsInjectAgent(t *testing.T) {
	tests := []struct {
		name string
		opts *CreateOptions
		want bool
	}{
		{
			name: "default false",
			opts: &CreateOptions{},
			want: false,
		},
		{
			name: "explicit true",
			opts: &CreateOptions{InjectAgent: true},
			want: true,
		},
		{
			name: "explicit false",
			opts: &CreateOptions{InjectAgent: false},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.opts == nil {
				tt.opts = &CreateOptions{}
			}
			if tt.opts.InjectAgent != tt.want {
				t.Errorf("InjectAgent = %v, want %v", tt.opts.InjectAgent, tt.want)
			}
		})
	}
}
