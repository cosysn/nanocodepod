package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/codepod-io/codepod/internal/config"
	"github.com/codepod-io/codepod/internal/types"
)

func TestGetConfigDir(t *testing.T) {
	dir, err := config.GetConfigDir()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	expected := filepath.Join(os.Getenv("HOME"), ".codepod")
	if dir != expected {
		t.Errorf("expected %s, got %s", expected, dir)
	}
}

func TestEnsureConfigDir(t *testing.T) {
	// Clean up test directory
	dir, _ := config.GetConfigDir()
	os.RemoveAll(dir)

	// Test creation
	err := config.EnsureConfigDir()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify directories exist
	dirs := []string{
		dir,
		filepath.Join(dir, "workspaces"),
		filepath.Join(dir, "keys"),
		filepath.Join(dir, "tools"),
	}

	for _, d := range dirs {
		if _, err := os.Stat(d); os.IsNotExist(err) {
			t.Errorf("directory %s does not exist", d)
		}
	}

	// Cleanup
	os.RemoveAll(dir)
}

func TestGetDefaultConfig(t *testing.T) {
	cfg := config.GetDefaultConfig()

	if cfg.Version != "1.0" {
		t.Errorf("expected version 1.0, got %s", cfg.Version)
	}

	if cfg.WSL.Distribution != "Ubuntu-22.04" {
		t.Errorf("expected Ubuntu-22.04, got %s", cfg.WSL.Distribution)
	}

	if cfg.WSL.DockerHost != "tcp://localhost:2375" {
		t.Errorf("expected tcp://localhost:2375, got %s", cfg.WSL.DockerHost)
	}

	if cfg.PortPool.Start != 22000 {
		t.Errorf("expected port pool start 22000, got %d", cfg.PortPool.Start)
	}

	if cfg.PortPool.End != 22999 {
		t.Errorf("expected port pool end 22999, got %d", cfg.PortPool.End)
	}
}

func TestSaveAndLoadConfig(t *testing.T) {
	// Setup
	dir, _ := config.GetConfigDir()
	os.RemoveAll(dir)
	config.EnsureConfigDir()

	// Test save
	cfg := &types.Config{
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
			Used:  []int{22000},
		},
	}

	err := config.SaveConfig(cfg)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Test load
	loaded, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if loaded.Version != cfg.Version {
		t.Errorf("expected version %s, got %s", cfg.Version, loaded.Version)
	}

	if loaded.WSL.Distribution != cfg.WSL.Distribution {
		t.Errorf("expected distribution %s, got %s", cfg.WSL.Distribution, loaded.WSL.Distribution)
	}

	if loaded.PortPool.Used[0] != 22000 {
		t.Errorf("expected used port 22000, got %d", loaded.PortPool.Used[0])
	}

	// Cleanup
	os.RemoveAll(dir)
}

func TestLoadConfigReturnsDefault(t *testing.T) {
	// Setup - ensure config doesn't exist
	dir, _ := config.GetConfigDir()
	os.RemoveAll(dir)

	// Test - should return default config
	cfg := config.GetDefaultConfig()
	if cfg == nil {
		t.Fatal("expected default config, got nil")
	}

	if cfg.Version != "1.0" {
		t.Errorf("expected version 1.0, got %s", cfg.Version)
	}
}

func TestWorkspacesDir(t *testing.T) {
	dir, err := config.GetWorkspacesDir()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	expected := filepath.Join(os.Getenv("HOME"), ".codepod", "workspaces")
	if dir != expected {
		t.Errorf("expected %s, got %s", expected, dir)
	}
}

func TestKeysDir(t *testing.T) {
	dir, err := config.GetKeysDir()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	expected := filepath.Join(os.Getenv("HOME"), ".codepod", "keys")
	if dir != expected {
		t.Errorf("expected %s, got %s", expected, dir)
	}
}

func TestToolsDir(t *testing.T) {
	dir, err := config.GetToolsDir()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	expected := filepath.Join(os.Getenv("HOME"), ".codepod", "tools")
	if dir != expected {
		t.Errorf("expected %s, got %s", expected, dir)
	}
}
