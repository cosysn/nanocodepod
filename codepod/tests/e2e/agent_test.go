package e2e

import (
	"fmt"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// TestCLI_AgentInjection tests agent injection via CLI flag
func TestCLI_AgentInjection(t *testing.T) {
	binary := getBinaryPath()
	workspaceName := "e2e-agent"

	// Cleanup
	cmd := exec.Command(binary, "delete", workspaceName)
	cmd.Dir = "/home/ubuntu/nanocodepod/codepod"
	cmd.Output()
	time.Sleep(1 * time.Second)

	// Create workspace with agent (default)
	cmd = exec.Command(binary, "up", workspaceName, "--image", "ubuntu:22.04")
	cmd.Dir = "/home/ubuntu/nanocodepod/codepod"
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("up with agent failed: %v, output: %s", err, string(out))
	}

	// Verify output contains agent status
	output := string(out)
	if !strings.Contains(output, "Agent:") {
		t.Errorf("expected 'Agent:' in output: %s", output)
	}

	// Verify container is running
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

// TestCLI_AgentDisabled tests agent disabled via CLI flag
func TestCLI_AgentDisabled(t *testing.T) {
	binary := getBinaryPath()
	workspaceName := "e2e-no-agent"

	// Cleanup
	cmd := exec.Command(binary, "delete", workspaceName)
	cmd.Dir = "/home/ubuntu/nanocodepod/codepod"
	cmd.Output()
	time.Sleep(1 * time.Second)

	// Create workspace without agent
	cmd = exec.Command(binary, "up", workspaceName, "--image", "ubuntu:22.04", "--no-agent")
	cmd.Dir = "/home/ubuntu/nanocodepod/codepod"
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("up with --no-agent failed: %v, output: %s", err, string(out))
	}

	// Verify output contains agent status
	output := string(out)
	if !strings.Contains(output, "Agent:") {
		t.Errorf("expected 'Agent:' in output: %s", output)
	}

	// Verify container is running
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

// TestCLI_StartWithAgent tests start command with agent flag
func TestCLI_StartWithAgent(t *testing.T) {
	binary := getBinaryPath()
	workspaceName := "e2e-start-agent"

	// Cleanup
	cmd := exec.Command(binary, "delete", workspaceName)
	cmd.Dir = "/home/ubuntu/nanocodepod/codepod"
	cmd.Output()
	time.Sleep(1 * time.Second)

	// Create workspace without agent first
	cmd = exec.Command(binary, "up", workspaceName, "--image", "ubuntu:22.04", "--no-agent")
	cmd.Dir = "/home/ubuntu/nanocodepod/codepod"
	cmd.Output()
	time.Sleep(2 * time.Second)

	// Stop workspace
	cmd = exec.Command(binary, "stop", workspaceName)
	cmd.Dir = "/home/ubuntu/nanocodepod/codepod"
	cmd.Output()
	time.Sleep(1 * time.Second)

	// Start with agent
	cmd = exec.Command(binary, "start", workspaceName, "--agent")
	cmd.Dir = "/home/ubuntu/nanocodepod/codepod"
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("start with --agent failed: %v, output: %s", err, string(out))
	}

	// Verify output contains agent status
	output := string(out)
	if !strings.Contains(output, "Agent:") {
		t.Errorf("expected 'Agent:' in output: %s", output)
	}

	// Cleanup
	cmd = exec.Command(binary, "delete", workspaceName)
	cmd.Dir = "/home/ubuntu/nanocodepod/codepod"
	cmd.Output()
}

// TestCLI_StartWithoutAgent tests start command without agent flag
func TestCLI_StartWithoutAgent(t *testing.T) {
	binary := getBinaryPath()
	workspaceName := "e2e-start-no-agent"

	// Cleanup
	cmd := exec.Command(binary, "delete", workspaceName)
	cmd.Dir = "/home/ubuntu/nanocodepod/codepod"
	cmd.Output()
	time.Sleep(1 * time.Second)

	// Create workspace with agent
	cmd = exec.Command(binary, "up", workspaceName, "--image", "ubuntu:22.04", "--agent")
	cmd.Dir = "/home/ubuntu/nanocodepod/codepod"
	cmd.Output()
	time.Sleep(2 * time.Second)

	// Stop workspace
	cmd = exec.Command(binary, "stop", workspaceName)
	cmd.Dir = "/home/ubuntu/nanocodepod/codepod"
	cmd.Output()
	time.Sleep(1 * time.Second)

	// Start without agent
	cmd = exec.Command(binary, "start", workspaceName, "--no-agent")
	cmd.Dir = "/home/ubuntu/nanocodepod/codepod"
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("start with --no-agent failed: %v, output: %s", err, string(out))
	}

	// Verify output contains agent status
	output := string(out)
	if !strings.Contains(output, "Agent:") {
		t.Errorf("expected 'Agent:' in output: %s", output)
	}

	// Cleanup
	cmd = exec.Command(binary, "delete", workspaceName)
	cmd.Dir = "/home/ubuntu/nanocodepod/codepod"
	cmd.Output()
}

// TestCLI_AgentHelp tests that agent flags are shown in help
func TestCLI_AgentHelp(t *testing.T) {
	binary := getBinaryPath()
	t.Logf("Using binary: %s", binary)

	// Check up command help
	cmd := exec.Command(binary, "up", "--help")
	cmd.Dir = "/home/ubuntu/nanocodepod/codepod"
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("up --help failed: %v, output: %s", err, string(out))
	}

	output := string(out)
	t.Logf("UP help output: %s", output)
	if !strings.Contains(output, "--agent") {
		t.Error("up help should contain --agent flag")
	}

	// Check start command help
	cmd = exec.Command(binary, "start", "--help")
	cmd.Dir = "/home/ubuntu/nanocodepod/codepod"
	out, err = cmd.Output()
	if err != nil {
		t.Fatalf("start --help failed: %v, output: %s", err, string(out))
	}

	output = string(out)
	t.Logf("START help output: %s", output)
	if !strings.Contains(output, "--agent") {
		t.Error("start help should contain --agent flag")
	}
}
