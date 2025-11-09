# Helm Deployment Guide

Complete guide for deploying Certificate Monkey using Helm in production and development environments.

## Table of Contents

- [Quick Start](#quick-start)
- [Installation](#installation)
- [Configuration](#configuration)
- [AWS Setup](#aws-setup)
- [Production Deployment](#production-deployment)
- [Upgrading](#upgrading)
- [Monitoring](#monitoring)
- [Troubleshooting](#troubleshooting)

## Quick Start

```bash
# Add the Helm repository
helm repo add certificate-monkey https://sefiris.github.io/CertificateMonkey/
helm repo update

# Create namespace
kubectl create namespace certificate-monkey

# Create required secrets (see detailed instructions below)
kubectl create secret generic certificate-monkey-api-keys \
  --namespace=certificate-monkey \
  --from-literal=API_KEY_1=your-primary-key \
  --from-literal=API_KEY_2=your-secondary-key

# Install the chart
helm install certificate-monkey certificate-monkey/certificate-monkey \
  --namespace=certificate-monkey \
  --set aws.region=us-east-1 \
  --set aws.dynamodbTable=certificate-monkey-prod \
  --set aws.kmsKeyId=alias/certificate-monkey-prod
```

## Installation

### Prerequisites

- Kubernetes 1.24+
- Helm 3.8+
- AWS Account with DynamoDB and KMS configured
- kubectl configured to access your cluster

### Add Helm Repository

```bash
# Add the Certificate Monkey Helm repository
helm repo add certificate-monkey https://sefiris.github.io/CertificateMonkey/

# Update repository information
helm repo update

# Search for available versions
helm search repo certificate-monkey
```

### Install from GitHub Pages

```bash
helm install certificate-monkey certificate-monkey/certificate-monkey \
  --namespace=certificate-monkey \
  --create-namespace \
  --values your-values.yaml
```

### Install from Local Chart

```bash
# Clone the repository
git clone https://github.com/dduivenbode/CertificateMonkey.git
cd CertificateMonkey

# Install from local chart
helm install certificate-monkey ./helm/certificate-monkey \
  --namespace=certificate-monkey \
  --create-namespace \
  --values your-values.yaml
```

## Configuration

### Required Secrets

Certificate Monkey requires two Kubernetes secrets:

#### 1. API Keys Secret (Required)

```bash
kubectl create secret generic certificate-monkey-api-keys \
  --namespace=certificate-monkey \
  --from-literal=API_KEY_1="$(openssl rand -hex 32)" \
  --from-literal=API_KEY_2="$(openssl rand -hex 32)"
```

**Store these keys securely** - you'll need them to authenticate API requests.

#### 2. AWS Credentials Secret (Optional - for non-EKS)

Only required for non-EKS environments (development, minikube, non-AWS Kubernetes):

```bash
kubectl create secret generic certificate-monkey-aws \
  --namespace=certificate-monkey \
  --from-literal=AWS_ACCESS_KEY_ID=your-access-key \
  --from-literal=AWS_SECRET_ACCESS_KEY=your-secret-key

# Enable in values.yaml
aws:
  credentials:
    useSecret: true
    secretName: certificate-monkey-aws
```

### Basic Configuration

Create a `values.yaml` file:

```yaml
# values.yaml - Basic Configuration

replicaCount: 2

image:
  repository: ghcr.io/dduivenbode/certificatemonkey
  tag: "0.1.0"
  pullPolicy: IfNotPresent

aws:
  region: us-east-1
  dynamodbTable: certificate-monkey-prod
  kmsKeyId: alias/certificate-monkey-prod

  serviceAccount:
    create: true
    annotations:
      eks.amazonaws.com/role-arn: arn:aws:iam::123456789012:role/certificate-monkey

resources:
  requests:
    cpu: 100m
    memory: 128Mi
  # Uncomment to set limits:
  # limits:
  #   cpu: 500m
  #   memory: 512Mi

service:
  type: ClusterIP
  port: 8080
```

### Advanced Configuration

For production deployments with autoscaling, ingress, and monitoring:

```yaml
# values-production.yaml - Advanced Configuration

replicaCount: 3

image:
  repository: ghcr.io/dduivenbode/certificatemonkey
  tag: "0.1.0"
  pullPolicy: IfNotPresent

aws:
  region: us-east-1
  dynamodbTable: certificate-monkey-prod
  kmsKeyId: alias/certificate-monkey-prod

  serviceAccount:
    create: true
    annotations:
      eks.amazonaws.com/role-arn: arn:aws:iam::123456789012:role/certificate-monkey

apiKeys:
  secretName: certificate-monkey-api-keys

# Resource configuration
# For production, consider setting limits to prevent resource contention
resources:
  requests:
    cpu: 250m
    memory: 256Mi
  limits:
    cpu: 1000m
    memory: 1Gi

# Horizontal Pod Autoscaler
autoscaling:
  enabled: true
  minReplicas: 3
  maxReplicas: 10
  targetCPUUtilizationPercentage: 70
  targetMemoryUtilizationPercentage: 80

# Service configuration
service:
  type: ClusterIP
  port: 8080
  annotations:
    service.beta.kubernetes.io/aws-load-balancer-type: nlb
    service.beta.kubernetes.io/aws-load-balancer-cross-zone-load-balancing-enabled: "true"

# Ingress configuration
ingress:
  enabled: true
  className: nginx
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt-prod
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    nginx.ingress.kubernetes.io/force-ssl-redirect: "true"
  hosts:
    - host: certificate-monkey.yourdomain.com
      paths:
        - path: /
          pathType: Prefix
  tls:
    - secretName: certificate-monkey-tls
      hosts:
        - certificate-monkey.yourdomain.com

# Pod disruption budget
podDisruptionBudget:
  enabled: true
  minAvailable: 1

# Pod anti-affinity for high availability
affinity:
  podAntiAffinity:
    preferredDuringSchedulingIgnoredDuringExecution:
      - weight: 100
        podAffinityTerm:
          labelSelector:
            matchExpressions:
              - key: app.kubernetes.io/name
                operator: In
                values:
                  - certificate-monkey
          topologyKey: topology.kubernetes.io/zone

# Monitoring annotations
podAnnotations:
  prometheus.io/scrape: "true"
  prometheus.io/port: "8080"
  prometheus.io/path: "/metrics"
```

## AWS Setup

### 1. DynamoDB Table

Create a DynamoDB table with the required schema:

```bash
aws dynamodb create-table \
    --table-name certificate-monkey-prod \
    --attribute-definitions \
        AttributeName=id,AttributeType=S \
        AttributeName=created_at,AttributeType=S \
    --key-schema \
        AttributeName=id,KeyType=HASH \
    --global-secondary-indexes \
        'IndexName=created_at-index,KeySchema=[{AttributeName=created_at,KeyType=HASH}],Projection={ProjectionType=ALL},BillingMode=PAY_PER_REQUEST' \
    --billing-mode PAY_PER_REQUEST \
    --region us-east-1 \
    --tags Key=Application,Value=CertificateMonkey Key=Environment,Value=production
```

Enable point-in-time recovery and encryption:

```bash
# Enable point-in-time recovery
aws dynamodb update-continuous-backups \
    --table-name certificate-monkey-prod \
    --point-in-time-recovery-specification PointInTimeRecoveryEnabled=true \
    --region us-east-1

# Table is encrypted by default with AWS managed keys
```

### 2. KMS Key

Create a KMS key for encrypting private keys:

```bash
# Create KMS key
KMS_KEY_ID=$(aws kms create-key \
    --description "Certificate Monkey production encryption key" \
    --region us-east-1 \
    --tags TagKey=Application,TagValue=CertificateMonkey TagKey=Environment,TagValue=production \
    --query 'KeyMetadata.KeyId' \
    --output text)

# Create alias
aws kms create-alias \
    --alias-name alias/certificate-monkey-prod \
    --target-key-id $KMS_KEY_ID \
    --region us-east-1

echo "KMS Key ID: $KMS_KEY_ID"
```

Enable automatic key rotation:

```bash
aws kms enable-key-rotation \
    --key-id $KMS_KEY_ID \
    --region us-east-1
```

### 3. IAM Role for Service Account (IRSA)

For EKS clusters, configure IRSA for secure AWS authentication:

#### Create IAM Policy

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
        "dynamodb:Scan",
        "dynamodb:Query"
      ],
      "Resource": [
        "arn:aws:dynamodb:us-east-1:123456789012:table/certificate-monkey-prod",
        "arn:aws:dynamodb:us-east-1:123456789012:table/certificate-monkey-prod/index/*"
      ]
    },
    {
      "Effect": "Allow",
      "Action": [
        "kms:Encrypt",
        "kms:Decrypt",
        "kms:GenerateDataKey"
      ],
      "Resource": "arn:aws:kms:us-east-1:123456789012:key/<key-id>"
    }
  ]
}
```

#### Create IAM Role

```bash
# Set variables
CLUSTER_NAME=your-eks-cluster
NAMESPACE=certificate-monkey
SERVICE_ACCOUNT=certificate-monkey
AWS_ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)
OIDC_PROVIDER=$(aws eks describe-cluster --name $CLUSTER_NAME --query "cluster.identity.oidc.issuer" --output text | sed -e "s/^https:\/\///")

# Create trust policy
cat > trust-policy.json <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Federated": "arn:aws:iam::${AWS_ACCOUNT_ID}:oidc-provider/${OIDC_PROVIDER}"
      },
      "Action": "sts:AssumeRoleWithWebIdentity",
      "Condition": {
        "StringEquals": {
          "${OIDC_PROVIDER}:sub": "system:serviceaccount:${NAMESPACE}:${SERVICE_ACCOUNT}",
          "${OIDC_PROVIDER}:aud": "sts.amazonaws.com"
        }
      }
    }
  ]
}
EOF

# Create IAM role
aws iam create-role \
  --role-name certificate-monkey-prod \
  --assume-role-policy-document file://trust-policy.json

# Attach policy
aws iam put-role-policy \
  --role-name certificate-monkey-prod \
  --policy-name certificate-monkey-access \
  --policy-document file://iam-policy.json

# Get role ARN
ROLE_ARN=$(aws iam get-role --role-name certificate-monkey-prod --query 'Role.Arn' --output text)
echo "Role ARN: $ROLE_ARN"
```

#### Configure Service Account Annotation

Add the role ARN to your `values.yaml`:

```yaml
aws:
  serviceAccount:
    create: true
    annotations:
      eks.amazonaws.com/role-arn: arn:aws:iam::123456789012:role/certificate-monkey-prod
```

## Production Deployment

### Pre-Deployment Checklist

- [ ] AWS infrastructure created (DynamoDB, KMS, IAM role)
- [ ] Kubernetes cluster ready (EKS recommended)
- [ ] Helm 3.8+ installed
- [ ] kubectl configured with cluster access
- [ ] Namespace created
- [ ] Secrets created (API keys)
- [ ] DNS configured (if using Ingress)
- [ ] TLS certificates ready (if using HTTPS)
- [ ] Monitoring/logging configured

### Deployment Steps

1. **Create Namespace**

```bash
kubectl create namespace certificate-monkey
```

2. **Create Secrets**

```bash
# API Keys
kubectl create secret generic certificate-monkey-api-keys \
  --namespace=certificate-monkey \
  --from-literal=API_KEY_1="$(openssl rand -hex 32)" \
  --from-literal=API_KEY_2="$(openssl rand -hex 32)"
```

3. **Deploy with Helm**

```bash
helm install certificate-monkey certificate-monkey/certificate-monkey \
  --namespace=certificate-monkey \
  --values values-production.yaml \
  --wait \
  --timeout 5m
```

4. **Verify Deployment**

```bash
# Check deployment status
kubectl get deployment certificate-monkey -n certificate-monkey

# Check pod status
kubectl get pods -n certificate-monkey

# Check logs
kubectl logs -l app.kubernetes.io/name=certificate-monkey -n certificate-monkey --tail=50

# Test health endpoint
kubectl exec -n certificate-monkey \
  $(kubectl get pod -n certificate-monkey -l app.kubernetes.io/name=certificate-monkey -o jsonpath="{.items[0].metadata.name}") \
  -- curl -s localhost:8080/health
```

5. **Test API Access**

```bash
# If using LoadBalancer service
export SERVICE_URL=$(kubectl get svc certificate-monkey -n certificate-monkey -o jsonpath='{.status.loadBalancer.ingress[0].hostname}')

# If using Ingress
export SERVICE_URL=https://certificate-monkey.yourdomain.com

# Test health endpoint
curl $SERVICE_URL/health

# Test authenticated endpoint
export API_KEY=$(kubectl get secret certificate-monkey-api-keys -n certificate-monkey -o jsonpath='{.data.API_KEY_1}' | base64 --decode)
curl -H "X-API-Key: $API_KEY" $SERVICE_URL/api/v1/keys
```

## Upgrading

### Upgrade to New Version

```bash
# Update Helm repository
helm repo update

# Check available versions
helm search repo certificate-monkey --versions

# Upgrade to specific version
helm upgrade certificate-monkey certificate-monkey/certificate-monkey \
  --namespace=certificate-monkey \
  --version 0.2.0 \
  --values values-production.yaml \
  --wait

# Check rollout status
kubectl rollout status deployment/certificate-monkey -n certificate-monkey
```

### Rollback

```bash
# View release history
helm history certificate-monkey -n certificate-monkey

# Rollback to previous version
helm rollback certificate-monkey -n certificate-monkey

# Rollback to specific revision
helm rollback certificate-monkey 2 -n certificate-monkey
```

### Update Configuration

```bash
# Upgrade with new values
helm upgrade certificate-monkey certificate-monkey/certificate-monkey \
  --namespace=certificate-monkey \
  --values values-production.yaml \
  --set replicaCount=5 \
  --wait
```

## Monitoring

### Metrics and Observability

Configure monitoring for production deployments:

#### Prometheus Integration

Add Prometheus annotations to your values:

```yaml
podAnnotations:
  prometheus.io/scrape: "true"
  prometheus.io/port: "8080"
  prometheus.io/path: "/metrics"
```

#### CloudWatch Container Insights

For EKS clusters:

```bash
# Install CloudWatch agent
eksctl utils install-cloudwatch-agent \
  --cluster=your-cluster-name \
  --region=us-east-1
```

#### Logging

Configure centralized logging:

```yaml
podAnnotations:
  fluentbit.io/parser: json
```

### Health Checks

Monitor application health:

```bash
# Create monitoring script
cat > monitor-health.sh <<'EOF'
#!/bin/bash
NAMESPACE=certificate-monkey
SERVICE_URL=$(kubectl get svc certificate-monkey -n $NAMESPACE -o jsonpath='{.status.loadBalancer.ingress[0].hostname}')

while true; do
  STATUS=$(curl -s -o /dev/null -w "%{http_code}" $SERVICE_URL/health)
  if [ "$STATUS" != "200" ]; then
    echo "$(date) - Health check failed: HTTP $STATUS"
    # Send alert
  fi
  sleep 30
done
EOF

chmod +x monitor-health.sh
```

## Troubleshooting

### Common Issues

#### 1. Pods Not Starting

```bash
# Check pod events
kubectl describe pod -n certificate-monkey -l app.kubernetes.io/name=certificate-monkey

# Check logs
kubectl logs -n certificate-monkey -l app.kubernetes.io/name=certificate-monkey --tail=100

# Common causes:
# - Missing secrets
# - Invalid AWS credentials
# - Insufficient resources
# - Image pull errors
```

#### 2. AWS Connection Failures

```bash
# Verify IRSA configuration
kubectl describe sa certificate-monkey -n certificate-monkey

# Check pod has correct IAM role
kubectl exec -n certificate-monkey \
  $(kubectl get pod -n certificate-monkey -l app.kubernetes.io/name=certificate-monkey -o jsonpath="{.items[0].metadata.name}") \
  -- env | grep AWS

# Test AWS connectivity
kubectl exec -n certificate-monkey \
  $(kubectl get pod -n certificate-monkey -l app.kubernetes.io/name=certificate-monkey -o jsonpath="{.items[0].metadata.name}") \
  -- sh -c 'curl -I https://dynamodb.us-east-1.amazonaws.com'
```

#### 3. Service Not Accessible

```bash
# Check service endpoints
kubectl get endpoints certificate-monkey -n certificate-monkey

# Check pod readiness
kubectl get pods -n certificate-monkey -o wide

# Test from within cluster
kubectl run test-pod --rm -it --restart=Never --image=curlimages/curl -n certificate-monkey -- curl -v http://certificate-monkey:8080/health
```

#### 4. High Memory Usage

```bash
# Check current resource usage
kubectl top pods -n certificate-monkey

# Increase memory limits
helm upgrade certificate-monkey certificate-monkey/certificate-monkey \
  --namespace=certificate-monkey \
  --values values-production.yaml \
  --set resources.limits.memory=1Gi \
  --set resources.requests.memory=512Mi
```

### Debug Mode

Enable debug logging:

```yaml
extraEnv:
  - name: LOG_LEVEL
    value: debug
```

## Security Best Practices

1. **Use IRSA for AWS Authentication** - Avoid static credentials in production
2. **Rotate API Keys Regularly** - Update secrets periodically
3. **Enable Network Policies** - Restrict pod-to-pod communication
4. **Use Private Container Registry** - Scan images for vulnerabilities
5. **Enable Pod Security Standards** - Enforce security contexts
6. **Encrypt Secrets at Rest** - Use AWS Secrets Manager or sealed-secrets
7. **Implement RBAC** - Restrict access to Kubernetes resources
8. **Monitor Access Logs** - Track API usage and anomalies
9. **Regular Updates** - Keep chart and application versions current
10. **Backup DynamoDB** - Enable point-in-time recovery

## Uninstalling

```bash
# Uninstall the Helm release
helm uninstall certificate-monkey -n certificate-monkey

# Delete secrets
kubectl delete secret certificate-monkey-api-keys -n certificate-monkey
kubectl delete secret certificate-monkey-aws -n certificate-monkey

# Delete namespace
kubectl delete namespace certificate-monkey
```

## Support and Resources

- **Documentation**: https://github.com/dduivenbode/CertificateMonkey
- **Helm Charts**: https://sefiris.github.io/CertificateMonkey/
- **Issues**: https://github.com/dduivenbode/CertificateMonkey/issues
- **Testing Guide**: [HELM_TESTING.md](HELM_TESTING.md)

## Next Steps

- Set up monitoring and alerting
- Configure backup and disaster recovery
- Implement CI/CD pipeline for automated deployments
- Set up multi-region deployment
- Configure API rate limiting
- Implement audit logging
