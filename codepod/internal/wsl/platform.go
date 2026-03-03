package wsl

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/codepod-io/codepod/internal/config"
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

// PlatformInterface defines the interface for platform operations
type PlatformInterface interface {
	GetType() PlatformType
	GetDistribution() string
	GetHostname() (string, error)
	RunCommand(cmd string) (string, error)
	FileExists(path string) bool
	CopyToWSL(src, dest string) error
	CopyFromWSL(src, dest string) error
}

// Platform provides platform-specific operations
type Platform struct {
	Type PlatformType
}

// Ensure Platform implements PlatformInterface
var _ PlatformInterface = (*Platform)(nil)

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
		distro := GetWSLDistributionFromConfig()
		wsl := New(distro)
		return wsl.RunCommand(cmd)
	case PlatformLinux:
		out, err := exec.Command("bash", "-c", cmd).Output()
		if err != nil {
			return "", fmt.Errorf("failed to run command: %w", err)
		}
		return strings.TrimSpace(string(out)), nil
	case PlatformWindows:
		// On Windows, run command in WSL
		distro := GetWSLDistributionFromConfig()
		wsl := New(distro)
		return wsl.RunCommand(cmd)
	default:
		return "", fmt.Errorf("unsupported platform: %s", p.Type)
	}
}

// GetHostname returns the hostname based on platform
func (p *Platform) GetHostname() (string, error) {
	switch p.Type {
	case PlatformWSL:
		distro := GetWSLDistributionFromConfig()
		return distro, nil
	case PlatformLinux:
		cmd := exec.Command("hostname")
		out, err := cmd.Output()
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(string(out)), nil
	case PlatformWindows:
		// On Windows, return WSL distribution name
		return GetWSLDistributionFromConfig(), nil
	default:
		return "", fmt.Errorf("unsupported platform")
	}
}

// FileExists checks if a file exists
func (p *Platform) FileExists(path string) bool {
	switch p.Type {
	case PlatformWSL:
		distro := GetWSLDistributionFromConfig()
		wsl := New(distro)
		return wsl.FileExists(path)
	case PlatformLinux:
		_, err := os.Stat(path)
		return err == nil
	case PlatformWindows:
		// On Windows, check file in WSL
		distro := GetWSLDistributionFromConfig()
		wsl := New(distro)
		return wsl.FileExists(path)
	default:
		return false
	}
}

// GetDistribution returns the WSL distribution name
func (p *Platform) GetDistribution() string {
	distro := os.Getenv("WSL_DISTRO_NAME")
	if distro == "" {
		distro = "Ubuntu"
	}
	return distro
}

// CopyToWSL copies a file to WSL
func (p *Platform) CopyToWSL(src, dest string) error {
	if p.Type == PlatformWSL {
		wsl := New(p.GetDistribution())
		return wsl.CopyToWSL(src, dest)
	}
	return nil
}

// CopyFromWSL copies a file from WSL
func (p *Platform) CopyFromWSL(src, dest string) error {
	if p.Type == PlatformWSL {
		wsl := New(p.GetDistribution())
		return wsl.CopyFromWSL(src, dest)
	}
	return nil
}

// GetType returns the platform type
func (p *Platform) GetType() PlatformType {
	return p.Type
}

// DockerAccessMode represents how Docker can be accessed
type DockerAccessMode string

const (
	DockerAccessNative    DockerAccessMode = "native"   // Docker available directly (Linux or Windows with Docker Desktop)
	DockerAccessWSL       DockerAccessMode = "wsl"      // Docker only available in WSL
	DockerAccessNone      DockerAccessMode = "none"     // Docker not available
)

// DetectDockerAccessMode detects how Docker can be accessed on the current system
func DetectDockerAccessMode() DockerAccessMode {
	platform := DetectPlatform()

	switch platform {
	case PlatformLinux, PlatformWSL:
		// On Linux or inside WSL, use Docker directly
		if IsDockerAvailable() {
			return DockerAccessNative
		}
		return DockerAccessNone
	case PlatformWindows:
		// On Windows, try native Docker first
		if IsDockerAvailable() {
			return DockerAccessNative
		}
		// Fall back to checking WSL
		if IsDockerAvailableInWSL() {
			return DockerAccessWSL
		}
		return DockerAccessNone
	default:
		if IsDockerAvailable() {
			return DockerAccessNative
		}
		return DockerAccessNone
	}
}

// GetDockerAccessModeDebug returns detailed debug info about Docker detection
func GetDockerAccessModeDebug() string {
	platform := DetectPlatform()
	result := fmt.Sprintf("Platform: %s\n", platform)
	result += fmt.Sprintf("runtime.GOOS: %s\n", runtime.GOOS)

	// Check native Docker
	nativeAvailable := IsDockerAvailable()
	result += fmt.Sprintf("Native Docker (docker info): %v\n", nativeAvailable)

	// Check WSL Docker
	if platform == PlatformWindows {
		result += IsDockerAvailableInWSLDebug()
		wslDocker := IsDockerAvailableInWSL()
		result += fmt.Sprintf("WSL Docker: %v\n", wslDocker)
	}

	result += fmt.Sprintf("Detected Access Mode: %s\n", DetectDockerAccessMode())
	return result
}

// GetWSLDistributionFromConfig returns the WSL distribution name from config
func GetWSLDistributionFromConfig() string {
	// Try to load config
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("[DEBUG] LoadConfig error: %v\n", err)
	} else if cfg == nil {
		fmt.Printf("[DEBUG] LoadConfig returned nil\n")
	} else {
		fmt.Printf("[DEBUG] LoadConfig success, WSL.Distribution='%s', DataDir='%s'\n", cfg.WSL.Distribution, cfg.DataDir)
		if cfg.WSL.Distribution != "" {
			return cfg.WSL.Distribution
		}
	}
	// Fallback to environment variable
	distro := os.Getenv("WSL_DISTRO_NAME")
	if distro != "" {
		fmt.Printf("[DEBUG] Using WSL distribution from env: %s\n", distro)
		return distro
	}
	// Fallback to default
	fmt.Printf("[DEBUG] Using WSL distribution default: Ubuntu-22.04\n")
	return "Ubuntu-22.04"
}

// IsDockerAvailableInWSL checks if Docker is available in any WSL distribution
func IsDockerAvailableInWSL() bool {
	// Get WSL distribution from config
	distro := GetWSLDistributionFromConfig()

	// Try to run docker info in WSL
	wslInstance := New(distro)
	_, err := wslInstance.RunCommand("docker info")
	return err == nil
}

// IsDockerAvailableInWSLDebug returns detailed debug info for WSL Docker check
func IsDockerAvailableInWSLDebug() string {
	result := ""
	distro := GetWSLDistributionFromConfig()
	result += fmt.Sprintf("WSL Distribution (from config): %s\n", distro)

	wslInstance := New(distro)
	output, err := wslInstance.RunCommand("docker info")
	result += fmt.Sprintf("docker info output: %s\n", output)
	result += fmt.Sprintf("docker info error: %v\n", err)
	return result
}

// GetWSLDistributionWithDocker returns the WSL distribution name that has Docker available
func GetWSLDistributionWithDocker() (string, error) {
	// Get WSL distribution from config first
	distro := GetWSLDistributionFromConfig()

	wsl := New(distro)
	if wsl.IsDockerRunning() {
		return distro, nil
	}

	// If config distribution doesn't have Docker, try to find one that does
	distros, err := ListDistributions()
	if err == nil {
		for _, d := range distros {
			wslTest := New(d)
			if wslTest.IsDockerRunning() {
				return d, nil
			}
		}
	}

	return "", fmt.Errorf("no WSL distribution with Docker found")
}

// WindowsPathToWSLPath returns the path as-is for WSL
// If user configures WSL path (e.g., /data/.codepod), use it directly
func WindowsPathToWSLPath(windowsPath string) string {
	return windowsPath
}
