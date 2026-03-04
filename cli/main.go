package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/codepod-io/cli/internal/client"
	"github.com/codepod-io/cli/internal/provider"

	"github.com/codepod-io/cli/cmd"
)

var version = "dev"

var (
	flagImage       string
	flagRepoURL     string
	flagRepoBranch  string
	flagLocalPath   string
	flagIDEType     string
	flagAutoConnect bool
	flagAgent       bool
	flagNoAgent     bool
)

var upCmd = cmd.NewUpCommand(runUp)

func init() {
	upCmd.Flags().StringVar(&flagImage, "image", "ubuntu:22.04", "Docker image to use")
	upCmd.Flags().StringVar(&flagRepoURL, "repo", "", "Git repository URL")
	upCmd.Flags().StringVar(&flagRepoBranch, "branch", "main", "Git repository branch")
	upCmd.Flags().StringVar(&flagLocalPath, "local", "", "Local directory path to use as workspace")
	upCmd.Flags().StringVar(&flagIDEType, "ide", "vscode", "IDE type (vscode, jetbrains)")
	upCmd.Flags().BoolVar(&flagAutoConnect, "connect", false, "Auto-connect after starting")
	upCmd.Flags().BoolVar(&flagAgent, "agent", true, "Enable agent injection")
	upCmd.Flags().BoolVar(&flagNoAgent, "no-agent", false, "Disable agent injection")

	cmd.RootCmd.AddCommand(upCmd)
}

func runUp(cmd *cobra.Command, args []string) error {
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

	// Check if workspace already exists
	existing, err := c.GetWorkspace(name)
	if err != nil {
		return fmt.Errorf("failed to check workspace: %w", err)
	}

	if existing != nil {
		// Workspace exists, start it
		fmt.Printf("Workspace %s already exists, starting...\n", name)
		if err := c.StartWorkspace(name); err != nil {
			return fmt.Errorf("failed to start workspace: %w", err)
		}
		fmt.Printf("Workspace %s started (Port: %d)\n", name, existing.Port)
		return nil
	}

	// Create new workspace
	fmt.Printf("Creating workspace %s...\n", name)
	workspace, err := c.CreateWorkspace(client.CreateWorkspaceRequest{
		Name:    name,
		Image:   flagImage,
		RepoURL: flagRepoURL,
		Branch:  flagRepoBranch,
	})
	if err != nil {
		return fmt.Errorf("failed to create workspace: %w", err)
	}

	fmt.Printf("Workspace %s created (Port: %d)\n", workspace.Name, workspace.Port)

	return nil
}

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

var style = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("86"))
