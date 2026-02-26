package controllers

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"hsduc.com/rag/database"
	"hsduc.com/rag/models"
)

func SetupTestDB() {
	// Initialize in-memory SQLite for testing
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		panic("Failed to connect database")
	}
	db.Migrator().DropTable(&models.Conversation{}, &models.Message{})
	db.AutoMigrate(&models.Conversation{}, &models.Message{})
	database.DB = db

	// Initialize mock redis (using go-redis mock or just simple connect if available)
	// For simplicity, we just use a placeholder to avoid nil pointer panic if redis is accessed.
	// You might want to use actual mocking libraries like "github.com/go-redis/redismock/v9" later
	// For now, let's connect to the local redis or ignore it if tests fail.
	// To prevent test failures, we can leave Redis nil and avoid writing to it during these tests, or use a mock.
	database.Redis = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	// Just testing the ping so we know if it's there
	_, _ = database.Redis.Ping(context.Background()).Result()
}

func GetTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.Default()
}
