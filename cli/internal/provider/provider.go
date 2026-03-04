package provider

import (
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Provider interface defines operations for different environment providers
type Provider interface {
	// Init initializes the environment
	Init() error

	// Command executes a command in the target environment
	Command(cmd string) (string, error)

	// Status returns the status of the target environment
	Status() (string, error)

	// DiscoverServer finds and returns server information
	DiscoverServer() (*ServerInfo, error)

	// GetServerURL returns the server URL
	GetServerURL() string
}

// ServerInfo holds server connection information
type ServerInfo struct {
	URL    string
	Status string
}

// Config holds provider configuration
type Config struct {
	Type      string `yaml:"type"`
	WSLDistro string `yaml:"wsl_distro,omitempty"`
	DataDir   string `yaml:"data_dir,omitempty"`
	SocketPath string `yaml:"socket_path,omitempty"`
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

// Status returns the status of the WSL distribution
func (p *WSLProvider) Status() (string, error) {
	out, err := exec.Command("wsl.exe", "-l", "-v").Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// DiscoverServer finds the server in WSL
func (p *WSLProvider) DiscoverServer() (*ServerInfo, error) {
	cmd := "cat /tmp/codepod-server-port 2>/dev/null || echo \"\""
	out, err := p.Command(cmd)
	if err != nil {
		return &ServerInfo{Status: "not_found"}, nil
	}

	port := strings.TrimSpace(out)
	if port == "" {
		return &ServerInfo{Status: "not_found"}, nil
	}

	url := "http://localhost:" + port + "/health"
	resp, err := http.Get(url)
	if err != nil {
		return &ServerInfo{Status: "not_running", URL: url}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return &ServerInfo{Status: "running", URL: url}, nil
	}

	return &ServerInfo{Status: "error", URL: url}, nil
}

// GetServerURL returns the server URL
func (p *WSLProvider) GetServerURL() string {
	info, _ := p.DiscoverServer()
	return info.URL
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

// Status returns the status of the local environment
func (p *LocalProvider) Status() (string, error) {
	return "running", nil
}

// DiscoverServer finds the server locally
func (p *LocalProvider) DiscoverServer() (*ServerInfo, error) {
	portFile := "/tmp/codepod-server-port"
	data, err := os.ReadFile(portFile)
	if err != nil {
		return &ServerInfo{Status: "not_found"}, nil
	}

	port := strings.TrimSpace(string(data))
	if port == "" {
		return &ServerInfo{Status: "not_found"}, nil
	}

	url := "http://localhost:" + port + "/health"
	resp, err := http.Get(url)
	if err != nil {
		return &ServerInfo{Status: "not_running", URL: url}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return &ServerInfo{Status: "running", URL: url}, nil
	}

	return &ServerInfo{Status: "error", URL: url}, nil
}

// GetServerURL returns the server URL
func (p *LocalProvider) GetServerURL() string {
	info, _ := p.DiscoverServer()
	return info.URL
}

// GetProvider returns the appropriate provider based on environment
func GetProvider() Provider {
	// Check if running in WSL
	if isWSL() {
		return NewWSLProvider("Ubuntu-22.04")
	}

	// Default to local provider
	return NewLocalProvider("/tmp/codepod", "")
}

func isWSL() bool {
	// Check if running on Windows
	out, err := exec.Command("uname", "-r").Output()
	if err != nil {
		return false
	}
	return strings.Contains(string(out), "microsoft")
}

// LoadConfig loads provider configuration
func LoadConfig(name string) (*Config, error) {
	configPath := filepath.Join(os.Getenv("USERPROFILE"), ".codepod", "provider", name, "config.yaml")

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
