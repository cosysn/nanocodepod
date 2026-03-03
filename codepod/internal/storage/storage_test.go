package storage

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/codepod-io/codepod/internal/config"
	"github.com/codepod-io/codepod/internal/types"
	"github.com/codepod-io/codepod/internal/wsl"
	yaml "gopkg.in/yaml.v3"
)

func setupTestStorage(t *testing.T) (*Manager, string, func()) {
	// Create a temp directory for testing
	tmpDir, err := os.MkdirTemp("", "codepod-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	// Create a test config file with the temp dir as data dir
	testDataDir := filepath.Join(tmpDir, "data")
	testCfg := &types.Config{
		Version: "1.0",
		WSL: types.WSLConfig{
			Distribution: "Ubuntu-22.04",
			DockerHost:   "tcp://localhost:2375",
		},
		General: types.GeneralConfig{
			DefaultIDE: "vscode",
			SSHPort:    2222,
		},
		PortPool: types.PortPool{
			Start: 22000,
			End:   22999,
			Used:  []int{},
		},
		DataDir: testDataDir,
	}

	// Write config file
	cfgBytes, _ := yaml.Marshal(testCfg)
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, cfgBytes, 0644); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to write config file: %v", err)
	}

	// Override config dir to use temp directory
	config.SetConfigDir(tmpDir)

	platform, _ := wsl.NewPlatform()

	// On Windows, we need to skip the WSL directory creation or it will fail
	// because the test data dir doesn't exist in WSL
	if runtime.GOOS == "windows" {
		// For Windows tests, we can't easily test storage creation that requires WSL
		// So we'll skip these tests
		t.Skip("storage tests requiring WSL not supported on Windows in CI")
		return nil, tmpDir, func() {
			config.ResetConfigDir()
			os.RemoveAll(tmpDir)
		}
	}

	mgr, err := New(platform)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to create storage manager: %v", err)
	}

	cleanup := func() {
		config.ResetConfigDir()
		os.RemoveAll(tmpDir)
	}

	return mgr, tmpDir, cleanup
}

func TestManager_CreateWorkspaceStorage(t *testing.T) {
	mgr, tmpDir, cleanup := setupTestStorage(t)
	defer cleanup()

	path, err := mgr.CreateWorkspaceStorage("test-ws")
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}

	expected := filepath.Join(tmpDir, "data", "workspaces", "test-ws")
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
	mgr, tmpDir, cleanup := setupTestStorage(t)
	defer cleanup()

	path := mgr.GetWorkspaceStorage("myproject")
	expected := filepath.Join(tmpDir, "data", "workspaces", "myproject")

	if path != expected {
		t.Errorf("expected %s, got %s", expected, path)
	}
}

func TestManager_DeleteWorkspaceStorage(t *testing.T) {
	mgr, _, cleanup := setupTestStorage(t)
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
	mgr, _, cleanup := setupTestStorage(t)
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
	mgr, _, cleanup := setupTestStorage(t)
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
	mgr, _, cleanup := setupTestStorage(t)
	defer cleanup()

	bind := mgr.BindWSLStorage("myproject", "/workspace")

	if bind == "" {
		t.Error("bind path should not be empty")
	}
}
