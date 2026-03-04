package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Client wraps HTTP calls to the server
type Client struct {
	BaseURL string
}

// NewClient creates a new server client
func NewClient(baseURL string) *Client {
	return &Client{BaseURL: baseURL}
}

// Workspace represents a workspace
type Workspace struct {
	Name   string `json:"name"`
	State  string `json:"state"`
	Port   int    `json:"port"`
	Image  string `json:"image"`
	RepoURL string `json:"repo_url,omitempty"`
	Branch string `json:"branch,omitempty"`
}

// CreateWorkspaceRequest is the request for creating a workspace
type CreateWorkspaceRequest struct {
	Name    string `json:"name"`
	Image   string `json:"image"`
	RepoURL string `json:"repo_url,omitempty"`
	Branch  string `json:"branch,omitempty"`
}

// ListWorkspaces returns all workspaces
func (c *Client) ListWorkspaces() ([]Workspace, error) {
	resp, err := http.Get(c.BaseURL + "/workspace/list")
	if err != nil {
		return nil, fmt.Errorf("failed to list workspaces: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to list workspaces: %s", string(body))
	}

	var workspaces []Workspace
	if err := json.NewDecoder(resp.Body).Decode(&workspaces); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return workspaces, nil
}

// GetWorkspace returns a workspace by name
func (c *Client) GetWorkspace(name string) (*Workspace, error) {
	resp, err := http.Get(c.BaseURL + "/workspace/get?name=" + name)
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get workspace: %s", string(body))
	}

	var workspace Workspace
	if err := json.NewDecoder(resp.Body).Decode(&workspace); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &workspace, nil
}

// CreateWorkspace creates a new workspace
func (c *Client) CreateWorkspace(req CreateWorkspaceRequest) (*Workspace, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := http.Post(c.BaseURL+"/workspace/create", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create workspace: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to create workspace: %s", string(respBody))
	}

	var workspace Workspace
	if err := json.NewDecoder(resp.Body).Decode(&workspace); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &workspace, nil
}

// StartWorkspace starts a workspace
func (c *Client) StartWorkspace(name string) error {
	resp, err := http.Get(c.BaseURL + "/workspace/start?name=" + name)
	if err != nil {
		return fmt.Errorf("failed to start workspace: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to start workspace: %s", string(body))
	}

	return nil
}

// StopWorkspace stops a workspace
func (c *Client) StopWorkspace(name string) error {
	resp, err := http.Get(c.BaseURL + "/workspace/stop?name=" + name)
	if err != nil {
		return fmt.Errorf("failed to stop workspace: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to stop workspace: %s", string(body))
	}

	return nil
}

// DeleteWorkspace deletes a workspace
func (c *Client) DeleteWorkspace(name string) error {
	resp, err := http.Get(c.BaseURL + "/workspace/delete?name=" + name)
	if err != nil {
		return fmt.Errorf("failed to delete workspace: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete workspace: %s", string(body))
	}

	return nil
}
