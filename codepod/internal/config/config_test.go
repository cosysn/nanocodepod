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
	SetConfigDir("/test/config")
	defer ResetConfigDir()

	dir, err := GetConfigDir()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dir != "/test/config" {
		t.Errorf("expected /test/config, got %s", dir)
	}
}

func TestGetConfigDir_WithoutOverride(t *testing.T) {
	// Make sure override is reset
	ResetConfigDir()

	// This will use os.UserHomeDir, which should work in test
	dir, err := GetConfigDir()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dir == "" {
		t.Error("expected non-empty directory")
	}
}

func TestResetConfigDir(t *testing.T) {
	SetConfigDir("/test/config")
	ResetConfigDir()

	// After reset, should try to get home dir (may fail in test env, but shouldn't panic)
	_, err := GetConfigDir()
	if err != nil && err.Error() != "failed to get home directory" {
		// Expected to potentially fail in test without home dir
	}
}

func TestGetWorkspacesDir(t *testing.T) {
	SetConfigDir("/test/config")
	defer ResetConfigDir()

	dir, err := GetWorkspacesDir()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := filepath.Join("/test/config", WorkspacesDir)
	if dir != expected {
		t.Errorf("expected %s, got %s", expected, dir)
	}
}

func TestGetKeysDir(t *testing.T) {
	SetConfigDir("/test/config")
	defer ResetConfigDir()

	dir, err := GetKeysDir()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := filepath.Join("/test/config", KeysDir)
	if dir != expected {
		t.Errorf("expected %s, got %s", expected, dir)
	}
}

func TestGetToolsDir(t *testing.T) {
	SetConfigDir("/test/config")
	defer ResetConfigDir()

	dir, err := GetToolsDir()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := filepath.Join("/test/config", ToolsDir)
	if dir != expected {
		t.Errorf("expected %s, got %s", expected, dir)
	}
}

func TestEnsureConfigDir(t *testing.T) {
	// Create temp dir
	tmpDir := t.TempDir()
	SetConfigDir(tmpDir)
	defer ResetConfigDir()

	err := EnsureConfigDir()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check that directories were created
	expectedDirs := []string{
		tmpDir,
		filepath.Join(tmpDir, WorkspacesDir),
		filepath.Join(tmpDir, KeysDir),
		filepath.Join(tmpDir, ToolsDir),
	}

	for _, dir := range expectedDirs {
		info, err := os.Stat(dir)
		if err != nil {
			t.Errorf("directory %s not created: %v", dir, err)
		} else if !info.IsDir() {
			t.Errorf("%s is not a directory", dir)
		}
	}
}

func TestLoadConfig_FileNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	SetConfigDir(tmpDir)
	defer ResetConfigDir()

	// Should return default config when file doesn't exist
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Version != "1.0" {
		t.Errorf("expected version 1.0, got %s", cfg.Version)
	}
}

func TestLoadConfig_InvalidFile(t *testing.T) {
	tmpDir := t.TempDir()
	SetConfigDir(tmpDir)
	defer ResetConfigDir()

	// Create invalid config file
	configPath := filepath.Join(tmpDir, ConfigFile)
	os.WriteFile(configPath, []byte("invalid: yaml: :content"), 0644)

	_, err := LoadConfig()
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}

func TestLoadConfig_ValidFile(t *testing.T) {
	tmpDir := t.TempDir()
	SetConfigDir(tmpDir)
	defer ResetConfigDir()

	// Create valid config file
	cfg := &types.Config{
		Version: "2.0",
		General: types.GeneralConfig{
			DefaultIDE: "goland",
			SSHPort:    3333,
		},
	}

	configPath := filepath.Join(tmpDir, ConfigFile)
	data, _ := yaml.Marshal(cfg)
	os.WriteFile(configPath, data, 0644)

	loaded, err := LoadConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if loaded.Version != "2.0" {
		t.Errorf("expected version 2.0, got %s", loaded.Version)
	}
}

func TestSaveConfig(t *testing.T) {
	tmpDir := t.TempDir()
	SetConfigDir(tmpDir)
	defer ResetConfigDir()

	cfg := &types.Config{
		Version: "3.0",
		General: types.GeneralConfig{
			DefaultIDE: "vscode",
			SSHPort:    4444,
		},
	}

	err := SaveConfig(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify file was created
	configPath := filepath.Join(tmpDir, ConfigFile)
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("config file not created: %v", err)
	}

	var loaded types.Config
	if err := yaml.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("failed to parse saved config: %v", err)
	}

	if loaded.Version != "3.0" {
		t.Errorf("expected version 3.0, got %s", loaded.Version)
	}
}

func TestGetDefaultConfig(t *testing.T) {
	cfg := GetDefaultConfig()

	if cfg.Version != "1.0" {
		t.Errorf("expected version 1.0, got %s", cfg.Version)
	}
	if cfg.WSL.Distribution != "Ubuntu-22.04" {
		t.Errorf("expected Ubuntu-22.04, got %s", cfg.WSL.Distribution)
	}
	if cfg.General.DefaultIDE != "vscode" {
		t.Errorf("expected vscode, got %s", cfg.General.DefaultIDE)
	}
	if cfg.PortPool.Start != 22000 {
		t.Errorf("expected 22000, got %d", cfg.PortPool.Start)
	}
	if cfg.PortPool.End != 22999 {
		t.Errorf("expected 22999, got %d", cfg.PortPool.End)
	}
}
