# TODO - Certificate Monkey

This document tracks technical debt, planned improvements, and future enhancements for Certificate Monkey.

## 🧪 Testing & Quality

### High Priority
- [ ] **Increase Unit Test Coverage**
  - Current: 33.5%
  - Target: 80%+
  - Focus areas:
    - [ ] `internal/api/handlers/` - Add comprehensive handler tests with proper mocking
    - [ ] `internal/storage/dynamodb.go` - Add tests for DynamoDB operations
    - [ ] `internal/crypto/operations.go` - Expand crypto operation test coverage
    - [ ] Error handling and edge cases across all packages
  - CI currently set to fail below 33% (temporary threshold)

### Medium Priority
- [ ] **Integration Tests**
  - [ ] End-to-end API workflow tests
  - [ ] Docker container integration tests
  - [ ] AWS service integration tests (with localstack)

## 🔧 Technical Debt

### Code Quality
- [ ] **Dependency Injection Improvements**
  - Current handlers use concrete types instead of interfaces
  - Makes testing with mocks difficult
  - Consider using interfaces for storage and crypto services

- [ ] **Error Handling Standardization**
  - [ ] Implement consistent error wrapping patterns
  - [ ] Add structured logging with context
  - [ ] Standard HTTP error responses

### Performance
- [ ] **Caching Layer**
  - [ ] Add Redis cache for certificate metadata
  - [ ] Cache frequently accessed certificates
  - [ ] Implement cache invalidation strategy

## 🚀 Features & Enhancements

### API Improvements
- [ ] **Certificate Validation**
  - [ ] Certificate chain validation
  - [ ] CRL (Certificate Revocation List) checking
  - [ ] OCSP (Online Certificate Status Protocol) support

- [ ] **Bulk Operations**
  - [ ] Bulk certificate import/export
  - [ ] Batch certificate renewal
  - [ ] CSV/Excel import/export functionality

### Security
- [ ] **Enhanced Authentication**
  - [ ] JWT token support
  - [ ] Role-based access control (RBAC)
  - [ ] API rate limiting per user/key

- [ ] **Audit Logging**
  - [ ] Comprehensive audit trail for all operations
  - [ ] Sensitive operation logging (private key access)
  - [ ] Compliance reporting features

### Monitoring & Observability
- [ ] **Metrics & Monitoring**
  - [ ] Prometheus metrics integration
  - [ ] Certificate expiration alerts
  - [ ] Health check improvements with dependency status

- [ ] **Distributed Tracing**
  - [ ] OpenTelemetry integration
  - [ ] Request tracing across services
  - [ ] Performance monitoring

## 📖 Documentation

- [ ] **API Documentation**
  - [ ] Interactive Swagger examples
  - [ ] Postman collection
  - [ ] SDK/client library documentation

- [ ] **Deployment Guides**
  - [ ] Kubernetes deployment manifests
  - [ ] Terraform infrastructure as code
  - [ ] Production deployment best practices

## 🏗️ Infrastructure

### Scalability
- [ ] **Database Optimizations**
  - [ ] DynamoDB GSI optimization
  - [ ] Query performance analysis
  - [ ] Partition key strategy review

- [ ] **Multi-Region Support**
  - [ ] Cross-region replication
  - [ ] Disaster recovery procedures
  - [ ] Region failover automation

### Development
- [ ] **Developer Experience**
  - [ ] Development environment automation (docker-compose)
  - [ ] Mock services for local development
  - [ ] Pre-commit hooks for code quality

## 📋 Maintenance

### Dependencies
- [ ] **Security Updates**
  - [ ] Regular dependency vulnerability scanning
  - [ ] Automated security patch management
  - [ ] Go version upgrade strategy

### Code Cleanup
- [ ] **Code Organization**
  - [ ] Package structure review
  - [ ] Remove unused code/dependencies
  - [ ] Consistent naming conventions

---

## 📝 How to Contribute

1. Pick an item from this TODO list
2. Create an issue on GitHub referencing the TODO item
3. Submit a PR with your implementation
4. Update this TODO list when items are completed

## 🎯 Current Sprint Focus

**H1 2025 Goals:**
- Achieve 80% test coverage
- Implement basic RBAC
- Add certificate expiration monitoring
- Improve error handling consistency

**Next Release (v0.2.0):**
- Unit test coverage improvements
- Enhanced error responses
- Basic audit logging
- Performance optimizations
