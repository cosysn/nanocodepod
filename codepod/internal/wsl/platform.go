package wsl

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// PlatformType represents the platform type
type PlatformType string

const (
	PlatformLinux PlatformType = "linux"
	PlatformWSL   PlatformType = "wsl"
	PlatformWindows PlatformType = "windows"
)

// DetectPlatform detects the current platform
func DetectPlatform() PlatformType {
	// Check if we're on WSL
	if isWSL() {
		return PlatformWSL
	}

	// Check if we're on Windows
	if runtime.GOOS == "windows" {
		return PlatformWindows
	}

	return PlatformLinux
}

// isWSL checks if running in WSL
func isWSL() bool {
	// Check /proc/version for WSL signature
	data, err := os.ReadFile("/proc/version")
	if err == nil {
		if strings.Contains(strings.ToLower(string(data)), "microsoft") {
			return true
		}
	}

	// Check WS_DISTRIBUTION_NAME environment variable
	if os.Getenv("WSL_DISTRO_NAME") != "" {
		return true
	}

	return false
}

// IsDockerAvailable checks if Docker is available
func IsDockerAvailable() bool {
	cmd := exec.Command("docker", "info")
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}

// GetDockerHost returns the appropriate Docker host
func GetDockerHost(platform PlatformType) string {
	switch platform {
	case PlatformWSL:
		return "tcp://localhost:2375"
	case PlatformLinux:
		return "unix:///var/run/docker.sock"
	default:
		return "tcp://localhost:2375"
	}
}

// Platform provides platform-specific operations
type Platform struct {
	Type PlatformType
}

// NewPlatform creates a new platform handler
func NewPlatform() (*Platform, error) {
	platform := DetectPlatform()
	return &Platform{Type: platform}, nil
}

// RunCommand runs a command based on platform
func (p *Platform) RunCommand(cmd string) (string, error) {
	switch p.Type {
	case PlatformWSL:
		// Try to detect WSL distribution
		distro := os.Getenv("WSL_DISTRO_NAME")
		if distro == "" {
			distro = "Ubuntu"
		}
		wsl := New(distro)
		return wsl.RunCommand(cmd)
	case PlatformLinux:
		out, err := exec.Command("bash", "-c", cmd).Output()
		if err != nil {
			return "", fmt.Errorf("failed to run command: %w", err)
		}
		return strings.TrimSpace(string(out)), nil
	default:
		return "", fmt.Errorf("unsupported platform: %s", p.Type)
	}
}

// GetHostname returns the hostname based on platform
func (p *Platform) GetHostname() (string, error) {
	switch p.Type {
	case PlatformWSL:
		distro := os.Getenv("WSL_DISTRO_NAME")
		if distro == "" {
			distro = "Ubuntu"
		}
		return distro, nil
	case PlatformLinux:
		cmd := exec.Command("hostname")
		out, err := cmd.Output()
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(string(out)), nil
	default:
		return "", fmt.Errorf("unsupported platform")
	}
}

// FileExists checks if a file exists
func (p *Platform) FileExists(path string) bool {
	switch p.Type {
	case PlatformWSL:
		distro := os.Getenv("WSL_DISTRO_NAME")
		if distro == "" {
			distro = "Ubuntu"
		}
		wsl := New(distro)
		return wsl.FileExists(path)
	case PlatformLinux:
		_, err := os.Stat(path)
		return err == nil
	default:
		return false
	}
}
