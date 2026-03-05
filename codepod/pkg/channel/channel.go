// Package channel provides extensible channel abstractions for CodePod.
// It supports multiple transport types: UDS, SSH, WSL, and StdIO.
package channel

import (
	"context"
	"errors"
	"net"
	"os"
	"os/exec"
	"syscall"
)

// ChannelType represents the type of channel.
type ChannelType string

const (
	ChannelTypeUDS   ChannelType = "uds"
	ChannelTypeSSH   ChannelType = "ssh"
	ChannelTypeWSL   ChannelType = "wsl"
	ChannelTypeStdIO ChannelType = "stdio"
)

// Errors
var (
	ErrConnectionRefused = errors.New("connection refused")
	ErrTimeout          = errors.New("connection timeout")
	ErrInvalidAddress   = errors.New("invalid address")
	ErrNotSupported     = errors.New("channel type not supported")
)

// Config represents channel configuration.
type Config struct {
	// Common fields
	Type ChannelType

	// UDS specific
	SocketPath string

	// SSH specific
	SSHAddr      string
	SSHUser      string
	SSHPassword  string
	SSHKeyPath   string

	// WSL specific
	WSLDistro string

	// StdIO specific
	Command string
	Args    []string
}

// Conn represents a connection.
type Conn interface {
	Read(p []byte) (n int, err error)
	Write(p []byte) (n int, err error)
	Close() error
	LocalAddr() net.Addr
	RemoteAddr() net.Addr
}

// Listener represents a listening endpoint.
type Listener interface {
	Accept() (Conn, error)
	Close() error
	Addr() net.Addr
}

// Channel is the interface for all channel types.
type Channel interface {
	Dial(ctx context.Context, addr string) (Conn, error)
	Listen(addr string) (Listener, error)
	Close() error
}

// channelCreators is a registry of channel creators.
var channelCreators = make(map[ChannelType]func(*Config) Channel)

// RegisterChannel registers a channel type with its creator.
func RegisterChannel(channelType ChannelType, creator func(*Config) Channel) {
	channelCreators[channelType] = creator
}

// NewChannel creates a new channel based on the config.
func NewChannel(cfg *Config) (Channel, error) {
	creator, ok := channelCreators[cfg.Type]
	if !ok {
		return nil, ErrNotSupported
	}
	return creator(cfg), nil
}

// UDSChannel implements Unix Domain Socket transport.
type UDSChannel struct {
	socketPath string
	listener   net.Listener
}

// NewUDSChannel creates a new UDS channel.
func NewUDSChannel(socketPath string) *UDSChannel {
	return &UDSChannel{socketPath: socketPath}
}

// Dial connects to a Unix Domain Socket.
func (c *UDSChannel) Dial(ctx context.Context, addr string) (Conn, error) {
	// Resolve the Unix socket address
	raddr := &net.UnixAddr{Name: addr, Net: "unix"}

	dialer := &net.Dialer{}
	conn, err := dialer.DialContext(ctx, "unix", raddr.Name)
	if err != nil {
		var pathErr *os.PathError
		if errors.As(err, &pathErr) && errors.Is(pathErr.Err, syscall.ENOENT) {
			return nil, ErrConnectionRefused
		}
		return nil, err
	}
	return conn, nil
}

// Listen creates a Unix Domain Socket listener.
func (c *UDSChannel) Listen(addr string) (Listener, error) {
	// Remove existing socket file
	os.Remove(addr)

	laddr := &net.UnixAddr{Name: addr, Net: "unix"}
	listener, err := net.ListenUnix("unix", laddr)
	if err != nil {
		return nil, err
	}
	c.listener = listener
	return &udsListener{listener: listener}, nil
}

// udsListener wraps net.UnixListener to implement Listener.
type udsListener struct {
	listener *net.UnixListener
}

func (l *udsListener) Accept() (Conn, error) {
	conn, err := l.listener.Accept()
	return conn, err
}

func (l *udsListener) Close() error {
	return l.listener.Close()
}

func (l *udsListener) Addr() net.Addr {
	return l.listener.Addr()
}

// Close closes the channel.
func (c *UDSChannel) Close() error {
	if c.listener != nil {
		return c.listener.Close()
	}
	return nil
}

// StdIOChannel implements stdio transport for subprocess communication.
type StdIOChannel struct {
	cmd *exec.Cmd
}

// NewStdIOChannel creates a new StdIO channel.
func NewStdIOChannel(cmd string, args ...string) *StdIOChannel {
	return &StdIOChannel{
		cmd: exec.Command(cmd, args...),
	}
}

// Dial connects via stdio (starts the process).
func (c *StdIOChannel) Dial(ctx context.Context, addr string) (Conn, error) {
	// Note: For stdio, we don't use addr
	// This starts the subprocess
	c.cmd.Stdin = os.Stdin
	c.cmd.Stdout = os.Stdout
	c.cmd.Stderr = os.Stderr

	if err := c.cmd.Start(); err != nil {
		return nil, err
	}

	return &stdioConn{cmd: c.cmd}, nil
}

// Listen is not supported for StdIO.
func (c *StdIOChannel) Listen(addr string) (Listener, error) {
	return nil, ErrNotSupported
}

// Close closes the channel.
func (c *StdIOChannel) Close() error {
	if c.cmd.Process != nil {
		return c.cmd.Process.Kill()
	}
	return nil
}

// stdioConn implements Conn for stdio.
type stdioConn struct {
	cmd *exec.Cmd
}

func (c *stdioConn) Read(p []byte) (n int, err error) {
	// This is a simplified implementation
	// In practice, you'd use pipes
	return 0, errors.New("stdio read not implemented")
}

func (c *stdioConn) Write(p []byte) (n int, err error) {
	// This is a simplified implementation
	// In practice, you'd use pipes
	return 0, errors.New("stdio write not implemented")
}

func (c *stdioConn) Close() error {
	if c.cmd.Process != nil {
		return c.cmd.Process.Kill()
	}
	return nil
}

func (c *stdioConn) LocalAddr() net.Addr {
	return nil
}

func (c *stdioConn) RemoteAddr() net.Addr {
	return nil
}

// SSHChannel implements SSH transport.
type SSHChannel struct {
	addr   string
	user   string
	keyPath string
	password string
}

// NewSSHChannel creates a new SSH channel.
func NewSSHChannel(addr, user, keyPath, password string) *SSHChannel {
	return &SSHChannel{
		addr:      addr,
		user:      user,
		keyPath:   keyPath,
		password:  password,
	}
}

// Dial connects via SSH.
func (c *SSHChannel) Dial(ctx context.Context, addr string) (Conn, error) {
	// Simplified SSH connection - in production would use golang.org/x/crypto/ssh
	// For now, return a placeholder
	return nil, errors.New("SSH dial not implemented - requires golang.org/x/crypto/ssh")
}

// Listen is not supported for SSH.
func (c *SSHChannel) Listen(addr string) (Listener, error) {
	return nil, ErrNotSupported
}

// Close closes the channel.
func (c *SSHChannel) Close() error {
	return nil
}

// WSLChannel implements WSL Interop transport.
type WSLChannel struct {
	distro string
}

// NewWSLChannel creates a new WSL channel.
func NewWSLChannel(distro string) *WSLChannel {
	return &WSLChannel{distro: distro}
}

// Dial connects via WSL Interop.
func (c *WSLChannel) Dial(ctx context.Context, addr string) (Conn, error) {
	// Simplified WSL connection
	return nil, errors.New("WSL dial not implemented")
}

// Listen is not supported for WSL.
func (c *WSLChannel) Listen(addr string) (Listener, error) {
	return nil, ErrNotSupported
}

// Close closes the channel.
func (c *WSLChannel) Close() error {
	return nil
}

// Register built-in channels
func init() {
	RegisterChannel(ChannelTypeUDS, func(cfg *Config) Channel {
		return NewUDSChannel(cfg.SocketPath)
	})
	RegisterChannel(ChannelTypeStdIO, func(cfg *Config) Channel {
		return NewStdIOChannel(cfg.Command, cfg.Args...)
	})
	RegisterChannel(ChannelTypeSSH, func(cfg *Config) Channel {
		return NewSSHChannel(cfg.SSHAddr, cfg.SSHUser, cfg.SSHKeyPath, cfg.SSHPassword)
	})
	RegisterChannel(ChannelTypeWSL, func(cfg *Config) Channel {
		return NewWSLChannel(cfg.WSLDistro)
	})
}
