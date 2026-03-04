package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var cfgFile string

// RootCmd is the root command
var RootCmd = &cobra.Command{
	Use:   "codepod",
	Short: "Container development environment management",
	Long:  `Manage container-based development environments.`,
}

// Execute executes the root command
func Execute() error {
	return RootCmd.Execute()
}

// NewUpCommand creates a new up command
func NewUpCommand(runE func(*cobra.Command, []string) error) *cobra.Command {
	return &cobra.Command{
		Use:   "up [workspace-name]",
		Short: "Create and start a workspace",
		Long:  `Create a new workspace and start it. If workspace already exists, just start it.`,
		Args:  cobra.ExactArgs(1),
		RunE:  runE,
	}
}

func init() {
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ~/.codepod/config.yaml)")
}

func ExecuteAndExit() {
	if err := Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
