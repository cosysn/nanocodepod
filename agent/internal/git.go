package agent

import (
	"fmt"
	"io"
	"net"
	"os/exec"
	"strings"
)

// GitForward handles Git protocol forwarding
type GitForward struct {
	localPort  int
	remoteHost string
	remotePort int
	listener   net.Listener
}

// NewGitForward creates a new Git forwarder
func NewGitForward(localPort int, remoteHost string, remotePort int) *GitForward {
	return &GitForward{
		localPort:  localPort,
		remoteHost: remoteHost,
		remotePort: remotePort,
	}
}

// Start starts the Git forwarder
func (g *GitForward) Start() error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", g.localPort))
	if err != nil {
		return fmt.Errorf("failed to listen on port %d: %w", g.localPort, err)
	}
	g.listener = listener

	go g.acceptConnections()
	return nil
}

// Stop stops the Git forwarder
func (g *GitForward) Stop() error {
	if g.listener != nil {
		return g.listener.Close()
	}
	return nil
}

func (g *GitForward) acceptConnections() {
	for {
		conn, err := g.listener.Accept()
		if err != nil {
			break
		}

		go g.forward(conn)
	}
}

func (g *GitForward) forward(clientConn net.Conn) {
	defer clientConn.Close()

	// Connect to remote
	serverConn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", g.remoteHost, g.remotePort))
	if err != nil {
		return
	}
	defer serverConn.Close()

	// Bidirectional copy
	go func() {
		io.Copy(serverConn, clientConn)
	}()
	io.Copy(clientConn, serverConn)
}

// SetupGitConfig sets up Git configuration in container
func SetupGitConfig(containerID, name, email string) error {
	cmds := [][]string{
		{"git", "config", "--global", "user.name", name},
		{"git", "config", "--global", "user.email", email},
		{"git", "config", "--global", "init.defaultBranch", "main"},
		{"git", "config", "--global", "pull.rebase", "false"},
	}

	for _, cmd := range cmds {
		c := exec.Command("docker", append([]string{"exec", containerID}, cmd...)...)
		if err := c.Run(); err != nil {
			return fmt.Errorf("failed to run %v: %w", cmd, err)
		}
	}

	return nil
}

// CloneRepo clones a Git repository in container
func CloneRepo(containerID, repoURL, branch, targetPath string) error {
	cmd := exec.Command("docker", "exec", containerID, "git", "clone", "-b", branch, repoURL, targetPath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to clone repo: %w", err)
	}
	return nil
}

// GetGitStatus returns the Git status in container
func GetGitStatus(containerID, path string) (string, error) {
	cmd := exec.Command("docker", "exec", containerID, "git", "-C", path, "status", "--porcelain")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// SetupSSHKeys sets up SSH keys for Git in container
func SetupSSHKeys(containerID, privateKey string) error {
	// Create .ssh directory
	cmd := exec.Command("docker", "exec", containerID, "mkdir", "-p", "/root/.ssh")
	if err := cmd.Run(); err != nil {
		return err
	}

	// Write private key
	echoCmd := fmt.Sprintf("echo '%s' > /root/.ssh/id_rsa", privateKey)
	cmd = exec.Command("docker", "exec", containerID, "bash", "-c", echoCmd)
	if err := cmd.Run(); err != nil {
		return err
	}

	// Set permissions
	cmd = exec.Command("docker", "exec", containerID, "chmod", "600", "/root/.ssh/id_rsa")
	if err := cmd.Run(); err != nil {
		return err
	}

	// Add github.com to known_hosts
	cmd = exec.Command("docker", "exec", containerID, "ssh-keyscan", "github.com", ">>", "/root/.ssh/known_hosts")
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}
