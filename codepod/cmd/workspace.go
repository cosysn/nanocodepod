package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/codepod-io/codepod/pkg/agent"
	"github.com/codepod-io/codepod/pkg/channel"
	"github.com/spf13/cobra"
)

var workspaceCmd = &cobra.Command{
	Use:   "workspace",
	Short: "Start workspace agent",
	Long:  `Start the workspace agent for handling remote workspace connections.`,
	RunE:  runWorkspaceAgent,
}

func runWorkspaceAgent(cmd *cobra.Command, args []string) error {
	socketPath := "/tmp/codepod-workspace.sock"

	// Allow socket path override
	if len(args) > 0 {
		socketPath = args[0]
	}

	fmt.Printf("Starting CodePod workspace agent on socket: %s\n", socketPath)

	// Create workspace agent
	a := agent.NewWorkspaceAgent()

	// Set up context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Println("\nShutting down workspace agent...")
		cancel()
	}()

	// Start UDS listener
	udsCh := channel.NewUDSChannel(socketPath)
	listener, err := udsCh.Listen(socketPath)
	if err != nil {
		return fmt.Errorf("failed to listen on socket: %w", err)
	}
	defer listener.Close()

	fmt.Printf("Workspace agent listening on %s\n", socketPath)

	// Handle connections
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		conn, err := listener.Accept()
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			continue
		}

		go handleAgentConnection(ctx, a, conn)
	}
}

var containerCmd = &cobra.Command{
	Use:   "container",
	Short: "Start container agent",
	Long:  `Start the container agent for handling container connections.`,
	RunE:  runContainerAgent,
}

func runContainerAgent(cmd *cobra.Command, args []string) error {
	socketPath := "/tmp/codepod-container.sock"

	// Allow socket path override
	if len(args) > 0 {
		socketPath = args[0]
	}

	fmt.Printf("Starting CodePod container agent on socket: %s\n", socketPath)

	// Create container agent
	a := agent.NewContainerAgent()

	// Set up context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Println("\nShutting down container agent...")
		cancel()
	}()

	// Start UDS listener
	udsCh := channel.NewUDSChannel(socketPath)
	listener, err := udsCh.Listen(socketPath)
	if err != nil {
		return fmt.Errorf("failed to listen on socket: %w", err)
	}
	defer listener.Close()

	fmt.Printf("Container agent listening on %s\n", socketPath)

	// Handle connections
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		conn, err := listener.Accept()
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			continue
		}

		go handleAgentConnection(ctx, a, conn)
	}
}

var localCmd = &cobra.Command{
	Use:   "local",
	Short: "Start local agent",
	Long:  `Start the local agent for handling local workspace connections.`,
	RunE:  runLocalAgent,
}

func runLocalAgent(cmd *cobra.Command, args []string) error {
	socketPath := "/tmp/codepod.sock"

	// Allow socket path override
	if len(args) > 0 {
		socketPath = args[0]
	}

	fmt.Printf("Starting CodePod local agent on socket: %s\n", socketPath)

	// Create local agent
	a := agent.NewLocalAgent()

	// Set up context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Println("\nShutting down local agent...")
		cancel()
	}()

	// Start UDS listener
	udsCh := channel.NewUDSChannel(socketPath)
	listener, err := udsCh.Listen(socketPath)
	if err != nil {
		return fmt.Errorf("failed to listen on socket: %w", err)
	}
	defer listener.Close()

	fmt.Printf("Local agent listening on %s\n", socketPath)

	// Handle connections
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		conn, err := listener.Accept()
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			continue
		}

		go handleAgentConnection(ctx, a, conn)
	}
}

// handleAgentConnection handles a single agent connection.
func handleAgentConnection(ctx context.Context, a *agent.Agent, conn channel.Conn) {
	defer conn.Close()

	rpcServer := a.RPCServer

	// Read and handle requests
	buf := make([]byte, 4096)
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		n, err := conn.Read(buf)
		if err != nil {
			return
		}

		// Handle RPC request
		response := rpcServer.HandleRequest(buf[:n])
		if len(response) > 0 {
			conn.Write(response)
		}
	}
}

// init registers agent-specific flags if needed
func init() {
	// Add flags for agent subcommands
	workspaceCmd.Flags().String("socket", "/tmp/codepod-workspace.sock", "socket path")
	containerCmd.Flags().String("socket", "/tmp/codepod-container.sock", "socket path")
	localCmd.Flags().String("socket", "/tmp/codepod.sock", "socket path")
}
