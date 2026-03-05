package embed

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// DetectTargetArchitecture detects the architecture of the target environment.
func DetectTargetArchitecture(ctx context.Context, targetType, target string) (Platform, error) {
	switch targetType {
	case "wsl":
		return detectWSLArchitecture(ctx, target)
	case "docker":
		return detectDockerArchitecture(ctx, target)
	case "ssh":
		return detectSSHArchitecture(ctx, target)
	default:
		// Default to current platform
		return GetCurrentPlatform(), nil
	}
}

// detectWSLArchitecture detects the architecture of a WSL distribution.
func detectWSLArchitecture(ctx context.Context, distro string) (Platform, error) {
	// Check if uname -m works in WSL
	cmd := exec.CommandContext(ctx, "wsl", "-d", distro, "uname", "-m")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		// Fall back to current platform
		return GetCurrentPlatform(), fmt.Errorf("failed to detect WSL architecture: %w", err)
	}

	arch := strings.TrimSpace(out.String())
	return wslArchToPlatform(arch), nil
}

// detectDockerArchitecture detects the architecture of a Docker container.
func detectDockerArchitecture(ctx context.Context, container string) (Platform, error) {
	// Use docker inspect to get architecture
	cmd := exec.CommandContext(ctx, "docker", "inspect", "--format={{.Architecture}}", container)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		// Fall back to current platform
		return GetCurrentPlatform(), fmt.Errorf("failed to detect Docker architecture: %w", err)
	}

	arch := strings.TrimSpace(out.String())
	return dockerArchToPlatform(arch), nil
}

// detectSSHArchitecture detects the architecture via SSH.
func detectSSHArchitecture(ctx context.Context, target string) (Platform, error) {
	// Use uname -m via SSH
	cmd := exec.CommandContext(ctx, "ssh", target, "uname", "-m")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		// Fall back to current platform
		return GetCurrentPlatform(), fmt.Errorf("failed to detect SSH architecture: %w", err)
	}

	arch := strings.TrimSpace(out.String())
	return sshArchToPlatform(arch), nil
}

// wslArchToPlatform converts WSL uname -m to Platform.
func wslArchToPlatform(arch string) Platform {
	switch arch {
	case "x86_64", "amd64":
		return Platform{OS: "linux", Arch: "amd64"}
	case "aarch64", "arm64":
		return Platform{OS: "linux", Arch: "arm64"}
	default:
		// Default to current platform
		return GetCurrentPlatform()
	}
}

// dockerArchToPlatform converts Docker architecture to Platform.
func dockerArchToPlatform(arch string) Platform {
	switch arch {
	case "x86_64", "amd64":
		return Platform{OS: "linux", Arch: "amd64"}
	case "aarch64", "arm64", "arm/v8":
		return Platform{OS: "linux", Arch: "arm64"}
	default:
		// Default to current platform
		return GetCurrentPlatform()
	}
}

// sshArchToPlatform converts SSH uname -m to Platform.
func sshArchToPlatform(arch string) Platform {
	switch arch {
	case "x86_64", "amd64":
		return Platform{OS: runtime.GOOS, Arch: "amd64"}
	case "aarch64", "arm64":
		return Platform{OS: runtime.GOOS, Arch: "arm64"}
	case "i386", "i686":
		return Platform{OS: runtime.GOOS, Arch: "386"}
	default:
		// Default to current platform
		return GetCurrentPlatform()
	}
}

// GetPlatformForTarget returns the appropriate platform for bootstrapping to a target.
// This is a convenience function that wraps DetectTargetArchitecture.
func GetPlatformForTarget(ctx context.Context, targetType, target string) (Platform, error) {
	// If target is empty, use current platform
	if target == "" {
		return GetCurrentPlatform(), nil
	}

	return DetectTargetArchitecture(ctx, targetType, target)
}
