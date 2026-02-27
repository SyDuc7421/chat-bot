package routes

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"hsduc.com/rag/controllers"
	_ "hsduc.com/rag/docs"
	"hsduc.com/rag/middleware"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

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
		}
	}

	return r
}
