package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Server   ServerConfig
	AWS      AWSConfig
	Security SecurityConfig
}

type ServerConfig struct {
	Port string
	Host string
}

type AWSConfig struct {
	Region        string
	DynamoDBTable string
	KMSKeyID      string
}

type SecurityConfig struct {
	APIKeys []string
}

func Load() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Port: getEnvWithDefault("SERVER_PORT", "8080"),
			Host: getEnvWithDefault("SERVER_HOST", "0.0.0.0"),
		},
		AWS: AWSConfig{
			Region:        getEnvWithDefault("AWS_REGION", "eu-central-1"),
			DynamoDBTable: getEnvWithDefault("DYNAMODB_TABLE", "certificate-monkey-dev"),
			KMSKeyID:      getEnvWithDefault("KMS_KEY_ID", "alias/certificate-monkey-dev"),
		},
		Security: SecurityConfig{
			APIKeys: []string{
				getEnvWithDefault("API_KEY_1", "cm_dev_12345"),  // TODO: remove this default value for production ready version
				getEnvWithDefault("API_KEY_2", "cm_prod_67890"), // TODO: remove this default value for production ready version
			},
		},
	}

	// Validate API keys are not empty
	if cfg.Security.APIKeys[0] == "" {
		return nil, fmt.Errorf("API_KEY_1 is required")
	}
	if cfg.Security.APIKeys[1] == "" {
		return nil, fmt.Errorf("API_KEY_2 is required")
	}

	// Validate KMS key ID is set
	if cfg.AWS.KMSKeyID == "" {
		return nil, fmt.Errorf("KMS_KEY_ID is required")
	}

	return cfg, nil
}

func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
