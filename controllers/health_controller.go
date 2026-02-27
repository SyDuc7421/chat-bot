package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"hsduc.com/rag/database"
)

// @Summary      Health Check
// @Description  Check the server health
// @Tags         Health
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Router       /health [get]
func HealthCheck(c *gin.Context) {
	// Optional: Check database connectivity
	var dbStatus = "ok"
	if database.DB != nil {
		sqlDB, err := database.DB.DB()
		if err != nil || sqlDB.Ping() != nil {
			dbStatus = "error"
		}
	} else {
		dbStatus = "error"
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   "ok",
		"database": dbStatus,
	})
}
