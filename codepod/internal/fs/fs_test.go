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
