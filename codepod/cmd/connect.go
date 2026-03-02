package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/codepod-io/codepod/internal/ide"
	"github.com/codepod-io/codepod/internal/workspace"
)

var connectCmd = &cobra.Command{
	Use:   "connect [workspace-name]",
	Short: "Connect to a workspace and launch IDE",
	Long:  `Connect to a workspace and launch the configured IDE.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runConnect,
}

func init() {
	rootCmd.AddCommand(connectCmd)
}

func runConnect(cmd *cobra.Command, args []string) error {
	name := args[0]

	wsm, err := workspace.New()
	if err != nil {
		return fmt.Errorf("failed to create workspace manager: %w", err)
	}

	ws, err := wsm.Get(name)
	if err != nil {
		return fmt.Errorf("failed to get workspace: %w", err)
	}

	// Start workspace if not running
	if ws.State != "running" {
		fmt.Printf("Starting workspace %s...\n", name)
		ws, err = wsm.Start(name, ws.Agent.Status == "running")
		if err != nil {
			return fmt.Errorf("failed to start workspace: %w", err)
		}
	}

	// Launch IDE
	launcher := ide.New()
	if err := launcher.Launch(ws); err != nil {
		return fmt.Errorf("failed to launch IDE: %w", err)
	}

	fmt.Printf("IDE launched for workspace '%s'!\n", name)
	fmt.Printf("  SSH Port: %d\n", ws.Port)
	return nil
}
