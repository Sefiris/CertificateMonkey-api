# Helm Chart Integration Guide

This document outlines the versioning strategy and integration approach for future Helm chart packaging of Certificate Monkey.

## Overview

Certificate Monkey is prepared for Helm chart packaging with:

- ✅ **Semantic Versioning** for both application and chart versions
- ✅ **OCI-compliant Docker images** with comprehensive metadata
- ✅ **Multi-platform support** (linux/amd64, linux/arm64)
- ✅ **Helm-friendly image tags** (major, minor, patch versions)
- ✅ **Security scanning** with SBOM generation
- ✅ **Automated releases** synchronized with application versions

## Versioning Strategy

### Application vs Chart Versioning

When Helm charts are created, they will follow this versioning approach:

```yaml
# Chart.yaml
apiVersion: v2
name: certificate-monkey
version: 0.1.0        # Chart version (follows SemVer)
appVersion: "0.1.0"   # Application version (matches VERSION file)
```

**Key Principles:**

1. **Chart version** (`version`) tracks Helm chart changes
   - Bump for chart template modifications
   - Bump for values.yaml changes
   - Bump for breaking chart API changes

2. **App version** (`appVersion`) tracks the Certificate Monkey application
   - Always matches the application VERSION file
   - References the Docker image tag
   - Updated automatically on each release

3. **Synchronization approach:**
   - Initial release: Chart version matches app version (1:1)
   - Subsequent updates: Chart version increments independently
   - Chart version ≥ app version (chart can be ahead)

### Version Synchronization Examples

```bash
# Scenario 1: New app feature (minor bump)
App: 0.1.0 -> 0.2.0
Chart: 0.1.0 -> 0.2.0  (templates unchanged)

# Scenario 2: Chart-only change (chart template update)
App: 0.2.0 (unchanged)
Chart: 0.2.0 -> 0.2.1  (fixed template bug)

# Scenario 3: Both app and chart change
App: 0.2.0 -> 0.3.0
Chart: 0.2.1 -> 0.3.0  (updated for new app features)

# Scenario 4: Breaking chart change
App: 0.3.0 (unchanged)
Chart: 0.3.0 -> 1.0.0  (major Helm API redesign)
```

## Docker Image Reference Patterns

### Image Tags for Helm

The Docker images are published with multiple tags suitable for different Helm deployment strategies:

```yaml
# values.yaml patterns

# 1. Pinned patch version (RECOMMENDED for production)
image:
  repository: ghcr.io/your-org/certificate-monkey
  tag: "0.1.0"  # or "v0.1.0"
  pullPolicy: IfNotPresent

# 2. Minor version tracking (auto-receive patches)
image:
  repository: ghcr.io/your-org/certificate-monkey
  tag: "0.1"  # or "v0.1" - gets latest 0.1.x
  pullPolicy: Always

# 3. Major version tracking (auto-receive features)
image:
  repository: ghcr.io/your-org/certificate-monkey
  tag: "0"  # or "v0" - gets latest 0.x.x
  pullPolicy: Always

# 4. SHA-based (immutable, for compliance)
image:
  repository: ghcr.io/your-org/certificate-monkey
  tag: "main-abc1234"
  pullPolicy: IfNotPresent

# 5. Latest (NOT recommended for production)
image:
  repository: ghcr.io/your-org/certificate-monkey
  tag: "latest"
  pullPolicy: Always
```

### Tag Selection Guide

| Use Case | Recommended Tag | Update Frequency | Stability |
|----------|----------------|------------------|-----------|
| Production | `0.1.0` (pinned) | Manual | Highest |
| Staging | `0.1` (minor) | Automatic patches | High |
| Development | `0` (major) | All updates | Medium |
| Testing | `main-sha` | On commit | Variable |
| CI/CD | `0.1.0` or SHA | Manual | Highest |

## Chart Structure (Future Implementation)

When creating the Helm chart, follow this structure:

```
helm/
├── certificate-monkey/
│   ├── Chart.yaml              # Chart metadata
│   ├── values.yaml             # Default configuration
│   ├── values.schema.json      # Values validation
│   ├── README.md               # Chart documentation
│   ├── .helmignore             # Files to ignore
│   ├── templates/
│   │   ├── NOTES.txt           # Post-install notes
│   │   ├── _helpers.tpl        # Template helpers
│   │   ├── deployment.yaml     # Main deployment
│   │   ├── service.yaml        # Service definition
│   │   ├── serviceaccount.yaml # Service account
│   │   ├── secret.yaml         # API keys, config
│   │   ├── configmap.yaml      # Configuration
│   │   ├── ingress.yaml        # Ingress (optional)
│   │   ├── hpa.yaml            # Horizontal Pod Autoscaler
│   │   └── tests/
│   │       └── test-connection.yaml
│   └── crds/                   # Custom Resource Definitions (if needed)
```

## Chart.yaml Template

```yaml
apiVersion: v2
name: certificate-monkey
description: Secure certificate and private key management API with AWS KMS encryption
type: application
version: 0.1.0
appVersion: "0.1.0"

keywords:
  - certificate
  - pki
  - tls
  - ssl
  - kms
  - security
  - aws

home: https://github.com/your-org/certificate-monkey
sources:
  - https://github.com/your-org/certificate-monkey

maintainers:
  - name: Certificate Monkey Team
    email: support@certificatemonkey.dev
    url: https://github.com/your-org/certificate-monkey

icon: https://raw.githubusercontent.com/your-org/certificate-monkey/main/docs/logo.png

dependencies: []

annotations:
  category: Security
  licenses: MIT
  artifacthub.io/changes: |
    - kind: added
      description: Initial Helm chart release
  artifacthub.io/containsSecurityUpdates: "false"
  artifacthub.io/prerelease: "false"
  artifacthub.io/recommendations: |
    - url: https://artifacthub.io/packages/helm/bitnami/aws-load-balancer-controller
```

## Values.yaml Template (Key Sections)

```yaml
# Image configuration
image:
  repository: ghcr.io/your-org/certificate-monkey
  pullPolicy: IfNotPresent
  tag: ""  # Defaults to .Chart.AppVersion

# Application configuration
config:
  serverPort: 8080
  serverHost: "0.0.0.0"

  # AWS configuration
  aws:
    region: us-east-1
    dynamodbTable: certificate-monkey
    kmsKeyId: alias/certificate-monkey

# Authentication
apiKeys:
  # Option 1: Use existing secret
  existingSecret: ""
  existingSecretKeys:
    primary: API_KEY_1
    secondary: API_KEY_2

  # Option 2: Create keys inline (less secure)
  primary: ""
  secondary: ""

# Service configuration
service:
  type: ClusterIP
  port: 80
  targetPort: 8080
  annotations: {}

# Ingress configuration
ingress:
  enabled: false
  className: ""
  annotations: {}
  hosts:
    - host: certificate-monkey.local
      paths:
        - path: /
          pathType: Prefix
  tls: []

# Resource limits
resources:
  limits:
    cpu: 500m
    memory: 512Mi
  requests:
    cpu: 250m
    memory: 256Mi

# Autoscaling
autoscaling:
  enabled: false
  minReplicas: 2
  maxReplicas: 10
  targetCPUUtilizationPercentage: 80
  targetMemoryUtilizationPercentage: 80

# Pod security
podSecurityContext:
  runAsNonRoot: true
  runAsUser: 1000
  fsGroup: 1000
  seccompProfile:
    type: RuntimeDefault

securityContext:
  allowPrivilegeEscalation: false
  capabilities:
    drop:
      - ALL
  readOnlyRootFilesystem: true

# Service Account
serviceAccount:
  create: true
  annotations:
    # AWS IRSA annotation
    eks.amazonaws.com/role-arn: ""
  name: ""

# Health checks
livenessProbe:
  httpGet:
    path: /health
    port: http
  initialDelaySeconds: 30
  periodSeconds: 10

readinessProbe:
  httpGet:
    path: /health
    port: http
  initialDelaySeconds: 5
  periodSeconds: 5
```

## Publishing Strategy

### OCI Registry (Recommended)

Helm 3 supports OCI registries (same as Docker images):

```bash
# Package and push to GitHub Container Registry
helm package ./helm/certificate-monkey
helm push certificate-monkey-0.1.0.tgz oci://ghcr.io/your-org/helm-charts

# Install from OCI registry
helm install certificate-monkey oci://ghcr.io/your-org/helm-charts/certificate-monkey --version 0.1.0
```

**Advantages:**
- ✅ Same infrastructure as Docker images
- ✅ GitHub Container Registry support
- ✅ Integrated authentication
- ✅ Version immutability

### Traditional Chart Repository (Alternative)

If needed, charts can be published to a traditional Helm repository:

```bash
# Create GitHub Pages repository
helm repo index . --url https://your-org.github.io/helm-charts/
git add index.yaml certificate-monkey-*.tgz
git commit -m "Add Certificate Monkey chart v0.1.0"
git push

# Users add and install
helm repo add certificate-monkey https://your-org.github.io/helm-charts/
helm install my-release certificate-monkey/certificate-monkey
```

## Automated Chart Releases

### GitHub Actions Workflow

Add to `.github/workflows/helm-release.yml`:

```yaml
name: Helm Chart Release

on:
  push:
    branches: [main]
    paths:
      - 'helm/**'
      - VERSION
  release:
    types: [published]

jobs:
  release-chart:
    runs-on: ubuntu-latest
    permissions:
      contents: read
      packages: write

    steps:
    - name: Checkout
      uses: actions/checkout@v4

    - name: Install Helm
      uses: azure/setup-helm@v3

    - name: Get version
      id: version
      run: echo "version=$(cat VERSION)" >> $GITHUB_OUTPUT

    - name: Update Chart version
      run: |
        # Update appVersion in Chart.yaml
        sed -i "s/^appVersion:.*$/appVersion: \"${{ steps.version.outputs.version }}\"/" helm/certificate-monkey/Chart.yaml

    - name: Lint chart
      run: helm lint helm/certificate-monkey

    - name: Package chart
      run: helm package helm/certificate-monkey

    - name: Login to GHCR
      run: |
        echo ${{ secrets.GITHUB_TOKEN }} | helm registry login ghcr.io -u ${{ github.actor }} --password-stdin

    - name: Push chart to GHCR
      run: |
        helm push certificate-monkey-*.tgz oci://ghcr.io/${{ github.repository_owner }}/helm-charts
```

## Version Management Integration

### Makefile Commands (Future)

Add these commands when charts are created:

```makefile
# Helm chart commands
.PHONY: helm-lint helm-package helm-install helm-test

helm-lint: ## Lint Helm chart
	@echo "🔍 Linting Helm chart..."
	@helm lint helm/certificate-monkey

helm-package: ## Package Helm chart with current version
	@echo "📦 Packaging Helm chart..."
	@sed -i.bak "s/^appVersion:.*$$/appVersion: \"$(CURRENT_VERSION)\"/" helm/certificate-monkey/Chart.yaml
	@helm package helm/certificate-monkey
	@rm -f helm/certificate-monkey/Chart.yaml.bak

helm-install: ## Install chart locally for testing
	@echo "🚀 Installing Helm chart..."
	@helm install certificate-monkey-test ./helm/certificate-monkey \
		--set image.tag=$(CURRENT_VERSION) \
		--set apiKeys.primary=test-key

helm-test: ## Run Helm chart tests
	@echo "🧪 Testing Helm chart..."
	@helm test certificate-monkey-test

helm-uninstall: ## Uninstall test chart
	@helm uninstall certificate-monkey-test
```

### Version Script Enhancement

The `version-manager.sh` script should validate Helm compatibility:

```bash
# Validate version is Helm-compatible
validate_helm_version() {
    local version="$1"

    # Helm requires SemVer format
    if ! echo "$version" | grep -qE '^[0-9]+\.[0-9]+\.[0-9]+$'; then
        log_error "Version $version is not valid SemVer (required for Helm)"
        exit 1
    fi

    log_success "Version $version is Helm-compatible"
}
```

## Installation Examples

### Basic Installation

```bash
# Install with default values
helm install certificate-monkey \
  oci://ghcr.io/your-org/helm-charts/certificate-monkey \
  --version 0.1.0

# Install with custom values
helm install certificate-monkey \
  oci://ghcr.io/your-org/helm-charts/certificate-monkey \
  --version 0.1.0 \
  --values custom-values.yaml

# Install with inline values
helm install certificate-monkey \
  oci://ghcr.io/your-org/helm-charts/certificate-monkey \
  --version 0.1.0 \
  --set config.aws.region=eu-west-1 \
  --set apiKeys.existingSecret=my-api-keys
```

### Production Installation

```bash
# Production configuration
helm install certificate-monkey \
  oci://ghcr.io/your-org/helm-charts/certificate-monkey \
  --version 0.1.0 \
  --namespace certificate-monkey \
  --create-namespace \
  --values production-values.yaml \
  --wait \
  --timeout 5m
```

### Upgrade Strategy

```bash
# Upgrade to new version
helm upgrade certificate-monkey \
  oci://ghcr.io/your-org/helm-charts/certificate-monkey \
  --version 0.2.0 \
  --reuse-values \
  --wait

# Rollback if needed
helm rollback certificate-monkey 1
```

## Security Considerations

### Image Pull Secrets

```yaml
# values.yaml
imagePullSecrets:
  - name: ghcr-credentials

# Create secret
kubectl create secret docker-registry ghcr-credentials \
  --docker-server=ghcr.io \
  --docker-username=$GITHUB_ACTOR \
  --docker-password=$GITHUB_TOKEN
```

### AWS Permissions (IRSA)

```yaml
# values.yaml
serviceAccount:
  annotations:
    eks.amazonaws.com/role-arn: arn:aws:iam::123456789012:role/certificate-monkey

# Required IAM policy (same as standalone deployment)
```

### Secret Management

```bash
# Use Kubernetes secrets (recommended)
kubectl create secret generic certificate-monkey-api-keys \
  --from-literal=API_KEY_1=your-secure-key-here \
  --from-literal=API_KEY_2=your-backup-key-here

# Or use external secret management
# - AWS Secrets Manager (via External Secrets Operator)
# - HashiCorp Vault
# - Sealed Secrets
```

## Testing Strategy

### Pre-release Testing

```bash
# 1. Template validation
helm template certificate-monkey ./helm/certificate-monkey

# 2. Dry-run installation
helm install certificate-monkey ./helm/certificate-monkey --dry-run --debug

# 3. Actual test installation
helm install certificate-monkey-test ./helm/certificate-monkey \
  --set image.tag=$(cat VERSION) \
  --wait

# 4. Run chart tests
helm test certificate-monkey-test

# 5. Cleanup
helm uninstall certificate-monkey-test
```

### CI/CD Integration

```yaml
# test-chart.yaml step
- name: Test Helm chart
  run: |
    kind create cluster
    helm install test ./helm/certificate-monkey --wait --timeout 5m
    helm test test
    helm uninstall test
```

## Migration Path

When ready to create Helm charts:

1. **Create chart structure** following the template above
2. **Update workflows** to include `helm-release.yml`
3. **Add Makefile commands** for chart management
4. **Update documentation** to include Helm installation instructions
5. **Test thoroughly** in dev/staging before production
6. **Publish initial chart** synchronized with current app version
7. **Document chart values** in chart's README.md

## Resources

- [Helm Best Practices](https://helm.sh/docs/chart_best_practices/)
- [OCI Registry Support](https://helm.sh/docs/topics/registries/)
- [Chart Testing](https://github.com/helm/chart-testing)
- [Artifact Hub](https://artifacthub.io/) - Chart discovery
- [Helm Docs](https://github.com/norwoodj/helm-docs) - Auto-generate docs

## Support

For questions about Helm integration:

1. Review this document
2. Check Helm official documentation
3. Test locally before deploying
4. Create an issue for chart-specific problems

---

**Note**: This is a preparation guide. Actual Helm charts will be created when needed. The current versioning and Docker image strategy is already Helm-ready.
