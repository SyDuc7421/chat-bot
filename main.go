package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"hsduc.com/rag/database"
	"hsduc.com/rag/routes"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using default environment variables")
	}

	// Initialize Database Connections
	database.ConnectMySQL()
	database.ConnectRedis()

	// Setup Routes
	r := routes.SetupRouter()

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting server on port %s...", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
