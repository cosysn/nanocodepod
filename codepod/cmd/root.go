package cmd

import (
	"github.com/codepod-io/codepod/internal/config"
	"github.com/spf13/cobra"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "codepod",
	Short: "CodePod - 容器开发环境管理工具",
	Long:  `CodePod 是一个容器开发环境管理工具，支持本地、WSL、SSH 和 Docker 容器。`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Skip config for agent subcommands
		if cmd.Name() == "workspace" || cmd.Name() == "container" || cmd.Name() == "local" {
			return
		}
		// Initialize config directory
		if cfgFile != "" {
			config.SetConfigDir(cfgFile)
		}
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ~/.codepod/config.yaml)")

	// Add agent subcommands
	rootCmd.AddCommand(workspaceCmd)
	rootCmd.AddCommand(containerCmd)
	rootCmd.AddCommand(localCmd)
}
