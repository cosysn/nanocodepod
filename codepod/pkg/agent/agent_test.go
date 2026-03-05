package agent

import (
	"context"
	"testing"

	"github.com/codepod-io/codepod/pkg/router"
	"github.com/codepod-io/codepod/pkg/rpc"
)

func TestNewAgent(t *testing.T) {
	a := NewLocalAgent()

	if a.Type != router.AgentTypeLocal {
		t.Errorf("expected local agent, got %v", a.Type)
	}
	if a.Router == nil {
		t.Error("router should not be nil")
	}
	if a.Resolver == nil {
		t.Error("resolver should not be nil")
	}
	if a.RPCServer == nil {
		t.Error("rpc server should not be nil")
	}
}

func TestNewWorkspaceAgent(t *testing.T) {
	a := NewWorkspaceAgent()

	if a.Type != router.AgentTypeWorkspace {
		t.Errorf("expected workspace agent, got %v", a.Type)
	}
}

func TestNewContainerAgent(t *testing.T) {
	a := NewContainerAgent()

	if a.Type != router.AgentTypeContainer {
		t.Errorf("expected container agent, got %v", a.Type)
	}
}

func TestAgentRegisterProvider(t *testing.T) {
	a := NewLocalAgent()

	// Test provider registration
	mockProvider := &mockProvider{providerType: "test"}
	a.RegisterProvider(mockProvider)

	p, ok := a.GetProvider("test")
	if !ok {
		t.Error("provider should be registered")
	}
	if p.Type() != "test" {
		t.Errorf("expected test provider, got %v", p.Type())
	}
}

func TestAgentRouteResolver(t *testing.T) {
	a := NewLocalAgent()

	// Test resolving a WSL URI
	auth, err := a.Resolver.Resolve("wsl+ubuntu")
	if err != nil {
		t.Fatalf("resolve error: %v", err)
	}

	if auth.Provider != "wsl" {
		t.Errorf("expected provider wsl, got %s", auth.Provider)
	}
	if auth.Identity != "ubuntu" {
		t.Errorf("expected identity ubuntu, got %s", auth.Identity)
	}
}

func TestAgentRouteResolverDevContainer(t *testing.T) {
	a := NewLocalAgent()

	// Test resolving a dev-container URI with Hex-JSON
	// This would require a valid Hex-JSON string
	// For now, test with placeholder
	_, err := a.Resolver.Resolve("dev-container+invalid")
	if err == nil {
		t.Error("should return error for invalid hex")
	}
}

func TestAgentRouter(t *testing.T) {
	a := NewLocalAgent()

	// Test router has correct agent type
	if a.Router.AgentType() != router.AgentTypeLocal {
		t.Errorf("expected local, got %v", a.Router.AgentType())
	}

	// Test service registration
	a.Router.RegisterService("test", func(params any) (any, error) {
		return "test result", nil
	})

	handler, ok := a.Router.GetService("test")
	if !ok {
		t.Error("service should be registered")
	}

	result, err := handler(nil)
	if err != nil {
		t.Errorf("handler error: %v", err)
	}
	if result != "test result" {
		t.Errorf("expected test result, got %v", result)
	}
}

// mockProvider implements Provider interface
type mockProvider struct {
	providerType string
}

func (m *mockProvider) Type() string {
	return m.providerType
}

func (m *mockProvider) Bootstrap(ctx context.Context, identity string) (*rpc.RPCClient, error) {
	return nil, nil
}

func (m *mockProvider) Connect(ctx context.Context, identity string) (*rpc.RPCClient, error) {
	return nil, nil
}
