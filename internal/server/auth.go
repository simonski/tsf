package server

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/simonski/task/internal/db"
)

// JWTSecret holds the signing key for JWT tokens
var JWTSecret []byte

func init() {
	// Generate a random secret if not set
	JWTSecret = make([]byte, 32)
	rand.Read(JWTSecret)
}

// SetJWTSecret allows setting a custom JWT secret
func SetJWTSecret(secret string) {
	JWTSecret = []byte(secret)
}

// Claims represents JWT claims
type Claims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Type     string `json:"type"`
	jwt.RegisteredClaims
}

// LoginRequest represents a login request body
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse represents a login response
type LoginResponse struct {
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	User         UserInfo  `json:"user"`
}

// UserInfo represents user info in login response
type UserInfo struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Type     string `json:"type"`
}

// RefreshRequest represents a token refresh request
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// generateToken creates a new JWT token for a user
func generateToken(user *db.User, duration time.Duration) (string, time.Time, error) {
	expiresAt := time.Now().Add(duration)
	claims := &Claims{
		UserID:   user.ID,
		Username: user.Username,
		Type:     user.Type,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "task-server",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(JWTSecret)
	return signedToken, expiresAt, err
}

// generateRefreshToken creates a random refresh token
func generateRefreshToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// handleLogin handles user login and returns JWT tokens
func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := decodeJSON(r, &req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Username == "" || req.Password == "" {
		sendError(w, http.StatusBadRequest, "username and password required")
		return
	}

	// Query user from database
	var user db.User
	var createdAt, updatedAt string
	err := s.db.Conn().QueryRow(`
		SELECT id, username, password_hash, type, is_active, created_at, updated_at
		FROM users
		WHERE username = ?
	`, req.Username).Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.Type,
		&user.IsActive,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		sendError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	if !user.IsActive {
		sendError(w, http.StatusUnauthorized, "user account is disabled")
		return
	}

	// Verify password
	valid, err := db.VerifyPassword(req.Password, user.PasswordHash)
	if err != nil || !valid {
		sendError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	// Generate JWT token (expires in 15 minutes)
	token, expiresAt, err := generateToken(&user, 15*time.Minute)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	// Generate refresh token (expires in 7 days, stored in sessions table)
	refreshToken := generateRefreshToken()
	refreshExpiresAt := time.Now().Add(7 * 24 * time.Hour)

	// Store session token in database (use SQLite format for timestamps)
	sessionID := uuid.New().String()
	_, err = s.db.Conn().Exec(`
		INSERT INTO sessions (id, token, user_id, expires_at, created_at)
		VALUES (?, ?, ?, ?, ?)
	`, sessionID, refreshToken, user.ID, refreshExpiresAt.Format("2006-01-02 15:04:05"), time.Now().Format("2006-01-02 15:04:05"))
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to create session")
		return
	}

	sendJSON(w, http.StatusOK, LoginResponse{
		Token:        token,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
		User: UserInfo{
			ID:       user.ID,
			Username: user.Username,
			Type:     user.Type,
		},
	})
}

// handleRefresh handles token refresh
func (s *Server) handleRefresh(w http.ResponseWriter, r *http.Request) {
	var req RefreshRequest
	if err := decodeJSON(r, &req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.RefreshToken == "" {
		sendError(w, http.StatusBadRequest, "refresh_token required")
		return
	}

	// Look up refresh token in sessions table
	var userID string
	var expiresAt string
	err := s.db.Conn().QueryRow(`
		SELECT user_id, expires_at FROM sessions WHERE token = ?
	`, req.RefreshToken).Scan(&userID, &expiresAt)
	if err != nil {
		sendError(w, http.StatusUnauthorized, "invalid refresh token")
		return
	}

	// Check if expired
	expiry := parseTimestamp(expiresAt)
	if time.Now().After(expiry) {
		// Delete expired token
		s.db.Conn().Exec("DELETE FROM sessions WHERE token = ?", req.RefreshToken)
		sendError(w, http.StatusUnauthorized, "refresh token expired")
		return
	}

	// Get user
	var user db.User
	var createdAtStr, updatedAtStr string
	err = s.db.Conn().QueryRow(`
		SELECT id, username, password_hash, type, is_active, created_at, updated_at
		FROM users WHERE id = ?
	`, userID).Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.Type,
		&user.IsActive,
		&createdAtStr,
		&updatedAtStr,
	)
	if err != nil {
		sendError(w, http.StatusUnauthorized, "user not found")
		return
	}

	if !user.IsActive {
		sendError(w, http.StatusUnauthorized, "user account is disabled")
		return
	}

	// Generate new JWT token
	token, newExpiresAt, err := generateToken(&user, 15*time.Minute)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	// Optionally rotate refresh token
	newRefreshToken := generateRefreshToken()
	refreshExpiresAt := time.Now().Add(7 * 24 * time.Hour)

	// Delete old token and insert new
	s.db.Conn().Exec("DELETE FROM sessions WHERE token = ?", req.RefreshToken)
	sessionID := uuid.New().String()
	s.db.Conn().Exec(`
		INSERT INTO sessions (id, token, user_id, expires_at, created_at)
		VALUES (?, ?, ?, ?, ?)
	`, sessionID, newRefreshToken, user.ID, refreshExpiresAt.Format("2006-01-02 15:04:05"), time.Now().Format("2006-01-02 15:04:05"))

	sendJSON(w, http.StatusOK, LoginResponse{
		Token:        token,
		RefreshToken: newRefreshToken,
		ExpiresAt:    newExpiresAt,
		User: UserInfo{
			ID:       user.ID,
			Username: user.Username,
			Type:     user.Type,
		},
	})
}

// handleLogout invalidates the refresh token
func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	var req RefreshRequest
	if err := decodeJSON(r, &req); err != nil {
		// Just return success even if no body
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if req.RefreshToken != "" {
		s.db.Conn().Exec("DELETE FROM sessions WHERE token = ?", req.RefreshToken)
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleMe returns the current user info
func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r.Context())
	if user == nil {
		sendError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	sendJSON(w, http.StatusOK, UserInfo{
		ID:       user.ID,
		Username: user.Username,
		Type:     user.Type,
	})
}
