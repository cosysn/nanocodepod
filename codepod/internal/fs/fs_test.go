package fs

import (
	"testing"
)

func TestClientStructure(t *testing.T) {
	client := &Client{
		wslDistro: "Ubuntu-22.04",
		useServer: false,
	}

	if client.wslDistro != "Ubuntu-22.04" {
		t.Errorf("Expected 'Ubuntu-22.04', got %s", client.wslDistro)
	}
	if client.useServer != false {
		t.Errorf("Expected false, got %v", client.useServer)
	}
}

func TestClientWithServer(t *testing.T) {
	client := &Client{
		wslDistro: "Ubuntu-22.04",
		useServer: true,
	}

	if client.useServer != true {
		t.Errorf("Expected true, got %v", client.useServer)
	}
}

func TestClientWithServerClient(t *testing.T) {
	client := &Client{
		wslDistro:    "Ubuntu-22.04",
		useServer:    true,
		serverClient: nil,
	}

	if client.wslDistro != "Ubuntu-22.04" {
		t.Errorf("Expected 'Ubuntu-22.04', got %s", client.wslDistro)
	}
	if client.useServer != true {
		t.Errorf("Expected true, got %v", client.useServer)
	}
}

func TestNewClientWithoutServer(t *testing.T) {
	// When serverURL is empty, New should return client with useServer=false
	client, err := New("", "Ubuntu-22.04")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if client.useServer != false {
		t.Errorf("Expected useServer=false, got %v", client.useServer)
	}
	if client.wslDistro != "Ubuntu-22.04" {
		t.Errorf("Expected wslDistro='Ubuntu-22.04', got %s", client.wslDistro)
	}
}

func TestNewClientWithInvalidServer(t *testing.T) {
	// When serverURL is invalid, should fall back to useServer=false
	client, err := New("http://invalid:99999", "Ubuntu-22.04")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if client.useServer != false {
		t.Errorf("Expected useServer=false (fallback), got %v", client.useServer)
	}
}

func TestReadFileWithoutServer(t *testing.T) {
	client := &Client{
		wslDistro: "Ubuntu-22.04",
		useServer: false,
	}

	_, err := client.ReadFile("/test/path")
	if err == nil {
		t.Errorf("Expected error for WSL direct access, got nil")
	}
}

func TestWriteFileWithoutServer(t *testing.T) {
	client := &Client{
		wslDistro: "Ubuntu-22.04",
		useServer: false,
	}

	err := client.WriteFile("/test/path", "content")
	if err == nil {
		t.Errorf("Expected error for WSL direct access, got nil")
	}
}
