package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/sirupsen/logrus"

	"certificate-monkey/internal/config"
	"certificate-monkey/internal/models"
)

// DynamoDBStorage handles all DynamoDB operations
type DynamoDBStorage struct {
	client    *dynamodb.Client
	kmsClient *kms.Client
	tableName string
	kmsKeyID  string
	logger    *logrus.Logger
}

// NewDynamoDBStorage creates a new DynamoDB storage instance
func NewDynamoDBStorage(client *dynamodb.Client, kmsClient *kms.Client, cfg *config.Config, logger *logrus.Logger) *DynamoDBStorage {
	return &DynamoDBStorage{
		client:    client,
		kmsClient: kmsClient,
		tableName: cfg.AWS.DynamoDBTable,
		kmsKeyID:  cfg.AWS.KMSKeyID,
		logger:    logger,
	}
}

// CreateCertificateEntity stores a new certificate entity in DynamoDB
func (d *DynamoDBStorage) CreateCertificateEntity(ctx context.Context, entity *models.CertificateEntity) error {
	// Encrypt the private key using KMS
	encryptedPrivateKey, err := d.encryptData(ctx, entity.EncryptedPrivateKey)
	if err != nil {
		return fmt.Errorf("failed to encrypt private key: %w", err)
	}

	// Create a copy with encrypted private key
	entityToStore := *entity
	entityToStore.EncryptedPrivateKey = encryptedPrivateKey

	// Convert to DynamoDB attribute value
	av, err := attributevalue.MarshalMap(entityToStore)
	if err != nil {
		return fmt.Errorf("failed to marshal entity: %w", err)
	}

	// Put item in DynamoDB
	input := &dynamodb.PutItemInput{
		TableName:           aws.String(d.tableName),
		Item:                av,
		ConditionExpression: aws.String("attribute_not_exists(id)"),
	}

	_, err = d.client.PutItem(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to put item in DynamoDB: %w", err)
	}

	d.logger.WithFields(logrus.Fields{
		"entity_id":   entity.ID,
		"common_name": entity.CommonName,
		"key_type":    entity.KeyType,
	}).Info("Certificate entity created successfully")

	return nil
}

// GetCertificateEntity retrieves a certificate entity by ID
func (d *DynamoDBStorage) GetCertificateEntity(ctx context.Context, id string) (*models.CertificateEntity, error) {
	input := &dynamodb.GetItemInput{
		TableName: aws.String(d.tableName),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		},
	}

	result, err := d.client.GetItem(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to get item from DynamoDB: %w", err)
	}

	if result.Item == nil {
		return nil, fmt.Errorf("certificate entity not found")
	}

	// Unmarshal the result
	var entity models.CertificateEntity
	err = attributevalue.UnmarshalMap(result.Item, &entity)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal entity: %w", err)
	}

	// Decrypt the private key
	decryptedPrivateKey, err := d.decryptData(ctx, entity.EncryptedPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt private key: %w", err)
	}
	entity.EncryptedPrivateKey = decryptedPrivateKey

	return &entity, nil
}

// UpdateCertificateEntity updates an existing certificate entity
func (d *DynamoDBStorage) UpdateCertificateEntity(ctx context.Context, entity *models.CertificateEntity) error {
	// Encrypt the private key if it's not already encrypted
	encryptedPrivateKey := entity.EncryptedPrivateKey
	if entity.EncryptedPrivateKey != "" {
		var err error
		encryptedPrivateKey, err = d.encryptData(ctx, entity.EncryptedPrivateKey)
		if err != nil {
			return fmt.Errorf("failed to encrypt private key: %w", err)
		}
	}

	// Update timestamp
	entity.UpdatedAt = time.Now()

	// Build update expression
	updateExpression := "SET #status = :status, #updated_at = :updated_at"
	expressionAttributeNames := map[string]string{
		"#status":     "status",
		"#updated_at": "updated_at",
	}
	expressionAttributeValues := map[string]types.AttributeValue{
		":status":     &types.AttributeValueMemberS{Value: string(entity.Status)},
		":updated_at": &types.AttributeValueMemberS{Value: entity.UpdatedAt.Format(time.RFC3339)},
	}

	// Add certificate fields if present
	if entity.Certificate != "" {
		updateExpression += ", #certificate = :certificate"
		expressionAttributeNames["#certificate"] = "certificate"
		expressionAttributeValues[":certificate"] = &types.AttributeValueMemberS{Value: entity.Certificate}
	}

	if entity.ValidFrom != nil {
		updateExpression += ", #valid_from = :valid_from"
		expressionAttributeNames["#valid_from"] = "valid_from"
		expressionAttributeValues[":valid_from"] = &types.AttributeValueMemberS{Value: entity.ValidFrom.Format(time.RFC3339)}
	}

	if entity.ValidTo != nil {
		updateExpression += ", #valid_to = :valid_to"
		expressionAttributeNames["#valid_to"] = "valid_to"
		expressionAttributeValues[":valid_to"] = &types.AttributeValueMemberS{Value: entity.ValidTo.Format(time.RFC3339)}
	}

	if entity.SerialNumber != "" {
		updateExpression += ", #serial_number = :serial_number"
		expressionAttributeNames["#serial_number"] = "serial_number"
		expressionAttributeValues[":serial_number"] = &types.AttributeValueMemberS{Value: entity.SerialNumber}
	}

	if entity.Fingerprint != "" {
		updateExpression += ", #fingerprint = :fingerprint"
		expressionAttributeNames["#fingerprint"] = "fingerprint"
		expressionAttributeValues[":fingerprint"] = &types.AttributeValueMemberS{Value: entity.Fingerprint}
	}

	if encryptedPrivateKey != "" {
		updateExpression += ", #encrypted_private_key = :encrypted_private_key"
		expressionAttributeNames["#encrypted_private_key"] = "encrypted_private_key"
		expressionAttributeValues[":encrypted_private_key"] = &types.AttributeValueMemberS{Value: encryptedPrivateKey}
	}

	// Perform the update
	input := &dynamodb.UpdateItemInput{
		TableName: aws.String(d.tableName),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: entity.ID},
		},
		UpdateExpression:          aws.String(updateExpression),
		ExpressionAttributeNames:  expressionAttributeNames,
		ExpressionAttributeValues: expressionAttributeValues,
		ConditionExpression:       aws.String("attribute_exists(id)"),
	}

	_, err := d.client.UpdateItem(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to update item in DynamoDB: %w", err)
	}

	d.logger.WithFields(logrus.Fields{
		"entity_id": entity.ID,
		"status":    entity.Status,
	}).Info("Certificate entity updated successfully")

	return nil
}

// ListCertificateEntities retrieves certificate entities with optional filtering
func (d *DynamoDBStorage) ListCertificateEntities(ctx context.Context, filters models.SearchFilters) ([]models.CertificateEntity, error) {
	input := &dynamodb.ScanInput{
		TableName: aws.String(d.tableName),
	}

	// Apply filters if provided
	var filterExpressions []string
	expressionAttributeNames := make(map[string]string)
	expressionAttributeValues := make(map[string]types.AttributeValue)

	if filters.Status != "" {
		filterExpressions = append(filterExpressions, "#status = :status")
		expressionAttributeNames["#status"] = "status"
		expressionAttributeValues[":status"] = &types.AttributeValueMemberS{Value: string(filters.Status)}
	}

	if filters.KeyType != "" {
		filterExpressions = append(filterExpressions, "#key_type = :key_type")
		expressionAttributeNames["#key_type"] = "key_type"
		expressionAttributeValues[":key_type"] = &types.AttributeValueMemberS{Value: string(filters.KeyType)}
	}

	if filters.DateFrom != nil {
		filterExpressions = append(filterExpressions, "#created_at >= :date_from")
		expressionAttributeNames["#created_at"] = "created_at"
		expressionAttributeValues[":date_from"] = &types.AttributeValueMemberS{Value: filters.DateFrom.Format(time.RFC3339)}
	}

	if filters.DateTo != nil {
		filterExpressions = append(filterExpressions, "#created_at <= :date_to")
		expressionAttributeNames["#created_at"] = "created_at"
		expressionAttributeValues[":date_to"] = &types.AttributeValueMemberS{Value: filters.DateTo.Format(time.RFC3339)}
	}

	// Add tag filters
	if len(filters.Tags) > 0 {
		// Define #tags attribute name once for all tag filters
		expressionAttributeNames["#tags"] = "tags"
	}

	tagIndex := 0
	for tagKey, tagValue := range filters.Tags {
		filterExpressions = append(filterExpressions, fmt.Sprintf("#tags.#tag_key_%d = :tag_value_%d", tagIndex, tagIndex))
		expressionAttributeNames[fmt.Sprintf("#tag_key_%d", tagIndex)] = tagKey
		expressionAttributeValues[fmt.Sprintf(":tag_value_%d", tagIndex)] = &types.AttributeValueMemberS{Value: tagValue}
		tagIndex++
	}

	if len(filterExpressions) > 0 {
		filterExpression := ""
		for i, expr := range filterExpressions {
			if i > 0 {
				filterExpression += " AND "
			}
			filterExpression += expr
		}
		input.FilterExpression = aws.String(filterExpression)
		input.ExpressionAttributeNames = expressionAttributeNames
		input.ExpressionAttributeValues = expressionAttributeValues
	}

	// Note: We'll retrieve all matching items first, then sort and paginate in memory
	// This is because DynamoDB Scan doesn't support sorting by arbitrary fields
	result, err := d.client.Scan(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to scan DynamoDB table: %w", err)
	}

	// Unmarshal results
	var entities []models.CertificateEntity
	for _, item := range result.Items {
		var entity models.CertificateEntity
		err = attributevalue.UnmarshalMap(item, &entity)
		if err != nil {
			d.logger.WithError(err).Error("Failed to unmarshal certificate entity")
			continue
		}

		// Decrypt the private key
		if entity.EncryptedPrivateKey != "" {
			decryptedPrivateKey, err := d.decryptData(ctx, entity.EncryptedPrivateKey)
			if err != nil {
				d.logger.WithError(err).WithField("entity_id", entity.ID).Error("Failed to decrypt private key")
				continue
			}
			entity.EncryptedPrivateKey = decryptedPrivateKey
		}

		entities = append(entities, entity)
	}

	// Apply sorting
	d.sortEntities(entities, filters.SortBy, filters.SortOrder)

	// Apply pagination after sorting
	totalCount := len(entities)
	page := filters.Page
	pageSize := filters.PageSize

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 50
	}

	startIndex := (page - 1) * pageSize
	endIndex := startIndex + pageSize

	if startIndex >= totalCount {
		return []models.CertificateEntity{}, nil
	}

	if endIndex > totalCount {
		endIndex = totalCount
	}

	return entities[startIndex:endIndex], nil
}

// GetCertificateEntityCount returns the total count of entities matching the filters
func (d *DynamoDBStorage) GetCertificateEntityCount(ctx context.Context, filters models.SearchFilters) (int, error) {
	input := &dynamodb.ScanInput{
		TableName: aws.String(d.tableName),
		Select:    types.SelectCount, // Only count, don't return items
	}

	// Apply the same filters as in ListCertificateEntities
	var filterExpressions []string
	expressionAttributeNames := make(map[string]string)
	expressionAttributeValues := make(map[string]types.AttributeValue)

	if filters.Status != "" {
		filterExpressions = append(filterExpressions, "#status = :status")
		expressionAttributeNames["#status"] = "status"
		expressionAttributeValues[":status"] = &types.AttributeValueMemberS{Value: string(filters.Status)}
	}

	if filters.KeyType != "" {
		filterExpressions = append(filterExpressions, "#key_type = :key_type")
		expressionAttributeNames["#key_type"] = "key_type"
		expressionAttributeValues[":key_type"] = &types.AttributeValueMemberS{Value: string(filters.KeyType)}
	}

	if filters.DateFrom != nil {
		filterExpressions = append(filterExpressions, "#created_at >= :date_from")
		expressionAttributeNames["#created_at"] = "created_at"
		expressionAttributeValues[":date_from"] = &types.AttributeValueMemberS{Value: filters.DateFrom.Format(time.RFC3339)}
	}

	if filters.DateTo != nil {
		filterExpressions = append(filterExpressions, "#created_at <= :date_to")
		expressionAttributeNames["#created_at"] = "created_at"
		expressionAttributeValues[":date_to"] = &types.AttributeValueMemberS{Value: filters.DateTo.Format(time.RFC3339)}
	}

	// Add tag filters
	if len(filters.Tags) > 0 {
		expressionAttributeNames["#tags"] = "tags"
	}

	tagIndex := 0
	for tagKey, tagValue := range filters.Tags {
		filterExpressions = append(filterExpressions, fmt.Sprintf("#tags.#tag_key_%d = :tag_value_%d", tagIndex, tagIndex))
		expressionAttributeNames[fmt.Sprintf("#tag_key_%d", tagIndex)] = tagKey
		expressionAttributeValues[fmt.Sprintf(":tag_value_%d", tagIndex)] = &types.AttributeValueMemberS{Value: tagValue}
		tagIndex++
	}

	if len(filterExpressions) > 0 {
		filterExpression := ""
		for i, expr := range filterExpressions {
			if i > 0 {
				filterExpression += " AND "
			}
			filterExpression += expr
		}
		input.FilterExpression = aws.String(filterExpression)
		input.ExpressionAttributeNames = expressionAttributeNames
		input.ExpressionAttributeValues = expressionAttributeValues
	}

	result, err := d.client.Scan(ctx, input)
	if err != nil {
		return 0, fmt.Errorf("failed to count items in DynamoDB table: %w", err)
	}

	return int(result.Count), nil
}

// sortEntities sorts the entities slice in-place based on the specified field and order
func (d *DynamoDBStorage) sortEntities(entities []models.CertificateEntity, sortBy, sortOrder string) {
	if len(entities) <= 1 {
		return
	}

	// Import sort package at the top of the file
	// sort.Slice(entities, func(i, j int) bool {
	// 	return d.compareEntities(entities[i], entities[j], sortBy, sortOrder)
	// })

	// Implement sorting using a simple approach
	for i := 0; i < len(entities)-1; i++ {
		for j := i + 1; j < len(entities); j++ {
			shouldSwap := d.compareEntities(entities[i], entities[j], sortBy, sortOrder)
			if shouldSwap {
				entities[i], entities[j] = entities[j], entities[i]
			}
		}
	}
}

// compareEntities compares two entities based on the sort field and order
// Returns true if entity i should come after entity j in the sorted order
func (d *DynamoDBStorage) compareEntities(entityI, entityJ models.CertificateEntity, sortBy, sortOrder string) bool {
	var comparison int

	switch sortBy {
	case "created_at":
		if entityI.CreatedAt.Before(entityJ.CreatedAt) {
			comparison = -1
		} else if entityI.CreatedAt.After(entityJ.CreatedAt) {
			comparison = 1
		} else {
			comparison = 0
		}
	case "updated_at":
		if entityI.UpdatedAt.Before(entityJ.UpdatedAt) {
			comparison = -1
		} else if entityI.UpdatedAt.After(entityJ.UpdatedAt) {
			comparison = 1
		} else {
			comparison = 0
		}
	case "common_name":
		if entityI.CommonName < entityJ.CommonName {
			comparison = -1
		} else if entityI.CommonName > entityJ.CommonName {
			comparison = 1
		} else {
			comparison = 0
		}
	case "status":
		statusI := string(entityI.Status)
		statusJ := string(entityJ.Status)
		if statusI < statusJ {
			comparison = -1
		} else if statusI > statusJ {
			comparison = 1
		} else {
			comparison = 0
		}
	case "key_type":
		keyTypeI := string(entityI.KeyType)
		keyTypeJ := string(entityJ.KeyType)
		if keyTypeI < keyTypeJ {
			comparison = -1
		} else if keyTypeI > keyTypeJ {
			comparison = 1
		} else {
			comparison = 0
		}
	case "valid_to":
		// Handle nil values
		if entityI.ValidTo == nil && entityJ.ValidTo == nil {
			comparison = 0
		} else if entityI.ValidTo == nil {
			comparison = -1 // nil comes first
		} else if entityJ.ValidTo == nil {
			comparison = 1
		} else if entityI.ValidTo.Before(*entityJ.ValidTo) {
			comparison = -1
		} else if entityI.ValidTo.After(*entityJ.ValidTo) {
			comparison = 1
		} else {
			comparison = 0
		}
	case "valid_from":
		// Handle nil values
		if entityI.ValidFrom == nil && entityJ.ValidFrom == nil {
			comparison = 0
		} else if entityI.ValidFrom == nil {
			comparison = -1 // nil comes first
		} else if entityJ.ValidFrom == nil {
			comparison = 1
		} else if entityI.ValidFrom.Before(*entityJ.ValidFrom) {
			comparison = -1
		} else if entityI.ValidFrom.After(*entityJ.ValidFrom) {
			comparison = 1
		} else {
			comparison = 0
		}
	default:
		// Default to created_at sorting
		if entityI.CreatedAt.Before(entityJ.CreatedAt) {
			comparison = -1
		} else if entityI.CreatedAt.After(entityJ.CreatedAt) {
			comparison = 1
		} else {
			comparison = 0
		}
	}

	// Apply sort order
	if sortOrder == "desc" {
		comparison = -comparison
	}

	return comparison > 0
}

// DeleteCertificateEntity deletes a certificate entity by ID
func (d *DynamoDBStorage) DeleteCertificateEntity(ctx context.Context, id string) error {
	input := &dynamodb.DeleteItemInput{
		TableName: aws.String(d.tableName),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		},
		ConditionExpression: aws.String("attribute_exists(id)"),
	}

	_, err := d.client.DeleteItem(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to delete item from DynamoDB: %w", err)
	}

	d.logger.WithField("entity_id", id).Info("Certificate entity deleted successfully")
	return nil
}

// encryptData encrypts data using AWS KMS
func (d *DynamoDBStorage) encryptData(ctx context.Context, plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}

	input := &kms.EncryptInput{
		KeyId:     aws.String(d.kmsKeyID),
		Plaintext: []byte(plaintext),
	}

	result, err := d.kmsClient.Encrypt(ctx, input)
	if err != nil {
		return "", err
	}

	// Encode the encrypted data as base64
	return fmt.Sprintf("%x", result.CiphertextBlob), nil
}

// decryptData decrypts data using AWS KMS
func (d *DynamoDBStorage) decryptData(ctx context.Context, encryptedData string) (string, error) {
	if encryptedData == "" {
		return "", nil
	}

	// Decode from hex
	ciphertext := make([]byte, len(encryptedData)/2)
	_, err := fmt.Sscanf(encryptedData, "%x", &ciphertext)
	if err != nil {
		return "", fmt.Errorf("failed to decode encrypted data: %w", err)
	}

	input := &kms.DecryptInput{
		CiphertextBlob: ciphertext,
	}

	result, err := d.kmsClient.Decrypt(ctx, input)
	if err != nil {
		return "", err
	}

	return string(result.Plaintext), nil
}
