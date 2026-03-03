package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/codepod-io/codepod/internal/workspace"
)

var startCmd = &cobra.Command{
	Use:   "start [workspace-name]",
	Short: "Start a workspace",
	Long:  `Start a stopped workspace.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runStart,
}

var (
	flagStartAgent bool
)

func init() {
	startCmd.Flags().BoolVar(&flagStartAgent, "agent", true, "Enable agent injection (PID 0, SSH+gRPC)")
	startCmd.Flags().BoolVar(&flagStartAgent, "no-agent", false, "Disable agent injection")

	rootCmd.AddCommand(startCmd)
}

func runStart(cmd *cobra.Command, args []string) error {
	name := args[0]

	wsm, err := workspace.New()
	if err != nil {
		return fmt.Errorf("failed to create workspace manager: %w", err)
	}

	ws, err := wsm.Start(name, flagStartAgent)
	if err != nil {
		return fmt.Errorf("failed to start workspace: %w", err)
	}

	fmt.Printf("Workspace '%s' started (State: %s)\n", name, ws.State)
	fmt.Printf("  SSH Port: %d\n", ws.Port)
	fmt.Printf("  Agent: %s\n", ws.Agent.Status)
	if ws.Agent.Status == "running" {
		fmt.Println("\nAgent Connection Info:")
		fmt.Printf("  Host: localhost\n")
		fmt.Printf("  Port: %d\n", ws.Agent.Port)
		fmt.Printf("  Username: root\n")
		fmt.Printf("  Password: codepod\n")
		fmt.Printf("  gRPC: localhost:%d (for command dispatch)\n\n", ws.Agent.Port+1)
	}
	return nil
}
