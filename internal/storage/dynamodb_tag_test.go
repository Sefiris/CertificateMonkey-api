package storage

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"certificate-monkey/internal/models"
)

// TestTagStorageFormat demonstrates how tags are stored in DynamoDB
func TestTagStorageFormat(t *testing.T) {
	// Create a test entity with tags
	entity := &models.CertificateEntity{
		ID:         "test-123",
		CommonName: "example.com",
		KeyType:    models.KeyTypeRSA2048,
		Status:     models.StatusCSRCreated,
		Tags: map[string]string{
			"environment": "dev",
			"project":     "api-gateway",
			"team":        "platform",
			"cost-center": "IT-001",
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Marshal to DynamoDB format
	av, err := attributevalue.MarshalMap(entity)
	require.NoError(t, err)

	// Check that tags are stored as a Map
	tagsAttr, exists := av["tags"]
	require.True(t, exists, "Tags should exist in marshaled data")

	// The tags should be stored as a Map type
	mapAttr, ok := tagsAttr.(*types.AttributeValueMemberM)
	require.True(t, ok, "Tags should be stored as DynamoDB Map type")

	// Verify individual tag values
	envAttr, exists := mapAttr.Value["environment"]
	require.True(t, exists, "Environment tag should exist")

	envStr, ok := envAttr.(*types.AttributeValueMemberS)
	require.True(t, ok, "Environment tag should be a String type")
	assert.Equal(t, "dev", envStr.Value)

	// Convert to JSON to see the actual DynamoDB format
	jsonBytes, err := json.MarshalIndent(av, "", "  ")
	require.NoError(t, err)

	t.Logf("DynamoDB storage format:\n%s", string(jsonBytes))

	// Unmarshal back to verify round-trip
	var unmarshaledEntity models.CertificateEntity
	err = attributevalue.UnmarshalMap(av, &unmarshaledEntity)
	require.NoError(t, err)

	// Verify tags are correctly unmarshaled
	assert.Equal(t, entity.Tags, unmarshaledEntity.Tags)
	assert.Equal(t, "dev", unmarshaledEntity.Tags["environment"])
	assert.Equal(t, "api-gateway", unmarshaledEntity.Tags["project"])
}

// TestTagFilteringFormat tests how tag filtering works in DynamoDB expressions
func TestTagFilteringFormat(t *testing.T) {
	// This test demonstrates the correct filter expression format for tags
	filters := models.SearchFilters{
		Tags: map[string]string{
			"environment": "production",
			"project":     "web-server",
		},
	}

	// Build filter expressions like the real implementation
	var filterExpressions []string
	expressionAttributeNames := make(map[string]string)
	expressionAttributeValues := make(map[string]types.AttributeValue)

	// Add tag filters - INCLUDING THE FIX
	if len(filters.Tags) > 0 {
		// Define #tags attribute name once for all tag filters
		expressionAttributeNames["#tags"] = "tags"
	}

	tagIndex := 0
	for tagKey, tagValue := range filters.Tags {
		// This is the CORRECT way to filter tags in DynamoDB
		filterExpressions = append(filterExpressions,
			fmt.Sprintf("#tags.#tag_key_%d = :tag_value_%d", tagIndex, tagIndex))
		expressionAttributeNames[fmt.Sprintf("#tag_key_%d", tagIndex)] = tagKey
		expressionAttributeValues[fmt.Sprintf(":tag_value_%d", tagIndex)] = &types.AttributeValueMemberS{Value: tagValue}
		tagIndex++
	}

	t.Logf("Filter expressions: %v", filterExpressions)
	t.Logf("Attribute names: %v", expressionAttributeNames)
	t.Logf("Attribute values: %v", expressionAttributeValues)

	// Verify the filter expressions are correct
	expectedFilterPattern := "#tags.#tag_key_"
	assert.Contains(t, filterExpressions[0], expectedFilterPattern,
		"Filter should reference tags as a nested attribute")

	// CRITICAL: Verify that #tags is defined in attribute names
	_, exists := expressionAttributeNames["#tags"]
	assert.True(t, exists, "#tags must be defined in expressionAttributeNames")
	assert.Equal(t, "tags", expressionAttributeNames["#tags"], "#tags should map to 'tags'")

	// Verify tag key mappings
	assert.Equal(t, "environment", expressionAttributeNames["#tag_key_0"])
	assert.Equal(t, "project", expressionAttributeNames["#tag_key_1"])

	// Verify tag value mappings
	envValue, exists := expressionAttributeValues[":tag_value_0"]
	assert.True(t, exists, ":tag_value_0 should exist")
	envStr, ok := envValue.(*types.AttributeValueMemberS)
	assert.True(t, ok, ":tag_value_0 should be a string")
	assert.Equal(t, "production", envStr.Value)
}

// TestIncorrectTagStorageDetection helps identify if tags are being stored incorrectly
func TestIncorrectTagStorageDetection(t *testing.T) {
	t.Log("=== How to identify incorrect tag storage ===")

	t.Log("✅ CORRECT DynamoDB format:")
	t.Log(`{
  "id": {"S": "123e4567-e89b-12d3-a456-426614174000"},
  "common_name": {"S": "example.com"},
  "tags": {
    "M": {
      "environment": {"S": "dev"},
      "project": {"S": "api-gateway"}
    }
  }
}`)

	t.Log("❌ INCORRECT format (tags flattened to top level):")
	t.Log(`{
  "id": {"S": "123e4567-e89b-12d3-a456-426614174000"},
  "common_name": {"S": "example.com"},
  "environment": {"S": "dev"},
  "project": {"S": "api-gateway"}
}`)

	t.Log("🔍 If you see tags at the top level, there's a storage issue!")
}

// TestDynamoDBJSONFormats shows different ways to view DynamoDB data
func TestDynamoDBJSONFormats(t *testing.T) {
	t.Log("=== DynamoDB JSON Formats ===")

	t.Log("1. DynamoDB JSON (what you see in AWS CLI/Console):")
	t.Log(`{
  "tags": {
    "M": {
      "environment": {"S": "dev"}
    }
  }
}`)

	t.Log("2. Document Format (what your application sees):")
	t.Log(`{
  "tags": {
    "environment": "dev"
  }
}`)

	t.Log("3. If using AWS CLI with --output table or --no-cli-pager:")
	t.Log("   You might see a simplified view that looks like:")
	t.Log(`   environment: dev`)

	t.Log("The 'S' and 'M' type indicators are DynamoDB's internal representation")
	t.Log("and are automatically handled by the AWS SDK")
}
