package routes

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"hsduc.com/rag/config"
	"hsduc.com/rag/controllers"
	_ "hsduc.com/rag/docs"
	"hsduc.com/rag/middleware"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	origins := []string{"http://localhost:3000", "http://localhost:5173"}
	if config.App.FRONTEND_BASE_URL != "" && config.App.FRONTEND_BASE_URL != "http://localhost:3000" && config.App.FRONTEND_BASE_URL != "http://localhost:5173" {
		origins = append(origins, config.App.FRONTEND_BASE_URL)
	}

	r.Use(cors.New(cors.Config{
		AllowOrigins:     origins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Swagger route
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Health Check Route
	r.GET("/health", controllers.HealthCheck)

	api := r.Group("/api/v1")
	{
		// Auth Routes
		auth := api.Group("/auth")
		{
			auth.POST("/register", controllers.Register)
			auth.POST("/login", controllers.Login)
			auth.POST("/refresh", controllers.RefreshToken)
		}

		// Protected Routes
		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware())
		{
			protected.POST("/auth/logout", controllers.Logout)
			protected.GET("/auth/me", controllers.GetMe)

			// Conversation Routes
			protected.POST("/conversations", controllers.CreateConversation)
			protected.GET("/conversations", controllers.GetConversations)
			protected.GET("/conversations/:id", controllers.GetConversation)
			protected.PUT("/conversations/:id", controllers.UpdateConversation)
			protected.DELETE("/conversations/:id", controllers.DeleteConversation)

			// Message Routes
			protected.POST("/messages", controllers.CreateMessage)
			protected.GET("/messages", controllers.GetMessages) // Use query ?conversation_id=X
			protected.GET("/messages/:id", controllers.GetMessage)
			protected.PUT("/messages/:id", controllers.UpdateMessage)
			protected.DELETE("/messages/:id", controllers.DeleteMessage)

			// Document Routes
			protected.POST("/documents", controllers.CreateDocument)
			protected.GET("/documents", controllers.GetDocuments)
			protected.GET("/documents/:id", controllers.GetDocument)
			protected.PUT("/documents/:id", controllers.UpdateDocument)
			protected.DELETE("/documents/:id", controllers.DeleteDocument)

			// Document File Routes
			protected.POST("/documents/:id/files", controllers.UploadDocumentFile)
			protected.GET("/documents/:id/files/:fileId/download", controllers.GetDocumentFileDownloadURL)
			protected.DELETE("/documents/:id/files/:fileId", controllers.DeleteDocumentFile)
		}
	}

	return r
}
