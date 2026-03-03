package workspace

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/codepod-io/codepod/internal/agent"
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
	dockerClient docker.DockerClient
	platform     *wsl.Platform
	storage      *storage.Manager
	portPool     *port.Pool
	devcon       *devcon.Devcon
	config       *types.Config
}

// New creates a new workspace manager
func New() (*Manager, error) {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		// If config loading fails, use default config but log warning
		fmt.Printf("[WARN] Failed to load config: %v, using defaults\n", err)
		cfg = config.GetDefaultConfig()
	}

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
		config:       cfg,
	}, nil
}

// NewWithMock creates a workspace manager with a mock Docker client for testing
func NewWithMock(mockClient docker.DockerClient) (*Manager, error) {
	platform, err := wsl.NewPlatform()
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
		dockerClient: mockClient,
		platform:     platform,
		storage:      storageManager,
		portPool:     portPool,
		devcon:       nil,
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
	var codePath string
	if runtime.GOOS == "windows" {
		codePath = strings.Join([]string{storagePath, "code"}, "/")
	} else {
		codePath = filepath.Join(storagePath, "code")
	}

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
			var devcontainerPath string
			if runtime.GOOS == "windows" {
				devcontainerPath = strings.Join([]string{codePath, ".devcontainer.json"}, "/")
			} else {
				devcontainerPath = filepath.Join(codePath, ".devcontainer.json")
			}
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
			var devcontainerPath string
			if runtime.GOOS == "windows" {
				devcontainerPath = strings.Join([]string{codePath, ".devcontainer.json"}, "/")
			} else {
				devcontainerPath = filepath.Join(codePath, ".devcontainer.json")
			}
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

	// Determine container entrypoint and environment
	// Note: Agent injection happens in Start() method, not here
	// Create always uses sleep infinity as entrypoint
	var containerCmd []string
	var containerEnv []string

	if opts.InjectAgent {
		// Create with sleep infinity, agent will be started in Start() method
		containerCmd = []string{"sleep", "infinity"}
		containerEnv = []string{}
	} else {
		// Default: use sleep infinity
		containerCmd = []string{"sleep", "infinity"}
		containerEnv = []string{}
	}

	// Create Docker container
	containerConfig := &docker.ContainerConfig{
		Name:    GetContainerName(name),
		Image:   imageToUse,
		Cmd:     containerCmd,
		Env:     containerEnv,
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

	// If agent injection is enabled, copy agent binary to container
	if opts.InjectAgent {
		agentPath, err := m.getAgentBinaryPath()
		if err != nil {
			m.dockerClient.RemoveContainer(containerID, true)
			m.storage.DeleteWorkspaceStorage(workspaceUUID)
			m.portPool.Release(allocatedPort)
			return nil, fmt.Errorf("failed to find agent binary: %w", err)
		}

		// Copy agent to container
		if err := m.dockerClient.CopyToContainer(containerID, agentPath, "/usr/local/bin/codepod-agent"); err != nil {
			m.dockerClient.RemoveContainer(containerID, true)
			m.storage.DeleteWorkspaceStorage(workspaceUUID)
			m.portPool.Release(allocatedPort)
			return nil, fmt.Errorf("failed to copy agent to container: %w", err)
		}

		// Make agent executable
		if err := m.dockerClient.ExecInContainer(containerID, []string{"chmod", "+x", "/usr/local/bin/codepod-agent"}); err != nil {
			m.dockerClient.RemoveContainer(containerID, true)
			m.storage.DeleteWorkspaceStorage(workspaceUUID)
			m.portPool.Release(allocatedPort)
			return nil, fmt.Errorf("failed to make agent executable: %w", err)
		}

		// Install openssh-client for ssh-keygen (needed for host key generation)
		if err := m.dockerClient.ExecInContainer(containerID, []string{"apt-get", "update"}); err != nil {
			fmt.Printf("Warning: failed to update apt: %v\n", err)
		}
		if err := m.dockerClient.ExecInContainer(containerID, []string{"apt-get", "install", "-y", "openssh-client"}); err != nil {
			fmt.Printf("Warning: failed to install openssh-client: %v\n", err)
		}
	}

	agentStatus := "stopped"
	if opts.InjectAgent {
		agentStatus = "running"
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
		Agent:       types.Agent{Port: allocatedPort, Status: agentStatus},
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
func (m *Manager) Start(name string, injectAgent bool) (*types.Workspace, error) {
	workspace, err := m.Get(name)
	if err != nil {
		return nil, err
	}

	// Use UUID for storage binding
	storageID := workspace.UUID
	if storageID == "" {
		// Fallback to name for backwards compatibility
		storageID = name
	}

	// Determine if we should use agent
	useAgent := injectAgent || workspace.Agent.Status == "running"

	exists := m.dockerClient.ContainerExists(workspace.Container.Name)

	if !exists {
		// Create container with sleep infinity (default)
		containerCmd := []string{"sleep", "infinity"}
		containerEnv := []string{}

		// Create container
		containerConfig := &docker.ContainerConfig{
			Name:    workspace.Container.Name,
			Image:   workspace.Container.Image,
			Cmd:     containerCmd,
			Env:     containerEnv,
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
		if _, err := m.dockerClient.CreateContainer(containerConfig); err != nil {
			return nil, err
		}
	}

	// Start the container
	if err := m.dockerClient.StartContainer(workspace.Container.Name); err != nil {
		return nil, err
	}

	// If agent is enabled, copy agent binary and start it
	if useAgent {
		// Always copy the agent binary (to ensure we have the latest version)
		agentPath, err := m.getAgentBinaryPath()
		if err != nil {
			return nil, fmt.Errorf("failed to find agent binary: %w", err)
		}
		if err := m.dockerClient.CopyToContainer(workspace.Container.Name, agentPath, "/usr/local/bin/codepod-agent"); err != nil {
			return nil, fmt.Errorf("failed to copy agent to container: %w", err)
		}
		// Make executable
		if err := m.dockerClient.ExecInContainer(workspace.Container.Name, []string{"chmod", "+x", "/usr/local/bin/codepod-agent"}); err != nil {
			return nil, fmt.Errorf("failed to make agent executable: %w", err)
		}

		// Install openssh-client for ssh-keygen (needed for host key generation)
		if err := m.dockerClient.ExecInContainer(workspace.Container.Name, []string{"apt-get", "update"}); err != nil {
			fmt.Printf("Warning: failed to update apt: %v\n", err)
		}
		if err := m.dockerClient.ExecInContainer(workspace.Container.Name, []string{"apt-get", "install", "-y", "openssh-client"}); err != nil {
			fmt.Printf("Warning: failed to install openssh-client: %v\n", err)
		}

		// Start the agent as a background process
		// Note: Inside the container, agent listens on port 22 (SSH port)
		// The port mapping (container:22 -> host:workspace.Port) is handled by Docker
		agentPort := 22
		if envPort := os.Getenv("CODEPOD_AGENT_PORT"); envPort != "" {
			if port, err := strconv.Atoi(envPort); err == nil && port > 0 && port < 1000 {
				// Only use custom port if it's a low port (for container internal use)
				agentPort = port
			}
		}
		agentPassword := "codepod"
		if envPass := os.Getenv("CODEPOD_AGENT_PASSWORD"); envPass != "" {
			agentPassword = envPass
		}

		// Kill sleep and start agent as background process
		startAgentCmd := []string{
			"bash", "-c",
			fmt.Sprintf("pkill -f 'sleep infinity' || true; nohup /usr/local/bin/codepod-agent --port %d --password %s > /tmp/agent.log 2>&1 &",
				agentPort, agentPassword),
		}
		if err := m.dockerClient.ExecInContainerDetached(workspace.Container.Name, startAgentCmd); err != nil {
			fmt.Printf("Warning: failed to start agent: %v\n", err)
		}
	}

	// Clone repository if specified (for backwards compatibility)
	if workspace.Repository.URL != "" && workspace.CodePath == "" {
		if err := m.cloneRepository(workspace); err != nil {
			return nil, err
		}
	}

	// Update agent status
	if useAgent {
		workspace.Agent.Status = "running"
	} else {
		workspace.Agent.Status = "stopped"
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
		// Skip directories, look for .yaml files
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if len(name) > 5 && name[len(name)-5:] == ".yaml" {
			workspaceName := name[:len(name)-5]
			workspace, err := m.Get(workspaceName)
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
	InjectAgent bool
}

// GetWorkspaceDir returns the workspace directory
func getWorkspacesDir() (string, error) {
	// Use data_dir from config
	cfg, err := config.LoadConfig()
	if err != nil || cfg.DataDir == "" {
		// Fallback to old behavior
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
	// Use data_dir from config
	var workspacesDir string
	if runtime.GOOS == "windows" {
		workspacesDir = strings.Join([]string{cfg.DataDir, "workspaces"}, "/")
	} else {
		workspacesDir = filepath.Join(cfg.DataDir, "workspaces")
	}
	if runtime.GOOS == "windows" {
		// Create in WSL
		distro := wsl.GetWSLDistributionFromConfig()
		wslInstance := wsl.New(distro)
		mkdirCmd := fmt.Sprintf("mkdir -p %s", workspacesDir)
		_, err := wslInstance.RunCommand(mkdirCmd)
		if err != nil {
			return "", fmt.Errorf("failed to create workspaces dir in WSL: %w", err)
		}
	} else {
		if err := os.MkdirAll(workspacesDir, 0755); err != nil {
			return "", err
		}
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

	// On Windows, run git clone in WSL
	if runtime.GOOS == "windows" {
		distro := wsl.GetWSLDistributionFromConfig()
		wslInstance := wsl.New(distro)

		// Create target directory in WSL
		mkdirCmd := fmt.Sprintf("mkdir -p %s", targetPath)
		_, err := wslInstance.RunCommand(mkdirCmd)
		if err != nil {
			return fmt.Errorf("failed to create code directory in WSL: %w", err)
		}

		// Clone the repository in WSL
		cloneCmd := fmt.Sprintf("cd %s && git clone -b %s --depth 1 %s .", targetPath, branch, repo.URL)
		output, err := wslInstance.RunCommand(cloneCmd)
		if err != nil {
			return fmt.Errorf("failed to clone repository in WSL: %w, output: %s", err, output)
		}
		return nil
	}

	// On Linux/WSL, use native git
	if err := os.MkdirAll(targetPath, 0755); err != nil {
		return fmt.Errorf("failed to create code directory: %w", err)
	}

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

	// On Windows, copy to WSL
	if runtime.GOOS == "windows" {
		distro := wsl.GetWSLDistributionFromConfig()
		wslInstance := wsl.New(distro)

		// Create target directory in WSL
		mkdirCmd := fmt.Sprintf("mkdir -p %s", targetPath)
		_, err := wslInstance.RunCommand(mkdirCmd)
		if err != nil {
			return fmt.Errorf("failed to create code directory in WSL: %w", err)
		}

		// Copy directory using cp command in WSL
		cpCmd := fmt.Sprintf("cp -r %s/. %s", sourcePath, targetPath)
		output, err := wslInstance.RunCommand(cpCmd)
		if err != nil {
			return fmt.Errorf("failed to copy local directory to WSL: %w, output: %s", err, output)
		}
		return nil
	}

	// On Linux/WSL, use native cp
	if err := os.MkdirAll(targetPath, 0755); err != nil {
		return fmt.Errorf("failed to create code directory: %w", err)
	}

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

// injectAgent injects and starts the agent in the workspace container
func (m *Manager) injectAgent(workspace *types.Workspace) error {
	// Get agent binary path
	agentPath, err := exec.LookPath("codepod-agent")
	if err != nil {
		// Try current directory
		agentPath = "codepod-agent"
		if _, err := os.Stat(agentPath); os.IsNotExist(err) {
			return fmt.Errorf("codepod-agent not found: %w", err)
		}
	}

	// Get absolute path
	agentPath, err = filepath.Abs(agentPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Check if agent already exists in container
	if agent.AgentExistsInContainer(workspace.Container.Name) {
		fmt.Printf("Agent already exists in container %s\n", workspace.Container.Name)
		return nil
	}

	// Install openssh-server required for agent
	if err := m.dockerClient.ExecInContainer(workspace.Container.Name, []string{"apt-get", "update"}); err != nil {
		return fmt.Errorf("failed to update apt: %w", err)
	}
	if err := m.dockerClient.ExecInContainer(workspace.Container.Name, []string{"apt-get", "install", "-y", "openssh-server"}); err != nil {
		return fmt.Errorf("failed to install openssh-server: %w", err)
	}

	// Copy agent to container
	if err := agent.CopyToContainer(workspace.Container.Name, agentPath); err != nil {
		return fmt.Errorf("failed to copy agent to container: %w", err)
	}

	// Start agent in container
	agentPort := workspace.Agent.Port
	if agentPort == 0 {
		agentPort = 22001
	}

	// Generate a random password or use default
	agentPassword := "codepod"

	if err := agent.StartAgentInContainer(workspace.Container.Name, agentPort, agentPassword); err != nil {
		return fmt.Errorf("failed to start agent in container: %w", err)
	}

	// Update workspace agent status
	workspace.Agent.Status = "running"
	workspace.Agent.Port = agentPort

	fmt.Printf("Agent started on port %d in container %s\n", agentPort, workspace.Container.Name)

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

// getAgentBinaryPath returns the path to the codepod-agent binary
func (m *Manager) getAgentBinaryPath() (string, error) {
	// First try to find it in PATH
	agentPath, err := exec.LookPath("codepod-agent")
	if err == nil {
		return agentPath, nil
	}

	// Check common locations
	locations := []string{
		"/tmp/codepod-agent",
		filepath.Join(os.Getenv("HOME"), "go/bin/codepod-agent"),
		"/usr/local/bin/codepod-agent",
		"/usr/bin/codepod-agent",
		filepath.Join(os.Getenv("HOME"), "codepod-agent"),
	}

	// On Windows, also check WSL paths
	if runtime.GOOS == "windows" {
		// Try to detect WSL architecture and find matching agent
		agentPath, err := m.findOrCopyAgentToWSL()
		if err == nil {
			return agentPath, nil
		}
		locations = append(locations, "/home/"+getCurrentUser()+"/go/bin/codepod-agent")
	}

	for _, loc := range locations {
		if _, err := os.Stat(loc); err == nil {
			return loc, nil
		}
	}

	return "", fmt.Errorf("codepod-agent binary not found")
}

// getWSLArchitecture returns the architecture of the WSL distribution (amd64 or arm64)
func (m *Manager) getWSLArchitecture() (string, error) {
	// Run uname -m in WSL to get architecture
	distro := wsl.GetWSLDistributionFromConfig()
	wslInstance := wsl.New(distro)
	output, err := wslInstance.RunCommand("uname -m")
	if err != nil {
		return "", fmt.Errorf("failed to detect WSL architecture: %w", err)
	}
	output = strings.TrimSpace(output)
	// Map architecture names
	if output == "x86_64" || output == "amd64" {
		return "amd64", nil
	}
	if output == "aarch64" || output == "arm64" {
		return "arm64", nil
	}
	return "", fmt.Errorf("unsupported WSL architecture: %s", output)
}

// findOrCopyAgentToWSL finds the matching Linux agent and copies it to WSL
func (m *Manager) findOrCopyAgentToWSL() (string, error) {
	// Get WSL architecture
	arch, err := m.getWSLArchitecture()
	if err != nil {
		return "", err
	}

	// Determine agent filename based on architecture
	agentFilename := "codepod-agent-" + arch

	// Get the directory where codepod.exe is running from
	execPath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("failed to get executable path: %w", err)
	}
	execDir := filepath.Dir(execPath)

	// Check if agent exists in the same directory
	localAgentPath := filepath.Join(execDir, agentFilename)
	fmt.Printf("[DEBUG] Looking for agent: %s\n", localAgentPath)
	if _, err := os.Stat(localAgentPath); err != nil {
		// Try without dash (some builds use codepod-agent-amd64, some use codepod-agentamd64)
		localAgentPath = filepath.Join(execDir, "codepod-agent-"+arch)
		fmt.Printf("[DEBUG] Trying agent: %s\n", localAgentPath)
		if _, err := os.Stat(localAgentPath); err != nil {
			return "", fmt.Errorf("agent binary %s not found in %s", agentFilename, execDir)
		}
	}
	fmt.Printf("[DEBUG] Found agent at: %s\n", localAgentPath)

	// Get data directory from config, default to /root/.codepod
	dataDir := m.config.DataDir
	if dataDir == "" {
		dataDir = "/root/.codepod"
	}

	// Target path in WSL (agent subdirectory under data dir)
	wslAgentDir := strings.Join([]string{dataDir, "agent"}, "/")
	wslAgentPath := strings.Join([]string{wslAgentDir, "codepod-agent"}, "/")
	fmt.Printf("[DEBUG] Target WSL path: %s\n", wslAgentPath)

	// Copy to WSL
	distro := wsl.GetWSLDistributionFromConfig()
	wslInstance := wsl.New(distro)

	// First check if already exists
	checkCmd := fmt.Sprintf("test -f %s && echo 'exists'", wslAgentPath)
	fmt.Printf("[DEBUG] Check agent exists: %s\n", checkCmd)
	output, _ := wslInstance.RunCommand(checkCmd)
	fmt.Printf("[DEBUG] Agent exists check output: '%s'\n", output)
	if strings.TrimSpace(output) == "exists" {
		fmt.Printf("[DEBUG] Agent already exists, returning\n")
		return wslAgentPath, nil
	}

	// Create agent directory in WSL
	fmt.Printf("[DEBUG] Creating agent directory: %s\n", wslAgentDir)
	mkdirCmd := fmt.Sprintf("mkdir -p %s", wslAgentDir)
	_, err = wslInstance.RunCommand(mkdirCmd)
	if err != nil {
		return "", fmt.Errorf("failed to create agent directory in WSL: %w", err)
	}
	fmt.Printf("[DEBUG] Agent directory created\n")

	// Copy to WSL agent directory
	// Use \\wsl$\<distro>\ path to copy directly from Windows
	distro = wsl.GetWSLDistributionFromConfig()
	windowsWSLPath := fmt.Sprintf("\\\\wsl$\\%s%s", distro, wslAgentPath)
	fmt.Printf("[DEBUG] Copying agent from %s to %s\n", localAgentPath, windowsWSLPath)

	// Copy file using Windows file system
	err = copyFile(localAgentPath, windowsWSLPath)
	if err != nil {
		return "", fmt.Errorf("failed to copy agent to WSL: %w", err)
	}

	return wslAgentPath, nil
}

// getCurrentUser returns the current username
func getCurrentUser() string {
	if user := os.Getenv("USER"); user != "" {
		return user
	}
	if user := os.Getenv("USERNAME"); user != "" {
		return user
	}
	return "ubuntu"
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	fmt.Printf("[DEBUG] copyFile: src=%s, dst=%s\n", src, dst)

	sourceFile, err := os.Open(src)
	if err != nil {
		fmt.Printf("[DEBUG] copyFile: failed to open src: %v\n", err)
		return err
	}
	defer sourceFile.Close()

	// Create destination directory if needed
	dstDir := filepath.Dir(dst)
	fmt.Printf("[DEBUG] copyFile: dstDir=%s\n", dstDir)
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		fmt.Printf("[DEBUG] copyFile: failed to create dstDir: %v\n", err)
		return err
	}

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}
