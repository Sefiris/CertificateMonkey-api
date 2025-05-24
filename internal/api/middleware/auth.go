package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"certificate-monkey/internal/config"
)

// AuthMiddleware creates authentication middleware for API key validation
func AuthMiddleware(cfg *config.Config, logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get API key from header
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			// Also check Authorization header with Bearer prefix
			authHeader := c.GetHeader("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				apiKey = strings.TrimPrefix(authHeader, "Bearer ")
			}
		}

		if apiKey == "" {
			logger.WithFields(logrus.Fields{
				"remote_addr": c.ClientIP(),
				"user_agent":  c.GetHeader("User-Agent"),
				"path":        c.Request.URL.Path,
			}).Warn("Missing API key in request")

			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "API key is required",
			})
			c.Abort()
			return
		}

		// Validate API key
		isValid := false
		for _, validKey := range cfg.Security.APIKeys {
			if apiKey == validKey {
				isValid = true
				break
			}
		}

		if !isValid {
			logger.WithFields(logrus.Fields{
				"remote_addr": c.ClientIP(),
				"user_agent":  c.GetHeader("User-Agent"),
				"path":        c.Request.URL.Path,
				"api_key":     maskAPIKey(apiKey),
			}).Warn("Invalid API key used")

			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "Invalid API key",
			})
			c.Abort()
			return
		}

		// Log successful authentication
		logger.WithFields(logrus.Fields{
			"remote_addr": c.ClientIP(),
			"path":        c.Request.URL.Path,
			"api_key":     maskAPIKey(apiKey),
		}).Debug("Request authenticated successfully")

		// Continue to the next handler
		c.Next()
	}
}

// maskAPIKey masks an API key for logging purposes
func maskAPIKey(apiKey string) string {
	if len(apiKey) < 8 {
		return "***"
	}
	return apiKey[:4] + "..." + apiKey[len(apiKey)-4:]
}
