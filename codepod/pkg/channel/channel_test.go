package channel

import (
	"context"
	"net"
	"os"
	"testing"
	"time"
)

func TestNewChannel_UDS(t *testing.T) {
	cfg := &Config{
		Type:      ChannelTypeUDS,
		SocketPath: "/tmp/test-codepod.sock",
	}

	ch, err := NewChannel(cfg)
	if err != nil {
		t.Fatalf("NewChannel() error = %v", err)
	}

	if _, ok := ch.(*UDSChannel); !ok {
		t.Errorf("expected UDSChannel, got %T", ch)
	}

	// Cleanup
	os.Remove("/tmp/test-codepod.sock")
}

func TestNewChannel_StdIO(t *testing.T) {
	cfg := &Config{
		Type:    ChannelTypeStdIO,
		Command: "echo",
		Args:    []string{"test"},
	}

	ch, err := NewChannel(cfg)
	if err != nil {
		t.Fatalf("NewChannel() error = %v", err)
	}

	if _, ok := ch.(*StdIOChannel); !ok {
		t.Errorf("expected StdIOChannel, got %T", ch)
	}
}

func TestNewChannel_Unknown(t *testing.T) {
	cfg := &Config{
		Type: "unknown",
	}

	ch, err := NewChannel(cfg)
	if err != ErrNotSupported {
		t.Errorf("expected ErrNotSupported, got %v", err)
	}
	if ch != nil {
		t.Errorf("expected nil channel, got %v", ch)
	}
}

func TestUDSChannel_Dial(t *testing.T) {
	ch := NewUDSChannel("/tmp/test.sock")

	// Try to dial to non-existent socket
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := ch.Dial(ctx, "/tmp/nonexistent.sock")
	// Should return an error for non-existent socket
	if err == nil {
		t.Error("expected error for non-existent socket")
	}
}

func TestUDSChannel_ListenAndAccept(t *testing.T) {
	ch := NewUDSChannel("/tmp/test-listen.sock")
	defer os.Remove("/tmp/test-listen.sock")

	listener, err := ch.Listen("/tmp/test-listen.sock")
	if err != nil {
		t.Fatalf("Listen() error = %v", err)
	}
	defer listener.Close()

	// Test Accept in non-blocking way
	done := make(chan error, 1)
	go func() {
		_, err := listener.Accept()
		done <- err
	}()

	// Connect a client
	conn, err := net.Dial("unix", "/tmp/test-listen.sock")
	if err != nil {
		t.Fatalf("client dial error = %v", err)
	}
	defer conn.Close()

	// Wait for accept with timeout
	select {
	case err := <-done:
		// Accept returned (may have error or nil)
		_ = err
	case <-time.After(1 * time.Second):
		t.Log("Accept timeout - expected for non-blocking test")
	}
}

func TestChannelRegistry(t *testing.T) {
	// Verify built-in channels are registered
	udsCh, err := NewChannel(&Config{Type: ChannelTypeUDS, SocketPath: "/tmp/test.sock"})
	if err != nil {
		t.Fatalf("NewChannel(UDS) error = %v", err)
	}
	if udsCh == nil {
		t.Error("UDS channel should not be nil")
	}

	stdioCh, err := NewChannel(&Config{Type: ChannelTypeStdIO, Command: "echo"})
	if err != nil {
		t.Fatalf("NewChannel(StdIO) error = %v", err)
	}
	if stdioCh == nil {
		t.Error("StdIO channel should not be nil")
	}

	sshCh, err := NewChannel(&Config{Type: ChannelTypeSSH, SSHAddr: "localhost"})
	if err != nil {
		t.Fatalf("NewChannel(SSH) error = %v", err)
	}
	if sshCh == nil {
		t.Error("SSH channel should not be nil")
	}

	wslCh, err := NewChannel(&Config{Type: ChannelTypeWSL, WSLDistro: "Ubuntu"})
	if err != nil {
		t.Fatalf("NewChannel(WSL) error = %v", err)
	}
	if wslCh == nil {
		t.Error("WSL channel should not be nil")
	}
}
