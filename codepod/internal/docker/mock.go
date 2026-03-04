package docker

import (
	"fmt"
)

// MockDockerClient is a mock implementation of DockerClient for testing
type MockDockerClient struct {
	Containers   []Container
	ImageExists map[string]bool
	ContainerErr error
}

// Ensure MockDockerClient implements DockerClient
var _ DockerClient = (*MockDockerClient)(nil)

// NewMockDockerClient creates a new mock Docker client
func NewMockDockerClient() *MockDockerClient {
	return &MockDockerClient{
		Containers:  []Container{},
		ImageExists: make(map[string]bool),
	}
}

// AddContainer adds a container to the mock
func (m *MockDockerClient) AddContainer(id, name, image, status string) {
	m.Containers = append(m.Containers, Container{
		ID:     id,
		Names:  name,
		Image:  image,
		Status: status,
	})
}

// AddImage adds an image to the mock
func (m *MockDockerClient) AddImage(image string) {
	m.ImageExists[image] = true
}

// SetError sets a mock error for all operations
func (m *MockDockerClient) SetError(err error) {
	m.ContainerErr = err
}

// CreateContainer creates a container (mock)
func (m *MockDockerClient) CreateContainer(config *ContainerConfig) (string, error) {
	if m.ContainerErr != nil {
		return "", m.ContainerErr
	}
	id := fmt.Sprintf("mock-container-%d", len(m.Containers)+1)
	m.Containers = append(m.Containers, Container{
		ID:     id,
		Names:  config.Name,
		Image:  config.Image,
		Status: "created",
	})
	return id, nil
}

// StartContainer starts a container (mock)
func (m *MockDockerClient) StartContainer(name string) error {
	if m.ContainerErr != nil {
		return m.ContainerErr
	}
	return nil
}

// StopContainer stops a container (mock)
func (m *MockDockerClient) StopContainer(name string) error {
	if m.ContainerErr != nil {
		return m.ContainerErr
	}
	return nil
}

// RemoveContainer removes a container (mock)
func (m *MockDockerClient) RemoveContainer(name string, force bool) error {
	if m.ContainerErr != nil {
		return m.ContainerErr
	}
	for i, c := range m.Containers {
		if c.Names == name {
			m.Containers = append(m.Containers[:i], m.Containers[i+1:]...)
			return nil
		}
	}
	return nil
}

// InspectContainer inspects a container (mock)
func (m *MockDockerClient) InspectContainer(name string) (*ContainerInfo, error) {
	if m.ContainerErr != nil {
		return nil, m.ContainerErr
	}
	return &ContainerInfo{
		Running:   true,
		IPAddress: "172.17.0.2",
	}, nil
}

// ContainerExists checks if container exists (mock)
func (m *MockDockerClient) ContainerExists(name string) bool {
	for _, c := range m.Containers {
		if c.Names == name {
			return true
		}
	}
	return false
}

// ExecInContainer execs in container (mock)
func (m *MockDockerClient) ExecInContainer(name string, cmd []string) error {
	if m.ContainerErr != nil {
		return m.ContainerErr
	}
	return nil
}

// ExecInContainerDetached execs in container detached (mock)
func (m *MockDockerClient) ExecInContainerDetached(name string, cmd []string) error {
	if m.ContainerErr != nil {
		return m.ContainerErr
	}
	return nil
}

// CopyToContainer copies to container (mock)
func (m *MockDockerClient) CopyToContainer(containerID, src, dest string) error {
	if m.ContainerErr != nil {
		return m.ContainerErr
	}
	return nil
}

// CommitContainer commits container (mock)
func (m *MockDockerClient) CommitContainer(containerName, imageName string) error {
	if m.ContainerErr != nil {
		return m.ContainerErr
	}
	return nil
}

// Close closes the client (mock)
func (m *MockDockerClient) Close() error {
	return nil
}

// ListContainers lists containers (mock)
func (m *MockDockerClient) ListContainers() ([]Container, error) {
	if m.ContainerErr != nil {
		return nil, m.ContainerErr
	}
	return m.Containers, nil
}

// PullImage pulls an image (mock)
func (m *MockDockerClient) PullImage(image string) error {
	if m.ContainerErr != nil {
		return m.ContainerErr
	}
	m.ImageExists[image] = true
	return nil
}

// GetContainerIP gets container IP (mock)
func (m *MockDockerClient) GetContainerIP(name string) (string, error) {
	if m.ContainerErr != nil {
		return "", m.ContainerErr
	}
	return "172.17.0.2", nil
}

// GetContainerByName gets container by name (mock)
func (m *MockDockerClient) GetContainerByName(name string) (*Container, error) {
	if m.ContainerErr != nil {
		return nil, m.ContainerErr
	}
	for _, c := range m.Containers {
		if c.Names == name {
			return &c, nil
		}
	}
	return nil, fmt.Errorf("container %s not found", name)
}
