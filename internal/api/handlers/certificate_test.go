package handlers

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"certificate-monkey/internal/crypto"
)

// TestNewCertificateHandler tests the constructor
func TestNewCertificateHandler(t *testing.T) {
	logger := logrus.New()
	cryptoService := crypto.NewCryptoService()

	// We can't easily create a real DynamoDB storage for testing without AWS setup
	// But we can test that the constructor doesn't panic
	handler := NewCertificateHandler(nil, cryptoService, logger)

	assert.NotNil(t, handler)
	assert.Equal(t, cryptoService, handler.cryptoService)
	assert.Equal(t, logger, handler.logger)
}

// TestCertificateHandlerFields tests that all fields are properly set
func TestCertificateHandlerFields(t *testing.T) {
	logger := logrus.New()
	cryptoService := crypto.NewCryptoService()

	handler := &CertificateHandler{
		// Note: storage is intentionally not set here as it would be nil anyway
		cryptoService: cryptoService,
		logger:        logger,
	}

	assert.NotNil(t, handler.cryptoService)
	assert.NotNil(t, handler.logger)
	assert.Nil(t, handler.storage) // Explicitly test that storage is nil when not set
}

// TestCertificateHandlerType tests the struct type
func TestCertificateHandlerType(t *testing.T) {
	handler := &CertificateHandler{}

	// Test that it's the right type
	assert.IsType(t, &CertificateHandler{}, handler)

	// Test that we can set fields
	logger := logrus.New()
	cryptoService := crypto.NewCryptoService()

	handler.logger = logger
	handler.cryptoService = cryptoService

	assert.Equal(t, logger, handler.logger)
	assert.Equal(t, cryptoService, handler.cryptoService)
}
