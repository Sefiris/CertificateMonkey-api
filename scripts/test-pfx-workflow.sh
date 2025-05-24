#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Configuration
API_BASE_URL="http://localhost:8080"
API_KEY="test-api-key-12345"
PFX_PASSWORD="MySecurePFXPassword123!"

echo -e "${BLUE}🚀 Certificate Monkey - Complete PFX Workflow Test${NC}"
echo -e "${BLUE}=================================================${NC}"
echo ""

# Function to check if server is running
check_server() {
    echo -e "${CYAN}📡 Checking if Certificate Monkey server is running...${NC}"
    if ! curl -s "$API_BASE_URL/health" > /dev/null; then
        echo -e "${RED}❌ Server is not running at $API_BASE_URL${NC}"
        echo -e "${YELLOW}💡 Please start the server with: go run cmd/server/main.go${NC}"
        exit 1
    fi
    echo -e "${GREEN}✅ Server is running${NC}"
    echo ""
}

# Function to create a key and CSR
create_key() {
    echo -e "${CYAN}🔑 Step 1: Creating private key and CSR...${NC}"

    local response=$(curl -s -X POST "$API_BASE_URL/api/v1/keys" \
        -H "Content-Type: application/json" \
        -H "X-API-Key: $API_KEY" \
        -d '{
            "common_name": "pfx-test.example.com",
            "organization": "Certificate Monkey Test Corp",
            "organizational_unit": "IT Department",
            "country": "US",
            "state": "California",
            "city": "San Francisco",
            "key_type": "RSA2048",
            "tags": {
                "purpose": "pfx-workflow-test",
                "created_by": "test-script"
            }
        }')

    # Extract key ID using basic string manipulation (works without jq)
    KEY_ID=$(echo "$response" | grep -o '"id":"[^"]*"' | cut -d'"' -f4)

    if [[ -z "$KEY_ID" ]]; then
        echo -e "${RED}❌ Failed to create key. Response:${NC}"
        echo "$response"
        exit 1
    fi

    echo -e "${GREEN}✅ Key created successfully!${NC}"
    echo -e "   Key ID: ${YELLOW}$KEY_ID${NC}"
    echo ""
}

# Function to extract CSR
extract_csr() {
    echo -e "${CYAN}📋 Step 2: Extracting CSR...${NC}"

    if [[ -f "scripts/extract-csr.sh" ]]; then
        ./scripts/extract-csr.sh -k "$KEY_ID" -a "$API_KEY" -o "test-workflow.csr"
        echo -e "${GREEN}✅ CSR extracted to test-workflow.csr${NC}"
    else
        echo -e "${YELLOW}⚠️  extract-csr.sh not found, extracting manually...${NC}"

        local response=$(curl -s -X GET "$API_BASE_URL/api/v1/keys/$KEY_ID" \
            -H "X-API-Key: $API_KEY")

        # Extract CSR (basic approach without jq)
        echo "$response" | grep -o '"csr":"[^"]*"' | cut -d'"' -f4 | sed 's/\\n/\n/g' > test-workflow.csr
        echo -e "${GREEN}✅ CSR extracted manually${NC}"
    fi
    echo ""
}

# Function to sign CSR
sign_csr() {
    echo -e "${CYAN}🔏 Step 3: Signing CSR with test CA...${NC}"

    if [[ -f "scripts/sign-csr.sh" ]]; then
        # Setup CA if it doesn't exist
        if [[ ! -f "test-ca.key" || ! -f "test-ca.crt" ]]; then
            echo -e "${YELLOW}📜 Setting up test CA...${NC}"
            ./scripts/sign-csr.sh --setup-ca
        fi

        # Sign the CSR
        ./scripts/sign-csr.sh -c test-workflow.csr -o test-workflow.crt
        echo -e "${GREEN}✅ Certificate signed successfully!${NC}"
    else
        echo -e "${RED}❌ sign-csr.sh not found${NC}"
        exit 1
    fi
    echo ""
}

# Function to upload certificate
upload_certificate() {
    echo -e "${CYAN}📤 Step 4: Uploading certificate...${NC}"

    if [[ ! -f "test-workflow.crt" ]]; then
        echo -e "${RED}❌ Certificate file not found${NC}"
        exit 1
    fi

    # Read certificate content
    CERT_CONTENT=$(cat test-workflow.crt)

    local response=$(curl -s -X PUT "$API_BASE_URL/api/v1/keys/$KEY_ID/certificate" \
        -H "Content-Type: application/json" \
        -H "X-API-Key: $API_KEY" \
        -d "{\"certificate\": \"$CERT_CONTENT\"}")

    # Check if upload was successful
    if echo "$response" | grep -q '"status":"CERT_UPLOADED"'; then
        echo -e "${GREEN}✅ Certificate uploaded successfully!${NC}"

        # Extract certificate details
        FINGERPRINT=$(echo "$response" | grep -o '"fingerprint":"[^"]*"' | cut -d'"' -f4)
        SERIAL=$(echo "$response" | grep -o '"serial_number":"[^"]*"' | cut -d'"' -f4)

        echo -e "   Status: ${GREEN}CERT_UPLOADED${NC}"
        echo -e "   Fingerprint: ${YELLOW}$FINGERPRINT${NC}"
        echo -e "   Serial: ${YELLOW}$SERIAL${NC}"
    else
        echo -e "${RED}❌ Failed to upload certificate. Response:${NC}"
        echo "$response"
        exit 1
    fi
    echo ""
}

# Function to generate PFX
generate_pfx() {
    echo -e "${CYAN}🔐 Step 5: Generating PFX file...${NC}"

    local response=$(curl -s -X POST "$API_BASE_URL/api/v1/keys/$KEY_ID/pfx" \
        -H "Content-Type: application/json" \
        -H "X-API-Key: $API_KEY" \
        -d "{\"password\": \"$PFX_PASSWORD\"}")

    # Check if PFX generation was successful
    if echo "$response" | grep -q '"pfx_data":"'; then
        echo -e "${GREEN}✅ PFX file generated successfully!${NC}"

        # Extract PFX data and filename
        PFX_DATA=$(echo "$response" | grep -o '"pfx_data":"[^"]*"' | cut -d'"' -f4)
        FILENAME=$(echo "$response" | grep -o '"filename":"[^"]*"' | cut -d'"' -f4)

        # Decode and save PFX file
        echo "$PFX_DATA" | base64 -d > "$FILENAME"

        echo -e "   Filename: ${YELLOW}$FILENAME${NC}"
        echo -e "   Password: ${YELLOW}$PFX_PASSWORD${NC}"
        echo -e "   Size: ${YELLOW}$(ls -lh "$FILENAME" | awk '{print $5}')${NC}"

        # Verify PFX file with OpenSSL (if available)
        if command -v openssl &> /dev/null; then
            echo ""
            echo -e "${CYAN}🔍 Verifying PFX file...${NC}"
            if openssl pkcs12 -in "$FILENAME" -passin "pass:$PFX_PASSWORD" -noout 2>/dev/null; then
                echo -e "${GREEN}✅ PFX file is valid!${NC}"

                # Show PFX contents
                echo -e "${CYAN}📜 PFX Contents:${NC}"
                openssl pkcs12 -in "$FILENAME" -passin "pass:$PFX_PASSWORD" -nokeys -noout -info 2>/dev/null || true
            else
                echo -e "${RED}❌ PFX file validation failed${NC}"
            fi
        fi
    else
        echo -e "${RED}❌ Failed to generate PFX. Response:${NC}"
        echo "$response"
        exit 1
    fi
    echo ""
}

# Function to cleanup
cleanup() {
    echo -e "${CYAN}🧹 Cleaning up test files...${NC}"
    rm -f test-workflow.csr test-workflow.crt
    echo -e "${GREEN}✅ Cleanup complete${NC}"
    echo ""
}

# Function to show summary
show_summary() {
    echo -e "${BLUE}📋 Workflow Summary${NC}"
    echo -e "${BLUE}==================${NC}"
    echo -e "${GREEN}✅ Private key and CSR created${NC}"
    echo -e "${GREEN}✅ CSR extracted from API${NC}"
    echo -e "${GREEN}✅ Certificate signed with test CA${NC}"
    echo -e "${GREEN}✅ Certificate uploaded to API${NC}"
    echo -e "${GREEN}✅ PFX file generated successfully${NC}"
    echo ""
    echo -e "${CYAN}📁 Generated Files:${NC}"
    echo -e "   🔐 ${YELLOW}$(ls pfx-test.example.com-*.pfx 2>/dev/null || echo "Generated PFX file")${NC}"
    echo -e "   🔑 ${YELLOW}test-ca.key${NC} (Test CA private key)"
    echo -e "   📜 ${YELLOW}test-ca.crt${NC} (Test CA certificate)"
    echo ""
    echo -e "${CYAN}🔒 Security Information:${NC}"
    echo -e "   Key ID: ${YELLOW}$KEY_ID${NC}"
    echo -e "   PFX Password: ${YELLOW}$PFX_PASSWORD${NC}"
    echo ""
    echo -e "${GREEN}🎉 PFX workflow test completed successfully!${NC}"
    echo -e "${YELLOW}💡 You can now use the PFX file in applications that require PKCS#12 format${NC}"
}

# Main execution
main() {
    check_server
    create_key
    extract_csr
    sign_csr
    upload_certificate
    generate_pfx
    cleanup
    show_summary
}

# Handle script interruption
trap 'echo -e "\n${RED}❌ Script interrupted${NC}"; cleanup; exit 1' INT TERM

# Run main function
main
