package storage

import (
	"context"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"certificate-monkey/internal/config"
	"certificate-monkey/internal/models"
)

// TestNewDynamoDBStorage tests the constructor
func TestNewDynamoDBStorage(t *testing.T) {
	logger := logrus.New()
	cfg := &config.Config{
		AWS: config.AWSConfig{
			DynamoDBTable: "test-table",
			KMSKeyID:      "test-key",
		},
	}

	// We can't easily create real AWS clients for testing without AWS setup
	// But we can test that the constructor doesn't panic
	storage := NewDynamoDBStorage(nil, nil, cfg, logger)

	assert.NotNil(t, storage)
	assert.Equal(t, cfg.AWS.DynamoDBTable, storage.tableName)
	assert.Equal(t, cfg.AWS.KMSKeyID, storage.kmsKeyID)
	assert.Equal(t, logger, storage.logger)
}

// TestSortEntitiesSliceEdgeCases tests edge cases in sorting
func TestSortEntitiesSliceEdgeCases(t *testing.T) {
	storage := &DynamoDBStorage{}

	// Test empty slice
	var emptySlice []models.CertificateEntity
	storage.sortEntities(emptySlice, "created_at", "desc")
	assert.Empty(t, emptySlice)

	// Test single item slice
	singleSlice := []models.CertificateEntity{
		{ID: "test-1", CommonName: "example.com"},
	}
	storage.sortEntities(singleSlice, "created_at", "desc")
	assert.Len(t, singleSlice, 1)
	assert.Equal(t, "test-1", singleSlice[0].ID)
}

// TestCompareEntitiesEdgeCases tests edge cases in entity comparison
func TestCompareEntitiesEdgeCases(t *testing.T) {
	storage := &DynamoDBStorage{}

	entity1 := models.CertificateEntity{
		ID:         "test-1",
		CommonName: "a.example.com",
		Status:     models.StatusCSRCreated,
		KeyType:    models.KeyTypeRSA2048,
	}

	entity2 := models.CertificateEntity{
		ID:         "test-2",
		CommonName: "b.example.com",
		Status:     models.StatusCertUploaded,
		KeyType:    models.KeyTypeRSA4096,
	}

	// Test common_name comparison (ascending)
	result := storage.compareEntities(entity1, entity2, "common_name", "asc")
	assert.False(t, result, "a.example.com should come before b.example.com in ascending order")

	// Test common_name comparison (descending)
	result = storage.compareEntities(entity1, entity2, "common_name", "desc")
	assert.True(t, result, "a.example.com should come after b.example.com in descending order")

	// Test status comparison
	result = storage.compareEntities(entity1, entity2, "status", "asc")
	// CSR_CREATED vs CERT_UPLOADED - CERT_UPLOADED should come first lexicographically
	assert.True(t, result, "CSR_CREATED should come after CERT_UPLOADED in ascending order")

	// Test key_type comparison
	result = storage.compareEntities(entity1, entity2, "key_type", "asc")
	// RSA2048 vs RSA4096 - RSA2048 should come first lexicographically
	assert.False(t, result, "RSA2048 should come before RSA4096")

	// Test default sorting (created_at) with identical entities
	result = storage.compareEntities(entity1, entity1, "created_at", "asc")
	assert.False(t, result, "Identical entities should not swap")

	// Test unknown sort field (should default to created_at)
	result = storage.compareEntities(entity1, entity1, "unknown_field", "asc")
	assert.False(t, result, "Unknown field should default to created_at comparison")
}

// TestCompareEntitiesTimeFields tests time-based comparisons with nil values
func TestCompareEntitiesTimeFields(t *testing.T) {
	storage := &DynamoDBStorage{}

	entity1 := models.CertificateEntity{
		ID:        "test-1",
		ValidTo:   nil,
		ValidFrom: nil,
	}

	entity2 := models.CertificateEntity{
		ID:        "test-2",
		ValidTo:   nil,
		ValidFrom: nil,
	}

	// Test valid_to comparison with both nil
	result := storage.compareEntities(entity1, entity2, "valid_to", "asc")
	assert.False(t, result, "Both nil ValidTo should be equal")

	// Test valid_from comparison with both nil
	result = storage.compareEntities(entity1, entity2, "valid_from", "asc")
	assert.False(t, result, "Both nil ValidFrom should be equal")
}

// TestCompareEntitiesDescendingOrder tests descending order logic
func TestCompareEntitiesDescendingOrder(t *testing.T) {
	storage := &DynamoDBStorage{}

	entity1 := models.CertificateEntity{
		ID:         "test-1",
		CommonName: "a.example.com",
	}

	entity2 := models.CertificateEntity{
		ID:         "test-2",
		CommonName: "b.example.com",
	}

	// Test descending order flips the comparison
	result := storage.compareEntities(entity1, entity2, "common_name", "desc")
	assert.True(t, result, "Descending order should flip comparison result")
}

// TestHealthCheckMethodSignatures verifies the health check methods have correct signatures
func TestHealthCheckMethodSignatures(t *testing.T) {
	logger := logrus.New()
	cfg := &config.Config{
		AWS: config.AWSConfig{
			DynamoDBTable: "test-table",
			KMSKeyID:      "test-key",
		},
	}

	storage := NewDynamoDBStorage(nil, nil, cfg, logger)

	// Verify storage was created
	assert.NotNil(t, storage)
	assert.Equal(t, "test-table", storage.tableName)
	assert.Equal(t, "test-key", storage.kmsKeyID)

	// Verify health check methods exist by checking they can be referenced
	// We don't call them because they require real AWS clients
	var dynamoHealthCheck func(context.Context) error = storage.CheckDynamoDBHealth
	var kmsHealthCheck func(context.Context) error = storage.CheckKMSHealth

	assert.NotNil(t, dynamoHealthCheck)
	assert.NotNil(t, kmsHealthCheck)
}
