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

	// Apply pagination
	if filters.PageSize > 0 {
		// Ensure PageSize doesn't exceed int32 max value to prevent overflow
		if filters.PageSize > 2147483647 {
			input.Limit = aws.Int32(2147483647) // Max int32 value
		} else {
			input.Limit = aws.Int32(int32(filters.PageSize))
		}
	} else {
		input.Limit = aws.Int32(50) // Default page size
	}

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

	return entities, nil
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
