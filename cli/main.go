package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/codepod-io/cli/internal/provider"
)

var version = "dev"

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "codepod",
	Short: "Container development environment management",
	Long:  `Manage container-based development environments.`,
}

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

var upCmd = &cobra.Command{
	Use:   "up [workspace-name]",
	Short: "Create and start a workspace",
	Long:  `Create a new workspace and start it. If workspace already exists, just start it.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runUp,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ~/.codepod/config.yaml)")

	upCmd.Flags().StringVar(&flagImage, "image", "ubuntu:22.04", "Docker image to use")
	upCmd.Flags().StringVar(&flagRepoURL, "repo", "", "Git repository URL")
	upCmd.Flags().StringVar(&flagRepoBranch, "branch", "main", "Git repository branch")
	upCmd.Flags().StringVar(&flagLocalPath, "local", "", "Local directory path to use as workspace")
	upCmd.Flags().StringVar(&flagIDEType, "ide", "vscode", "IDE type (vscode, jetbrains)")
	upCmd.Flags().BoolVar(&flagAutoConnect, "connect", false, "Auto-connect after starting")
	upCmd.Flags().BoolVar(&flagAgent, "agent", true, "Enable agent injection")
	upCmd.Flags().BoolVar(&flagNoAgent, "no-agent", false, "Disable agent injection")

	rootCmd.AddCommand(upCmd)
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

	// Create workspace via server
	fmt.Printf("Creating workspace %s...\n", name)
	fmt.Printf("Server: %s\n", serverInfo.URL)
	fmt.Printf("Status: running\n")

	return nil
}

func main() {
	if err := Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

var style = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("86"))
