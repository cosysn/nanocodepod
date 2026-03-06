// Package main provides the codepod-agent entry point.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/codepod-io/codepod/pkg/agent"
	"github.com/codepod-io/codepod/pkg/channel"
	"github.com/codepod-io/codepod/pkg/provider"
	"github.com/codepod-io/codepod/pkg/router"
)

func main() {
	socketPath := "/tmp/codepod.sock"
	agentType := router.AgentTypeLocal

	// Parse arguments
	for i := 1; i < len(os.Args); i++ {
		switch os.Args[i] {
		case "--socket":
			if i+1 < len(os.Args) {
				socketPath = os.Args[i+1]
				i++
			}
		case "--local":
			agentType = router.AgentTypeLocal
		case "--workspace":
			agentType = router.AgentTypeWorkspace
		case "--container":
			agentType = router.AgentTypeContainer
		case "--port":
			// Legacy support - port flag ignored for UDS
			if i+1 < len(os.Args) {
				i++
			}
		}
	}

	if err := runAgent(agentType, socketPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runAgent(agentType router.AgentType, socketPath string) error {
	fmt.Printf("Starting CodePod agent (%s) on socket: %s\n", agentType, socketPath)

	// Create agent based on type
	var a *agent.Agent
	switch agentType {
	case router.AgentTypeLocal:
		a = createLocalAgent()
	case router.AgentTypeWorkspace:
		a = agent.NewWorkspaceAgent()
	case router.AgentTypeContainer:
		a = agent.NewContainerAgent()
	default:
		a = agent.NewLocalAgent()
	}

	// Set up signal handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		fmt.Println("\nShutting down agent...")
		cancel()
	}()

	// Start UDS listener for RPC
	udsCh := channel.NewUDSChannel(socketPath)
	listener, err := udsCh.Listen(socketPath)
	if err != nil {
		return fmt.Errorf("failed to listen on socket: %w", err)
	}
	defer listener.Close()

	fmt.Printf("Agent listening on %s\n", socketPath)

	// Handle connections
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		// Set accept timeout to allow checking context
		listener.SetDeadline(time.Now().Add(500 * time.Millisecond))

		conn, err := listener.Accept()
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			// Timeout is expected, continue to check context
			continue
		}

		// Handle RPC connection
		go handleConnection(ctx, a, conn)
	}
}

func createLocalAgent() *agent.Agent {
	a := agent.NewLocalAgent()

	// Find agent binary path
	binaryPath := findAgentBinary()

	// Register providers
	a.RegisterProvider(provider.NewWSLProvider("Ubuntu", binaryPath))
	a.RegisterProvider(provider.NewSSHProvider("", "", "", "", binaryPath))
	a.RegisterProvider(provider.NewDockerProvider("", binaryPath))

	return a
}

func findAgentBinary() string {
	// Check current directory
	if _, err := os.Stat("./codepod-agent"); err == nil {
		return "./codepod-agent"
	}

	// Check /usr/local/bin
	if _, err := os.Stat("/usr/local/bin/codepod-agent"); err == nil {
		return "/usr/local/bin/codepod-agent"
	}

	return "codepod-agent"
}

func handleConnection(ctx context.Context, a *agent.Agent, conn channel.Conn) {
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

var (
	flagPort     int
	flagPassword string
)

func init() {
	defaultPort := 22001
	if envPort := os.Getenv("CODEPOD_AGENT_PORT"); envPort != "" {
		if port, err := strconv.Atoi(envPort); err == nil && port > 0 {
			defaultPort = port
		}
	}
	defaultPassword := "codepod"
	if envPass := os.Getenv("CODEPOD_AGENT_PASSWORD"); envPass != "" {
		defaultPassword = envPass
	}

	_ = defaultPort  // Silence unused warning
	_ = defaultPassword
	_ = flagPort     // Silence unused warning
	_ = flagPassword // Silence unused warning
}
