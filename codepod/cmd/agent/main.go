package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/codepod-io/codepod/internal/agent"
	"github.com/spf13/cobra"
)

var (
	flagPort     int
	flagPassword string
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

	rootCmd.Flags().IntVar(&flagPort, "port", defaultPort, "Agent SSH port")
	rootCmd.Flags().StringVar(&flagPassword, "password", defaultPassword, "Agent password")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runAgent(cmd *cobra.Command, args []string) error {
	fmt.Printf("Starting CodePod agent on port %d...\n", flagPort)

	if err := agent.RunAgent(flagPort, flagPassword); err != nil {
		return fmt.Errorf("failed to run agent: %w", err)
	}

	return nil
}
