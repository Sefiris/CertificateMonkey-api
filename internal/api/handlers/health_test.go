package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"certificate-monkey/internal/storage"
)

func TestNewHealthHandler(t *testing.T) {
	logger := logrus.New()
	storage := &storage.DynamoDBStorage{}

	handler := NewHealthHandler(storage, logger)

	assert.NotNil(t, handler)
	assert.NotNil(t, handler.storage)
	assert.NotNil(t, handler.logger)
	assert.Equal(t, logger, handler.logger)
	assert.Equal(t, storage, handler.storage)
}

func TestBasicHealth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	storage := &storage.DynamoDBStorage{}
	logger := logrus.New()
	logger.SetOutput(nil) // Suppress log output during tests

	handler := NewHealthHandler(storage, logger)

	router := gin.New()
	router.GET("/health", handler.BasicHealth)

	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "healthy", response["status"])
	assert.Equal(t, "certificate-monkey", response["service"])
	assert.NotEmpty(t, response["version"])
}

func TestBasicHealthResponseStructure(t *testing.T) {
	gin.SetMode(gin.TestMode)

	storage := &storage.DynamoDBStorage{}
	logger := logrus.New()
	logger.SetOutput(nil)

	handler := NewHealthHandler(storage, logger)
	router := gin.New()
	router.GET("/health", handler.BasicHealth)

	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"))

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Verify all expected fields are present
	assert.Contains(t, response, "status")
	assert.Contains(t, response, "service")
	assert.Contains(t, response, "version")
}

// TestAWSHealthEndpointExists verifies the AWS health endpoint can be registered
func TestAWSHealthEndpointExists(t *testing.T) {
	gin.SetMode(gin.TestMode)

	storage := &storage.DynamoDBStorage{}
	logger := logrus.New()
	logger.SetOutput(nil)

	handler := NewHealthHandler(storage, logger)

	// Verify the AWSHealth method exists and is callable
	assert.NotNil(t, handler.AWSHealth)

	// Verify the handler has the required components
	assert.NotNil(t, handler.storage)
	assert.NotNil(t, handler.logger)
}

// Note: Full integration testing of /health/aws requires real AWS credentials and resources.
// This test verifies the handler structure is correct without calling AWS.

func TestHandlerLoggerIsUsed(t *testing.T) {
	storage := &storage.DynamoDBStorage{}
	logger := logrus.New()

	handler := NewHealthHandler(storage, logger)

	assert.Same(t, logger, handler.logger, "Handler should use the provided logger")
	assert.Same(t, storage, handler.storage, "Handler should use the provided storage")
}
