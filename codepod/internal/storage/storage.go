package storage

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/codepod-io/codepod/internal/config"
	"github.com/codepod-io/codepod/internal/wsl"
)

// Manager handles persistent storage for workspaces
type Manager struct {
	basePath string
	platform *wsl.Platform
}

// New creates a new storage manager
func New(platform *wsl.Platform) (*Manager, error) {
	// Load config to get data directory
	cfg, err := config.LoadConfig()
	var basePath string

	if err != nil || cfg.DataDir == "" {
		// Fallback to old behavior
		configDir, err := getConfigDir()
		if err != nil {
			return nil, err
		}
		basePath = filepath.Join(configDir, "workspaces")
	} else {
		// Use data_dir from config
		basePath = filepath.Join(cfg.DataDir, "workspaces")
		fmt.Printf("[DEBUG] Using data_dir from config: %s\n", cfg.DataDir)
	}

	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	return &Manager{
		basePath: basePath,
		platform: platform,
	}, nil
}

// CreateWorkspaceStorage creates storage for a new workspace
func (m *Manager) CreateWorkspaceStorage(workspaceName string) (string, error) {
	workspacePath := filepath.Join(m.basePath, workspaceName)

	if err := os.MkdirAll(workspacePath, 0755); err != nil {
		return "", fmt.Errorf("failed to create workspace directory: %w", err)
	}

	// Create subdirectories
	subdirs := []string{"data", "home", "keys", "projects"}
	for _, subdir := range subdirs {
		subdirPath := filepath.Join(workspacePath, subdir)
		if err := os.MkdirAll(subdirPath, 0755); err != nil {
			return "", fmt.Errorf("failed to create subdirectory %s: %w", subdir, err)
		}
	}

	return workspacePath, nil
}

// GetWorkspaceStorage returns the storage path for a workspace
func (m *Manager) GetWorkspaceStorage(workspaceName string) string {
	return filepath.Join(m.basePath, workspaceName)
}

// DeleteWorkspaceStorage deletes storage for a workspace
func (m *Manager) DeleteWorkspaceStorage(workspaceName string) error {
	workspacePath := filepath.Join(m.basePath, workspaceName)
	if err := os.RemoveAll(workspacePath); err != nil {
		return fmt.Errorf("failed to delete workspace storage: %w", err)
	}
	return nil
}

// WorkspaceStorageExists checks if workspace storage exists
func (m *Manager) WorkspaceStorageExists(workspaceName string) bool {
	workspacePath := filepath.Join(m.basePath, workspaceName)
	_, err := os.Stat(workspacePath)
	return err == nil
}

// ListWorkspacesStorage lists all workspace storage directories
func (m *Manager) ListWorkspacesStorage() ([]string, error) {
	entries, err := os.ReadDir(m.basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read storage directory: %w", err)
	}

	var workspaces []string
	for _, entry := range entries {
		if entry.IsDir() {
			workspaces = append(workspaces, entry.Name())
		}
	}
	return workspaces, nil
}

// CreateWSLStorage creates storage in WSL for mounting
func (m *Manager) CreateWSLStorage(workspaceName string) error {
	wslPath := filepath.Join("/home", getCurrentUser(), ".codepod", "workspaces", workspaceName)

	switch m.platform.Type {
	case wsl.PlatformWSL:
		_, err := m.platform.RunCommand(fmt.Sprintf("mkdir -p %s", wslPath))
		return err
	case wsl.PlatformLinux:
		// On Linux, just use local path
		localPath := filepath.Join(m.basePath, workspaceName)
		return os.MkdirAll(localPath, 0755)
	case wsl.PlatformWindows:
		// On Windows, create storage in WSL
		wslPath := filepath.Join("/home", getCurrentUser(), ".codepod", "workspaces", workspaceName)
		_, err := m.platform.RunCommand(fmt.Sprintf("mkdir -p %s", wslPath))
		return err
	default:
		return fmt.Errorf("unsupported platform: %s", m.platform.Type)
	}
}

// GetWSLStoragePath returns the storage path for mounting to container
func (m *Manager) GetWSLStoragePath(workspaceName string) string {
	switch m.platform.Type {
	case wsl.PlatformWSL:
		return filepath.Join("/home", getCurrentUser(), ".codepod", "workspaces", workspaceName)
	case wsl.PlatformLinux:
		return filepath.Join(m.basePath, workspaceName)
	case wsl.PlatformWindows:
		// On Windows, storage is in WSL
		return filepath.Join("/home", getCurrentUser(), ".codepod", "workspaces", workspaceName)
	default:
		return filepath.Join(m.basePath, workspaceName)
	}
}

// BindWSLStorage binds storage path to container mount point
func (m *Manager) BindWSLStorage(workspaceName, containerPath string) string {
	storagePath := m.GetWSLStoragePath(workspaceName)
	return fmt.Sprintf("%s:%s", storagePath, containerPath)
}

func getConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, ".codepod"), nil
}

func getCurrentUser() string {
	if user := os.Getenv("USER"); user != "" {
		return user
	}
	if user := os.Getenv("USERNAME"); user != "" {
		return user
	}
	return "ubuntu"
}
