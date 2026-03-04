package agent

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// Stats represents container resource statistics
type Stats struct {
	CPU       float64   `json:"cpu"`
	Memory    int       `json:"memory"`
	MemoryLimit int     `json:"memory_limit"`
	Disk      int       `json:"disk"`
	NetworkRx int       `json:"network_rx"`
	NetworkTx int       `json:"network_tx"`
	Timestamp time.Time `json:"timestamp"`
}

// Monitor monitors container resources
type Monitor struct {
	containerID string
	interval    time.Duration
	statsChan   chan *Stats
	stopChan    chan bool
}

// NewMonitor creates a new container monitor
func NewMonitor(containerID string, interval time.Duration) *Monitor {
	return &Monitor{
		containerID: containerID,
		interval:    interval,
		statsChan:   make(chan *Stats, 10),
		stopChan:    make(chan bool),
	}
}

// Start starts the monitor
func (m *Monitor) Start() {
	go m.run()
}

// Stop stops the monitor
func (m *Monitor) Stop() {
	m.stopChan <- true
}

// Stats returns the stats channel
func (m *Monitor) Stats() <-chan *Stats {
	return m.statsChan
}

func (m *Monitor) run() {
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			stats, err := m.collectStats()
			if err == nil {
				m.statsChan <- stats
			}
		case <-m.stopChan:
			close(m.statsChan)
			return
		}
	}
}

func (m *Monitor) collectStats() (*Stats, error) {
	cmd := exec.Command("docker", "stats", "--no-stream", "--format", "{{.CPUPerc}}|{{.MemUsage}}|{{.NetIO}}|{{.BlockIO}}", m.containerID)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	parts := strings.Split(strings.TrimSpace(string(out)), "|")
	if len(parts) < 4 {
		return nil, fmt.Errorf("unexpected output format")
	}

	stats := &Stats{
		Timestamp: time.Now(),
	}

	// Parse CPU
	cpu := strings.TrimSuffix(parts[0], "%")
	fmt.Sscanf(cpu, "%f", &stats.CPU)

	// Parse Memory
	memParts := strings.Split(parts[1], "/")
	if len(memParts) >= 2 {
		stats.Memory = parseSize(memParts[0])
		stats.MemoryLimit = parseSize(memParts[1])
	}

	// Parse Network
	netParts := strings.Split(parts[2], "/")
	if len(netParts) >= 2 {
		stats.NetworkRx = parseSize(netParts[0])
		stats.NetworkTx = parseSize(netParts[1])
	}

	// Parse Block IO
	diskParts := strings.Split(parts[3], "/")
	if len(diskParts) >= 2 {
		stats.Disk = parseSize(diskParts[1])
	}

	return stats, nil
}

func parseSize(s string) int {
	s = strings.TrimSpace(s)
	s = strings.ToUpper(s)

	var multiplier float64 = 1
	if strings.HasSuffix(s, "GB") || strings.HasSuffix(s, "GIB") {
		multiplier = 1024 * 1024 * 1024
		s = strings.TrimSuffix(s, "GB")
		s = strings.TrimSuffix(s, "GIB")
	} else if strings.HasSuffix(s, "MB") || strings.HasSuffix(s, "MIB") {
		multiplier = 1024 * 1024
		s = strings.TrimSuffix(s, "MB")
		s = strings.TrimSuffix(s, "MIB")
	} else if strings.HasSuffix(s, "KB") || strings.HasSuffix(s, "KIB") {
		multiplier = 1024
		s = strings.TrimSuffix(s, "KB")
		s = strings.TrimSuffix(s, "KIB")
	} else if strings.HasSuffix(s, "B") {
		s = strings.TrimSuffix(s, "B")
	}

	var value float64
	fmt.Sscanf(s, "%f", &value)
	return int(value * multiplier)
}

// GetStatsJSON returns stats as JSON
func (s *Stats) GetStatsJSON() (string, error) {
	data, err := json.Marshal(s)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// GetContainerHealth checks container health
func GetContainerHealth(containerID string) (string, error) {
	cmd := exec.Command("docker", "inspect", "--format", "{{.State.Health.Status}}", containerID)
	out, err := cmd.Output()
	if err != nil {
		// If no health check, return "unknown"
		return "unknown", nil
	}
	return strings.TrimSpace(string(out)), nil
}

// GetContainerUptime returns container uptime
func GetContainerUptime(containerID string) (time.Duration, error) {
	cmd := exec.Command("docker", "inspect", "--format", "{{.State.StartedAt}}", containerID)
	out, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	started, err := time.Parse(time.RFC3339, strings.TrimSpace(string(out)))
	if err != nil {
		return 0, err
	}

	return time.Since(started), nil
}
