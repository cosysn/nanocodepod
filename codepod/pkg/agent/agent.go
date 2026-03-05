// Package agent provides the unified agent core for CodePod.
// It integrates resolver, router, channel, and RPC to implement
// the recursive routing "peeling onion" mechanism.
package agent

import (
	"context"
	"errors"

	"github.com/codepod-io/codepod/pkg/channel"
	"github.com/codepod-io/codepod/pkg/router"
	"github.com/codepod-io/codepod/pkg/rpc"
	"github.com/codepod-io/codepod/pkg/resolver"
)

// Agent represents a CodePod agent (Local, Workspace, or Container).
type Agent struct {
	Type           router.AgentType
	Router         *router.Router
	Resolver       *resolver.Resolver
	RPCServer      *rpc.RPCServer
	RPCClient      *rpc.RPCClient
	Channel        channel.Channel
	providers      map[string]Provider
	mu             int // for future use
}

// Provider defines the interface for environment providers.
type Provider interface {
	Type() string
	Bootstrap(ctx context.Context, identity string) (*rpc.RPCClient, error)
	Connect(ctx context.Context, identity string) (*rpc.RPCClient, error)
}

// Option is a functional option for Agent.
type Option func(*Agent)

// WithRouter sets the router for the agent.
func WithRouter(r *router.Router) Option {
	return func(a *Agent) {
		a.Router = r
	}
}

// WithResolver sets the resolver for the agent.
func WithResolver(r *resolver.Resolver) Option {
	return func(a *Agent) {
		a.Resolver = r
	}
}

// WithChannel sets the channel for the agent.
func WithChannel(ch channel.Channel) Option {
	return func(a *Agent) {
		a.Channel = ch
	}
}

// NewAgent creates a new Agent instance.
func NewAgent(agentType router.AgentType, opts ...Option) *Agent {
	a := &Agent{
		Type:     agentType,
		Resolver: resolver.NewResolver(),
		Router:  router.NewRouter(agentType, string(agentType)+"-1"),
		RPCServer: rpc.NewRPCServer(),
		providers: make(map[string]Provider),
	}

	// Apply options
	for _, opt := range opts {
		opt(a)
	}

	// Register default RPC handlers
	a.registerHandlers()

	return a
}

// NewLocalAgent creates a new Local Agent.
func NewLocalAgent(opts ...Option) *Agent {
	return NewAgent(router.AgentTypeLocal, opts...)
}

// NewWorkspaceAgent creates a new Workspace Agent.
func NewWorkspaceAgent(opts ...Option) *Agent {
	return NewAgent(router.AgentTypeWorkspace, opts...)
}

// NewContainerAgent creates a new Container Agent.
func NewContainerAgent(opts ...Option) *Agent {
	return NewAgent(router.AgentTypeContainer, opts...)
}

// RegisterProvider registers a provider for an agent type.
func (a *Agent) RegisterProvider(p Provider) {
	a.providers[p.Type()] = p
}

// GetProvider returns a provider by type.
func (a *Agent) GetProvider(providerType string) (Provider, bool) {
	p, ok := a.providers[providerType]
	return p, ok
}

// Route handles the routing of a URI through the agent hierarchy.
// This implements the "peeling onion" recursive routing.
func (a *Agent) Route(ctx context.Context, authority, path string) error {
	// 1. Parse the authority
	auth, err := a.Resolver.Resolve(authority)
	if err != nil {
		return err
	}

	// 2. Check if it's a local service
	if a.isLocalService(auth.Provider) {
		return a.handleLocalService(path)
	}

	// 3. Look up in routing table
	child := a.Router.FindChild(authority)
	if child != nil {
		// 4a. Forward to existing child
		return a.forwardToChild(ctx, child, path)
	}

	// 4b. Bootstrap new child via provider
	return a.bootstrapAndForward(ctx, auth, path)
}

// isLocalService checks if the provider refers to a local service.
func (a *Agent) isLocalService(provider string) bool {
	// Local services would be handled directly by the agent
	// For now, only container-level providers need forwarding
	switch provider {
	case "wsl", "ssh-remote", "docker-container", "dev-container":
		return false
	default:
		return true
	}
}

// handleLocalService handles requests for local services.
func (a *Agent) handleLocalService(path string) error {
	// Extract service name from path
	// e.g., "/pty" -> service = "pty"
	if path == "" || path == "/" {
		return errors.New("no service specified")
	}

	// Get service name (first segment after /)
	var serviceName string
	if path[0] == '/' {
		serviceName = path[1:]
	} else {
		serviceName = path
	}

	// Look up service handler
	handler, ok := a.Router.GetService(serviceName)
	if !ok {
		return errors.New("unknown service: " + serviceName)
	}

	// Call service handler
	_, err := handler(nil)
	return err
}

// forwardToChild forwards a request to an existing child agent.
func (a *Agent) forwardToChild(ctx context.Context, child *router.RouteNode, path string) error {
	// Get RPC client from connection
	client, ok := child.Connection.(*rpc.RPCClient)
	if !ok {
		return errors.New("invalid child connection type")
	}

	// Call child's Route method via RPC
	_, err := client.Call(ctx, "Agent.Route", map[string]string{
		"path": path,
	})
	return err
}

// bootstrapAndForward bootstraps a new child agent and forwards the request.
func (a *Agent) bootstrapAndForward(ctx context.Context, auth *resolver.Authority, path string) error {
	// Get provider
	provider, ok := a.GetProvider(auth.Provider)
	if !ok {
		return errors.New("no provider for: " + auth.Provider)
	}

	// Determine target identity
	var target string
	switch id := auth.IdentityParsed.(type) {
	case string:
		target = id
	case *resolver.DevContainerConfig:
		// For dev-container, use the path as target
		target = id.Path
	default:
		target = auth.Identity
	}

	// Bootstrap the child agent
	client, err := provider.Bootstrap(ctx, target)
	if err != nil {
		return err
	}

	// Add to routing table
	a.Router.AddChild(authorityToKey(auth), &router.RouteNode{
		AgentType:  router.AgentTypeWorkspace,
		AgentID:    target,
		Connection: client,
	})

	// Forward the remaining path
	return a.forwardToChild(ctx, &router.RouteNode{Connection: client}, path)
}

// authorityToKey creates a routing key from authority.
func authorityToKey(auth *resolver.Authority) string {
	return auth.Provider + "+" + auth.Identity
}

// registerHandlers registers the default RPC handlers.
func (a *Agent) registerHandlers() {
	// Register Agent service
	a.RPCServer.Register("Agent.Route", func(ctx context.Context, params any) (any, error) {
		type RouteParams struct {
			Authority string `json:"authority"`
			Path      string `json:"path"`
		}
		var p RouteParams
		if params != nil {
			if m, ok := params.(map[string]any); ok {
				if v, ok := m["authority"]; ok {
					p.Authority, _ = v.(string)
				}
				if v, ok := m["path"]; ok {
					p.Path, _ = v.(string)
				}
			}
		}
		err := a.Route(ctx, p.Authority, p.Path)
		return map[string]any{"success": err == nil, "error": err}, err
	})

	a.RPCServer.Register("Agent.ListChildren", func() (any, error) {
		return a.Router.GetChildren(), nil
	})

	// Register Resolver service
	a.RPCServer.Register("Resolver.Resolve", func(ctx context.Context, params any) (any, error) {
		type ResolveParams struct {
			Authority string `json:"authority"`
		}
		var p ResolveParams
		if params != nil {
			if m, ok := params.(map[string]any); ok {
				if v, ok := m["authority"]; ok {
					p.Authority, _ = v.(string)
				}
			}
		}
		auth, err := a.Resolver.Resolve(p.Authority)
		if err != nil {
			return nil, err
		}
		return auth, nil
	})
}
