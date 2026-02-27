package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"hsduc.com/rag/config"
)

// TokenDetails holds info extracted from JWT
type TokenDetails struct {
	UserID    uint
	SessionID string
}

func GenerateTokens(userID uint, sessionID string) (string, string, error) {
	// For production, the secret key should be loaded once from env, but here we read it for simplicity
	secret := []byte(config.App.JWTSecretKey)
	if len(secret) == 0 {
		secret = []byte("default_secret_key") // Fallback
	}

	// Access Token (Expires in 15 minutes)
	accessClaims := jwt.MapClaims{
		"user_id":    userID,
		"session_id": sessionID,
		"exp":        time.Now().Add(15 * time.Minute).Unix(),
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessStr, err := accessToken.SignedString(secret)
	if err != nil {
		return "", "", err
	}

	// Refresh Token (Expires in 7 days)
	refreshClaims := jwt.MapClaims{
		"user_id":    userID,
		"session_id": sessionID,
		"exp":        time.Now().Add(7 * 24 * time.Hour).Unix(),
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshStr, err := refreshToken.SignedString(secret)
	if err != nil {
		return "", "", err
	}

	return accessStr, refreshStr, nil
}

func ValidateToken(tokenStr string) (*TokenDetails, error) {
	secret := []byte(config.App.JWTSecretKey)
	if len(secret) == 0 {
		secret = []byte("default_secret_key") // Fallback
	}

	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		// Verify signature method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return secret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		var details TokenDetails

		if userIDFloat, ok := claims["user_id"].(float64); ok {
			details.UserID = uint(userIDFloat)
		} else {
			return nil, errors.New("invalid user_id in token")
		}

		if sessionID, ok := claims["session_id"].(string); ok {
			details.SessionID = sessionID
		} else {
			return nil, errors.New("invalid session_id in token")
		}

		return &details, nil
	}

	return nil, errors.New("invalid token")
}
