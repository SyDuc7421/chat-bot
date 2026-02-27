package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"hsduc.com/rag/config"
	"hsduc.com/rag/database"
	"hsduc.com/rag/models"
	"hsduc.com/rag/utils"
)

func setupAuthConfig() {
	config.App = &config.Config{
		JWTSecretKey: "test_secret_key",
	}
}

func TestRegister(t *testing.T) {
	SetupTestDB()
	setupAuthConfig()
	r := GetTestRouter()
	r.POST("/auth/register", Register)

	payload := []byte(`{"name":"Test User","email":"test@example.com","password":"password123"}`)
	req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	var user models.User
	database.DB.First(&user, "email = ?", "test@example.com")
	assert.Equal(t, "Test User", user.Name)

	// Test duplicate email
	wDup := httptest.NewRecorder()
	reqDup, _ := http.NewRequest("POST", "/auth/register", bytes.NewBuffer(payload))
	reqDup.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(wDup, reqDup)
	assert.Equal(t, http.StatusConflict, wDup.Code)
}

func TestLogin(t *testing.T) {
	SetupTestDB()
	setupAuthConfig()

	hash, _ := utils.HashPassword("password123")
	user := models.User{Name: "Login User", Email: "login@example.com", PasswordHash: hash}
	database.DB.Create(&user)

	r := GetTestRouter()
	r.POST("/auth/login", Login)

	payload := []byte(`{"email":"login@example.com","password":"password123"}`)
	req, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.NotEmpty(t, response["access_token"])
	assert.NotEmpty(t, response["refresh_token"])

	// Test invalid password
	payloadInvalid := []byte(`{"email":"login@example.com","password":"wrongpassword"}`)
	reqInvalid, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(payloadInvalid))
	reqInvalid.Header.Set("Content-Type", "application/json")
	wInvalid := httptest.NewRecorder()

	r.ServeHTTP(wInvalid, reqInvalid)
	assert.Equal(t, http.StatusUnauthorized, wInvalid.Code)
}

func TestRefreshToken(t *testing.T) {
	SetupTestDB()
	setupAuthConfig()

	r := GetTestRouter()
	r.POST("/auth/refresh", RefreshToken)

	sessionID := uuid.New().String()
	userID := uint(1)
	_, refreshStr, _ := utils.GenerateTokens(userID, sessionID)

	database.Redis.Set(context.Background(), "session:"+sessionID, userID, 1*time.Minute)

	payload := []byte(`{"refresh_token":"` + refreshStr + `"}`)
	req, _ := http.NewRequest("POST", "/auth/refresh", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.NotEmpty(t, response["access_token"])
	assert.NotEmpty(t, response["refresh_token"])
}

func TestLogout(t *testing.T) {
	SetupTestDB()
	setupAuthConfig()

	r := GetTestRouter()
	r.POST("/auth/logout", Logout)

	sessionID := uuid.New().String()
	userID := uint(1)
	accessStr, _, _ := utils.GenerateTokens(userID, sessionID)

	database.Redis.Set(context.Background(), "session:"+sessionID, userID, 1*time.Minute)

	req, _ := http.NewRequest("POST", "/auth/logout", nil)
	req.Header.Set("Authorization", "Bearer "+accessStr)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Check that session is deleted
	val, _ := database.Redis.Get(context.Background(), "session:"+sessionID).Result()
	assert.Empty(t, val)
}
