package routes

import (
	"github.com/gin-gonic/gin"
	"hsduc.com/rag/controllers"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	// Health Check Route
	r.GET("/health", controllers.HealthCheck)

	api := r.Group("/api/v1")
	{
		// Conversation Routes
		api.POST("/conversations", controllers.CreateConversation)
		api.GET("/conversations", controllers.GetConversations)
		api.GET("/conversations/:id", controllers.GetConversation)
		api.PUT("/conversations/:id", controllers.UpdateConversation)
		api.DELETE("/conversations/:id", controllers.DeleteConversation)

		// Message Routes
		api.POST("/messages", controllers.CreateMessage)
		api.GET("/messages", controllers.GetMessages) // Use query ?conversation_id=X
		api.GET("/messages/:id", controllers.GetMessage)
		api.PUT("/messages/:id", controllers.UpdateMessage)
		api.DELETE("/messages/:id", controllers.DeleteMessage)
	}

	return r
}
