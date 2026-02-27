package utils

import (
	"testing"
)

func TestHashPassword(t *testing.T) {
	password := "my_secret_password"
	hash, err := HashPassword(password)

	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	if len(hash) == 0 {
		t.Errorf("Expected hashed password to not be empty")
	}
}

func TestCheckPasswordHash(t *testing.T) {
	password := "my_secret_password"
	wrongPassword := "wrong_password"

	hash, _ := HashPassword(password)

	if !CheckPasswordHash(password, hash) {
		t.Errorf("Expected CheckPasswordHash to return true for correct password")
	}

	if CheckPasswordHash(wrongPassword, hash) {
		t.Errorf("Expected CheckPasswordHash to return false for wrong password")
	}
}
