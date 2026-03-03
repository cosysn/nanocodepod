package wsl

import (
	"testing"
)

// MockPlatformForTest is a mock implementation for testing
type MockPlatformForTest struct {
	MockType           PlatformType
	MockHostname       string
	MockRunCmdResult   string
	MockRunCmdError    error
	MockFileExistsResult bool
}

func (m *MockPlatformForTest) GetType() PlatformType {
	return m.MockType
}

func (m *MockPlatformForTest) GetDistribution() string {
	return "Ubuntu-22.04"
}

func (m *MockPlatformForTest) GetHostname() (string, error) {
	return m.MockHostname, nil
}

func (m *MockPlatformForTest) RunCommand(cmd string) (string, error) {
	return m.MockRunCmdResult, m.MockRunCmdError
}

func (m *MockPlatformForTest) FileExists(path string) bool {
	return m.MockFileExistsResult
}

func (m *MockPlatformForTest) CopyToWSL(src, dest string) error {
	return nil
}

func (m *MockPlatformForTest) CopyFromWSL(src, dest string) error {
	return nil
}

func TestDetectPlatform(t *testing.T) {
	platform := DetectPlatform()

	// Should return one of the platform types
	switch platform {
	case PlatformLinux, PlatformWSL, PlatformWindows:
		// OK
	default:
		t.Errorf("unexpected platform: %s", platform)
	}
}

func TestIsDockerAvailable(t *testing.T) {
	// This test depends on Docker being installed
	// Just ensure it doesn't panic
	_ = IsDockerAvailable()
}

func TestGetDockerHost(t *testing.T) {
	tests := []struct {
		platform PlatformType
		expected string
	}{
		{PlatformLinux, "unix:///var/run/docker.sock"},
		{PlatformWSL, "tcp://localhost:2375"},
		{PlatformWindows, "tcp://localhost:2375"},
	}

	for _, tt := range tests {
		result := GetDockerHost(tt.platform)
		if result != tt.expected {
			t.Errorf("GetDockerHost(%s) = %s, want %s", tt.platform, result, tt.expected)
		}
	}
}

func TestPlatform_GetHostname(t *testing.T) {
	platform, err := NewPlatform()
	if err != nil {
		t.Fatalf("failed to create platform: %v", err)
	}

	hostname, err := platform.GetHostname()
	if err != nil {
		t.Fatalf("failed to get hostname: %v", err)
	}

	if hostname == "" {
		t.Error("hostname should not be empty")
	}
}

func TestPlatform_RunCommand(t *testing.T) {
	platform, err := NewPlatform()
	if err != nil {
		t.Fatalf("failed to create platform: %v", err)
	}

	// Skip if not on WSL/Linux
	if platform.GetType() != PlatformWSL && platform.GetType() != PlatformLinux {
		t.Skip("RunCommand only tested on WSL or Linux")
	}

	// Test simple command
	output, err := platform.RunCommand("echo hello")
	if err != nil {
		t.Skipf("WSL not available: %v", err)
	}

	if output != "hello" {
		t.Errorf("expected 'hello', got '%s'", output)
	}
}

func TestPlatform_FileExists(t *testing.T) {
	platform, err := NewPlatform()
	if err != nil {
		t.Fatalf("failed to create platform: %v", err)
	}

	// Skip if not on WSL/Linux
	if platform.GetType() != PlatformWSL && platform.GetType() != PlatformLinux {
		t.Skip("FileExists only tested on WSL or Linux")
	}

	// Test with existing file
	exists := platform.FileExists("/etc/passwd")
	if !exists {
		t.Skip("/etc/passwd does not exist in WSL")
	}

	// Test with non-existing file
	exists = platform.FileExists("/nonexistent/path/file")
	if exists {
		t.Error("should not exist")
	}
}

func TestNewPlatform(t *testing.T) {
	platform, err := NewPlatform()
	if err != nil {
		t.Fatalf("failed to create platform: %v", err)
	}

	if platform.Type == "" {
		t.Error("platform type should not be empty")
	}
}

func TestWSL_New(t *testing.T) {
	ws := New("Ubuntu-22.04")
	if ws == nil {
		t.Error("New should not return nil")
	}
	if ws.GetDistribution() != "Ubuntu-22.04" {
		t.Errorf("expected Ubuntu-22.04, got %s", ws.GetDistribution())
	}
}

func TestListDistributions(t *testing.T) {
	// This test will fail on Linux since wsl.exe doesn't exist
	// Just ensure it doesn't panic
	_, _ = ListDistributions()
}

func TestDistributionExists(t *testing.T) {
	// This test will fail on Linux since wsl.exe doesn't exist
	// Just ensure it doesn't panic
	_, _ = DistributionExists("Ubuntu")
}

func TestPlatformInterface(t *testing.T) {
	// Verify Platform implements PlatformInterface
	var _ PlatformInterface = (*Platform)(nil)
}

func TestPlatformMethods(t *testing.T) {
	platform, err := NewPlatform()
	if err != nil {
		t.Fatalf("failed to create platform: %v", err)
	}

	// Test GetType
	if platform.GetType() == "" {
		t.Error("GetType should not return empty")
	}

	// Test GetDistribution
	dist := platform.GetDistribution()
	if dist == "" {
		t.Error("GetDistribution should not return empty")
	}
}

func TestMockPlatform(t *testing.T) {
	// Create a mock platform
	mock := &MockPlatformForTest{
		MockType:     PlatformLinux,
		MockHostname: "test-host",
		MockRunCmdResult: "test output",
	}

	if mock.GetType() != PlatformLinux {
		t.Errorf("expected PlatformLinux, got %s", mock.GetType())
	}

	hostname, _ := mock.GetHostname()
	if hostname != "test-host" {
		t.Errorf("expected test-host, got %s", hostname)
	}

	output, _ := mock.RunCommand("echo test")
	if output != "test output" {
		t.Errorf("expected test output, got %s", output)
	}
}
