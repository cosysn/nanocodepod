// Package channel provides extensible channel abstractions for CodePod.
// It supports multiple transport types: UDS, SSH, WSL, and StdIO.
package channel

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"golang.org/x/crypto/ssh"
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
	SetDeadline(t time.Time) error
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

// SetDeadline sets the accept deadline.
func (l *udsListener) SetDeadline(t time.Time) error {
	return l.listener.SetDeadline(t)
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
	addr      string
	user      string
	keyPath   string
	password  string
	client    *ssh.Client
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
	var auth []ssh.AuthMethod

	if c.password != "" {
		auth = append(auth, ssh.Password(c.password))
	}

	if c.keyPath != "" {
		key, err := os.ReadFile(c.keyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read SSH key: %w", err)
		}
		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return nil, fmt.Errorf("failed to parse SSH key: %w", err)
		}
		auth = append(auth, ssh.PublicKeys(signer))
	}

	if len(auth) == 0 {
		return nil, errors.New("no authentication method provided for SSH")
	}

	cfg := &ssh.ClientConfig{
		User:            c.user,
		Auth:            auth,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	client, err := ssh.Dial("tcp", addr, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to dial SSH: %w", err)
	}

	c.client = client
	return &sshConn{client: client}, nil
}

// Exec executes a command on the remote host.
func (c *SSHChannel) Exec(cmd string) (io.ReadCloser, error) {
	if c.client == nil {
		return nil, errors.New("not connected")
	}

	session, err := c.client.NewSession()
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	stdout, err := session.StdoutPipe()
	if err != nil {
		session.Close()
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	if err := session.Start(cmd); err != nil {
		session.Close()
		return nil, fmt.Errorf("failed to start command: %w", err)
	}

	// Return a ReadCloser that also closes the session
	return &sshReadCloser{Reader: stdout, Session: session}, nil
}

// sshReadCloser wraps ssh.Session stdout to provide io.ReadCloser.
type sshReadCloser struct {
	io.Reader
	Session *ssh.Session
}

func (r *sshReadCloser) Close() error {
	return r.Session.Close()
}

// Session creates a new SSH session for interactive use.
func (c *SSHChannel) Session() (*ssh.Session, error) {
	if c.client == nil {
		return nil, errors.New("not connected")
	}
	return c.client.NewSession()
}

// Listen is not supported for SSH.
func (c *SSHChannel) Listen(addr string) (Listener, error) {
	return nil, ErrNotSupported
}

// Close closes the channel.
func (c *SSHChannel) Close() error {
	if c.client != nil {
		c.client.Close()
	}
	return nil
}

// sshConn wraps ssh.Client to implement Conn interface.
type sshConn struct {
	client *ssh.Client
}

func (c *sshConn) Read(p []byte) (n int, err error) {
	// SSH doesn't provide direct read/write - use sessions
	return 0, errors.New("use Session() for I/O")
}

func (c *sshConn) Write(p []byte) (n int, err error) {
	return 0, errors.New("use Session() for I/O")
}

func (c *sshConn) Close() error {
	return c.client.Close()
}

func (c *sshConn) LocalAddr() net.Addr {
	return c.client.LocalAddr()
}

func (c *sshConn) RemoteAddr() net.Addr {
	return c.client.RemoteAddr()
}

// WSLChannel implements WSL Interop transport.
type WSLChannel struct {
	distro string
	cmd    *exec.Cmd
}

// NewWSLChannel creates a new WSL channel.
func NewWSLChannel(distro string) *WSLChannel {
	return &WSLChannel{distro: distro}
}

// Dial connects via WSL Interop.
// On Windows, uses wsl.exe to execute commands in the WSL distribution.
func (c *WSLChannel) Dial(ctx context.Context, addr string) (Conn, error) {
	// For WSL, we spawn a process that connects to the specified socket/address
	// Use wsl.exe to run commands in the WSL distribution
	distro := c.distro
	if distro == "" {
		distro = "Ubuntu" // default
	}

	// Build wsl command to execute
	cmd := exec.Command("wsl.exe", "-d", distro, "--", "bash", "-c", addr)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		// Set up for WSL interop if available
	}

	c.cmd = cmd
	return &wslConn{cmd: cmd}, nil
}

// Exec executes a command in WSL.
func (c *WSLChannel) Exec(cmd string) (io.ReadCloser, error) {
	distro := c.distro
	if distro == "" {
		distro = "Ubuntu"
	}

	execCmd := exec.Command("wsl.exe", "-d", distro, "--", "bash", "-c", cmd)
	stdout, err := execCmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	if err := execCmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start WSL command: %w", err)
	}

	// Return a ReadCloser that also closes the command
	return &wslReadCloser{Reader: stdout, Cmd: execCmd}, nil
}

// wslReadCloser wraps exec.Cmd stdout to provide io.ReadCloser.
type wslReadCloser struct {
	io.Reader
	Cmd *exec.Cmd
}

func (r *wslReadCloser) Close() error {
	if r.Cmd.Process != nil {
		r.Cmd.Process.Kill()
	}
	return r.Cmd.Wait()
}

// Distribution returns the default WSL distribution name.
func (c *WSLChannel) Distribution() string {
	return c.distro
}

// ListDistributions lists available WSL distributions.
func ListDistributions() ([]string, error) {
	cmd := exec.Command("wsl.exe", "--list", "--quiet")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list WSL distributions: %w", err)
	}

	var distros []string
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "NAME") {
			distros = append(distros, line)
		}
	}
	return distros, nil
}

// Listen is not supported for WSL.
func (c *WSLChannel) Listen(addr string) (Listener, error) {
	return nil, ErrNotSupported
}

// Close closes the channel.
func (c *WSLChannel) Close() error {
	if c.cmd != nil && c.cmd.Process != nil {
		c.cmd.Process.Kill()
	}
	return nil
}

// wslConn wraps exec.Cmd to implement Conn interface.
type wslConn struct {
	cmd *exec.Cmd
}

func (c *wslConn) Read(p []byte) (n int, err error) {
	return 0, errors.New("use Exec() for command execution")
}

func (c *wslConn) Write(p []byte) (n int, err error) {
	return 0, errors.New("use Exec() for command execution")
}

func (c *wslConn) Close() error {
	if c.cmd != nil && c.cmd.Process != nil {
		return c.cmd.Process.Kill()
	}
	return nil
}

func (c *wslConn) LocalAddr() net.Addr {
	return nil
}

func (c *wslConn) RemoteAddr() net.Addr {
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
