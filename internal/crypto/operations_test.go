package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"software.sslmate.com/src/go-pkcs12"

	"certificate-monkey/internal/models"
)

// CryptoTestSuite groups related crypto tests
type CryptoTestSuite struct {
	suite.Suite
	cryptoService *CryptoService
}

func (suite *CryptoTestSuite) SetupTest() {
	suite.cryptoService = NewCryptoService()
}

func TestCryptoTestSuite(t *testing.T) {
	suite.Run(t, new(CryptoTestSuite))
}

// Test NewCryptoService
func (suite *CryptoTestSuite) TestNewCryptoService() {
	cs := NewCryptoService()
	assert.NotNil(suite.T(), cs)
}

// Test GenerateKeyAndCSR with various key types and field combinations
func (suite *CryptoTestSuite) TestGenerateKeyAndCSR() {
	tests := []struct {
		name        string
		request     models.CreateKeyRequest
		expectError bool
		errorMsg    string
	}{
		{
			name: "RSA2048 with all fields",
			request: models.CreateKeyRequest{
				CommonName:              "example.com",
				SubjectAlternativeNames: []string{"www.example.com", "api.example.com", "192.168.1.1"},
				Organization:            "ACME Corp",
				OrganizationalUnit:      "IT Department",
				Country:                 "US",
				State:                   "California",
				City:                    "San Francisco",
				EmailAddress:            "admin@example.com",
				KeyType:                 models.KeyTypeRSA2048,
				Tags:                    map[string]string{"env": "test"},
			},
			expectError: false,
		},
		{
			name: "RSA4096 minimal fields",
			request: models.CreateKeyRequest{
				CommonName: "minimal.example.com",
				KeyType:    models.KeyTypeRSA4096,
			},
			expectError: false,
		},
		{
			name: "ECDSA-P256 with organization",
			request: models.CreateKeyRequest{
				CommonName:   "ecdsa.example.com",
				Organization: "ECDSA Corp",
				Country:      "CA",
				KeyType:      models.KeyTypeECDSAP256,
			},
			expectError: false,
		},
		{
			name: "ECDSA-P384 with email",
			request: models.CreateKeyRequest{
				CommonName:   "John Doe",
				EmailAddress: "john.doe@example.com",
				KeyType:      models.KeyTypeECDSAP384,
			},
			expectError: false,
		},
		{
			name: "Invalid key type",
			request: models.CreateKeyRequest{
				CommonName: "invalid.example.com",
				KeyType:    "INVALID",
			},
			expectError: true,
			errorMsg:    "unsupported key type",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			privateKeyPEM, csrPEM, err := suite.cryptoService.GenerateKeyAndCSR(tt.request)

			if tt.expectError {
				assert.Error(suite.T(), err)
				assert.Contains(suite.T(), err.Error(), tt.errorMsg)
				assert.Empty(suite.T(), privateKeyPEM)
				assert.Empty(suite.T(), csrPEM)
				return
			}

			// Should not error
			require.NoError(suite.T(), err)
			assert.NotEmpty(suite.T(), privateKeyPEM)
			assert.NotEmpty(suite.T(), csrPEM)

			// Validate private key PEM format
			privateKeyBlock, _ := pem.Decode([]byte(privateKeyPEM))
			require.NotNil(suite.T(), privateKeyBlock)
			assert.Contains(suite.T(), []string{"RSA PRIVATE KEY", "EC PRIVATE KEY"}, privateKeyBlock.Type)

			// Validate CSR PEM format
			csrBlock, _ := pem.Decode([]byte(csrPEM))
			require.NotNil(suite.T(), csrBlock)
			assert.Equal(suite.T(), "CERTIFICATE REQUEST", csrBlock.Type)

			// Parse and validate CSR contents
			csr, err := x509.ParseCertificateRequest(csrBlock.Bytes)
			require.NoError(suite.T(), err)

			// Verify subject fields
			assert.Equal(suite.T(), tt.request.CommonName, csr.Subject.CommonName)

			if tt.request.Organization != "" {
				assert.Contains(suite.T(), csr.Subject.Organization, tt.request.Organization)
			}
			if tt.request.OrganizationalUnit != "" {
				assert.Contains(suite.T(), csr.Subject.OrganizationalUnit, tt.request.OrganizationalUnit)
			}
			if tt.request.Country != "" {
				assert.Contains(suite.T(), csr.Subject.Country, tt.request.Country)
			}
			if tt.request.State != "" {
				assert.Contains(suite.T(), csr.Subject.Province, tt.request.State)
			}
			if tt.request.City != "" {
				assert.Contains(suite.T(), csr.Subject.Locality, tt.request.City)
			}
			if tt.request.EmailAddress != "" {
				assert.Contains(suite.T(), csr.EmailAddresses, tt.request.EmailAddress)
			}

			// Verify SAN fields
			for _, san := range tt.request.SubjectAlternativeNames {
				if strings.Contains(san, ".") && !strings.Contains(san, ":") && !strings.Contains(san, "/") {
					// Check if it looks like an IP address (simple heuristic)
					if strings.Count(san, ".") == 3 {
						// Likely an IP address - check IPAddresses
						found := false
						for _, ip := range csr.IPAddresses {
							if ip.String() == san {
								found = true
								break
							}
						}
						assert.True(suite.T(), found, "IP SAN %s not found in CSR", san)
					} else {
						// Likely a domain name - check DNSNames
						found := false
						for _, dns := range csr.DNSNames {
							if dns == san {
								found = true
								break
							}
						}
						assert.True(suite.T(), found, "Domain SAN %s not found in CSR", san)
					}
				}
			}

			// Verify key type by parsing the private key
			switch tt.request.KeyType {
			case models.KeyTypeRSA2048, models.KeyTypeRSA4096:
				assert.Equal(suite.T(), "RSA PRIVATE KEY", privateKeyBlock.Type)
				rsaKey, err := x509.ParsePKCS1PrivateKey(privateKeyBlock.Bytes)
				require.NoError(suite.T(), err)

				expectedBits := 2048
				if tt.request.KeyType == models.KeyTypeRSA4096 {
					expectedBits = 4096
				}
				assert.Equal(suite.T(), expectedBits, rsaKey.N.BitLen())

			case models.KeyTypeECDSAP256, models.KeyTypeECDSAP384:
				assert.Equal(suite.T(), "EC PRIVATE KEY", privateKeyBlock.Type)
				ecKey, err := x509.ParseECPrivateKey(privateKeyBlock.Bytes)
				require.NoError(suite.T(), err)

				expectedCurve := elliptic.P256()
				if tt.request.KeyType == models.KeyTypeECDSAP384 {
					expectedCurve = elliptic.P384()
				}
				assert.Equal(suite.T(), expectedCurve, ecKey.Curve)
			}
		})
	}
}

// Test ParseCertificate
func (suite *CryptoTestSuite) TestParseCertificate() {
	// Create a test certificate
	testCert := suite.createTestCertificate()

	tests := []struct {
		name        string
		certPEM     string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Valid certificate",
			certPEM:     testCert,
			expectError: false,
		},
		{
			name:        "Invalid PEM",
			certPEM:     "invalid pem data",
			expectError: true,
			errorMsg:    "failed to decode PEM block",
		},
		{
			name: "Wrong PEM type",
			certPEM: `-----BEGIN PRIVATE KEY-----
MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQC7
-----END PRIVATE KEY-----`,
			expectError: true,
			errorMsg:    "invalid certificate PEM block type",
		},
		{
			name: "Invalid certificate data",
			certPEM: `-----BEGIN CERTIFICATE-----
invaliddata
-----END CERTIFICATE-----`,
			expectError: true,
			errorMsg:    "failed to decode PEM block",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			cert, err := suite.cryptoService.ParseCertificate(tt.certPEM)

			if tt.expectError {
				assert.Error(suite.T(), err)
				assert.Contains(suite.T(), err.Error(), tt.errorMsg)
				assert.Nil(suite.T(), cert)
			} else {
				assert.NoError(suite.T(), err)
				assert.NotNil(suite.T(), cert)
				assert.Equal(suite.T(), "test.example.com", cert.Subject.CommonName)
			}
		})
	}
}

// Test GenerateCertificateFingerprint
func (suite *CryptoTestSuite) TestGenerateCertificateFingerprint() {
	testCert := suite.createTestCertificate()

	tests := []struct {
		name        string
		certPEM     string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Valid certificate",
			certPEM:     testCert,
			expectError: false,
		},
		{
			name:        "Invalid certificate",
			certPEM:     "invalid",
			expectError: true,
			errorMsg:    "failed to decode PEM block",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			fingerprint, err := suite.cryptoService.GenerateCertificateFingerprint(tt.certPEM)

			if tt.expectError {
				assert.Error(suite.T(), err)
				assert.Contains(suite.T(), err.Error(), tt.errorMsg)
				assert.Empty(suite.T(), fingerprint)
			} else {
				assert.NoError(suite.T(), err)
				assert.NotEmpty(suite.T(), fingerprint)

				// Fingerprint should be uppercase hex with colons
				assert.Regexp(suite.T(), `^[A-F0-9:]+$`, fingerprint)
				assert.Contains(suite.T(), fingerprint, ":")

				// Should be consistent
				fingerprint2, err := suite.cryptoService.GenerateCertificateFingerprint(tt.certPEM)
				assert.NoError(suite.T(), err)
				assert.Equal(suite.T(), fingerprint, fingerprint2)
			}
		})
	}
}

// Test ValidateCertificateWithCSR
func (suite *CryptoTestSuite) TestValidateCertificateWithCSR() {
	// Generate a key and CSR
	req := models.CreateKeyRequest{
		CommonName: "validate.example.com",
		KeyType:    models.KeyTypeRSA2048,
	}
	privateKeyPEM, csrPEM, err := suite.cryptoService.GenerateKeyAndCSR(req)
	require.NoError(suite.T(), err)

	// Create a matching certificate
	matchingCert := suite.createMatchingCertificate(privateKeyPEM, csrPEM)

	// Create a non-matching certificate
	nonMatchingCert := suite.createTestCertificate()

	tests := []struct {
		name        string
		certPEM     string
		csrPEM      string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Matching certificate and CSR",
			certPEM:     matchingCert,
			csrPEM:      csrPEM,
			expectError: false,
		},
		{
			name:        "Non-matching certificate",
			certPEM:     nonMatchingCert,
			csrPEM:      csrPEM,
			expectError: true,
			errorMsg:    "certificate public key does not match CSR public key",
		},
		{
			name:        "Invalid certificate",
			certPEM:     "invalid",
			csrPEM:      csrPEM,
			expectError: true,
			errorMsg:    "failed to parse certificate",
		},
		{
			name:        "Invalid CSR",
			certPEM:     matchingCert,
			csrPEM:      "invalid",
			expectError: true,
			errorMsg:    "failed to decode CSR PEM block",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			err := suite.cryptoService.ValidateCertificateWithCSR(tt.certPEM, tt.csrPEM)

			if tt.expectError {
				assert.Error(suite.T(), err)
				assert.Contains(suite.T(), err.Error(), tt.errorMsg)
			} else {
				assert.NoError(suite.T(), err)
			}
		})
	}
}

// Test Base64 encoding/decoding
func (suite *CryptoTestSuite) TestBase64Operations() {
	testData := []byte("Hello, Certificate Monkey!")

	// Test encoding
	encoded := suite.cryptoService.EncodeToBase64(testData)
	assert.NotEmpty(suite.T(), encoded)
	assert.NotEqual(suite.T(), string(testData), encoded)

	// Test decoding
	decoded, err := suite.cryptoService.DecodeFromBase64(encoded)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), testData, decoded)

	// Test invalid base64
	_, err = suite.cryptoService.DecodeFromBase64("invalid base64!")
	assert.Error(suite.T(), err)
}

// Test GeneratePFX with both RSA and ECDSA keys
func (suite *CryptoTestSuite) TestGeneratePFX() {
	tests := []struct {
		name        string
		keyType     models.KeyType
		password    string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "RSA2048 PFX generation",
			keyType:     models.KeyTypeRSA2048,
			password:    "test-password-123",
			expectError: false,
		},
		{
			name:        "RSA4096 PFX generation",
			keyType:     models.KeyTypeRSA4096,
			password:    "secure-pfx-password",
			expectError: false,
		},
		{
			name:        "ECDSA P256 PFX generation",
			keyType:     models.KeyTypeECDSAP256,
			password:    "ecdsa-test-password",
			expectError: false,
		},
		{
			name:        "ECDSA P384 PFX generation",
			keyType:     models.KeyTypeECDSAP384,
			password:    "another-secure-password",
			expectError: false,
		},
		{
			name:        "Empty password",
			keyType:     models.KeyTypeRSA2048,
			password:    "",
			expectError: false, // Empty password should work for PKCS#12
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			// Generate a key and CSR
			req := models.CreateKeyRequest{
				CommonName: "pfx-test.example.com",
				KeyType:    tt.keyType,
			}
			privateKeyPEM, csrPEM, err := suite.cryptoService.GenerateKeyAndCSR(req)
			require.NoError(suite.T(), err)

			// Create a matching certificate
			certificatePEM := suite.createMatchingCertificate(privateKeyPEM, csrPEM)

			// Generate PFX
			pfxData, err := suite.cryptoService.GeneratePFX(privateKeyPEM, certificatePEM, tt.password)

			if tt.expectError {
				assert.Error(suite.T(), err)
				if tt.errorMsg != "" {
					assert.Contains(suite.T(), err.Error(), tt.errorMsg)
				}
				assert.Nil(suite.T(), pfxData)
			} else {
				assert.NoError(suite.T(), err)
				assert.NotNil(suite.T(), pfxData)
				assert.Greater(suite.T(), len(pfxData), 0, "PFX data should not be empty")

				// Verify we can decode the PFX back (basic validation)
				decodedKey, decodedCert, err := pkcs12.Decode(pfxData, tt.password)
				assert.NoError(suite.T(), err, "Should be able to decode generated PFX")
				assert.NotNil(suite.T(), decodedKey, "Decoded private key should not be nil")
				assert.NotNil(suite.T(), decodedCert, "Decoded certificate should not be nil")

				// Verify the decoded certificate matches the original
				originalCert, err := suite.cryptoService.ParseCertificate(certificatePEM)
				require.NoError(suite.T(), err)
				assert.Equal(suite.T(), originalCert.Subject, decodedCert.Subject)
				assert.Equal(suite.T(), originalCert.SerialNumber, decodedCert.SerialNumber)
			}
		})
	}

	// Test error cases
	suite.Run("Invalid private key", func() {
		certificatePEM := suite.createTestCertificate()
		_, err := suite.cryptoService.GeneratePFX("invalid-private-key", certificatePEM, "password")
		assert.Error(suite.T(), err)
		assert.Contains(suite.T(), err.Error(), "failed to parse private key")
	})

	suite.Run("Invalid certificate", func() {
		req := models.CreateKeyRequest{
			CommonName: "test.example.com",
			KeyType:    models.KeyTypeRSA2048,
		}
		privateKeyPEM, _, err := suite.cryptoService.GenerateKeyAndCSR(req)
		require.NoError(suite.T(), err)

		_, err = suite.cryptoService.GeneratePFX(privateKeyPEM, "invalid-certificate", "password")
		assert.Error(suite.T(), err)
		assert.Contains(suite.T(), err.Error(), "failed to parse certificate")
	})
}

// Test private key parsing with different formats
func (suite *CryptoTestSuite) TestParsePrivateKeyFromPEM() {
	// Generate test keys for each supported type
	rsaKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	ecKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

	tests := []struct {
		name        string
		privateKey  interface{}
		expectError bool
	}{
		{
			name:        "RSA private key",
			privateKey:  rsaKey,
			expectError: false,
		},
		{
			name:        "ECDSA private key",
			privateKey:  ecKey,
			expectError: false,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			// Encode the key
			pemData, err := suite.cryptoService.encodePrivateKeyToPEM(tt.privateKey)
			require.NoError(suite.T(), err)

			// Parse it back
			parsedKey, err := suite.cryptoService.parsePrivateKeyFromPEM(pemData)

			if tt.expectError {
				assert.Error(suite.T(), err)
				assert.Nil(suite.T(), parsedKey)
			} else {
				assert.NoError(suite.T(), err)
				assert.NotNil(suite.T(), parsedKey)

				// Type should match
				switch tt.privateKey.(type) {
				case *rsa.PrivateKey:
					_, ok := parsedKey.(*rsa.PrivateKey)
					assert.True(suite.T(), ok)
				case *ecdsa.PrivateKey:
					_, ok := parsedKey.(*ecdsa.PrivateKey)
					assert.True(suite.T(), ok)
				}
			}
		})
	}

	// Test invalid PEM
	_, err := suite.cryptoService.parsePrivateKeyFromPEM("invalid")
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "failed to decode PEM block")

	// Test unsupported key type
	invalidPEM := `-----BEGIN UNKNOWN KEY-----
MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQC7
-----END UNKNOWN KEY-----`
	_, err = suite.cryptoService.parsePrivateKeyFromPEM(invalidPEM)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "unsupported private key type")
}

// Helper function to create a test certificate
func (suite *CryptoTestSuite) createTestCertificate() string {
	// Generate a private key
	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)

	// Create certificate template
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: "test.example.com",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	// Create the certificate
	certDER, _ := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)

	// Encode to PEM
	return string(pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDER,
	}))
}

// Helper function to create a certificate that matches a given CSR
func (suite *CryptoTestSuite) createMatchingCertificate(privateKeyPEM, csrPEM string) string {
	// Parse the private key
	privateKey, err := suite.cryptoService.parsePrivateKeyFromPEM(privateKeyPEM)
	require.NoError(suite.T(), err)

	// Parse the CSR
	csrBlock, _ := pem.Decode([]byte(csrPEM))
	require.NotNil(suite.T(), csrBlock)
	csr, err := x509.ParseCertificateRequest(csrBlock.Bytes)
	require.NoError(suite.T(), err)

	// Create certificate template matching the CSR
	template := x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               csr.Subject,
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              csr.DNSNames,
		IPAddresses:           csr.IPAddresses,
		EmailAddresses:        csr.EmailAddresses,
	}

	// Create the certificate
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, csr.PublicKey, privateKey)
	require.NoError(suite.T(), err)

	// Encode to PEM
	return string(pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDER,
	}))
}
