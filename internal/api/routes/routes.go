package routes

import (
	"crypto/rand"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"certificate-monkey/internal/api/handlers"
	"certificate-monkey/internal/api/middleware"
	"certificate-monkey/internal/config"
	"certificate-monkey/internal/crypto"
	"certificate-monkey/internal/storage"
	"certificate-monkey/internal/version"
)

// SetupRoutes configures all API routes
func SetupRoutes(
	cfg *config.Config,
	storage *storage.DynamoDBStorage,
	cryptoService *crypto.CryptoService,
	logger *logrus.Logger,
) *gin.Engine {
	// Set Gin mode
	if strings.Contains(cfg.Server.Host, "0.0.0.0") {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	// Create Gin router
	router := gin.New()

	// Add middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(corsMiddleware())
	router.Use(requestIDMiddleware())

	// Health check endpoint (no auth required)
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "certificate-monkey",
			"version": version.GetVersion(),
		})
	})

	// Swagger documentation endpoint (no authentication required)
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API version 1 routes
	v1 := router.Group("/api/v1")

	// Apply authentication middleware to all v1 routes
	v1.Use(middleware.AuthMiddleware(cfg, logger))

	// Create handlers
	certHandler := handlers.NewCertificateHandler(storage, cryptoService, logger)

	// Certificate management endpoints
	keys := v1.Group("/keys")
	{
		keys.POST("", certHandler.CreateKey)                        // POST /api/v1/keys
		keys.GET("", certHandler.ListCertificates)                  // GET /api/v1/keys
		keys.GET("/:id", certHandler.GetCertificate)                // GET /api/v1/keys/{id}
		keys.GET("/:id/private-key", certHandler.ExportPrivateKey)  // GET /api/v1/keys/{id}/private-key
		keys.PUT("/:id/certificate", certHandler.UploadCertificate) // PUT /api/v1/keys/{id}/certificate
		keys.POST("/:id/pfx", certHandler.GeneratePFX)              // POST /api/v1/keys/{id}/pfx
	}

	// Add a catch-all route for undefined endpoints
	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Not Found",
			"message": "The requested endpoint does not exist",
			"path":    c.Request.URL.Path,
		})
	})

	return router
}

// corsMiddleware adds CORS headers
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-API-Key")
		c.Header("Access-Control-Max-Age", "3600")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// requestIDMiddleware adds a unique request ID to each request
func requestIDMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}
		c.Header("X-Request-ID", requestID)
		c.Set("request_id", requestID)
		c.Next()
	})
}

// generateRequestID generates a simple request ID
func generateRequestID() string {
	// Simple implementation - in production you might want to use UUID
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		// Fallback to timestamp-based ID if crypto/rand fails
		return fmt.Sprintf("req_%d", 12345678)
	}
	return fmt.Sprintf("req_%x", b)
}
