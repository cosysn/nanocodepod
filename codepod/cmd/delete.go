package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/codepod-io/codepod/internal/workspace"
)

var deleteCmd = &cobra.Command{
	Use:   "delete [workspace-name]",
	Short: "Delete a workspace",
	Long:  `Delete a workspace and its storage.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runDelete,
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}

func runDelete(cmd *cobra.Command, args []string) error {
	name := args[0]

	wsm, err := workspace.New()
	if err != nil {
		return fmt.Errorf("failed to create workspace manager: %w", err)
	}

	if err := wsm.Delete(name); err != nil {
		return fmt.Errorf("failed to delete workspace: %w", err)
	}

	fmt.Printf("Workspace '%s' deleted.\n", name)
	return nil
}
