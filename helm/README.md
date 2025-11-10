# Certificate Monkey Helm Charts

Official Helm charts for deploying Certificate Monkey components - secure certificate management with AWS KMS encryption.

## Charts

- **certificate-monkey-api**: Certificate management REST API backend

## About Certificate Monkey

Certificate Monkey provides a complete solution for managing the certificate lifecycle:
- Generate private keys (RSA 2048/4096, ECDSA P-256/P-384)
- Create certificate signing requests (CSRs)
- Upload and validate certificates
- Generate PFX/PKCS#12 files for legacy applications
- Export private keys (with comprehensive audit logging)

All private keys are encrypted with AWS KMS and stored in DynamoDB.

## Quick Start

```bash
# Add the Helm repository
helm repo add certificate-monkey https://sefiris.github.io/CertificateMonkey/
helm repo update

# Create namespace
kubectl create namespace certificate-monkey

# Create required secrets
kubectl create secret generic certificate-monkey-api-keys \
  --namespace=certificate-monkey \
  --from-literal=API_KEY_1=your-primary-key \
  --from-literal=API_KEY_2=your-secondary-key

# Install the API chart
helm install certificate-monkey-api certificate-monkey/certificate-monkey-api \
  --namespace=certificate-monkey \
  --set aws.region=us-east-1 \
  --set aws.dynamodbTable=certificate-monkey \
  --set aws.kmsKeyId=alias/certificate-monkey
```

## Chart Repository

The Helm chart repository is hosted via GitHub Pages at:

```
https://sefiris.github.io/CertificateMonkey/
```

## Available Charts

| Chart | Description | App Version | Chart Version |
|-------|-------------|-------------|---------------|
| certificate-monkey-api | API backend for certificate management | 0.1.0 | 0.1.0 |

## Prerequisites

- Kubernetes 1.24+
- Helm 3.8+
- AWS Account with DynamoDB and KMS configured

## Documentation

- **[Helm Deployment Guide](../docs/HELM_DEPLOYMENT.md)** - Complete production deployment guide
- **[Helm Testing Guide](../docs/HELM_TESTING.md)** - Minikube testing instructions
- **[Main README](../README.md)** - Application documentation

## Configuration

The chart can be configured via `values.yaml`. Key configuration options:

### Required Configuration

```yaml
aws:
  region: us-east-1
  dynamodbTable: certificate-monkey
  kmsKeyId: alias/certificate-monkey

apiKeys:
  secretName: certificate-monkey-api-keys  # Must be created before installation
```

### AWS Authentication

**Default: IRSA (IAM Roles for Service Accounts)**

For EKS clusters (recommended):

```yaml
aws:
  serviceAccount:
    create: true
    annotations:
      eks.amazonaws.com/role-arn: arn:aws:iam::ACCOUNT:role/certificate-monkey
```

**Optional: Static Credentials**

For non-EKS environments (minikube, testing):

```yaml
aws:
  credentials:
    useSecret: true
    secretName: certificate-monkey-aws
```

### Scaling and Resources

```yaml
replicaCount: 2

# Resource requests are set by default
# Limits are optional - uncomment to restrict resource usage
resources:
  requests:
    cpu: 100m
    memory: 128Mi
  # limits:
  #   cpu: 500m
  #   memory: 512Mi

autoscaling:
  enabled: true
  minReplicas: 2
  maxReplicas: 10
  targetCPUUtilizationPercentage: 70
```

### Ingress Configuration

```yaml
ingress:
  enabled: true
  className: nginx
  hosts:
    - host: certificate-monkey.yourdomain.com
      paths:
        - path: /
          pathType: Prefix
  tls:
    - secretName: certificate-monkey-tls
      hosts:
        - certificate-monkey.yourdomain.com
```

## Installation Examples

### Basic Installation

```bash
helm install certificate-monkey certificate-monkey/certificate-monkey \
  --namespace=certificate-monkey \
  --create-namespace
```

### Production Installation

```bash
helm install certificate-monkey certificate-monkey/certificate-monkey \
  --namespace=certificate-monkey \
  --create-namespace \
  --set replicaCount=3 \
  --set autoscaling.enabled=true \
  --set aws.serviceAccount.annotations."eks\.amazonaws\.com/role-arn"=arn:aws:iam::123456789012:role/certificate-monkey \
  --set ingress.enabled=true \
  --set ingress.hosts[0].host=certificate-monkey.yourdomain.com
```

### Minikube Testing

```bash
# Create secrets first
kubectl create secret generic certificate-monkey-aws \
  --namespace=certificate-monkey \
  --from-literal=AWS_ACCESS_KEY_ID=your-key \
  --from-literal=AWS_SECRET_ACCESS_KEY=your-secret

kubectl create secret generic certificate-monkey-api-keys \
  --namespace=certificate-monkey \
  --from-literal=API_KEY_1=test-key-1 \
  --from-literal=API_KEY_2=test-key-2

# Install with minikube values
helm install certificate-monkey ./certificate-monkey \
  --namespace=certificate-monkey \
  --create-namespace \
  --values ./certificate-monkey/values-minikube.yaml
```

## Upgrading

```bash
# Update repository
helm repo update

# Upgrade to latest version
helm upgrade certificate-monkey certificate-monkey/certificate-monkey \
  --namespace=certificate-monkey \
  --reuse-values
```

## Uninstalling

```bash
helm uninstall certificate-monkey --namespace=certificate-monkey
```

## Values Reference

For a complete list of configurable values, see:
- [values.yaml](certificate-monkey/values.yaml) - All available options
- [values-minikube.yaml](certificate-monkey/values-minikube.yaml) - Minikube testing configuration

## Support

For issues and questions:
- **GitHub Issues**: https://github.com/dduivenbode/CertificateMonkey/issues
- **Documentation**: https://github.com/dduivenbode/CertificateMonkey

## License

MIT License - See [LICENSE](../LICENSE) for details.
