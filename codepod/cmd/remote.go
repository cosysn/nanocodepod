// Package cmd provides CLI commands for CodePod.
package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/codepod-io/codepod/pkg/agent"
	"github.com/codepod-io/codepod/pkg/channel"
	"github.com/codepod-io/codepod/pkg/provider"
	"github.com/codepod-io/codepod/pkg/resolver"
	"github.com/codepod-io/codepod/pkg/rpc"
	"github.com/spf13/cobra"
)

const (
	defaultSocketPath = "/tmp/codepod.sock"
	localAgentBinary  = "codepod-agent"
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
	// Check if socket exists
	if _, err := os.Stat(defaultSocketPath); err == nil {
		// Socket exists, try to connect
		return connectToLocalAgent()
	}

	// Socket doesn't exist, start Local Agent
	fmt.Println("Starting Local Agent...")
	if err := startLocalAgent(); err != nil {
		return nil, fmt.Errorf("failed to start local agent: %w", err)
	}

	// Wait for agent to be ready
	for i := 0; i < 30; i++ {
		time.Sleep(100 * time.Millisecond)
		if _, err := os.Stat(defaultSocketPath); err == nil {
			return connectToLocalAgent()
		}
	}

	return nil, fmt.Errorf("local agent failed to start")
}

// connectToLocalAgent connects to the Local Agent via UDS.
func connectToLocalAgent() (*rpc.RPCClient, error) {
	// Create UDS channel
	udsCh := channel.NewUDSChannel(defaultSocketPath)

	// Dial to the socket
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := udsCh.Dial(ctx, defaultSocketPath)
	if err != nil {
		return nil, fmt.Errorf("failed to dial local agent: %w", err)
	}

	// Create RPC client that sends over the connection
	rpcClient := rpc.NewRPCClient(func(data []byte) error {
		_, err := conn.Write(data)
		return err
	})

	return rpcClient, nil
}

// startLocalAgent starts the Local Agent process.
func startLocalAgent() error {
	// Find the agent binary
	binaryPath, err := findAgentBinary()
	if err != nil {
		return err
	}

	// Start the agent
	cmd := exec.Command(binaryPath, "--local", "--socket", defaultSocketPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Start()
}

// findAgentBinary finds the codepod-agent binary.
func findAgentBinary() (string, error) {
	// Check current directory
	wd, _ := os.Getwd()
	localPath := filepath.Join(wd, localAgentBinary)
	if _, err := os.Stat(localPath); err == nil {
		return localPath, nil
	}

	// Check PATH
	path, err := exec.LookPath(localAgentBinary)
	if err == nil {
		return path, nil
	}

	return "", fmt.Errorf("agent binary not found: %s", localAgentBinary)
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
