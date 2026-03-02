package e2e

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// findCodePath finds the code path in workspaces directory by looking for .git
func findCodePath(workspaceName string) string {
	workspacesDir := "/root/.codepod/workspaces"
	entries, _ := os.ReadDir(workspacesDir)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		codePath := filepath.Join(workspacesDir, entry.Name(), "code")
		gitPath := filepath.Join(codePath, ".git")
		if _, err := os.Stat(gitPath); err == nil {
			return codePath
		}
	}
	return ""
}

// TestCLI_LocalPath tests local directory functionality
func TestCLI_LocalPath(t *testing.T) {
	binary := getBinaryPath()
	workspaceName := "e2e-local"

	// Create temp local directory in /tmp to avoid permission issues
	tmpDir := "/tmp/codepod-test-local"
	os.MkdirAll(tmpDir, 0755)

	// Create test file
	if err := os.WriteFile(filepath.Join(tmpDir, "test.txt"), []byte("hello"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Cleanup
	cmd := exec.Command(binary, "delete", workspaceName)
	cmd.Dir = "/home/ubuntu/nanocodepod/codepod"
	cmd.Output()
	time.Sleep(1 * time.Second)

	// Create workspace with local path
	cmd = exec.Command(binary, "up", workspaceName, "--image", "ubuntu:22.04", "--local", tmpDir)
	cmd.Dir = "/home/ubuntu/nanocodepod/codepod"
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("up with local failed: %v, output: %s", err, string(out))
	}

	time.Sleep(2 * time.Second)

	// Verify code was copied - find by looking for .git or test.txt
	codePath := findCodePath(workspaceName)
	if codePath == "" {
		t.Error("code directory should exist")
	}

	testFile := filepath.Join(codePath, "test.txt")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Error("test.txt should exist in code directory")
	}

	// Cleanup
	cmd = exec.Command(binary, "delete", workspaceName)
	cmd.Dir = "/home/ubuntu/nanocodepod/codepod"
	cmd.Output()
}

// TestCLI_WithoutRepo tests workspace without git repo
func TestCLI_WithoutRepo(t *testing.T) {
	binary := getBinaryPath()
	workspaceName := "e2e-norepo"

	// Cleanup
	cmd := exec.Command(binary, "delete", workspaceName)
	cmd.Dir = "/home/ubuntu/nanocodepod/codepod"
	cmd.Output()
	time.Sleep(1 * time.Second)

	// Create workspace without repo
	cmd = exec.Command(binary, "up", workspaceName, "--image", "ubuntu:22.04")
	cmd.Dir = "/home/ubuntu/nanocodepod/codepod"
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("up without repo failed: %v, output: %s", err, string(out))
	}

	time.Sleep(2 * time.Second)

	// Verify workspace still works
	cmd = exec.Command(binary, "list")
	cmd.Dir = "/home/ubuntu/nanocodepod/codepod"
	listOut, _ := cmd.Output()
	if !strings.Contains(string(listOut), workspaceName) {
		t.Error("workspace should be in list")
	}

	// Cleanup
	cmd = exec.Command(binary, "delete", workspaceName)
	cmd.Dir = "/home/ubuntu/nanocodepod/codepod"
	cmd.Output()
}
