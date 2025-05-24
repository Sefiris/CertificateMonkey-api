#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
API_BASE_URL="http://localhost:8080"
API_KEY=""
KEY_ID=""
OUTPUT_FILE=""

# Function to display usage
usage() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Extract CSR from Certificate Monkey API response and save to file"
    echo ""
    echo "Options:"
    echo "  -k, --key-id ID          Key ID to fetch CSR for (required)"
    echo "  -a, --api-key KEY        API key for authentication (required)"
    echo "  -o, --output FILE        Output CSR file (default: {key-id}.csr)"
    echo "  -u, --url URL           API base URL (default: http://localhost:8080)"
    echo "  -h, --help              Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 -k abc123 -a your-api-key"
    echo "  $0 -k abc123 -a your-api-key -o my-request.csr"
    echo "  $0 -k abc123 -a your-api-key -u https://api.example.com"
}

# Function to extract CSR
extract_csr() {
    local key_id="$1"
    local api_key="$2"
    local output_file="$3"

    echo -e "${BLUE}🔍 Fetching key details from API...${NC}"
    echo "  Key ID: $key_id"
    echo "  API URL: $API_BASE_URL/api/v1/keys/$key_id"

    # Make API request
    local response
    local http_code

    response=$(curl -s -w "HTTPSTATUS:%{http_code}" \
        -X GET "$API_BASE_URL/api/v1/keys/$key_id" \
        -H "Content-Type: application/json" \
        -H "X-API-Key: $api_key")

    # Extract HTTP status code
    http_code=$(echo "$response" | grep -o "HTTPSTATUS:[0-9]*" | cut -d: -f2)

    # Extract response body
    local body=$(echo "$response" | sed -E 's/HTTPSTATUS:[0-9]*$//')

    if [[ "$http_code" != "200" ]]; then
        echo -e "${RED}❌ API request failed with status code: $http_code${NC}"
        echo "Response: $body"
        exit 1
    fi

    echo -e "${GREEN}✅ Successfully fetched key details${NC}"

    # Extract CSR using jq if available, otherwise use basic text processing
    local csr_content
    if command -v jq &> /dev/null; then
        csr_content=$(echo "$body" | jq -r '.certificate_signing_request // empty')
    else
        # Fallback: extract CSR using grep/sed (less reliable but works without jq)
        csr_content=$(echo "$body" | grep -o '"certificate_signing_request":"[^"]*"' | cut -d'"' -f4 | sed 's/\\n/\n/g')
    fi

    if [[ -z "$csr_content" || "$csr_content" == "null" ]]; then
        echo -e "${RED}❌ No CSR found in API response${NC}"
        echo "Response body:"
        echo "$body" | head -20
        exit 1
    fi

    # Validate CSR format
    if [[ ! "$csr_content" =~ -----BEGIN.*(CERTIFICATE REQUEST|NEW CERTIFICATE REQUEST).* ]]; then
        echo -e "${YELLOW}⚠️  CSR content doesn't appear to be in PEM format${NC}"
        echo "First few lines:"
        echo "$csr_content" | head -3
    fi

    # Save CSR to file
    echo "$csr_content" > "$output_file"

    echo -e "${GREEN}✅ CSR extracted and saved to: $output_file${NC}"

    # Display CSR details
    if command -v openssl &> /dev/null; then
        echo ""
        echo "CSR Details:"
        openssl req -in "$output_file" -text -noout | grep -E "(Subject:|Public Key|Subject Alternative Name)" 2>/dev/null || echo "Could not parse CSR details"
    fi

    # Show file size
    echo ""
    echo "File info:"
    ls -lh "$output_file"
}

# Function to show complete workflow example
show_workflow() {
    cat << 'EOF'
🚀 Complete Workflow: Create Key → Extract CSR → Sign → Upload Certificate

1. Create a new key via API:
```bash
RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/keys \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key" \
  -d '{
    "common_name": "example.com",
    "organization": "Test Corp",
    "country": "US",
    "key_type": "RSA2048"
  }')

# Extract the key ID
KEY_ID=$(echo "$RESPONSE" | jq -r '.id')
echo "Created key with ID: $KEY_ID"
```

2. Extract CSR using this script:
```bash
./scripts/extract-csr.sh -k "$KEY_ID" -a your-api-key
```

3. Sign the CSR:
```bash
./scripts/sign-csr.sh --setup-ca              # First time only
./scripts/sign-csr.sh -c "${KEY_ID}.csr"      # Creates ${KEY_ID}.crt
```

4. Upload the certificate:
```bash
CERT_CONTENT=$(cat "${KEY_ID}.crt")
curl -X PUT "http://localhost:8080/api/v1/keys/$KEY_ID/certificate" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key" \
  -d "{\"certificate\": \"$CERT_CONTENT\"}"
```

5. Verify upload:
```bash
curl -X GET "http://localhost:8080/api/v1/keys/$KEY_ID" \
  -H "X-API-Key: your-api-key" | jq .
```

💡 Pro tip: Use this one-liner for the complete flow:
```bash
KEY_ID="your-key-id"
API_KEY="your-api-key"
./scripts/extract-csr.sh -k "$KEY_ID" -a "$API_KEY" && \
./scripts/sign-csr.sh -c "${KEY_ID}.csr" && \
curl -X PUT "http://localhost:8080/api/v1/keys/$KEY_ID/certificate" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: $API_KEY" \
  -d "{\"certificate\": \"$(cat "${KEY_ID}.crt")\"}"
```
EOF
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -k|--key-id)
            KEY_ID="$2"
            shift 2
            ;;
        -a|--api-key)
            API_KEY="$2"
            shift 2
            ;;
        -o|--output)
            OUTPUT_FILE="$2"
            shift 2
            ;;
        -u|--url)
            API_BASE_URL="$2"
            shift 2
            ;;
        --workflow)
            show_workflow
            exit 0
            ;;
        -h|--help)
            usage
            exit 0
            ;;
        *)
            echo -e "${RED}Unknown option: $1${NC}"
            usage
            exit 1
            ;;
    esac
done

# Validate required parameters
if [[ -z "$KEY_ID" ]]; then
    echo -e "${RED}❌ Key ID is required. Use -k or --key-id to specify.${NC}"
    echo ""
    usage
    exit 1
fi

if [[ -z "$API_KEY" ]]; then
    echo -e "${RED}❌ API key is required. Use -a or --api-key to specify.${NC}"
    echo ""
    usage
    exit 1
fi

# Set default output filename if not provided
if [[ -z "$OUTPUT_FILE" ]]; then
    OUTPUT_FILE="${KEY_ID}.csr"
fi

# Check if curl is available
if ! command -v curl &> /dev/null; then
    echo -e "${RED}❌ curl is required but not installed.${NC}"
    exit 1
fi

# Extract the CSR
extract_csr "$KEY_ID" "$API_KEY" "$OUTPUT_FILE"

echo ""
echo -e "${GREEN}🎉 CSR ready for signing!${NC}"
echo ""
echo "Next steps:"
echo "1. Sign the CSR: ./scripts/sign-csr.sh -c $OUTPUT_FILE"
echo "2. Upload certificate: PUT /api/v1/keys/$KEY_ID/certificate"
echo "3. Run './scripts/extract-csr.sh --workflow' for complete example"
