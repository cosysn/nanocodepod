package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/codepod-io/codepod/internal/config"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize codepod configuration",
	Long:  `Initialize codepod configuration directory and default config.`,
	RunE:  runInit,
}

var flagForce bool

func init() {
	initCmd.Flags().BoolVar(&flagForce, "force", false, "Force initialization, overwrite existing config")
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	// Check if config already exists
	_, err := config.LoadConfig()
	if err == nil && !flagForce {
		fmt.Println("Config already exists. Use --force to overwrite.")
		fmt.Printf("Config location: ~/.codepod/config.yaml\n")
		return nil
	}

	// Ensure config directory exists
	if err := config.EnsureConfigDir(); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Save default config
	defaultCfg := config.GetDefaultConfig()
	if err := config.SaveConfig(defaultCfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Println("Codepod initialized successfully!")
	fmt.Println("Config location: ~/.codepod/config.yaml")
	fmt.Printf("Default WSL distribution: %s\n", defaultCfg.WSL.Distribution)
	fmt.Printf("Default Docker host: %s\n", defaultCfg.WSL.DockerHost)
	fmt.Printf("Port pool: %d-%d\n", defaultCfg.PortPool.Start, defaultCfg.PortPool.End)

	return nil
}
