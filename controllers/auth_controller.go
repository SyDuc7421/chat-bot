package controllers

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"hsduc.com/rag/database"
	"hsduc.com/rag/dtos"
	"hsduc.com/rag/models"
	"hsduc.com/rag/utils"
)

// @Summary      Register User
// @Description  Register a new user
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body body dtos.RegisterRequest true "Register Request"
// @Success      201  {object}  map[string]interface{}
// @Router       /api/v1/auth/register [post]
func Register(c *gin.Context) {
	var input dtos.RegisterRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hash, err := utils.HashPassword(input.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	user := models.User{
		Name:         input.Name,
		Email:        input.Email,
		PasswordHash: hash,
	}

	if err := database.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Email already exists"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully", "id": user.ID})
}

// @Summary      Login User
// @Description  Login and get JWT tokens
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body body dtos.LoginRequest true "Login Request"
// @Success      200  {object}  dtos.AuthResponse
// @Router       /api/v1/auth/login [post]
func Login(c *gin.Context) {
	var input dtos.LoginRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := database.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	if !utils.CheckPasswordHash(input.Password, user.PasswordHash) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	// Generate a session ID
	sessionID := uuid.New().String()

	accessStr, refreshStr, err := utils.GenerateTokens(user.ID, sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tokens"})
		return
	}

	// Save session in Redis with 7 days expiration
	err = database.Redis.Set(context.Background(), "session:"+sessionID, user.ID, 7*24*time.Hour).Err()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save session"})
		return
	}

	c.JSON(http.StatusOK, dtos.AuthResponse{
		AccessToken:  accessStr,
		RefreshToken: refreshStr,
	})
}

// @Summary      Refresh Token
// @Description  Refresh expired access token
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body body dtos.RefreshRequest true "Refresh Request"
// @Success      200  {object}  dtos.AuthResponse
// @Router       /api/v1/auth/refresh [post]
func RefreshToken(c *gin.Context) {
	var input dtos.RefreshRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Parsing token
	details, err := utils.ValidateToken(input.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired refresh token"})
		return
	}

	// Check if session still active in Redis
	sessionKey := "session:" + details.SessionID
	val, err := database.Redis.Get(context.Background(), sessionKey).Result()
	if err != nil || val == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Session expired, please login again"})
		return
	}

	// Invalidate old session (optional based on strict security preferences, or generate new ones and delete old)
	database.Redis.Del(context.Background(), sessionKey)

	// Create new session
	newSessionID := uuid.New().String()
	accessStr, refreshStr, err := utils.GenerateTokens(details.UserID, newSessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate new tokens"})
		return
	}

	// Save new session
	database.Redis.Set(context.Background(), "session:"+newSessionID, details.UserID, 7*24*time.Hour)

	c.JSON(http.StatusOK, dtos.AuthResponse{
		AccessToken:  accessStr,
		RefreshToken: refreshStr,
	})
}

// @Summary      Logout User
// @Description  Logout and invalidate session
// @Tags         Auth
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Router       /api/v1/auth/logout [post]
func Logout(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token missing"})
		return
	}

	tokenStr := parts[1]
	details, err := utils.ValidateToken(tokenStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	// Delete from Redis
	sessionKey := "session:" + details.SessionID
	database.Redis.Del(context.Background(), sessionKey)

	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}
