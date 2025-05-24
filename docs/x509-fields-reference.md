# X.509 Certificate Fields Reference

This document provides detailed information about all supported X.509 certificate fields in Certificate Monkey.

## Standard X.509 Distinguished Name Fields

Certificate Monkey supports all standard X.509 certificate fields that appear in the Subject Distinguished Name (DN):

| Field | API Parameter | X.509 Attribute | Description | Example |
|-------|---------------|-----------------|-------------|---------|
| **Common Name** | `common_name` | CN | Primary identifier, usually a domain name or person's name | `example.com`, `John Doe` |
| **Organization** | `organization` | O | Legal organization name | `ACME Corporation Ltd` |
| **Organizational Unit** | `organizational_unit` | OU | Department or division | `Information Technology` |
| **Country** | `country` | C | Two-letter ISO country code | `US`, `CA`, `GB`, `DE` |
| **State/Province** | `state` | ST | State or province name | `California`, `Ontario` |
| **City/Locality** | `city` | L | City or locality name | `San Francisco`, `Toronto` |
| **Email Address** | `email_address` | emailAddress | Email address | `admin@example.com` |

## Additional Fields

| Field | API Parameter | Description | Example |
|-------|---------------|-------------|---------|
| **Subject Alternative Names** | `subject_alternative_names` | Additional domain names, IP addresses | `["www.example.com", "192.168.1.1"]` |
| **Key Type** | `key_type` | Cryptographic algorithm and key size | `RSA2048`, `ECDSA-P256` |
| **Tags** | `tags` | Custom metadata for organization | `{"environment": "prod"}` |

## Example CSR Subject Lines

### Minimal Certificate (Common Name only)
```
Subject: CN=example.com
```

### Basic Web Server Certificate
```
Subject: C=US, ST=California, L=San Francisco, O=ACME Corp, CN=example.com
```

### Full Corporate Certificate
```
Subject: emailAddress=admin@example.com, CN=secure.example.com, OU=Information Technology, O=ACME Corporation Ltd, L=San Francisco, ST=California, C=US
```

### Code Signing Certificate
```
Subject: emailAddress=john.doe@example.com, CN=John Doe, O=ACME Corporation, C=US
```

## Field Validation Rules

### Required Fields
- `common_name`: Must be provided, 1-64 characters
- `key_type`: Must be one of the supported key types

### Optional Fields
- All other fields are optional
- Empty fields are automatically excluded from the CSR
- No minimum length requirements for optional fields

### Country Code Validation
- Must be exactly 2 characters if provided
- Should follow ISO 3166-1 alpha-2 standard
- Examples: `US`, `CA`, `GB`, `DE`, `FR`, `JP`, `AU`

### Email Address Format
- Must be a valid email format if provided
- Used for contact information and some certificate types
- Can be included in SAN (Subject Alternative Names) automatically

## Certificate Types and Common Field Combinations

### 1. SSL/TLS Web Server Certificates
```json
{
  "common_name": "www.example.com",
  "subject_alternative_names": ["example.com", "api.example.com"],
  "organization": "ACME Corporation",
  "country": "US",
  "state": "California",
  "city": "San Francisco"
}
```

### 2. Code Signing Certificates
```json
{
  "common_name": "John Doe",
  "email_address": "john.doe@company.com",
  "organization": "Software Company Inc",
  "country": "US"
}
```

### 3. Client Authentication Certificates
```json
{
  "common_name": "John Doe",
  "email_address": "john.doe@company.com",
  "organization": "ACME Corporation",
  "organizational_unit": "Engineering",
  "country": "US"
}
```

### 4. Internal Infrastructure Certificates
```json
{
  "common_name": "internal-api.local",
  "subject_alternative_names": ["10.0.1.100", "internal-api.company.local"],
  "organization": "ACME Corporation",
  "organizational_unit": "Infrastructure",
  "country": "US"
}
```

## Best Practices

### 1. Organization Certificates
- Always include `organization` for commercial certificates
- Use full legal entity name
- Include `organizational_unit` for large organizations

### 2. Geographic Information
- Include `country` for certificates that will be validated internationally
- Add `state` and `city` for enhanced verification
- Use standardized names (avoid abbreviations)

### 3. Email Addresses
- Use role-based emails for service certificates (`ssl-admin@company.com`)
- Use personal emails for individual certificates
- Ensure email address is monitored for certificate lifecycle notifications

### 4. Subject Alternative Names
- Include all hostnames that will use the certificate
- Add IP addresses for internal services
- Consider future subdomains and services

### 5. Tags for Organization
- Use consistent tagging strategies
- Include environment (`prod`, `staging`, `dev`)
- Add cost center or team ownership information
- Include expiry notification contacts

## OpenSSL Verification

You can verify the CSR contents using OpenSSL:

```bash
# View CSR subject and details
echo "CERTIFICATE_CSR_HERE" | openssl req -noout -text

# View just the subject line
echo "CERTIFICATE_CSR_HERE" | openssl req -noout -subject

# Verify CSR signature
echo "CERTIFICATE_CSR_HERE" | openssl req -noout -verify
```

## Common Issues

### 1. Missing Required Fields
- Ensure `common_name` is always provided
- Verify `key_type` is one of the supported values

### 2. Invalid Country Codes
- Use ISO 3166-1 alpha-2 codes only
- Common mistake: Using `USA` instead of `US`

### 3. Email Format Issues
- Ensure proper email format: `user@domain.com`
- Avoid special characters that might cause parsing issues

### 4. Character Encoding
- Use UTF-8 encoding for international characters
- Be cautious with special characters in organization names
