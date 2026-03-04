package devcon

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config represents devcontainer.json configuration
type Config struct {
	Image           string   `json:"image,omitempty"`
	Build           *Build   `json:"build,omitempty"`
	Features        Features `json:"features,omitempty"`
	OverrideCommand string   `json:"overrideCommand,omitempty"`
	Mounts          []string `json:"mounts,omitempty"`
	WorkspaceFolder string   `json:"workspaceFolder,omitempty"`
}

// Build represents build configuration
type Build struct {
	Dockerfile string            `json:"dockerfile,omitempty"`
	Context    string            `json:"context,omitempty"`
	Args      map[string]string `json:"args,omitempty"`
}

// Features represents devcontainer features
type Features map[string]map[string]interface{}

// LoadConfig loads devcontainer.json from a directory
func LoadConfig(workspaceDir string) (*Config, error) {
	configPath := filepath.Join(workspaceDir, ".devcontainer", "devcontainer.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read devcontainer.json: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse devcontainer.json: %w", err)
	}

	return &config, nil
}

// GetImage returns the image to use for the container
func (c *Config) GetImage() string {
	if c.Image != "" {
		return c.Image
	}
	if c.Build != nil && c.Build.Dockerfile != "" {
		// Return empty to indicate build is needed
		return ""
	}
	return ""
}
