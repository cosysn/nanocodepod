package storage

import (
	"fmt"
	"os"
	"path/filepath"
)

// Manager manages workspace storage
type Manager struct {
	baseDir string
}

// NewManager creates a new storage manager
func NewManager(baseDir string) *Manager {
	return &Manager{
		baseDir: baseDir,
	}
}

// EnsureStorage creates the storage directory for a workspace
func (m *Manager) EnsureStorage(workspaceName string) (string, error) {
	workspaceDir := filepath.Join(m.baseDir, workspaceName)
	if err := os.MkdirAll(workspaceDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create workspace directory: %w", err)
	}
	return workspaceDir, nil
}

// GetStoragePath returns the storage path for a workspace
func (m *Manager) GetStoragePath(workspaceName string) string {
	return filepath.Join(m.baseDir, workspaceName)
}

// DeleteStorage deletes the storage for a workspace
func (m *Manager) DeleteStorage(workspaceName string) error {
	workspaceDir := filepath.Join(m.baseDir, workspaceName)
	if err := os.RemoveAll(workspaceDir); err != nil {
		return fmt.Errorf("failed to delete workspace directory: %w", err)
	}
	return nil
}

// StorageExists checks if storage exists for a workspace
func (m *Manager) StorageExists(workspaceName string) bool {
	workspaceDir := filepath.Join(m.baseDir, workspaceName)
	_, err := os.Stat(workspaceDir)
	return err == nil
}
