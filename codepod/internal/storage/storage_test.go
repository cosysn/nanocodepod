package storage

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/codepod-io/codepod/internal/config"
	"github.com/codepod-io/codepod/internal/wsl"
)

func setupTestStorage(t *testing.T) (*Manager, func()) {
	dir, _ := config.GetConfigDir()
	os.RemoveAll(dir)
	config.EnsureConfigDir()

	platform, _ := wsl.NewPlatform()
	mgr, err := New(platform)
	if err != nil {
		t.Fatalf("failed to create storage manager: %v", err)
	}

	cleanup := func() {
		os.RemoveAll(dir)
	}

	return mgr, cleanup
}

func TestManager_CreateWorkspaceStorage(t *testing.T) {
	mgr, cleanup := setupTestStorage(t)
	defer cleanup()

	path, err := mgr.CreateWorkspaceStorage("test-ws")
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}

	expected := filepath.Join(os.Getenv("HOME"), ".codepod", "workspaces", "test-ws")
	if path != expected {
		t.Errorf("expected %s, got %s", expected, path)
	}

	// Verify subdirectories
	subdirs := []string{"data", "home", "keys", "projects"}
	for _, subdir := range subdirs {
		subdirPath := filepath.Join(path, subdir)
		if _, err := os.Stat(subdirPath); os.IsNotExist(err) {
			t.Errorf("subdirectory %s not created", subdir)
		}
	}
}

func TestManager_GetWorkspaceStorage(t *testing.T) {
	mgr, cleanup := setupTestStorage(t)
	defer cleanup()

	path := mgr.GetWorkspaceStorage("myproject")
	expected := filepath.Join(os.Getenv("HOME"), ".codepod", "workspaces", "myproject")

	if path != expected {
		t.Errorf("expected %s, got %s", expected, path)
	}
}

func TestManager_DeleteWorkspaceStorage(t *testing.T) {
	mgr, cleanup := setupTestStorage(t)
	defer cleanup()

	// Create first
	mgr.CreateWorkspaceStorage("delete-test")

	// Delete
	err := mgr.DeleteWorkspaceStorage("delete-test")
	if err != nil {
		t.Fatalf("failed to delete: %v", err)
	}

	// Verify deleted
	if mgr.WorkspaceStorageExists("delete-test") {
		t.Error("storage should not exist after deletion")
	}
}

func TestManager_WorkspaceStorageExists(t *testing.T) {
	mgr, cleanup := setupTestStorage(t)
	defer cleanup()

	if mgr.WorkspaceStorageExists("exists-test") {
		t.Error("should not exist before creation")
	}

	mgr.CreateWorkspaceStorage("exists-test")

	if !mgr.WorkspaceStorageExists("exists-test") {
		t.Error("should exist after creation")
	}
}

func TestManager_ListWorkspacesStorage(t *testing.T) {
	mgr, cleanup := setupTestStorage(t)
	defer cleanup()

	mgr.CreateWorkspaceStorage("ws1")
	mgr.CreateWorkspaceStorage("ws2")

	workspaces, err := mgr.ListWorkspacesStorage()
	if err != nil {
		t.Fatalf("failed to list: %v", err)
	}

	if len(workspaces) != 2 {
		t.Errorf("expected 2 workspaces, got %d", len(workspaces))
	}
}

func TestManager_BindWSLStorage(t *testing.T) {
	mgr, cleanup := setupTestStorage(t)
	defer cleanup()

	bind := mgr.BindWSLStorage("myproject", "/workspace")

	if bind == "" {
		t.Error("bind path should not be empty")
	}
}
