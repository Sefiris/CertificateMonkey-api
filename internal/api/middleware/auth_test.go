package middleware

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"certificate-monkey/internal/config"
)

func TestAuthMiddleware(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create test config
	cfg := &config.Config{
		Security: config.SecurityConfig{
			APIKeys: []string{"valid_key_1", "valid_key_2", "test_key_123"},
		},
	}

	// Create logger for testing
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel) // Reduce noise in tests

	// Create test router with auth middleware
	router := gin.New()
	router.Use(AuthMiddleware(cfg, logger))

	// Add a test endpoint
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	tests := []struct {
		name           string
		headers        map[string]string
		expectedStatus int
		expectedError  string
		expectedMsg    string
	}{
		{
			name: "Valid API key in X-API-Key header",
			headers: map[string]string{
				"X-API-Key": "valid_key_1",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Valid API key in Authorization Bearer header",
			headers: map[string]string{
				"Authorization": "Bearer valid_key_2",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Valid API key with different valid key",
			headers: map[string]string{
				"X-API-Key": "test_key_123",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Missing API key",
			headers:        map[string]string{},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Unauthorized",
			expectedMsg:    "API key is required",
		},
		{
			name: "Invalid API key in X-API-Key header",
			headers: map[string]string{
				"X-API-Key": "invalid_key",
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Unauthorized",
			expectedMsg:    "Invalid API key",
		},
		{
			name: "Invalid API key in Authorization Bearer header",
			headers: map[string]string{
				"Authorization": "Bearer invalid_key",
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Unauthorized",
			expectedMsg:    "Invalid API key",
		},
		{
			name: "Empty X-API-Key header",
			headers: map[string]string{
				"X-API-Key": "",
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Unauthorized",
			expectedMsg:    "API key is required",
		},
		{
			name: "Empty Authorization Bearer",
			headers: map[string]string{
				"Authorization": "Bearer ",
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Unauthorized",
			expectedMsg:    "API key is required",
		},
		{
			name: "Authorization without Bearer prefix",
			headers: map[string]string{
				"Authorization": "valid_key_1",
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Unauthorized",
			expectedMsg:    "API key is required",
		},
		{
			name: "X-API-Key takes precedence over Authorization",
			headers: map[string]string{
				"X-API-Key":     "valid_key_1",
				"Authorization": "Bearer invalid_key",
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)

			// Set headers
			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				// Check successful response
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, "success", response["message"])
			} else {
				// Check error response
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, tt.expectedError, response["error"])
				assert.Equal(t, tt.expectedMsg, response["message"])
			}
		})
	}
}

// Test AuthMiddleware with empty API keys configuration
func TestAuthMiddlewareEmptyConfig(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Config with no valid API keys
	cfg := &config.Config{
		Security: config.SecurityConfig{
			APIKeys: []string{},
		},
	}

	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	router := gin.New()
	router.Use(AuthMiddleware(cfg, logger))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-API-Key", "any_key")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "Unauthorized", response["error"])
	assert.Equal(t, "Invalid API key", response["message"])
}

// Test AuthMiddleware with nil configuration
func TestAuthMiddlewareNilConfig(t *testing.T) {
	gin.SetMode(gin.TestMode)

	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	// This should panic with nil config, so we test that it panics
	assert.Panics(t, func() {
		router := gin.New()
		router.Use(AuthMiddleware(nil, logger))
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-API-Key", "any_key")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	})
}

// Test maskAPIKey function
func TestMaskAPIKey(t *testing.T) {
	tests := []struct {
		name     string
		apiKey   string
		expected string
	}{
		{
			name:     "Normal API key",
			apiKey:   "cm_dev_12345678",
			expected: "cm_d...5678",
		},
		{
			name:     "Long API key",
			apiKey:   "very_long_api_key_with_many_characters_12345",
			expected: "very...2345",
		},
		{
			name:     "Short API key",
			apiKey:   "short",
			expected: "***",
		},
		{
			name:     "Very short API key",
			apiKey:   "ab",
			expected: "***",
		},
		{
			name:     "Empty API key",
			apiKey:   "",
			expected: "***",
		},
		{
			name:     "Exactly 8 characters",
			apiKey:   "12345678",
			expected: "1234...5678",
		},
		{
			name:     "Exactly 7 characters",
			apiKey:   "1234567",
			expected: "***",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := maskAPIKey(tt.apiKey)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test AuthMiddleware with different HTTP methods
func TestAuthMiddlewareHTTPMethods(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		Security: config.SecurityConfig{
			APIKeys: []string{"valid_key"},
		},
	}

	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	router := gin.New()
	router.Use(AuthMiddleware(cfg, logger))

	// Add endpoints for different HTTP methods
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"method": "GET"})
	})
	router.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"method": "POST"})
	})
	router.PUT("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"method": "PUT"})
	})
	router.DELETE("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"method": "DELETE"})
	})

	methods := []string{"GET", "POST", "PUT", "DELETE"}

	for _, method := range methods {
		t.Run("Method_"+method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/test", nil)
			req.Header.Set("X-API-Key", "valid_key")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)
			assert.Equal(t, method, response["method"])
		})
	}
}

// Test AuthMiddleware with request body
func TestAuthMiddlewareWithRequestBody(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		Security: config.SecurityConfig{
			APIKeys: []string{"valid_key"},
		},
	}

	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	router := gin.New()
	router.Use(AuthMiddleware(cfg, logger))

	router.POST("/test", func(c *gin.Context) {
		var body map[string]interface{}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"received": body})
	})

	testData := map[string]interface{}{
		"test":  "data",
		"count": 42,
	}

	jsonData, _ := json.Marshal(testData)

	req := httptest.NewRequest("POST", "/test", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", "valid_key")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	received := response["received"].(map[string]interface{})
	assert.Equal(t, "data", received["test"])
	assert.Equal(t, float64(42), received["count"]) // JSON unmarshals numbers as float64
}

// Test AuthMiddleware logging behavior
func TestAuthMiddlewareLogging(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		Security: config.SecurityConfig{
			APIKeys: []string{"valid_key"},
		},
	}

	// Create logger that captures output
	logger := logrus.New()
	var logBuffer bytes.Buffer
	logger.SetOutput(&logBuffer)
	logger.SetLevel(logrus.DebugLevel) // Enable debug logs

	router := gin.New()
	router.Use(AuthMiddleware(cfg, logger))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	t.Run("Successful authentication logs debug message", func(t *testing.T) {
		logBuffer.Reset()

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-API-Key", "valid_key")
		req.Header.Set("User-Agent", "Test-Agent/1.0")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		logOutput := logBuffer.String()
		assert.Contains(t, logOutput, "Request authenticated successfully")
		assert.Contains(t, logOutput, "vali..._key") // Adjusted to match actual format
	})

	t.Run("Missing API key logs warning", func(t *testing.T) {
		logBuffer.Reset()

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("User-Agent", "Test-Agent/1.0")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		logOutput := logBuffer.String()
		assert.Contains(t, logOutput, "Missing API key in request")
	})

	t.Run("Invalid API key logs warning", func(t *testing.T) {
		logBuffer.Reset()

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-API-Key", "invalid_key")
		req.Header.Set("User-Agent", "Test-Agent/1.0")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		logOutput := logBuffer.String()
		assert.Contains(t, logOutput, "Invalid API key used")
		assert.Contains(t, logOutput, "inva..._key") // Adjusted to match actual format
	})
}

// Benchmark the auth middleware
func BenchmarkAuthMiddleware(b *testing.B) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		Security: config.SecurityConfig{
			APIKeys: []string{"benchmark_key_1", "benchmark_key_2", "benchmark_key_3"},
		},
	}

	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel) // Disable logging for benchmark

	router := gin.New()
	router.Use(AuthMiddleware(cfg, logger))
	router.GET("/benchmark", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/benchmark", nil)
	req.Header.Set("X-API-Key", "benchmark_key_2")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			b.Fatalf("Expected 200, got %d", w.Code)
		}
	}
}
