package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client represents a client for the WSL server
type Client struct {
	baseURL  string
	httpClient *http.Client
}

// Container represents a Docker container
type Container struct {
	ID      string `json:"id"`
	Image   string `json:"image"`
	Status  string `json:"status"`
	Names   string `json:"names"`
	Created string `json:"created"`
}

// New creates a new server client
func New(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// HealthCheck checks if the server is running
func (c *Client) HealthCheck() error {
	resp, err := c.httpClient.Get(c.baseURL + "/health")
	if err != nil {
		return fmt.Errorf("server not reachable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status %d", resp.StatusCode)
	}
	return nil
}

// ListContainers lists containers via the server
func (c *Client) ListContainers() ([]Container, error) {
	resp, err := c.httpClient.Get(c.baseURL + "/docker/ps")
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("server error: %s", string(body))
	}

	var containers []Container
	if err := json.NewDecoder(resp.Body).Decode(&containers); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return containers, nil
}

// RunContainer runs a container via the server
func (c *Client) RunContainer(image string) (string, error) {
	resp, err := c.httpClient.Post(c.baseURL+"/docker/run", "text/plain", bytes.NewBufferString(image))
	if err != nil {
		return "", fmt.Errorf("failed to run container: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("server error: %s", string(body))
	}

	output, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	return string(output), nil
}

// PullImage pulls an image via the server
func (c *Client) PullImage(image string) error {
	resp, err := c.httpClient.Post(c.baseURL+"/docker/pull", "text/plain", bytes.NewBufferString(image))
	if err != nil {
		return fmt.Errorf("failed to pull image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server error: %s", string(body))
	}

	return nil
}

// ReadFile reads a file via the server
func (c *Client) ReadFile(path string) (string, error) {
	resp, err := c.httpClient.Get(c.baseURL + "/fs/read?path=" + path)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("server error: %s", string(body))
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	return string(content), nil
}

// WriteFile writes a file via the server
func (c *Client) WriteFile(path string, content string) error {
	resp, err := c.httpClient.Post(c.baseURL+"/fs/write?path="+path, "text/plain", bytes.NewBufferString(content))
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server error: %s", string(body))
	}

	return nil
}
