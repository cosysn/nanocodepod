// Package mux provides Yamux stream multiplexing integration.
package mux

import (
	"io"
	"net"

	"github.com/hashicorp/yamux"
)

// MuxedConn is a connection that supports stream multiplexing.
type MuxedConn interface {
	net.Conn
	Stream() (MuxStream, error)
}

// MuxStream represents a multiplexed stream.
type MuxStream interface {
	io.ReadWriteCloser
	StreamID() uint32
}

// YamuxMuxer wraps a net.Conn with Yamux multiplexing.
type YamuxMuxer struct {
	session *yamux.Session
	conn    net.Conn
}

// NewYamuxMuxer creates a new Yamux muxer from a connection.
func NewYamuxMuxer(conn net.Conn, isServer bool) (*YamuxMuxer, error) {
	var session *yamux.Session
	var err error

	config := yamux.DefaultConfig()
	config.EnableKeepAlive = true

	if isServer {
		session, err = yamux.Server(conn, config)
	} else {
		session, err = yamux.Client(conn, config)
	}

	if err != nil {
		return nil, err
	}

	return &YamuxMuxer{
		session: session,
		conn:    conn,
	}, nil
}

// Stream creates a new multiplexed stream.
func (m *YamuxMuxer) Stream() (MuxStream, error) {
	stream, err := m.session.Open()
	if err != nil {
		return nil, err
	}
	return &yamuxStream{stream: stream}, nil
}

// AcceptStream accepts a new stream from the remote.
func (m *YamuxMuxer) AcceptStream() (MuxStream, error) {
	stream, err := m.session.Accept()
	if err != nil {
		return nil, err
	}
	return &yamuxStream{stream: stream}, nil
}

// Close closes the muxer and underlying connection.
func (m *YamuxMuxer) Close() error {
	if m.session != nil {
		m.session.Close()
	}
	return m.conn.Close()
}

// LocalAddr returns the local network address.
func (m *YamuxMuxer) LocalAddr() net.Addr {
	return m.conn.LocalAddr()
}

// RemoteAddr returns the remote network address.
func (m *YamuxMuxer) RemoteAddr() net.Addr {
	return m.conn.RemoteAddr()
}

// yamuxStream wraps yamux.Stream to implement MuxStream.
type yamuxStream struct {
	stream net.Conn
}

func (s *yamuxStream) Read(p []byte) (n int, err error) {
	return s.stream.Read(p)
}

func (s *yamuxStream) Write(p []byte) (n int, err error) {
	return s.stream.Write(p)
}

func (s *yamuxStream) Close() error {
	return s.stream.Close()
}

func (s *yamuxStream) StreamID() uint32 {
	// Yamux doesn't expose StreamID directly, return 0 as placeholder
	return 0
}

// StreamType represents the type of stream.
type StreamType string

const (
	StreamTypeControl StreamType = "control" // RPC
	StreamTypeData    StreamType = "data"    // PTY, FS
)
