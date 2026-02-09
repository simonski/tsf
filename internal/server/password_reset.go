package server

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/simonski/task/internal/db"
)

// ResetPasswordRequest represents a password reset request
type ResetPasswordRequest struct {
	NewPassword string `json:"new_password"`
}

// handleResetUserPassword resets a user's password (admin only)
func (s *Server) handleResetUserPassword(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("user_id")

	var req ResetPasswordRequest
	if err := decodeJSON(r, &req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.NewPassword == "" {
		sendError(w, http.StatusBadRequest, "new_password is required")
		return
	}

	// Verify user exists
	var exists int
	err := s.db.Conn().QueryRow(`SELECT 1 FROM users WHERE id = ?`, userID).Scan(&exists)
	if err == sql.ErrNoRows {
		sendError(w, http.StatusNotFound, "user not found")
		return
	}

	// Hash new password
	passwordHash, err := db.HashPassword(req.NewPassword)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to hash password")
		return
	}

	// Update password
	_, err = s.db.Conn().Exec(`
		UPDATE users SET password_hash = ?, updated_at = ? WHERE id = ?
	`, passwordHash, time.Now(), userID)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to update password")
		return
	}

	// Invalidate all sessions for this user
	s.db.Conn().Exec(`DELETE FROM sessions WHERE user_id = ?`, userID)

	w.WriteHeader(http.StatusNoContent)
}
