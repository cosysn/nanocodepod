package docker

import (
	"fmt"
	"testing"
)

func TestNew(t *testing.T) {
	client, err := New("unix:///var/run/docker.sock")
	if err != nil {
		t.Skipf("Docker is not available: %v", err)
	}

	if client == nil {
		t.Error("client should not be nil")
	}
}

func TestClient_Close(t *testing.T) {
	client, err := New("unix:///var/run/docker.sock")
	if err != nil {
		t.Skipf("Docker is not available: %v", err)
	}
	err = client.Close()
	if err != nil {
		t.Fatalf("failed to close: %v", err)
	}
}

func TestContainerConfig(t *testing.T) {
	config := &ContainerConfig{
		Name:    "test-container",
		Image:   "ubuntu:22.04",
		Cmd:     []string{"sleep", "infinity"},
		Env:     []string{"VAR=value"},
		Labels:  map[string]string{"app": "test"},
		PortBindings: map[string][]PortBinding{
			"22/tcp": {{HostIP: "0.0.0.0", HostPort: "22000"}},
		},
		Binds:       []string{"/tmp:/workspace"},
		NetworkMode: "bridge",
		Privileged:  false,
	}

	if config.Name != "test-container" {
		t.Errorf("expected name test-container, got %s", config.Name)
	}
	if config.Image != "ubuntu:22.04" {
		t.Errorf("expected image ubuntu:22.04, got %s", config.Image)
	}
	if len(config.PortBindings) != 1 {
		t.Errorf("expected 1 port binding, got %d", len(config.PortBindings))
	}
}

func TestPortBinding(t *testing.T) {
	binding := PortBinding{
		HostIP:   "0.0.0.0",
		HostPort: "22000",
	}

	if binding.HostIP != "0.0.0.0" {
		t.Errorf("expected HostIP 0.0.0.0, got %s", binding.HostIP)
	}
	if binding.HostPort != "22000" {
		t.Errorf("expected HostPort 22000, got %s", binding.HostPort)
	}
}

func TestContainer(t *testing.T) {
	container := Container{
		ID:     "abc123",
		Names:  "test-container",
		Status: "running",
		Image:  "ubuntu:22.04",
	}

	if container.ID != "abc123" {
		t.Errorf("expected ID abc123, got %s", container.ID)
	}
	if container.Names != "test-container" {
		t.Errorf("expected Names test-container, got %s", container.Names)
	}
}

func TestContainerInfo(t *testing.T) {
	info := &ContainerInfo{
		Running:   true,
		IPAddress: "172.17.0.2",
	}

	if !info.Running {
		t.Error("expected Running to be true")
	}
	if info.IPAddress != "172.17.0.2" {
		t.Errorf("expected IPAddress 172.17.0.2, got %s", info.IPAddress)
	}
}

func TestContainerConfigWithAllOptions(t *testing.T) {
	config := &ContainerConfig{
		Name:    "full-test-container",
		Image:   "alpine:latest",
		Cmd:     []string{"sh", "-c", "echo hello"},
		Env:     []string{"VAR1=value1", "VAR2=value2"},
		Labels:  map[string]string{"env": "test", "version": "1.0"},
		Binds:   []string{"/host/path:/container/path:ro"},
		NetworkMode: "host",
		Privileged:  true,
	}

	if config.Name != "full-test-container" {
		t.Errorf("expected Name full-test-container, got %s", config.Name)
	}
	if len(config.Env) != 2 {
		t.Errorf("expected 2 env vars, got %d", len(config.Env))
	}
	if !config.Privileged {
		t.Error("expected Privileged to be true")
	}
}

func TestClientImplementsInterface(t *testing.T) {
	var _ DockerClient = (*Client)(nil)
}

// TestMockDockerClient tests the mock client
func TestMockDockerClient(t *testing.T) {
	mock := NewMockDockerClient()

	// Test CreateContainer
	id, err := mock.CreateContainer(&ContainerConfig{
		Name:  "test-container",
		Image: "ubuntu:22.04",
	})
	if err != nil {
		t.Errorf("CreateContainer failed: %v", err)
	}
	if id == "" {
		t.Error("container ID should not be empty")
	}

	// Test container exists
	if !mock.ContainerExists("test-container") {
		t.Error("container should exist")
	}

	// Test StartContainer
	err = mock.StartContainer("test-container")
	if err != nil {
		t.Errorf("StartContainer failed: %v", err)
	}

	// Test StopContainer
	err = mock.StopContainer("test-container")
	if err != nil {
		t.Errorf("StopContainer failed: %v", err)
	}

	// Test RemoveContainer
	err = mock.RemoveContainer("test-container", false)
	if err != nil {
		t.Errorf("RemoveContainer failed: %v", err)
	}

	// Test container no longer exists
	if mock.ContainerExists("test-container") {
		t.Error("container should not exist after removal")
	}

	// Test InspectContainer
	mock.CreateContainer(&ContainerConfig{Name: "test2", Image: "alpine"})
	info, err := mock.InspectContainer("test2")
	if err != nil {
		t.Errorf("InspectContainer failed: %v", err)
	}
	if info == nil {
		t.Error("info should not be nil")
	}

	// Test GetContainerIP
	ip, err := mock.GetContainerIP("test2")
	if err != nil {
		t.Errorf("GetContainerIP failed: %v", err)
	}
	_ = ip

	// Test ListContainers
	mock.CreateContainer(&ContainerConfig{Name: "test3", Image: "nginx"})
	containers, err := mock.ListContainers()
	if err != nil {
		t.Errorf("ListContainers failed: %v", err)
	}
	if len(containers) != 2 {
		t.Errorf("expected 2 containers, got %d", len(containers))
	}

	// Test PullImage
	err = mock.PullImage("nginx:latest")
	if err != nil {
		t.Errorf("PullImage failed: %v", err)
	}

	// Test CommitContainer
	err = mock.CommitContainer("test3", "myimage:v1")
	if err != nil {
		t.Errorf("CommitContainer failed: %v", err)
	}

	// Test ExecInContainer
	err = mock.ExecInContainer("test3", []string{"echo", "hello"})
	if err != nil {
		t.Errorf("ExecInContainer failed: %v", err)
	}

	// Test ExecInContainerDetached
	err = mock.ExecInContainerDetached("test3", []string{"sleep", "infinity"})
	if err != nil {
		t.Errorf("ExecInContainerDetached failed: %v", err)
	}

	// Test CopyToContainer
	err = mock.CopyToContainer("test3", "/tmp/file", "/dest")
	if err != nil {
		t.Errorf("CopyToContainer failed: %v", err)
	}

	// Test GetContainerByName
	c, err := mock.GetContainerByName("test3")
	if err != nil {
		t.Errorf("GetContainerByName failed: %v", err)
	}
	if c.Names != "test3" {
		t.Errorf("expected test3, got %s", c.Names)
	}

	// Test Close
	err = mock.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}
}

// MockDockerClient is a mock implementation for testing
type MockDockerClient struct {
	Containers      map[string]*MockContainer
	NextContainerID int
}

type MockContainer struct {
	Name    string
	Image   string
	Running bool
	IP      string
}

func NewMockDockerClient() *MockDockerClient {
	return &MockDockerClient{
		Containers:      make(map[string]*MockContainer),
		NextContainerID: 1,
	}
}

func (m *MockDockerClient) CreateContainer(config *ContainerConfig) (string, error) {
	id := fmt.Sprintf("mock-container-%d", m.NextContainerID)
	m.NextContainerID++
	m.Containers[config.Name] = &MockContainer{
		Name:    config.Name,
		Image:   config.Image,
		Running: false,
	}
	return id, nil
}

func (m *MockDockerClient) StartContainer(name string) error {
	if c, ok := m.Containers[name]; ok {
		c.Running = true
	}
	return nil
}

func (m *MockDockerClient) StopContainer(name string) error {
	if c, ok := m.Containers[name]; ok {
		c.Running = false
	}
	return nil
}

func (m *MockDockerClient) RemoveContainer(name string, force bool) error {
	delete(m.Containers, name)
	return nil
}

func (m *MockDockerClient) InspectContainer(name string) (*ContainerInfo, error) {
	c, ok := m.Containers[name]
	if !ok {
		return &ContainerInfo{Running: false}, nil
	}
	return &ContainerInfo{Running: c.Running, IPAddress: c.IP}, nil
}

func (m *MockDockerClient) ContainerExists(name string) bool {
	_, ok := m.Containers[name]
	return ok
}

func (m *MockDockerClient) ExecInContainer(name string, cmd []string) error {
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

func (m *MockDockerClient) Close() error {
	return nil
}

func (m *MockDockerClient) ListContainers() ([]Container, error) {
	var result []Container
	for _, c := range m.Containers {
		result = append(result, Container{
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
	if c, ok := m.Containers[name]; ok {
		return c.IP, nil
	}
	return "", nil
}

func (m *MockDockerClient) GetContainerByName(name string) (*Container, error) {
	c, ok := m.Containers[name]
	if !ok {
		return nil, fmt.Errorf("container %s not found", name)
	}
	return &Container{
		ID:     c.Name,
		Names:  c.Name,
		Status: "running",
		Image:  c.Image,
	}, nil
}
