package routes

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"certificate-monkey/internal/config"
	"certificate-monkey/internal/crypto"
	"certificate-monkey/internal/storage"
)

// MockStorage and MockCrypto for testing
type MockStorage struct{}

func (m *MockStorage) CreateCertificateEntity(ctx interface{}, entity interface{}) error {
	return nil
}

func (m *MockStorage) GetCertificateEntity(ctx interface{}, id string) (interface{}, error) {
	return nil, nil
}

func (m *MockStorage) UpdateCertificateEntity(ctx interface{}, entity interface{}) error {
	return nil
}

func (m *MockStorage) ListCertificateEntities(ctx interface{}, filters interface{}) (interface{}, error) {
	return nil, nil
}

type MockCrypto struct{}

func (m *MockCrypto) GenerateKeyAndCSR(req interface{}) (string, string, error) {
	return "", "", nil
}

func (m *MockCrypto) ParseCertificate(certPEM string) (interface{}, error) {
	return nil, nil
}

func (m *MockCrypto) GenerateCertificateFingerprint(certPEM string) (string, error) {
	return "", nil
}

func (m *MockCrypto) ValidateCertificateWithCSR(certPEM, csrPEM string) error {
	return nil
}

func (m *MockCrypto) GeneratePFX(privateKeyPEM, certificatePEM, password string) ([]byte, error) {
	return nil, nil
}

func (m *MockCrypto) EncodeToBase64(data []byte) string {
	return ""
}

// Test SetupRoutes basic functionality
func TestSetupRoutes(t *testing.T) {
	// Set gin to test mode
	gin.SetMode(gin.TestMode)

	// Create test config
	cfg := &config.Config{
		Server: config.ServerConfig{
			Host: "localhost",
			Port: "8080",
		},
		Security: config.SecurityConfig{
			APIKeys: []string{"test_key"},
		},
	}

	// Create mock dependencies
	storage := &storage.DynamoDBStorage{}
	cryptoService := crypto.NewCryptoService()
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	// This should not panic
	assert.NotPanics(t, func() {
		router := SetupRoutes(cfg, storage, cryptoService, logger)
		assert.NotNil(t, router)
	})
}

// Test health endpoint
func TestHealthEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		Server: config.ServerConfig{
			Host: "localhost",
			Port: "8080",
		},
		Security: config.SecurityConfig{
			APIKeys: []string{"test_key"},
		},
	}

	storage := &storage.DynamoDBStorage{}
	cryptoService := crypto.NewCryptoService()
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	router := SetupRoutes(cfg, storage, cryptoService, logger)

	// Test health endpoint
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "healthy", response["status"])
	assert.Equal(t, "certificate-monkey", response["service"])
	assert.Equal(t, "0.1.0", response["version"])
}

// Test CORS middleware
func TestCorsMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(corsMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "test"})
	})

	tests := []struct {
		name         string
		method       string
		expectedCode int
	}{
		{
			name:         "GET request",
			method:       "GET",
			expectedCode: http.StatusOK,
		},
		{
			name:         "OPTIONS request",
			method:       "OPTIONS",
			expectedCode: http.StatusNoContent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/test", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)

			// Check CORS headers
			assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
			assert.Equal(t, "GET, POST, PUT, DELETE, OPTIONS", w.Header().Get("Access-Control-Allow-Methods"))
			assert.Equal(t, "Content-Type, Authorization, X-API-Key", w.Header().Get("Access-Control-Allow-Headers"))
			assert.Equal(t, "3600", w.Header().Get("Access-Control-Max-Age"))
		})
	}
}

// Test request ID middleware
func TestRequestIDMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(requestIDMiddleware())
	router.GET("/test", func(c *gin.Context) {
		requestID := c.GetString("request_id")
		c.JSON(http.StatusOK, gin.H{"request_id": requestID})
	})

	t.Run("generates request ID when not provided", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Check that request ID header is set
		requestID := w.Header().Get("X-Request-ID")
		assert.NotEmpty(t, requestID)
		assert.True(t, strings.HasPrefix(requestID, "req_"))

		// Check response contains request ID
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, requestID, response["request_id"])
	})

	t.Run("uses provided request ID", func(t *testing.T) {
		customRequestID := "custom-request-id-12345"

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Request-ID", customRequestID)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Check that the custom request ID is used
		assert.Equal(t, customRequestID, w.Header().Get("X-Request-ID"))

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, customRequestID, response["request_id"])
	})
}

// Test generateRequestID
func TestGenerateRequestID(t *testing.T) {
	// Generate multiple request IDs to ensure they're unique
	requestIDs := make(map[string]bool)
	for i := 0; i < 100; i++ {
		id := generateRequestID()
		assert.True(t, strings.HasPrefix(id, "req_"))
		assert.False(t, requestIDs[id], "Request ID should be unique: %s", id)
		requestIDs[id] = true

		// Check format: req_ followed by 8 hex characters
		matched, err := regexp.MatchString(`^req_[a-f0-9]{8}$`, id)
		assert.NoError(t, err)
		assert.True(t, matched, "Request ID format should be req_[8hexchars]: %s", id)
	}
}

// Test protected routes require authentication
func TestProtectedRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		Server: config.ServerConfig{
			Host: "localhost",
			Port: "8080",
		},
		Security: config.SecurityConfig{
			APIKeys: []string{"valid_key"},
		},
	}

	storage := &storage.DynamoDBStorage{}
	cryptoService := crypto.NewCryptoService()
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	router := SetupRoutes(cfg, storage, cryptoService, logger)

	protectedEndpoints := []struct {
		method string
		path   string
	}{
		{"GET", "/api/v1/keys"},
		{"POST", "/api/v1/keys"},
		{"GET", "/api/v1/keys/test-id"},
		{"GET", "/api/v1/keys/test-id/private-key"},
		{"PUT", "/api/v1/keys/test-id/certificate"},
		{"POST", "/api/v1/keys/test-id/pfx"},
	}

	for _, endpoint := range protectedEndpoints {
		t.Run(endpoint.method+"_"+endpoint.path+"_without_auth", func(t *testing.T) {
			req := httptest.NewRequest(endpoint.method, endpoint.path, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusUnauthorized, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)
			assert.Equal(t, "Unauthorized", response["error"])
		})

		t.Run(endpoint.method+"_"+endpoint.path+"_with_auth", func(t *testing.T) {
			req := httptest.NewRequest(endpoint.method, endpoint.path, nil)
			req.Header.Set("X-API-Key", "valid_key")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Should not be unauthorized (might be other errors due to missing implementation)
			assert.NotEqual(t, http.StatusUnauthorized, w.Code)
		})
	}
}

// Test NoRoute handler
func TestNoRouteHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		Server: config.ServerConfig{
			Host: "localhost",
			Port: "8080",
		},
		Security: config.SecurityConfig{
			APIKeys: []string{"test_key"},
		},
	}

	storage := &storage.DynamoDBStorage{}
	cryptoService := crypto.NewCryptoService()
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	router := SetupRoutes(cfg, storage, cryptoService, logger)

	testPaths := []string{
		"/nonexistent",
		"/api/v2/keys",
		"/api/v1/nonexistent",
		"/health/detailed",
	}

	for _, path := range testPaths {
		t.Run("NoRoute_"+path, func(t *testing.T) {
			req := httptest.NewRequest("GET", path, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusNotFound, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			assert.Equal(t, "Not Found", response["error"])
			assert.Equal(t, "The requested endpoint does not exist", response["message"])
			assert.Equal(t, path, response["path"])
		})
	}
}

// Test Gin mode setting based on configuration
func TestGinModeConfiguration(t *testing.T) {
	originalMode := gin.Mode()
	defer gin.SetMode(originalMode) // Restore original mode

	tests := []struct {
		name         string
		serverHost   string
		expectedMode string
	}{
		{
			name:         "Production mode for 0.0.0.0",
			serverHost:   "0.0.0.0",
			expectedMode: gin.ReleaseMode,
		},
		{
			name:         "Debug mode for localhost",
			serverHost:   "localhost",
			expectedMode: gin.DebugMode,
		},
		{
			name:         "Debug mode for 127.0.0.1",
			serverHost:   "127.0.0.1",
			expectedMode: gin.DebugMode,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				Server: config.ServerConfig{
					Host: tt.serverHost,
					Port: "8080",
				},
				Security: config.SecurityConfig{
					APIKeys: []string{"test_key"},
				},
			}

			storage := &storage.DynamoDBStorage{}
			cryptoService := crypto.NewCryptoService()
			logger := logrus.New()
			logger.SetLevel(logrus.ErrorLevel)

			SetupRoutes(cfg, storage, cryptoService, logger)
			assert.Equal(t, tt.expectedMode, gin.Mode())
		})
	}
}

// Test route grouping
func TestRouteGrouping(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		Server: config.ServerConfig{
			Host: "localhost",
			Port: "8080",
		},
		Security: config.SecurityConfig{
			APIKeys: []string{"valid_key"},
		},
	}

	storage := &storage.DynamoDBStorage{}
	cryptoService := crypto.NewCryptoService()
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	router := SetupRoutes(cfg, storage, cryptoService, logger)

	// Test that all expected routes are properly grouped under /api/v1/keys
	keyRoutes := []struct {
		method string
		path   string
	}{
		{"POST", "/api/v1/keys"},
		{"GET", "/api/v1/keys"},
		{"GET", "/api/v1/keys/test-id"},
		{"GET", "/api/v1/keys/test-id/private-key"},
		{"PUT", "/api/v1/keys/test-id/certificate"},
		{"POST", "/api/v1/keys/test-id/pfx"},
	}

	for _, route := range keyRoutes {
		t.Run("Route_exists_"+route.method+"_"+route.path, func(t *testing.T) {
			req := httptest.NewRequest(route.method, route.path, nil)
			req.Header.Set("X-API-Key", "valid_key")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Should not return 404 (not found), indicating the route exists
			assert.NotEqual(t, http.StatusNotFound, w.Code)
		})
	}
}

// Benchmark route setup
func BenchmarkSetupRoutes(b *testing.B) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Host: "localhost",
			Port: "8080",
		},
		Security: config.SecurityConfig{
			APIKeys: []string{"benchmark_key"},
		},
	}

	storage := &storage.DynamoDBStorage{}
	cryptoService := crypto.NewCryptoService()
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		router := SetupRoutes(cfg, storage, cryptoService, logger)
		_ = router // Avoid unused variable
	}
}

// Benchmark request processing
func BenchmarkHealthEndpoint(b *testing.B) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		Server: config.ServerConfig{
			Host: "localhost",
			Port: "8080",
		},
		Security: config.SecurityConfig{
			APIKeys: []string{"benchmark_key"},
		},
	}

	storage := &storage.DynamoDBStorage{}
	cryptoService := crypto.NewCryptoService()
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	router := SetupRoutes(cfg, storage, cryptoService, logger)

	req := httptest.NewRequest("GET", "/health", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			b.Fatalf("Expected 200, got %d", w.Code)
		}
	}
}
