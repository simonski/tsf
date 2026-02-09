package db

import (
	"database/sql"
	"embed"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed schema.sql
var schemaFS embed.FS

//go:embed scripts/initdb/users.md
//go:embed scripts/initdb/workers.md
//go:embed scripts/initdb/orchestrators.md
//go:embed scripts/initdb/projects.md
//go:embed scripts/initdb/roles.md
var scriptsFS embed.FS

// DB wraps the SQLite database connection
type DB struct {
	conn *sql.DB
	path string
}

// Open opens a connection to the SQLite database
func Open(dbPath string) (*DB, error) {
	// Ensure directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// Open database
	conn, err := sql.Open("sqlite3", dbPath+"?_foreign_keys=on")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := conn.Ping(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{
		conn: conn,
		path: dbPath,
	}, nil
}

// Close closes the database connection
func (db *DB) Close() error {
	if db.conn != nil {
		return db.conn.Close()
	}
	return nil
}

// Conn returns the underlying database connection
func (db *DB) Conn() *sql.DB {
	return db.conn
}

// Path returns the database file path
func (db *DB) Path() string {
	return db.path
}

// InitSchema applies the database schema
func (db *DB) InitSchema() error {
	schema, err := schemaFS.ReadFile("schema.sql")
	if err != nil {
		return fmt.Errorf("failed to read schema file: %w", err)
	}

	if _, err := db.conn.Exec(string(schema)); err != nil {
		return fmt.Errorf("failed to execute schema: %w", err)
	}

	return nil
}

// Transaction executes a function within a database transaction
func (db *DB) Transaction(fn func(*sql.Tx) error) error {
	tx, err := db.conn.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// UpsertHeartbeat inserts or updates a heartbeat record
func (db *DB) UpsertHeartbeat(userID, status string, taskID *string) error {
	query := `
		INSERT INTO heartbeats (id, user_id, status, task_id, last_seen)
		VALUES (?, ?, ?, ?, datetime('now'))
		ON CONFLICT(user_id) DO UPDATE SET
			status = excluded.status,
			task_id = excluded.task_id,
			last_seen = excluded.last_seen
	`

	// Generate a unique ID for new heartbeats
	id := userID + "-heartbeat"

	_, err := db.conn.Exec(query, id, userID, status, taskID)
	if err != nil {
		return fmt.Errorf("failed to upsert heartbeat: %w", err)
	}

	return nil
}

// GetHeartbeat retrieves a heartbeat by user ID
func (db *DB) GetHeartbeat(userID string) (*Heartbeat, error) {
	query := `
		SELECT id, user_id, status, task_id, last_seen, created_at
		FROM heartbeats
		WHERE user_id = ?
	`

	var hb Heartbeat
	var taskID sql.NullString

	err := db.conn.QueryRow(query, userID).Scan(
		&hb.ID,
		&hb.UserID,
		&hb.Status,
		&taskID,
		&hb.LastSeen,
		&hb.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get heartbeat: %w", err)
	}

	if taskID.Valid {
		hb.TaskID = &taskID.String
	}

	return &hb, nil
}

// ListHeartbeats lists all heartbeats
func (db *DB) ListHeartbeats() ([]Heartbeat, error) {
	query := `
		SELECT id, user_id, status, task_id, last_seen, created_at
		FROM heartbeats
		ORDER BY last_seen DESC
	`

	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list heartbeats: %w", err)
	}
	defer rows.Close()

	var heartbeats []Heartbeat
	for rows.Next() {
		var hb Heartbeat
		var taskID sql.NullString

		if err := rows.Scan(&hb.ID, &hb.UserID, &hb.Status, &taskID, &hb.LastSeen, &hb.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan heartbeat: %w", err)
		}

		if taskID.Valid {
			hb.TaskID = &taskID.String
		}

		heartbeats = append(heartbeats, hb)
	}

	return heartbeats, nil
}
