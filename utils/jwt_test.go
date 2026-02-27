package utils

import (
	"testing"

	"hsduc.com/rag/config"
)

func TestGenerateAndValidateTokens(t *testing.T) {
	// Setup mock config
	config.App = &config.Config{
		JWTSecretKey: "test_secret_key",
	}

	userID := uint(123)
	sessionID := "test_session_id"

	// Generate tokens
	accessToken, refreshToken, err := GenerateTokens(userID, sessionID)
	if err != nil {
		t.Fatalf("Expected no error while generating tokens, got %v", err)
	}

	if len(accessToken) == 0 {
		t.Errorf("Expected non-empty access token")
	}

	if len(refreshToken) == 0 {
		t.Errorf("Expected non-empty refresh token")
	}

	// Validate the access token
	details, err := ValidateToken(accessToken)
	if err != nil {
		t.Fatalf("Expected no error while validating token, got %v", err)
	}

	if details == nil {
		t.Fatalf("Expected token details, got nil")
	}

	if details.UserID != userID {
		t.Errorf("Expected UserID %d, got %d", userID, details.UserID)
	}

	if details.SessionID != sessionID {
		t.Errorf("Expected SessionID %s, got %s", sessionID, details.SessionID)
	}
}

func TestValidateInvalidToken(t *testing.T) {
	config.App = &config.Config{
		JWTSecretKey: "test_secret_key",
	}

	invalidToken := "invalid.jwt.token"

	details, err := ValidateToken(invalidToken)
	if err == nil {
		t.Errorf("Expected error for invalid token, got nil")
	}

	if details != nil {
		t.Errorf("Expected details to be nil for invalid token, got %v", details)
	}
}
