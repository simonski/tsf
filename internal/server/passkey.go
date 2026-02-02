package server

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
	"github.com/simonski/task/internal/db"
)

// WebAuthnUser implements the webauthn.User interface
type WebAuthnUser struct {
	user        *db.User
	credentials []webauthn.Credential
}

func (u *WebAuthnUser) WebAuthnID() []byte {
	return []byte(u.user.ID)
}

func (u *WebAuthnUser) WebAuthnName() string {
	return u.user.Username
}

func (u *WebAuthnUser) WebAuthnDisplayName() string {
	return u.user.Username
}

func (u *WebAuthnUser) WebAuthnIcon() string {
	return ""
}

func (u *WebAuthnUser) WebAuthnCredentials() []webauthn.Credential {
	return u.credentials
}

// PasskeyRegisterBeginRequest represents the request to start passkey registration
type PasskeyRegisterBeginRequest struct {
	Username string `json:"username"`
}

// PasskeyRegisterBeginResponse represents the response to start passkey registration
type PasskeyRegisterBeginResponse struct {
	SessionID string                       `json:"session_id"`
	Options   *protocol.CredentialCreation `json:"options"`
}

// PasskeyRegisterFinishRequest represents the request to complete passkey registration
type PasskeyRegisterFinishRequest struct {
	SessionID  string                               `json:"session_id"`
	Credential *protocol.CredentialCreationResponse `json:"credential"`
	DeviceName string                               `json:"device_name,omitempty"`
}

// PasskeyAuthBeginRequest represents the request to start passkey authentication
type PasskeyAuthBeginRequest struct {
	Username string `json:"username,omitempty"`
}

// PasskeyAuthBeginResponse represents the response to start passkey authentication
type PasskeyAuthBeginResponse struct {
	SessionID string                        `json:"session_id"`
	Options   *protocol.CredentialAssertion `json:"options"`
}

// PasskeyAuthFinishRequest represents the request to complete passkey authentication
type PasskeyAuthFinishRequest struct {
	SessionID  string                                `json:"session_id"`
	Credential *protocol.CredentialAssertionResponse `json:"credential"`
}

// getWebAuthnUser fetches a user and their credentials for WebAuthn
func (s *Server) getWebAuthnUser(username string) (*WebAuthnUser, error) {
	var user db.User
	var createdAt, updatedAt string
	err := s.db.Conn().QueryRow(`
		SELECT id, username, password_hash, type, is_active, created_at, updated_at
		FROM users
		WHERE username = ?
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

	user.CreatedAt = parseTimestamp(createdAt)
	user.UpdatedAt = parseTimestamp(updatedAt)

	// Fetch credentials
	rows, err := s.db.Conn().Query(`
		SELECT id, credential_id, public_key, attestation_type, aaguid, 
		       sign_count, clone_warning, transports, backup_eligible, backup_state
		FROM passkey_credentials
		WHERE user_id = ?
	`, user.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var credentials []webauthn.Credential
	for rows.Next() {
		var cred db.PasskeyCredential
		var transportsJSON *string
		err := rows.Scan(
			&cred.ID,
			&cred.CredentialID,
			&cred.PublicKey,
			&cred.AttestationType,
			&cred.AAGUID,
			&cred.SignCount,
			&cred.CloneWarning,
			&transportsJSON,
			&cred.BackupEligible,
			&cred.BackupState,
		)
		if err != nil {
			continue
		}

		var transports []string
		if transportsJSON != nil && *transportsJSON != "" {
			json.Unmarshal([]byte(*transportsJSON), &transports)
		}

		// Convert to protocol.AuthenticatorTransport
		var authTransports []protocol.AuthenticatorTransport
		for _, t := range transports {
			authTransports = append(authTransports, protocol.AuthenticatorTransport(t))
		}

		credentials = append(credentials, webauthn.Credential{
			ID:              cred.CredentialID,
			PublicKey:       cred.PublicKey,
			AttestationType: cred.AttestationType,
			Authenticator: webauthn.Authenticator{
				AAGUID:       cred.AAGUID,
				SignCount:    cred.SignCount,
				CloneWarning: cred.CloneWarning,
				Attachment:   protocol.Platform,
			},
			Transport: authTransports,
			Flags: webauthn.CredentialFlags{
				BackupEligible: cred.BackupEligible,
				BackupState:    cred.BackupState,
			},
		})
	}

	return &WebAuthnUser{
		user:        &user,
		credentials: credentials,
	}, nil
}

// handlePasskeyRegisterBegin starts the passkey registration process
func (s *Server) handlePasskeyRegisterBegin(w http.ResponseWriter, r *http.Request) {
	var req PasskeyRegisterBeginRequest
	if err := decodeJSON(r, &req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Username == "" {
		sendError(w, http.StatusBadRequest, "username required")
		return
	}

	// Check if user exists
	var userID string
	err := s.db.Conn().QueryRow("SELECT id FROM users WHERE username = ?", req.Username).Scan(&userID)
	if err != nil {
		sendError(w, http.StatusNotFound, "user not found")
		return
	}

	// Get user for WebAuthn
	user, err := s.getWebAuthnUser(req.Username)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to get user")
		return
	}

	// Generate registration options
	options, session, err := s.webAuthn.BeginRegistration(user)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to begin registration")
		return
	}

	// Store session in database
	sessionID := uuid.New().String()
	// Store the raw challenge bytes
	challengeBytes := []byte(session.Challenge)
	expiresAt := time.Now().Add(5 * time.Minute)

	_, err = s.db.Conn().Exec(`
		INSERT INTO webauthn_sessions (id, user_id, challenge, user_verification, expires_at, session_type, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, sessionID, user.user.ID, challengeBytes, session.UserVerification, expiresAt.Format("2006-01-02 15:04:05"), "registration", time.Now().Format("2006-01-02 15:04:05"))
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to create session")
		return
	}

	sendJSON(w, http.StatusOK, PasskeyRegisterBeginResponse{
		SessionID: sessionID,
		Options:   options,
	})
}

// handlePasskeyRegisterFinish completes the passkey registration process
func (s *Server) handlePasskeyRegisterFinish(w http.ResponseWriter, r *http.Request) {
	var req PasskeyRegisterFinishRequest
	if err := decodeJSON(r, &req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.SessionID == "" || req.Credential == nil {
		sendError(w, http.StatusBadRequest, "session_id and credential required")
		return
	}

	// Retrieve session
	var userID string
	var challengeJSON []byte
	var expiresAtStr string
	err := s.db.Conn().QueryRow(`
		SELECT user_id, challenge, expires_at FROM webauthn_sessions 
		WHERE id = ? AND session_type = 'registration'
	`, req.SessionID).Scan(&userID, &challengeJSON, &expiresAtStr)
	if err != nil {
		sendError(w, http.StatusUnauthorized, "invalid session")
		return
	}

	// Check if expired
	expiresAt := parseTimestamp(expiresAtStr)
	if time.Now().After(expiresAt) {
		s.db.Conn().Exec("DELETE FROM webauthn_sessions WHERE id = ?", req.SessionID)
		sendError(w, http.StatusUnauthorized, "session expired")
		return
	}

	// Get user
	var username string
	err = s.db.Conn().QueryRow("SELECT username FROM users WHERE id = ?", userID).Scan(&username)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "user not found")
		return
	}

	user, err := s.getWebAuthnUser(username)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to get user")
		return
	}

	// Reconstruct session data
	// The challenge is stored as raw bytes, convert to base64url string
	challenge := protocol.URLEncodedBase64(challengeJSON)
	sessionData := webauthn.SessionData{
		Challenge:        challenge.String(),
		UserID:           []byte(userID),
		UserVerification: protocol.VerificationPreferred,
	}

	// Parse the credential creation response
	parsedResponse, err := req.Credential.Parse()
	if err != nil {
		sendError(w, http.StatusBadRequest, "failed to parse credential")
		return
	}

	// Verify the credential
	credential, err := s.webAuthn.CreateCredential(user, sessionData, parsedResponse)
	if err != nil {
		sendError(w, http.StatusBadRequest, "failed to verify credential")
		return
	}

	// Store credential in database
	credID := uuid.New().String()
	transportsJSON, _ := json.Marshal(credential.Transport)
	deviceName := req.DeviceName
	if deviceName == "" {
		deviceName = "Unnamed Device"
	}

	_, err = s.db.Conn().Exec(`
		INSERT INTO passkey_credentials (
			id, user_id, credential_id, public_key, attestation_type, aaguid,
			sign_count, clone_warning, transports, backup_eligible, backup_state,
			device_name, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, credID, userID, credential.ID, credential.PublicKey, credential.AttestationType,
		credential.Authenticator.AAGUID, credential.Authenticator.SignCount,
		credential.Authenticator.CloneWarning, string(transportsJSON),
		credential.Flags.BackupEligible, credential.Flags.BackupState,
		deviceName, time.Now().Format("2006-01-02 15:04:05"))
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to store credential")
		return
	}

	// Delete session
	s.db.Conn().Exec("DELETE FROM webauthn_sessions WHERE id = ?", req.SessionID)

	sendJSON(w, http.StatusOK, map[string]string{
		"status":  "success",
		"message": "passkey registered successfully",
	})
}

// handlePasskeyAuthBegin starts the passkey authentication process
func (s *Server) handlePasskeyAuthBegin(w http.ResponseWriter, r *http.Request) {
	var req PasskeyAuthBeginRequest
	if err := decodeJSON(r, &req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	var user *WebAuthnUser
	var userID string

	if req.Username != "" {
		// User-specific authentication
		u, err := s.getWebAuthnUser(req.Username)
		if err != nil {
			sendError(w, http.StatusNotFound, "user not found")
			return
		}
		user = u
		userID = user.user.ID
	}

	// Generate authentication options
	var options *protocol.CredentialAssertion
	var session *webauthn.SessionData
	var err error

	if user != nil {
		options, session, err = s.webAuthn.BeginLogin(user)
	} else {
		// Discoverable credential (usernameless) flow
		options, session, err = s.webAuthn.BeginDiscoverableLogin()
	}

	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to begin authentication")
		return
	}

	// Store session in database
	sessionID := uuid.New().String()
	// Store the raw challenge bytes
	challengeBytes := []byte(session.Challenge)
	expiresAt := time.Now().Add(5 * time.Minute)

	var userIDPtr *string
	if userID != "" {
		userIDPtr = &userID
	}

	_, err = s.db.Conn().Exec(`
		INSERT INTO webauthn_sessions (id, user_id, challenge, user_verification, expires_at, session_type, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, sessionID, userIDPtr, challengeBytes, session.UserVerification, expiresAt.Format("2006-01-02 15:04:05"), "authentication", time.Now().Format("2006-01-02 15:04:05"))
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to create session")
		return
	}

	sendJSON(w, http.StatusOK, PasskeyAuthBeginResponse{
		SessionID: sessionID,
		Options:   options,
	})
}

// handlePasskeyAuthFinish completes the passkey authentication process
func (s *Server) handlePasskeyAuthFinish(w http.ResponseWriter, r *http.Request) {
	var req PasskeyAuthFinishRequest
	if err := decodeJSON(r, &req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.SessionID == "" || req.Credential == nil {
		sendError(w, http.StatusBadRequest, "session_id and credential required")
		return
	}

	// Retrieve session
	var userID sql.NullString
	var challengeJSON []byte
	var expiresAtStr string
	err := s.db.Conn().QueryRow(`
		SELECT user_id, challenge, expires_at FROM webauthn_sessions 
		WHERE id = ? AND session_type = 'authentication'
	`, req.SessionID).Scan(&userID, &challengeJSON, &expiresAtStr)
	if err != nil {
		sendError(w, http.StatusUnauthorized, "invalid session")
		return
	}

	// Check if expired
	expiresAt := parseTimestamp(expiresAtStr)
	if time.Now().After(expiresAt) {
		s.db.Conn().Exec("DELETE FROM webauthn_sessions WHERE id = ?", req.SessionID)
		sendError(w, http.StatusUnauthorized, "session expired")
		return
	}

	// Look up credential to find user
	var credUserID string
	var username string
	err = s.db.Conn().QueryRow(`
		SELECT pc.user_id, u.username 
		FROM passkey_credentials pc
		JOIN users u ON pc.user_id = u.id
		WHERE pc.credential_id = ?
	`, req.Credential.RawID).Scan(&credUserID, &username)
	if err != nil {
		sendError(w, http.StatusUnauthorized, "credential not found")
		return
	}

	// Get user with credentials
	user, err := s.getWebAuthnUser(username)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to get user")
		return
	}

	// Reconstruct session data
	// The challenge is stored as raw bytes, convert to base64url string
	challenge := protocol.URLEncodedBase64(challengeJSON)
	sessionData := webauthn.SessionData{
		Challenge:        challenge.String(),
		UserID:           []byte(credUserID),
		UserVerification: protocol.VerificationPreferred,
	}

	// Parse the credential assertion response
	parsedResponse, err := req.Credential.Parse()
	if err != nil {
		sendError(w, http.StatusBadRequest, "failed to parse credential")
		return
	}

	// Verify the assertion
	credential, err := s.webAuthn.ValidateLogin(user, sessionData, parsedResponse)
	if err != nil {
		sendError(w, http.StatusUnauthorized, "failed to verify credential")
		return
	}

	// Update credential sign count and last used
	_, err = s.db.Conn().Exec(`
		UPDATE passkey_credentials 
		SET sign_count = ?, last_used_at = ?, clone_warning = ?
		WHERE credential_id = ?
	`, credential.Authenticator.SignCount, time.Now().Format("2006-01-02 15:04:05"),
		credential.Authenticator.CloneWarning, credential.ID)
	if err != nil {
		// Non-fatal, log and continue
	}

	// Delete session
	s.db.Conn().Exec("DELETE FROM webauthn_sessions WHERE id = ?", req.SessionID)

	// Generate JWT token
	token, tokenExpiresAt, err := generateToken(user.user, 15*time.Minute)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	// Generate refresh token
	refreshToken := generateRefreshToken()
	refreshExpiresAt := time.Now().Add(7 * 24 * time.Hour)
	_, err = s.db.Conn().Exec(`
		INSERT OR REPLACE INTO refresh_tokens (token, user_id, expires_at, created_at)
		VALUES (?, ?, ?, ?)
	`, refreshToken, user.user.ID, refreshExpiresAt.Format("2006-01-02 15:04:05"), time.Now().Format("2006-01-02 15:04:05"))

	sendJSON(w, http.StatusOK, LoginResponse{
		Token:        token,
		RefreshToken: refreshToken,
		ExpiresAt:    tokenExpiresAt,
		User: UserInfo{
			ID:       user.user.ID,
			Username: user.user.Username,
			Type:     user.user.Type,
		},
	})
}

// handlePasskeyList lists all passkeys for the authenticated user
func (s *Server) handlePasskeyList(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r.Context())
	if user == nil {
		sendError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	rows, err := s.db.Conn().Query(`
		SELECT id, credential_id, attestation_type, device_name, created_at, last_used_at
		FROM passkey_credentials
		WHERE user_id = ?
		ORDER BY created_at DESC
	`, user.ID)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to list passkeys")
		return
	}
	defer rows.Close()

	type PasskeyInfo struct {
		ID              string  `json:"id"`
		CredentialID    string  `json:"credential_id"`
		AttestationType string  `json:"attestation_type"`
		DeviceName      *string `json:"device_name"`
		CreatedAt       string  `json:"created_at"`
		LastUsedAt      *string `json:"last_used_at"`
	}

	var passkeys []PasskeyInfo
	for rows.Next() {
		var p PasskeyInfo
		var credID []byte
		err := rows.Scan(&p.ID, &credID, &p.AttestationType, &p.DeviceName, &p.CreatedAt, &p.LastUsedAt)
		if err != nil {
			continue
		}
		p.CredentialID = protocol.URLEncodedBase64(credID).String()
		passkeys = append(passkeys, p)
	}

	if passkeys == nil {
		passkeys = []PasskeyInfo{}
	}

	sendJSON(w, http.StatusOK, passkeys)
}

// handlePasskeyDelete deletes a passkey for the authenticated user
func (s *Server) handlePasskeyDelete(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r.Context())
	if user == nil {
		sendError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	passkeyID := r.URL.Query().Get("id")
	if passkeyID == "" {
		sendError(w, http.StatusBadRequest, "passkey id required")
		return
	}

	result, err := s.db.Conn().Exec(`
		DELETE FROM passkey_credentials WHERE id = ? AND user_id = ?
	`, passkeyID, user.ID)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to delete passkey")
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		sendError(w, http.StatusNotFound, "passkey not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
