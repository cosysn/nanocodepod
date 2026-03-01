package devcon

import (
	"os"
	"path/filepath"
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
