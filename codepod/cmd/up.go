package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/codepod-io/codepod/internal/config"
	"github.com/codepod-io/codepod/internal/ide"
	"github.com/codepod-io/codepod/internal/types"
	"github.com/codepod-io/codepod/internal/workspace"
	"github.com/codepod-io/codepod/internal/wsl"
)

var upCmd = &cobra.Command{
	Use:   "up [workspace-name]",
	Short: "Create and start a workspace",
	Long:  `Create a new workspace and start it. If workspace already exists, just start it.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runUp,
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

func init() {
	upCmd.Flags().StringVar(&flagImage, "image", "ubuntu:22.04", "Docker image to use")
	upCmd.Flags().StringVar(&flagRepoURL, "repo", "", "Git repository URL")
	upCmd.Flags().StringVar(&flagRepoBranch, "branch", "main", "Git repository branch")
	upCmd.Flags().StringVar(&flagLocalPath, "local", "", "Local directory path to use as workspace")
	upCmd.Flags().StringVar(&flagIDEType, "ide", "vscode", "IDE type (vscode, jetbrains)")
	upCmd.Flags().BoolVar(&flagAutoConnect, "connect", false, "Auto-connect after starting")
	upCmd.Flags().BoolVar(&flagAgent, "agent", true, "Enable agent injection (PID 0, SSH+gRPC)")
	upCmd.Flags().BoolVar(&flagNoAgent, "no-agent", false, "Disable agent injection")

	rootCmd.AddCommand(upCmd)
}

func runUp(cmd *cobra.Command, args []string) error {
	name := args[0]

	// Ensure config directory exists
	if err := config.EnsureConfigDir(); err != nil {
		return fmt.Errorf("failed to ensure config directory: %w", err)
	}

	// Detect platform
	platform, err := wsl.NewPlatform()
	if err != nil {
		return fmt.Errorf("failed to detect platform: %w", err)
	}

	fmt.Printf("Detected platform: %s\n", platform.Type)

	// Check Docker availability
	accessMode := wsl.DetectDockerAccessMode()
	if accessMode == wsl.DockerAccessNone {
		return fmt.Errorf("Docker is not available. Please ensure Docker is running\n\nDebug info:\n%s", wsl.GetDockerAccessModeDebug())
	}
	fmt.Printf("Docker access mode: %s\n", accessMode)

	// Create workspace manager
	wsm, err := workspace.New()
	if err != nil {
		return fmt.Errorf("failed to create workspace manager: %w", err)
	}

	// Check if workspace exists
	exists, err := wsm.Exists(name)
	if err != nil {
		return fmt.Errorf("failed to check workspace: %w", err)
	}

	// Handle agent flag: --no-agent overrides --agent
	injectAgent := flagAgent && !flagNoAgent

	var ws *types.Workspace

	if exists {
		fmt.Printf("Workspace %s already exists, starting...\n", name)
		ws, err = wsm.Start(name, injectAgent)
		if err != nil {
			return fmt.Errorf("failed to start workspace: %w", err)
		}
	} else {
		fmt.Printf("Creating workspace %s...\n", name)

		// Get IDE type
		ideType := ide.ParseIDEType(flagIDEType)

		createOpts := &workspace.CreateOptions{
			Image: flagImage,
			Repository: types.Repository{
				URL:       flagRepoURL,
				Branch:    flagRepoBranch,
				LocalPath: flagLocalPath,
			},
			IDE: types.IDE{
				Type: ideType,
			},
			InjectAgent: injectAgent,
		}

		ws, err = wsm.Create(name, createOpts)
		if err != nil {
			return fmt.Errorf("failed to create workspace: %w", err)
		}

		// Start the workspace
		ws, err = wsm.Start(name, injectAgent)
		if err != nil {
			return fmt.Errorf("failed to start workspace: %w", err)
		}
	}

	// Print workspace info
	fmt.Printf("\nWorkspace '%s' is running!\n", name)
	fmt.Printf("  SSH Port: %d\n", ws.Port)
	fmt.Printf("  State: %s\n", ws.State)
	fmt.Printf("  Agent: %s\n", ws.Agent.Status)
	if ws.Agent.Status == "running" {
		fmt.Println("\nAgent Connection Info:")
		fmt.Printf("  Host: localhost\n")
		fmt.Printf("  Port: %d\n", ws.Agent.Port)
		fmt.Printf("  Username: root\n")
		fmt.Printf("  Password: codepod\n")
		fmt.Printf("  gRPC: localhost:%d (for command dispatch)\n\n", ws.Agent.Port+1)
	}

	// Auto-connect if requested
	if flagAutoConnect {
		launcher := ide.New()
		if err := launcher.Launch(ws); err != nil {
			return fmt.Errorf("failed to launch IDE: %w", err)
		}
		fmt.Println("IDE launched!")
	}

	return nil
}
