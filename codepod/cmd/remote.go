// Package cmd provides CLI commands for CodePod.
package cmd

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"github.com/codepod-io/codepod/pkg/agent"
	"github.com/codepod-io/codepod/pkg/bootstrapper"
	"github.com/codepod-io/codepod/pkg/channel"
	"github.com/codepod-io/codepod/pkg/provider"
	"github.com/codepod-io/codepod/pkg/resolver"
	"github.com/codepod-io/codepod/pkg/rpc"
	"github.com/spf13/cobra"
)

const (
	defaultSocketPath = "/tmp/codepod.sock"
	defaultPipeName   = "codepod"
)

// remoteCmd represents the remote connection command
var remoteCmd = &cobra.Command{
	Use:   "remote <uri>",
	Short: "Connect to a remote environment via URI",
	Long: `Connect to a remote environment using codepod-remote:// URI format.

Examples:
  codepod remote codepod-remote://wsl+ubuntu/home/user
  codepod remote codepod-remote://ssh-remote+192.168.1.1/home/user
  codepod remote codepod-remote://docker-container+mycontainer/bin/sh`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		uri := args[0]
		return runRemote(uri)
	},
}

func runRemote(uri string) error {
	fmt.Printf("Connecting to: %s\n", uri)

	// 1. Parse the URI
	authority, path, err := resolver.SplitAuthorityAndPath(uri)
	if err != nil {
		return fmt.Errorf("failed to parse URI: %w", err)
	}

	fmt.Printf("Authority: %s, Path: %s\n", authority, path)

	// 2. Check if Local Agent is running, if not start it
	rpcClient, err := ensureLocalAgent()
	if err != nil {
		return fmt.Errorf("failed to connect to local agent: %w", err)
	}

	// 3. Forward the request to Local Agent
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := rpcClient.Call(ctx, "Agent.Route", map[string]string{
		"authority": authority,
		"path":      path,
	})
	if err != nil {
		return fmt.Errorf("failed to route: %w", err)
	}

	fmt.Printf("Result: %v\n", result)
	return nil
}

// ensureLocalAgent ensures Local Agent is running and returns RPC client.
func ensureLocalAgent() (*rpc.RPCClient, error) {
	socketPath := getSocketPath()

	// Check if agent is running
	if isAgentRunning(socketPath) {
		return connectToLocalAgent(socketPath)
	}

	// Not running, start Local Agent
	fmt.Println("Starting Local Agent...")
	if err := startLocalAgent(); err != nil {
		return nil, fmt.Errorf("failed to start local agent: %w", err)
	}

	// Wait for agent to be ready
	for i := 0; i < 50; i++ {
		time.Sleep(100 * time.Millisecond)
		if isAgentRunning(socketPath) {
			time.Sleep(200 * time.Millisecond)
			return connectToLocalAgent(socketPath)
		}
	}

	return nil, fmt.Errorf("local agent failed to start")
}

// getSocketPath returns the socket path based on OS.
func getSocketPath() string {
	if runtime.GOOS == "windows" {
		return `\\.\pipe\` + defaultPipeName
	}
	return defaultSocketPath
}

// isAgentRunning checks if the local agent is running.
func isAgentRunning(socketPath string) bool {
	if runtime.GOOS == "windows" {
		_, err := os.Open(socketPath)
		return err == nil
	}
	_, err := os.Stat(socketPath)
	return err == nil
}

// connectToLocalAgent connects to the Local Agent via UDS or Named Pipe.
func connectToLocalAgent(socketPath string) (*rpc.RPCClient, error) {
	var conn channel.Conn
	var err error

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if runtime.GOOS == "windows" {
		conn, err = connectToNamedPipe(ctx, socketPath)
	} else {
		udsCh := channel.NewUDSChannel(socketPath)
		conn, err = udsCh.Dial(ctx, socketPath)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to local agent: %w", err)
	}

	rpcClient := rpc.NewRPCClient(func(data []byte) error {
		_, err := conn.Write(data)
		return err
	})

	return rpcClient, nil
}

// connectToNamedPipe connects to a Windows named pipe.
func connectToNamedPipe(ctx context.Context, pipePath string) (channel.Conn, error) {
	for i := 0; i < 10; i++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		pipe, err := os.OpenFile(pipePath, os.O_RDWR, 0)
		if err == nil {
			return &pipeConn{file: pipe}, nil
		}

		time.Sleep(100 * time.Millisecond)
	}

	return nil, fmt.Errorf("failed to connect to named pipe: %s", pipePath)
}

// pipeConn implements channel.Conn for Windows named pipe.
type pipeConn struct {
	file *os.File
}

func (c *pipeConn) Read(p []byte) (n int, err error)  { return c.file.Read(p) }
func (c *pipeConn) Write(p []byte) (n int, err error) { return c.file.Write(p) }
func (c *pipeConn) Close() error                       { return c.file.Close() }
func (c *pipeConn) LocalAddr() net.Addr               { return nil }
func (c *pipeConn) RemoteAddr() net.Addr              { return nil }

// startLocalAgent starts the Local Agent process.
func startLocalAgent() error {
	socketPath := getSocketPath()

	// Try to use embedded binary via bootstrapper
	binaryPath, err := bootstrapper.GetAgentBinaryPath("")
	if err != nil {
		// Fall back to finding codepod binary and running as subcommand
		binaryPath, err = findCodepodBinary()
		if err != nil {
			return fmt.Errorf("failed to find codepod binary: %w", err)
		}

		// Run as subcommand: codepod local --socket <path>
		cmd := exec.Command(binaryPath, "local", "--socket", socketPath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		cmd.SysProcAttr = &syscall.SysProcAttr{
			Setpgid: true,
		}

		return cmd.Start()
	}

	// Use extracted embedded binary directly with --local flag
	cmd := exec.Command(binaryPath, "--local", "--socket", socketPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	return cmd.Start()
}

// findCodepodBinary finds the codepod binary.
func findCodepodBinary() (string, error) {
	wd, _ := os.Getwd()
	localPath := filepath.Join(wd, "codepod")
	if runtime.GOOS == "windows" {
		localPath += ".exe"
	}
	if _, err := os.Stat(localPath); err == nil {
		return localPath, nil
	}

	path, err := exec.LookPath("codepod")
	if err == nil {
		return path, nil
	}

	usrLocalPath := "/usr/local/bin/codepod"
	if runtime.GOOS == "windows" {
		usrLocalPath = "C:\\Program Files\\codepod\\codepod.exe"
	}
	if _, err := os.Stat(usrLocalPath); err == nil {
		return usrLocalPath, nil
	}

	return "", fmt.Errorf("codepod binary not found")
}

// findAgentBinary finds the agent binary (for provider bootstrap).
func findAgentBinary() (string, error) {
	// Try embedded binary first
	binaryPath, err := bootstrapper.GetAgentBinaryPath("")
	if err == nil {
		return binaryPath, nil
	}

	// Fall back to external binary
	wd, _ := os.Getwd()
	localPath := filepath.Join(wd, "codepod-agent")
	if runtime.GOOS == "windows" {
		localPath += ".exe"
	}
	if _, err := os.Stat(localPath); err == nil {
		return localPath, nil
	}

	path, err := exec.LookPath("codepod-agent")
	if err == nil {
		return path, nil
	}

	return "", fmt.Errorf("agent binary not found")
}

// createLocalAgent creates and configures a Local Agent with providers.
func createLocalAgent() *agent.Agent {
	a := agent.NewLocalAgent()

	// Register providers
	binaryPath, _ := findAgentBinary()

	// Register WSL provider
	wslProvider := provider.NewWSLProvider("Ubuntu", binaryPath)
	a.RegisterProvider(wslProvider)

	// Register SSH provider
	sshProvider := provider.NewSSHProvider("", "", "", "", binaryPath)
	a.RegisterProvider(sshProvider)

	// Register Docker provider
	dockerProvider := provider.NewDockerProvider("", binaryPath)
	a.RegisterProvider(dockerProvider)

	return a
}

func init() {
	rootCmd.AddCommand(remoteCmd)
}
