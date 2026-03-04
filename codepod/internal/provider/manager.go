package provider

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Manager manages providers
type Manager struct {
	providerDir string
}

// NewManager creates a new provider manager
func NewManager() *Manager {
	return &Manager{
		providerDir: filepath.Join(os.Getenv("USERPROFILE"), ".codepod", "provider"),
	}
}

// List returns all available providers
func (m *Manager) List() ([]string, error) {
	entries, err := os.ReadDir(m.providerDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}

	var providers []string
	for _, entry := range entries {
		if entry.IsDir() {
			providers = append(providers, entry.Name())
		}
	}
	return providers, nil
}

// Get returns a provider by name
func (m *Manager) Get(name string) (Provider, error) {
	config, err := LoadConfig(name)
	if err != nil {
		return nil, fmt.Errorf("failed to load provider config: %w", err)
	}

	switch config.Type {
	case "wsl":
		return NewWSLProvider(config.WSLDistro), nil
	case "local":
		return NewLocalProvider(config.DataDir, config.SocketPath), nil
	default:
		return nil, fmt.Errorf("unknown provider type: %s", config.Type)
	}
}

// Add creates a new provider
func (m *Manager) Add(name string, config *Config) error {
	providerPath := filepath.Join(m.providerDir, name)
	if err := os.MkdirAll(providerPath, 0755); err != nil {
		return fmt.Errorf("failed to create provider directory: %w", err)
	}

	configPath := filepath.Join(providerPath, "config.yaml")
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}
