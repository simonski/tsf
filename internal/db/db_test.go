package db

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"
)

func TestOpenDatabase(t *testing.T) {
	// Create temporary directory for test database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	
	// Open database
	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()
	
	// Check that database file was created
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("Database file was not created")
	}
	
	// Test connection
	if err := db.Conn().Ping(); err != nil {
		t.Errorf("Failed to ping database: %v", err)
	}
	
	// Verify path
	if db.Path() != dbPath {
		t.Errorf("Expected path %s, got %s", dbPath, db.Path())
	}
}

func TestInitSchema(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	
	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()
	
	// Initialize schema
	if err := db.InitSchema(); err != nil {
		t.Fatalf("Failed to initialize schema: %v", err)
	}
	
	// Verify tables exist
	tables := []string{"users", "projects", "roles", "tasks", "task_history", "config", "project_members"}
	
	for _, table := range tables {
		var name string
		err := db.Conn().QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&name)
		if err != nil {
			t.Errorf("Table %s does not exist: %v", table, err)
		}
	}
}

func TestInitializeDatabase(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	
	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()
	
	// Initialize database
	adminPassword, orchestratorPassword, err := db.InitializeDatabase()
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	
	// Verify passwords were generated
	if adminPassword == "" {
		t.Error("Admin password is empty")
	}
	if orchestratorPassword == "" {
		t.Error("Orchestrator password is empty")
	}
	if adminPassword == orchestratorPassword {
		t.Error("Admin and orchestrator passwords are identical")
	}
	
	// Verify admin user exists
	var username string
	err = db.Conn().QueryRow("SELECT username FROM users WHERE username = ?", "admin").Scan(&username)
	if err != nil {
		t.Errorf("Admin user was not created: %v", err)
	}
	
	// Verify orchestrator user exists
	err = db.Conn().QueryRow("SELECT username FROM users WHERE username = ?", "orchestrator").Scan(&username)
	if err != nil {
		t.Errorf("Orchestrator user was not created: %v", err)
	}
	
	// Verify default project exists
	var projectName string
	err = db.Conn().QueryRow("SELECT name FROM projects WHERE name = ?", "default").Scan(&projectName)
	if err != nil {
		t.Errorf("Default project was not created: %v", err)
	}
	
	// Verify system roles exist
	roles := []string{"programmer", "tester", "analyst", "reviewer"}
	for _, role := range roles {
		var name string
		err = db.Conn().QueryRow("SELECT name FROM roles WHERE name = ? AND scope = 'system'", role).Scan(&name)
		if err != nil {
			t.Errorf("System role %s was not created: %v", role, err)
		}
	}
	
	// Verify config values exist
	configs := []string{"orchestrator.heartbeat", "orchestrator.idle", "worker.heartbeat", "worker.idle"}
	for _, key := range configs {
		var value string
		err = db.Conn().QueryRow("SELECT value FROM config WHERE key = ?", key).Scan(&value)
		if err != nil {
			t.Errorf("Config %s was not created: %v", key, err)
		}
	}
}

func TestTransaction(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	
	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()
	
	if err := db.InitSchema(); err != nil {
		t.Fatalf("Failed to initialize schema: %v", err)
	}
	
	// Test successful transaction
	err = db.Transaction(func(tx *sql.Tx) error {
		_, err := tx.Exec("INSERT INTO config (key, value) VALUES (?, ?)", "test.key", "test.value")
		return err
	})
	
	if err != nil {
		t.Errorf("Transaction failed: %v", err)
	}
	
	// Verify data was committed
	var value string
	err = db.Conn().QueryRow("SELECT value FROM config WHERE key = ?", "test.key").Scan(&value)
	if err != nil {
		t.Errorf("Data was not committed: %v", err)
	}
	if value != "test.value" {
		t.Errorf("Expected value 'test.value', got '%s'", value)
	}
}
