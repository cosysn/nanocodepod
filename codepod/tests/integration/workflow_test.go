package integration

import (
	"os"
	"testing"

	"github.com/codepod-io/codepod/internal/config"
	"github.com/codepod-io/codepod/internal/workspace"
)

// TestFullWorkflow tests the complete workspace lifecycle
func TestFullWorkflow(t *testing.T) {
	// Setup - ensure clean state
	testName := "test-workflow-" + string(rune(os.Getpid()))
	cleanupTestWorkspace(testName)
	defer cleanupTestWorkspace(testName)

	// Initialize config
	if err := config.EnsureConfigDir(); err != nil {
		t.Fatalf("failed to ensure config dir: %v", err)
	}

	// Test 1: Create workspace manager
	wsm, err := workspace.New()
	if err != nil {
		t.Fatalf("failed to create workspace manager: %v", err)
	}

	// Test 2: Create workspace
	opts := &workspace.CreateOptions{
		Image: "ubuntu:22.04",
	}

	ws, err := wsm.Create(testName, opts)
	if err != nil {
		t.Fatalf("failed to create workspace: %v", err)
	}

	if ws.Name != testName {
		t.Errorf("expected workspace name %s, got %s", testName, ws.Name)
	}

	// Test 3: Start workspace
	ws, err = wsm.Start(testName)
	if err != nil {
		t.Fatalf("failed to start workspace: %v", err)
	}

	if ws.State != "running" {
		t.Errorf("expected state running, got %s", ws.State)
	}

	// Test 4: List workspaces
	workspaces, err := wsm.List()
	if err != nil {
		t.Fatalf("failed to list workspaces: %v", err)
	}

	found := false
	for _, w := range workspaces {
		if w.Name == testName {
			found = true
			break
		}
	}

	if !found {
		t.Error("created workspace not found in list")
	}

	// Test 5: Stop workspace
	ws, err = wsm.Stop(testName)
	if err != nil {
		t.Fatalf("failed to stop workspace: %v", err)
	}

	if ws.State != "stopped" {
		t.Errorf("expected state stopped, got %s", ws.State)
	}

	// Test 6: Start again (idempotency)
	ws, err = wsm.Start(testName)
	if err != nil {
		t.Fatalf("failed to start workspace again: %v", err)
	}

	// Test 7: Delete workspace
	err = wsm.Delete(testName)
	if err != nil {
		t.Fatalf("failed to delete workspace: %v", err)
	}

	// Test 8: Verify deletion
	exists, _ := wsm.Exists(testName)
	if exists {
		t.Error("workspace still exists after deletion")
	}
}

// TestIdempotency tests command idempotency
func TestIdempotency(t *testing.T) {
	testName := "test-idempotent"
	cleanupTestWorkspace(testName)
	defer cleanupTestWorkspace(testName)

	// Initialize
	config.EnsureConfigDir()
	wsm, _ := workspace.New()

	// Create multiple times
	_, err := wsm.Create(testName, &workspace.CreateOptions{Image: "ubuntu:22.04"})
	if err != nil {
		t.Fatalf("failed to create workspace: %v", err)
	}

	// Second create should return existing
	ws, err := wsm.Create(testName, &workspace.CreateOptions{Image: "ubuntu:22.04"})
	if err != nil {
		t.Fatalf("idempotent create failed: %v", err)
	}

	if ws.Name != testName {
		t.Errorf("expected existing workspace, got new one")
	}

	// Start multiple times
	_, err = wsm.Start(testName)
	if err != nil {
		t.Fatalf("first start failed: %v", err)
	}

	_, err = wsm.Start(testName)
	if err != nil {
		t.Fatalf("idempotent start failed: %v", err)
	}

	// Stop multiple times
	_, err = wsm.Stop(testName)
	if err != nil {
		t.Fatalf("first stop failed: %v", err)
	}

	_, err = wsm.Stop(testName)
	if err != nil {
		t.Fatalf("idempotent stop failed: %v", err)
	}

	// Cleanup
	wsm.Delete(testName)
}

// TestConfigPersistence tests config persistence across restarts
func TestConfigPersistence(t *testing.T) {
	testName := "test-config-persist"
	cleanupTestWorkspace(testName)
	defer cleanupTestWorkspace(testName)

	// Initialize config
	config.EnsureConfigDir()
	cfg := config.GetDefaultConfig()
	config.SaveConfig(cfg)

	// Create workspace manager
	wsm, _ := workspace.New()

	// Create and start
	wsm.Create(testName, &workspace.CreateOptions{Image: "ubuntu:22.04"})
	wsm.Start(testName)

	// Simulate restart - new manager
	wsm2, _ := workspace.New()

	// Get workspace
	ws, err := wsm2.Get(testName)
	if err != nil {
		t.Fatalf("failed to get workspace after restart: %v", err)
	}

	if ws.Name != testName {
		t.Errorf("workspace name mismatch: %s != %s", ws.Name, testName)
	}

	// Cleanup
	wsm2.Delete(testName)
}

func cleanupTestWorkspace(name string) {
	wsm, err := workspace.New()
	if err != nil {
		return
	}
	wsm.Delete(name)
}
