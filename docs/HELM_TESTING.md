# Helm Chart Testing Guide

This guide provides comprehensive instructions for testing the Certificate Monkey Helm chart in a local minikube environment.

## Prerequisites

Before starting, ensure you have the following tools installed:

- **minikube** (v1.25.0+): Local Kubernetes cluster
- **kubectl** (v1.24.0+): Kubernetes CLI
- **helm** (v3.8.0+): Kubernetes package manager
- **AWS CLI** (v2.0+): For AWS credentials configuration
- **AWS Account**: With DynamoDB and KMS resources configured

### Installation Commands

```bash
# macOS
brew install minikube kubectl helm awscli

# Linux
# Follow official installation guides:
# - minikube: https://minikube.sigs.k8s.io/docs/start/
# - kubectl: https://kubernetes.io/docs/tasks/tools/
# - helm: https://helm.sh/docs/intro/install/
```

## AWS Infrastructure Setup

Before deploying the application, ensure your AWS infrastructure is ready:

### 1. Create DynamoDB Table

```bash
aws dynamodb create-table \
    --table-name certificate-monkey-dev \
    --attribute-definitions \
        AttributeName=id,AttributeType=S \
        AttributeName=created_at,AttributeType=S \
    --key-schema \
        AttributeName=id,KeyType=HASH \
    --global-secondary-indexes \
        'IndexName=created_at-index,KeySchema=[{AttributeName=created_at,KeyType=HASH}],Projection={ProjectionType=ALL},BillingMode=PAY_PER_REQUEST' \
    --billing-mode PAY_PER_REQUEST \
    --region us-east-1
```

### 2. Create KMS Key

```bash
# Create KMS key
KMS_KEY_ID=$(aws kms create-key \
    --description "Certificate Monkey dev encryption key" \
    --region us-east-1 \
    --query 'KeyMetadata.KeyId' \
    --output text)

# Create alias
aws kms create-alias \
    --alias-name alias/certificate-monkey-dev \
    --target-key-id $KMS_KEY_ID \
    --region us-east-1

echo "KMS Key ID: $KMS_KEY_ID"
echo "KMS Alias: alias/certificate-monkey-dev"
```

### 3. Verify AWS Credentials

```bash
# Check AWS credentials are configured
aws sts get-caller-identity

# Verify permissions
aws dynamodb describe-table --table-name certificate-monkey-dev --region us-east-1
aws kms describe-key --key-id alias/certificate-monkey-dev --region us-east-1
```

## Minikube Setup

### 1. Start Minikube

```bash
# Start minikube with sufficient resources
minikube start --cpus=4 --memory=4096 --driver=docker

# Verify cluster is running
kubectl cluster-info
kubectl get nodes
```

### 2. Enable Required Addons (Optional)

```bash
# Enable metrics server for HPA testing
minikube addons enable metrics-server

# Enable ingress for ingress testing
minikube addons enable ingress
```

## Kubernetes Secrets Setup

### 1. Create AWS Credentials Secret

For minikube testing, we use static AWS credentials (production should use IRSA):

```bash
# Create namespace (optional, or use default)
kubectl create namespace certificate-monkey

# Create AWS credentials secret
kubectl create secret generic certificate-monkey-aws \
  --namespace=certificate-monkey \
  --from-literal=AWS_ACCESS_KEY_ID="$(aws configure get aws_access_key_id)" \
  --from-literal=AWS_SECRET_ACCESS_KEY="$(aws configure get aws_secret_access_key)"

# Verify secret
kubectl get secret certificate-monkey-aws -n certificate-monkey
```

**⚠️ Security Note**: Never commit AWS credentials to version control. These are for local testing only.

### 2. Create API Keys Secret

```bash
# Generate secure API keys or use test keys
kubectl create secret generic certificate-monkey-api-keys \
  --namespace=certificate-monkey \
  --from-literal=API_KEY_1="cm_minikube_test_$(openssl rand -hex 12)" \
  --from-literal=API_KEY_2="cm_minikube_backup_$(openssl rand -hex 12)"

# Save API keys for testing
export API_KEY=$(kubectl get secret certificate-monkey-api-keys -n certificate-monkey -o jsonpath='{.data.API_KEY_1}' | base64 --decode)
echo "API Key: $API_KEY"
```

## Helm Chart Installation

### 1. Add Local Helm Repository (Optional)

If testing from local repository:

```bash
# From the project root directory
cd /Users/dduivenbode/Documents/project_buckaroo/repos/personal_projects/CertificateMonkey
```

### 2. Lint the Chart

```bash
# Validate chart syntax
helm lint helm/certificate-monkey

# Validate chart with minikube values
helm lint helm/certificate-monkey -f helm/certificate-monkey/values-minikube.yaml
```

### 3. Dry Run Installation

```bash
# Test template rendering without installing
helm install certificate-monkey-test helm/certificate-monkey \
  --namespace=certificate-monkey \
  --create-namespace \
  --values helm/certificate-monkey/values-minikube.yaml \
  --dry-run --debug
```

### 4. Install the Chart

```bash
# Install with minikube configuration
helm install certificate-monkey helm/certificate-monkey \
  --namespace=certificate-monkey \
  --create-namespace \
  --values helm/certificate-monkey/values-minikube.yaml

# Watch deployment progress
kubectl get pods -n certificate-monkey -w
```

Expected output:
```
NAME                                   READY   STATUS    RESTARTS   AGE
certificate-monkey-7d8f6b4c9d-abcde    1/1     Running   0          30s
certificate-monkey-7d8f6b4c9d-fghij    1/1     Running   0          30s
```

## Verification and Testing

### 1. Check Deployment Status

```bash
# Check all resources
kubectl get all -n certificate-monkey

# Check deployment details
kubectl describe deployment certificate-monkey -n certificate-monkey

# Check pod logs
kubectl logs -l app.kubernetes.io/name=certificate-monkey -n certificate-monkey --tail=50
```

### 2. Verify Health Endpoints

**Recommended: Use port-forward** (more reliable than NodePort in minikube):

```bash
# Start port forwarding in the background
kubectl port-forward -n certificate-monkey svc/certificate-monkey 8080:8080 &

# Wait for port-forward to initialize
sleep 2

# Test basic health endpoint
curl http://localhost:8080/health
# Expected: {"service":"certificate-monkey","status":"healthy","version":"0.1.0"}

# Test AWS connectivity health check (DynamoDB + KMS)
curl http://localhost:8080/health/aws | jq .
# Expected: Detailed status for DynamoDB and KMS connectivity

# Test build info endpoint
curl http://localhost:8080/build-info | jq .
```

**Alternative: Using NodePort** (may have connectivity issues on some systems):

```bash
# Get the NodePort service URL
export NODE_PORT=$(kubectl get svc certificate-monkey -n certificate-monkey -o jsonpath='{.spec.ports[0].nodePort}')
export NODE_IP=$(minikube ip)
export SERVICE_URL="http://$NODE_IP:$NODE_PORT"

echo "Service URL: $SERVICE_URL"

# Test health endpoint (may timeout on some minikube configurations)
curl $SERVICE_URL/health

# If NodePort doesn't work, use port-forward method above
```

### 3. Test API Functionality

**Note:** Ensure port-forward is running (see step 2 above). If not:
```bash
kubectl port-forward -n certificate-monkey svc/certificate-monkey 8080:8080 &
sleep 2
```

```bash
# Get API key from secret
export API_KEY=$(kubectl get secret certificate-monkey-api-keys -n certificate-monkey -o jsonpath='{.data.API_KEY_1}' | base64 --decode)
echo "Using API Key: $API_KEY"

# Create a test certificate
curl -X POST http://localhost:8080/api/v1/keys \
  -H "Content-Type: application/json" \
  -H "X-API-Key: $API_KEY" \
  -d '{
    "common_name": "minikube-test.example.com",
    "key_type": "RSA2048",
    "tags": {
      "environment": "minikube",
      "test": "true"
    }
  }' | jq .

# List certificates
curl -H "X-API-Key: $API_KEY" \
  http://localhost:8080/api/v1/keys | jq .
```

### 4. Verify AWS Integration

**Test AWS Connectivity Endpoint:**

```bash
# Ensure port-forward is running
kubectl port-forward -n certificate-monkey svc/certificate-monkey 8080:8080 &
sleep 2

# Test AWS health endpoint - verifies DynamoDB and KMS connectivity
curl http://localhost:8080/health/aws | jq .

# Expected successful response:
# {
#   "status": "healthy",
#   "service": "certificate-monkey",
#   "version": "0.1.0",
#   "timestamp": "2025-11-09T17:30:00Z",
#   "checks": {
#     "dynamodb": {
#       "status": "healthy",
#       "message": "DynamoDB table is accessible",
#       "response_ms": 45
#     },
#     "kms": {
#       "status": "healthy",
#       "message": "KMS key is accessible",
#       "response_ms": 32
#     }
#   }
# }
```

**Verify Data in DynamoDB:**

```bash
# Check if data is written to DynamoDB
aws dynamodb scan \
  --table-name certificate-monkey-dev \
  --region eu-central-1 \
  --max-items 5

# Check pod environment variables
kubectl exec -n certificate-monkey \
  $(kubectl get pod -n certificate-monkey -l app.kubernetes.io/name=certificate-monkey -o jsonpath="{.items[0].metadata.name}") \
  -- env | grep -E "AWS_|DYNAMODB|KMS"
```

### 5. Test Multi-Replica Behavior

```bash
# Check both replicas are running
kubectl get pods -n certificate-monkey -l app.kubernetes.io/name=certificate-monkey

# Test load balancing by making multiple requests
for i in {1..10}; do
  echo "Request $i:"
  curl -s $SERVICE_URL/build-info | jq -r .timestamp
  sleep 0.5
done

# Check logs from different pods
kubectl logs -n certificate-monkey -l app.kubernetes.io/name=certificate-monkey --tail=20 --prefix=true
```

## Testing Scenarios

### Scenario 1: Scale Replicas

```bash
# Scale up to 3 replicas
helm upgrade certificate-monkey helm/certificate-monkey \
  --namespace=certificate-monkey \
  --values helm/certificate-monkey/values-minikube.yaml \
  --set replicaCount=3

# Watch scaling
kubectl get pods -n certificate-monkey -w

# Scale back down
helm upgrade certificate-monkey helm/certificate-monkey \
  --namespace=certificate-monkey \
  --values helm/certificate-monkey/values-minikube.yaml \
  --set replicaCount=2
```

### Scenario 2: Rolling Update

```bash
# Trigger a rolling update by changing an annotation
helm upgrade certificate-monkey helm/certificate-monkey \
  --namespace=certificate-monkey \
  --values helm/certificate-monkey/values-minikube.yaml \
  --set podAnnotations.test="rolling-update-$(date +%s)"

# Watch the rolling update
kubectl rollout status deployment/certificate-monkey -n certificate-monkey
```

### Scenario 3: Resource Limits Testing

```bash
# Update with different resource limits
helm upgrade certificate-monkey helm/certificate-monkey \
  --namespace=certificate-monkey \
  --values helm/certificate-monkey/values-minikube.yaml \
  --set resources.limits.memory=512Mi \
  --set resources.requests.memory=256Mi

# Verify new limits
kubectl describe pod -n certificate-monkey -l app.kubernetes.io/name=certificate-monkey | grep -A 5 "Limits\|Requests"
```

### Scenario 4: Enable Horizontal Pod Autoscaler

```bash
# Enable HPA
helm upgrade certificate-monkey helm/certificate-monkey \
  --namespace=certificate-monkey \
  --values helm/certificate-monkey/values-minikube.yaml \
  --set autoscaling.enabled=true \
  --set autoscaling.minReplicas=2 \
  --set autoscaling.maxReplicas=5 \
  --set autoscaling.targetCPUUtilizationPercentage=70

# Check HPA status
kubectl get hpa -n certificate-monkey

# Generate load to test autoscaling
kubectl run -n certificate-monkey load-generator --rm -it --restart=Never --image=busybox -- /bin/sh -c "while true; do wget -q -O- http://certificate-monkey:8080/health; done"

# Watch HPA scale up (in another terminal)
kubectl get hpa -n certificate-monkey -w
```

## Troubleshooting

### Pod Not Starting

```bash
# Check pod status and events
kubectl describe pod -n certificate-monkey -l app.kubernetes.io/name=certificate-monkey

# Check logs
kubectl logs -n certificate-monkey -l app.kubernetes.io/name=certificate-monkey --tail=100

# Common issues:
# 1. Secret not found - verify secrets exist
kubectl get secrets -n certificate-monkey

# 2. AWS credentials invalid - verify credentials
kubectl exec -n certificate-monkey \
  $(kubectl get pod -n certificate-monkey -l app.kubernetes.io/name=certificate-monkey -o jsonpath="{.items[0].metadata.name}") \
  -- env | grep AWS
```

### AWS Connection Failures

```bash
# Test AWS connectivity from pod
kubectl exec -n certificate-monkey \
  $(kubectl get pod -n certificate-monkey -l app.kubernetes.io/name=certificate-monkey -o jsonpath="{.items[0].metadata.name}") \
  -- sh -c 'apk add curl && curl -I https://dynamodb.us-east-1.amazonaws.com'

# Verify IAM permissions
aws sts get-caller-identity
aws dynamodb describe-table --table-name certificate-monkey-dev --region us-east-1
```

### Image Pull Errors

```bash
# Check image pull secrets
kubectl get pods -n certificate-monkey -o jsonpath='{.items[*].spec.imagePullSecrets}'

# Manually pull image to verify it's accessible
minikube ssh
docker pull ghcr.io/dduivenbode/certificatemonkey:0.1.0
```

### Service Not Accessible

**If `curl` hangs or times out:**

This is a common issue with minikube NodePort networking. **Solution: Use port-forward instead:**

```bash
# Stop any existing port-forwards
pkill -f "port-forward.*certificate-monkey"

# Start fresh port-forward
kubectl port-forward -n certificate-monkey svc/certificate-monkey 8080:8080 &
sleep 2

# Test connection
curl http://localhost:8080/health
```

**Verify service and endpoints:**

```bash
# Check service endpoints
kubectl get endpoints -n certificate-monkey

# Check if pods are ready
kubectl get pods -n certificate-monkey

# Test from within cluster (this should always work)
kubectl run -n certificate-monkey test-curl --rm -it --restart=Never --image=curlimages/curl -- curl -v http://certificate-monkey:8080/health
```

**If port-forward doesn't work:**
- Ensure minikube is running: `minikube status`
- Check if port 8080 is already in use: `lsof -i :8080`
- Try a different local port: `kubectl port-forward -n certificate-monkey svc/certificate-monkey 9090:8080`

## Cleanup

### Uninstall the Chart

```bash
# Uninstall the release
helm uninstall certificate-monkey -n certificate-monkey

# Verify all resources are removed
kubectl get all -n certificate-monkey

# Delete secrets (optional)
kubectl delete secret certificate-monkey-aws -n certificate-monkey
kubectl delete secret certificate-monkey-api-keys -n certificate-monkey

# Delete namespace (optional)
kubectl delete namespace certificate-monkey
```

### Stop Minikube

```bash
# Stop minikube
minikube stop

# Delete minikube cluster (complete cleanup)
minikube delete
```

### Cleanup AWS Resources

```bash
# Delete DynamoDB table (WARNING: deletes all data)
aws dynamodb delete-table \
  --table-name certificate-monkey-dev \
  --region us-east-1

# Schedule KMS key deletion (WARNING: irreversible after waiting period)
aws kms schedule-key-deletion \
  --key-id alias/certificate-monkey-dev \
  --pending-window-in-days 7 \
  --region us-east-1
```

## Quick Reference Commands

```bash
# Makefile shortcuts (from project root)
make helm-lint                # Lint Helm chart
make helm-template            # Test template rendering
make helm-package             # Package chart locally
make helm-test-minikube       # Run complete minikube test

# Helm operations
helm list -n certificate-monkey                    # List releases
helm status certificate-monkey -n certificate-monkey  # Check status
helm get values certificate-monkey -n certificate-monkey  # Show values
helm history certificate-monkey -n certificate-monkey # Show history
helm rollback certificate-monkey -n certificate-monkey  # Rollback

# Kubernetes operations
kubectl get all -n certificate-monkey             # View all resources
kubectl logs -f -l app.kubernetes.io/name=certificate-monkey -n certificate-monkey  # Stream logs
kubectl port-forward -n certificate-monkey svc/certificate-monkey 8080:8080  # Port forward
kubectl exec -it -n certificate-monkey <pod-name> -- sh  # Shell into pod
```

## Next Steps

After successful testing in minikube:

1. Review the [Helm Deployment Guide](HELM_DEPLOYMENT.md) for production deployment
2. Configure IRSA for EKS deployments
3. Set up monitoring and alerting
4. Configure ingress with TLS certificates
5. Implement backup and disaster recovery procedures

## Support

For issues and questions:
- GitHub Issues: https://github.com/dduivenbode/CertificateMonkey/issues
- Documentation: https://github.com/dduivenbode/CertificateMonkey/blob/main/README.md
