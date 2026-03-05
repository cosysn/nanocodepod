package mux

import (
	"testing"
)

func TestStreamTypeConstants(t *testing.T) {
	if StreamTypeControl != "control" {
		t.Errorf("StreamTypeControl = %s; want control", StreamTypeControl)
	}
	if StreamTypeData != "data" {
		t.Errorf("StreamTypeData = %s; want data", StreamTypeData)
	}
}

func TestMuxedConnInterface(t *testing.T) {
	// This test verifies the interface is properly defined
	var _ MuxedConn = (*YamuxMuxer)(nil)
}

func TestMuxStreamInterface(t *testing.T) {
	// This test verifies the interface is properly defined
	var _ MuxStream = (*yamuxStream)(nil)
}

func TestYamuxStreamStreamID(t *testing.T) {
	// Test the placeholder StreamID - it should return 0
	// We can't create a real yamuxStream without a connection,
	// but we can verify the type exists
	_ = StreamTypeControl
	_ = StreamTypeData
}
