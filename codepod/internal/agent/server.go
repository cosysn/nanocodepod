package agent

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/codepod-io/codepod/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"golang.org/x/crypto/ssh"
)

// HTTP/2 magic bytes for gRPC protocol detection
var grpcMagicBytes = []byte("PRI * HTTP/2.0\r\n\r\nSM\r\n\r\n")

// Server represents an SSH server for the agent
type Server struct {
	port           int
	config         *ServerConfig
	listener       net.Listener
	grpcServer     *grpc.Server
	startTime      time.Time
	activeCommands int32
}

// GrpcServer wraps the gRPC server for the agent
type GrpcServer struct {
	proto.UnimplementedAgentServer
	port    int
	server  *grpc.Server
	startTime time.Time
}

// ServerConfig holds SSH server configuration
type ServerConfig struct {
	HostKeyPath string
	Port        int
	User        string
	Password    string
}

// NewServer creates a new SSH server
func NewServer(config *ServerConfig) *Server {
	return &Server{
		port:   config.Port,
		config: config,
	}
}

// Start starts the SSH server
func (s *Server) Start() error {
	// Generate host keys if not exist
	if err := s.ensureHostKeys(); err != nil {
		return fmt.Errorf("failed to ensure host keys: %w", err)
	}

	// Create SSH server configuration
	serverConfig := &ssh.ServerConfig{
		PasswordCallback: s.passwordAuth,
	}

	// Load host key
	privateKey, err := os.ReadFile(s.config.HostKeyPath)
	if err != nil {
		return fmt.Errorf("failed to read host key: %w", err)
	}

	signer, err := ssh.ParsePrivateKey(privateKey)
	if err != nil {
		return fmt.Errorf("failed to parse host key: %w", err)
	}
	serverConfig.AddHostKey(signer)

	// Start listening
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return fmt.Errorf("failed to listen on port %d: %w", s.port, err)
	}
	s.listener = listener

	// Accept connections
	go s.acceptConnections(serverConfig)

	return nil
}

// Stop stops the SSH server
func (s *Server) Stop() error {
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}

func (s *Server) acceptConnections(config *ssh.ServerConfig) {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			break
		}

		go s.handleConnection(conn, config)
	}
}

func (s *Server) handleConnection(conn net.Conn, config *ssh.ServerConfig) {
	defer conn.Close()

	_, chans, reqs, err := ssh.NewServerConn(conn, config)
	if err != nil {
		return
	}

	go ssh.DiscardRequests(reqs)

	for newChannel := range chans {
		if newChannel.ChannelType() != "session" {
			newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
			continue
		}

		channel, requests, err := newChannel.Accept()
		if err != nil {
			continue
		}

		go s.handleSession(channel, requests)
	}
}

func (s *Server) handleSession(channel ssh.Channel, requests <-chan *ssh.Request) {
	defer channel.Close()

	for req := range requests {
		switch req.Type {
		case "exec":
			// Execute command
			cmd := string(req.Payload[4:])
			s.executeCommand(cmd, channel)
			req.Reply(true, nil)
		default:
			req.Reply(false, nil)
		}
	}
}

func (s *Server) executeCommand(cmd string, channel ssh.Channel) {
	command := exec.Command("bash", "-c", cmd)
	command.Stdout = channel
	command.Stderr = channel.Stderr()
	command.Stdin = channel

	command.Run()
}

func (s *Server) passwordAuth(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
	if string(pass) == s.config.Password {
		return &ssh.Permissions{}, nil
	}
	return nil, fmt.Errorf("invalid password")
}

func (s *Server) ensureHostKeys() error {
	dir := filepath.Dir(s.config.HostKeyPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	if _, err := os.Stat(s.config.HostKeyPath); os.IsNotExist(err) {
		// Generate new host key
		cmd := exec.Command("ssh-keygen", "-t", "rsa", "-b", "4096", "-f", s.config.HostKeyPath, "-N", "")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to generate host key: %w", err)
		}
	}

	return nil
}

// MuxListener wraps a net.Listener to detect SSH vs gRPC protocol
type MuxListener struct {
	net.Listener
	grpcServer *grpc.Server
	sshServer  *Server
}

// NewMuxListener creates a new mux listener that routes connections based on protocol
func NewMuxListener(port int, sshServ *Server, grpcServ *grpc.Server) *MuxListener {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil
	}

	return &MuxListener{
		Listener:   lis,
		grpcServer: grpcServ,
		sshServer:  sshServ,
	}
}

// Accept implements Accept method with protocol detection
func (m *MuxListener) Accept() (net.Conn, error) {
	conn, err := m.Listener.Accept()
	if err != nil {
		return nil, err
	}

	// Peek at the first few bytes to detect protocol
	peekBuf := make([]byte, len(grpcMagicBytes))
	n, err := io.ReadFull(conn, peekBuf)
	if err != nil || n < len(grpcMagicBytes) {
		// Not enough data to detect, treat as SSH
		return conn, nil
	}

	// Check for gRPC HTTP/2 magic bytes
	if string(peekBuf) == string(grpcMagicBytes) {
		// This is a gRPC connection - would need to use a different approach
		// For now, we'll handle this in a connection wrapper
		return &grpcConnWrapper{conn, peekBuf}, nil
	}

	// SSH connection - put the peeked bytes back
	conn = &sshConnWrapper{conn, peekBuf}
	return conn, nil
}

// grpcConnWrapper wraps a connection for gRPC
type grpcConnWrapper struct {
	net.Conn
	peekBuf []byte
}

// Read reads data, prepending the peeked bytes on first read
func (g *grpcConnWrapper) Read(b []byte) (int, error) {
	if len(g.peekBuf) > 0 {
		n := copy(b, g.peekBuf)
		g.peekBuf = g.peekBuf[n:]
		return n, nil
	}
	return g.Conn.Read(b)
}

// sshConnWrapper wraps a connection for SSH with buffered peek bytes
type sshConnWrapper struct {
	net.Conn
	peekBuf []byte
}

// Read reads data, prepending the peeked bytes on first read
func (s *sshConnWrapper) Read(b []byte) (int, error) {
	if len(s.peekBuf) > 0 {
		n := copy(b, s.peekBuf)
		s.peekBuf = s.peekBuf[n:]
		return n, nil
	}
	return s.Conn.Read(b)
}

// NewGrpcServer creates a new gRPC server
func NewGrpcServer(port int) *GrpcServer {
	return &GrpcServer{
		port:      port,
		startTime: time.Now(),
	}
}

// Start starts the gRPC server
func (g *GrpcServer) Start() error {
	// Listen on the same port as SSH (will use protocol detection)
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", g.port))
	if err != nil {
		return fmt.Errorf("failed to listen on port %d: %w", g.port, err)
	}

	g.server = grpc.NewServer()
	proto.RegisterAgentServer(g.server, g)

	go func() {
		if err := g.server.Serve(lis); err != nil {
			fmt.Printf("gRPC server error: %v\n", err)
		}
	}()

	return nil
}

// Stop stops the gRPC server
func (g *GrpcServer) Stop() {
	if g.server != nil {
		g.server.GracefulStop()
	}
}

// ExecuteCommand executes a command via gRPC
func (g *GrpcServer) ExecuteCommand(ctx context.Context, req *proto.CommandRequest) (*proto.CommandResponse, error) {
	// Check for password in metadata
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok || len(md["password"]) == 0 {
		return nil, fmt.Errorf("unauthorized: missing password")
	}

	// Build command
	command := exec.Command("bash", "-c", req.Command)
	command.Dir = req.WorkingDir
	if req.WorkingDir == "" {
		command.Dir = "/workspace"
	}

	// Set environment variables
	for k, v := range req.Env {
		command.Env = append(command.Env, fmt.Sprintf("%s=%s", k, v))
	}

	output, err := command.CombinedOutput()
	exitCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			exitCode = 1
		}
	}

	return &proto.CommandResponse{
		ExitCode: int32(exitCode),
		Output:   string(output),
	}, nil
}

// GetStatus returns the agent status
func (g *GrpcServer) GetStatus(ctx context.Context, req *proto.StatusRequest) (*proto.StatusResponse, error) {
	return &proto.StatusResponse{
		Status:            "running",
		Pid:               int32(os.Getpid()),
		UptimeSeconds:     int64(time.Since(g.startTime).Seconds()),
		ActiveConnections: 0,
	}, nil
}

// StreamCommand streams command output via gRPC
func (g *GrpcServer) StreamCommand(req *proto.CommandRequest, srv proto.Agent_StreamCommandServer) error {
	command := exec.Command("bash", "-c", req.Command)
	command.Dir = req.WorkingDir
	if req.WorkingDir == "" {
		command.Dir = "/workspace"
	}

	stdout, err := command.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := command.StderrPipe()
	if err != nil {
		return err
	}

	if err := command.Start(); err != nil {
		return err
	}

	// Stream stdout
	buf := make([]byte, 1024)
	for {
		n, err := stdout.Read(buf)
		if n > 0 {
			srv.Send(&proto.CommandOutput{Data: buf[:n], Stderr: false})
		}
		if err != nil {
			break
		}
	}

	// Stream stderr
	for {
		n, err := stderr.Read(buf)
		if n > 0 {
			srv.Send(&proto.CommandOutput{Data: buf[:n], Stderr: true})
		}
		if err != nil {
			break
		}
	}

	command.Wait()
	return nil
}

// RunAgent runs the agent with SSH server
func RunAgent(port int, password string) error {
	config := &ServerConfig{
		HostKeyPath: "/etc/codepod/ssh_host_rsa_key",
		Port:        port,
		User:        "root",
		Password:    password,
	}

	// Start zombie reaper (for running as PID 0/init process)
	go reapZombies()

	// Handle shutdown signals
	go handleSignals()

	// Start SSH server
	server := NewServer(config)
	if err := server.Start(); err != nil {
		return err
	}
	fmt.Println("Agent SSH server started")

	// Start gRPC server on same port (protocol detection will route connections)
	grpcServer := NewGrpcServer(port)
	if err := grpcServer.Start(); err != nil {
		fmt.Printf("Warning: failed to start gRPC server: %v\n", err)
	} else {
		fmt.Println("Agent gRPC server started")
	}

	// Wait forever (agent runs as PID 0/init process)
	select {}
}

// reapZombies reaps zombie processes
// This is essential when running as PID 0 (init process)
func reapZombies() {
	for {
		var status syscall.WaitStatus
		pid, err := syscall.Wait4(-1, &status, syscall.WNOHANG, nil)
		if err != nil {
			// No more children to reap, sleep briefly and try again
		}
		if pid <= 0 {
			// Use simple sleep to avoid busy loop
			fmt.Println("Waiting for child processes...")
		}
	}
}

// handleSignals handles shutdown signals for graceful termination
func handleSignals() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)

	for {
		sig := <-sigChan
		fmt.Printf("Received signal: %v\n", sig)
		// Cleanup and exit
		os.Exit(0)
	}
}

// CopyToContainer copies agent binary to container
func CopyToContainer(containerID, agentPath string) error {
	cmd := exec.Command("docker", "cp", agentPath, fmt.Sprintf("%s:/usr/local/bin/codepod-agent", containerID))
	return cmd.Run()
}

// AgentExistsInContainer checks if agent exists in container
func AgentExistsInContainer(containerID string) bool {
	cmd := exec.Command("docker", "exec", containerID, "test", "-f", "/usr/local/bin/codepod-agent")
	return cmd.Run() == nil
}

// StartAgentInContainer starts the agent in container
func StartAgentInContainer(containerID string, port int, password string) error {
	cmd := exec.Command("docker", "exec", "-d", containerID, "/usr/local/bin/codepod-agent", "--port", fmt.Sprintf("%d", port), "--password", password)
	return cmd.Run()
}

// GetAgentLogs gets agent logs from container
func GetAgentLogs(containerID string) (string, error) {
	cmd := exec.Command("docker", "logs", containerID)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}
