package port

import (
	"os"
	"testing"

	"github.com/codepod-io/codepod/internal/config"
)

func TestPool_Allocate(t *testing.T) {
	// Setup clean config
	dir, _ := config.GetConfigDir()
	os.RemoveAll(dir)
	config.EnsureConfigDir()

	// Save default config
	cfg := config.GetDefaultConfig()
	cfg.PortPool.Used = []int{}
	config.SaveConfig(cfg)

	pool, err := New()
	if err != nil {
		t.Fatalf("failed to create pool: %v", err)
	}

	// Test allocation
	port, err := pool.Allocate()
	if err != nil {
		t.Fatalf("failed to allocate: %v", err)
	}

	if port < 22000 || port > 22999 {
		t.Errorf("expected port in range 22000-22999, got %d", port)
	}

	// Cleanup
	pool.Release(port)
	os.RemoveAll(dir)
}

func TestPool_Release(t *testing.T) {
	// Setup clean config
	dir, _ := config.GetConfigDir()
	os.RemoveAll(dir)
	config.EnsureConfigDir()

	cfg := config.GetDefaultConfig()
	cfg.PortPool.Used = []int{}
	config.SaveConfig(cfg)

	pool, err := New()
	if err != nil {
		t.Fatalf("failed to create pool: %v", err)
	}

	// Allocate
	port, _ := pool.Allocate()

	// Release
	err = pool.Release(port)
	if err != nil {
		t.Fatalf("failed to release: %v", err)
	}

	// Verify it's released
	if pool.IsAllocated(port) {
		t.Error("port should be released")
	}

	os.RemoveAll(dir)
}

func TestPool_IsAllocated(t *testing.T) {
	dir, _ := config.GetConfigDir()
	os.RemoveAll(dir)
	config.EnsureConfigDir()

	cfg := config.GetDefaultConfig()
	cfg.PortPool.Used = []int{}
	config.SaveConfig(cfg)

	pool, _ := New()

	if pool.IsAllocated(22000) {
		t.Error("port should not be allocated")
	}

	pool.Allocate()

	if !pool.IsAllocated(22000) {
		t.Error("port should be allocated")
	}

	os.RemoveAll(dir)
}

func TestPool_GetAllocatedPorts(t *testing.T) {
	dir, _ := config.GetConfigDir()
	os.RemoveAll(dir)
	config.EnsureConfigDir()

	cfg := config.GetDefaultConfig()
	cfg.PortPool.Used = []int{}
	config.SaveConfig(cfg)

	pool, _ := New()

	p1, _ := pool.Allocate()
	p2, _ := pool.Allocate()

	ports := pool.GetAllocatedPorts()

	if len(ports) != 2 {
		t.Errorf("expected 2 ports, got %d", len(ports))
	}

	if ports[0] != p1 || ports[1] != p2 {
		t.Errorf("unexpected ports: %v", ports)
	}

	os.RemoveAll(dir)
}

func TestPool_Allocate_Exhausted(t *testing.T) {
	dir, _ := config.GetConfigDir()
	os.RemoveAll(dir)
	config.EnsureConfigDir()

	cfg := config.GetDefaultConfig()
	cfg.PortPool.Start = 22000
	cfg.PortPool.End = 22001
	cfg.PortPool.Used = []int{}
	config.SaveConfig(cfg)

	pool, _ := New()

	// Allocate all ports
	p1, err := pool.Allocate()
	if err != nil {
		t.Fatalf("first allocation failed: %v", err)
	}

	p2, err := pool.Allocate()
	if err != nil {
		t.Fatalf("second allocation failed: %v", err)
	}

	// Third should fail
	_, err = pool.Allocate()
	if err == nil {
		t.Error("expected error for exhausted pool")
	}

	// Cleanup
	pool.Release(p1)
	pool.Release(p2)
	os.RemoveAll(dir)
}

func TestPool_SetUsed(t *testing.T) {
	dir, _ := config.GetConfigDir()
	os.RemoveAll(dir)
	config.EnsureConfigDir()

	cfg := config.GetDefaultConfig()
	cfg.PortPool.Used = []int{}
	config.SaveConfig(cfg)

	pool, _ := New()

	// Set used ports
	err := pool.SetUsed([]int{22005, 22006, 22007})
	if err != nil {
		t.Fatalf("failed to set used: %v", err)
	}

	// Verify
	if !pool.IsAllocated(22005) {
		t.Error("port 22005 should be allocated")
	}
	if !pool.IsAllocated(22006) {
		t.Error("port 22006 should be allocated")
	}
	if !pool.IsAllocated(22007) {
		t.Error("port 22007 should be allocated")
	}

	os.RemoveAll(dir)
}

func TestPool_GetConfig(t *testing.T) {
	dir, _ := config.GetConfigDir()
	os.RemoveAll(dir)
	config.EnsureConfigDir()

	cfg := config.GetDefaultConfig()
	cfg.PortPool.Start = 22000
	cfg.PortPool.End = 22010
	cfg.PortPool.Used = []int{}
	config.SaveConfig(cfg)

	pool, _ := New()

	// Get config
	configResult := pool.GetConfig()

	if configResult.Start != 22000 {
		t.Errorf("want start 22000, got %d", configResult.Start)
	}
	if configResult.End != 22010 {
		t.Errorf("want end 22010, got %d", configResult.End)
	}

	os.RemoveAll(dir)
}
