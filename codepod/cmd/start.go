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

func init() {
	rootCmd.AddCommand(startCmd)
}

func runStart(cmd *cobra.Command, args []string) error {
	name := args[0]

	wsm, err := workspace.New()
	if err != nil {
		return fmt.Errorf("failed to create workspace manager: %w", err)
	}

	ws, err := wsm.Start(name)
	if err != nil {
		return fmt.Errorf("failed to start workspace: %w", err)
	}

	fmt.Printf("Workspace '%s' started (State: %s)\n", name, ws.State)
	fmt.Printf("  SSH Port: %d\n", ws.Port)
	return nil
}
