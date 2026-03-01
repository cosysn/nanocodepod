package port

import (
	"fmt"
	"sync"

	"github.com/codepod-io/codepod/internal/config"
	"github.com/codepod-io/codepod/internal/types"
)

// Pool manages port allocation for workspaces
type Pool struct {
	mu     sync.Mutex
	config *types.PortPool
}

// New creates a new port pool
func New() (*Pool, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		// Use default config if loading fails
		cfg = config.GetDefaultConfig()
	}

	return &Pool{
		config: &cfg.PortPool,
	}, nil
}

// Allocate allocates an available port
func (p *Pool) Allocate() (int, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Find first available port
	for port := p.config.Start; port <= p.config.End; port++ {
		if !p.isUsed(port) {
			p.config.Used = append(p.config.Used, port)
			if err := p.save(); err != nil {
				return 0, err
			}
			return port, nil
		}
	}

	return 0, fmt.Errorf("no available ports in range %d-%d", p.config.Start, p.config.End)
}

// Release releases an allocated port
func (p *Pool) Release(port int) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.config.Used = removePort(p.config.Used, port)
	return p.save()
}

// IsAllocated checks if a port is allocated
func (p *Pool) IsAllocated(port int) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.isUsed(port)
}

// GetAllocatedPorts returns all allocated ports
func (p *Pool) GetAllocatedPorts() []int {
	p.mu.Lock()
	defer p.mu.Unlock()

	result := make([]int, len(p.config.Used))
	copy(result, p.config.Used)
	return result
}

// SetUsed sets the list of used ports directly
func (p *Pool) SetUsed(ports []int) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.config.Used = ports
	return p.save()
}

// GetConfig returns the port pool configuration
func (p *Pool) GetConfig() *types.PortPool {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.config
}

func (p *Pool) isUsed(port int) bool {
	for _, used := range p.config.Used {
		if used == port {
			return true
		}
	}
	return false
}

func removePort(ports []int, port int) []int {
	result := make([]int, 0)
	for _, p := range ports {
		if p != port {
			result = append(result, p)
		}
	}
	return result
}

func (p *Pool) save() error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	cfg.PortPool = *p.config
	return config.SaveConfig(cfg)
}
