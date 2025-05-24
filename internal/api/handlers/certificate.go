package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"certificate-monkey/internal/crypto"
	"certificate-monkey/internal/models"
	"certificate-monkey/internal/storage"
)

// CertificateHandler handles certificate-related HTTP requests
type CertificateHandler struct {
	storage       *storage.DynamoDBStorage
	cryptoService *crypto.CryptoService
	logger        *logrus.Logger
}

// NewCertificateHandler creates a new certificate handler
func NewCertificateHandler(storage *storage.DynamoDBStorage, cryptoService *crypto.CryptoService, logger *logrus.Logger) *CertificateHandler {
	return &CertificateHandler{
		storage:       storage,
		cryptoService: cryptoService,
		logger:        logger,
	}
}

// CreateKey creates a new private key and CSR
// @Summary Create a new private key and certificate signing request
// @Description Generates a new private key pair and creates a certificate signing request (CSR) with the provided details
// @Tags Certificate Management
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Security BearerAuth
// @Param request body models.CreateKeyRequest true "Certificate creation request"
// @Success 201 {object} models.CreateKeyResponse "Successfully created private key and CSR"
// @Failure 400 {object} map[string]interface{} "Bad request - invalid input parameters"
// @Failure 401 {object} map[string]interface{} "Unauthorized - invalid or missing API key"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /keys [post]
func (h *CertificateHandler) CreateKey(c *gin.Context) {
	var req models.CreateKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Failed to bind JSON request")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad Request",
			"message": "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Validate key type
	validKeyTypes := []models.KeyType{
		models.KeyTypeRSA2048,
		models.KeyTypeRSA4096,
		models.KeyTypeECDSAP256,
		models.KeyTypeECDSAP384,
	}
	isValidKeyType := false
	for _, validType := range validKeyTypes {
		if req.KeyType == validType {
			isValidKeyType = true
			break
		}
	}
	if !isValidKeyType {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad Request",
			"message": "Invalid key type",
			"valid_types": []string{
				string(models.KeyTypeRSA2048),
				string(models.KeyTypeRSA4096),
				string(models.KeyTypeECDSAP256),
				string(models.KeyTypeECDSAP384),
			},
		})
		return
	}

	// Generate UUID for the certificate entity
	entityID := uuid.New().String()

	// Generate private key and CSR
	privateKeyPEM, csrPEM, err := h.cryptoService.GenerateKeyAndCSR(req)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"entity_id":   entityID,
			"common_name": req.CommonName,
			"key_type":    req.KeyType,
		}).Error("Failed to generate private key and CSR")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal Server Error",
			"message": "Failed to generate cryptographic material",
		})
		return
	}

	// Create certificate entity
	now := time.Now()
	entity := &models.CertificateEntity{
		ID:                      entityID,
		CommonName:              req.CommonName,
		SubjectAlternativeNames: req.SubjectAlternativeNames,
		Organization:            req.Organization,
		OrganizationalUnit:      req.OrganizationalUnit,
		Country:                 req.Country,
		State:                   req.State,
		City:                    req.City,
		EmailAddress:            req.EmailAddress,
		KeyType:                 req.KeyType,
		EncryptedPrivateKey:     privateKeyPEM,
		CSR:                     csrPEM,
		Status:                  models.StatusCSRCreated,
		Tags:                    req.Tags,
		CreatedAt:               now,
		UpdatedAt:               now,
	}

	// Store in DynamoDB
	err = h.storage.CreateCertificateEntity(c.Request.Context(), entity)
	if err != nil {
		h.logger.WithError(err).WithField("entity_id", entityID).Error("Failed to store certificate entity")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal Server Error",
			"message": "Failed to store certificate data",
		})
		return
	}

	// Prepare response
	response := models.CreateKeyResponse{
		ID:         entityID,
		CommonName: req.CommonName,
		KeyType:    req.KeyType,
		CSR:        csrPEM,
		Status:     models.StatusCSRCreated,
		Tags:       req.Tags,
		CreatedAt:  now,
	}

	h.logger.WithFields(logrus.Fields{
		"entity_id":   entityID,
		"common_name": req.CommonName,
		"key_type":    req.KeyType,
	}).Info("Private key and CSR created successfully")

	c.JSON(http.StatusCreated, response)
}

// UploadCertificate uploads a certificate for an existing CSR
// @Summary Upload certificate for existing CSR
// @Description Uploads and validates a certificate against an existing certificate signing request
// @Tags Certificate Management
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Security BearerAuth
// @Param id path string true "Certificate entity ID (UUID format)"
// @Param request body models.UploadCertificateRequest true "Certificate upload request containing PEM-encoded certificate"
// @Success 200 {object} models.UploadCertificateResponse "Certificate uploaded successfully"
// @Failure 400 {object} map[string]interface{} "Bad request - invalid certificate or ID format"
// @Failure 401 {object} map[string]interface{} "Unauthorized - invalid or missing API key"
// @Failure 404 {object} map[string]interface{} "Certificate entity not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /keys/{id}/certificate [put]
func (h *CertificateHandler) UploadCertificate(c *gin.Context) {
	entityID := c.Param("id")
	if entityID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad Request",
			"message": "Entity ID is required",
		})
		return
	}

	var req models.UploadCertificateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Failed to bind JSON request")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad Request",
			"message": "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Retrieve existing entity
	entity, err := h.storage.GetCertificateEntity(c.Request.Context(), entityID)
	if err != nil {
		h.logger.WithError(err).WithField("entity_id", entityID).Error("Failed to retrieve certificate entity")
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Not Found",
			"message": "Certificate entity not found",
		})
		return
	}

	// Validate that certificate matches the CSR
	err = h.cryptoService.ValidateCertificateWithCSR(req.Certificate, entity.CSR)
	if err != nil {
		h.logger.WithError(err).WithField("entity_id", entityID).Error("Certificate validation failed")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad Request",
			"message": "Certificate does not match the CSR",
			"details": err.Error(),
		})
		return
	}

	// Parse certificate to extract details
	cert, err := h.cryptoService.ParseCertificate(req.Certificate)
	if err != nil {
		h.logger.WithError(err).WithField("entity_id", entityID).Error("Failed to parse certificate")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad Request",
			"message": "Invalid certificate format",
			"details": err.Error(),
		})
		return
	}

	// Generate certificate fingerprint
	fingerprint, err := h.cryptoService.GenerateCertificateFingerprint(req.Certificate)
	if err != nil {
		h.logger.WithError(err).WithField("entity_id", entityID).Error("Failed to generate certificate fingerprint")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal Server Error",
			"message": "Failed to process certificate",
		})
		return
	}

	// Update entity with certificate information
	entity.Certificate = req.Certificate
	entity.Status = models.StatusCertUploaded
	entity.ValidFrom = &cert.NotBefore
	entity.ValidTo = &cert.NotAfter
	entity.SerialNumber = cert.SerialNumber.String()
	entity.Fingerprint = fingerprint

	// Update in DynamoDB
	err = h.storage.UpdateCertificateEntity(c.Request.Context(), entity)
	if err != nil {
		h.logger.WithError(err).WithField("entity_id", entityID).Error("Failed to update certificate entity")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal Server Error",
			"message": "Failed to update certificate data",
		})
		return
	}

	// Prepare response
	response := models.UploadCertificateResponse{
		ID:           entityID,
		Status:       entity.Status,
		ValidFrom:    entity.ValidFrom,
		ValidTo:      entity.ValidTo,
		SerialNumber: entity.SerialNumber,
		Fingerprint:  entity.Fingerprint,
		UpdatedAt:    entity.UpdatedAt,
	}

	h.logger.WithFields(logrus.Fields{
		"entity_id":     entityID,
		"serial_number": entity.SerialNumber,
		"fingerprint":   entity.Fingerprint,
	}).Info("Certificate uploaded successfully")

	c.JSON(http.StatusOK, response)
}

// GeneratePFX generates a PKCS#12 file for a completed certificate
// @Summary Generate PFX/P12 file
// @Description Creates a password-protected PKCS#12 file containing the private key and certificate
// @Tags Certificate Management
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Security BearerAuth
// @Param id path string true "Certificate entity ID (UUID format)"
// @Param request body models.GeneratePFXRequest true "PFX generation request with password"
// @Success 200 {object} models.GeneratePFXResponse "PFX file generated successfully (base64 encoded)"
// @Failure 400 {object} map[string]interface{} "Bad request - certificate not ready or invalid password"
// @Failure 401 {object} map[string]interface{} "Unauthorized - invalid or missing API key"
// @Failure 404 {object} map[string]interface{} "Certificate entity not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /keys/{id}/pfx [post]
func (h *CertificateHandler) GeneratePFX(c *gin.Context) {
	entityID := c.Param("id")
	if entityID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad Request",
			"message": "Entity ID is required",
		})
		return
	}

	var req models.GeneratePFXRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Failed to bind JSON request")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad Request",
			"message": "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	if req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad Request",
			"message": "Password is required for PFX generation",
		})
		return
	}

	// Retrieve entity
	entity, err := h.storage.GetCertificateEntity(c.Request.Context(), entityID)
	if err != nil {
		h.logger.WithError(err).WithField("entity_id", entityID).Error("Failed to retrieve certificate entity")
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Not Found",
			"message": "Certificate entity not found",
		})
		return
	}

	// Validate that both private key and certificate are available
	if entity.EncryptedPrivateKey == "" || entity.Certificate == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad Request",
			"message": "Both private key and certificate must be available to generate PFX",
		})
		return
	}

	// Generate PFX
	pfxData, err := h.cryptoService.GeneratePFX(entity.EncryptedPrivateKey, entity.Certificate, req.Password)
	if err != nil {
		h.logger.WithError(err).WithField("entity_id", entityID).Error("Failed to generate PFX")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal Server Error",
			"message": "Failed to generate PFX file",
			"details": err.Error(),
		})
		return
	}

	// Encode PFX data as base64
	pfxBase64 := h.cryptoService.EncodeToBase64(pfxData)

	// Generate filename
	filename := fmt.Sprintf("%s-%s.pfx", entity.CommonName, entityID[:8])

	// Prepare response
	response := models.GeneratePFXResponse{
		ID:       entityID,
		PFXData:  pfxBase64,
		Filename: filename,
	}

	h.logger.WithFields(logrus.Fields{
		"entity_id":   entityID,
		"common_name": entity.CommonName,
		"filename":    filename,
	}).Info("PFX file generated successfully")

	c.JSON(http.StatusOK, response)
}

// GetCertificate retrieves a certificate entity by ID
// @Summary Get certificate by ID
// @Description Retrieves a specific certificate entity including its private key, CSR, and certificate details
// @Tags Certificate Management
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Security BearerAuth
// @Param id path string true "Certificate ID (UUID format)"
// @Success 200 {object} models.CertificateEntity "Certificate entity details"
// @Failure 400 {object} map[string]interface{} "Bad request - invalid ID format"
// @Failure 401 {object} map[string]interface{} "Unauthorized - invalid or missing API key"
// @Failure 404 {object} map[string]interface{} "Certificate not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /keys/{id} [get]
func (h *CertificateHandler) GetCertificate(c *gin.Context) {
	entityID := c.Param("id")
	if entityID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad Request",
			"message": "Entity ID is required",
		})
		return
	}

	// Retrieve entity
	entity, err := h.storage.GetCertificateEntity(c.Request.Context(), entityID)
	if err != nil {
		h.logger.WithError(err).WithField("entity_id", entityID).Error("Failed to retrieve certificate entity")
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Not Found",
			"message": "Certificate entity not found",
		})
		return
	}

	// Remove sensitive data from response
	entity.EncryptedPrivateKey = "[REDACTED]"

	h.logger.WithField("entity_id", entityID).Debug("Certificate entity retrieved")

	c.JSON(http.StatusOK, entity)
}

// ListCertificates retrieves a list of certificates with optional filtering
// @Summary List certificates with filtering
// @Description Retrieves a paginated list of certificate entities with optional filtering by tags, status, key type, and date range
// @Tags Certificate Management
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Security BearerAuth
// @Param status query string false "Filter by certificate status" Enums(CSR_CREATED, CERT_UPLOADED, EXPIRED, REVOKED)
// @Param key_type query string false "Filter by key type" Enums(RSA2048, RSA4096, ECDSA-P256, ECDSA-P384)
// @Param date_from query string false "Filter certificates created after this date (RFC3339 format)"
// @Param date_to query string false "Filter certificates created before this date (RFC3339 format)"
// @Param page query int false "Page number for pagination (default: 1)" minimum(1)
// @Param page_size query int false "Number of items per page (default: 50, max: 100)" minimum(1) maximum(100)
// @Param environment query string false "Filter by environment tag"
// @Param project query string false "Filter by project tag"
// @Param team query string false "Filter by team tag"
// @Success 200 {object} models.ListKeysResponse "List of certificate entities"
// @Failure 401 {object} map[string]interface{} "Unauthorized - invalid or missing API key"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /keys [get]
func (h *CertificateHandler) ListCertificates(c *gin.Context) {
	// Parse query parameters
	var filters models.SearchFilters

	// Status filter
	if status := c.Query("status"); status != "" {
		filters.Status = models.CertificateStatus(status)
	}

	// Key type filter
	if keyType := c.Query("key_type"); keyType != "" {
		filters.KeyType = models.KeyType(keyType)
	}

	// Date filters
	if dateFrom := c.Query("date_from"); dateFrom != "" {
		if parsedDate, err := time.Parse(time.RFC3339, dateFrom); err == nil {
			filters.DateFrom = &parsedDate
		}
	}

	if dateTo := c.Query("date_to"); dateTo != "" {
		if parsedDate, err := time.Parse(time.RFC3339, dateTo); err == nil {
			filters.DateTo = &parsedDate
		}
	}

	// Pagination
	if page := c.Query("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil && p > 0 {
			filters.Page = p
		}
	}

	if pageSize := c.Query("page_size"); pageSize != "" {
		if ps, err := strconv.Atoi(pageSize); err == nil && ps > 0 && ps <= 100 {
			filters.PageSize = ps
		}
	}

	// Tag filters - expecting format: tag_key=tag_value
	filters.Tags = make(map[string]string)
	for key, values := range c.Request.URL.Query() {
		if len(values) > 0 && key != "status" && key != "key_type" && key != "date_from" && key != "date_to" && key != "page" && key != "page_size" {
			filters.Tags[key] = values[0]
		}
	}

	// Retrieve entities
	entities, err := h.storage.ListCertificateEntities(c.Request.Context(), filters)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list certificate entities")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal Server Error",
			"message": "Failed to retrieve certificate list",
		})
		return
	}

	// Remove sensitive data from response
	for i := range entities {
		entities[i].EncryptedPrivateKey = "[REDACTED]"
	}

	// Prepare response
	response := models.ListKeysResponse{
		Keys:       entities,
		TotalCount: len(entities),
		Page:       filters.Page,
		PageSize:   filters.PageSize,
	}

	h.logger.WithFields(logrus.Fields{
		"count":     len(entities),
		"page":      filters.Page,
		"page_size": filters.PageSize,
	}).Debug("Certificate entities listed")

	c.JSON(http.StatusOK, response)
}

// ExportPrivateKey exports the private key for a certificate entity
// @Summary Export private key (SENSITIVE OPERATION)
// @Description Exports the decrypted private key in PEM format. WARNING: This operation exposes sensitive cryptographic material and should be used with extreme caution. Ensure proper access controls and audit logging.
// @Tags Certificate Management
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Security BearerAuth
// @Param id path string true "Certificate entity ID (UUID format)"
// @Success 200 {object} models.ExportPrivateKeyResponse "Private key exported successfully"
// @Failure 400 {object} map[string]interface{} "Bad request - invalid ID format"
// @Failure 401 {object} map[string]interface{} "Unauthorized - invalid or missing API key"
// @Failure 404 {object} map[string]interface{} "Certificate entity not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /keys/{id}/private-key [get]
func (h *CertificateHandler) ExportPrivateKey(c *gin.Context) {
	entityID := c.Param("id")
	if entityID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad Request",
			"message": "Entity ID is required",
		})
		return
	}

	// Retrieve entity
	entity, err := h.storage.GetCertificateEntity(c.Request.Context(), entityID)
	if err != nil {
		h.logger.WithError(err).WithField("entity_id", entityID).Error("Failed to retrieve certificate entity")
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Not Found",
			"message": "Certificate entity not found",
		})
		return
	}

	// Validate that private key exists
	if entity.EncryptedPrivateKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad Request",
			"message": "No private key available for this certificate entity",
		})
		return
	}

	// Log the private key export for audit purposes
	h.logger.WithFields(logrus.Fields{
		"entity_id":   entityID,
		"common_name": entity.CommonName,
		"key_type":    entity.KeyType,
		"operation":   "export_private_key",
		"user_agent":  c.GetHeader("User-Agent"),
		"remote_addr": c.ClientIP(),
		"request_id":  c.GetString("request_id"),
	}).Warn("SENSITIVE: Private key exported")

	// Prepare response
	response := models.ExportPrivateKeyResponse{
		ID:         entityID,
		PrivateKey: entity.EncryptedPrivateKey, // Note: This is actually the decrypted private key in PEM format
		KeyType:    entity.KeyType,
		CommonName: entity.CommonName,
		ExportedAt: time.Now().Format(time.RFC3339),
	}

	h.logger.WithFields(logrus.Fields{
		"entity_id":   entityID,
		"common_name": entity.CommonName,
		"key_type":    entity.KeyType,
	}).Info("Private key export completed")

	c.JSON(http.StatusOK, response)
}
