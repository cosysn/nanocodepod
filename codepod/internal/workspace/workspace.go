package workspace

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/codepod-io/codepod/internal/config"
	"github.com/codepod-io/codepod/internal/docker"
	"github.com/codepod-io/codepod/internal/port"
	"github.com/codepod-io/codepod/internal/storage"
	"github.com/codepod-io/codepod/internal/types"
	"github.com/codepod-io/codepod/internal/wsl"
	"gopkg.in/yaml.v3"
)

// Manager manages workspaces
type Manager struct {
	dockerClient *docker.Client
	platform     *wsl.Platform
	storage      *storage.Manager
	portPool     *port.Pool
}

// New creates a new workspace manager
func New() (*Manager, error) {
	platform, err := wsl.NewPlatform()
	if err != nil {
		return nil, err
	}

	dockerHost := wsl.GetDockerHost(platform.Type)
	dockerClient, err := docker.New(dockerHost)
	if err != nil {
		return nil, err
	}

	storageManager, err := storage.New(platform)
	if err != nil {
		return nil, err
	}

	portPool, err := port.New()
	if err != nil {
		return nil, err
	}

	return &Manager{
		dockerClient: dockerClient,
		platform:     platform,
		storage:      storageManager,
		portPool:     portPool,
	}, nil
}

// Create creates a new workspace
func (m *Manager) Create(name string, opts *CreateOptions) (*types.Workspace, error) {
	// Check if workspace already exists
	if exists, err := m.Exists(name); err != nil {
		return nil, err
	} else if exists {
		return m.Get(name)
	}

	// Set defaults
	if opts == nil {
		opts = &CreateOptions{}
	}
	if opts.Image == "" {
		opts.Image = "ubuntu:22.04"
	}
	if opts.Repository.Branch == "" {
		opts.Repository.Branch = "main"
	}

	// Allocate port
	allocatedPort, err := m.portPool.Allocate()
	if err != nil {
		return nil, err
	}

	// Create storage
	storagePath, err := m.storage.CreateWorkspaceStorage(name)
	if err != nil {
		m.portPool.Release(allocatedPort)
		return nil, err
	}

	// Create workspace directory in WSL
	if err := m.storage.CreateWSLStorage(name); err != nil {
		m.storage.DeleteWorkspaceStorage(name)
		m.portPool.Release(allocatedPort)
		return nil, err
	}

	// Create Docker container
	containerConfig := &docker.ContainerConfig{
		Name:    GetContainerName(name),
		Image:   opts.Image,
		Cmd:     []string{"sleep", "infinity"},
		Env:     []string{},
		Labels:  map[string]string{"codepod.workspace": name},
		PortBindings: map[string][]docker.PortBinding{
			"22/tcp": {{HostIP: "0.0.0.0", HostPort: fmt.Sprintf("%d", allocatedPort)}},
		},
		Binds: []string{
			m.storage.BindWSLStorage(name, "/workspace"),
		},
		NetworkMode: "bridge",
		Privileged:  false,
	}

	containerID, err := m.dockerClient.CreateContainer(containerConfig)
	if err != nil {
		m.storage.DeleteWorkspaceStorage(name)
		m.portPool.Release(allocatedPort)
		return nil, err
	}

	workspace := &types.Workspace{
		Name:        name,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		State:       types.WorkspaceStateCreated,
		Repository:  opts.Repository,
		IDE:         opts.IDE,
		Container:   types.Container{Image: opts.Image, Name: containerConfig.Name},
		Domain:      fmt.Sprintf("%s.local", name),
		SSH:         types.SSH{},
		Agent:       types.Agent{Port: 22001, Status: "stopped"},
		Port:        allocatedPort,
		StoragePath: storagePath,
	}

	// Save workspace
	if err := m.saveWorkspace(workspace); err != nil {
		m.dockerClient.RemoveContainer(containerID, true)
		m.storage.DeleteWorkspaceStorage(name)
		m.portPool.Release(allocatedPort)
		return nil, err
	}

	return workspace, nil
}

// Start starts a workspace
func (m *Manager) Start(name string) (*types.Workspace, error) {
	workspace, err := m.Get(name)
	if err != nil {
		return nil, err
	}

	if workspace.State == types.WorkspaceStateRunning {
		return workspace, nil
	}

	exists := m.dockerClient.ContainerExists(workspace.Container.Name)
	if !exists {
		// Recreate container
		containerConfig := &docker.ContainerConfig{
			Name:    workspace.Container.Name,
			Image:   workspace.Container.Image,
			Cmd:     []string{"sleep", "infinity"},
			Labels:  map[string]string{"codepod.workspace": name},
			PortBindings: map[string][]docker.PortBinding{
				"22/tcp": {{HostIP: "0.0.0.0", HostPort: fmt.Sprintf("%d", workspace.Port)}},
			},
			Binds: []string{
				m.storage.BindWSLStorage(name, "/workspace"),
			},
			NetworkMode: "bridge",
			Privileged:  true,
		}
		_, err := m.dockerClient.CreateContainer(containerConfig)
		if err != nil {
			return nil, err
		}
	}

	if err := m.dockerClient.StartContainer(workspace.Container.Name); err != nil {
		return nil, err
	}

	workspace.State = types.WorkspaceStateRunning
	workspace.UpdatedAt = time.Now()
	if err := m.saveWorkspace(workspace); err != nil {
		return nil, err
	}

	return workspace, nil
}

// Stop stops a workspace
func (m *Manager) Stop(name string) (*types.Workspace, error) {
	workspace, err := m.Get(name)
	if err != nil {
		return nil, err
	}

	if workspace.State == types.WorkspaceStateStopped {
		return workspace, nil
	}

	if err := m.dockerClient.StopContainer(workspace.Container.Name); err != nil {
		return nil, err
	}

	workspace.State = types.WorkspaceStateStopped
	workspace.UpdatedAt = time.Now()
	if err := m.saveWorkspace(workspace); err != nil {
		return nil, err
	}

	return workspace, nil
}

// Delete deletes a workspace
func (m *Manager) Delete(name string) error {
	workspace, err := m.Get(name)
	if err != nil {
		return err
	}

	// Stop and remove container
	if err := m.dockerClient.StopContainer(workspace.Container.Name); err != nil {
		// Try to force remove
		m.dockerClient.RemoveContainer(workspace.Container.Name, true)
	} else {
		m.dockerClient.RemoveContainer(workspace.Container.Name, false)
	}

	// Release port
	m.portPool.Release(workspace.Port)

	// Delete storage
	if err := m.storage.DeleteWorkspaceStorage(name); err != nil {
		return err
	}

	// Remove from workspace file
	workspaceFile := m.getWorkspaceFile(name)
	if err := os.Remove(workspaceFile); err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}

// Get returns a workspace by name
func (m *Manager) Get(name string) (*types.Workspace, error) {
	workspaceFile := m.getWorkspaceFile(name)
	data, err := os.ReadFile(workspaceFile)
	if err != nil {
		return nil, fmt.Errorf("workspace %s not found", name)
	}

	var workspace types.Workspace
	if err := unmarshalWorkspace(data, &workspace); err != nil {
		return nil, err
	}

	// Update state based on current container state
	exists := m.dockerClient.ContainerExists(workspace.Container.Name)

	if !exists {
		workspace.State = types.WorkspaceStateError
	} else {
		inspect, err := m.dockerClient.InspectContainer(workspace.Container.Name)
		if err == nil && inspect.Running {
			workspace.State = types.WorkspaceStateRunning
		} else {
			workspace.State = types.WorkspaceStateStopped
		}
	}

	return &workspace, nil
}

// List lists all workspaces
func (m *Manager) List() ([]*types.Workspace, error) {
	workspacesDir, err := getWorkspacesDir()
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(workspacesDir)
	if err != nil {
		return nil, err
	}

	var workspaces []*types.Workspace
	for _, entry := range entries {
		if entry.IsDir() {
			workspace, err := m.Get(entry.Name())
			if err != nil {
				continue
			}
			workspaces = append(workspaces, workspace)
		}
	}

	return workspaces, nil
}

// Exists checks if a workspace exists
func (m *Manager) Exists(name string) (bool, error) {
	workspaceFile := m.getWorkspaceFile(name)
	_, err := os.Stat(workspaceFile)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// CreateOptions holds options for creating a workspace
type CreateOptions struct {
	Image       string
	Repository  types.Repository
	IDE         types.IDE
}

// GetWorkspaceDir returns the workspace directory
func getWorkspacesDir() (string, error) {
	configDir, err := config.GetConfigDir()
	if err != nil {
		return "", err
	}
	workspacesDir := filepath.Join(configDir, "workspaces")
	if err := os.MkdirAll(workspacesDir, 0755); err != nil {
		return "", err
	}
	return workspacesDir, nil
}

func (m *Manager) getWorkspaceFile(name string) string {
	workspacesDir, _ := getWorkspacesDir()
	return filepath.Join(workspacesDir, name+".yaml")
}

func (m *Manager) saveWorkspace(workspace *types.Workspace) error {
	workspacesDir, err := getWorkspacesDir()
	if err != nil {
		return err
	}

	workspaceFile := filepath.Join(workspacesDir, workspace.Name+".yaml")
	data, err := marshalWorkspace(workspace)
	if err != nil {
		return err
	}

	return os.WriteFile(workspaceFile, data, 0644)
}

func marshalWorkspace(workspace *types.Workspace) ([]byte, error) {
	return yaml.Marshal(workspace)
}

func unmarshalWorkspace(data []byte, workspace *types.Workspace) error {
	return yaml.Unmarshal(data, workspace)
}
