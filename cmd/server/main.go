package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/sirupsen/logrus"

	_ "certificate-monkey/docs" // Import generated docs
	"certificate-monkey/internal/api/routes"
	appConfig "certificate-monkey/internal/config"
	"certificate-monkey/internal/crypto"
	"certificate-monkey/internal/storage"
)

// @title Certificate Monkey API
// @version 1.0
// @description A secure certificate management API for private keys, CSRs, and certificates
// @termsOfService https://github.com/your-org/certificate-monkey/blob/main/TERMS.md

// @contact.name Certificate Monkey Support
// @contact.url https://github.com/your-org/certificate-monkey
// @contact.email support@certificate-monkey.com

// @license.name MIT
// @license.url https://github.com/your-org/certificate-monkey/blob/main/LICENSE

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-API-Key
// @description API key for authentication. Can also be provided as Bearer token in Authorization header.

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Bearer token authentication. Format: "Bearer <api-key>"

func main() {
	// Load configuration
	cfg, err := appConfig.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Setup logger
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(logrus.InfoLevel)

	// Log startup information
	logger.WithFields(logrus.Fields{
		"service": "certificate-monkey",
		"version": "1.0.0",
		"host":    cfg.Server.Host,
		"port":    cfg.Server.Port,
	}).Info("Starting Certificate Monkey API server")

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

	// Setup routes
	router := routes.SetupRoutes(cfg, dbStorage, cryptoService, logger)

	// Create HTTP server
	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port),
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		logger.WithFields(logrus.Fields{
			"address": server.Addr,
		}).Info("HTTP server starting")

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.WithError(err).Fatal("Failed to start HTTP server")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Create a context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		logger.WithError(err).Error("Server forced to shutdown")
	} else {
		logger.Info("Server shutdown completed gracefully")
	}
}
