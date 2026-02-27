package controllers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestHealthCheck(t *testing.T) {
	SetupTestDB()
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.GET("/health", HealthCheck)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected to get status %d but got %d", http.StatusOK, w.Code)
	}

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse response body: %v", err)
	}

	if response["status"] != "ok" {
		t.Errorf("Expected status to be ok, got %v", response["status"])
	}

	if response["database"] != "ok" {
		t.Errorf("Expected database to be ok, got %v", response["database"])
	}
}
