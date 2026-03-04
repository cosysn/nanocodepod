package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/codepod-io/server/internal/agent"
	"github.com/codepod-io/server/internal/db"
	"github.com/codepod-io/server/internal/docker"
	"github.com/codepod-io/server/internal/filesystem"
	"github.com/codepod-io/server/internal/workspace"
)

func main() {
	// Get port from command line args or environment, or use default
	port := os.Getenv("CODEPOD_SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	// Get data directory
	dataDir := os.Getenv("CODEPOD_DATA_DIR")
	if dataDir == "" {
		dataDir = "/tmp/codepod"
	}

	// Initialize database
	dbPath := dataDir + "/codepod.db"
	if err := db.InitDB(dbPath); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	log.Printf("Database initialized at %s", dbPath)

	// Create workspace manager
	dockerClient := docker.New()
	wsManager := workspace.NewManager(dockerClient, dataDir+"/workspaces")

	// Print port to stdout so CLI can read it
	fmt.Printf("CODEPOD_SERVER_PORT=%s\n", port)

	// Register handlers
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/docker/ps", docker.ListContainers)
	http.HandleFunc("/docker/run", docker.RunContainer)
	http.HandleFunc("/docker/pull", docker.PullImage)
	http.HandleFunc("/fs/read", filesystem.ReadFile)
	http.HandleFunc("/fs/write", filesystem.WriteFile)

	// Workspace handlers
	http.HandleFunc("/workspace/create", wsManager.Create)
	http.HandleFunc("/workspace/list", wsManager.List)
	http.HandleFunc("/workspace/get", wsManager.Get)
	http.HandleFunc("/workspace/delete", wsManager.Delete)
	http.HandleFunc("/workspace/start", wsManager.Start)
	http.HandleFunc("/workspace/stop", wsManager.Stop)

	// Agent handlers
	http.HandleFunc("/agent/heartbeat", agent.RegisterAgent)
	http.HandleFunc("/agent/list", agent.ListAgents)

	addr := ":" + port
	log.Printf("Starting CodePod Server on %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
