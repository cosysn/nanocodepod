package docker

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os/exec"
	"strings"
)

// Container represents a Docker container
type Container struct {
	ID      string `json:"id"`
	Image   string `json:"image"`
	Status  string `json:"status"`
	Names   string `json:"names"`
	Created string `json:"created"`
}

// DockerClient defines the interface for Docker operations
type DockerClient interface {
	CreateContainer(config *ContainerConfig) (string, error)
	StartContainer(name string) error
	StopContainer(name string) error
	RemoveContainer(name string, force bool) error
	ListContainers() ([]Container, error)
	PullImage(image string) error
}

// ContainerConfig holds container configuration
type ContainerConfig struct {
	Name         string
	Image        string
	Cmd          []string
	Env          []string
	Labels       map[string]string
	PortBindings map[string][]PortBinding
	Binds        []string
	NetworkMode  string
}

type PortBinding struct {
	HostIP   string
	HostPort string
}

// Client is the Docker CLI client
type Client struct{}

// New creates a new Docker client
func New() *Client {
	return &Client{}
}

// CreateContainer creates a new container
func (c *Client) CreateContainer(config *ContainerConfig) (string, error) {
	args := []string{"run", "-d", "--name", config.Name}

	for _, env := range config.Env {
		args = append(args, "-e", env)
	}

	for containerPort, hostBindings := range config.PortBindings {
		for _, binding := range hostBindings {
			args = append(args, "-p", fmt.Sprintf("%s:%s", binding.HostPort, containerPort))
		}
	}

	for _, bind := range config.Binds {
		args = append(args, "-v", bind)
	}

	if config.NetworkMode != "" {
		args = append(args, "--network", config.NetworkMode)
	}

	for k, v := range config.Labels {
		args = append(args, "--label", fmt.Sprintf("%s=%s", k, v))
	}

	args = append(args, config.Image)

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

// PullImage pulls a Docker image
func (c *Client) PullImage(image string) error {
	cmd := exec.Command("docker", "pull", image)
	cmd.Stdout = log.Writer()
	cmd.Stderr = log.Writer()
	return cmd.Run()
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

// ListContainers handles GET /docker/ps
func ListContainers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cmd := exec.Command("docker", "ps", "--format", "{{.ID}}|{{.Image}}|{{.Status}}|{{.Names}}|{{.Created}}")
	output, err := cmd.Output()
	if err != nil {
		log.Printf("docker ps failed: %v", err)
		http.Error(w, fmt.Sprintf("docker ps failed: %v", err), http.StatusInternalServerError)
		return
	}

	var containers []Container
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Split(line, "|")
		if len(parts) >= 5 {
			containers = append(containers, Container{
				ID:      parts[0],
				Image:   parts[1],
				Status:  parts[2],
				Names:   parts[3],
				Created: parts[4],
			})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(containers)
}

// RunContainer handles POST /docker/run
func RunContainer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	// Parse docker run command from body (simplified - in production use proper JSON)
	image := string(body)
	if image == "" {
		image = "ubuntu:22.04"
	}

	cmd := exec.Command("docker", "run", "-d", image)
	output, err := cmd.Output()
	if err != nil {
		log.Printf("docker run failed: %v", err)
		http.Error(w, fmt.Sprintf("docker run failed: %v", err), http.StatusInternalServerError)
		return
	}

	w.Write(output)
}

// PullImage handles POST /docker/pull
func PullImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	image := strings.TrimSpace(string(body))
	if image == "" {
		http.Error(w, "Image name required", http.StatusBadRequest)
		return
	}

	cmd := exec.Command("docker", "pull", image)
	cmd.Stdout = w
	cmd.Stderr = w

	if err := cmd.Run(); err != nil {
		log.Printf("docker pull failed: %v", err)
		// Already wrote output, just log
	}
}
