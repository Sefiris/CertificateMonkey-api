# Tag Search Optimization Guide

## Current Implementation vs. Scalable Solutions

Certificate Monkey currently supports tag-based searching using DynamoDB FilterExpression with Scan operations. While this works for small to medium datasets, here are optimization strategies for production scale.

## Current Implementation (Works for < 10K certificates)

```go
// Current approach in internal/storage/dynamodb.go
for tagKey, tagValue := range filters.Tags {
    filterExpressions = append(filterExpressions,
        fmt.Sprintf("#tags.#tag_key_%d = :tag_value_%d", tagIndex, tagIndex))
    // Uses Scan + FilterExpression
}
```

**Pros:**
- ✅ Simple implementation
- ✅ Supports any tag combination
- ✅ No additional infrastructure

**Cons:**
- ❌ Full table scan (expensive)
- ❌ Performance degrades with table size
- ❌ High RCU consumption
- ❌ Limited by DynamoDB scan limits

## Production-Scale Solutions

### Option 1: Tag-Based GSI (Recommended for Common Tags)

Create GSIs for frequently searched tags:

```hcl
# Terraform example - Add to existing table
resource "aws_dynamodb_table" "certificate_monkey" {
  # ... existing configuration ...

  # GSI for environment tag
  global_secondary_index {
    name     = "environment-created_at-index"
    hash_key = "environment_tag"
    range_key = "created_at"
    projection_type = "ALL"
  }

  # GSI for project tag
  global_secondary_index {
    name     = "project-created_at-index"
    hash_key = "project_tag"
    range_key = "created_at"
    projection_type = "ALL"
  }
}
```

**Implementation Changes:**

```go
// Update CertificateEntity model
type CertificateEntity struct {
    // ... existing fields ...

    // Add computed tag fields for indexing
    EnvironmentTag string `json:"environment_tag,omitempty" dynamodbav:"environment_tag,omitempty"`
    ProjectTag     string `json:"project_tag,omitempty" dynamodbav:"project_tag,omitempty"`
    TeamTag        string `json:"team_tag,omitempty" dynamodbav:"team_tag,omitempty"`
}

// Update storage layer
func (d *DynamoDBStorage) CreateCertificateEntity(ctx context.Context, entity *models.CertificateEntity) error {
    // Populate searchable tag fields
    if env, exists := entity.Tags["environment"]; exists {
        entity.EnvironmentTag = env
    }
    if project, exists := entity.Tags["project"]; exists {
        entity.ProjectTag = project
    }
    // ... continue for other common tags

    // Store as usual
}

// Optimized search function
func (d *DynamoDBStorage) ListCertificatesByEnvironment(ctx context.Context, environment string) ([]models.CertificateEntity, error) {
    input := &dynamodb.QueryInput{
        TableName:              aws.String(d.tableName),
        IndexName:              aws.String("environment-created_at-index"),
        KeyConditionExpression: aws.String("environment_tag = :env"),
        ExpressionAttributeValues: map[string]types.AttributeValue{
            ":env": &types.AttributeValueMemberS{Value: environment},
        },
    }

    // Uses Query instead of Scan - much faster!
    result, err := d.client.Query(ctx, input)
    // ... handle result
}
```

### Option 2: Inverted Index Pattern

Create a separate table for tag-to-certificate mappings:

```bash
# Primary table: certificate-monkey (existing)
# New table: certificate-tags

# certificate-tags table structure:
# PK: tag_key#tag_value (e.g., "environment#production")
# SK: certificate_id
# Attributes: certificate_id, tag_key, tag_value, created_at
```

**Implementation:**

```go
type TagEntity struct {
    TagKeyValue     string `dynamodbav:"pk"`           // "environment#production"
    CertificateID   string `dynamodbav:"sk"`           // certificate UUID
    TagKey          string `dynamodbav:"tag_key"`      // "environment"
    TagValue        string `dynamodbav:"tag_value"`    // "production"
    CreatedAt       string `dynamodbav:"created_at"`   // timestamp
}

func (d *DynamoDBStorage) CreateCertificateEntity(ctx context.Context, entity *models.CertificateEntity) error {
    // 1. Create main certificate entity
    err := d.createMainEntity(ctx, entity)
    if err != nil {
        return err
    }

    // 2. Create tag index entries
    for tagKey, tagValue := range entity.Tags {
        tagEntity := TagEntity{
            TagKeyValue:   fmt.Sprintf("%s#%s", tagKey, tagValue),
            CertificateID: entity.ID,
            TagKey:        tagKey,
            TagValue:      tagValue,
            CreatedAt:     entity.CreatedAt.Format(time.RFC3339),
        }

        err = d.createTagEntity(ctx, tagEntity)
        if err != nil {
            // Consider implementing cleanup/rollback
            d.logger.WithError(err).Error("Failed to create tag entity")
        }
    }

    return nil
}

func (d *DynamoDBStorage) SearchByTag(ctx context.Context, tagKey, tagValue string) ([]string, error) {
    input := &dynamodb.QueryInput{
        TableName:              aws.String("certificate-tags"),
        KeyConditionExpression: aws.String("pk = :tag_key_value"),
        ExpressionAttributeValues: map[string]types.AttributeValue{
            ":tag_key_value": &types.AttributeValueMemberS{Value: fmt.Sprintf("%s#%s", tagKey, tagValue)},
        },
    }

    result, err := d.client.Query(ctx, input)
    if err != nil {
        return nil, err
    }

    var certificateIDs []string
    for _, item := range result.Items {
        var tagEntity TagEntity
        err = attributevalue.UnmarshalMap(item, &tagEntity)
        if err != nil {
            continue
        }
        certificateIDs = append(certificateIDs, tagEntity.CertificateID)
    }

    return certificateIDs, nil
}
```

### Option 3: Composite Search Attributes

Add computed search fields to the main table:

```go
type CertificateEntity struct {
    // ... existing fields ...

    // Composite search fields
    SearchTags    []string `json:"search_tags,omitempty" dynamodbav:"search_tags,omitempty"`       // ["environment:production", "project:api"]
    TaggedStatus  string   `json:"tagged_status,omitempty" dynamodbav:"tagged_status,omitempty"`  // "environment:production#CERT_UPLOADED"
}

// During entity creation
func computeSearchFields(entity *models.CertificateEntity) {
    var searchTags []string
    for key, value := range entity.Tags {
        searchTags = append(searchTags, fmt.Sprintf("%s:%s", key, value))
    }
    entity.SearchTags = searchTags

    // Combined status + primary tag for common searches
    if env, exists := entity.Tags["environment"]; exists {
        entity.TaggedStatus = fmt.Sprintf("%s:%s#%s", "environment", env, entity.Status)
    }
}

// GSI on search_tags for contains queries
// GSI on tagged_status for environment + status queries
```

## Performance Comparison

| Method | Best For | Query Speed | Storage Cost | Complexity |
|--------|----------|-------------|--------------|------------|
| Current (Scan + Filter) | < 10K items | Slow | Low | Low |
| Tag-based GSI | Common tags | Fast | Medium | Medium |
| Inverted Index | Complex tag queries | Fast | High | High |
| Composite Attributes | Mixed queries | Medium | Low | Medium |

## Migration Strategy

### Phase 1: Add GSI for Top Tags (Immediate)
```bash
# Identify most queried tags from logs
aws logs filter-log-events --log-group-name certificate-monkey \
  --filter-pattern "?environment ?project ?team"

# Add GSIs for top 3-5 most common tags
```

### Phase 2: Hybrid Approach (Recommended)
```go
func (d *DynamoDBStorage) ListCertificateEntities(ctx context.Context, filters models.SearchFilters) ([]models.CertificateEntity, error) {
    // Check if we can use optimized GSI query
    if len(filters.Tags) == 1 {
        for tagKey, tagValue := range filters.Tags {
            if d.hasGSIForTag(tagKey) {
                return d.queryByTagGSI(ctx, tagKey, tagValue, filters)
            }
        }
    }

    // Fall back to current scan-based approach for other cases
    return d.scanWithFilters(ctx, filters)
}
```

### Phase 3: Full Optimization (Future)
- Implement inverted index for complex queries
- Add tag analytics and query optimization
- Consider migrating to dedicated search service (ElasticSearch/OpenSearch)

## Monitoring & Optimization

```go
// Add metrics for tag search performance
type TagSearchMetrics struct {
    SearchMethod     string        // "gsi", "scan", "inverted_index"
    TagsQueried      []string      // tags in the query
    ResultCount      int           // number of results
    QueryTime        time.Duration // time taken
    RCUConsumed      int           // DynamoDB read capacity units
}

func (d *DynamoDBStorage) recordTagSearchMetrics(metrics TagSearchMetrics) {
    d.logger.WithFields(logrus.Fields{
        "search_method": metrics.SearchMethod,
        "tags_queried":  metrics.TagsQueried,
        "result_count":  metrics.ResultCount,
        "query_time_ms": metrics.QueryTime.Milliseconds(),
        "rcu_consumed":  metrics.RCUConsumed,
    }).Info("Tag search performed")
}
```

## Recommended Implementation Order

1. **Immediate (Current works fine for most use cases)**
   - Continue using current implementation
   - Add monitoring for query patterns

2. **Short term (when > 1K certificates)**
   - Add GSI for `environment` tag
   - Add GSI for `project` tag
   - Implement hybrid search logic

3. **Medium term (when > 10K certificates)**
   - Implement inverted index pattern
   - Add search result caching
   - Consider search service integration

4. **Long term (enterprise scale)**
   - Full search service (OpenSearch)
   - Real-time search analytics
   - Advanced query optimization

## Testing Tag Search Performance

```bash
# Test current implementation
scripts/test-tag-search-performance.sh

# Load test with various tag combinations
go test ./internal/storage -run TestTagSearchPerformance -v
```

This guide provides a clear path from the current working implementation to enterprise-scale tag searching capabilities.
