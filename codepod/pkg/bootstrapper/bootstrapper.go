// Package bootstrapper provides agent bootstrap functionality for CodePod.
// It handles injecting agent binaries into different environments.
package bootstrapper

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
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

// Bootstrap starts an agent in the target environment.
func (b *Bootstrapper) Bootstrap(ctx context.Context, config *BootstrapConfig) error {
	// Validate config
	if config.Target == "" {
		return errors.New("invalid bootstrap config: empty target")
	}

	// Check binary exists
	if _, err := os.Stat(config.BinaryPath); err != nil {
		if os.IsNotExist(err) {
			return errors.New("agent binary not found: " + config.BinaryPath)
		}
		return err
	}

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
