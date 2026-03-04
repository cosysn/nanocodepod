package docker

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os/exec"
	"strings"
)

// Container represents a Docker container
type Container struct {
	ID      string `json:"id"`
	Image   string `json:"image"`
	Status  string `json:"status"`
	Names   string `json:"names"`
	Created string `json:"created"`
}

// ListContainers handles GET /docker/ps
func ListContainers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cmd := exec.Command("docker", "ps", "--format", "{{.ID}}|{{.Image}}|{{.Status}}|{{.Names}}|{{.Created}}")
	output, err := cmd.Output()
	if err != nil {
		log.Printf("docker ps failed: %v", err)
		http.Error(w, fmt.Sprintf("docker ps failed: %v", err), http.StatusInternalServerError)
		return
	}

	var containers []Container
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Split(line, "|")
		if len(parts) >= 5 {
			containers = append(containers, Container{
				ID:      parts[0],
				Image:   parts[1],
				Status:  parts[2],
				Names:   parts[3],
				Created: parts[4],
			})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(containers)
}

// RunContainer handles POST /docker/run
func RunContainer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	// Parse docker run command from body (simplified - in production use proper JSON)
	image := string(body)
	if image == "" {
		image = "ubuntu:22.04"
	}

	cmd := exec.Command("docker", "run", "-d", image)
	output, err := cmd.Output()
	if err != nil {
		log.Printf("docker run failed: %v", err)
		http.Error(w, fmt.Sprintf("docker run failed: %v", err), http.StatusInternalServerError)
		return
	}

	w.Write(output)
}

// PullImage handles POST /docker/pull
func PullImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	image := strings.TrimSpace(string(body))
	if image == "" {
		http.Error(w, "Image name required", http.StatusBadRequest)
		return
	}

	cmd := exec.Command("docker", "pull", image)
	cmd.Stdout = w
	cmd.Stderr = w

	if err := cmd.Run(); err != nil {
		log.Printf("docker pull failed: %v", err)
		// Already wrote output, just log
	}
}
