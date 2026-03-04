package port

import (
	"fmt"
	"sync"
)

const (
	// DefaultPortRangeStart is the default start of port pool
	DefaultPortRangeStart = 22000
	// DefaultPortRangeEnd is the default end of port pool
	DefaultPortRangeEnd = 22999
)

// Pool manages port allocation
type Pool struct {
	start     int
	end       int
	allocated map[int]bool
	mu        sync.Mutex
}

// NewPool creates a new port pool
func NewPool(start, end int) *Pool {
	return &Pool{
		start:     start,
		end:       end,
		allocated: make(map[int]bool),
	}
}

// Allocate allocates an available port
func (p *Pool) Allocate() (int, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for port := p.start; port <= p.end; port++ {
		if !p.allocated[port] {
			p.allocated[port] = true
			return port, nil
		}
	}
	return 0, fmt.Errorf("no available ports in range %d-%d", p.start, p.end)
}

// Release releases an allocated port
func (p *Pool) Release(port int) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if port < p.start || port > p.end {
		return fmt.Errorf("port %d is outside the allocated range", port)
	}

	if !p.allocated[port] {
		return fmt.Errorf("port %d is not allocated", port)
	}

	delete(p.allocated, port)
	return nil
}

// IsAllocated checks if a port is allocated
func (p *Pool) IsAllocated(port int) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.allocated[port]
}
