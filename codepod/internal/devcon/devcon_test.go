package devcon

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildOptions(t *testing.T) {
	opts := &BuildOptions{
		WorkspacePath: "/path/to/workspace",
		ImageTag:      "myimage:latest",
		RepositoryURL: "https://github.com/user/repo",
		Dockerfile:   "/path/to/Dockerfile",
		Context:       "/path/to/context",
	}

	if opts.ImageTag != "myimage:latest" {
		t.Errorf("expected ImageTag myimage:latest, got %s", opts.ImageTag)
	}
	if opts.Dockerfile != "/path/to/Dockerfile" {
		t.Errorf("expected Dockerfile /path/to/Dockerfile, got %s", opts.Dockerfile)
	}
}

func TestBuildWithDockerfile_NoDockerfile(t *testing.T) {
	d := &Devcon{}

	opts := &BuildOptions{
		WorkspacePath: "/tmp",
		ImageTag:      "test:latest",
	}

	_, err := d.BuildWithDockerfile(opts)
	if err == nil {
		t.Error("expected error for missing Dockerfile")
	}
}

func TestHasDockerfile(t *testing.T) {
	d := &Devcon{}

	// Create temp dir with Dockerfile
	tmpDir := t.TempDir()
	dockerfilePath := filepath.Join(tmpDir, "Dockerfile")
	os.WriteFile(dockerfilePath, []byte("FROM ubuntu\n"), 0644)

	if !d.HasDockerfile(tmpDir) {
		t.Error("should detect Dockerfile exists")
	}

	// Test non-existing
	if d.HasDockerfile("/nonexistent") {
		t.Error("should return false for non-existent path")
	}
}

func TestGetDockerfilePath(t *testing.T) {
	d := &Devcon{}

	path := d.GetDockerfilePath("/workspace/myproject")
	expected := "/workspace/myproject/Dockerfile"

	if path != expected {
		t.Errorf("expected %s, got %s", expected, path)
	}
}

func TestHasDevfile(t *testing.T) {
	d := &Devcon{}

	// Create temp dir with devfile
	tmpDir := t.TempDir()
	devfilePath := filepath.Join(tmpDir, "devfile.yaml")
	os.WriteFile(devfilePath, []byte("schemaVersion: 2.0.0\n"), 0644)

	if !d.HasDevfile(tmpDir) {
		t.Error("should detect devfile exists")
	}

	// Test non-existing
	if d.HasDevfile("/nonexistent") {
		t.Error("should return false for non-existent path")
	}
}

func TestHasDevcontainer(t *testing.T) {
	d := &Devcon{}

	// Create temp dir with .devcontainer.json
	tmpDir := t.TempDir()
	devcontainerPath := filepath.Join(tmpDir, ".devcontainer.json")
	os.WriteFile(devcontainerPath, []byte("{\n}\n"), 0644)

	if !d.HasDevcontainer(tmpDir) {
		t.Error("should detect devcontainer exists")
	}

	// Test non-existing
	if d.HasDevcontainer("/nonexistent") {
		t.Error("should return false for non-existent path")
	}
}

func TestExtractImageName(t *testing.T) {
	d := &Devcon{}

	tests := []struct {
		output    string
		fallback  string
		expected  string
	}{
		{"Image built: myimage:latest\n", "", "myimage:latest"},
		{"Some output\nImage built: custom:tag\n", "", "custom:tag"},
		{"No image line here", "fallback:tag", "fallback:tag"},
		{"", "", "devcontainer:latest"},
	}

	for _, tt := range tests {
		result := d.extractImageName(tt.output, tt.fallback)
		if result != tt.expected {
			t.Errorf("extractImageName(%q, %q) = %q, want %q", tt.output, tt.fallback, result, tt.expected)
		}
	}
}

func TestGetDevcontainerConfig(t *testing.T) {
	d := &Devcon{}

	// Create temp dir with .devcontainer.json
	tmpDir := t.TempDir()
	devcontainerPath := filepath.Join(tmpDir, ".devcontainer.json")
	content := `{"name": "test", "image": "ubuntu:latest"}`
	os.WriteFile(devcontainerPath, []byte(content), 0644)

	config, err := d.GetDevcontainerConfig(tmpDir)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !strings.Contains(config, "ubuntu:latest") {
		t.Error("expected config to contain ubuntu:latest")
	}

	// Test non-existing
	_, err = d.GetDevcontainerConfig("/nonexistent")
	if err == nil {
		t.Error("expected error for non-existent path")
	}
}
