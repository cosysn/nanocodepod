package fs

import (
	"fmt"

	"github.com/codepod-io/codepod/internal/server"
)

// Client handles file operations via server or direct WSL
type Client struct {
	serverClient *server.Client
	wslDistro   string
	useServer   bool
}

// New creates a new filesystem client
func New(serverURL string, wslDistro string) (*Client, error) {
	client := &Client{
		wslDistro: wslDistro,
	}

	// Try to connect to server
	if serverURL != "" {
		serverClient := server.New(serverURL)
		if err := serverClient.HealthCheck(); err == nil {
			client.serverClient = serverClient
			client.useServer = true
			return client, nil
		}
	}

	// Fall back to direct WSL commands
	client.useServer = false
	return client, nil
}

// ReadFile reads a file from WSL
func (c *Client) ReadFile(path string) (string, error) {
	if c.useServer {
		return c.serverClient.ReadFile(path)
	}

	// Fall back to direct WSL command
	cmd := fmt.Sprintf("cat %s", path)
	out, err := c.runWSLCommand(cmd)
	return string(out), err
}

// WriteFile writes a file to WSL
func (c *Client) WriteFile(path string, content string) error {
	if c.useServer {
		return c.serverClient.WriteFile(path, content)
	}

	// Fall back to direct WSL command using tee
	cmd := fmt.Sprintf("echo '%s' | tee %s > /dev/null", content, path)
	_, err := c.runWSLCommand(cmd)
	return err
}

func (c *Client) runWSLCommand(cmd string) (string, error) {
	// Import wsl package here to avoid circular imports
	// This is a simplified version - in production use the actual wsl package
	return "", fmt.Errorf("WSL direct access not implemented: use server mode")
}
