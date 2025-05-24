#!/bin/bash

# Certificate Monkey - Tag Storage Debugging Script
# This script helps verify that tags are stored correctly in DynamoDB

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

API_BASE_URL="${API_BASE_URL:-http://localhost:8080}"
API_KEY="${API_KEY:-test-api-key-12345}"
TABLE_NAME="${DYNAMODB_TABLE:-certificate-monkey}"
AWS_REGION="${AWS_REGION:-us-east-1}"

echo -e "${BLUE}🔍 Certificate Monkey - Tag Storage Debugging${NC}"
echo -e "${BLUE}=============================================${NC}"
echo ""

# Function to check API connectivity
check_api() {
    echo -e "${CYAN}📡 Checking API connectivity...${NC}"
    if curl -s "$API_BASE_URL/health" > /dev/null; then
        echo -e "${GREEN}✅ API is accessible${NC}"
    else
        echo -e "${RED}❌ API is not accessible at $API_BASE_URL${NC}"
        echo -e "${YELLOW}💡 Make sure the server is running${NC}"
        return 1
    fi
    echo ""
}

# Function to create a test certificate with tags
create_test_certificate() {
    echo -e "${CYAN}🏗️  Creating test certificate with tags...${NC}"

    response=$(curl -s -X POST "$API_BASE_URL/api/v1/keys" \
        -H "Content-Type: application/json" \
        -H "X-API-Key: $API_KEY" \
        -d '{
            "common_name": "tag-debug.example.com",
            "organization": "Debug Corp",
            "key_type": "RSA2048",
            "tags": {
                "environment": "debug",
                "purpose": "tag-storage-test",
                "created_by": "debug-script"
            }
        }')

    # Extract ID (basic parsing without jq)
    CERT_ID=$(echo "$response" | grep -o '"id":"[^"]*"' | cut -d'"' -f4)

    if [[ -n "$CERT_ID" ]]; then
        echo -e "${GREEN}✅ Test certificate created with ID: $CERT_ID${NC}"
        echo -e "   Tags: environment=debug, purpose=tag-storage-test${NC}"
    else
        echo -e "${RED}❌ Failed to create certificate${NC}"
        echo "Response: $response"
        return 1
    fi
    echo ""
}

# Function to verify via API
verify_via_api() {
    echo -e "${CYAN}🔎 Verifying tags via API...${NC}"

    # Get the certificate back via API
    api_response=$(curl -s "$API_BASE_URL/api/v1/keys/$CERT_ID" \
        -H "X-API-Key: $API_KEY")

    echo -e "${YELLOW}API Response (what your application sees):${NC}"
    echo "$api_response" | grep -A 10 '"tags"' || echo "No tags found in API response"
    echo ""

    # Test tag search
    echo -e "${YELLOW}Testing tag search:${NC}"
    search_response=$(curl -s "$API_BASE_URL/api/v1/keys?environment=debug" \
        -H "X-API-Key: $API_KEY")

    search_count=$(echo "$search_response" | grep -o '"id"' | wc -l)
    echo -e "Search for environment=debug: ${GREEN}$search_count results${NC}"

    if [[ $search_count -gt 0 ]]; then
        echo -e "${GREEN}✅ Tag search is working${NC}"
    else
        echo -e "${RED}❌ Tag search returned no results - possible storage issue${NC}"
    fi
    echo ""
}

# Function to check DynamoDB directly
check_dynamodb_storage() {
    echo -e "${CYAN}🗄️  Checking DynamoDB storage directly...${NC}"

    if ! command -v aws &> /dev/null; then
        echo -e "${YELLOW}⚠️  AWS CLI not found, skipping DynamoDB direct check${NC}"
        echo ""
        return
    fi

    # Try to get the item directly from DynamoDB
    echo -e "${YELLOW}Fetching from DynamoDB table: $TABLE_NAME${NC}"

    ddb_response=$(aws dynamodb get-item \
        --table-name "$TABLE_NAME" \
        --key "{\"id\": {\"S\": \"$CERT_ID\"}}" \
        --region "$AWS_REGION" 2>/dev/null || echo "DynamoDB_ACCESS_FAILED")

    if [[ "$ddb_response" == "DynamoDB_ACCESS_FAILED" ]]; then
        echo -e "${YELLOW}⚠️  Cannot access DynamoDB directly (permissions/credentials)${NC}"
        echo -e "${CYAN}💡 This is normal in many environments - using API verification only${NC}"
    else
        echo -e "${YELLOW}Raw DynamoDB storage format:${NC}"
        echo "$ddb_response"

        # Check if tags are stored correctly
        if echo "$ddb_response" | grep -q '"tags".*"M"'; then
            echo -e "${GREEN}✅ Tags are stored correctly as a nested Map${NC}"
        elif echo "$ddb_response" | grep -q '"environment".*"S"' && ! echo "$ddb_response" | grep -q '"tags"'; then
            echo -e "${RED}❌ ISSUE DETECTED: Tags appear to be flattened to top level${NC}"
            echo -e "${RED}   This will break tag searching functionality${NC}"
        else
            echo -e "${YELLOW}🤔 Unclear tag storage format - manual inspection needed${NC}"
        fi
    fi
    echo ""
}

# Function to show expected vs actual format
show_format_comparison() {
    echo -e "${CYAN}📋 Expected vs Actual Storage Format${NC}"
    echo ""

    echo -e "${GREEN}✅ CORRECT DynamoDB format should be:${NC}"
    cat << 'EOF'
{
  "Item": {
    "id": {"S": "123e4567-..."},
    "common_name": {"S": "tag-debug.example.com"},
    "tags": {
      "M": {
        "environment": {"S": "debug"},
        "purpose": {"S": "tag-storage-test"}
      }
    }
  }
}
EOF
    echo ""

    echo -e "${RED}❌ INCORRECT format (tags flattened):${NC}"
    cat << 'EOF'
{
  "Item": {
    "id": {"S": "123e4567-..."},
    "common_name": {"S": "tag-debug.example.com"},
    "environment": {"S": "debug"},
    "purpose": {"S": "tag-storage-test"}
  }
}
EOF
    echo ""
}

# Function to cleanup
cleanup() {
    if [[ -n "$CERT_ID" ]]; then
        echo -e "${CYAN}🧹 Cleaning up test certificate...${NC}"
        curl -s -X DELETE "$API_BASE_URL/api/v1/keys/$CERT_ID" \
            -H "X-API-Key: $API_KEY" > /dev/null || true
        echo -e "${GREEN}✅ Cleanup complete${NC}"
    fi
}

# Function to provide next steps
show_next_steps() {
    echo -e "${CYAN}🔧 Next Steps Based on Results:${NC}"
    echo ""

    echo -e "${YELLOW}If tags are stored correctly:${NC}"
    echo -e "  • Your implementation is working as expected"
    echo -e "  • The {\"S\": \"value\"} format is normal DynamoDB JSON representation"
    echo ""

    echo -e "${YELLOW}If tags are flattened to top level:${NC}"
    echo -e "  • Check your handler code for tag processing"
    echo -e "  • Verify the CertificateEntity model has correct dynamodbav tags"
    echo -e "  • Test with a simple struct marshaling to isolate the issue"
    echo ""

    echo -e "${YELLOW}If tag search isn't working:${NC}"
    echo -e "  • Check filter expression syntax in ListCertificateEntities()"
    echo -e "  • Verify attribute names are mapped correctly"
    echo -e "  • Test with simpler tag combinations"
    echo ""
}

# Main execution
main() {
    check_api || exit 1
    create_test_certificate || exit 1
    verify_via_api
    check_dynamodb_storage
    show_format_comparison
    show_next_steps

    echo -e "${GREEN}🎉 Tag storage debugging complete!${NC}"
    echo -e "${CYAN}📖 For more details, see: docs/TAG_SEARCH_OPTIMIZATION.md${NC}"
}

# Cleanup on exit
trap cleanup EXIT

# Run main
main
