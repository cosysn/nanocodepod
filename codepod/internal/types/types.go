package types

import "time"

// WorkspaceState represents the state of a workspace
type WorkspaceState string

const (
	WorkspaceStateCreated  WorkspaceState = "created"
	WorkspaceStateRunning WorkspaceState = "running"
	WorkspaceStateStopped WorkspaceState = "stopped"
	WorkspaceStateError   WorkspaceState = "error"
)

// IDEType represents the type of IDE
type IDEType string

const (
	IDETypeVSCode    IDEType = "vscode"
	IDETypeJetBrains IDEType = "jetbrains"
)

// Workspace represents a development workspace
type Workspace struct {
	Name        string         `yaml:"name"`
	UUID        string         `yaml:"uuid"`
	CreatedAt   time.Time      `yaml:"created_at"`
	UpdatedAt   time.Time      `yaml:"updated_at"`
	State       WorkspaceState `yaml:"state"`
	Repository  Repository     `yaml:"repository"`
	IDE         IDE            `yaml:"ide"`
	Container   Container      `yaml:"container"`
	Domain      string         `yaml:"domain"`
	SSH         SSH            `yaml:"ssh"`
	Agent       Agent          `yaml:"agent"`
	Port        int            `yaml:"port"`
	StoragePath string         `yaml:"storage_path"`
	CodePath    string         `yaml:"code_path"`
}

// Repository represents a Git repository or local directory
type Repository struct {
	URL       string `yaml:"url"`
	Branch    string `yaml:"branch"`
	LocalPath string `yaml:"local_path"`
}

// IDE represents IDE configuration
type IDE struct {
	Type     IDEType            `yaml:"type"`
	Settings map[string]string   `yaml:"settings"`
}

// Container represents container configuration
type Container struct {
	Image string `yaml:"image"`
	Name  string `yaml:"name"`
}

// SSH represents SSH configuration
type SSH struct {
	ConfigPath string `yaml:"config_path"`
	KeyPath    string `yaml:"key_path"`
}

// Agent represents agent configuration
type Agent struct {
	Port   int            `yaml:"port"`
	Status string         `yaml:"status"`
}

// Config represents the global configuration
type Config struct {
	Version     string        `yaml:"version"`
	WSL         WSLConfig     `yaml:"wsl"`
	General     GeneralConfig `yaml:"general"`
	PortPool    PortPool     `yaml:"port_pool"`
	DataDir     string       `yaml:"data_dir"` // Data directory for workspaces and agent
}

// WSLConfig represents WSL configuration
type WSLConfig struct {
	Distribution string `yaml:"distribution"`
	DockerHost   string `yaml:"docker_host"`
}

// GeneralConfig represents general configuration
type GeneralConfig struct {
	DefaultIDE string `yaml:"default_ide"`
	SSHPort    int    `yaml:"ssh_port"`
}

// PortPool represents the port pool for allocation
type PortPool struct {
	Start  int `yaml:"start"`
	End    int `yaml:"end"`
	Used   []int `yaml:"used"`
}
