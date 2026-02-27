package main

import (
	"log"

	"github.com/joho/godotenv"
	"hsduc.com/rag/config"
	"hsduc.com/rag/database"
	"hsduc.com/rag/routes"
)

// @title           Chatbot RAG API
// @version         1.0
// @description     This is a RAG chatbot server.
// @host            localhost:8080
// @BasePath        /

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using default environment variables")
	}

	config.LoadConfig()

	// Initialize Database Connections
	database.ConnectMySQL()
	database.ConnectRedis()

	// Setup Routes
	r := routes.SetupRouter()

	// Start server
	port := config.App.Port
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting server on port %s...", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
