package agent

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

// AgentInfo holds agent registration information
type AgentInfo struct {
	Port     int       `json:"port"`
	Status   string    `json:"status"`
	Hostname string    `json:"hostname"`
	LastSeen time.Time `json:"last_seen"`
}

// Registry manages agent registrations
type Registry struct {
	agents map[string]*AgentInfo
	mu     sync.RWMutex
}

// Global registry
var registry = &Registry{
	agents: make(map[string]*AgentInfo),
}

// RegisterAgent handles agent heartbeat/registration
func RegisterAgent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req HeartbeatRequest
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Register or update agent
	registry.mu.Lock()
	registry.agents[req.Hostname] = &AgentInfo{
		Port:     req.AgentPort,
		Status:   req.Status,
		Hostname: req.Hostname,
		LastSeen: time.Now(),
	}
	registry.mu.Unlock()

	log.Printf("Agent registered: %s (port: %d)", req.Hostname, req.AgentPort)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

// ListAgents lists all registered agents
func ListAgents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	registry.mu.RLock()
	agents := make([]*AgentInfo, 0, len(registry.agents))
	for _, agent := range registry.agents {
		agents = append(agents, agent)
	}
	registry.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(agents)
}

// HeartbeatRequest represents a heartbeat request
type HeartbeatRequest struct {
	AgentPort int    `json:"agent_port"`
	Status    string `json:"status"`
	Hostname  string `json:"hostname"`
}

// GetAgent returns agent info by hostname
func GetAgent(hostname string) (*AgentInfo, error) {
	registry.mu.RLock()
	defer registry.mu.RUnlock()

	agent, ok := registry.agents[hostname]
	if !ok {
		return nil, fmt.Errorf("agent not found")
	}
	return agent, nil
}
