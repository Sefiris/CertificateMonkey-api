package models

import (
	"time"
)

// KeyType represents the supported cryptographic key types
type KeyType string

const (
	KeyTypeRSA2048   KeyType = "RSA2048"
	KeyTypeRSA4096   KeyType = "RSA4096"
	KeyTypeECDSAP256 KeyType = "ECDSA-P256"
	KeyTypeECDSAP384 KeyType = "ECDSA-P384"
)

// CertificateStatus represents the current status of a certificate
type CertificateStatus string

const (
	StatusPendingCSR   CertificateStatus = "PENDING_CSR"
	StatusCSRCreated   CertificateStatus = "CSR_CREATED"
	StatusCertUploaded CertificateStatus = "CERT_UPLOADED"
	StatusCompleted    CertificateStatus = "COMPLETED"
)

// CertificateEntity represents the main entity stored in DynamoDB
type CertificateEntity struct {
	// DynamoDB Primary Key
	ID string `json:"id" dynamodbav:"id"`

	// Certificate Information
	CommonName              string   `json:"common_name" dynamodbav:"common_name"`
	SubjectAlternativeNames []string `json:"subject_alternative_names,omitempty" dynamodbav:"subject_alternative_names,omitempty"`
	Organization            string   `json:"organization,omitempty" dynamodbav:"organization,omitempty"`
	OrganizationalUnit      string   `json:"organizational_unit,omitempty" dynamodbav:"organizational_unit,omitempty"`
	Country                 string   `json:"country,omitempty" dynamodbav:"country,omitempty"`
	State                   string   `json:"state,omitempty" dynamodbav:"state,omitempty"`
	City                    string   `json:"city,omitempty" dynamodbav:"city,omitempty"`
	EmailAddress            string   `json:"email_address,omitempty" dynamodbav:"email_address,omitempty"`

	// Cryptographic Details
	KeyType             KeyType `json:"key_type" dynamodbav:"key_type"`
	EncryptedPrivateKey string  `json:"encrypted_private_key" dynamodbav:"encrypted_private_key"`
	CSR                 string  `json:"csr,omitempty" dynamodbav:"csr,omitempty"`
	Certificate         string  `json:"certificate,omitempty" dynamodbav:"certificate,omitempty"`

	// Metadata
	Status    CertificateStatus `json:"status" dynamodbav:"status"`
	Tags      map[string]string `json:"tags,omitempty" dynamodbav:"tags,omitempty"`
	CreatedAt time.Time         `json:"created_at" dynamodbav:"created_at"`
	UpdatedAt time.Time         `json:"updated_at" dynamodbav:"updated_at"`

	// Certificate Details (populated when certificate is uploaded)
	ValidFrom    *time.Time `json:"valid_from,omitempty" dynamodbav:"valid_from,omitempty"`
	ValidTo      *time.Time `json:"valid_to,omitempty" dynamodbav:"valid_to,omitempty"`
	SerialNumber string     `json:"serial_number,omitempty" dynamodbav:"serial_number,omitempty"`
	Fingerprint  string     `json:"fingerprint,omitempty" dynamodbav:"fingerprint,omitempty"`
}

// CreateKeyRequest represents the request to create a new private key and CSR
type CreateKeyRequest struct {
	CommonName              string            `json:"common_name" binding:"required"`
	SubjectAlternativeNames []string          `json:"subject_alternative_names,omitempty"`
	Organization            string            `json:"organization,omitempty"`
	OrganizationalUnit      string            `json:"organizational_unit,omitempty"`
	Country                 string            `json:"country,omitempty"`
	State                   string            `json:"state,omitempty"`
	City                    string            `json:"city,omitempty"`
	EmailAddress            string            `json:"email_address,omitempty"`
	KeyType                 KeyType           `json:"key_type" binding:"required"`
	Tags                    map[string]string `json:"tags,omitempty"`
}

// CreateKeyResponse represents the response after creating a key and CSR
type CreateKeyResponse struct {
	ID         string            `json:"id"`
	CommonName string            `json:"common_name"`
	KeyType    KeyType           `json:"key_type"`
	CSR        string            `json:"csr"`
	Status     CertificateStatus `json:"status"`
	Tags       map[string]string `json:"tags,omitempty"`
	CreatedAt  time.Time         `json:"created_at"`
}

// UploadCertificateRequest represents the request to upload a certificate
type UploadCertificateRequest struct {
	Certificate string `json:"certificate" binding:"required"`
}

// UploadCertificateResponse represents the response after uploading a certificate
type UploadCertificateResponse struct {
	ID           string            `json:"id"`
	Status       CertificateStatus `json:"status"`
	ValidFrom    *time.Time        `json:"valid_from,omitempty"`
	ValidTo      *time.Time        `json:"valid_to,omitempty"`
	SerialNumber string            `json:"serial_number,omitempty"`
	Fingerprint  string            `json:"fingerprint,omitempty"`
	UpdatedAt    time.Time         `json:"updated_at"`
}

// GeneratePFXRequest represents the request to generate a PFX file
type GeneratePFXRequest struct {
	Password string `json:"password" binding:"required"`
}

// GeneratePFXResponse represents the response for PFX generation
type GeneratePFXResponse struct {
	ID       string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	PFXData  string `json:"pfx_data" example:"base64_encoded_pfx_data"`
	Filename string `json:"filename" example:"example.com-550e8400.pfx"`
}

// ExportPrivateKeyResponse represents the response for private key export
type ExportPrivateKeyResponse struct {
	ID         string  `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	PrivateKey string  `json:"private_key" example:"-----BEGIN PRIVATE KEY-----\nMIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQC7VJTUt9Us8cKB...\n-----END PRIVATE KEY-----"`
	KeyType    KeyType `json:"key_type" example:"RSA2048"`
	CommonName string  `json:"common_name" example:"example.com"`
	ExportedAt string  `json:"exported_at" example:"2024-01-15T10:30:00Z"`
}

// ListKeysResponse represents the response for listing keys
type ListKeysResponse struct {
	Keys       []CertificateEntity `json:"keys"`
	TotalCount int                 `json:"total_count"`
	Page       int                 `json:"page"`
	PageSize   int                 `json:"page_size"`
}

// SearchFilters represents filters for searching certificates
type SearchFilters struct {
	Tags     map[string]string `form:"tags"`
	Status   CertificateStatus `form:"status"`
	KeyType  KeyType           `form:"key_type"`
	DateFrom *time.Time        `form:"date_from"`
	DateTo   *time.Time        `form:"date_to"`
	Page     int               `form:"page"`
	PageSize int               `form:"page_size"`
}
