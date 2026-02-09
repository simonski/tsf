package db

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
)

// InitializeDatabase sets up a new database with default data
// Returns: adminPassword, userPassword, workerPassword, orchestratorPassword, error
func (db *DB) InitializeDatabase() (adminPassword, userPassword, workerPassword, orchestratorPassword string, err error) {
	return db.InitializeDatabaseWithPasswords("")
}

// InitializeDatabaseWithPasswords sets up a new database with custom passwords
// If password is an empty string, random passwords will be generated for all users
// If password is provided, all users will have the same password
// Returns: adminPassword, userPassword, workerPassword, orchestratorPassword, error
func (db *DB) InitializeDatabaseWithPasswords(password string) (string, string, string, string, error) {
	// Apply schema
	if err := db.InitSchema(); err != nil {
		return "", "", "", "", fmt.Errorf("failed to initialize schema: %w", err)
	}

	// Generate or use provided password for all users
	var err error
	var adminPassword, userPassword, workerPassword, orchestratorPassword string

	if password == "" {
		// Generate unique random passwords for each user
		adminPassword, err = GeneratePassword(16)
		if err != nil {
			return "", "", "", "", fmt.Errorf("failed to generate admin password: %w", err)
		}

		userPassword, err = GeneratePassword(16)
		if err != nil {
			return "", "", "", "", fmt.Errorf("failed to generate user password: %w", err)
		}

		workerPassword, err = GeneratePassword(16)
		if err != nil {
			return "", "", "", "", fmt.Errorf("failed to generate worker password: %w", err)
		}

		orchestratorPassword, err = GeneratePassword(16)
		if err != nil {
			return "", "", "", "", fmt.Errorf("failed to generate orchestrator password: %w", err)
		}
	} else {
		// Use the same password for all users
		adminPassword = password
		userPassword = password
		workerPassword = password
		orchestratorPassword = password
	}

	// Hash passwords
	adminHash, err := HashPassword(adminPassword)
	if err != nil {
		return "", "", "", "", fmt.Errorf("failed to hash admin password: %w", err)
	}

	userHash, err := HashPassword(userPassword)
	if err != nil {
		return "", "", "", "", fmt.Errorf("failed to hash user password: %w", err)
	}

	workerHash, err := HashPassword(workerPassword)
	if err != nil {
		return "", "", "", "", fmt.Errorf("failed to hash worker password: %w", err)
	}

	orchestratorHash, err := HashPassword(orchestratorPassword)
	if err != nil {
		return "", "", "", "", fmt.Errorf("failed to hash orchestrator password: %w", err)
	}

	// Create admin, user, worker and orchestrator users, default project, system roles, and config
	err = db.Transaction(func(tx *sql.Tx) error {
		// Create admin user
		adminID := uuid.New().String()
		_, err := tx.Exec(`
			INSERT INTO users (id, username, password_hash, type, is_active)
			VALUES (?, ?, ?, ?, ?)
		`, adminID, "admin", adminHash, "human", true)
		if err != nil {
			return fmt.Errorf("failed to create admin user: %w", err)
		}

		// Create normal user
		userID := uuid.New().String()
		_, err = tx.Exec(`
			INSERT INTO users (id, username, password_hash, type, is_active)
			VALUES (?, ?, ?, ?, ?)
		`, userID, "user", userHash, "human", true)
		if err != nil {
			return fmt.Errorf("failed to create normal user: %w", err)
		}

		// Create worker
		workerID := uuid.New().String()
		_, err = tx.Exec(`
			INSERT INTO users (id, username, password_hash, type, is_active)
			VALUES (?, ?, ?, ?, ?)
		`, workerID, "worker", workerHash, "worker", true)
		if err != nil {
			return fmt.Errorf("failed to create worker user: %w", err)
		}

		// Create orchestrator user
		orchestratorID := uuid.New().String()
		_, err = tx.Exec(`
			INSERT INTO users (id, username, password_hash, type, is_active)
			VALUES (?, ?, ?, ?, ?)
		`, orchestratorID, "orchestrator", orchestratorHash, "orchestrator", true)
		if err != nil {
			return fmt.Errorf("failed to create orchestrator user: %w", err)
		}

		// Create default project
		defaultProjectID := uuid.New().String()
		_, err = tx.Exec(`
			INSERT INTO projects (id, name, description, status, visibility, created_by, updated_by)
			VALUES (?, ?, ?, ?, ?, ?, ?)
		`, defaultProjectID, "default", "Default project for general tasks", "active", "public", adminID, adminID)
		if err != nil {
			return fmt.Errorf("failed to create default project: %w", err)
		}

		// Create system roles
		roles := []struct {
			name        string
			description string
			goals       string
		}{
			{
				name:        "programmer",
				description: "Software developer who implements features and fixes bugs",
				goals: `You are a skilled software developer. Your responsibilities include:
- Writing clean, maintainable, and well-tested code
- Following best practices and coding standards
- Implementing features based on specifications
- Fixing bugs and addressing technical debt
- Documenting your code appropriately
- Collaborating with other team members`,
			},
			{
				name:        "tester",
				description: "Quality assurance specialist who verifies functionality",
				goals: `You are a quality assurance specialist. Your responsibilities include:
- Writing comprehensive test cases
- Executing manual and automated tests
- Identifying and documenting bugs
- Verifying that acceptance criteria are met
- Ensuring software quality and reliability
- Working with developers to reproduce and fix issues`,
			},
			{
				name:        "analyst",
				description: "Business analyst who breaks down requirements",
				goals: `You are a business analyst. Your responsibilities include:
- Understanding business requirements and user needs
- Breaking down large requirements into manageable tasks
- Writing clear and detailed specifications
- Defining acceptance criteria
- Collaborating with stakeholders
- Ensuring requirements are feasible and well-understood`,
			},
			{
				name:        "reviewer",
				description: "Code reviewer who ensures quality and standards",
				goals: `You are a code reviewer. Your responsibilities include:
- Reviewing code for quality, correctness, and maintainability
- Ensuring adherence to coding standards
- Identifying potential bugs and security issues
- Providing constructive feedback
- Verifying test coverage
- Ensuring documentation is adequate`,
			},
		}

		for _, role := range roles {
			roleID := uuid.New().String()
			_, err = tx.Exec(`
				INSERT INTO roles (id, name, description, goals, scope, is_active, created_by, updated_by)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?)
			`, roleID, role.name, role.description, role.goals, "system", true, adminID, adminID)
			if err != nil {
				return fmt.Errorf("failed to create role %s: %w", role.name, err)
			}
		}

		// Create default config values
		configs := []struct {
			key         string
			value       string
			description string
		}{
			{"orchestrator.heartbeat", "1000", "Orchestrator heartbeat interval in milliseconds"},
			{"orchestrator.idle", "10000", "Orchestrator idle timeout in milliseconds"},
			{"worker.heartbeat", "1000", "Worker heartbeat interval in milliseconds"},
			{"worker.idle", "10000", "Worker idle timeout in milliseconds"},
		}

		for _, cfg := range configs {
			_, err = tx.Exec(`
				INSERT INTO config (key, value, description)
				VALUES (?, ?, ?)
			`, cfg.key, cfg.value, cfg.description)
			if err != nil {
				return fmt.Errorf("failed to create config %s: %w", cfg.key, err)
			}
		}

		return nil
	})

	if err != nil {
		return "", "", "", "", err
	}

	return adminPassword, userPassword, workerPassword, orchestratorPassword, nil
}
