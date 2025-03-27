package tests

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"travel-agent/internal/config"
	"travel-agent/internal/server"

	"github.com/gin-gonic/gin"
)

func TestHealthHandler(t *testing.T) {
	// Test setup
	cfg := &config.Config{
		ServerPort: ":8080",
		LogLevel:   "info",
	}
	srv := server.New(cfg)

	// Create test request
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	// Execute request
	router := gin.Default()
	router.GET("/health", func(c *gin.Context) {
		srv.HealthHandler(c.Writer, c.Request)
	})
	router.ServeHTTP(w, req)

	// Check status code
	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	// Check Content-Type header
	contentType := w.Header().Get("Content-Type")
	expectedContentType := "application/json"
	if contentType != expectedContentType {
		t.Errorf("Expected Content-Type %s, got %s", expectedContentType, contentType)
	}

	// Read and parse response body
	body, err := io.ReadAll(w.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	var response map[string]string
	if err := json.Unmarshal(body, &response); err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}

	// Check response content
	expectedStatus := "OK"
	if status, exists := response["status"]; !exists || status != expectedStatus {
		t.Errorf("Expected status %q, got %q", expectedStatus, status)
	}
}
