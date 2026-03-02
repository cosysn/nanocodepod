package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/codepod-io/codepod/internal/types"
	yaml "gopkg.in/yaml.v3"
)

const (
	ConfigDir = ".codepod"
	ConfigFile = "config.yaml"
	WorkspacesDir = "workspaces"
	KeysDir = "keys"
	ToolsDir = "tools"
)

var (
	configDirOverride string
	configDirOnce sync.Once
)

// SetConfigDir sets the config directory override
func SetConfigDir(dir string) {
	configDirOverride = dir
}

// GetConfigDir returns the config directory path
func GetConfigDir() (string, error) {
	if configDirOverride != "" {
		return configDirOverride, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, ConfigDir), nil
}

// ResetConfigDir resets the config directory override (for testing)
func ResetConfigDir() {
	configDirOverride = ""
}

// GetWorkspacesDir returns the workspaces directory path
func GetWorkspacesDir() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, WorkspacesDir), nil
}

// GetKeysDir returns the keys directory path
func GetKeysDir() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, KeysDir), nil
}

// GetToolsDir returns the tools directory path
func GetToolsDir() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, ToolsDir), nil
}

// EnsureConfigDir ensures the config directory exists
func EnsureConfigDir() error {
	configDir, err := GetConfigDir()
	if err != nil {
		return err
	}

	dirs := []string{
		configDir,
		filepath.Join(configDir, WorkspacesDir),
		filepath.Join(configDir, KeysDir),
		filepath.Join(configDir, ToolsDir),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}

// LoadConfig loads the configuration from file
func LoadConfig() (*types.Config, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return nil, err
	}

	configPath := filepath.Join(configDir, ConfigFile)
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return default config if file doesn't exist
			return GetDefaultConfig(), nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg types.Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &cfg, nil
}

// SaveConfig saves the configuration to file
func SaveConfig(cfg *types.Config) error {
	configDir, err := GetConfigDir()
	if err != nil {
		return err
	}

	if err := EnsureConfigDir(); err != nil {
		return err
	}

	configPath := filepath.Join(configDir, ConfigFile)
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GetDefaultConfig returns the default configuration
func GetDefaultConfig() *types.Config {
	return &types.Config{
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
	}
}
