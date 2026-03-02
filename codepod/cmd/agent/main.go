package main

import (
	"fmt"
	"os"

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

	rootCmd.Flags().IntVar(&flagPort, "port", 22001, "Agent SSH port")
	rootCmd.Flags().StringVar(&flagPassword, "password", "codepod", "Agent password")

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
