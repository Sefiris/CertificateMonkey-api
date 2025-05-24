# Certificate Monkey Infrastructure

This Pulumi Python project deploys the AWS infrastructure required for the Certificate Monkey API.

## Resources Created

- **DynamoDB Table**: Stores certificate entities with proper schema and GSI
- **KMS Key**: Encrypts private keys at rest
- **IAM Policy**: Reference policy for application permissions (optional)

## Prerequisites

- [Pulumi CLI](https://www.pulumi.com/docs/get-started/install/) installed
- AWS CLI configured with appropriate credentials
- Python 3.8+ with pip

## Quick Start

### 1. Install Dependencies

```bash
cd infrastructure
pip install -r requirements.txt
```

### 2. Initialize Pulumi Stack

```bash
# Initialize the stack (local backend)
pulumi stack init dev --secrets-provider passphrase

# Or use Pulumi Cloud (requires account)
pulumi stack init dev
```

### 3. Configure AWS Region

```bash
pulumi config set aws:region us-east-1
```

### 4. Deploy Infrastructure

```bash
pulumi up
```

### 5. Get Output Values

```bash
# Get all outputs
pulumi stack output

# Get specific values
pulumi stack output dynamodb_table_name
pulumi stack output kms_alias_name
```

## Configuration Options

| Config Key | Default | Description |
|------------|---------|-------------|
| `environment` | `dev` | Environment name (dev, staging, prod) |
| `table_name` | `certificate-monkey-{environment}` | DynamoDB table name |
| `aws:region` | `us-east-1` | AWS region for resources |

### Example Configurations

**Development:**
```bash
pulumi config set environment dev
pulumi config set table_name certificate-monkey-dev
```

**Production:**
```bash
pulumi config set environment prod
pulumi config set table_name certificate-monkey-prod
```

## Environment-Specific Deployments

### Development Environment
```bash
pulumi stack init dev
pulumi config set aws:region us-east-1
pulumi config set environment dev
pulumi up
```

### Production Environment
```bash
pulumi stack init prod
pulumi config set aws:region us-east-1
pulumi config set environment prod
pulumi config set table_name certificate-monkey-prod
pulumi up
```

## Outputs

After deployment, the following outputs are available:

- `dynamodb_table_name`: Name of the created DynamoDB table
- `dynamodb_table_arn`: ARN of the DynamoDB table
- `kms_key_id`: KMS key ID for encryption
- `kms_key_arn`: KMS key ARN
- `kms_alias_name`: KMS key alias (use this in application config)
- `iam_policy_arn`: ARN of the IAM policy for application
- `environment_variables`: Ready-to-use environment variables for the app

## Application Configuration

After deployment, configure your Certificate Monkey application with these environment variables:

```bash
# Get the values from Pulumi outputs
export DYNAMODB_TABLE=$(pulumi stack output dynamodb_table_name)
export KMS_KEY_ID=$(pulumi stack output kms_alias_name)
export AWS_REGION=$(pulumi stack output aws:region || echo "us-east-1")

# Add your API keys
export API_KEY_1=your_secure_api_key_here
export API_KEY_2=your_backup_api_key_here

# Run the application
cd ../
go run cmd/server/main.go
```

## Infrastructure as Code Benefits

- **Version Control**: All infrastructure changes are tracked
- **Reproducible**: Deploy identical environments easily
- **Environment Separation**: Different configs for dev/staging/prod
- **State Management**: Pulumi tracks resource state automatically
- **Rollback**: Easy to revert changes if needed

## Security Features

- **KMS Encryption**: All private keys encrypted with customer-managed keys
- **IAM Least Privilege**: Minimal permissions for application
- **Point-in-time Recovery**: DynamoDB backups enabled
- **Server-side Encryption**: DynamoDB encrypted at rest

## Managing Multiple Environments

```bash
# List stacks
pulumi stack ls

# Switch between environments
pulumi stack select dev
pulumi stack select prod

# Deploy to specific environment
pulumi stack select prod
pulumi up
```

## Cleanup

To destroy the infrastructure:

```bash
pulumi destroy
```

**Note**: This will delete all data in DynamoDB. Make sure you have backups!

## Troubleshooting

### Common Issues

1. **AWS Credentials**: Ensure AWS CLI is configured
   ```bash
   aws configure
   ```

2. **Permissions**: Your AWS user/role needs permissions to create:
   - DynamoDB tables
   - KMS keys
   - IAM policies

3. **Region Mismatch**: Ensure Pulumi and AWS CLI use same region
   ```bash
   pulumi config get aws:region
   aws configure get region
   ```

### Getting Help

```bash
# View stack configuration
pulumi config

# View current stack
pulumi stack

# Preview changes without applying
pulumi preview
```
