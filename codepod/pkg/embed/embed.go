// Package embed provides embedded agent binaries for cross-platform deployment.
package embed

import (
	"archive/zip"
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
)

//go:embed agents.zip
var agentsZip []byte

// Platform represents a target platform/architecture.
type Platform struct {
	OS   string // linux, darwin, windows
	Arch string // amd64, arm64
}

var (
	// SupportedPlatforms lists all available embedded platforms.
	SupportedPlatforms = []Platform{
		{"linux", "amd64"},
		{"linux", "arm64"},
		{"darwin", "amd64"},
		{"darwin", "arm64"},
		{"windows", "amd64"},
	}

	// platformToFile maps platform to embedded filename.
	platformToFile = map[Platform]string{
		{"linux", "amd64"}:   "agent-linux-amd64",
		{"linux", "arm64"}:   "agent-linux-arm64",
		{"darwin", "amd64"}:  "agent-darwin-amd64",
		{"darwin", "arm64"}:  "agent-darwin-arm64",
		{"windows", "amd64"}: "agent-windows-amd64.exe",
	}
)

// GetAgentBinary returns the embedded agent binary for the given platform.
// Returns nil if not embedded.
func GetAgentBinary(platform Platform) ([]byte, error) {
	if len(agentsZip) == 0 {
		return nil, fmt.Errorf("no embedded agents available")
	}

	filename := platformToFile[platform]
	if filename == "" {
		return nil, fmt.Errorf("unsupported platform: %s/%s", platform.OS, platform.Arch)
	}

	zr, err := zip.NewReader(bytes.NewReader(agentsZip), int64(len(agentsZip)))
	if err != nil {
		return nil, fmt.Errorf("failed to read embedded agents: %w", err)
	}

	for _, f := range zr.File {
		if f.Name == filename {
			rc, err := f.Open()
			if err != nil {
				return nil, fmt.Errorf("failed to open %s: %w", filename, err)
			}
			defer rc.Close()

			return io.ReadAll(rc)
		}
	}

	return nil, fmt.Errorf("agent not found for platform: %s/%s", platform.OS, platform.Arch)
}

// GetCurrentPlatform returns the platform for the current runtime.
func GetCurrentPlatform() Platform {
	os := runtime.GOOS
	if os == "windows" {
		os = "windows"
	}
	return Platform{
		OS:   os,
		Arch: runtime.GOARCH,
	}
}

// GetTargetPlatform returns the target platform for embedding based on build tags.
// This is used during build time to determine which platform to embed.
func GetTargetPlatform() []Platform {
	return SupportedPlatforms
}

// ExtractAgent extracts the embedded agent binary to a temporary file
// and returns the path. Caller should delete the file when done.
func ExtractAgent(platform Platform) (string, error) {
	data, err := GetAgentBinary(platform)
	if err != nil {
		return "", err
	}

	// Get temp directory
	tmpDir := os.TempDir()
	filename := platformToFile[platform]
	if filename == "" {
		filename = "codepod-agent"
	}

	// Add .exe on windows
	if platform.OS == "windows" && filepath.Ext(filename) == "" {
		filename += ".exe"
	}

	tmpPath := filepath.Join(tmpDir, filename)

	if err := os.WriteFile(tmpPath, data, 0755); err != nil {
		return "", fmt.Errorf("failed to write agent binary: %w", err)
	}

	return tmpPath, nil
}

// HasEmbeddedAgents returns true if agents are embedded.
func HasEmbeddedAgents() bool {
	return len(agentsZip) > 0
}
