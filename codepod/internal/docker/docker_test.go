package docker

import (
	"testing"
)

func TestNew(t *testing.T) {
	client, err := New("unix:///var/run/docker.sock")
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	if client == nil {
		t.Error("client should not be nil")
	}
}

func TestClient_Close(t *testing.T) {
	client, _ := New("unix:///var/run/docker.sock")
	err := client.Close()
	if err != nil {
		t.Fatalf("failed to close: %v", err)
	}
}

func TestContainerConfig(t *testing.T) {
	config := &ContainerConfig{
		Name:    "test-container",
		Image:   "ubuntu:22.04",
		Cmd:     []string{"sleep", "infinity"},
		Env:     []string{"VAR=value"},
		Labels:  map[string]string{"app": "test"},
		PortBindings: map[string][]PortBinding{
			"22/tcp": {{HostIP: "0.0.0.0", HostPort: "22000"}},
		},
		Binds:       []string{"/tmp:/workspace"},
		NetworkMode: "bridge",
		Privileged:  false,
	}

	if config.Name != "test-container" {
		t.Errorf("expected name test-container, got %s", config.Name)
	}
	if config.Image != "ubuntu:22.04" {
		t.Errorf("expected image ubuntu:22.04, got %s", config.Image)
	}
	if len(config.PortBindings) != 1 {
		t.Errorf("expected 1 port binding, got %d", len(config.PortBindings))
	}
}

func TestPortBinding(t *testing.T) {
	binding := PortBinding{
		HostIP:   "0.0.0.0",
		HostPort: "22000",
	}

	if binding.HostIP != "0.0.0.0" {
		t.Errorf("expected HostIP 0.0.0.0, got %s", binding.HostIP)
	}
	if binding.HostPort != "22000" {
		t.Errorf("expected HostPort 22000, got %s", binding.HostPort)
	}
}
