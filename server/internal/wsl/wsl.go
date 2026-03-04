package wsl

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

// GetDistroName returns the default WSL distribution name
func GetDistroName() string {
	// Check WSL_DISTRO_NAME environment variable (set by WSL)
	distro := os.Getenv("WSL_DISTRO_NAME")
	if distro != "" {
		return distro
	}
	return "Ubuntu-22.04"
}

// RunInWSL runs a command inside the WSL distribution
func RunInWSL(distro, command string) (string, error) {
	cmd := exec.Command("wsl", "-d", distro, "bash", "-c", command)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("WSL command failed: %v", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// ListDistros returns a list of available WSL distributions
func ListDistros() ([]string, error) {
	cmd := exec.Command("wsl", "--list", "--quiet")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list WSL distributions: %v", err)
	}

	var distros []string
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			distros = append(distros, line)
		}
	}
	return distros, nil
}

// IsWSL returns true if running inside WSL
func IsWSL() bool {
	_, ok := os.LookupEnv("WSL_DISTRO_NAME")
	return ok
}

// GetIP returns the IP address of the WSL instance
func GetIP() (string, error) {
	cmd := exec.Command("wsl", "hostname", "-I")
	output, err := cmd.Output()
	if err != nil {
		log.Printf("Failed to get WSL IP: %v", err)
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}
