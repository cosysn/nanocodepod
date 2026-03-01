package ide

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/codepod-io/codepod/internal/types"
)

// Launcher launches IDEs for workspaces
type Launcher struct{}

// New creates a new IDE launcher
func New() *Launcher {
	return &Launcher{}
}

// Launch launches an IDE for a workspace
func (l *Launcher) Launch(workspace *types.Workspace) error {
	switch workspace.IDE.Type {
	case types.IDETypeVSCode:
		return l.launchVSCode(workspace)
	case types.IDETypeJetBrains:
		return l.launchJetBrains(workspace)
	default:
		return l.launchVSCode(workspace)
	}
}

// launchVSCode launches VS Code with Remote SSH
func (l *Launcher) launchVSCode(workspace *types.Workspace) error {
	// For now, use simple SSH URL approach
	// In production, use vscode.dev or code-server
	user := "root"
	host := fmt.Sprintf("localhost")
	port := workspace.Port

	vscodeURL := fmt.Sprintf("vscode://vscode-remote/ssh-remote+%s@%s:%d/home/%s/workspace",
		user, host, port, user)

	args := []string{"--remote", vscodeURL}

	switch runtime.GOOS {
	case "windows":
		cmd := exec.Command("code", args...)
		cmd.Start()
	case "darwin":
		cmd := exec.Command("open", append([]string{"-a", "Visual Studio Code"}, args...)...)
		cmd.Start()
	case "linux":
		cmd := exec.Command("code", args...)
		cmd.Start()
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	return nil
}

// launchJetBrains launches a JetBrains IDE
func (l *Launcher) launchJetBrains(workspace *types.Workspace) error {
	ide := workspace.IDE.Settings["ide"]
	if ide == "" {
		ide = "idea"
	}

	args := []string{}

	switch runtime.GOOS {
	case "windows":
		cmd := exec.Command(ide, args...)
		cmd.Start()
	case "darwin":
		cmd := exec.Command("open", []string{"-a", ide}...)
		cmd.Start()
	case "linux":
		cmd := exec.Command(ide, args...)
		cmd.Start()
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	return nil
}

// GetSupportedIDEs returns a list of supported IDEs
func (l *Launcher) GetSupportedIDEs() []string {
	return []string{
		"vscode",
		"idea",
		"goland",
		"pycharm",
		"webstorm",
		"rider",
		"clion",
		"webstorm",
	}
}

// IsIDEAvailable checks if an IDE is available on the system
func (l *Launcher) IsIDEAvailable(ide string) bool {
	switch runtime.GOOS {
	case "windows":
		_, err := exec.LookPath("code")
		return err == nil
	case "darwin":
		cmd := exec.Command("mdfind", "kMDItemCFBundleIdentifier==com.microsoft.VSCode")
		out, _ := cmd.Output()
		return len(out) > 0
	case "linux":
		_, err := exec.LookPath(ide)
		return err == nil
	}
	return false
}

// GetVSCodeRemoteArgs returns the VS Code remote arguments for SSH
func GetVSCodeRemoteArgs(host string, port int, user, path string) []string {
	return []string{
		"--remote",
		fmt.Sprintf("ssh-remote+%s@%s:%d%s", user, host, port, path),
	}
}

// GetVSCodeWebURL returns the VS Code Web URL for the workspace
func GetVSCodeWebURL(workspace *types.Workspace) string {
	return fmt.Sprintf("https://vscode.dev/tunnel/%s@localhost:%d",
		"root", workspace.Port)
}

// ParseIDEType parses the IDE type from a string
func ParseIDEType(s string) types.IDEType {
	s = strings.ToLower(s)
	switch s {
	case "vscode", "code":
		return types.IDETypeVSCode
	case "jetbrains", "idea", "goland", "pycharm":
		return types.IDETypeJetBrains
	default:
		return types.IDETypeVSCode
	}
}
