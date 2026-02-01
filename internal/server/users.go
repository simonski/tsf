package server

import (
	"database/sql"
	"net/http"

	"github.com/google/uuid"
	"github.com/simonski/task/internal/db"
)

// RegisterRequest represents a user registration request
type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Type     string `json:"type"`
}

// handleRegisterUser creates a new user
func (s *Server) handleRegisterUser(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := decodeJSON(r, &req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Username == "" || req.Password == "" || req.Type == "" {
		sendError(w, http.StatusBadRequest, "username, password, and type are required")
		return
	}

	if req.Type != "human" && req.Type != "worker" && req.Type != "orchestrator" {
		sendError(w, http.StatusBadRequest, "type must be 'human', 'worker', or 'orchestrator'")
		return
	}

	// Hash password
	passwordHash, err := db.HashPassword(req.Password)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to hash password")
		return
	}

	userID := uuid.New().String()
	_, err = s.db.Conn().Exec(`
		INSERT INTO users (id, username, password_hash, type, is_active)
		VALUES (?, ?, ?, ?, ?)
	`, userID, req.Username, passwordHash, req.Type, true)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to create user")
		return
	}

	var user db.User
	var createdAt, updatedAt string
	err = s.db.Conn().QueryRow(`
		SELECT id, username, password_hash, type, is_active, created_at, updated_at
		FROM users
		WHERE id = ?
	`, userID).Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.Type,
		&user.IsActive,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to retrieve user")
		return
	}
	user.CreatedAt = parseTimestamp(createdAt)
	user.UpdatedAt = parseTimestamp(updatedAt)

	// Don't send password hash back
	user.PasswordHash = ""
	sendJSON(w, http.StatusCreated, user)
}

// handleListUsers returns all users
func (s *Server) handleListUsers(w http.ResponseWriter, r *http.Request) {
	query := "SELECT id, username, type, is_active, created_at, updated_at FROM users WHERE 1=1"
	args := []interface{}{}

	if userType := r.URL.Query().Get("type"); userType != "" {
		query += " AND type = ?"
		args = append(args, userType)
	}

	query += " ORDER BY username"

	rows, err := s.db.Conn().Query(query, args...)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to query users")
		return
	}
	defer rows.Close()

	var users []db.User
	for rows.Next() {
		var user db.User
		var createdAt, updatedAt string
		err := rows.Scan(&user.ID, &user.Username, &user.Type, &user.IsActive, &createdAt, &updatedAt)
		if err != nil {
			sendError(w, http.StatusInternalServerError, "failed to scan user")
			return
		}
		user.CreatedAt = parseTimestamp(createdAt)
		user.UpdatedAt = parseTimestamp(updatedAt)
		users = append(users, user)
	}

	sendJSON(w, http.StatusOK, users)
}

// handleGetUser returns a specific user
func (s *Server) handleGetUser(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("user_id")

	var user db.User
	var createdAt, updatedAt string
	err := s.db.Conn().QueryRow(`
		SELECT id, username, type, is_active, created_at, updated_at
		FROM users
		WHERE id = ?
	`, userID).Scan(&user.ID, &user.Username, &user.Type, &user.IsActive, &createdAt, &updatedAt)
	if err == sql.ErrNoRows {
		sendError(w, http.StatusNotFound, "user not found")
		return
	}
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to query user")
		return
	}
	user.CreatedAt = parseTimestamp(createdAt)
	user.UpdatedAt = parseTimestamp(updatedAt)

	sendJSON(w, http.StatusOK, user)
}

// handleEnableUser enables a user
func (s *Server) handleEnableUser(w http.ResponseWriter, r *http.Request) {
	username := r.PathValue("username")

	result, err := s.db.Conn().Exec("UPDATE users SET is_active = ? WHERE username = ?", true, username)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to enable user")
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		sendError(w, http.StatusNotFound, "user not found")
		return
	}

	var user db.User
	var createdAt, updatedAt string
	err = s.db.Conn().QueryRow(`
		SELECT id, username, type, is_active, created_at, updated_at
		FROM users
		WHERE username = ?
	`, username).Scan(&user.ID, &user.Username, &user.Type, &user.IsActive, &createdAt, &updatedAt)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to retrieve user")
		return
	}

	sendJSON(w, http.StatusOK, user)
}

// handleDisableUser disables a user
func (s *Server) handleDisableUser(w http.ResponseWriter, r *http.Request) {
	username := r.PathValue("username")

	// Prevent disabling admin
	if username == "admin" {
		sendError(w, http.StatusBadRequest, "cannot disable admin user")
		return
	}

	result, err := s.db.Conn().Exec("UPDATE users SET is_active = ? WHERE username = ?", false, username)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to disable user")
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		sendError(w, http.StatusNotFound, "user not found")
		return
	}

	var user db.User
	var createdAt, updatedAt string
	err = s.db.Conn().QueryRow(`
		SELECT id, username, type, is_active, created_at, updated_at
		FROM users
		WHERE username = ?
	`, username).Scan(&user.ID, &user.Username, &user.Type, &user.IsActive, &createdAt, &updatedAt)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to retrieve user")
		return
	}

	sendJSON(w, http.StatusOK, user)
}
