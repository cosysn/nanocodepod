package devcon

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/codepod-io/codepod/internal/wsl"
)

// Devcon handles devcontainer.json integration
type Devcon struct {
	devconPath string
	wsl        *wsl.WSL
}

// New creates a new devcon handler
func New(wslInstance *wsl.WSL) (*Devcon, error) {
	devconPath, err := findDevcon()
	if err != nil {
		return nil, err
	}

	return &Devcon{
		devconPath: devconPath,
		wsl:        wslInstance,
	}, nil
}

// BuildOptions holds options for building a devcontainer
type BuildOptions struct {
	WorkspacePath string
	ImageTag      string
	RepositoryURL string
	Dockerfile    string // Custom Dockerfile path
	Context       string // Build context path
}

// Build builds a devcontainer image from .devcontainer.json
func (d *Devcon) Build(opts *BuildOptions) (string, error) {
	// Check if devcontainer.json exists
	devcontainerPath := filepath.Join(opts.WorkspacePath, ".devcontainer.json")
	if _, err := os.Stat(devcontainerPath); os.IsNotExist(err) {
		return "", fmt.Errorf("no .devcontainer.json found in %s", opts.WorkspacePath)
	}

	// Determine build command based on environment
	buildCmd, err := d.getBuildCommand(opts)
	if err != nil {
		return "", err
	}

	output, err := exec.Command(buildCmd[0], buildCmd[1:]...).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("devcontainer build failed: %w\noutput: %s", err, string(output))
	}

	// Extract image name from output
	imageName := d.extractImageName(string(output), opts.ImageTag)
	return imageName, nil
}

// Up starts a devcontainer
func (d *Devcon) Up(opts *BuildOptions) error {
	buildCmd, err := d.getUpCommand(opts)
	if err != nil {
		return err
	}

	cmd := exec.Command(buildCmd[0], buildCmd[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}

// getBuildCommand returns the appropriate build command
func (d *Devcon) getBuildCommand(opts *BuildOptions) ([]string, error) {
	if d.wsl != nil {
		// WSL: copy devcon to WSL and run there
		wslDevconPath := "/tmp/devcon"
		if err := d.wsl.CopyToWSL(d.devconPath, wslDevconPath); err != nil {
			return nil, err
		}

		return []string{
			"wsl.exe", "-d", d.wsl.GetDistribution(), "--",
			"bash", "-c",
			fmt.Sprintf("%s build %s --image-tag %s", wslDevconPath, opts.WorkspacePath, opts.ImageTag),
		}, nil
	}

	// Linux: run devcon directly
	return []string{d.devconPath, "build", opts.WorkspacePath, "--image-tag", opts.ImageTag}, nil
}

// getUpCommand returns the appropriate up command
func (d *Devcon) getUpCommand(opts *BuildOptions) ([]string, error) {
	if d.wsl != nil {
		wslDevconPath := "/tmp/devcon"
		if err := d.wsl.CopyToWSL(d.devconPath, wslDevconPath); err != nil {
			return nil, err
		}

		return []string{
			"wsl.exe", "-d", d.wsl.GetDistribution(), "--",
			"bash", "-c",
			fmt.Sprintf("%s up %s", wslDevconPath, opts.WorkspacePath),
		}, nil
	}

	return []string{d.devconPath, "up", opts.WorkspacePath}, nil
}

// findDevcon finds the devcon binary
func findDevcon() (string, error) {
	// Check common locations
	paths := []string{
		"/home/ubuntu/devcon/devcon",
		"/usr/local/bin/devcon",
		"/usr/bin/devcon",
		filepath.Join(os.Getenv("HOME"), "devcon", "devcon"),
	}

	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	// Check if devcon is in PATH
	path, err := exec.LookPath("devcon")
	if err == nil {
		return path, nil
	}

	return "", fmt.Errorf("devcon not found in common paths")
}

// extractImageName extracts the image name from devcon output
func (d *Devcon) extractImageName(output, fallback string) string {
	// Look for "Image built: <name>" pattern
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "Image built:") {
			parts := strings.Split(line, "Image built:")
			if len(parts) > 1 {
				return strings.TrimSpace(parts[1])
			}
		}
	}

	if fallback != "" {
		return fallback
	}

	return "devcontainer:latest"
}

// HasDevcontainer checks if a workspace has a .devcontainer.json
func (d *Devcon) HasDevcontainer(workspacePath string) bool {
	devcontainerPath := filepath.Join(workspacePath, ".devcontainer.json")
	_, err := os.Stat(devcontainerPath)
	return err == nil
}

// GetDevcontainerConfig reads the .devcontainer.json config
func (d *Devcon) GetDevcontainerConfig(workspacePath string) (string, error) {
	devcontainerPath := filepath.Join(workspacePath, ".devcontainer.json")
	data, err := os.ReadFile(devcontainerPath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// BuildWithDockerfile builds a container using custom Dockerfile
func (d *Devcon) BuildWithDockerfile(opts *BuildOptions) (string, error) {
	if opts.Dockerfile == "" {
		return "", fmt.Errorf("dockerfile path is required")
	}

	// Verify Dockerfile exists
	if _, err := os.Stat(opts.Dockerfile); os.IsNotExist(err) {
		return "", fmt.Errorf("dockerfile not found: %s", opts.Dockerfile)
	}

	// Determine build context
	context := opts.WorkspacePath
	if opts.Context != "" {
		context = opts.Context
	}

	// Build using docker build
	cmd := exec.Command("docker", "build",
		"-t", opts.ImageTag,
		"-f", opts.Dockerfile,
		context,
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("docker build failed: %w", err)
	}

	return opts.ImageTag, nil
}

// HasDockerfile checks if a custom Dockerfile exists
func (d *Devcon) HasDockerfile(workspacePath string) bool {
	dockerfilePath := filepath.Join(workspacePath, "Dockerfile")
	_, err := os.Stat(dockerfilePath)
	return err == nil
}

// GetDockerfilePath returns the path to Dockerfile
func (d *Devcon) GetDockerfilePath(workspacePath string) string {
	return filepath.Join(workspacePath, "Dockerfile")
}

// HasDevfile checks if a devfile exists
func (d *Devcon) HasDevfile(workspacePath string) bool {
	devfilePath := filepath.Join(workspacePath, "devfile.yaml")
	_, err := os.Stat(devfilePath)
	return err == nil
}
