package server

import (
	"testing"
)

func TestNewClient(t *testing.T) {
	client := New("http://localhost:8080")

	if client.baseURL != "http://localhost:8080" {
		t.Errorf("Expected 'http://localhost:8080', got %s", client.baseURL)
	}
	if client.httpClient == nil {
		t.Errorf("Expected httpClient to be set")
	}
	if client.httpClient.Timeout.Seconds() != 30 {
		t.Errorf("Expected timeout 30s, got %v", client.httpClient.Timeout)
	}
}

func TestContainerStructure(t *testing.T) {
	container := Container{
		ID:     "abc123",
		Image:  "ubuntu:22.04",
		Status: "running",
		Names:  "test-container",
		Created: "2024-01-01",
	}

	if container.ID != "abc123" {
		t.Errorf("Expected 'abc123', got %s", container.ID)
	}
	if container.Image != "ubuntu:22.04" {
		t.Errorf("Expected 'ubuntu:22.04', got %s", container.Image)
	}
	if container.Status != "running" {
		t.Errorf("Expected 'running', got %s", container.Status)
	}
	if container.Names != "test-container" {
		t.Errorf("Expected 'test-container', got %s", container.Names)
	}
	if container.Created != "2024-01-01" {
		t.Errorf("Expected '2024-01-01', got %s", container.Created)
	}
}

func TestClientWithDifferentURLs(t *testing.T) {
	tests := []struct {
		url      string
		expected string
	}{
		{"http://localhost:8080", "http://localhost:8080"},
		{"http://192.168.1.1:9000", "http://192.168.1.1:9000"},
		{"http://127.0.0.1:80", "http://127.0.0.1:80"},
	}

	for _, tc := range tests {
		client := New(tc.url)
		if client.baseURL != tc.expected {
			t.Errorf("Expected %s, got %s", tc.expected, client.baseURL)
		}
	}
}

func TestWorkspaceStructure(t *testing.T) {
	ws := Workspace{
		Name:        "test-workspace",
		Status:      "running",
		Provider:    "local",
		Image:       "ubuntu:22.04",
		Port:        22001,
		ContainerID: "abc123",
		DataDir:     "/home/user/codepod/test-workspace",
	}

	if ws.Name != "test-workspace" {
		t.Errorf("Expected 'test-workspace', got %s", ws.Name)
	}
	if ws.Status != "running" {
		t.Errorf("Expected 'running', got %s", ws.Status)
	}
	if ws.Provider != "local" {
		t.Errorf("Expected 'local', got %s", ws.Provider)
	}
	if ws.Image != "ubuntu:22.04" {
		t.Errorf("Expected 'ubuntu:22.04', got %s", ws.Image)
	}
	if ws.Port != 22001 {
		t.Errorf("Expected 22001, got %d", ws.Port)
	}
	if ws.ContainerID != "abc123" {
		t.Errorf("Expected 'abc123', got %s", ws.ContainerID)
	}
	if ws.DataDir != "/home/user/codepod/test-workspace" {
		t.Errorf("Expected '/home/user/codepod/test-workspace', got %s", ws.DataDir)
	}
}

func TestHealthCheckWithInvalidURL(t *testing.T) {
	client := New("http://invalid:99999")
	err := client.HealthCheck()
	if err == nil {
		t.Errorf("Expected error for invalid URL, got nil")
	}
}
