package server

import (
	"database/sql"
	"net/http"

	"github.com/simonski/task/internal/db"
)

// ConfigRequest represents a config set request
type ConfigRequest struct {
	Value       string  `json:"value"`
	Description *string `json:"description,omitempty"`
}

// handleListConfig returns all configuration
func (s *Server) handleListConfig(w http.ResponseWriter, r *http.Request) {
	rows, err := s.db.Conn().Query(`
		SELECT key, value, description, created_at, updated_at
		FROM config
		ORDER BY key
	`)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to query config")
		return
	}
	defer rows.Close()

	var configs []db.Config
	for rows.Next() {
		var cfg db.Config
		var description sql.NullString
		var createdAt, updatedAt string
		err := rows.Scan(&cfg.Key, &cfg.Value, &description, &createdAt, &updatedAt)
		if err != nil {
			sendError(w, http.StatusInternalServerError, "failed to scan config")
			return
		}
		cfg.CreatedAt = parseTimestamp(createdAt)
		cfg.UpdatedAt = parseTimestamp(updatedAt)
		if description.Valid {
			cfg.Description = &description.String
		}
		configs = append(configs, cfg)
	}

	sendJSON(w, http.StatusOK, configs)
}

// handleGetConfig returns a specific config value
func (s *Server) handleGetConfig(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")

	var cfg db.Config
	var description sql.NullString
	var createdAt, updatedAt string
	err := s.db.Conn().QueryRow(`
		SELECT key, value, description, created_at, updated_at
		FROM config
		WHERE key = ?
	`, key).Scan(&cfg.Key, &cfg.Value, &description, &createdAt, &updatedAt)
	if err == sql.ErrNoRows {
		sendError(w, http.StatusNotFound, "config key not found")
		return
	}
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to query config")
		return
	}
	cfg.CreatedAt = parseTimestamp(createdAt)
	cfg.UpdatedAt = parseTimestamp(updatedAt)
	if description.Valid {
		cfg.Description = &description.String
	}

	sendJSON(w, http.StatusOK, cfg)
}

// handleSetConfig sets a config value
func (s *Server) handleSetConfig(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")

	var req ConfigRequest
	if err := decodeJSON(r, &req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Value == "" {
		sendError(w, http.StatusBadRequest, "value is required")
		return
	}

	// Upsert config
	_, err := s.db.Conn().Exec(`
		INSERT INTO config (key, value, description)
		VALUES (?, ?, ?)
		ON CONFLICT(key) DO UPDATE SET value = ?, description = COALESCE(?, description)
	`, key, req.Value, req.Description, req.Value, req.Description)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to set config")
		return
	}

	var cfg db.Config
	var description sql.NullString
	var createdAt, updatedAt string
	err = s.db.Conn().QueryRow(`
		SELECT key, value, description, created_at, updated_at
		FROM config
		WHERE key = ?
	`, key).Scan(&cfg.Key, &cfg.Value, &description, &createdAt, &updatedAt)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to retrieve config")
		return
	}
	cfg.CreatedAt = parseTimestamp(createdAt)
	cfg.UpdatedAt = parseTimestamp(updatedAt)
	if description.Valid {
		cfg.Description = &description.String
	}

	sendJSON(w, http.StatusOK, cfg)
}

// handleDeleteConfig deletes a config value
func (s *Server) handleDeleteConfig(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")

	result, err := s.db.Conn().Exec("DELETE FROM config WHERE key = ?", key)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to delete config")
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		sendError(w, http.StatusNotFound, "config key not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
