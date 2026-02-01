package db

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"golang.org/x/crypto/argon2"
)

const (
	// Argon2id parameters
	Memory      = 64 * 1024 // 64 MB
	Iterations  = 3
	Parallelism = 2
	SaltLength  = 16
	KeyLength   = 32
)

// HashPassword generates an Argon2id hash of the password
func HashPassword(password string) (string, error) {
	// Generate a random salt
	salt := make([]byte, SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	// Generate the hash
	hash := argon2.IDKey([]byte(password), salt, Iterations, Memory, Parallelism, KeyLength)

	// Encode salt and hash together
	// Format: $argon2id$v=19$m=65536,t=3,p=2$<salt>$<hash>
	encoded := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		Memory,
		Iterations,
		Parallelism,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	)

	return encoded, nil
}

// VerifyPassword checks if the provided password matches the hash
func VerifyPassword(password, encodedHash string) (bool, error) {
	// Parse the encoded hash by splitting on $
	parts := []string{}
	current := ""
	for _, ch := range encodedHash {
		if ch == '$' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(ch)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	
	if len(parts) != 5 {
		return false, fmt.Errorf("invalid hash format: expected 5 parts, got %d", len(parts))
	}
	
	if parts[0] != "argon2id" {
		return false, fmt.Errorf("unsupported hash algorithm: %s", parts[0])
	}
	
	// Parse version
	var version int
	if _, err := fmt.Sscanf(parts[1], "v=%d", &version); err != nil {
		return false, fmt.Errorf("failed to parse version: %w", err)
	}
	
	// Parse parameters
	var memory, iterations uint32
	var parallelism uint8
	if _, err := fmt.Sscanf(parts[2], "m=%d,t=%d,p=%d", &memory, &iterations, &parallelism); err != nil {
		return false, fmt.Errorf("failed to parse parameters: %w", err)
	}
	
	saltB64 := parts[3]
	hashB64 := parts[4]

	// Decode salt and hash
	salt, err := base64.RawStdEncoding.DecodeString(saltB64)
	if err != nil {
		return false, fmt.Errorf("failed to decode salt: %w", err)
	}

	hash, err := base64.RawStdEncoding.DecodeString(hashB64)
	if err != nil {
		return false, fmt.Errorf("failed to decode hash: %w", err)
	}

	// Generate hash for the provided password
	testHash := argon2.IDKey([]byte(password), salt, iterations, memory, parallelism, uint32(len(hash)))

	// Compare hashes
	if len(testHash) != len(hash) {
		return false, nil
	}

	for i := range testHash {
		if testHash[i] != hash[i] {
			return false, nil
		}
	}

	return true, nil
}

// GeneratePassword generates a random password
func GeneratePassword(length int) (string, error) {
	if length < 8 {
		length = 8
	}

	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate password: %w", err)
	}

	// Use URL-safe encoding and trim to desired length
	password := base64.URLEncoding.EncodeToString(bytes)
	if len(password) > length {
		password = password[:length]
	}

	return password, nil
}
