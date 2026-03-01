package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/codepod-io/codepod/internal/workspace"
)

var stopCmd = &cobra.Command{
	Use:   "stop [workspace-name]",
	Short: "Stop a workspace",
	Long:  `Stop a running workspace.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runStop,
}

func init() {
	rootCmd.AddCommand(stopCmd)
}

func runStop(cmd *cobra.Command, args []string) error {
	name := args[0]

	wsm, err := workspace.New()
	if err != nil {
		return fmt.Errorf("failed to create workspace manager: %w", err)
	}

	ws, err := wsm.Stop(name)
	if err != nil {
		return fmt.Errorf("failed to stop workspace: %w", err)
	}

	fmt.Printf("Workspace '%s' stopped (State: %s)\n", name, ws.State)
	return nil
}
