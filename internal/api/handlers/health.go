package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"certificate-monkey/internal/storage"
	"certificate-monkey/internal/version"
)

// HealthHandler handles health check HTTP requests
type HealthHandler struct {
	storage *storage.DynamoDBStorage
	logger  *logrus.Logger
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(storage *storage.DynamoDBStorage, logger *logrus.Logger) *HealthHandler {
	return &HealthHandler{
		storage: storage,
		logger:  logger,
	}
}

// HealthResponse represents the basic health check response
type HealthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
	Version string `json:"version"`
}

// AWSHealthResponse represents the AWS connectivity health check response
type AWSHealthResponse struct {
	Status    string                 `json:"status"`
	Service   string                 `json:"service"`
	Version   string                 `json:"version"`
	Timestamp string                 `json:"timestamp"`
	Checks    map[string]HealthCheck `json:"checks"`
}

// HealthCheck represents individual service check result
type HealthCheck struct {
	Status     string `json:"status"`
	Message    string `json:"message,omitempty"`
	ResponseMs int64  `json:"response_ms"`
	Error      string `json:"error,omitempty"`
}

// BasicHealth returns basic health status
// @Summary Basic health check
// @Description Returns basic service health status
// @Tags Health
// @Produce json
// @Success 200 {object} HealthResponse "Service is healthy"
// @Router /health [get]
func (h *HealthHandler) BasicHealth(c *gin.Context) {
	c.JSON(http.StatusOK, HealthResponse{
		Status:  "healthy",
		Service: "certificate-monkey",
		Version: version.GetVersion(),
	})
}

// AWSHealth checks AWS services connectivity
// @Summary AWS connectivity health check
// @Description Verifies connectivity to DynamoDB and KMS services
// @Tags Health
// @Produce json
// @Success 200 {object} AWSHealthResponse "All AWS services are accessible"
// @Failure 503 {object} AWSHealthResponse "One or more AWS services are unavailable"
// @Router /health/aws [get]
func (h *HealthHandler) AWSHealth(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	checks := make(map[string]HealthCheck)
	overallHealthy := true

	// Check DynamoDB connectivity
	dynamoCheck := h.checkDynamoDB(ctx)
	checks["dynamodb"] = dynamoCheck
	if dynamoCheck.Status != "healthy" {
		overallHealthy = false
	}

	// Check KMS connectivity
	kmsCheck := h.checkKMS(ctx)
	checks["kms"] = kmsCheck
	if kmsCheck.Status != "healthy" {
		overallHealthy = false
	}

	// Determine overall status
	status := "healthy"
	httpStatus := http.StatusOK
	if !overallHealthy {
		status = "unhealthy"
		httpStatus = http.StatusServiceUnavailable
	}

	response := AWSHealthResponse{
		Status:    status,
		Service:   "certificate-monkey",
		Version:   version.GetVersion(),
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Checks:    checks,
	}

	h.logger.WithFields(logrus.Fields{
		"overall_status": status,
		"dynamodb":       dynamoCheck.Status,
		"kms":            kmsCheck.Status,
	}).Info("AWS health check completed")

	c.JSON(httpStatus, response)
}

// checkDynamoDB verifies DynamoDB table accessibility
func (h *HealthHandler) checkDynamoDB(ctx context.Context) HealthCheck {
	start := time.Now()

	err := h.storage.CheckDynamoDBHealth(ctx)
	elapsed := time.Since(start).Milliseconds()

	if err != nil {
		h.logger.WithError(err).Error("DynamoDB health check failed")
		return HealthCheck{
			Status:     "unhealthy",
			Message:    "Failed to access DynamoDB table",
			ResponseMs: elapsed,
			Error:      err.Error(),
		}
	}

	return HealthCheck{
		Status:     "healthy",
		Message:    "DynamoDB table is accessible",
		ResponseMs: elapsed,
	}
}

// checkKMS verifies KMS key accessibility
func (h *HealthHandler) checkKMS(ctx context.Context) HealthCheck {
	start := time.Now()

	err := h.storage.CheckKMSHealth(ctx)
	elapsed := time.Since(start).Milliseconds()

	if err != nil {
		h.logger.WithError(err).Error("KMS health check failed")
		return HealthCheck{
			Status:     "unhealthy",
			Message:    "Failed to access KMS key",
			ResponseMs: elapsed,
			Error:      err.Error(),
		}
	}

	return HealthCheck{
		Status:     "healthy",
		Message:    "KMS key is accessible",
		ResponseMs: elapsed,
	}
}
