package workspace

import (
	"fmt"

	"github.com/codepod-io/codepod/internal/docker"
	"github.com/codepod-io/codepod/internal/port"
	"github.com/codepod-io/codepod/internal/storage"
	"github.com/codepod-io/codepod/internal/wsl"
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

func (m *MockDockerClient) ExecInContainerDetached(name string, cmd []string) error {
	return nil
}

func (m *MockDockerClient) CopyToContainer(containerID, src, dest string) error {
	return nil
}

func (m *MockDockerClient) CommitContainer(containerName, imageName string) error {
	return nil
}

func (m *MockDockerClient) ListContainers() ([]docker.Container, error) {
	var result []docker.Container
	for _, c := range m.Containers {
		result = append(result, docker.Container{
			ID:     c.Name,
			Names:  c.Name,
			Status: "running",
			Image:  c.Image,
		})
	}
	return result, nil
}

func (m *MockDockerClient) PullImage(image string) error {
	return nil
}

func (m *MockDockerClient) GetContainerIP(name string) (string, error) {
	return "172.17.0.2", nil
}

func (m *MockDockerClient) GetContainerByName(name string) (*docker.Container, error) {
	c, ok := m.Containers[name]
	if !ok {
		return nil, fmt.Errorf("container %s not found", name)
	}
	return &docker.Container{
		ID:     c.Name,
		Names:  c.Name,
		Status: "running",
		Image:  c.Image,
	}, nil
}

// MockPlatform is a mock implementation of WSL platform for testing
type MockPlatform struct {
	PlatformType        wsl.PlatformType
	Hostname            string
	RunCmdResult        string
	RunCmdError         error
	FileExistsResult    bool
	StorageManager      *storage.Manager
	PortPool            *port.Pool
}

func NewMockPlatform() *MockPlatform {
	return &MockPlatform{
		PlatformType: wsl.PlatformLinux,
		Hostname:    "test-host",
	}
}

func (m *MockPlatform) GetType() wsl.PlatformType {
	return m.PlatformType
}

func (m *MockPlatform) GetDistribution() string {
	return "Ubuntu-22.04"
}

func (m *MockPlatform) GetHostname() (string, error) {
	return m.Hostname, nil
}

func (m *MockPlatform) RunCommand(cmd string) (string, error) {
	return m.RunCmdResult, m.RunCmdError
}

func (m *MockPlatform) FileExists(path string) bool {
	return m.FileExistsResult
}

func (m *MockPlatform) CopyToWSL(src, dest string) error {
	return nil
}

func (m *MockPlatform) CopyFromWSL(src, dest string) error {
	return nil
}
