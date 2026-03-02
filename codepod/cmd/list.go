package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/codepod-io/codepod/internal/workspace"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all workspaces",
	Long:  `List all workspaces and their status.`,
	RunE:  runList,
}

func init() {
	rootCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) error {
	wsm, err := workspace.New()
	if err != nil {
		return fmt.Errorf("failed to create workspace manager: %w", err)
	}

	workspaces, err := wsm.List()
	if err != nil {
		return fmt.Errorf("failed to list workspaces: %w", err)
	}

	if len(workspaces) == 0 {
		fmt.Println("No workspaces found.")
		return nil
	}

	fmt.Println("Workspaces:")
	fmt.Println("------------")
	for _, ws := range workspaces {
		fmt.Printf("  %s\t%s\tPort: %d\tAgent: %s\n", ws.Name, ws.State, ws.Port, ws.Agent.Status)
	}

	return nil
}
