package e2e

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// getBinaryPath returns the path to the codepod binary
func getBinaryPath() string {
	// Check current directory
	if _, err := os.Stat("./codepod"); err == nil {
		abs, _ := filepath.Abs("./codepod")
		return abs
	}
	// Check parent directory
	if _, err := os.Stat("../codepod"); err == nil {
		abs, _ := filepath.Abs("../codepod")
		return abs
	}
	// Fallback to current dir
	return "./codepod"
}

// getTestConfigPath returns a config path for tests
func getTestConfigPath() string {
	return "/tmp/codepod-e2e-config"  // This is passed as --config which expects a directory
}

// TestCLI_Config_List tests config list
func TestCLI_Config_List(t *testing.T) {
	binary := getBinaryPath()
	cmd := exec.Command(binary, "config", "list")
	cmd.Dir = "/home/ubuntu/nanocodepod/codepod"
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("config list failed: %v, output: %s", err, string(out))
	}

	output := string(out)
	if !strings.Contains(output, "WSL Distribution:") {
		t.Error("output should contain WSL Distribution")
	}
}

// TestCLI_Config_Set_Get tests config set and get
func TestCLI_Config_Set_Get(t *testing.T) {
	binary := getBinaryPath()

	// Set value
	cmd := exec.Command(binary, "config", "set", "wsl.distribution", "Ubuntu-20.04")
	cmd.Dir = "/home/ubuntu/nanocodepod/codepod"
	cmd.Output()

	// Get value
	cmd = exec.Command(binary, "config", "get", "wsl.distribution")
	cmd.Dir = "/home/ubuntu/nanocodepod/codepod"
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("config get failed: %v", err)
	}

	output := strings.TrimSpace(string(out))
	if output != "Ubuntu-20.04" {
		t.Errorf("expected Ubuntu-20.04, got %s", output)
	}

	// Reset to default
	cmd = exec.Command(binary, "config", "set", "wsl.distribution", "Ubuntu-22.04")
	cmd.Dir = "/home/ubuntu/nanocodepod/codepod"
	cmd.Output()
}

// TestCLI_Up_CreateStart tests up command (create + start)
func TestCLI_Up_CreateStart(t *testing.T) {
	binary := getBinaryPath()
	workspaceName := "e2e-test"

	// Cleanup
	cmd := exec.Command(binary, "delete", workspaceName)
	cmd.Dir = "/home/ubuntu/nanocodepod/codepod"
	cmd.Output()
	time.Sleep(1 * time.Second)

	// Run up (use --no-agent since test binary doesn't include agent)
	cmd = exec.Command(binary, "up", workspaceName, "--image", "ubuntu:22.04", "--no-agent")
	cmd.Dir = "/home/ubuntu/nanocodepod/codepod"
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("up failed: %v, output: %s", err, string(out))
	}

	// Verify output contains success message
	output := string(out)
	if !strings.Contains(output, "running") {
		t.Errorf("expected 'running' in output, got: %s", output)
	}

	// Verify container exists
	dockerCmd := exec.Command("docker", "ps", "--filter", fmt.Sprintf("name=codepod-%s", workspaceName), "--format", "{{.Names}}")
	out, _ = dockerCmd.Output()
	if !strings.Contains(string(out), workspaceName) {
		t.Error("container should be running")
	}

	// Cleanup
	cmd = exec.Command(binary, "delete", workspaceName)
	cmd.Dir = "/home/ubuntu/nanocodepod/codepod"
	cmd.Output()
}

// TestCLI_List tests list command
func TestCLI_List(t *testing.T) {
	binary := getBinaryPath()
	configPath := getTestConfigPath()
	workspaceName := "e2e-list-test"

	// Create a workspace first
	cmd := exec.Command(binary, "--config", configPath, "delete", workspaceName)
	cmd.Dir = "/home/ubuntu/nanocodepod/codepod"
	cmd.Output()

	cmd = exec.Command(binary, "--config", configPath, "up", workspaceName, "--image", "ubuntu:22.04", "--no-agent")
	cmd.Dir = "/home/ubuntu/nanocodepod/codepod"
	cmd.Output()
	time.Sleep(2 * time.Second)

	// List
	cmd = exec.Command(binary, "--config", configPath, "list")
	cmd.Dir = "/home/ubuntu/nanocodepod/codepod"
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}

	output := string(out)
	if !strings.Contains(output, workspaceName) {
		t.Errorf("list should contain workspace: %s", output)
	}

	// Cleanup
	cmd = exec.Command(binary, "--config", configPath, "delete", workspaceName)
	cmd.Dir = "/home/ubuntu/nanocodepod/codepod"
	cmd.Output()
}

// TestCLI_Stop tests stop command
func TestCLI_Stop(t *testing.T) {
	binary := getBinaryPath()
	workspaceName := "e2e-stop-test"

	// Create workspace
	cmd := exec.Command(binary, "delete", workspaceName)
	cmd.Dir = "/home/ubuntu/nanocodepod/codepod"
	cmd.Output()

	cmd = exec.Command(binary, "up", workspaceName, "--image", "ubuntu:22.04")
	cmd.Dir = "/home/ubuntu/nanocodepod/codepod"
	cmd.Output()
	time.Sleep(2 * time.Second)

	// Stop
	cmd = exec.Command(binary, "stop", workspaceName)
	cmd.Dir = "/home/ubuntu/nanocodepod/codepod"
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("stop failed: %v, output: %s", err, string(out))
	}

	output := string(out)
	if !strings.Contains(output, "stopped") {
		t.Errorf("expected 'stopped' in output: %s", output)
	}

	// Cleanup
	cmd = exec.Command(binary, "delete", workspaceName)
	cmd.Dir = "/home/ubuntu/nanocodepod/codepod"
	cmd.Output()
}

// TestCLI_Start tests start command
func TestCLI_Start(t *testing.T) {
	binary := getBinaryPath()
	workspaceName := "e2e-start-test"

	// Create and stop
	cmd := exec.Command(binary, "delete", workspaceName)
	cmd.Dir = "/home/ubuntu/nanocodepod/codepod"
	cmd.Output()

	cmd = exec.Command(binary, "up", workspaceName, "--image", "ubuntu:22.04")
	cmd.Dir = "/home/ubuntu/nanocodepod/codepod"
	cmd.Output()

	cmd = exec.Command(binary, "stop", workspaceName)
	cmd.Dir = "/home/ubuntu/nanocodepod/codepod"
	cmd.Output()
	time.Sleep(1 * time.Second)

	// Start
	cmd = exec.Command(binary, "start", workspaceName)
	cmd.Dir = "/home/ubuntu/nanocodepod/codepod"
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("start failed: %v, output: %s", err, string(out))
	}

	output := string(out)
	if !strings.Contains(output, "started") {
		t.Errorf("expected 'started' in output: %s", output)
	}

	// Cleanup
	cmd = exec.Command(binary, "delete", workspaceName)
	cmd.Dir = "/home/ubuntu/nanocodepod/codepod"
	cmd.Output()
}

// TestCLI_Delete tests delete command
func TestCLI_Delete(t *testing.T) {
	binary := getBinaryPath()
	workspaceName := "e2e-delete-test"

	// Create workspace
	cmd := exec.Command(binary, "up", workspaceName, "--image", "ubuntu:22.04")
	cmd.Dir = "/home/ubuntu/nanocodepod/codepod"
	cmd.Output()
	time.Sleep(2 * time.Second)

	// Delete
	cmd = exec.Command(binary, "delete", workspaceName)
	cmd.Dir = "/home/ubuntu/nanocodepod/codepod"
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("delete failed: %v, output: %s", err, string(out))
	}

	output := string(out)
	if !strings.Contains(output, "deleted") {
		t.Errorf("expected 'deleted' in output: %s", output)
	}
}

// TestCLI_Idempotency tests command idempotency
func TestCLI_Idempotency(t *testing.T) {
	binary := getBinaryPath()
	workspaceName := "e2e-idempotent"

	// Cleanup
	cmd := exec.Command(binary, "delete", workspaceName)
	cmd.Dir = "/home/ubuntu/nanocodepod/codepod"
	cmd.Output()
	time.Sleep(1 * time.Second)

	// Create twice
	cmd = exec.Command(binary, "up", workspaceName, "--image", "ubuntu:22.04")
	cmd.Dir = "/home/ubuntu/nanocodepod/codepod"
	cmd.Output()
	time.Sleep(2 * time.Second)

	// Second up should succeed (idempotent)
	cmd = exec.Command(binary, "up", workspaceName, "--image", "ubuntu:22.04")
	cmd.Dir = "/home/ubuntu/nanocodepod/codepod"
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("idempotent up failed: %v, output: %s", err, string(out))
	}

	// Cleanup
	cmd = exec.Command(binary, "delete", workspaceName)
	cmd.Dir = "/home/ubuntu/nanocodepod/codepod"
	cmd.Output()
}

// TestCLI_Help tests help command
func TestCLI_Help(t *testing.T) {
	binary := getBinaryPath()
	cmd := exec.Command(binary, "help")
	cmd.Dir = "/home/ubuntu/nanocodepod/codepod"
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("help failed: %v", err)
	}

	output := string(out)
	if !strings.Contains(output, "Usage:") {
		t.Error("help should show usage")
	}
}

// TestCLI_InvalidCommand tests invalid command
func TestCLI_InvalidCommand(t *testing.T) {
	binary := getBinaryPath()
	cmd := exec.Command(binary, "invalid-cmd")
	cmd.Dir = "/home/ubuntu/nanocodepod/codepod"
	_, err := cmd.Output()
	if err == nil {
		t.Error("invalid command should fail")
	}
}

// TestCLI_PortAllocation tests port allocation
func TestCLI_PortAllocation(t *testing.T) {
	binary := getBinaryPath()
	workspaceName1 := "e2e-port1"
	workspaceName2 := "e2e-port2"

	// Cleanup
	cmd := exec.Command(binary, "delete", workspaceName1)
	cmd.Dir = "/home/ubuntu/nanocodepod/codepod"
	cmd.Output()

	cmd = exec.Command(binary, "delete", workspaceName2)
	cmd.Dir = "/home/ubuntu/nanocodepod/codepod"
	cmd.Output()
	time.Sleep(1 * time.Second)

	// Create first workspace
	cmd = exec.Command(binary, "up", workspaceName1, "--image", "ubuntu:22.04")
	cmd.Dir = "/home/ubuntu/nanocodepod/codepod"
	out1, _ := cmd.Output()

	// Create second workspace
	cmd = exec.Command(binary, "up", workspaceName2, "--image", "ubuntu:22.04")
	cmd.Dir = "/home/ubuntu/nanocodepod/codepod"
	out2, _ := cmd.Output()

	// Extract ports
	port1 := extractPort(string(out1))
	port2 := extractPort(string(out2))

	// Verify different ports
	if port1 == port2 && port1 != 0 {
		t.Errorf("workspaces should have different ports, got %d and %d", port1, port2)
	}

	// Cleanup
	cmd = exec.Command(binary, "delete", workspaceName1)
	cmd.Dir = "/home/ubuntu/nanocodepod/codepod"
	cmd.Output()

	cmd = exec.Command(binary, "delete", workspaceName2)
	cmd.Dir = "/home/ubuntu/nanocodepod/codepod"
	cmd.Output()
}

// TestCLI_MultipleWorkspaces tests multiple workspace management
func TestCLI_MultipleWorkspaces(t *testing.T) {
	binary := getBinaryPath()
	configPath := getTestConfigPath()
	workspaces := []string{"e2e-multi1", "e2e-multi2", "e2e-multi3"}

	// Cleanup
	for _, ws := range workspaces {
		cmd := exec.Command(binary, "--config", configPath, "delete", ws)
		cmd.Dir = "/home/ubuntu/nanocodepod/codepod"
		cmd.Output()
	}
	time.Sleep(1 * time.Second)

	// Create multiple workspaces
	for _, ws := range workspaces {
		cmd := exec.Command(binary, "--config", configPath, "up", ws, "--image", "ubuntu:22.04", "--no-agent")
		cmd.Dir = "/home/ubuntu/nanocodepod/codepod"
		cmd.Output()
	}
	time.Sleep(4 * time.Second)

	// List
	cmd := exec.Command(binary, "--config", configPath, "list")
	cmd.Dir = "/home/ubuntu/nanocodepod/codepod"
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}

	// Verify all workspaces in list
	for _, ws := range workspaces {
		if !strings.Contains(string(out), ws) {
			t.Errorf("workspace %s should be in list", ws)
		}
	}

	// Cleanup
	for _, ws := range workspaces {
		cmd := exec.Command(binary, "--config", configPath, "delete", ws)
		cmd.Dir = "/home/ubuntu/nanocodepod/codepod"
		cmd.Output()
	}
}

// TestCLI_PersistentStorage tests persistent storage
func TestCLI_PersistentStorage(t *testing.T) {
	binary := getBinaryPath()
	workspaceName := "e2e-storage"

	// Cleanup
	cmd := exec.Command(binary, "delete", workspaceName)
	cmd.Dir = "/home/ubuntu/nanocodepod/codepod"
	cmd.Output()
	time.Sleep(1 * time.Second)

	// Create workspace
	cmd = exec.Command(binary, "up", workspaceName, "--image", "ubuntu:22.04")
	cmd.Dir = "/home/ubuntu/nanocodepod/codepod"
	cmd.Output()
	time.Sleep(2 * time.Second)

	// Verify storage directory exists
	storagePath := filepath.Join("/root", ".codepod", "workspaces", workspaceName)
	if _, err := os.Stat(storagePath); os.IsNotExist(err) {
		t.Error("storage directory should exist")
	}

	// Delete
	cmd = exec.Command(binary, "delete", workspaceName)
	cmd.Dir = "/home/ubuntu/nanocodepod/codepod"
	cmd.Output()

	// Storage should be cleaned up
	if _, err := os.Stat(storagePath); !os.IsNotExist(err) {
		t.Error("storage should be cleaned up after delete")
	}
}

func extractPort(output string) int {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "Port:") {
			var port int
			fmt.Sscanf(line, "Port: %d", &port)
			return port
		}
	}
	return 0
}
