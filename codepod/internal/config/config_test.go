package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/codepod-io/codepod/internal/types"
	yaml "gopkg.in/yaml.v3"
)

func TestGetConfigDir_WithOverride(t *testing.T) {
	// Set override
	SetConfigDir("/test/config/dir")

	// Get config dir
	dir, err := GetConfigDir()

	// Verify
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dir != "/test/config/dir" {
		t.Fatalf("expected /test/config/dir, got %s", dir)
	}

	// Reset
	ResetConfigDir()
}

func TestGetConfigDir_Default(t *testing.T) {
	// Reset to ensure we're using default
	ResetConfigDir()

	// This should use os.UserHomeDir()
	// We can't easily test this without mocking, but we can verify it doesn't error
	_, err := GetConfigDir()
	if err != nil {
		t.Logf("Note: GetConfigDir failed (likely no home dir in test env): %v", err)
	}
}

func TestGetWorkspacesDir(t *testing.T) {
	SetConfigDir("/test/config")
	defer ResetConfigDir()

	dir, err := GetWorkspacesDir()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "/test/config/workspaces"
	if dir != expected {
		t.Fatalf("expected %s, got %s", expected, dir)
	}
}

func TestGetKeysDir(t *testing.T) {
	SetConfigDir("/test/config")
	defer ResetConfigDir()

	dir, err := GetKeysDir()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "/test/config/keys"
	if dir != expected {
		t.Fatalf("expected %s, got %s", expected, dir)
	}
}

func TestGetToolsDir(t *testing.T) {
	SetConfigDir("/test/config")
	defer ResetConfigDir()

	dir, err := GetToolsDir()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "/test/config/tools"
	if dir != expected {
		t.Fatalf("expected %s, got %s", expected, dir)
	}
}

func TestEnsureConfigDir(t *testing.T) {
	// Create temp dir
	tmpDir := t.TempDir()
	SetConfigDir(tmpDir)
	defer ResetConfigDir()

	// Ensure config dir
	err := EnsureConfigDir()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify directories exist
	dirs := []string{
		tmpDir,
		filepath.Join(tmpDir, "workspaces"),
		filepath.Join(tmpDir, "keys"),
		filepath.Join(tmpDir, "tools"),
	}

	for _, dir := range dirs {
		info, err := os.Stat(dir)
		if err != nil {
			t.Fatalf("directory %s should exist: %v", dir, err)
		}
		if !info.IsDir() {
			t.Fatalf("expected %s to be a directory", dir)
		}
	}
}

func TestLoadConfig_FileNotFound(t *testing.T) {
	SetConfigDir("/nonexistent")
	defer ResetConfigDir()

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should return default config
	if cfg == nil {
		t.Fatal("expected default config, got nil")
	}
	if cfg.Version != "1.0" {
		t.Fatalf("expected version 1.0, got %s", cfg.Version)
	}
}

func TestLoadConfig_InvalidFile(t *testing.T) {
	tmpDir := t.TempDir()
	SetConfigDir(tmpDir)
	defer ResetConfigDir()

	// Write invalid YAML
	configPath := filepath.Join(tmpDir, "config.yaml")
	err := os.WriteFile(configPath, []byte("invalid: yaml: content:"), 0644)
	if err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	_, err = LoadConfig()
	if err == nil {
		t.Fatal("expected error for invalid YAML, got nil")
	}
}

func TestLoadConfig_ValidFile(t *testing.T) {
	tmpDir := t.TempDir()
	SetConfigDir(tmpDir)
	defer ResetConfigDir()

	// Write valid YAML
	cfg := &types.Config{
		Version: "2.0",
		WSL: types.WSLConfig{
			Distribution: "Ubuntu-20.04",
			DockerHost:   "tcp://localhost:2376",
		},
		General: types.GeneralConfig{
			DefaultIDE: "jetbrains",
			SSHPort:    3333,
		},
		PortPool: types.PortPool{
			Start: 23000,
			End:   23999,
			Used:  []int{23000, 23001},
		},
		DataDir: "/custom/data/dir",
	}

	// Marshal and write
	data, err := yaml.Marshal(cfg)
	if err != nil {
		t.Fatalf("failed to marshal config: %v", err)
	}

	configPath := filepath.Join(tmpDir, "config.yaml")
	err = os.WriteFile(configPath, data, 0644)
	if err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// Load config
	loadedCfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify
	if loadedCfg.Version != "2.0" {
		t.Fatalf("expected version 2.0, got %s", loadedCfg.Version)
	}
	if loadedCfg.WSL.Distribution != "Ubuntu-20.04" {
		t.Fatalf("expected Ubuntu-20.04, got %s", loadedCfg.WSL.Distribution)
	}
	if loadedCfg.General.DefaultIDE != "jetbrains" {
		t.Fatalf("expected jetbrains, got %s", loadedCfg.General.DefaultIDE)
	}
}

func TestSaveConfig(t *testing.T) {
	tmpDir := t.TempDir()
	SetConfigDir(tmpDir)
	defer ResetConfigDir()

	cfg := &types.Config{
		Version: "3.0",
		DataDir: "/test/data",
	}

	err := SaveConfig(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify file exists
	configPath := filepath.Join(tmpDir, "config.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("config file should exist: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("config file should not be empty")
	}
}

func TestGetDefaultConfig(t *testing.T) {
	cfg := GetDefaultConfig()

	if cfg.Version != "1.0" {
		t.Fatalf("expected version 1.0, got %s", cfg.Version)
	}
	if cfg.WSL.Distribution != "Ubuntu-22.04" {
		t.Fatalf("expected Ubuntu-22.04, got %s", cfg.WSL.Distribution)
	}
	if cfg.WSL.DockerHost != "tcp://localhost:2375" {
		t.Fatalf("expected tcp://localhost:2375, got %s", cfg.WSL.DockerHost)
	}
	if cfg.General.DefaultIDE != "vscode" {
		t.Fatalf("expected vscode, got %s", cfg.General.DefaultIDE)
	}
	if cfg.General.SSHPort != 2222 {
		t.Fatalf("expected SSHPort 2222, got %d", cfg.General.SSHPort)
	}
	if cfg.PortPool.Start != 22000 {
		t.Fatalf("expected PortPool.Start 22000, got %d", cfg.PortPool.Start)
	}
	if cfg.PortPool.End != 22999 {
		t.Fatalf("expected PortPool.End 22999, got %d", cfg.PortPool.End)
	}
	if cfg.DataDir != "/root/.codepod" {
		t.Fatalf("expected /root/.codepod, got %s", cfg.DataDir)
	}
}
