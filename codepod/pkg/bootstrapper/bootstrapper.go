// Package bootstrapper provides agent bootstrap functionality for CodePod.
// It handles injecting agent binaries into different environments.
package bootstrapper

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/codepod-io/codepod/pkg/embed"
)

// TargetType represents the type of bootstrap target.
type TargetType string

const (
	TargetTypeWSL      TargetType = "wsl"
	TargetTypeDocker  TargetType = "docker"
	TargetTypeSSH     TargetType = "ssh"
	TargetTypeUDS     TargetType = "uds"
)

// AgentType represents the type of agent to bootstrap.
type AgentType string

const (
	AgentTypeLocal     AgentType = "local"
	AgentTypeWorkspace AgentType = "workspace"
	AgentTypeContainer AgentType = "container"
)

// BootstrapConfig contains configuration for bootstrapping an agent.
type BootstrapConfig struct {
	TargetType TargetType // Target environment type
	Target     string     // Target identifier (e.g., wsl distro name, container name)
	AgentType  AgentType  // Type of agent to start
	BinaryPath string     // Path to agent binary
}

// Bootstrapper handles bootstrapping agents in different environments.
type Bootstrapper struct {
	binaryPath string
}

// NewBootstrapper creates a new Bootstrapper.
func NewBootstrapper(binaryPath string) *Bootstrapper {
	return &Bootstrapper{binaryPath: binaryPath}
}

// GetAgentBinaryPath returns the path to the agent binary.
// If binaryPath is provided, it uses that. Otherwise, it tries to extract
// from embedded binaries.
func GetAgentBinaryPath(binaryPath string) (string, error) {
	return GetAgentBinaryPathForTarget(binaryPath, "", "")
}

// GetAgentBinaryPathForTarget returns the path to the agent binary for a specific target.
// If binaryPath is provided, it uses that. Otherwise, it detects the target architecture
// and extracts the appropriate embedded binary.
func GetAgentBinaryPathForTarget(binaryPath string, targetType, target string) (string, error) {
	if binaryPath != "" {
		return binaryPath, nil
	}

	// Try to get embedded binary
	if embed.HasEmbeddedAgents() {
		// Determine target platform
		var platform embed.Platform
		var err error

		if target != "" {
			// Detect target architecture
			platform, err = embed.GetPlatformForTarget(context.Background(), targetType, target)
			if err != nil {
				// Fall back to current platform
				platform = embed.GetCurrentPlatform()
			}
		} else {
			// Use current platform
			platform = embed.GetCurrentPlatform()
		}

		extractedPath, err := embed.ExtractAgent(platform)
		if err != nil {
			return "", fmt.Errorf("failed to extract embedded agent: %w", err)
		}
		return extractedPath, nil
	}

	// Fall back to searching in common locations
	locations := []string{
		"./codepod-agent",
		"/usr/local/bin/codepod-agent",
		filepath.Join(os.Getenv("HOME"), "bin", "codepod-agent"),
	}

	for _, loc := range locations {
		if _, err := os.Stat(loc); err == nil {
			return loc, nil
		}
	}

	return "", errors.New("agent binary not found and no embedded agents available")
}

// Bootstrap starts an agent in the target environment.
func (b *Bootstrapper) Bootstrap(ctx context.Context, config *BootstrapConfig) error {
	// Validate config
	if config.Target == "" {
		return errors.New("invalid bootstrap config: empty target")
	}

	// Get binary path (handles embedded vs external)
	binaryPath, err := GetAgentBinaryPath(config.BinaryPath)
	if err != nil {
		return err
	}

	// Update config with resolved binary path
	config.BinaryPath = binaryPath

	// Bootstrap based on target type
	switch config.TargetType {
	case TargetTypeWSL:
		return b.bootstrapWSL(ctx, config)
	case TargetTypeDocker:
		return b.bootstrapDocker(ctx, config)
	case TargetTypeSSH:
		return b.bootstrapSSH(ctx, config)
	case TargetTypeUDS:
		return b.bootstrapUDS(ctx, config)
	default:
		return errors.New("unknown target type: " + string(config.TargetType))
	}
}

// bootstrapWSL bootstraps an agent in WSL.
func (b *Bootstrapper) bootstrapWSL(ctx context.Context, config *BootstrapConfig) error {
	// Copy binary to WSL and start agent
	destPath := filepath.Join("/tmp", filepath.Base(config.BinaryPath))

	// Use wsl command to copy and execute
	cmd := exec.CommandContext(ctx, "wsl", "-d", config.Target, "cp", config.BinaryPath, destPath)
	if err := cmd.Run(); err != nil {
		return err
	}

	// Start agent in WSL
	agentFlag := b.getAgentFlag(config.AgentType)
	cmd = exec.CommandContext(ctx, "wsl", "-d", config.Target, "chmod", "+x", destPath)
	if err := cmd.Run(); err != nil {
		return err
	}

	cmd = exec.CommandContext(ctx, "wsl", "-d", config.Target, destPath, agentFlag)
	return cmd.Start()
}

// bootstrapDocker bootstraps an agent in a Docker container.
func (b *Bootstrapper) bootstrapDocker(ctx context.Context, config *BootstrapConfig) error {
	// Copy binary to container
	destPath := "/tmp/codepod-agent"

	cmd := exec.CommandContext(ctx, "docker", "cp", config.BinaryPath, config.Target+":"+destPath)
	if err := cmd.Run(); err != nil {
		return err
	}

	// Start agent in container
	agentFlag := b.getAgentFlag(config.AgentType)
	cmd = exec.CommandContext(ctx, "docker", "exec", "-d", config.Target, "chmod", "+x", destPath)
	if err := cmd.Run(); err != nil {
		return err
	}

	cmd = exec.CommandContext(ctx, "docker", "exec", "-d", config.Target, destPath, agentFlag)
	return cmd.Start()
}

// bootstrapSSH bootstraps an agent via SSH.
func (b *Bootstrapper) bootstrapSSH(ctx context.Context, config *BootstrapConfig) error {
	// Copy binary via scp
	destPath := "/tmp/codepod-agent"

	cmd := exec.CommandContext(ctx, "scp", config.BinaryPath, config.Target+":"+destPath)
	if err := cmd.Run(); err != nil {
		return err
	}

	// Make executable and start
	agentFlag := b.getAgentFlag(config.AgentType)
	cmd = exec.CommandContext(ctx, "ssh", config.Target, "chmod", "+x", destPath)
	if err := cmd.Run(); err != nil {
		return err
	}

	cmd = exec.CommandContext(ctx, "ssh", "-f", config.Target, destPath, agentFlag)
	return cmd.Start()
}

// bootstrapUDS bootstraps a local agent via UDS.
func (b *Bootstrapper) bootstrapUDS(ctx context.Context, config *BootstrapConfig) error {
	// Start agent as a local process
	agentFlag := b.getAgentFlag(config.AgentType)
	cmd := exec.CommandContext(ctx, config.BinaryPath, agentFlag)
	return cmd.Start()
}

// getAgentFlag returns the command line flag for the agent type.
func (b *Bootstrapper) getAgentFlag(agentType AgentType) string {
	switch agentType {
	case AgentTypeLocal:
		return "--local"
	case AgentTypeWorkspace:
		return "--workspace"
	case AgentTypeContainer:
		return "--container"
	default:
		return ""
	}
}
