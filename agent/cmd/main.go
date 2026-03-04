package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

var (
	flagPort      int
	flagPassword  string
	flagServerURL string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "codepod-agent",
		Short: "CodePod agent for remote command execution",
		RunE:  runAgent,
	}

	// Read from environment variables with defaults
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
	defaultServerURL := os.Getenv("CODEPOD_SERVER_URL")

	rootCmd.Flags().IntVar(&flagPort, "port", defaultPort, "Agent SSH port")
	rootCmd.Flags().StringVar(&flagPassword, "password", defaultPassword, "Agent password")
	rootCmd.Flags().StringVar(&flagServerURL, "server", defaultServerURL, "Server URL for heartbeat")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runAgent(cmd *cobra.Command, args []string) error {
	fmt.Printf("Starting CodePod agent on port %d...\n", flagPort)
	fmt.Printf("Server URL: %s\n", flagServerURL)

	// Start heartbeat if server URL is provided
	if flagServerURL != "" {
		go startHeartbeat(flagServerURL, flagPort)
	}

	// TODO: Run SSH+gRPC server

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	fmt.Println("Shutting down agent...")
	return nil
}

// HeartbeatRequest represents a heartbeat request
type HeartbeatRequest struct {
	AgentPort int    `json:"agent_port"`
	Status    string `json:"status"`
	Hostname  string `json:"hostname"`
}

// startHeartbeat sends periodic heartbeats to the server
func startHeartbeat(serverURL string, port int) {
	hostname, _ := os.Hostname()
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		heartbeat(serverURL, port, hostname)
		<-ticker.C
	}
}

func heartbeat(serverURL, port int, hostname string) {
	req := HeartbeatRequest{
		AgentPort: port,
		Status:    "running",
		Hostname:  hostname,
	}

	jsonBody, err := json.Marshal(req)
	if err != nil {
		fmt.Printf("Failed to marshal heartbeat: %v\n", err)
		return
	}

	resp, err := http.Post(serverURL+"/agent/heartbeat", "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		fmt.Printf("Failed to send heartbeat: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Heartbeat failed with status: %d\n", resp.StatusCode)
	}
}
