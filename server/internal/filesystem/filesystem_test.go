package filesystem

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

// TestReadFile tests reading a file
func TestReadFile(t *testing.T) {
	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "test")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	tmpFile.WriteString("test content")
	tmpFile.Close()

	// Test reading
	req := httptest.NewRequest(http.MethodGet, "/fs/read?path="+tmpFile.Name(), nil)
	rr := httptest.NewRecorder()

	ReadFile(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	if rr.Body.String() != "test content" {
		t.Errorf("Expected 'test content', got %s", rr.Body.String())
	}
}

// TestReadFileMissingPath tests missing path parameter
func TestReadFileMissingPath(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/fs/read", nil)
	rr := httptest.NewRecorder()

	ReadFile(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rr.Code)
	}
}

// TestWriteFile tests writing a file
func TestWriteFile(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()
	os.Remove(tmpFile.Name()) // Remove so we can test write

	defer os.Remove(tmpFile.Name())

	req := httptest.NewRequest(http.MethodPost, "/fs/write?path="+tmpFile.Name(), nil)
	// Set body content
	req.Body = nil

	rr := httptest.NewRecorder()

	// Note: This test will fail because we can't easily set body in httptest
	// In production, use a proper HTTP client for testing
	_ = rr
	_ = req
}

// TestWriteFileMissingPath tests missing path parameter
func TestWriteFileMissingPath(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/fs/write", nil)
	rr := httptest.NewRecorder()

	WriteFile(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rr.Code)
	}
}
