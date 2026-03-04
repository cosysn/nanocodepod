package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// DB is the database connection
var DB *sql.DB

// Workspace represents a workspace in the database
type Workspace struct {
	ID        int64
	Name      string
	Status    string
	Provider  string
	Image     string
	Port      int
	ContainerID string
	DataDir   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// InitDB initializes the SQLite database
func InitDB(dbPath string) error {
	var err error
	DB, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// Create tables
	_, err = DB.Exec(`
		CREATE TABLE IF NOT EXISTS workspaces (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT UNIQUE NOT NULL,
			status TEXT NOT NULL,
			provider TEXT NOT NULL,
			image TEXT,
			port INTEGER,
			container_id TEXT,
			data_dir TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS config (
			key TEXT PRIMARY KEY,
			value TEXT
		);
	`)
	if err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}

	return nil
}

// CreateWorkspace creates a new workspace
func CreateWorkspace(ws *Workspace) (int64, error) {
	result, err := DB.ExecContext(context.Background(),
		`INSERT INTO workspaces (name, status, provider, image, port, container_id, data_dir)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		ws.Name, ws.Status, ws.Provider, ws.Image, ws.Port, ws.ContainerID, ws.DataDir)
	if err != nil {
		return 0, fmt.Errorf("failed to create workspace: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return id, nil
}

// GetWorkspace returns a workspace by name
func GetWorkspace(name string) (*Workspace, error) {
	ws := &Workspace{}
	err := DB.QueryRowContext(context.Background(),
		`SELECT id, name, status, provider, image, port, container_id, data_dir, created_at, updated_at
		 FROM workspaces WHERE name = ?`, name).
		Scan(&ws.ID, &ws.Name, &ws.Status, &ws.Provider, &ws.Image, &ws.Port,
			&ws.ContainerID, &ws.DataDir, &ws.CreatedAt, &ws.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace: %w", err)
	}
	return ws, nil
}

// ListWorkspaces returns all workspaces
func ListWorkspaces() ([]Workspace, error) {
	rows, err := DB.QueryContext(context.Background(),
		`SELECT id, name, status, provider, image, port, container_id, data_dir, created_at, updated_at
		 FROM workspaces`)
	if err != nil {
		return nil, fmt.Errorf("failed to list workspaces: %w", err)
	}
	defer rows.Close()

	var workspaces []Workspace
	for rows.Next() {
		var ws Workspace
		err := rows.Scan(&ws.ID, &ws.Name, &ws.Status, &ws.Provider, &ws.Image, &ws.Port,
			&ws.ContainerID, &ws.DataDir, &ws.CreatedAt, &ws.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan workspace: %w", err)
		}
		workspaces = append(workspaces, ws)
	}
	return workspaces, nil
}

// UpdateWorkspace updates a workspace
func UpdateWorkspace(ws *Workspace) error {
	_, err := DB.ExecContext(context.Background(),
		`UPDATE workspaces SET status = ?, port = ?, container_id = ?, updated_at = CURRENT_TIMESTAMP
		 WHERE name = ?`,
		ws.Status, ws.Port, ws.ContainerID, ws.Name)
	if err != nil {
		return fmt.Errorf("failed to update workspace: %w", err)
	}
	return nil
}

// DeleteWorkspace deletes a workspace
func DeleteWorkspace(name string) error {
	_, err := DB.ExecContext(context.Background(), "DELETE FROM workspaces WHERE name = ?", name)
	if err != nil {
		return fmt.Errorf("failed to delete workspace: %w", err)
	}
	return nil
}

// GetConfig returns a config value
func GetConfig(key string) (string, error) {
	var value string
	err := DB.QueryRowContext(context.Background(), "SELECT value FROM config WHERE key = ?", key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("failed to get config: %w", err)
	}
	return value, nil
}

// SetConfig sets a config value
func SetConfig(key, value string) error {
	_, err := DB.ExecContext(context.Background(),
		"INSERT OR REPLACE INTO config (key, value) VALUES (?, ?)", key, value)
	if err != nil {
		return fmt.Errorf("failed to set config: %w", err)
	}
	return nil
}
