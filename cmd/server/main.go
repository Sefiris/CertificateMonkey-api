package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"certificate-monkey/docs"
	_ "certificate-monkey/docs" // Import generated docs
	"certificate-monkey/internal/api/routes"
	appConfig "certificate-monkey/internal/config"
	"certificate-monkey/internal/crypto"
	"certificate-monkey/internal/storage"
	"certificate-monkey/internal/version"
)

// @title Certificate Monkey API
// @description Secure certificate management API for private keys, CSRs, and certificates
// @description
// @description Certificate Monkey provides a complete solution for managing the certificate lifecycle:
// @description - Generate private keys (RSA 2048/4096, ECDSA P-256/P-384)
// @description - Create certificate signing requests (CSRs)
// @description - Upload and validate certificates
// @description - Generate PFX/PKCS#12 files for legacy applications
// @description - Export private keys (with comprehensive audit logging)
// @description
// @description All private keys are encrypted with AWS KMS and stored in DynamoDB.
// @description The API provides comprehensive search and filtering capabilities.
// @version 0.1.0
// @contact.name Certificate Monkey Support
// @contact.url https://github.com/your-username/certificate-monkey
// @contact.email support@certificatemonkey.dev
// @license.name MIT
// @license.url https://github.com/your-username/certificate-monkey/blob/main/LICENSE
// @host localhost:8080
// @BasePath /api/v1
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-API-Key
// @description API key for authentication. Use 'demo-api-key-12345' for testing.
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Bearer token for authentication. Format: 'Bearer <your-api-key>'

func main() {
	// Set up logger
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(logrus.InfoLevel)

	// Load configuration
	cfg, err := appConfig.Load()
	if err != nil {
		logger.WithError(err).Fatal("Failed to load configuration")
	}

	// Update Swagger info with current version
	docs.SwaggerInfo.Version = version.GetVersion()

	logger.WithFields(logrus.Fields{
		"version":    version.GetVersion(),
		"build_info": version.Get(),
	}).Info("Starting Certificate Monkey API")

	// Initialize AWS configuration
	awsCfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(cfg.AWS.Region),
	)
	if err != nil {
		logger.WithError(err).Fatal("Failed to load AWS configuration")
	}

	// Initialize AWS clients
	dynamoClient := dynamodb.NewFromConfig(awsCfg)
	kmsClient := kms.NewFromConfig(awsCfg)

	// Initialize storage layer
	dbStorage := storage.NewDynamoDBStorage(dynamoClient, kmsClient, cfg, logger)

	// Initialize crypto service
	cryptoService := crypto.NewCryptoService()

	// Set up routes
	router := routes.SetupRoutes(cfg, dbStorage, cryptoService, logger)

	// Add build info endpoint
	router.GET("/build-info", func(c *gin.Context) {
		c.JSON(http.StatusOK, version.GetBuildInfo())
	})

	// Create HTTP server
	server := &http.Server{
		Addr:              fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port),
		Handler:           router,
		ReadTimeout:       15 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		logger.WithFields(logrus.Fields{
			"host":    cfg.Server.Host,
			"port":    cfg.Server.Port,
			"version": version.GetVersion(),
		}).Info("Server starting")

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.WithError(err).Fatal("Server failed to start")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Server shutting down...")

	// Give outstanding requests 5 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.WithError(err).Fatal("Server forced to shutdown")
	}

	logger.Info("Server exited")
}
