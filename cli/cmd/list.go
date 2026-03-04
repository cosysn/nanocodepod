package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/codepod-io/cli/internal/client"
	"github.com/codepod-io/cli/internal/provider"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all workspaces",
	Long:  `List all workspaces and their status.`,
	RunE:  runList,
}

func init() {
	RootCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) error {
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
	baseURL := serverInfo.URL[:len(serverInfo.URL)-len("/health")] // Remove /health suffix
	c := client.NewClient(baseURL)

	// List workspaces
	workspaces, err := c.ListWorkspaces()
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
		fmt.Printf("  %s\t%s\tPort: %d\n", ws.Name, ws.State, ws.Port)
	}

	return nil
}
