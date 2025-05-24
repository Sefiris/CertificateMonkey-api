#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
CA_KEY_FILE="test-ca.key"
CA_CERT_FILE="test-ca.crt"
CSR_FILE=""
OUTPUT_CERT=""
VALIDITY_DAYS=365
SERIAL_NUMBER=""

# Function to display usage
usage() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Sign a CSR and generate a certificate for testing Certificate Monkey API"
    echo ""
    echo "Options:"
    echo "  -c, --csr FILE           CSR file to sign (required)"
    echo "  -o, --output FILE        Output certificate file (default: derived from CSR filename)"
    echo "  -d, --days DAYS          Certificate validity in days (default: 365)"
    echo "  -s, --serial NUMBER      Serial number for certificate (default: auto-generated)"
    echo "  --ca-key FILE           CA private key file (default: test-ca.key)"
    echo "  --ca-cert FILE          CA certificate file (default: test-ca.crt)"
    echo "  --setup-ca              Create a new test CA"
    echo "  --example               Show example usage with curl"
    echo "  -h, --help              Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 --setup-ca                          # Create test CA"
    echo "  $0 -c request.csr                      # Sign CSR with test CA"
    echo "  $0 -c request.csr -o certificate.crt   # Sign CSR, save to specific file"
    echo "  $0 --example                           # Show API usage example"
}

# Function to create a test CA
setup_ca() {
    echo -e "${BLUE}🔧 Setting up test Certificate Authority...${NC}"

    if [[ -f "$CA_KEY_FILE" && -f "$CA_CERT_FILE" ]]; then
        echo -e "${YELLOW}⚠️  CA files already exist. Use --ca-key and --ca-cert to specify different files.${NC}"
        read -p "Overwrite existing CA? (y/N): " confirm
        if [[ ! "$confirm" =~ ^[Yy]$ ]]; then
            echo "Aborted."
            exit 1
        fi
    fi

    # Generate CA private key
    echo "Generating CA private key..."
    openssl genrsa -out "$CA_KEY_FILE" 4096

    # Generate CA certificate
    echo "Generating CA certificate..."
    openssl req -new -x509 -key "$CA_KEY_FILE" -out "$CA_CERT_FILE" -days 3650 \
        -subj "/C=US/ST=Test/L=Test/O=Certificate Monkey Test CA/CN=Test CA" \
        -extensions v3_ca -config <(
        echo '[req]'
        echo 'distinguished_name = req'
        echo '[v3_ca]'
        echo 'basicConstraints = CA:TRUE'
        echo 'keyUsage = keyCertSign, cRLSign'
        echo 'subjectKeyIdentifier = hash'
        echo 'authorityKeyIdentifier = keyid:always,issuer:always'
    )

    echo -e "${GREEN}✅ Test CA created successfully!${NC}"
    echo "  CA Key:  $CA_KEY_FILE"
    echo "  CA Cert: $CA_CERT_FILE"
    echo ""
    echo "CA Certificate Details:"
    openssl x509 -in "$CA_CERT_FILE" -text -noout | grep -E "(Subject:|Validity|Not Before|Not After)"
}

# Function to sign a CSR
sign_csr() {
    local csr_file="$1"
    local output_file="$2"

    if [[ ! -f "$csr_file" ]]; then
        echo -e "${RED}❌ CSR file not found: $csr_file${NC}"
        exit 1
    fi

    if [[ ! -f "$CA_KEY_FILE" || ! -f "$CA_CERT_FILE" ]]; then
        echo -e "${RED}❌ CA files not found. Run with --setup-ca first.${NC}"
        exit 1
    fi

    # Generate serial number if not provided
    if [[ -z "$SERIAL_NUMBER" ]]; then
        SERIAL_NUMBER=$(date +%s)
    fi

    echo -e "${BLUE}🔏 Signing CSR...${NC}"
    echo "  CSR file: $csr_file"
    echo "  Output:   $output_file"
    echo "  Validity: $VALIDITY_DAYS days"
    echo "  Serial:   $SERIAL_NUMBER"

    # Display CSR details
    echo ""
    echo "CSR Details:"
    openssl req -in "$csr_file" -text -noout | grep -E "(Subject:|Public Key|Subject Alternative Name)" || true

    # Sign the CSR
    openssl x509 -req -in "$csr_file" \
        -CA "$CA_CERT_FILE" \
        -CAkey "$CA_KEY_FILE" \
        -out "$output_file" \
        -days "$VALIDITY_DAYS" \
        -set_serial "$SERIAL_NUMBER" \
        -extensions v3_req -extfile <(
        echo '[v3_req]'
        echo 'basicConstraints = CA:FALSE'
        echo 'keyUsage = nonRepudiation, digitalSignature, keyEncipherment'
        echo 'extendedKeyUsage = serverAuth, clientAuth'
        echo 'subjectKeyIdentifier = hash'
        echo 'authorityKeyIdentifier = keyid,issuer'

        # Extract SAN from CSR if present
        san_line=$(openssl req -in "$csr_file" -text -noout | grep -A1 "Subject Alternative Name" | tail -1 | sed 's/^ *//')
        if [[ -n "$san_line" ]]; then
            echo "subjectAltName = $san_line"
        fi
    )

    echo ""
    echo -e "${GREEN}✅ Certificate signed successfully!${NC}"
    echo "  Certificate: $output_file"

    # Display certificate details
    echo ""
    echo "Certificate Details:"
    openssl x509 -in "$output_file" -text -noout | grep -E "(Subject:|Issuer:|Validity|Not Before|Not After|Subject Alternative Name)" || true

    # Show fingerprint
    echo ""
    echo "Certificate Fingerprint (SHA-256):"
    openssl x509 -in "$output_file" -fingerprint -sha256 -noout
}

# Function to show example usage
show_example() {
    cat << 'EOF'
🚀 Example Usage with Certificate Monkey API

1. First, create a key and CSR using the API:
```bash
curl -X POST http://localhost:8080/api/v1/keys \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key" \
  -d '{
    "common_name": "example.com",
    "organization": "Test Corp",
    "country": "US",
    "key_type": "RSA2048"
  }'
```

2. Save the CSR from the response to a file (e.g., example.csr)

3. Sign the CSR using this script:
```bash
./scripts/sign-csr.sh --setup-ca                    # First time only
./scripts/sign-csr.sh -c example.csr -o example.crt
```

4. Upload the certificate using the API:
```bash
# Get the certificate content
CERT_CONTENT=$(cat example.crt)

# Upload to Certificate Monkey
curl -X PUT http://localhost:8080/api/v1/keys/{key-id}/certificate \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key" \
  -d "{
    \"certificate\": \"$CERT_CONTENT\"
  }"
```

5. Verify the certificate was uploaded:
```bash
curl -X GET http://localhost:8080/api/v1/keys/{key-id} \
  -H "X-API-Key: your-api-key"
```

📁 Files created:
- test-ca.key / test-ca.crt  # Test CA (keep secure!)
- example.csr               # Certificate Signing Request
- example.crt               # Signed Certificate

🔒 Security Note:
This creates a test CA for development only. Never use in production!
EOF
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -c|--csr)
            CSR_FILE="$2"
            shift 2
            ;;
        -o|--output)
            OUTPUT_CERT="$2"
            shift 2
            ;;
        -d|--days)
            VALIDITY_DAYS="$2"
            shift 2
            ;;
        -s|--serial)
            SERIAL_NUMBER="$2"
            shift 2
            ;;
        --ca-key)
            CA_KEY_FILE="$2"
            shift 2
            ;;
        --ca-cert)
            CA_CERT_FILE="$2"
            shift 2
            ;;
        --setup-ca)
            setup_ca
            exit 0
            ;;
        --example)
            show_example
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

# Main logic
if [[ -z "$CSR_FILE" ]]; then
    echo -e "${RED}❌ CSR file is required. Use -c or --csr to specify.${NC}"
    echo ""
    usage
    exit 1
fi

# Set default output filename if not provided
if [[ -z "$OUTPUT_CERT" ]]; then
    OUTPUT_CERT="${CSR_FILE%.csr}.crt"
fi

# Check if OpenSSL is available
if ! command -v openssl &> /dev/null; then
    echo -e "${RED}❌ OpenSSL is required but not installed.${NC}"
    exit 1
fi

# Sign the CSR
sign_csr "$CSR_FILE" "$OUTPUT_CERT"

echo ""
echo -e "${GREEN}🎉 Ready to upload certificate to Certificate Monkey API!${NC}"
echo ""
echo "Next steps:"
echo "1. Use the certificate file: $OUTPUT_CERT"
echo "2. Upload via PUT /api/v1/keys/{id}/certificate"
echo "3. Run './scripts/sign-csr.sh --example' for full API usage example"
