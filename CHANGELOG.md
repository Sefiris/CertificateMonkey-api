# Changelog

All notable changes to Certificate Monkey will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- TBD

### Changed
- TBD

### Deprecated
- TBD

### Removed
- TBD

### Fixed
- TBD

### Security
- TBD

## [0.1.0] - 2024-01-15

### Added
- Initial release of Certificate Monkey API
- Private key generation support (RSA 2048/4096, ECDSA P-256/P-384)
- Certificate Signing Request (CSR) creation
- Certificate upload and validation
- PFX/PKCS#12 file generation with password protection
- Private key export functionality with security audit logging
- RESTful API with 6 endpoints for complete certificate lifecycle
- AWS DynamoDB storage with KMS encryption for private keys
- Comprehensive API authentication using API keys
- Interactive Swagger UI documentation at `/swagger/index.html`
- Search and filtering capabilities (by status, key type, date range, tags)
- Docker containerization support
- Infrastructure as Code with Pulumi (DynamoDB, KMS)
- Comprehensive test suite with 95%+ coverage
- Development tooling (Makefile, test scripts, demo environment)
- Security features:
  - API key authentication with masking in logs
  - Private key encryption with AWS KMS
  - Input validation and sanitization
  - Comprehensive audit logging for sensitive operations
  - CORS support for web applications
- Certificate lifecycle management:
  - Status tracking (PENDING_CSR, CSR_CREATED, CERT_UPLOADED, COMPLETED)
  - Certificate validation against CSRs
  - Automatic extraction of certificate metadata (serial number, fingerprint, validity dates)
- Tag-based organization and metadata support
- MIT License for open source distribution

### Security
- All private keys encrypted with AWS KMS before storage
- API endpoints protected with configurable API keys
- Sensitive operations logged with client information for audit trails
- Private key export includes comprehensive security warnings
- No sensitive data exposure in API responses (keys redacted by default)

---

## Version History

### Semantic Versioning Guide

This project follows [Semantic Versioning](https://semver.org/):

- **MAJOR** version (X.y.z): Incompatible API changes
- **MINOR** version (x.Y.z): Backwards-compatible functionality additions
- **PATCH** version (x.y.Z): Backwards-compatible bug fixes

### Pre-1.0 Development

During the 0.x.x series, the API is considered unstable and may include breaking changes in minor versions. Once the API stabilizes, version 1.0.0 will be released with a commitment to backwards compatibility.

### Release Notes

For detailed technical information about each release, see the [GitHub Releases](https://github.com/your-username/certificate-monkey/releases) page.
