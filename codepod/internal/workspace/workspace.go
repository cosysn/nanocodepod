package workspace

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/codepod-io/codepod/internal/config"
	"github.com/codepod-io/codepod/internal/devcon"
	"github.com/codepod-io/codepod/internal/docker"
	"github.com/codepod-io/codepod/internal/port"
	"github.com/codepod-io/codepod/internal/storage"
	"github.com/codepod-io/codepod/internal/types"
	"github.com/codepod-io/codepod/internal/wsl"
	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
)

// Manager manages workspaces
type Manager struct {
	dockerClient *docker.Client
	platform     *wsl.Platform
	storage      *storage.Manager
	portPool     *port.Pool
	devcon       *devcon.Devcon
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

	// Initialize devcon handler
	devconHandler, err := devcon.New(&wsl.WSL{})
	if err != nil {
		// Devcon is optional, continue without it
		devconHandler = nil
	}

	return &Manager{
		dockerClient: dockerClient,
		platform:     platform,
		storage:      storageManager,
		portPool:     portPool,
		devcon:       devconHandler,
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

	// Generate UUID for workspace
	workspaceUUID := uuid.New().String()

	// Allocate port
	allocatedPort, err := m.portPool.Allocate()
	if err != nil {
		return nil, err
	}

	// Create storage with UUID-based directory
	storagePath, err := m.storage.CreateWorkspaceStorage(workspaceUUID)
	if err != nil {
		m.portPool.Release(allocatedPort)
		return nil, err
	}

	// Create workspace directory in WSL
	if err := m.storage.CreateWSLStorage(workspaceUUID); err != nil {
		m.storage.DeleteWorkspaceStorage(workspaceUUID)
		m.portPool.Release(allocatedPort)
		return nil, err
	}

	// Code will be cloned/copied here
	codePath := filepath.Join(storagePath, "code")

	// Clone repository or copy local directory (on host before container creation)
	var imageToUse = opts.Image

	// Handle local directory first
	if opts.Repository.LocalPath != "" {
		if err := m.copyLocalDirectory(opts.Repository.LocalPath, codePath); err != nil {
			m.storage.DeleteWorkspaceStorage(workspaceUUID)
			m.portPool.Release(allocatedPort)
			return nil, err
		}

		// Check for .devcontainer.json and build image if exists
		if m.devcon != nil {
			devcontainerPath := filepath.Join(codePath, ".devcontainer.json")
			if _, err := os.Stat(devcontainerPath); err == nil {
				// Build custom image from devcontainer
				imageTag := fmt.Sprintf("codepod-%s:latest", workspaceUUID)
				builtImage, err := m.devcon.Build(&devcon.BuildOptions{
					WorkspacePath: codePath,
					ImageTag:     imageTag,
				})
				if err != nil {
					// Log but continue with default image
					fmt.Printf("Warning: failed to build devcontainer: %v, using default image\n", err)
				} else if builtImage != "" {
					imageToUse = builtImage
				}
			}
		}
	} else if opts.Repository.URL != "" {
		// Clone from git repository
		if err := m.cloneRepositoryOnHost(opts.Repository, codePath); err != nil {
			m.storage.DeleteWorkspaceStorage(workspaceUUID)
			m.portPool.Release(allocatedPort)
			return nil, err
		}

		// Check for .devcontainer.json and build image if exists
		if m.devcon != nil {
			devcontainerPath := filepath.Join(codePath, ".devcontainer.json")
			if _, err := os.Stat(devcontainerPath); err == nil {
				// Build custom image from devcontainer
				imageTag := fmt.Sprintf("codepod-%s:latest", workspaceUUID)
				builtImage, err := m.devcon.Build(&devcon.BuildOptions{
					WorkspacePath: codePath,
					ImageTag:     imageTag,
				})
				if err != nil {
					// Log but continue with default image
					fmt.Printf("Warning: failed to build devcontainer: %v, using default image\n", err)
				} else if builtImage != "" {
					imageToUse = builtImage
				}
			}
		}
	}

	// Create Docker container
	containerConfig := &docker.ContainerConfig{
		Name:    GetContainerName(name),
		Image:   imageToUse,
		Cmd:     []string{"sleep", "infinity"},
		Env:     []string{},
		Labels:  map[string]string{"codepod.workspace": name},
		PortBindings: map[string][]docker.PortBinding{
			"22/tcp": {{HostIP: "0.0.0.0", HostPort: fmt.Sprintf("%d", allocatedPort)}},
		},
		Binds: []string{
			m.storage.BindWSLStorage(workspaceUUID, "/workspace"),
		},
		NetworkMode: "bridge",
		Privileged:  false,
	}

	containerID, err := m.dockerClient.CreateContainer(containerConfig)
	if err != nil {
		m.storage.DeleteWorkspaceStorage(workspaceUUID)
		m.portPool.Release(allocatedPort)
		return nil, err
	}

	workspace := &types.Workspace{
		Name:        name,
		UUID:        workspaceUUID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		State:       types.WorkspaceStateCreated,
		Repository:  opts.Repository,
		IDE:         opts.IDE,
		Container:   types.Container{Image: imageToUse, Name: containerConfig.Name},
		Domain:      fmt.Sprintf("%s.local", name),
		SSH:         types.SSH{},
		Agent:       types.Agent{Port: 22001, Status: "stopped"},
		Port:        allocatedPort,
		StoragePath: storagePath,
		CodePath:    codePath,
	}

	// Save workspace
	if err := m.saveWorkspace(workspace); err != nil {
		m.dockerClient.RemoveContainer(containerID, true)
		m.storage.DeleteWorkspaceStorage(workspaceUUID)
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

	// Use UUID for storage binding
	storageID := workspace.UUID
	if storageID == "" {
		// Fallback to name for backwards compatibility
		storageID = name
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
				m.storage.BindWSLStorage(storageID, "/workspace"),
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

	// Clone repository if specified (for backwards compatibility)
	if workspace.Repository.URL != "" && workspace.CodePath == "" {
		if err := m.cloneRepository(workspace); err != nil {
			return nil, err
		}
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

	// Delete storage using UUID (or fallback to name for backwards compatibility)
	storageID := workspace.UUID
	if storageID == "" {
		storageID = name
	}
	if err := m.storage.DeleteWorkspaceStorage(storageID); err != nil {
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

// cloneRepositoryOnHost clones a Git repository on the host (before container creation)
func (m *Manager) cloneRepositoryOnHost(repo types.Repository, targetPath string) error {
	if repo.URL == "" {
		return nil
	}

	branch := repo.Branch
	if branch == "" {
		branch = "main"
	}

	// Create target directory
	if err := os.MkdirAll(targetPath, 0755); err != nil {
		return fmt.Errorf("failed to create code directory: %w", err)
	}

	// Clone the repository on host using git command
	cmd := exec.Command("git", "clone", "-b", branch, "--depth", "1", repo.URL, targetPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	return nil
}

// copyLocalDirectory copies a local directory to the workspace
func (m *Manager) copyLocalDirectory(sourcePath, targetPath string) error {
	// Validate source exists
	sourceInfo, err := os.Stat(sourcePath)
	if err != nil {
		return fmt.Errorf("local path does not exist: %w", err)
	}

	if !sourceInfo.IsDir() {
		return fmt.Errorf("local path is not a directory: %s", sourcePath)
	}

	// Create target directory
	if err := os.MkdirAll(targetPath, 0755); err != nil {
		return fmt.Errorf("failed to create code directory: %w", err)
	}

	// Copy directory using cp command
	cmd := exec.Command("cp", "-r", sourcePath+"/.", targetPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to copy local directory: %w", err)
	}

	return nil
}

// cloneRepository clones a Git repository into the workspace (legacy, for running containers)
func (m *Manager) cloneRepository(workspace *types.Workspace) error {
	if workspace.Repository.URL == "" {
		return nil
	}

	branch := workspace.Repository.Branch
	if branch == "" {
		branch = "main"
	}

	// Install git if not present
	if err := m.dockerClient.ExecInContainer(workspace.Container.Name, []string{"apt-get", "update"}); err != nil {
		return fmt.Errorf("failed to update apt: %w", err)
	}
	if err := m.dockerClient.ExecInContainer(workspace.Container.Name, []string{"apt-get", "install", "-y", "git"}); err != nil {
		return fmt.Errorf("failed to install git: %w", err)
	}

	// Get the storage path
	storagePath := m.storage.GetWSLStoragePath(workspace.UUID)
	targetPath := filepath.Join(storagePath, "projects")

	// Clone the repository
	cmd := []string{
		"git", "clone",
		"-b", branch,
		"--depth", "1",
		workspace.Repository.URL,
		targetPath,
	}

	err := m.dockerClient.ExecInContainer(workspace.Container.Name, cmd)
	if err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	return nil
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
