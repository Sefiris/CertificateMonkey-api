#!/bin/bash

# Certificate Monkey - Test All X.509 Certificate Fields
# This script demonstrates how to use all available certificate fields

set -e

API_URL="http://localhost:8080"
API_KEY="cm_dev_12345"

echo "🐒 Certificate Monkey - Testing All X.509 Fields"
echo "================================================="

# Test 1: RSA Certificate with All Fields
echo ""
echo "📝 Test 1: Creating RSA certificate with all X.509 fields..."

RSA_RESPONSE=$(curl -s -X POST "$API_URL/api/v1/keys" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: $API_KEY" \
  -d '{
    "common_name": "secure.example.com",
    "subject_alternative_names": [
      "www.secure.example.com",
      "api.secure.example.com",
      "mail.secure.example.com",
      "192.168.1.100",
      "10.0.0.50"
    ],
    "organization": "ACME Corporation Ltd",
    "organizational_unit": "Information Technology Department",
    "country": "US",
    "state": "California",
    "city": "San Francisco",
    "email_address": "ssl-admin@secure.example.com",
    "key_type": "RSA2048",
    "tags": {
      "environment": "production",
      "project": "web-platform",
      "cost-center": "IT-001",
      "owner": "ssl-team@example.com",
      "expiry-notification": "true"
    }
  }')

RSA_ID=$(echo "$RSA_RESPONSE" | jq -r '.id')
echo "✅ RSA Certificate created with ID: $RSA_ID"

# Test 2: ECDSA Certificate with Minimal Fields
echo ""
echo "📝 Test 2: Creating ECDSA certificate with minimal fields..."

ECDSA_RESPONSE=$(curl -s -X POST "$API_URL/api/v1/keys" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: $API_KEY" \
  -d '{
    "common_name": "minimal.example.com",
    "key_type": "ECDSA-P256"
  }')

ECDSA_ID=$(echo "$ECDSA_RESPONSE" | jq -r '.id')
echo "✅ ECDSA Certificate created with ID: $ECDSA_ID"

# Test 3: Email-only Certificate (for code signing, etc.)
echo ""
echo "📝 Test 3: Creating certificate with email address only..."

EMAIL_RESPONSE=$(curl -s -X POST "$API_URL/api/v1/keys" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: $API_KEY" \
  -d '{
    "common_name": "John Doe",
    "email_address": "john.doe@example.com",
    "organization": "ACME Corporation",
    "country": "US",
    "key_type": "RSA4096",
    "tags": {
      "type": "code-signing",
      "user": "john.doe"
    }
  }')

EMAIL_ID=$(echo "$EMAIL_RESPONSE" | jq -r '.id')
echo "✅ Email Certificate created with ID: $EMAIL_ID"

# Display CSRs for verification
echo ""
echo "🔍 Generated CSRs:"
echo "=================="

echo ""
echo "RSA Certificate CSR:"
echo "$RSA_RESPONSE" | jq -r '.csr'

echo ""
echo "ECDSA Certificate CSR:"
echo "$ECDSA_RESPONSE" | jq -r '.csr'

echo ""
echo "Email Certificate CSR:"
echo "$EMAIL_RESPONSE" | jq -r '.csr'

# Verify CSR contents with OpenSSL (if available)
if command -v openssl &> /dev/null; then
    echo ""
    echo "🔍 CSR Verification with OpenSSL:"
    echo "=================================="

    echo ""
    echo "RSA Certificate Subject:"
    echo "$RSA_RESPONSE" | jq -r '.csr' | openssl req -noout -subject -text | head -10

    echo ""
    echo "ECDSA Certificate Subject:"
    echo "$ECDSA_RESPONSE" | jq -r '.csr' | openssl req -noout -subject -text | head -10

    echo ""
    echo "Email Certificate Subject:"
    echo "$EMAIL_RESPONSE" | jq -r '.csr' | openssl req -noout -subject -text | head -10
else
    echo ""
    echo "⚠️  Install OpenSSL to verify CSR contents"
fi

echo ""
echo "🎉 All tests completed successfully!"
echo "📋 Created certificates:"
echo "   - RSA (all fields): $RSA_ID"
echo "   - ECDSA (minimal):  $ECDSA_ID"
echo "   - Email certificate: $EMAIL_ID"
echo ""
echo "💡 You can now upload certificates to these CSRs using:"
echo "   curl -X PUT $API_URL/api/v1/keys/{id}/certificate -H \"X-API-Key: $API_KEY\" -d '{\"certificate\": \"...\"}'"
