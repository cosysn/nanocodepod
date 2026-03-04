package workspace

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/codepod-io/server/internal/db"
	"github.com/codepod-io/server/internal/docker"
	"github.com/codepod-io/server/internal/port"
	"github.com/codepod-io/server/internal/storage"
)

// Manager handles workspace operations
type Manager struct {
	dockerClient docker.DockerClient
	portPool     *port.Pool
	storageMgr   *storage.Manager
}

// NewManager creates a new workspace manager
func NewManager(dockerClient docker.DockerClient, storageDir string) *Manager {
	return &Manager{
		dockerClient: dockerClient,
		portPool:     port.NewPool(port.DefaultPortRangeStart, port.DefaultPortRangeEnd),
		storageMgr:   storage.NewManager(storageDir),
	}
}

// CreateRequest represents a workspace create request
type CreateRequest struct {
	Name     string `json:"name"`
	Provider string `json:"provider"`
	Image    string `json:"image"`
	RepoURL  string `json:"repo_url,omitempty"`
}

// CreateResponse represents a workspace create response
type CreateResponse struct {
	Name      string `json:"name"`
	Status    string `json:"status"`
	Port      int    `json:"port"`
	ContainerID string `json:"container_id,omitempty"`
}

// ListResponse represents a workspace list response
type ListResponse struct {
	Workspaces []db.Workspace `json:"workspaces"`
}

// Create handles workspace creation
func (m *Manager) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req CreateRequest
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Allocate port
	port, err := m.portPool.Allocate()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Ensure storage
	storagePath, err := m.storageMgr.EnsureStorage(req.Name)
	if err != nil {
		m.portPool.Release(port)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Create container
	containerID, err := m.dockerClient.CreateContainer(&docker.ContainerConfig{
		Name:    fmt.Sprintf("codepod-%s", req.Name),
		Image:   req.Image,
		PortBindings: map[string][]docker.PortBinding{
			"22/tcp": {{HostPort: fmt.Sprintf("%d", port)}},
		},
		Binds: []string{fmt.Sprintf("%s:/workspace", storagePath)},
		Labels: map[string]string{
			"codepod.workspace": req.Name,
		},
	})
	if err != nil {
		m.portPool.Release(port)
		m.storageMgr.DeleteStorage(req.Name)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Start container
	if err := m.dockerClient.StartContainer(containerID); err != nil {
		m.portPool.Release(port)
		m.storageMgr.DeleteStorage(req.Name)
		m.dockerClient.RemoveContainer(containerID, true)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Save to database
	ws := &db.Workspace{
		Name:        req.Name,
		Status:      "running",
		Provider:    req.Provider,
		Image:       req.Image,
		Port:        port,
		ContainerID: containerID,
		DataDir:     storagePath,
	}
	_, err = db.CreateWorkspace(ws)
	if err != nil {
		log.Printf("failed to save workspace to database: %v", err)
	}

	resp := CreateResponse{
		Name:        req.Name,
		Status:      "running",
		Port:        port,
		ContainerID: containerID,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// List handles workspace listing
func (m *Manager) List(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	workspaces, err := db.ListWorkspaces()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := ListResponse{Workspaces: workspaces}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// Get handles workspace get
func (m *Manager) Get(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	name := r.URL.Query().Get("name")
	if name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	ws, err := db.GetWorkspace(name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if ws == nil {
		http.Error(w, "workspace not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ws)
}

// Delete handles workspace deletion
func (m *Manager) Delete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	name := r.URL.Query().Get("name")
	if name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	ws, err := db.GetWorkspace(name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if ws == nil {
		http.Error(w, "workspace not found", http.StatusNotFound)
		return
	}

	// Stop and remove container
	if ws.ContainerID != "" {
		m.dockerClient.StopContainer(ws.ContainerID)
		m.dockerClient.RemoveContainer(ws.ContainerID, true)
	}

	// Release port
	if ws.Port > 0 {
		m.portPool.Release(ws.Port)
	}

	// Delete storage
	m.storageMgr.DeleteStorage(name)

	// Delete from database
	db.DeleteWorkspace(name)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("deleted"))
}

// Start handles workspace start
func (m *Manager) Start(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	name := r.URL.Query().Get("name")
	if name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	ws, err := db.GetWorkspace(name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if ws == nil {
		http.Error(w, "workspace not found", http.StatusNotFound)
		return
	}

	if err := m.dockerClient.StartContainer(ws.ContainerID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ws.Status = "running"
	db.UpdateWorkspace(ws)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("started"))
}

// Stop handles workspace stop
func (m *Manager) Stop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	name := r.URL.Query().Get("name")
	if name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	ws, err := db.GetWorkspace(name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if ws == nil {
		http.Error(w, "workspace not found", http.StatusNotFound)
		return
	}

	if err := m.dockerClient.StopContainer(ws.ContainerID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ws.Status = "stopped"
	db.UpdateWorkspace(ws)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("stopped"))
}
