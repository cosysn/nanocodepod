package router

import (
	"testing"
	"time"
)

func TestNewRouter(t *testing.T) {
	r := NewRouter(AgentTypeLocal, "local-1")

	if r.AgentType() != AgentTypeLocal {
		t.Errorf("NewRouter() got agentType = %v, want local", r.AgentType())
	}
	if r.AgentID() != "local-1" {
		t.Errorf("NewRouter() got agentID = %v, want local-1", r.AgentID())
	}
}

func TestSetParent(t *testing.T) {
	parent := &RouteNode{
		AgentType: AgentTypeLocal,
		AgentID:   "parent-1",
	}

	child := NewRouter(AgentTypeWorkspace, "workspace-1")
	child.SetParent(parent)

	if child.GetParent() != parent {
		t.Errorf("SetParent() parent not set correctly")
	}
}

func TestAddChild(t *testing.T) {
	parent := NewRouter(AgentTypeLocal, "local-1")

	child := &RouteNode{
		AgentType:  AgentTypeWorkspace,
		AgentID:    "workspace-1",
		Connection: "mock-conn",
	}

	parent.AddChild("wsl+ubuntu", child)

	found := parent.FindChild("wsl+ubuntu")
	if found == nil {
		t.Errorf("AddChild() child not found")
	}
	if found.AgentID != "workspace-1" {
		t.Errorf("AddChild() got agentID = %v, want workspace-1", found.AgentID)
	}
}

func TestRemoveChild(t *testing.T) {
	r := NewRouter(AgentTypeLocal, "local-1")

	child := &RouteNode{
		AgentType: AgentTypeWorkspace,
		AgentID:   "workspace-1",
	}

	r.AddChild("wsl+ubuntu", child)
	r.RemoveChild("wsl+ubuntu")

	found := r.FindChild("wsl+ubuntu")
	if found != nil {
		t.Errorf("RemoveChild() child still exists")
	}
}

func TestFindChild(t *testing.T) {
	r := NewRouter(AgentTypeLocal, "local-1")

	// Add multiple children
	r.AddChild("wsl+ubuntu", &RouteNode{AgentType: AgentTypeWorkspace, AgentID: "ws-1"})
	r.AddChild("docker+container1", &RouteNode{AgentType: AgentTypeWorkspace, AgentID: "ws-2"})

	tests := []struct {
		name      string
		authority string
		wantID    string
		wantFound bool
	}{
		{"existing", "wsl+ubuntu", "ws-1", true},
		{"existing2", "docker+container1", "ws-2", true},
		{"not found", "ssh+host", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			found := r.FindChild(tt.authority)
			if (found != nil) != tt.wantFound {
				t.Errorf("FindChild() got found = %v, want %v", found != nil, tt.wantFound)
			}
			if tt.wantFound && found.AgentID != tt.wantID {
				t.Errorf("FindChild() got agentID = %v, want %v", found.AgentID, tt.wantID)
			}
		})
	}
}

func TestRegisterService(t *testing.T) {
	r := NewRouter(AgentTypeLocal, "local-1")

	handler := func(params any) (any, error) {
		return "result", nil
	}

	r.RegisterService("pty", handler)

	found, ok := r.GetService("pty")
	if !ok {
		t.Errorf("RegisterService() service not found")
	}

	result, err := found(nil)
	if err != nil {
		t.Errorf("RegisterService() handler error = %v", err)
	}
	if result != "result" {
		t.Errorf("RegisterService() got result = %v, want result", result)
	}
}

func TestConcurrentAccess(t *testing.T) {
	r := NewRouter(AgentTypeLocal, "local-1")

	done := make(chan bool)

	// Concurrent writes
	go func() {
		for i := 0; i < 100; i++ {
			r.AddChild("key"+string(rune(i)), &RouteNode{AgentID: "child"})
		}
		done <- true
	}()

	// Concurrent reads
	go func() {
		for i := 0; i < 100; i++ {
			r.FindChild("key0")
			r.GetParent()
			r.GetService("test")
		}
		done <- true
	}()

	// Concurrent writes
	go func() {
		for i := 0; i < 100; i++ {
			r.RemoveChild("key0")
		}
		done <- true
	}()

	timeout := time.After(2 * time.Second)
	for i := 0; i < 3; i++ {
		select {
		case <-done:
		case <-timeout:
			t.Errorf("Concurrent access timeout - possible race condition")
			return
		}
	}
}

func TestValidateAgentType(t *testing.T) {
	tests := []struct {
		name    string
		t       string
		wantErr bool
	}{
		{"valid local", "local", false},
		{"valid workspace", "workspace", false},
		{"valid container", "container", false},
		{"invalid", "invalid", true},
		{"empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAgentType(tt.t)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAgentType() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
