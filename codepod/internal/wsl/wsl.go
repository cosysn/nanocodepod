package wsl

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// WSL represents a WSL2 instance manager
type WSL struct {
	distribution string
}

// New creates a new WSL manager
func New(distribution string) *WSL {
	return &WSL{
		distribution: distribution,
	}
}

// RunCommand executes a command in WSL and returns the output
func (w *WSL) RunCommand(cmd string) (string, error) {
	args := []string{"-d", w.distribution, "--", "bash", "-c", cmd}
	out, err := exec.Command("wsl.exe", args...).Output()
	if err != nil {
		return "", fmt.Errorf("failed to run command in WSL: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// RunCommandWithStdin executes a command in WSL with stdin input
func (w *WSL) RunCommandWithStdin(cmd string, stdin string) (string, error) {
	args := []string{"-d", w.distribution, "--", "bash", "-c", cmd}
	proc := exec.Command("wsl.exe", args...)
	proc.Stdin = strings.NewReader(stdin)
	out, err := proc.Output()
	if err != nil {
		return "", fmt.Errorf("failed to run command in WSL: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// ListDistributions lists all WSL distributions
func ListDistributions() ([]string, error) {
	out, err := exec.Command("wsl.exe", "-l", "-q").Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list WSL distributions: %w", err)
	}

	var distros []string
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			distros = append(distros, line)
		}
	}
	return distros, nil
}

// DistributionExists checks if a WSL distribution exists
func DistributionExists(name string) (bool, error) {
	distros, err := ListDistributions()
	if err != nil {
		return false, err
	}
	for _, d := range distros {
		if strings.Contains(d, name) {
			return true, nil
		}
	}
	return false, nil
}

// CopyToWSL copies a file to WSL
func (w *WSL) CopyToWSL(localPath, wslPath string) error {
	args := []string{"-d", w.distribution, "--", "cp", localPath, wslPath}
	cmd := exec.Command("wsl.exe", args...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to copy to WSL: %w", err)
	}
	return nil
}

// CopyFromWSL copies a file from WSL
func (w *WSL) CopyFromWSL(wslPath, localPath string) error {
	args := []string{"-d", w.distribution, "--", "cp", wslPath, localPath}
	cmd := exec.Command("wsl.exe", args...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to copy from WSL: %w", err)
	}
	return nil
}

// GetWSLIP returns the IP address of the WSL instance
func (w *WSL) GetWSLIP() (string, error) {
	cmd := "ip addr show eth0 | grep 'inet ' | awk '{print $2}' | cut -d'/' -f1"
	output, err := w.RunCommand(cmd)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(output), nil
}

// Exec executes a command in WSL without capturing output
func (w *WSL) Exec(args ...string) error {
	fullArgs := []string{"-d", w.distribution, "--"}
	fullArgs = append(fullArgs, args...)
	cmd := exec.Command("wsl.exe", fullArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// IsDockerRunning checks if Docker is running in WSL
func (w *WSL) IsDockerRunning() bool {
	cmd := "docker info"
	_, err := w.RunCommand(cmd)
	return err == nil
}

// GetDockerVersion returns the Docker version in WSL
func (w *WSL) GetDockerVersion() (string, error) {
	cmd := "docker version --format '{{.Server.Version}}'"
	return w.RunCommand(cmd)
}

// EnsureDocker starts Docker daemon in WSL if not running
func (w *WSL) EnsureDocker() error {
	if w.IsDockerRunning() {
		return nil
	}
	// Try to start docker service
	_, err := w.RunCommand("sudo service docker start")
	if err != nil {
		return fmt.Errorf("failed to start Docker: %w", err)
	}
	return nil
}

// FileExists checks if a file exists in WSL
func (w *WSL) FileExists(path string) bool {
	cmd := fmt.Sprintf("test -e %s && echo 'exists' || echo 'not exists'", path)
	out, err := w.RunCommand(cmd)
	return err == nil && strings.Contains(out, "exists")
}

// CreateDir creates a directory in WSL
func (w *WSL) CreateDir(path string) error {
	cmd := fmt.Sprintf("mkdir -p %s", path)
	_, err := w.RunCommand(cmd)
	return err
}

// RemoveFile removes a file in WSL
func (w *WSL) RemoveFile(path string) error {
	cmd := fmt.Sprintf("rm -rf %s", path)
	_, err := w.RunCommand(cmd)
	return err
}

// WriteFile writes content to a file in WSL
func (w *WSL) WriteFile(path, content string) error {
	cmd := fmt.Sprintf("cat > %s << 'EOF'\n%s\nEOF", path, content)
	_, err := w.RunCommand(cmd)
	return err
}

// ReadFile reads content from a file in WSL
func (w *WSL) ReadFile(path string) (string, error) {
	cmd := fmt.Sprintf("cat %s", path)
	return w.RunCommand(cmd)
}
