package server

import (
	"context"
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/simonski/task/internal/db"
)

type contextKey string

const (
	contextKeyUser contextKey = "user"
)

// withAuth middleware checks Basic Auth credentials
func (s *Server) withAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get Authorization header
		auth := r.Header.Get("Authorization")
		if auth == "" {
			sendError(w, http.StatusUnauthorized, "missing authorization header")
			return
		}

		// Parse Basic Auth
		if !strings.HasPrefix(auth, "Basic ") {
			sendError(w, http.StatusUnauthorized, "invalid authorization scheme")
			return
		}

		// Decode credentials
		payload, err := base64.StdEncoding.DecodeString(auth[6:])
		if err != nil {
			sendError(w, http.StatusUnauthorized, "invalid authorization format")
			return
		}

		// Split username:password
		credentials := string(payload)
		parts := strings.SplitN(credentials, ":", 2)
		if len(parts) != 2 {
			sendError(w, http.StatusUnauthorized, "invalid credentials format")
			return
		}

		username := parts[0]
		password := parts[1]

		// Query user from database
		var user db.User
		err = s.db.Conn().QueryRow(`
			SELECT id, username, password_hash, type, is_active, created_at, updated_at
			FROM users
			WHERE username = ?
		`, username).Scan(
			&user.ID,
			&user.Username,
			&user.PasswordHash,
			&user.Type,
			&user.IsActive,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			sendError(w, http.StatusUnauthorized, "invalid credentials")
			return
		}

		// Check if user is active
		if !user.IsActive {
			sendError(w, http.StatusUnauthorized, "user account is disabled")
			return
		}

		// Verify password
		valid, err := db.VerifyPassword(password, user.PasswordHash)
		if err != nil || !valid {
			sendError(w, http.StatusUnauthorized, "invalid credentials")
			return
		}

		// Add user to context
		ctx := context.WithValue(r.Context(), contextKeyUser, &user)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// withAdmin middleware checks if user is admin
func (s *Server) withAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := getUserFromContext(r.Context())
		if user == nil {
			sendError(w, http.StatusUnauthorized, "authentication required")
			return
		}
		if user.Username != "admin" {
			sendError(w, http.StatusForbidden, "admin access required")
			return
		}
		next.ServeHTTP(w, r)
	}
}

// getUserFromContext retrieves the authenticated user from context
func getUserFromContext(ctx context.Context) *db.User {
	user, ok := ctx.Value(contextKeyUser).(*db.User)
	if !ok {
		return nil
	}
	return user
}
