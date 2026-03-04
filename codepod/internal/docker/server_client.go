package docker

import (
	"fmt"

	"github.com/codepod-io/codepod/internal/server"
)

// ServerDockerClient uses the WSL server for Docker operations
type ServerDockerClient struct {
	serverClient *server.Client
}

// Ensure ServerDockerClient implements DockerClient
var _ DockerClient = (*ServerDockerClient)(nil)

// NewServerDockerClient creates a new Docker client that uses the WSL server
func NewServerDockerClient(serverURL string) (DockerClient, error) {
	client := server.New(serverURL)

	// Verify server is reachable
	if err := client.HealthCheck(); err != nil {
		return nil, fmt.Errorf("server not available: %w", err)
	}

	return &ServerDockerClient{
		serverClient: client,
	}, nil
}

// ListContainers lists all containers
func (c *ServerDockerClient) ListContainers() ([]Container, error) {
	containers, err := c.serverClient.ListContainers()
	if err != nil {
		return nil, err
	}

	// Convert server.Container to docker.Container
	result := make([]Container, len(containers))
	for i, c := range containers {
		result[i] = Container{
			ID:     c.ID,
			Names:  c.Names,
			Image:  c.Image,
			Status: c.Status,
		}
	}
	return result, nil
}

// PullImage pulls an image
func (c *ServerDockerClient) PullImage(image string) error {
	return c.serverClient.PullImage(image)
}

// CreateContainer is not implemented for server client
// Use the regular client for create operations
func (c *ServerDockerClient) CreateContainer(config *ContainerConfig) (string, error) {
	return "", fmt.Errorf("not implemented: use direct Docker client for create operations")
}

// StartContainer is not implemented for server client
func (c *ServerDockerClient) StartContainer(name string) error {
	return fmt.Errorf("not implemented: use direct Docker client for start operations")
}

// StopContainer is not implemented for server client
func (c *ServerDockerClient) StopContainer(name string) error {
	return fmt.Errorf("not implemented: use direct Docker client for stop operations")
}

// RemoveContainer is not implemented for server client
func (c *ServerDockerClient) RemoveContainer(name string, force bool) error {
	return fmt.Errorf("not implemented: use direct Docker client for remove operations")
}

// InspectContainer is not implemented for server client
func (c *ServerDockerClient) InspectContainer(name string) (*ContainerInfo, error) {
	return nil, fmt.Errorf("not implemented: use direct Docker client for inspect operations")
}

// ContainerExists is not implemented for server client
func (c *ServerDockerClient) ContainerExists(name string) bool {
	return false
}

// ExecInContainer is not implemented for server client
func (c *ServerDockerClient) ExecInContainer(name string, cmd []string) error {
	return fmt.Errorf("not implemented: use direct Docker client for exec operations")
}

// ExecInContainerDetached is not implemented for server client
func (c *ServerDockerClient) ExecInContainerDetached(name string, cmd []string) error {
	return fmt.Errorf("not implemented: use direct Docker client for exec operations")
}

// CopyToContainer is not implemented for server client
func (c *ServerDockerClient) CopyToContainer(containerID, src, dest string) error {
	return fmt.Errorf("not implemented: use direct Docker client for copy operations")
}

// CommitContainer is not implemented for server client
func (c *ServerDockerClient) CommitContainer(containerName, imageName string) error {
	return fmt.Errorf("not implemented: use direct Docker client for commit operations")
}

// Close closes the client
func (c *ServerDockerClient) Close() error {
	return nil
}

// GetContainerIP is not implemented for server client
func (c *ServerDockerClient) GetContainerIP(name string) (string, error) {
	return "", fmt.Errorf("not implemented: use direct Docker client for IP operations")
}

// GetContainerByName is not implemented for server client
func (c *ServerDockerClient) GetContainerByName(name string) (*Container, error) {
	return nil, fmt.Errorf("not implemented: use direct Docker client for container lookup")
}
