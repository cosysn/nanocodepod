package docker

import (
	"testing"
)

// TestContainerStruct tests Container struct
func TestContainerStruct(t *testing.T) {
	container := Container{
		ID:     "abc123",
		Image:  "ubuntu:22.04",
		Status: "running",
		Names:  "test-container",
		Created: "2024-01-01",
	}

	if container.ID != "abc123" {
		t.Errorf("Expected ID 'abc123', got %s", container.ID)
	}
	if container.Image != "ubuntu:22.04" {
		t.Errorf("Expected Image 'ubuntu:22.04', got %s", container.Image)
	}
	if container.Status != "running" {
		t.Errorf("Expected Status 'running', got %s", container.Status)
	}
	if container.Names != "test-container" {
		t.Errorf("Expected Names 'test-container', got %s", container.Names)
	}
	if container.Created != "2024-01-01" {
		t.Errorf("Expected Created '2024-01-01', got %s", container.Created)
	}
}
