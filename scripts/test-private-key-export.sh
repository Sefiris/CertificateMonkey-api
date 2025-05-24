#!/bin/bash

# Certificate Monkey - Private Key Export Test Script

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

# Configuration
API_BASE_URL="${API_BASE_URL:-http://localhost:8080}"
API_KEY="${API_KEY:-demo-api-key-12345}"

echo -e "${BLUE}🔐 Certificate Monkey - Private Key Export Test${NC}"
echo -e "${BLUE}===============================================${NC}"
echo ""

# Function to check if server is running
check_server() {
    echo -e "${CYAN}📡 Checking if Certificate Monkey server is running...${NC}"
    if ! curl -s "$API_BASE_URL/health" > /dev/null; then
        echo -e "${RED}❌ Server is not running at $API_BASE_URL${NC}"
        echo -e "${YELLOW}💡 Please start the server with: make swagger-serve${NC}"
        exit 1
    fi
    echo -e "${GREEN}✅ Server is running${NC}"
    echo ""
}

# Function to create a test certificate
create_test_certificate() {
    echo -e "${CYAN}🔧 Creating test certificate...${NC}"

    local response=$(curl -s -X POST "$API_BASE_URL/api/v1/keys" \
        -H "Content-Type: application/json" \
        -H "X-API-Key: $API_KEY" \
        -d '{
            "common_name": "private-key-test.example.com",
            "key_type": "RSA2048",
            "organization": "Test Organization",
            "country": "US",
            "tags": {
                "purpose": "private-key-export-test",
                "created_by": "test-script"
            }
        }')

    if [[ $? -ne 0 ]]; then
        echo -e "${RED}❌ Failed to create certificate${NC}"
        exit 1
    fi

    local certificate_id=$(echo "$response" | jq -r '.id // empty')
    if [[ -z "$certificate_id" || "$certificate_id" == "null" ]]; then
        echo -e "${RED}❌ Failed to extract certificate ID from response${NC}"
        echo "Response: $response"
        exit 1
    fi

    echo -e "${GREEN}✅ Certificate created successfully${NC}"
    echo -e "   ID: ${YELLOW}$certificate_id${NC}"
    echo -e "   Common Name: private-key-test.example.com"
    echo ""

    echo "$certificate_id"
}

# Function to export private key
export_private_key() {
    local certificate_id="$1"

    echo -e "${CYAN}🔑 Exporting private key...${NC}"
    echo -e "${YELLOW}⚠️  WARNING: This is a sensitive operation that exposes cryptographic material!${NC}"
    echo ""

    local response=$(curl -s -X GET "$API_BASE_URL/api/v1/keys/$certificate_id/private-key" \
        -H "Content-Type: application/json" \
        -H "X-API-Key: $API_KEY")

    if [[ $? -ne 0 ]]; then
        echo -e "${RED}❌ Failed to export private key${NC}"
        exit 1
    fi

    # Check if response contains error
    local error_msg=$(echo "$response" | jq -r '.error // empty')
    if [[ -n "$error_msg" && "$error_msg" != "null" ]]; then
        echo -e "${RED}❌ API Error: $error_msg${NC}"
        local message=$(echo "$response" | jq -r '.message // empty')
        if [[ -n "$message" && "$message" != "null" ]]; then
            echo -e "   Details: $message"
        fi
        exit 1
    fi

    # Extract response fields
    local exported_id=$(echo "$response" | jq -r '.id // empty')
    local key_type=$(echo "$response" | jq -r '.key_type // empty')
    local common_name=$(echo "$response" | jq -r '.common_name // empty')
    local exported_at=$(echo "$response" | jq -r '.exported_at // empty')
    local private_key=$(echo "$response" | jq -r '.private_key // empty')

    if [[ -z "$private_key" || "$private_key" == "null" ]]; then
        echo -e "${RED}❌ No private key in response${NC}"
        echo "Response: $response"
        exit 1
    fi

    echo -e "${GREEN}✅ Private key exported successfully${NC}"
    echo -e "   Certificate ID: ${YELLOW}$exported_id${NC}"
    echo -e "   Key Type: $key_type"
    echo -e "   Common Name: $common_name"
    echo -e "   Exported At: $exported_at"
    echo ""

    echo -e "${CYAN}🔍 Private Key Details:${NC}"
    echo -e "   Length: ${YELLOW}$(echo "$private_key" | wc -c) characters${NC}"
    echo -e "   Format: PEM"
    echo -e "   Header: $(echo "$private_key" | head -1)"
    echo -e "   Footer: $(echo "$private_key" | tail -1)"
    echo ""

    # Save to file for verification
    local key_file="/tmp/exported-private-key-$certificate_id.pem"
    echo "$private_key" > "$key_file"
    echo -e "${CYAN}💾 Private key saved to: ${YELLOW}$key_file${NC}"
    echo ""

    # Verify the key format using OpenSSL (if available)
    if command -v openssl &> /dev/null; then
        echo -e "${CYAN}🔬 Verifying private key with OpenSSL...${NC}"
        if openssl rsa -in "$key_file" -text -noout > /dev/null 2>&1; then
            echo -e "${GREEN}✅ Private key is valid RSA format${NC}"
            local key_bits=$(openssl rsa -in "$key_file" -text -noout 2>/dev/null | grep "Private-Key:" | sed 's/.*(\([0-9]*\) bit).*/\1/')
            echo -e "   Key Size: ${YELLOW}$key_bits bits${NC}"
        elif openssl ec -in "$key_file" -text -noout > /dev/null 2>&1; then
            echo -e "${GREEN}✅ Private key is valid EC format${NC}"
            local curve=$(openssl ec -in "$key_file" -text -noout 2>/dev/null | grep "NIST CURVE" | sed 's/.*NIST CURVE: \(.*\)/\1/')
            echo -e "   Curve: ${YELLOW}$curve${NC}"
        else
            echo -e "${YELLOW}⚠️  Could not verify private key format with OpenSSL${NC}"
        fi
        echo ""
    fi

    return 0
}

# Function to test error cases
test_error_cases() {
    echo -e "${CYAN}🧪 Testing error cases...${NC}"

    # Test with non-existent certificate ID
    echo -e "   Testing non-existent certificate ID..."
    local error_response=$(curl -s -X GET "$API_BASE_URL/api/v1/keys/non-existent-id/private-key" \
        -H "Content-Type: application/json" \
        -H "X-API-Key: $API_KEY")

    local error_msg=$(echo "$error_response" | jq -r '.error // empty')
    if [[ "$error_msg" == "Not Found" ]]; then
        echo -e "${GREEN}   ✅ Correctly returned 'Not Found' for non-existent ID${NC}"
    else
        echo -e "${YELLOW}   ⚠️  Unexpected error response: $error_msg${NC}"
    fi

    # Test without authentication
    echo -e "   Testing without authentication..."
    local auth_response=$(curl -s -X GET "$API_BASE_URL/api/v1/keys/some-id/private-key" \
        -H "Content-Type: application/json")

    local auth_error=$(echo "$auth_response" | jq -r '.error // empty')
    if [[ "$auth_error" == "Unauthorized" ]]; then
        echo -e "${GREEN}   ✅ Correctly returned 'Unauthorized' without API key${NC}"
    else
        echo -e "${YELLOW}   ⚠️  Unexpected auth response: $auth_error${NC}"
    fi

    echo ""
}

# Function to cleanup
cleanup() {
    echo -e "${CYAN}🧹 Cleaning up temporary files...${NC}"
    rm -f /tmp/exported-private-key-*.pem
    echo -e "${GREEN}✅ Cleanup complete${NC}"
}

# Main execution
main() {
    check_server

    echo -e "${YELLOW}🔒 SECURITY NOTICE${NC}"
    echo -e "${YELLOW}=================${NC}"
    echo -e "This script will demonstrate the private key export functionality."
    echo -e "Private keys are sensitive cryptographic material and should be"
    echo -e "handled with extreme care in production environments."
    echo ""
    echo -e "Press ${CYAN}Enter${NC} to continue or ${CYAN}Ctrl+C${NC} to cancel..."
    read -r
    echo ""

    # Create test certificate and export its private key
    certificate_id=$(create_test_certificate)
    export_private_key "$certificate_id"

    # Test error cases
    test_error_cases

    # Cleanup
    cleanup

    echo -e "${GREEN}🎉 Private key export test completed successfully!${NC}"
    echo ""
    echo -e "${CYAN}📖 Usage Examples:${NC}"
    echo -e "   List all certificates:"
    echo -e "   ${YELLOW}curl -H 'X-API-Key: $API_KEY' $API_BASE_URL/api/v1/keys${NC}"
    echo ""
    echo -e "   Export private key:"
    echo -e "   ${YELLOW}curl -H 'X-API-Key: $API_KEY' $API_BASE_URL/api/v1/keys/\$CERT_ID/private-key${NC}"
    echo ""
    echo -e "${CYAN}📚 Documentation:${NC}"
    echo -e "   Swagger UI: ${BLUE}$API_BASE_URL/swagger/index.html${NC}"
    echo ""
}

# Run main function
main "$@"
