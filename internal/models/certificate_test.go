package models

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test KeyType constants
func TestKeyTypeConstants(t *testing.T) {
	assert.Equal(t, KeyType("RSA2048"), KeyTypeRSA2048)
	assert.Equal(t, KeyType("RSA4096"), KeyTypeRSA4096)
	assert.Equal(t, KeyType("ECDSA-P256"), KeyTypeECDSAP256)
	assert.Equal(t, KeyType("ECDSA-P384"), KeyTypeECDSAP384)
}

// Test CertificateStatus constants
func TestCertificateStatusConstants(t *testing.T) {
	assert.Equal(t, CertificateStatus("PENDING_CSR"), StatusPendingCSR)
	assert.Equal(t, CertificateStatus("CSR_CREATED"), StatusCSRCreated)
	assert.Equal(t, CertificateStatus("CERT_UPLOADED"), StatusCertUploaded)
	assert.Equal(t, CertificateStatus("COMPLETED"), StatusCompleted)
}

// Test CertificateEntity JSON marshaling/unmarshaling
func TestCertificateEntityJSONSerialization(t *testing.T) {
	now := time.Now()
	validFrom := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	validTo := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	entity := &CertificateEntity{
		ID:                      "test-id-123",
		CommonName:              "example.com",
		SubjectAlternativeNames: []string{"www.example.com", "api.example.com"},
		Organization:            "ACME Corp",
		OrganizationalUnit:      "IT Department",
		Country:                 "US",
		State:                   "California",
		City:                    "San Francisco",
		EmailAddress:            "admin@example.com",
		KeyType:                 KeyTypeRSA2048,
		EncryptedPrivateKey:     "encrypted-private-key-data",
		CSR:                     "csr-pem-data",
		Certificate:             "certificate-pem-data",
		Status:                  StatusCertUploaded,
		Tags:                    map[string]string{"env": "prod", "team": "infra"},
		CreatedAt:               now,
		UpdatedAt:               now,
		ValidFrom:               &validFrom,
		ValidTo:                 &validTo,
		SerialNumber:            "123456789",
		Fingerprint:             "AA:BB:CC:DD:EE:FF",
	}

	// Test marshaling
	jsonData, err := json.Marshal(entity)
	require.NoError(t, err)
	assert.NotEmpty(t, jsonData)

	// Test unmarshaling
	var unmarshaled CertificateEntity
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err)

	// Verify all fields
	assert.Equal(t, entity.ID, unmarshaled.ID)
	assert.Equal(t, entity.CommonName, unmarshaled.CommonName)
	assert.Equal(t, entity.SubjectAlternativeNames, unmarshaled.SubjectAlternativeNames)
	assert.Equal(t, entity.Organization, unmarshaled.Organization)
	assert.Equal(t, entity.OrganizationalUnit, unmarshaled.OrganizationalUnit)
	assert.Equal(t, entity.Country, unmarshaled.Country)
	assert.Equal(t, entity.State, unmarshaled.State)
	assert.Equal(t, entity.City, unmarshaled.City)
	assert.Equal(t, entity.EmailAddress, unmarshaled.EmailAddress)
	assert.Equal(t, entity.KeyType, unmarshaled.KeyType)
	assert.Equal(t, entity.EncryptedPrivateKey, unmarshaled.EncryptedPrivateKey)
	assert.Equal(t, entity.CSR, unmarshaled.CSR)
	assert.Equal(t, entity.Certificate, unmarshaled.Certificate)
	assert.Equal(t, entity.Status, unmarshaled.Status)
	assert.Equal(t, entity.Tags, unmarshaled.Tags)
	assert.Equal(t, entity.SerialNumber, unmarshaled.SerialNumber)
	assert.Equal(t, entity.Fingerprint, unmarshaled.Fingerprint)

	// Time fields require special handling due to precision
	assert.WithinDuration(t, entity.CreatedAt, unmarshaled.CreatedAt, time.Second)
	assert.WithinDuration(t, entity.UpdatedAt, unmarshaled.UpdatedAt, time.Second)

	require.NotNil(t, unmarshaled.ValidFrom)
	require.NotNil(t, unmarshaled.ValidTo)
	assert.Equal(t, entity.ValidFrom.UTC(), unmarshaled.ValidFrom.UTC())
	assert.Equal(t, entity.ValidTo.UTC(), unmarshaled.ValidTo.UTC())
}

// Test CertificateEntity with minimal fields
func TestCertificateEntityMinimalFields(t *testing.T) {
	entity := &CertificateEntity{
		ID:         "minimal-id",
		CommonName: "minimal.example.com",
		KeyType:    KeyTypeECDSAP256,
		Status:     StatusCSRCreated,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	jsonData, err := json.Marshal(entity)
	require.NoError(t, err)

	var unmarshaled CertificateEntity
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, entity.ID, unmarshaled.ID)
	assert.Equal(t, entity.CommonName, unmarshaled.CommonName)
	assert.Equal(t, entity.KeyType, unmarshaled.KeyType)
	assert.Equal(t, entity.Status, unmarshaled.Status)

	// Optional fields should be empty/nil
	assert.Empty(t, unmarshaled.SubjectAlternativeNames)
	assert.Empty(t, unmarshaled.Organization)
	assert.Empty(t, unmarshaled.OrganizationalUnit)
	assert.Empty(t, unmarshaled.Country)
	assert.Empty(t, unmarshaled.State)
	assert.Empty(t, unmarshaled.City)
	assert.Empty(t, unmarshaled.EmailAddress)
	assert.Empty(t, unmarshaled.CSR)
	assert.Empty(t, unmarshaled.Certificate)
	assert.Empty(t, unmarshaled.SerialNumber)
	assert.Empty(t, unmarshaled.Fingerprint)
	assert.Nil(t, unmarshaled.ValidFrom)
	assert.Nil(t, unmarshaled.ValidTo)
}

// Test CreateKeyRequest validation and marshaling
func TestCreateKeyRequest(t *testing.T) {
	tests := []struct {
		name           string
		request        CreateKeyRequest
		expectJSONTags []string
	}{
		{
			name: "Full request with all fields",
			request: CreateKeyRequest{
				CommonName:              "test.example.com",
				SubjectAlternativeNames: []string{"www.test.example.com"},
				Organization:            "Test Corp",
				OrganizationalUnit:      "Test Dept",
				Country:                 "US",
				State:                   "California",
				City:                    "San Francisco",
				EmailAddress:            "test@example.com",
				KeyType:                 KeyTypeRSA2048,
				Tags:                    map[string]string{"test": "true"},
			},
			expectJSONTags: []string{
				"common_name", "subject_alternative_names", "organization",
				"organizational_unit", "country", "state", "city",
				"email_address", "key_type", "tags",
			},
		},
		{
			name: "Minimal request",
			request: CreateKeyRequest{
				CommonName: "minimal.example.com",
				KeyType:    KeyTypeECDSAP256,
			},
			expectJSONTags: []string{"common_name", "key_type"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData, err := json.Marshal(tt.request)
			require.NoError(t, err)

			// Check that expected JSON tags are present
			jsonStr := string(jsonData)
			for _, tag := range tt.expectJSONTags {
				assert.Contains(t, jsonStr, tag)
			}

			// Test unmarshaling
			var unmarshaled CreateKeyRequest
			err = json.Unmarshal(jsonData, &unmarshaled)
			require.NoError(t, err)

			assert.Equal(t, tt.request.CommonName, unmarshaled.CommonName)
			assert.Equal(t, tt.request.KeyType, unmarshaled.KeyType)
			assert.Equal(t, tt.request.SubjectAlternativeNames, unmarshaled.SubjectAlternativeNames)
			assert.Equal(t, tt.request.Organization, unmarshaled.Organization)
			assert.Equal(t, tt.request.OrganizationalUnit, unmarshaled.OrganizationalUnit)
			assert.Equal(t, tt.request.Country, unmarshaled.Country)
			assert.Equal(t, tt.request.State, unmarshaled.State)
			assert.Equal(t, tt.request.City, unmarshaled.City)
			assert.Equal(t, tt.request.EmailAddress, unmarshaled.EmailAddress)
			assert.Equal(t, tt.request.Tags, unmarshaled.Tags)
		})
	}
}

// Test CreateKeyResponse
func TestCreateKeyResponse(t *testing.T) {
	now := time.Now()
	response := CreateKeyResponse{
		ID:         "response-id-123",
		CommonName: "response.example.com",
		KeyType:    KeyTypeRSA4096,
		CSR:        "csr-pem-data",
		Status:     StatusCSRCreated,
		Tags:       map[string]string{"response": "test"},
		CreatedAt:  now,
	}

	jsonData, err := json.Marshal(response)
	require.NoError(t, err)

	var unmarshaled CreateKeyResponse
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, response.ID, unmarshaled.ID)
	assert.Equal(t, response.CommonName, unmarshaled.CommonName)
	assert.Equal(t, response.KeyType, unmarshaled.KeyType)
	assert.Equal(t, response.CSR, unmarshaled.CSR)
	assert.Equal(t, response.Status, unmarshaled.Status)
	assert.Equal(t, response.Tags, unmarshaled.Tags)
	assert.WithinDuration(t, response.CreatedAt, unmarshaled.CreatedAt, time.Second)
}

// Test UploadCertificateRequest
func TestUploadCertificateRequest(t *testing.T) {
	request := UploadCertificateRequest{
		Certificate: "certificate-pem-data",
	}

	jsonData, err := json.Marshal(request)
	require.NoError(t, err)
	assert.Contains(t, string(jsonData), "certificate")

	var unmarshaled UploadCertificateRequest
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, request.Certificate, unmarshaled.Certificate)
}

// Test UploadCertificateResponse
func TestUploadCertificateResponse(t *testing.T) {
	now := time.Now()
	validFrom := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	validTo := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	response := UploadCertificateResponse{
		ID:           "upload-id-123",
		Status:       StatusCertUploaded,
		ValidFrom:    &validFrom,
		ValidTo:      &validTo,
		SerialNumber: "987654321",
		Fingerprint:  "FF:EE:DD:CC:BB:AA",
		UpdatedAt:    now,
	}

	jsonData, err := json.Marshal(response)
	require.NoError(t, err)

	var unmarshaled UploadCertificateResponse
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, response.ID, unmarshaled.ID)
	assert.Equal(t, response.Status, unmarshaled.Status)
	assert.Equal(t, response.SerialNumber, unmarshaled.SerialNumber)
	assert.Equal(t, response.Fingerprint, unmarshaled.Fingerprint)
	assert.WithinDuration(t, response.UpdatedAt, unmarshaled.UpdatedAt, time.Second)

	require.NotNil(t, unmarshaled.ValidFrom)
	require.NotNil(t, unmarshaled.ValidTo)
	assert.Equal(t, response.ValidFrom.UTC(), unmarshaled.ValidFrom.UTC())
	assert.Equal(t, response.ValidTo.UTC(), unmarshaled.ValidTo.UTC())
}

// Test GeneratePFXRequest
func TestGeneratePFXRequest(t *testing.T) {
	request := GeneratePFXRequest{
		Password: "secure-password-123",
	}

	jsonData, err := json.Marshal(request)
	require.NoError(t, err)
	assert.Contains(t, string(jsonData), "password")

	var unmarshaled GeneratePFXRequest
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, request.Password, unmarshaled.Password)
}

// Test GeneratePFXResponse
func TestGeneratePFXResponse(t *testing.T) {
	response := GeneratePFXResponse{
		ID:       "pfx-id-123",
		PFXData:  "base64-encoded-pfx-data",
		Filename: "test.example.com-pfx123.pfx",
	}

	jsonData, err := json.Marshal(response)
	require.NoError(t, err)

	var unmarshaled GeneratePFXResponse
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, response.ID, unmarshaled.ID)
	assert.Equal(t, response.PFXData, unmarshaled.PFXData)
	assert.Equal(t, response.Filename, unmarshaled.Filename)
}

// Test ListKeysResponse
func TestListKeysResponse(t *testing.T) {
	entities := []CertificateEntity{
		{
			ID:         "entity-1",
			CommonName: "entity1.example.com",
			KeyType:    KeyTypeRSA2048,
			Status:     StatusCSRCreated,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		},
		{
			ID:         "entity-2",
			CommonName: "entity2.example.com",
			KeyType:    KeyTypeECDSAP256,
			Status:     StatusCertUploaded,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		},
	}

	response := ListKeysResponse{
		Keys:       entities,
		TotalCount: 2,
		Page:       1,
		PageSize:   10,
	}

	jsonData, err := json.Marshal(response)
	require.NoError(t, err)

	var unmarshaled ListKeysResponse
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, response.TotalCount, unmarshaled.TotalCount)
	assert.Equal(t, response.Page, unmarshaled.Page)
	assert.Equal(t, response.PageSize, unmarshaled.PageSize)
	assert.Len(t, unmarshaled.Keys, 2)

	for i, entity := range unmarshaled.Keys {
		assert.Equal(t, entities[i].ID, entity.ID)
		assert.Equal(t, entities[i].CommonName, entity.CommonName)
		assert.Equal(t, entities[i].KeyType, entity.KeyType)
		assert.Equal(t, entities[i].Status, entity.Status)
	}
}

// Test SearchFilters
func TestSearchFilters(t *testing.T) {
	dateFrom := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	dateTo := time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC)

	filters := SearchFilters{
		Tags:     map[string]string{"env": "prod", "team": "infra"},
		Status:   StatusCertUploaded,
		KeyType:  KeyTypeRSA2048,
		DateFrom: &dateFrom,
		DateTo:   &dateTo,
		Page:     2,
		PageSize: 25,
	}

	// SearchFilters is used with form tags, so we test the struct directly
	assert.Equal(t, map[string]string{"env": "prod", "team": "infra"}, filters.Tags)
	assert.Equal(t, StatusCertUploaded, filters.Status)
	assert.Equal(t, KeyTypeRSA2048, filters.KeyType)
	assert.Equal(t, dateFrom, *filters.DateFrom)
	assert.Equal(t, dateTo, *filters.DateTo)
	assert.Equal(t, 2, filters.Page)
	assert.Equal(t, 25, filters.PageSize)
}

// Test empty/nil handling
func TestEmptyAndNilHandling(t *testing.T) {
	// Test entity with nil time pointers
	entity := CertificateEntity{
		ID:         "nil-test",
		CommonName: "nil.example.com",
		KeyType:    KeyTypeRSA2048,
		Status:     StatusCSRCreated,
		ValidFrom:  nil,
		ValidTo:    nil,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	jsonData, err := json.Marshal(entity)
	require.NoError(t, err)

	var unmarshaled CertificateEntity
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err)

	assert.Nil(t, unmarshaled.ValidFrom)
	assert.Nil(t, unmarshaled.ValidTo)

	// Test empty arrays and maps
	assert.Empty(t, unmarshaled.SubjectAlternativeNames)
	assert.Empty(t, unmarshaled.Tags)
}

// Test JSON omitempty behavior
func TestJSONOmitEmpty(t *testing.T) {
	// Create entity with only required fields
	entity := CertificateEntity{
		ID:         "omit-test",
		CommonName: "omit.example.com",
		KeyType:    KeyTypeECDSAP256,
		Status:     StatusCSRCreated,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	jsonData, err := json.Marshal(entity)
	require.NoError(t, err)

	jsonStr := string(jsonData)

	// These fields should be omitted when empty
	assert.NotContains(t, jsonStr, "subject_alternative_names")
	assert.NotContains(t, jsonStr, "organization")
	assert.NotContains(t, jsonStr, "organizational_unit")
	assert.NotContains(t, jsonStr, "country")
	assert.NotContains(t, jsonStr, "state")
	assert.NotContains(t, jsonStr, "city")
	assert.NotContains(t, jsonStr, "email_address")
	assert.NotContains(t, jsonStr, "csr")
	assert.NotContains(t, jsonStr, "certificate")
	assert.NotContains(t, jsonStr, "tags")
	assert.NotContains(t, jsonStr, "valid_from")
	assert.NotContains(t, jsonStr, "valid_to")
	assert.NotContains(t, jsonStr, "serial_number")
	assert.NotContains(t, jsonStr, "fingerprint")

	// These fields should always be present
	assert.Contains(t, jsonStr, "id")
	assert.Contains(t, jsonStr, "common_name")
	assert.Contains(t, jsonStr, "key_type")
	assert.Contains(t, jsonStr, "status")
	assert.Contains(t, jsonStr, "created_at")
	assert.Contains(t, jsonStr, "updated_at")
}
