// Package provider provides environment providers for CodePod agents.
package provider

import (
	"context"
	"errors"

	"github.com/codepod-io/codepod/pkg/bootstrapper"
	"github.com/codepod-io/codepod/pkg/rpc"
)

// Provider defines the interface for environment providers.
type Provider interface {
	Type() string
	Bootstrap(ctx context.Context, identity string) (*rpc.RPCClient, error)
	Connect(ctx context.Context, identity string) (*rpc.RPCClient, error)
}

// BaseProvider provides common functionality for all providers.
type BaseProvider struct {
	providerType string
	bootstrapper *bootstrapper.Bootstrapper
	binaryPath   string
}

// WSLProvider handles WSL environment.
type WSLProvider struct {
	BaseProvider
	distro string
}

// NewWSLProvider creates a new WSL provider.
func NewWSLProvider(distro, binaryPath string) *WSLProvider {
	return &WSLProvider{
		BaseProvider: BaseProvider{
			providerType: "wsl",
			binaryPath:   binaryPath,
			bootstrapper: bootstrapper.NewBootstrapper(binaryPath),
		},
		distro: distro,
	}
}

// Type returns the provider type.
func (p *WSLProvider) Type() string {
	return p.providerType
}

// Bootstrap starts a Workspace Agent in WSL.
func (p *WSLProvider) Bootstrap(ctx context.Context, identity string) (*rpc.RPCClient, error) {
	// Bootstrap agent in WSL
	err := p.bootstrapper.Bootstrap(ctx, &bootstrapper.BootstrapConfig{
		TargetType: bootstrapper.TargetTypeWSL,
		Target:     identity,
		AgentType:  bootstrapper.AgentTypeWorkspace,
		BinaryPath: p.binaryPath,
	})
	if err != nil {
		return nil, err
	}

	// Connect to the bootstrapped agent
	return p.Connect(ctx, identity)
}

// Connect connects to an existing Workspace Agent in WSL.
func (p *WSLProvider) Connect(ctx context.Context, identity string) (*rpc.RPCClient, error) {
	// In production, would connect via WSL channel
	// For now, return placeholder
	return nil, errors.New("WSL connect not implemented")
}

// SSHProvider handles SSH environment.
type SSHProvider struct {
	BaseProvider
	host     string
	user     string
	password string
	keyPath  string
}

// NewSSHProvider creates a new SSH provider.
func NewSSHProvider(host, user, password, keyPath, binaryPath string) *SSHProvider {
	return &SSHProvider{
		BaseProvider: BaseProvider{
			providerType: "ssh-remote",
			binaryPath:   binaryPath,
			bootstrapper: bootstrapper.NewBootstrapper(binaryPath),
		},
		host:     host,
		user:     user,
		password: password,
		keyPath:  keyPath,
	}
}

// Type returns the provider type.
func (p *SSHProvider) Type() string {
	return p.providerType
}

// Bootstrap starts a Workspace Agent via SSH.
func (p *SSHProvider) Bootstrap(ctx context.Context, identity string) (*rpc.RPCClient, error) {
	target := p.user + "@" + identity

	err := p.bootstrapper.Bootstrap(ctx, &bootstrapper.BootstrapConfig{
		TargetType: bootstrapper.TargetTypeSSH,
		Target:     target,
		AgentType:  bootstrapper.AgentTypeWorkspace,
		BinaryPath: p.binaryPath,
	})
	if err != nil {
		return nil, err
	}

	return p.Connect(ctx, identity)
}

// Connect connects to an existing Workspace Agent via SSH.
func (p *SSHProvider) Connect(ctx context.Context, identity string) (*rpc.RPCClient, error) {
	// In production, would connect via SSH channel
	return nil, errors.New("SSH connect not implemented")
}

// DockerProvider handles Docker container environment.
type DockerProvider struct {
	BaseProvider
	container string
}

// NewDockerProvider creates a new Docker provider.
func NewDockerProvider(container, binaryPath string) *DockerProvider {
	return &DockerProvider{
		BaseProvider: BaseProvider{
			providerType: "docker-container",
			binaryPath:   binaryPath,
			bootstrapper: bootstrapper.NewBootstrapper(binaryPath),
		},
		container: container,
	}
}

// Type returns the provider type.
func (p *DockerProvider) Type() string {
	return p.providerType
}

// Bootstrap starts a Container Agent in Docker.
func (p *DockerProvider) Bootstrap(ctx context.Context, identity string) (*rpc.RPCClient, error) {
	err := p.bootstrapper.Bootstrap(ctx, &bootstrapper.BootstrapConfig{
		TargetType: bootstrapper.TargetTypeDocker,
		Target:     identity,
		AgentType:  bootstrapper.AgentTypeContainer,
		BinaryPath: p.binaryPath,
	})
	if err != nil {
		return nil, err
	}

	return p.Connect(ctx, identity)
}

// Connect connects to an existing Container Agent in Docker.
func (p *DockerProvider) Connect(ctx context.Context, identity string) (*rpc.RPCClient, error) {
	// In production, would connect via Docker exec
	return nil, errors.New("Docker connect not implemented")
}

// DevContainerProvider handles DevContainer environment.
type DevContainerProvider struct {
	BaseProvider
	configPath string
}

// NewDevContainerProvider creates a new DevContainer provider.
func NewDevContainerProvider(configPath, binaryPath string) *DevContainerProvider {
	return &DevContainerProvider{
		BaseProvider: BaseProvider{
			providerType: "dev-container",
			binaryPath:   binaryPath,
			bootstrapper: bootstrapper.NewBootstrapper(binaryPath),
		},
		configPath: configPath,
	}
}

// Type returns the provider type.
func (p *DevContainerProvider) Type() string {
	return p.providerType
}

// Bootstrap starts a Container Agent via DevContainer config.
func (p *DevContainerProvider) Bootstrap(ctx context.Context, identity string) (*rpc.RPCClient, error) {
	// For dev-container, identity should be parsed as DevContainerConfig
	// In production, would read config and bootstrap accordingly
	return nil, errors.New("DevContainer bootstrap not implemented")
}

// Connect connects to an existing DevContainer.
func (p *DevContainerProvider) Connect(ctx context.Context, identity string) (*rpc.RPCClient, error) {
	return nil, errors.New("DevContainer connect not implemented")
}
