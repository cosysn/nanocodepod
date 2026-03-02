package docker

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// DockerClient defines the interface for Docker operations
type DockerClient interface {
	CreateContainer(config *ContainerConfig) (string, error)
	StartContainer(name string) error
	StopContainer(name string) error
	RemoveContainer(name string, force bool) error
	InspectContainer(name string) (*ContainerInfo, error)
	ContainerExists(name string) bool
	ExecInContainer(name string, cmd []string) error
	ExecInContainerDetached(name string, cmd []string) error
	CopyToContainer(containerID, src, dest string) error
	Close() error
}

// Client represents a Docker client using CLI
type Client struct {
	daemonAddr string
}

// Ensure Client implements DockerClient
var _ DockerClient = (*Client)(nil)

// New creates a new Docker client
func New(daemonAddr string) (*Client, error) {
	return &Client{
		daemonAddr: daemonAddr,
	}, nil
}

// Close closes the Docker client (no-op for CLI-based client)
func (c *Client) Close() error {
	return nil
}

// ListContainers lists all containers
func (c *Client) ListContainers() ([]Container, error) {
	cmd := exec.Command("docker", "ps", "-a", "--format", "{{.ID}}|{{.Names}}|{{.Status}}|{{.Image}}")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	var containers []Container
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, "|")
		if len(parts) >= 4 {
			containers = append(containers, Container{
				ID:     parts[0],
				Names:  parts[1],
				Status: parts[2],
				Image:  parts[3],
			})
		}
	}
	return containers, nil
}

// Container represents a Docker container
type Container struct {
	ID     string
	Names  string
	Status string
	Image  string
}

// PullImage pulls a Docker image
func (c *Client) PullImage(image string) error {
	cmd := exec.Command("docker", "pull", image)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// CreateContainer creates a new container
func (c *Client) CreateContainer(config *ContainerConfig) (string, error) {
	args := []string{"run", "-d", "--name", config.Name}

	// Add environment variables
	for _, env := range config.Env {
		args = append(args, "-e", env)
	}

	// Add port mappings
	for containerPort, hostBindings := range config.PortBindings {
		for _, binding := range hostBindings {
			// Format: hostPort:containerPort or hostIP:hostPort:containerPort
			if binding.HostIP != "" && binding.HostIP != "0.0.0.0" {
				args = append(args, "-p", fmt.Sprintf("%s:%s:%s", binding.HostIP, binding.HostPort, containerPort))
			} else {
				args = append(args, "-p", fmt.Sprintf("%s:%s", binding.HostPort, containerPort))
			}
		}
	}

	// Add volume bindings
	for _, bind := range config.Binds {
		args = append(args, "-v", bind)
	}

	// Add network
	if config.NetworkMode != "" {
		args = append(args, "--network", config.NetworkMode)
	}

	// Add privileged
	if config.Privileged {
		args = append(args, "--privileged")
	}

	// Add labels
	for k, v := range config.Labels {
		args = append(args, "--label", fmt.Sprintf("%s=%s", k, v))
	}

	// Add image
	args = append(args, config.Image)

	// Add command
	if len(config.Cmd) > 0 {
		args = append(args, config.Cmd...)
	}

	cmd := exec.Command("docker", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to create container: %w, output: %s", err, string(out))
	}

	return strings.TrimSpace(string(out)), nil
}

// StartContainer starts a container
func (c *Client) StartContainer(name string) error {
	cmd := exec.Command("docker", "start", name)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to start container: %w", err)
	}
	return nil
}

// StopContainer stops a container
func (c *Client) StopContainer(name string) error {
	cmd := exec.Command("docker", "stop", name)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to stop container: %w", err)
	}
	return nil
}

// RemoveContainer removes a container
func (c *Client) RemoveContainer(name string, force bool) error {
	args := []string{"rm"}
	if force {
		args = append(args, "-f")
	}
	args = append(args, name)

	cmd := exec.Command("docker", args...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to remove container: %w", err)
	}
	return nil
}

// InspectContainer returns container info
func (c *Client) InspectContainer(name string) (*ContainerInfo, error) {
	cmd := exec.Command("docker", "inspect", name)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to inspect container: %w", err)
	}

	// Parse basic info
	info := &ContainerInfo{}
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		if strings.Contains(line, "\"Running\":") {
			info.Running = strings.Contains(line, "true")
		}
		if strings.Contains(line, "\"IPAddress\"") {
			parts := strings.Split(line, "\"")
			for i, p := range parts {
				if p == "IPAddress" && i+2 < len(parts) {
					info.IPAddress = parts[i+2]
					break
				}
			}
		}
	}

	return info, nil
}

// ContainerInfo holds container inspection data
type ContainerInfo struct {
	Running   bool
	IPAddress string
}

// GetContainerIP returns the container's IP address
func (c *Client) GetContainerIP(name string) (string, error) {
	cmd := exec.Command("docker", "inspect", "-f", "{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}", name)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// ExecInContainer executes a command in a container
func (c *Client) ExecInContainer(name string, cmdArgs []string) error {
	args := []string{"exec", name}
	args = append(args, cmdArgs...)

	cmd := exec.Command("docker", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}

// ExecInContainerDetached executes a command in a container in detached mode
func (c *Client) ExecInContainerDetached(name string, cmdArgs []string) error {
	args := []string{"exec", "-d", name}
	args = append(args, cmdArgs...)

	cmd := exec.Command("docker", args...)
	return cmd.Run()
}

// CopyToContainer copies a file to a container
func (c *Client) CopyToContainer(containerID, src, dest string) error {
	cmd := exec.Command("docker", "cp", src, fmt.Sprintf("%s:%s", containerID, dest))
	return cmd.Run()
}

// ContainerExists checks if a container exists
func (c *Client) ContainerExists(name string) bool {
	containers, err := c.ListContainers()
	if err != nil {
		return false
	}
	for _, c := range containers {
		if c.Names == name {
			return true
		}
	}
	return false
}

// GetContainerByName returns container by name
func (c *Client) GetContainerByName(name string) (*Container, error) {
	containers, err := c.ListContainers()
	if err != nil {
		return nil, err
	}
	for _, c := range containers {
		if c.Names == name {
			return &c, nil
		}
	}
	return nil, fmt.Errorf("container %s not found", name)
}

// ContainerConfig holds container configuration
type ContainerConfig struct {
	Name          string
	Image         string
	Cmd           []string
	Env           []string
	Labels        map[string]string
	PortBindings  map[string][]PortBinding
	Binds         []string
	NetworkMode   string
	Privileged    bool
}

type PortBinding struct {
	HostIP   string
	HostPort string
}
