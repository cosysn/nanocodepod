package cmd

import (
	"github.com/codepod-io/codepod/internal/config"
	"github.com/spf13/cobra"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "codepod",
	Short: "WSL容器开发环境管理工具",
	Long:  `在Windows上基于WSL构建容器开发环境，类似devpod但更简单。`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
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
}
