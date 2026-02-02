package server

import (
	"encoding/json"
	"path/filepath"
	"testing"
	"time"

	"github.com/simonski/task/internal/db"
)

func setupTestDB(t *testing.T) *db.DB {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	
	database, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	
	if err := database.InitSchema(); err != nil {
		t.Fatalf("Failed to initialize schema: %v", err)
	}
	
	return database
}

func TestPasskeySessionStorage(t *testing.T) {
	// Create test database
	database := setupTestDB(t)
	defer database.Close()

	// Create test user
	userID := "test-user-id"
	username := "testuser"
	passwordHash, _ := db.HashPassword("testpass")
	
	_, err := database.Conn().Exec(`
		INSERT INTO users (id, username, password_hash, type, is_active, created_at, updated_at)
		VALUES (?, ?, ?, 'human', 1, datetime('now'), datetime('now'))
	`, userID, username, passwordHash)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create WebAuthn session
	sessionID := "test-session-id"
	challenge := []byte("test-challenge-bytes")
	challengeJSON, _ := json.Marshal(challenge)
	expiresAt := time.Now().Add(5 * time.Minute)

	_, err = database.Conn().Exec(`
		INSERT INTO webauthn_sessions (id, user_id, challenge, user_verification, expires_at, session_type, created_at)
		VALUES (?, ?, ?, 'preferred', ?, 'registration', datetime('now'))
	`, sessionID, userID, challengeJSON, expiresAt.Format("2006-01-02 15:04:05"))
	if err != nil {
		t.Fatalf("Failed to create WebAuthn session: %v", err)
	}

	// Verify session was stored
	var storedSessionID string
	var storedUserID string
	var storedChallengeJSON []byte
	var storedExpiresAt string

	err = database.Conn().QueryRow(`
		SELECT id, user_id, challenge, expires_at 
		FROM webauthn_sessions 
		WHERE id = ? AND session_type = 'registration'
	`, sessionID).Scan(&storedSessionID, &storedUserID, &storedChallengeJSON, &storedExpiresAt)
	
	if err != nil {
		t.Fatalf("Failed to retrieve WebAuthn session: %v", err)
	}

	if storedSessionID != sessionID {
		t.Errorf("Expected session ID %s, got %s", sessionID, storedSessionID)
	}

	if storedUserID != userID {
		t.Errorf("Expected user ID %s, got %s", userID, storedUserID)
	}

	// Verify challenge can be deserialized
	var retrievedChallenge []byte
	err = json.Unmarshal(storedChallengeJSON, &retrievedChallenge)
	if err != nil {
		t.Errorf("Failed to unmarshal challenge: %v", err)
	}

	if string(retrievedChallenge) != string(challenge) {
		t.Errorf("Challenge mismatch: expected %v, got %v", challenge, retrievedChallenge)
	}
}

func TestPasskeyCredentialStorage(t *testing.T) {
	// Create test database
	database := setupTestDB(t)
	defer database.Close()

	// Create test user
	userID := "test-user-id"
	username := "testuser"
	passwordHash, _ := db.HashPassword("testpass")
	
	_, err := database.Conn().Exec(`
		INSERT INTO users (id, username, password_hash, type, is_active, created_at, updated_at)
		VALUES (?, ?, ?, 'human', 1, datetime('now'), datetime('now'))
	`, userID, username, passwordHash)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create passkey credential
	credID := "test-cred-id"
	credentialID := []byte("test-credential-id-bytes")
	publicKey := []byte("test-public-key-bytes")
	deviceName := "Test Device"

	_, err = database.Conn().Exec(`
		INSERT INTO passkey_credentials (
			id, user_id, credential_id, public_key, attestation_type, aaguid,
			sign_count, clone_warning, transports, backup_eligible, backup_state,
			device_name, created_at
		) VALUES (?, ?, ?, ?, 'none', ?, 0, 0, '[]', 0, 0, ?, datetime('now'))
	`, credID, userID, credentialID, publicKey, make([]byte, 16), deviceName)
	
	if err != nil {
		t.Fatalf("Failed to create passkey credential: %v", err)
	}

	// Verify credential was stored
	var storedCredID string
	var storedUserID string
	var storedCredentialID []byte
	var storedDeviceName string

	err = database.Conn().QueryRow(`
		SELECT id, user_id, credential_id, device_name
		FROM passkey_credentials
		WHERE id = ?
	`, credID).Scan(&storedCredID, &storedUserID, &storedCredentialID, &storedDeviceName)
	
	if err != nil {
		t.Fatalf("Failed to retrieve passkey credential: %v", err)
	}

	if storedCredID != credID {
		t.Errorf("Expected credential ID %s, got %s", credID, storedCredID)
	}

	if storedUserID != userID {
		t.Errorf("Expected user ID %s, got %s", userID, storedUserID)
	}

	if string(storedCredentialID) != string(credentialID) {
		t.Errorf("Credential ID mismatch")
	}

	if storedDeviceName != deviceName {
		t.Errorf("Expected device name %s, got %s", deviceName, storedDeviceName)
	}
}

func TestPasskeyCredentialLookup(t *testing.T) {
	// Create test database
	database := setupTestDB(t)
	defer database.Close()

	// Create test user
	userID := "test-user-id"
	username := "testuser"
	passwordHash, _ := db.HashPassword("testpass")
	
	_, err := database.Conn().Exec(`
		INSERT INTO users (id, username, password_hash, type, is_active, created_at, updated_at)
		VALUES (?, ?, ?, 'human', 1, datetime('now'), datetime('now'))
	`, userID, username, passwordHash)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create multiple passkey credentials
	for i := 1; i <= 3; i++ {
		credID := "test-cred-" + string(rune('0'+i))
		credentialID := []byte("credential-" + string(rune('0'+i)))
		publicKey := []byte("public-key-" + string(rune('0'+i)))

		_, err = database.Conn().Exec(`
			INSERT INTO passkey_credentials (
				id, user_id, credential_id, public_key, attestation_type, aaguid,
				sign_count, clone_warning, transports, backup_eligible, backup_state,
				device_name, created_at
			) VALUES (?, ?, ?, ?, 'none', ?, 0, 0, '[]', 0, 0, 'Device', datetime('now'))
		`, credID, userID, credentialID, publicKey, make([]byte, 16))
		
		if err != nil {
			t.Fatalf("Failed to create passkey credential %d: %v", i, err)
		}
	}

	// Verify we can list all credentials for user
	rows, err := database.Conn().Query(`
		SELECT id, credential_id
		FROM passkey_credentials
		WHERE user_id = ?
		ORDER BY created_at
	`, userID)
	if err != nil {
		t.Fatalf("Failed to query passkey credentials: %v", err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var id string
		var credID []byte
		err := rows.Scan(&id, &credID)
		if err != nil {
			t.Errorf("Failed to scan row: %v", err)
		}
		count++
	}

	if count != 3 {
		t.Errorf("Expected 3 credentials, got %d", count)
	}
}

func TestPasskeySessionExpiration(t *testing.T) {
	// Create test database
	database := setupTestDB(t)
	defer database.Close()

	// Create test user
	userID := "test-user-id"
	username := "testuser"
	passwordHash, _ := db.HashPassword("testpass")
	
	_, err := database.Conn().Exec(`
		INSERT INTO users (id, username, password_hash, type, is_active, created_at, updated_at)
		VALUES (?, ?, ?, 'human', 1, datetime('now'), datetime('now'))
	`, userID, username, passwordHash)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create expired session
	sessionID := "expired-session"
	challenge := []byte("test-challenge")
	challengeJSON, _ := json.Marshal(challenge)
	expiresAt := time.Now().Add(-1 * time.Hour) // Expired 1 hour ago

	_, err = database.Conn().Exec(`
		INSERT INTO webauthn_sessions (id, user_id, challenge, user_verification, expires_at, session_type, created_at)
		VALUES (?, ?, ?, 'preferred', ?, 'registration', datetime('now'))
	`, sessionID, userID, challengeJSON, expiresAt.Format("2006-01-02 15:04:05"))
	if err != nil {
		t.Fatalf("Failed to create expired session: %v", err)
	}

	// Retrieve session and check expiration
	var storedExpiresAt string
	err = database.Conn().QueryRow(`
		SELECT expires_at FROM webauthn_sessions WHERE id = ?
	`, sessionID).Scan(&storedExpiresAt)
	
	if err != nil {
		t.Fatalf("Failed to retrieve session: %v", err)
	}

	// Parse timestamp (using simple format for SQLite)
	expiry, err := time.Parse("2006-01-02 15:04:05", storedExpiresAt)
	if err != nil {
		t.Fatalf("Failed to parse expiry time: %v", err)
	}

	if !time.Now().After(expiry) {
		t.Error("Session should be expired but isn't")
	}

	// Clean up expired session
	result, err := database.Conn().Exec(`
		DELETE FROM webauthn_sessions WHERE expires_at < datetime('now')
	`)
	if err != nil {
		t.Fatalf("Failed to delete expired session: %v", err)
	}

	affected, _ := result.RowsAffected()
	if affected != 1 {
		t.Errorf("Expected 1 expired session to be deleted, got %d", affected)
	}
}
