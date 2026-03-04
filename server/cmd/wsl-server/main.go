package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/codepod-io/server/internal/docker"
	"github.com/codepod-io/server/internal/filesystem"
)

func main() {
	// Get port from command line args or environment, or use default
	port := os.Getenv("CODEPOD_SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	// Print port to stdout so CLI can read it
	fmt.Printf("CODEPOD_SERVER_PORT=%s\n", port)

	// Register handlers
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/docker/ps", docker.ListContainers)
	http.HandleFunc("/docker/run", docker.RunContainer)
	http.HandleFunc("/docker/pull", docker.PullImage)
	http.HandleFunc("/fs/read", filesystem.ReadFile)
	http.HandleFunc("/fs/write", filesystem.WriteFile)

	addr := ":" + port
	log.Printf("Starting CodePod WSL Server on %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
