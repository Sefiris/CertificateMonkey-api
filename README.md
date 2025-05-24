# Certificate Monkey 🐒

A secure, scalable API for managing private keys, certificate signing requests (CSRs), and certificates with AWS DynamoDB storage and KMS encryption.

## Disclaimer

**This is my first public Go project and none of this should be ran in Production until it has been properly peer reviewed for security vulnerabilities.**
**It is currently my way of gaining Go development experience**

## Features

- **Private Key Generation**: Support for RSA (2048/4096) and ECDSA (P-256/P-384) key types
- **CSR Creation**: Automatic generation of certificate signing requests
- **Certificate Management**: Upload and validate certificates against CSRs
- **PFX Generation**: Create password-protected PKCS#12 files for legacy application compatibility
- **Secure Storage**: Private keys encrypted with AWS KMS and stored in DynamoDB
- **Search & Filtering**: Query certificates by date, tags, status, and key type
- **RESTful API**: Clean, well-documented endpoints
- **Docker Support**: Production-ready containerization

## Quick Start

### Prerequisites

- Go 1.22+
- AWS account with DynamoDB and KMS access
- Docker (for containerized deployment)

### Local Development

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd CertificateMonkey
   ```

2. **Install dependencies**
   ```bash
   go mod tidy
   ```

3. **Set environment variables**
   ```bash
   export AWS_REGION=us-east-1
   export DYNAMODB_TABLE=certificate-monkey
   export KMS_KEY_ID=alias/certificate-monkey
   export API_KEY_1=your_api_key_here
   export SERVER_PORT=8080
   ```

4. **Run the application**
   ```bash
   go run cmd/server/main.go
   ```

### Docker Deployment

1. **Build the image**
   ```bash
   docker build -t certificate-monkey .
   ```

2. **Run the container**
   ```bash
   docker run -p 8080:8080 \
     -e AWS_REGION=us-east-1 \
     -e DYNAMODB_TABLE=certificate-monkey \
     -e KMS_KEY_ID=alias/certificate-monkey \
     -e API_KEY_1=your_api_key_here \
     certificate-monkey
   ```

## API Documentation

### Authentication

All API endpoints (except `/health`) require authentication via API key:

```bash
# Using X-API-Key header
curl -H "X-API-Key: your_api_key_here" http://localhost:8080/api/v1/keys

# Using Authorization header
curl -H "Authorization: Bearer your_api_key_here" http://localhost:8080/api/v1/keys
```

### Endpoints

#### Health Check
```
GET /health
```

#### Build Information
```
GET /build-info
```

Returns version and build information:
```json
{
  "service": "certificate-monkey",
  "version": "0.1.0",
  "build_time": "2025-05-24_21:16:57_UTC",
  "git_commit": "b739e97",
  "go_version": "go1.24.3",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

#### Create Private Key and CSR

Creates a new private key and certificate signing request (CSR).

**Endpoint**: `POST /api/v1/keys`

**Request Body**:
```json
{
  "common_name": "example.com",
  "subject_alternative_names": ["www.example.com", "api.example.com"],
  "organization": "ACME Corporation",
  "organizational_unit": "IT Department",
  "country": "US",
  "state": "California",
  "city": "San Francisco",
  "email_address": "admin@example.com",
  "key_type": "RSA2048",
  "tags": {
    "environment": "production",
    "project": "web-server"
  }
}
```

**X.509 Certificate Fields**:
- `common_name` (required): CN - Common Name, typically the primary domain name
- `subject_alternative_names` (optional): SAN - Alternative domain names or IP addresses
- `organization` (optional): O - Organization name
- `organizational_unit` (optional): OU - Department or division within the organization
- `country` (optional): C - Two-letter country code (e.g., "US", "CA", "GB")
- `state` (optional): ST - State or province name
- `city` (optional): L - City or locality name
- `email_address` (optional): Email address associated with the certificate
- `key_type` (required): Cryptographic algorithm and key size
- `tags` (optional): Custom metadata for organization and searching

**Supported Key Types**:
- `RSA2048`: RSA 2048-bit key
- `RSA4096`: RSA 4096-bit key
- `ECDSA-P256`: Elliptic Curve P-256 key
- `ECDSA-P384`: Elliptic Curve P-384 key

**Example - Simple Certificate**:
```bash
curl -X POST http://localhost:8080/api/v1/keys \
  -H "Content-Type: application/json" \
  -H "X-API-Key: cm_dev_12345" \
  -d '{
    "common_name": "example.com",
    "key_type": "RSA2048"
  }'
```

**Example - Full Certificate with All Fields**:
```bash
curl -X POST http://localhost:8080/api/v1/keys \
  -H "Content-Type: application/json" \
  -H "X-API-Key: cm_dev_12345" \
  -d '{
    "common_name": "secure.example.com",
    "subject_alternative_names": [
      "www.secure.example.com",
      "api.secure.example.com",
      "192.168.1.100"
    ],
    "organization": "ACME Corporation Ltd",
    "organizational_unit": "Information Technology",
    "country": "US",
    "state": "California",
    "city": "San Francisco",
    "email_address": "ssl-admin@example.com",
    "key_type": "ECDSA-P256",
    "tags": {
      "environment": "production",
      "project": "api-gateway",
      "cost-center": "IT-001",
      "expiry-notification": "ssl-team@example.com"
    }
  }'
```

#### Upload Certificate
```
PUT /api/v1/keys/{id}/certificate
```

**Request Body:**
```json
{
  "certificate": "-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----"
}
```

**Response:**
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "status": "CERT_UPLOADED",
  "valid_from": "2024-01-01T10:00:00Z",
  "valid_to": "2025-01-01T10:00:00Z",
  "serial_number": "123456789",
  "fingerprint": "AA:BB:CC:DD:EE:FF:00:11:22:33:44:55:66:77:88:99:AA:BB:CC:DD",
  "updated_at": "2024-01-01T10:05:00Z"
}
```

#### Generate PFX File
```
POST /api/v1/keys/{id}/pfx
```

**Request Body:**
```json
{
  "password": "your_secure_password"
}
```

**Response:**
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "pfx_data": "base64_encoded_pfx_data",
  "filename": "example.com-123e4567.pfx"
}
```

#### Get Certificate Details
```
GET /api/v1/keys/{id}
```

#### Export Private Key (SENSITIVE)
```
GET /api/v1/keys/{id}/private-key
```

**⚠️ Security Warning**: This endpoint exposes sensitive cryptographic material. Use with extreme caution and ensure proper access controls.

**Response:**
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "private_key": "-----BEGIN PRIVATE KEY-----\nMIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQC7VJTUt9Us8cKB...\n-----END PRIVATE KEY-----",
  "key_type": "RSA2048",
  "common_name": "example.com",
  "exported_at": "2024-01-15T10:30:00Z"
}
```

**Security Features:**
- Comprehensive audit logging with client IP and User-Agent
- Authentication required (API key or Bearer token)
- Structured response with metadata
- RFC3339 timestamp for export tracking

#### List and Search Certificates
```
GET /api/v1/keys?status=CERT_UPLOADED&key_type=RSA2048&environment=production
```

**Query Parameters:**
- `status`: Filter by certificate status
- `key_type`: Filter by key type
- `date_from`: Filter by creation date (RFC3339 format)
- `date_to`: Filter by creation date (RFC3339 format)
- `page`: Page number for pagination
- `page_size`: Number of results per page (max 100)
- Any tag key: Filter by tag value (e.g., `environment=production`)

## API Documentation

### Swagger UI

Certificate Monkey includes comprehensive API documentation with an interactive Swagger UI interface.

**Access Swagger UI:**
```
http://localhost:8080/swagger/index.html
```

**Features:**
- 📋 Complete API endpoint documentation
- 🔧 Interactive API testing interface
- 📝 Request/response schema definitions
- 🔑 Built-in authentication support
- 📖 Model definitions and examples

**Quick Start with Swagger:**

1. **Start the server:**
   ```bash
   ./scripts/start-swagger-demo.sh
   ```

2. **Open Swagger UI:**
   ```
   http://localhost:8080/swagger/index.html
   ```

3. **Authenticate:**
   - Click "Authorize" button
   - Enter API key: `demo-api-key-12345`
   - Or use Bearer token: `Bearer demo-api-key-12345`

4. **Test APIs:**
   - Try the health endpoint first
   - Create a new certificate with POST /keys
   - List certificates with GET /keys
   - Upload certificates and generate PFX files

**Regenerating Documentation:**

If you modify API endpoints or models, regenerate the Swagger docs:

```bash
# Install swag CLI (one time)
go install github.com/swaggo/swag/cmd/swag@latest

# Regenerate documentation
swag init -g cmd/server/main.go -o docs --parseInternal
```

### API Examples

Here are some common API usage examples:

## Configuration

The application can be configured using environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `SERVER_HOST` | `0.0.0.0` | Server bind address |
| `SERVER_PORT` | `8080` | Server port |
| `AWS_REGION` | `us-east-1` | AWS region |
| `DYNAMODB_TABLE` | `certificate-monkey` | DynamoDB table name |
| `KMS_KEY_ID` | `alias/certificate-monkey` | KMS key for encryption |
| `API_KEY_1` | `cm_dev_12345` | Primary API key |
| `API_KEY_2` | `cm_prod_67890` | Secondary API key |

## AWS Infrastructure Requirements

### DynamoDB Table

**⚠️ IMPORTANT**: You must create the DynamoDB table before running the application. The application follows infrastructure-as-code best practices and does not manage its own infrastructure.

**Required Table Configuration:**

```bash
# Table Name: certificate-monkey (configurable via DYNAMODB_TABLE env var)

# Primary Key
- Partition Key: id (String)

# Global Secondary Index
- Index Name: created_at-index
- Partition Key: created_at (String)
- Projection: ALL

# Recommended Settings for Production
- Billing Mode: On-Demand (or Provisioned based on your needs)
- Encryption: Enabled with AWS managed key
- Point-in-time Recovery: Enabled
- Backup: Enabled
```

**Using AWS CLI:**
```bash
# Create the table
aws dynamodb create-table \
    --table-name certificate-monkey \
    --attribute-definitions \
        AttributeName=id,AttributeType=S \
        AttributeName=created_at,AttributeType=S \
    --key-schema \
        AttributeName=id,KeyType=HASH \
    --global-secondary-indexes \
        'IndexName=created_at-index,KeySchema=[{AttributeName=created_at,KeyType=HASH}],Projection={ProjectionType=ALL},BillingMode=PAY_PER_REQUEST' \
    --billing-mode PAY_PER_REQUEST
```

**Using Terraform:**
```hcl
resource "aws_dynamodb_table" "certificate_monkey" {
  name           = "certificate-monkey"
  billing_mode   = "PAY_PER_REQUEST"
  hash_key       = "id"

  attribute {
    name = "id"
    type = "S"
  }

  attribute {
    name = "created_at"
    type = "S"
  }

  global_secondary_index {
    name     = "created_at-index"
    hash_key = "created_at"
    projection_type = "ALL"
  }

  server_side_encryption {
    enabled = true
  }

  point_in_time_recovery {
    enabled = true
  }

  tags = {
    Name        = "certificate-monkey"
    Environment = "production"
  }
}
```

### KMS Key

Create a KMS key for encrypting private keys:

```bash
# Create KMS key
aws kms create-key --description "Certificate Monkey encryption key"

# Create alias
aws kms create-alias --alias-name alias/certificate-monkey --target-key-id <key-id>
```

### IAM Permissions

The application requires the following AWS permissions (following least privilege principle):

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "dynamodb:PutItem",
        "dynamodb:GetItem",
        "dynamodb:UpdateItem",
        "dynamodb:DeleteItem",
        "dynamodb:Scan"
      ],
      "Resource": [
        "arn:aws:dynamodb:*:*:table/certificate-monkey",
        "arn:aws:dynamodb:*:*:table/certificate-monkey/index/*"
      ]
    },
    {
      "Effect": "Allow",
      "Action": [
        "kms:Encrypt",
        "kms:Decrypt"
      ],
      "Resource": "arn:aws:kms:*:*:key/your-kms-key-id",
      "Condition": {
        "StringEquals": {
          "kms:ViaService": "dynamodb.your-region.amazonaws.com"
        }
      }
    }
  ]
}
```

**Note**: Replace `your-kms-key-id` and `your-region` with your actual values. The application does **not** require admin permissions like `CreateTable`, `DescribeTable`, or `GenerateDataKey`.

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   API Client    │───▶│  Certificate    │───▶│   DynamoDB      │
│                 │    │    Monkey       │    │   (Storage)     │
└─────────────────┘    │                 │    └─────────────────┘
                       │                 │           │
                       │                 │           │ encrypted
                       │                 │           ▼
                       │                 │    ┌─────────────────┐
                       │                 │───▶│   AWS KMS       │
                       │                 │    │ (Encryption)    │
                       └─────────────────┘    └─────────────────┘
```

## Security Features

- **Private Key Encryption**: All private keys are encrypted using AWS KMS before storage
- **API Key Authentication**: Secure access control with configurable API keys
- **Input Validation**: Comprehensive validation of all inputs
- **Certificate Validation**: Ensures uploaded certificates match their CSRs
- **No Sensitive Data Exposure**: Private keys are redacted in API responses

## Development

### Project Structure

```
certificate-monkey/
├── cmd/server/           # Application entry point
├── internal/
│   ├── api/
│   │   ├── handlers/     # HTTP request handlers
│   │   ├── middleware/   # Authentication, logging
│   │   └── routes/       # Route definitions
│   ├── config/           # Configuration management
│   ├── crypto/           # Cryptographic operations
│   ├── models/           # Data structures
│   ├── storage/          # DynamoDB operations
│   └── version/          # Version management
├── Dockerfile            # Container configuration
├── VERSION               # Current version number
├── CHANGELOG.md          # Version history and changes
└── README.md
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./internal/crypto -v
```

### Helper Scripts

The project includes several utility scripts to streamline development and testing:

#### Certificate Workflow Scripts

**CSR Signing Script** (`scripts/sign-csr.sh`):
```bash
# Setup test CA (first time only)
./scripts/sign-csr.sh --setup-ca

# Sign a CSR
./scripts/sign-csr.sh -c request.csr -o certificate.crt

# See all options
./scripts/sign-csr.sh --help
```

**CSR Extraction Script** (`scripts/extract-csr.sh`):
```bash
# Extract CSR from API
./scripts/extract-csr.sh -k key-id -a api-key -o output.csr

# See complete workflow examples
./scripts/extract-csr.sh --workflow
```

**Complete PFX Workflow Test** (`scripts/test-pfx-workflow.sh`):
```bash
# Run end-to-end PFX generation test
./scripts/test-pfx-workflow.sh
```

This script demonstrates the complete workflow:
1. Create private key and CSR via API
2. Extract CSR from API response
3. Sign CSR with test CA
4. Upload certificate via API
5. Generate PFX file via API
6. Validate the generated PFX

#### Test Execution Script

**Comprehensive Test Runner** (`scripts/run-tests.sh`):
```bash
# Run all tests with detailed output
./scripts/run-tests.sh

# Features:
# - Colored output for better readability
# - Individual package test results
# - Coverage reporting
# - Test summary and timing
# - CI/CD integration ready
```

### Building for Production

```bash
# Build binary
go build -o certificate-monkey cmd/server/main.go

# Build Docker image
docker build -t certificate-monkey .
```

## Versioning

Certificate Monkey follows [Semantic Versioning (SemVer)](https://semver.org/) for releases.

### Version Format

```
MAJOR.MINOR.PATCH
```

- **MAJOR**: Incompatible API changes
- **MINOR**: Backwards-compatible functionality additions
- **PATCH**: Backwards-compatible bug fixes

### Current Version

```bash
# Check current version
make version

# Show project information
make info
```

### Version Management

```bash
# Bump patch version (0.1.0 → 0.1.1)
make version-patch

# Bump minor version (0.1.0 → 0.2.0)
make version-minor

# Bump major version (0.1.0 → 1.0.0)
make version-major
```

### Build Information

The application embeds build information at compile time:

```bash
# Build with version information
make build

# Build release binary (optimized)
make build-release

# Check build info via API
curl http://localhost:8080/build-info
```

### Release Process

```bash
# Prepare for release (tests, docs, validation)
make release-prepare

# The release process will guide you through:
# 1. Update CHANGELOG.md with new version
# 2. Commit changes
# 3. Tag the release
# 4. Push to repository
```

### Changelog

All notable changes are documented in [CHANGELOG.md](CHANGELOG.md) following the [Keep a Changelog](https://keepachangelog.com/) format.

The changelog includes:
- **Added**: New features
- **Changed**: Changes to existing functionality
- **Deprecated**: Features that will be removed
- **Removed**: Features that have been removed
- **Fixed**: Bug fixes
- **Security**: Security improvements

### Pre-1.0 Development

During the 0.x.x series, the API is considered unstable and may include breaking changes in minor versions. Once the API stabilizes, version 1.0.0 will be released with a commitment to backwards compatibility.

## CI/CD Pipeline

Certificate Monkey uses GitHub Actions for continuous integration and deployment with comprehensive automated workflows.

### Workflows Overview

#### 1. Pull Request Validation (`.github/workflows/pr-validation.yml`)

Automatically runs on every pull request to ensure code quality:

- **Testing**: Full test suite with race detection and coverage reporting
- **Linting**: Code quality checks with golangci-lint
- **Security Scanning**: Vulnerability analysis with Gosec
- **Docker Build Test**: Verifies container builds and runs correctly
- **Coverage Enforcement**: Ensures minimum 80% test coverage
- **Multi-platform Support**: Tests Linux AMD64 and ARM64 builds

```yaml
# Triggered on:
# - Pull requests to main/develop
# - Pushes to develop branch
```

#### 2. Release Pipeline (`.github/workflows/release.yml`)

Builds and publishes Docker containers on releases and main branch pushes:

- **Automated Testing**: Validates all tests pass before deployment
- **Docker Registry**: Publishes to GitHub Container Registry (ghcr.io)
- **Multi-platform Images**: Builds for Linux AMD64 and ARM64
- **Version Tagging**: Creates semantic version tags and latest tag
- **Security Scanning**: Trivy vulnerability scanning and SBOM generation
- **Release Creation**: Automated GitHub releases with changelog integration
- **Staging Deployment**: Optional staging environment deployment

```yaml
# Triggered on:
# - Pushes to main branch
# - Git tags (v*)
# - Published releases
```

#### 3. Security Analysis (`.github/workflows/codeql.yml`)

Continuous security monitoring:

- **CodeQL Analysis**: GitHub's semantic code analysis
- **Scheduled Scans**: Weekly security scans
- **Dependency Scanning**: Monitors for vulnerable dependencies
- **Security Alerts**: Integration with GitHub Security tab

```yaml
# Triggered on:
# - Pushes to main/develop
# - Pull requests to main
# - Weekly schedule (Mondays 2:30 AM UTC)
```

### Container Registry

Docker images are published to GitHub Container Registry:

```bash
# Pull latest image
docker pull ghcr.io/username/certificate-monkey:latest

# Pull specific version
docker pull ghcr.io/username/certificate-monkey:0.1.0

# Available tags:
# - latest (main branch)
# - semver versions (v0.1.0, 0.1.0, 0.1)
# - branch names (main, develop)
# - commit SHAs (main-abc1234)
```

### Local Docker Commands

```bash
# Build Docker image locally
make docker-build

# Run container locally
make docker-run

# Test container health
make docker-test

# View container logs
make docker-logs

# Stop and cleanup
make docker-stop
make docker-clean
```

### Build Information

All builds include embedded metadata:

```bash
# Check build information
curl http://localhost:8080/build-info

# Example response:
{
  "service": "certificate-monkey",
  "version": "0.1.0",
  "build_time": "2024-01-15_14:30:25_UTC",
  "git_commit": "abc1234",
  "go_version": "go1.22",
  "timestamp": "2024-01-15T14:30:25Z"
}
```

### Environment Variables for CI/CD

Required for production deployments:

```bash
# Application configuration
API_KEY_1=your-production-api-key
DYNAMODB_TABLE=certificate-monkey-prod
KMS_KEY_ID=alias/certificate-monkey-prod

# Optional: AWS credentials (if not using IAM roles)
AWS_ACCESS_KEY_ID=your-access-key
AWS_SECRET_ACCESS_KEY=your-secret-key
AWS_REGION=us-east-1
```

### Development Workflow

1. **Feature Development**:
   ```bash
   git checkout -b feature/new-feature
   # Make changes
   git push origin feature/new-feature
   # Create pull request
   ```

2. **PR Validation**: Automated checks run on pull request
   - All tests must pass
   - Linting must pass
   - Docker build must succeed
   - Coverage must be ≥80%

3. **Merge to Main**: After PR approval
   ```bash
   git checkout main
   git merge feature/new-feature
   git push origin main
   ```

4. **Release Process**:
   ```bash
   # Update version and changelog
   make version-minor
   # Edit CHANGELOG.md
   make release-prepare

   # Create release
   git add .
   git commit -m "Release v0.2.0"
   git tag v0.2.0
   git push origin main --tags
   ```

5. **Automated Deployment**: GitHub Actions automatically:
   - Builds and tests the release
   - Creates Docker images
   - Publishes to container registry
   - Creates GitHub release with changelog
   - Deploys to staging environment

### Monitoring and Observability

- **Health Checks**: Built-in container health checks
- **Build Artifacts**: Coverage reports and SBOMs
- **Security Reports**: Vulnerability scanning results
- **Performance**: Multi-platform build optimization
- **Dependency Tracking**: Automated dependency updates

## Future Enhancements

- [x] Complete PFX generation implementation
- [ ] Certificate template system
- [ ] Certificate expiration monitoring
- [ ] Audit logging
- [ ] Role-based access control
- [ ] Webhook notifications
- [ ] Certificate chain validation
- [ ] Integration with external CAs

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Support

For issues and questions, please open a GitHub issue or contact the development team.

---

**Certificate Monkey** - Secure certificate management made simple! 🐒🔐
