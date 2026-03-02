package workspace

import (
	"github.com/codepod-io/codepod/internal/docker"
)

// MockDockerClient is a mock implementation of DockerClient for testing
type MockDockerClient struct {
	Containers       map[string]*MockContainer
	NextContainerID  int
	CreateContainerFunc func(config *docker.ContainerConfig) (string, error)
	StartContainerFunc func(name string) error
	StopContainerFunc  func(name string) error
	RemoveContainerFunc func(name string, force bool) error
	InspectContainerFunc func(name string) (*docker.ContainerInfo, error)
	ContainerExistsFunc   func(name string) bool
	ExecInContainerFunc   func(name string, cmd []string) error
	CloseFunc             func() error
}

type MockContainer struct {
	Name    string
	Image   string
	Running bool
}

// NewMockDockerClient creates a new mock Docker client
func NewMockDockerClient() *MockDockerClient {
	return &MockDockerClient{
		Containers:      make(map[string]*MockContainer),
		NextContainerID: 1,
	}
}

func (m *MockDockerClient) CreateContainer(config *docker.ContainerConfig) (string, error) {
	if m.CreateContainerFunc != nil {
		return m.CreateContainerFunc(config)
	}
	containerID := "mock-container-1"
	m.NextContainerID++
	m.Containers[config.Name] = &MockContainer{
		Name:    config.Name,
		Image:   config.Image,
		Running: false,
	}
	return containerID, nil
}

func (m *MockDockerClient) StartContainer(name string) error {
	if m.StartContainerFunc != nil {
		return m.StartContainerFunc(name)
	}
	if container, ok := m.Containers[name]; ok {
		container.Running = true
	}
	return nil
}

func (m *MockDockerClient) StopContainer(name string) error {
	if m.StopContainerFunc != nil {
		return m.StopContainerFunc(name)
	}
	if container, ok := m.Containers[name]; ok {
		container.Running = false
	}
	return nil
}

func (m *MockDockerClient) RemoveContainer(name string, force bool) error {
	if m.RemoveContainerFunc != nil {
		return m.RemoveContainerFunc(name, force)
	}
	delete(m.Containers, name)
	return nil
}

func (m *MockDockerClient) InspectContainer(name string) (*docker.ContainerInfo, error) {
	if m.InspectContainerFunc != nil {
		return m.InspectContainerFunc(name)
	}
	container, ok := m.Containers[name]
	if !ok {
		return &docker.ContainerInfo{Running: false}, nil
	}
	return &docker.ContainerInfo{Running: container.Running}, nil
}

func (m *MockDockerClient) ContainerExists(name string) bool {
	if m.ContainerExistsFunc != nil {
		return m.ContainerExistsFunc(name)
	}
	_, ok := m.Containers[name]
	return ok
}

func (m *MockDockerClient) ExecInContainer(name string, cmd []string) error {
	if m.ExecInContainerFunc != nil {
		return m.ExecInContainerFunc(name, cmd)
	}
	return nil
}

func (m *MockDockerClient) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}
