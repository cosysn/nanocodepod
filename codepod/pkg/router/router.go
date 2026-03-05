// Package router provides the routing engine for CodePod agents.
// It maintains parent-child relationships between agents and supports
// service registration and lookup.
package router

import (
	"errors"
	"sync"
)

// AgentType represents the type of agent in the hierarchy.
type AgentType string

const (
	AgentTypeLocal      AgentType = "local"
	AgentTypeWorkspace  AgentType = "workspace"
	AgentTypeContainer  AgentType = "container"
)

// ServiceHandler is a function that handles service requests.
type ServiceHandler func(params any) (any, error)

// RouteNode represents a node in the routing tree (external child representation).
type RouteNode struct {
	AgentType   AgentType
	AgentID     string
	Connection  interface{} // RPC client connection to the agent
	Parent      *Router     // Reference to parent router
	Services    map[string]ServiceHandler
}

// Router manages routing table with parent-child relationships.
type Router struct {
	agentType AgentType
	agentID   string
	parent    *RouteNode
	children  map[string]*RouteNode
	services  map[string]ServiceHandler
	mu        sync.RWMutex
}

// NewRouter creates a new Router instance.
func NewRouter(agentType AgentType, agentID string) *Router {
	return &Router{
		agentType: agentType,
		agentID:   agentID,
		children:  make(map[string]*RouteNode),
		services:  make(map[string]ServiceHandler),
	}
}

// SetParent sets the parent node for this router.
func (r *Router) SetParent(parent *RouteNode) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.parent = parent
}

// GetParent returns the parent node.
func (r *Router) GetParent() *RouteNode {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.parent
}

// AddChild adds a child route to the routing table.
func (r *Router) AddChild(authority string, node *RouteNode) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Set the parent reference
	node.Parent = &Router{
		agentType: r.agentType,
		agentID:   r.agentID,
		parent:    r.parent,
		services:  r.services,
	}

	r.children[authority] = node
}

// RemoveChild removes a child route from the routing table.
func (r *Router) RemoveChild(authority string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.children, authority)
}

// FindChild finds a child route by authority.
func (r *Router) FindChild(authority string) *RouteNode {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.children[authority]
}

// GetChildren returns all child routes.
func (r *Router) GetChildren() map[string]*RouteNode {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[string]*RouteNode)
	for k, v := range r.children {
		result[k] = v
	}
	return result
}

// RegisterService registers a service handler.
func (r *Router) RegisterService(name string, handler ServiceHandler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.services[name] = handler
}

// GetService returns a service handler by name.
func (r *Router) GetService(name string) (ServiceHandler, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	handler, ok := r.services[name]
	return handler, ok
}

// GetServices returns all registered services.
func (r *Router) GetServices() map[string]ServiceHandler {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[string]ServiceHandler)
	for k, v := range r.services {
		result[k] = v
	}
	return result
}

// AgentType returns the agent type.
func (r *Router) AgentType() AgentType {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.agentType
}

// AgentID returns the agent ID.
func (r *Router) AgentID() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.agentID
}

// ValidateAgentType validates if the agent type is valid.
func ValidateAgentType(t string) error {
	switch AgentType(t) {
	case AgentTypeLocal, AgentTypeWorkspace, AgentTypeContainer:
		return nil
	default:
		return errors.New("invalid agent type: " + t)
	}
}
