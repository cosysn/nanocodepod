package tui

import (
	"testing"
)

func TestPrintSuccess(t *testing.T) {
	// Just ensure it doesn't panic
	PrintSuccess("test message")
}

func TestPrintError(t *testing.T) {
	// Just ensure it doesn't panic
	PrintError("test error")
}

func TestPrintWarning(t *testing.T) {
	// Just ensure it doesn't panic
	PrintWarning("test warning")
}

func TestPrintInfo(t *testing.T) {
	// Just ensure it doesn't panic
	PrintInfo("test info")
}

func TestPrintDim(t *testing.T) {
	// Just ensure it doesn't panic
	PrintDim("test dim")
}

func TestProgressBar(t *testing.T) {
	result := ProgressBar("Test", 50, 100)

	if result == "" {
		t.Error("should return non-empty string")
	}
}

func TestSpinner(t *testing.T) {
	frame0 := Spinner(0)
	frame10 := Spinner(10)

	if frame0 == "" {
		t.Error("should return non-empty string")
	}

	// Should wrap around
	if frame0 != frame10 {
		t.Error("spinner should wrap around")
	}
}

func TestConfirm(t *testing.T) {
	// This will wait for input, so we just ensure it doesn't panic
	// In actual use, you'd mock this
	_ = Confirm("Test prompt")
}
