package provider

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/codepod-io/codepod/internal/wsl"
	"gopkg.in/yaml.v3"
)

// Provider interface defines operations for different environment providers
type Provider interface {
	// Init initializes the environment (installs server, agent, sets up container)
	Init() error

	// Command executes a command in the target environment
	Command(cmd string) (string, error)

	// CommandWithStdin executes a command with stdin input
	CommandWithStdin(cmd string, stdin string) (string, error)

	// Create creates the target environment
	Create() error

	// Delete deletes the target environment
	Delete() error

	// Start starts the target environment
	Start() error

	// Stop stops the target environment
	Stop() error

	// Status returns the status of the target environment
	Status() (string, error)
}

// Config holds provider configuration
type Config struct {
	Type      string `yaml:"type"`
	WSLDistro string `yaml:"wsl_distro,omitempty"`
	DataDir   string `yaml:"data_dir,omitempty"`
	SocketPath string `yaml:"socket_path,omitempty"`
}

// LoadConfig loads provider configuration from ~/.codepod/provider/<name>/config.yaml
func LoadConfig(name string) (*Config, error) {
	configPath := filepath.Join(os.Getenv("USERPROFILE"), ".codepod", "provider", name, "config.yaml")

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// Unmarshal parses YAML configuration
func Unmarshal(data []byte, v interface{}) error {
	return yaml.Unmarshal(data, v)
}

// WSLProvider implements Provider for WSL environments
type WSLProvider struct {
	distro string
}

// NewWSLProvider creates a new WSL provider
func NewWSLProvider(distro string) *WSLProvider {
	return &WSLProvider{
		distro: distro,
	}
}

// Init initializes the WSL environment
func (p *WSLProvider) Init() error {
	// This would trigger the deployment script
	// The script downloads and installs server/agent
	return nil
}

// Command executes a command in WSL
func (p *WSLProvider) Command(cmd string) (string, error) {
	args := []string{"-d", p.distro, "--", "bash", "-c", cmd}
	out, err := exec.Command("wsl.exe", args...).Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// CommandWithStdin executes a command with stdin
func (p *WSLProvider) CommandWithStdin(cmd string, stdin string) (string, error) {
	args := []string{"-d", p.distro, "--", "bash", "-c", cmd}
	proc := exec.Command("wsl.exe", args...)
	proc.Stdin = os.Stdin
	out, err := proc.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// Create is not applicable for WSL (distribution already exists)
func (p *WSLProvider) Create() error {
	return nil
}

// Delete is not applicable for WSL
func (p *WSLProvider) Delete() error {
	return nil
}

// Start starts the WSL distribution
func (p *WSLProvider) Start() error {
	args := []string{"-d", p.distro}
	return exec.Command("wsl.exe", args...).Run()
}

// Stop stops the WSL distribution
func (p *WSLProvider) Stop() error {
	args := []string{"-d", p.distro, "-t"}
	return exec.Command("wsl.exe", args...).Run()
}

// Status returns the status of the WSL distribution
func (p *WSLProvider) Status() (string, error) {
	out, err := exec.Command("wsl.exe", "-l", "-v").Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// LocalProvider implements Provider for local Linux/macOS environments
type LocalProvider struct {
	dataDir   string
	socketPath string
}

// NewLocalProvider creates a new local provider
func NewLocalProvider(dataDir, socketPath string) *LocalProvider {
	return &LocalProvider{
		dataDir:   dataDir,
		socketPath: socketPath,
	}
}

// Init initializes the local environment
func (p *LocalProvider) Init() error {
	// Check if Docker is available
	if !wsl.IsDockerAvailable() {
		return fmt.Errorf("Docker is not available. Please ensure Docker is installed and running")
	}
	return nil
}

// Command executes a command locally
func (p *LocalProvider) Command(cmd string) (string, error) {
	out, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// CommandWithStdin executes a command with stdin
func (p *LocalProvider) CommandWithStdin(cmd string, stdin string) (string, error) {
	proc := exec.Command("bash", "-c", cmd)
	proc.Stdin = os.Stdin
	out, err := proc.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// Create is not applicable for local
func (p *LocalProvider) Create() error {
	return nil
}

// Delete is not applicable for local
func (p *LocalProvider) Delete() error {
	return nil
}

// Start starts the local server
func (p *LocalProvider) Start() error {
	// Start server in background
	return nil
}

// Stop stops the local server
func (p *LocalProvider) Stop() error {
	// Stop server
	return nil
}

// Status returns the status of the local environment
func (p *LocalProvider) Status() (string, error) {
	// Check if Docker is available
	if !wsl.IsDockerAvailable() {
		return "docker unavailable", fmt.Errorf("Docker is not available")
	}

	// Check if Docker daemon is running by running docker info
	cmd := exec.Command("docker", "info")
	if err := cmd.Run(); err != nil {
		return "docker not running", fmt.Errorf("Docker daemon is not running: %w", err)
	}

	return "running", nil
}
