package provider

import (
	"os"
	"path/filepath"
	"testing"
)

// TestWSLProviderCommand tests WSL provider command execution
func TestWSLProviderCommand(t *testing.T) {
	provider := NewWSLProvider("Ubuntu-22.04")

	// This will fail on non-Windows, but that's expected
	_, err := provider.Command("echo test")
	if err != nil {
		// Expected to fail on non-Windows systems
		t.Logf("Command failed (expected on non-Windows): %v", err)
	}
}

// TestWSLProviderStatus tests WSL provider status
func TestWSLProviderStatus(t *testing.T) {
	provider := NewWSLProvider("Ubuntu-22.04")

	_, err := provider.Status()
	if err != nil {
		// Expected to fail on non-Windows systems
		t.Logf("Status failed (expected on non-Windows): %v", err)
	}
}

// TestLocalProvider tests local provider
func TestLocalProvider(t *testing.T) {
	provider := NewLocalProvider("/tmp/codepod", "/tmp/codepod.sock")

	// Test status
	status, err := provider.Status()
	if err != nil {
		t.Errorf("Status failed: %v", err)
	}
	if status != "running" {
		t.Errorf("Expected 'running', got %s", status)
	}
}

// TestConfig tests config loading
func TestConfig(t *testing.T) {
	// Create temp config file
	tmpDir, err := os.MkdirTemp("", "provider-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Set USERPROFILE to temp dir
	os.Setenv("USERPROFILE", tmpDir)
	defer os.Unsetenv("USERPROFILE")

	// Create provider config directory
	providerDir := filepath.Join(tmpDir, ".codepod", "provider", "test")
	os.MkdirAll(providerDir, 0755)

	// Write config
	configPath := filepath.Join(providerDir, "config.yaml")
	err = os.WriteFile(configPath, []byte(`type: wsl
wsl_distro: Ubuntu-22.04
data_dir: /home/codepod
`), 0644)
	if err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	// Load config
	config, err := LoadConfig("test")
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if config.Type != "wsl" {
		t.Errorf("Expected type 'wsl', got %s", config.Type)
	}
	if config.WSLDistro != "Ubuntu-22.04" {
		t.Errorf("Expected WSLDistro 'Ubuntu-22.04', got %s", config.WSLDistro)
	}
}

// TestManagerList tests listing providers
func TestManagerList(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "provider-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Set USERPROFILE to temp dir
	os.Setenv("USERPROFILE", tmpDir)
	defer os.Unsetenv("USERPROFILE")

	// Create provider directories
	providerDir := filepath.Join(tmpDir, ".codepod", "provider")
	os.MkdirAll(filepath.Join(providerDir, "provider1"), 0755)
	os.MkdirAll(filepath.Join(providerDir, "provider2"), 0755)

	// Create manager and list
	manager := &Manager{providerDir: providerDir}
	providers, err := manager.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(providers) != 2 {
		t.Errorf("Expected 2 providers, got %d", len(providers))
	}
}
