package db

import (
	"testing"
)

func TestHashPassword(t *testing.T) {
	password := "testpassword123"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	if hash == "" {
		t.Error("HashPassword returned empty string")
	}

	if hash == password {
		t.Error("HashPassword returned plaintext password")
	}
}

func TestVerifyPassword(t *testing.T) {
	password := "testpassword123"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	// Correct password should verify
	valid, err := VerifyPassword(password, hash)
	if err != nil {
		t.Fatalf("VerifyPassword failed: %v", err)
	}
	if !valid {
		t.Error("VerifyPassword returned false for correct password")
	}

	// Wrong password should not verify
	valid, err = VerifyPassword("wrongpassword", hash)
	if err != nil {
		t.Fatalf("VerifyPassword failed: %v", err)
	}
	if valid {
		t.Error("VerifyPassword returned true for incorrect password")
	}
}

func TestGeneratePassword(t *testing.T) {
	// Test various lengths
	lengths := []int{8, 12, 16, 24, 32}

	for _, length := range lengths {
		password, err := GeneratePassword(length)
		if err != nil {
			t.Fatalf("GeneratePassword(%d) failed: %v", length, err)
		}

		if len(password) < 8 {
			t.Errorf("GeneratePassword(%d) returned password shorter than 8 characters: %d", length, len(password))
		}

		if password == "" {
			t.Errorf("GeneratePassword(%d) returned empty string", length)
		}
	}

	// Test that generated passwords are unique
	pwd1, _ := GeneratePassword(16)
	pwd2, _ := GeneratePassword(16)

	if pwd1 == pwd2 {
		t.Error("GeneratePassword returned identical passwords")
	}
}
