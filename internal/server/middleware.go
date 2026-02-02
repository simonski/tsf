package server

import (
	"context"
	"encoding/base64"
	"errors"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/simonski/task/internal/db"
)

type contextKey string

const (
	contextKeyUser contextKey = "user"
)

// withAuth middleware checks Basic Auth or Bearer token credentials
func (s *Server) withAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get Authorization header
		auth := r.Header.Get("Authorization")
		if auth == "" {
			sendError(w, http.StatusUnauthorized, "missing authorization header")
			return
		}

		var user *db.User
		var err error

		if strings.HasPrefix(auth, "Bearer ") {
			// JWT token authentication
			user, err = s.authenticateJWT(auth[7:])
			if err != nil {
				sendError(w, http.StatusUnauthorized, "invalid or expired token")
				return
			}
		} else if strings.HasPrefix(auth, "Basic ") {
			// Basic Auth
			user, err = s.authenticateBasic(auth[6:])
			if err != nil {
				sendError(w, http.StatusUnauthorized, err.Error())
				return
			}
		} else {
			sendError(w, http.StatusUnauthorized, "invalid authorization scheme")
			return
		}

		// Ensure user was successfully authenticated
		if user == nil {
			sendError(w, http.StatusUnauthorized, "authentication failed")
			return
		}

		// Check if user is active
		if !user.IsActive {
			sendError(w, http.StatusUnauthorized, "user account is disabled")
			return
		}

		// Add user to context
		ctx := context.WithValue(r.Context(), contextKeyUser, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// authenticateJWT validates a JWT token and returns the user
func (s *Server) authenticateJWT(tokenString string) (*db.User, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return JWTSecret, nil
	})
	if err != nil || !token.Valid {
		return nil, err
	}

	// Get user from database
	var user db.User
	var createdAt, updatedAt string
	err = s.db.Conn().QueryRow(`
		SELECT id, username, password_hash, type, is_active, created_at, updated_at
		FROM users WHERE id = ?
	`, claims.UserID).Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.Type,
		&user.IsActive,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// authenticateBasic validates Basic Auth credentials and returns the user
func (s *Server) authenticateBasic(encoded string) (*db.User, error) {
	payload, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, err
	}

	credentials := string(payload)
	parts := strings.SplitN(credentials, ":", 2)
	if len(parts) != 2 {
		return nil, errors.New("invalid credentials format")
	}

	username := parts[0]
	password := parts[1]

	var user db.User
	var createdAt, updatedAt string
	err = s.db.Conn().QueryRow(`
		SELECT id, username, password_hash, type, is_active, created_at, updated_at
		FROM users WHERE username = ?
	`, username).Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.Type,
		&user.IsActive,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return nil, err
	}

	valid, err := db.VerifyPassword(password, user.PasswordHash)
	if err != nil || !valid {
		return nil, err
	}

	return &user, nil
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
