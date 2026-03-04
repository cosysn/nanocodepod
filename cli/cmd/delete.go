package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/codepod-io/cli/internal/client"
	"github.com/codepod-io/cli/internal/provider"
)

var deleteCmd = &cobra.Command{
	Use:   "delete [workspace-name]",
	Short: "Delete a workspace",
	Long:  `Delete a workspace and its data.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runDelete,
}

func init() {
	RootCmd.AddCommand(deleteCmd)
}

func runDelete(cmd *cobra.Command, args []string) error {
	name := args[0]

	// Get provider
	p := provider.GetProvider()
	if p == nil {
		return fmt.Errorf("no provider configured")
	}

	// Discover server
	serverInfo, err := p.DiscoverServer()
	if err != nil {
		return fmt.Errorf("failed to discover server: %w", err)
	}

	if serverInfo.Status != "running" {
		return fmt.Errorf("server is not running. Status: %s", serverInfo.Status)
	}

	// Create client
	baseURL := serverInfo.URL[:len(serverInfo.URL)-len("/health")]
	c := client.NewClient(baseURL)

	// Delete workspace
	fmt.Printf("Deleting workspace %s...\n", name)
	if err := c.DeleteWorkspace(name); err != nil {
		return fmt.Errorf("failed to delete workspace: %w", err)
	}

	fmt.Printf("Workspace %s deleted\n", name)
	return nil
}
