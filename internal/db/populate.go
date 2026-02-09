package db

import (
	"bufio"
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// PopulateFromScripts reads markdown files from embedded scripts/initdb and populates the database
func (db *DB) PopulateFromScripts() error {
	// Populate in order: users, workers, orchestrators, projects, roles
	if err := db.populateUsers("scripts/initdb/users.md"); err != nil {
		return fmt.Errorf("failed to populate users: %w", err)
	}

	if err := db.populateWorkers("scripts/initdb/workers.md"); err != nil {
		return fmt.Errorf("failed to populate workers: %w", err)
	}

	if err := db.populateOrchestrators("scripts/initdb/orchestrators.md"); err != nil {
		return fmt.Errorf("failed to populate orchestrators: %w", err)
	}

	if err := db.populateProjects("scripts/initdb/projects.md"); err != nil {
		return fmt.Errorf("failed to populate projects: %w", err)
	}

	if err := db.populateRoles("scripts/initdb/roles.md"); err != nil {
		return fmt.Errorf("failed to populate roles: %w", err)
	}

	return nil
}

// parseMarkdownTable extracts data from markdown code blocks
func parseMarkdownTable(filePath string) ([][]string, error) {
	file, err := scriptsFS.Open(filePath)
	if err != nil {
		return nil, nil // File doesn't exist, skip
	}
	defer file.Close()

	var rows [][]string
	scanner := bufio.NewScanner(file)
	inCodeBlock := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if strings.HasPrefix(line, "```") {
			inCodeBlock = !inCodeBlock
			continue
		}

		if inCodeBlock && line != "" {
			// Split by pipe and trim spaces
			parts := strings.Split(line, "|")
			var row []string
			for _, part := range parts {
				row = append(row, strings.TrimSpace(part))
			}
			rows = append(rows, row)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return rows, nil
}

// populateUsers creates users from users.md
// Format: username | type | password
func (db *DB) populateUsers(filePath string) error {
	rows, err := parseMarkdownTable(filePath)
	if err != nil || len(rows) == 0 {
		return err
	}

	return db.Transaction(func(tx *sql.Tx) error {
		for _, row := range rows {
			if len(row) < 3 {
				continue
			}

			username := row[0]
			userType := row[1]
			password := row[2]

			// Hash password
			hash, err := HashPassword(password)
			if err != nil {
				return fmt.Errorf("failed to hash password for %s: %w", username, err)
			}

			id := uuid.New().String()
			_, err = tx.Exec(`
				INSERT INTO users (id, username, password_hash, type, is_active)
				VALUES (?, ?, ?, ?, ?)
			`, id, username, hash, userType, true)
			if err != nil {
				return fmt.Errorf("failed to create user %s: %w", username, err)
			}
		}
		return nil
	})
}

// populateWorkers creates worker users from workers.md
// Format: username | password
func (db *DB) populateWorkers(filePath string) error {
	rows, err := parseMarkdownTable(filePath)
	if err != nil || len(rows) == 0 {
		return err
	}

	return db.Transaction(func(tx *sql.Tx) error {
		for _, row := range rows {
			if len(row) < 2 {
				continue
			}

			username := row[0]
			password := row[1]

			// Hash password
			hash, err := HashPassword(password)
			if err != nil {
				return fmt.Errorf("failed to hash password for %s: %w", username, err)
			}

			id := uuid.New().String()
			_, err = tx.Exec(`
				INSERT INTO users (id, username, password_hash, type, is_active)
				VALUES (?, ?, ?, ?, ?)
			`, id, username, hash, "worker", true)
			if err != nil {
				return fmt.Errorf("failed to create worker %s: %w", username, err)
			}
		}
		return nil
	})
}

// populateOrchestrators creates orchestrator users from orchestrators.md
// Format: username | password
func (db *DB) populateOrchestrators(filePath string) error {
	rows, err := parseMarkdownTable(filePath)
	if err != nil || len(rows) == 0 {
		return err
	}

	return db.Transaction(func(tx *sql.Tx) error {
		for _, row := range rows {
			if len(row) < 2 {
				continue
			}

			username := row[0]
			password := row[1]

			// Hash password
			hash, err := HashPassword(password)
			if err != nil {
				return fmt.Errorf("failed to hash password for %s: %w", username, err)
			}

			id := uuid.New().String()
			_, err = tx.Exec(`
				INSERT INTO users (id, username, password_hash, type, is_active)
				VALUES (?, ?, ?, ?, ?)
			`, id, username, hash, "orchestrator", true)
			if err != nil {
				return fmt.Errorf("failed to create orchestrator %s: %w", username, err)
			}
		}
		return nil
	})
}

// populateProjects creates projects from projects.md
// Format: name | description | repository | status | visibility
func (db *DB) populateProjects(filePath string) error {
	rows, err := parseMarkdownTable(filePath)
	if err != nil || len(rows) == 0 {
		return err
	}

	// Get admin user ID
	var adminID string
	err = db.Conn().QueryRow("SELECT id FROM users WHERE username = 'admin' LIMIT 1").Scan(&adminID)
	if err != nil {
		return fmt.Errorf("failed to get admin user: %w", err)
	}

	return db.Transaction(func(tx *sql.Tx) error {
		for _, row := range rows {
			if len(row) < 5 {
				continue
			}

			name := row[0]
			description := row[1]
			repository := row[2]
			status := row[3]
			visibility := row[4]

			var repoPtr *string
			if repository != "" {
				repoPtr = &repository
			}

			id := uuid.New().String()
			_, err = tx.Exec(`
				INSERT INTO projects (id, name, description, repository, status, visibility, created_by, updated_by)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?)
			`, id, name, description, repoPtr, status, visibility, adminID, adminID)
			if err != nil {
				return fmt.Errorf("failed to create project %s: %w", name, err)
			}
		}
		return nil
	})
}

// populateRoles creates roles from roles.md
// Format: name | description | goals | scope | project_name
func (db *DB) populateRoles(filePath string) error {
	rows, err := parseMarkdownTable(filePath)
	if err != nil || len(rows) == 0 {
		return err
	}

	// Get admin user ID
	var adminID string
	err = db.Conn().QueryRow("SELECT id FROM users WHERE username = 'admin' LIMIT 1").Scan(&adminID)
	if err != nil {
		return fmt.Errorf("failed to get admin user: %w", err)
	}

	return db.Transaction(func(tx *sql.Tx) error {
		for _, row := range rows {
			if len(row) < 5 {
				continue
			}

			name := row[0]
			description := row[1]
			goals := row[2]
			scope := row[3]
			projectName := row[4]

			var projectID *string
			if scope == "project" && projectName != "" {
				var pid string
				err := tx.QueryRow("SELECT id FROM projects WHERE name = ?", projectName).Scan(&pid)
				if err != nil {
					return fmt.Errorf("failed to find project %s for role %s: %w", projectName, name, err)
				}
				projectID = &pid
			}

			id := uuid.New().String()
			_, err = tx.Exec(`
				INSERT INTO roles (id, name, description, goals, scope, project_id, is_active, created_by, updated_by)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
			`, id, name, description, goals, scope, projectID, true, adminID, adminID)
			if err != nil {
				return fmt.Errorf("failed to create role %s: %w", name, err)
			}
		}
		return nil
	})
}
