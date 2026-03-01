package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/codepod-io/codepod/internal/config"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage codepod configuration",
	Long:  `Manage codepod configuration settings.`,
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Args:  cobra.ExactArgs(2),
	RunE:  runConfigSet,
}

var configGetCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Get a configuration value",
	Args:  cobra.ExactArgs(1),
	RunE:  runConfigGet,
}

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configuration",
	RunE:  runConfigList,
}

var configResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset configuration to defaults",
	RunE:  runConfigReset,
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configListCmd)
	configCmd.AddCommand(configResetCmd)
}

func runConfigSet(cmd *cobra.Command, args []string) error {
	key := args[0]
	value := args[1]

	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Set value based on key
	switch key {
	case "wsl.distribution":
		cfg.WSL.Distribution = value
	case "wsl.docker-host":
		cfg.WSL.DockerHost = value
	case "general.default-ide":
		cfg.General.DefaultIDE = value
	case "general.ssh-port":
		fmt.Sscanf(value, "%d", &cfg.General.SSHPort)
	default:
		return fmt.Errorf("unknown config key: %s", key)
	}

	if err := config.SaveConfig(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("Set %s = %s\n", key, value)
	return nil
}

func runConfigGet(cmd *cobra.Command, args []string) error {
	key := args[0]

	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	switch key {
	case "wsl.distribution":
		fmt.Println(cfg.WSL.Distribution)
	case "wsl.docker-host":
		fmt.Println(cfg.WSL.DockerHost)
	case "general.default-ide":
		fmt.Println(cfg.General.DefaultIDE)
	case "general.ssh-port":
		fmt.Println(cfg.General.SSHPort)
	case "port-pool.start":
		fmt.Println(cfg.PortPool.Start)
	case "port-pool.end":
		fmt.Println(cfg.PortPool.End)
	default:
		return fmt.Errorf("unknown config key: %s", key)
	}

	return nil
}

func runConfigList(cmd *cobra.Command, args []string) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	fmt.Println("Current configuration:")
	fmt.Println("----------------------")
	fmt.Printf("WSL Distribution: %s\n", cfg.WSL.Distribution)
	fmt.Printf("Docker Host: %s\n", cfg.WSL.DockerHost)
	fmt.Printf("Default IDE: %s\n", cfg.General.DefaultIDE)
	fmt.Printf("SSH Port: %d\n", cfg.General.SSHPort)
	fmt.Printf("Port Pool: %d-%d\n", cfg.PortPool.Start, cfg.PortPool.End)
	fmt.Printf("Used Ports: %v\n", cfg.PortPool.Used)

	return nil
}

func runConfigReset(cmd *cobra.Command, args []string) error {
	defaultCfg := config.GetDefaultConfig()
	if err := config.SaveConfig(defaultCfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Println("Configuration reset to defaults.")
	return nil
}
