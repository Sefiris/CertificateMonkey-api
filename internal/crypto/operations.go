package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"net"
	"net/url"
	"strings"

	"software.sslmate.com/src/go-pkcs12"

	"certificate-monkey/internal/models"
)

// CryptoService handles all cryptographic operations
type CryptoService struct{}

// NewCryptoService creates a new instance of CryptoService
func NewCryptoService() *CryptoService {
	return &CryptoService{}
}

// GenerateKeyAndCSR generates a private key and certificate signing request
func (cs *CryptoService) GenerateKeyAndCSR(req models.CreateKeyRequest) (privateKeyPEM, csrPEM string, err error) {
	// Generate the private key based on the key type
	var privateKey interface{}
	switch req.KeyType {
	case models.KeyTypeRSA2048:
		privateKey, err = rsa.GenerateKey(rand.Reader, 2048)
	case models.KeyTypeRSA4096:
		privateKey, err = rsa.GenerateKey(rand.Reader, 4096)
	case models.KeyTypeECDSAP256:
		privateKey, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	case models.KeyTypeECDSAP384:
		privateKey, err = ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	default:
		return "", "", fmt.Errorf("unsupported key type: %s", req.KeyType)
	}

	if err != nil {
		return "", "", fmt.Errorf("failed to generate private key: %w", err)
	}

	// Encode private key to PEM format
	privateKeyPEM, err = cs.encodePrivateKeyToPEM(privateKey)
	if err != nil {
		return "", "", fmt.Errorf("failed to encode private key: %w", err)
	}

	// Create certificate signing request template
	template := x509.CertificateRequest{
		Subject: pkix.Name{
			CommonName: req.CommonName,
		},
		EmailAddresses: []string{},
	}

	// Add Subject fields only if they are not empty
	if req.Organization != "" {
		template.Subject.Organization = []string{req.Organization}
	}
	if req.OrganizationalUnit != "" {
		template.Subject.OrganizationalUnit = []string{req.OrganizationalUnit}
	}
	if req.Country != "" {
		template.Subject.Country = []string{req.Country}
	}
	if req.State != "" {
		template.Subject.Province = []string{req.State}
	}
	if req.City != "" {
		template.Subject.Locality = []string{req.City}
	}
	if req.EmailAddress != "" {
		template.EmailAddresses = []string{req.EmailAddress}
	}

	// Add Subject Alternative Names
	for _, san := range req.SubjectAlternativeNames {
		if ip := net.ParseIP(san); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else if u, err := url.Parse("https://" + san); err == nil && u.Host == san {
			template.DNSNames = append(template.DNSNames, san)
		} else {
			template.DNSNames = append(template.DNSNames, san)
		}
	}

	// Create CSR
	csrDER, err := x509.CreateCertificateRequest(rand.Reader, &template, privateKey)
	if err != nil {
		return "", "", fmt.Errorf("failed to create certificate request: %w", err)
	}

	// Encode CSR to PEM format
	csrPEM = string(pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE REQUEST",
		Bytes: csrDER,
	}))

	return privateKeyPEM, csrPEM, nil
}

// encodePrivateKeyToPEM encodes a private key to PEM format
func (cs *CryptoService) encodePrivateKeyToPEM(privateKey interface{}) (string, error) {
	var privateKeyBytes []byte
	var blockType string
	var err error

	switch key := privateKey.(type) {
	case *rsa.PrivateKey:
		privateKeyBytes = x509.MarshalPKCS1PrivateKey(key)
		blockType = "RSA PRIVATE KEY"
	case *ecdsa.PrivateKey:
		privateKeyBytes, err = x509.MarshalECPrivateKey(key)
		if err != nil {
			return "", err
		}
		blockType = "EC PRIVATE KEY"
	default:
		return "", fmt.Errorf("unsupported private key type")
	}

	return string(pem.EncodeToMemory(&pem.Block{
		Type:  blockType,
		Bytes: privateKeyBytes,
	})), nil
}

// ParseCertificate parses a PEM-encoded certificate and returns certificate details
func (cs *CryptoService) ParseCertificate(certPEM string) (*x509.Certificate, error) {
	block, _ := pem.Decode([]byte(certPEM))
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	if block.Type != "CERTIFICATE" {
		return nil, fmt.Errorf("invalid certificate PEM block type: %s", block.Type)
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	return cert, nil
}

// GenerateCertificateFingerprint generates SHA256 fingerprint of a certificate
func (cs *CryptoService) GenerateCertificateFingerprint(certPEM string) (string, error) {
	cert, err := cs.ParseCertificate(certPEM)
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256(cert.Raw)
	fingerprint := fmt.Sprintf("%x", hash)

	// Format as XX:XX:XX... for readability
	var formatted strings.Builder
	for i, b := range fingerprint {
		if i > 0 && i%2 == 0 {
			formatted.WriteString(":")
		}
		formatted.WriteString(string(b))
	}

	return strings.ToUpper(formatted.String()), nil
}

// ValidateCertificateWithCSR validates that a certificate matches the CSR
func (cs *CryptoService) ValidateCertificateWithCSR(certPEM, csrPEM string) error {
	// Parse certificate
	cert, err := cs.ParseCertificate(certPEM)
	if err != nil {
		return fmt.Errorf("failed to parse certificate: %w", err)
	}

	// Parse CSR
	csrBlock, _ := pem.Decode([]byte(csrPEM))
	if csrBlock == nil {
		return fmt.Errorf("failed to decode CSR PEM block")
	}

	csr, err := x509.ParseCertificateRequest(csrBlock.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse CSR: %w", err)
	}

	// Verify that the certificate's public key matches the CSR's public key
	certPubKey, err := x509.MarshalPKIXPublicKey(cert.PublicKey)
	if err != nil {
		return fmt.Errorf("failed to marshal certificate public key: %w", err)
	}

	csrPubKey, err := x509.MarshalPKIXPublicKey(csr.PublicKey)
	if err != nil {
		return fmt.Errorf("failed to marshal CSR public key: %w", err)
	}

	if string(certPubKey) != string(csrPubKey) {
		return fmt.Errorf("certificate public key does not match CSR public key")
	}

	// Verify that the subject matches
	if cert.Subject.CommonName != csr.Subject.CommonName {
		return fmt.Errorf("certificate CommonName does not match CSR CommonName")
	}

	return nil
}

// GeneratePFX creates a PFX (PKCS#12) file from private key and certificate
func (cs *CryptoService) GeneratePFX(privateKeyPEM, certificatePEM, password string) ([]byte, error) {
	// Parse the private key
	privateKey, err := cs.parsePrivateKeyFromPEM(privateKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	// Parse the certificate
	cert, err := cs.ParseCertificate(certificatePEM)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	// Create PKCS#12 bundle
	// Using Modern.Encode for better security instead of the deprecated Encode method
	pfxData, err := pkcs12.Modern.Encode(privateKey, cert, nil, password)
	if err != nil {
		return nil, fmt.Errorf("failed to encode PKCS#12: %w", err)
	}

	return pfxData, nil
}

// parsePrivateKeyFromPEM parses a PEM-encoded private key
func (cs *CryptoService) parsePrivateKeyFromPEM(privateKeyPEM string) (interface{}, error) {
	block, _ := pem.Decode([]byte(privateKeyPEM))
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	switch block.Type {
	case "RSA PRIVATE KEY":
		return x509.ParsePKCS1PrivateKey(block.Bytes)
	case "EC PRIVATE KEY":
		return x509.ParseECPrivateKey(block.Bytes)
	case "PRIVATE KEY":
		return x509.ParsePKCS8PrivateKey(block.Bytes)
	default:
		return nil, fmt.Errorf("unsupported private key type: %s", block.Type)
	}
}

// EncodeToBase64 encodes bytes to base64 string
func (cs *CryptoService) EncodeToBase64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

// DecodeFromBase64 decodes base64 string to bytes
func (cs *CryptoService) DecodeFromBase64(encoded string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(encoded)
}
