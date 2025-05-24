"""
Certificate Monkey Infrastructure

This Pulumi program creates the AWS infrastructure required for Certificate Monkey:
- DynamoDB table with proper schema and GSI
- KMS key for encrypting private keys
- IAM policy (optional, for reference)
"""

import pulumi
import pulumi_aws as aws

# Get configuration values
config = pulumi.Config()
environment = config.get("environment", "dev")
table_name = config.get("table_name", f"certificate-monkey-{environment}")

# Create KMS key for encrypting private keys
kms_key = aws.kms.Key(
    "certificate-monkey-kms-key",
    description="Certificate Monkey private key encryption",
    deletion_window_in_days=7,  # Allow recovery if accidentally deleted
    tags={
        "Name": f"certificate-monkey-{environment}",
        "Environment": environment,
        "Application": "certificate-monkey",
        "Purpose": "private-key-encryption"
    }
)

# Create KMS key alias for easier reference
kms_alias = aws.kms.Alias(
    "certificate-monkey-kms-alias",
    name=f"alias/certificate-monkey-{environment}",
    target_key_id=kms_key.key_id
)

# Create DynamoDB table
dynamodb_table = aws.dynamodb.Table(
    "certificate-monkey-table",
    name=table_name,
    billing_mode="PAY_PER_REQUEST",  # On-demand pricing
    hash_key="id",
    attributes=[
        aws.dynamodb.TableAttributeArgs(
            name="id",
            type="S"  # String
        ),
        aws.dynamodb.TableAttributeArgs(
            name="created_at",
            type="S"  # String (ISO 8601 timestamp)
        )
    ],
    # Global Secondary Index for date-based queries
    global_secondary_indexes=[
        aws.dynamodb.TableGlobalSecondaryIndexArgs(
            name="created_at-index",
            hash_key="created_at",
            projection_type="ALL",  # Include all attributes
        )
    ],
    # Enable server-side encryption with KMS
    server_side_encryption=aws.dynamodb.TableServerSideEncryptionArgs(
        enabled=True,
        kms_key_arn=kms_key.arn
    ),
    # Enable point-in-time recovery
    point_in_time_recovery=aws.dynamodb.TablePointInTimeRecoveryArgs(
        enabled=True
    ),
    tags={
        "Name": table_name,
        "Environment": environment,
        "Application": "certificate-monkey",
        "Purpose": "certificate-storage"
    },
    # deletion_protection_enabled=True # TODO: enable in prod
)

# Create IAM policy for the application (for reference)
app_policy_document = aws.iam.get_policy_document(
    statements=[
        # DynamoDB permissions
        aws.iam.GetPolicyDocumentStatementArgs(
            effect="Allow",
            actions=[
                "dynamodb:PutItem",
                "dynamodb:GetItem",
                "dynamodb:UpdateItem",
                "dynamodb:DeleteItem",
                "dynamodb:Scan"
            ],
            resources=[
                dynamodb_table.arn,
                pulumi.Output.concat(dynamodb_table.arn, "/index/*")
            ]
        ),
        # KMS permissions
        aws.iam.GetPolicyDocumentStatementArgs(
            effect="Allow",
            actions=[
                "kms:Encrypt",
                "kms:Decrypt"
            ],
            resources=[kms_key.arn],
            conditions=[
                aws.iam.GetPolicyDocumentStatementConditionArgs(
                    test="StringEquals",
                    variable="kms:ViaService",
                    values=[pulumi.Output.concat("dynamodb.", aws.get_region().name, ".amazonaws.com")]
                )
            ]
        )
    ]
)

# Create the IAM policy (optional - for documentation/reference)
app_policy = aws.iam.Policy(
    "certificate-monkey-app-policy",
    name=f"certificate-monkey-{environment}-policy",
    description="IAM policy for Certificate Monkey application",
    policy=app_policy_document.json,
    tags={
        "Name": f"certificate-monkey-{environment}-policy",
        "Environment": environment,
        "Application": "certificate-monkey"
    }
)

# Outputs for easy reference
pulumi.export("dynamodb_table_name", dynamodb_table.name)
pulumi.export("dynamodb_table_arn", dynamodb_table.arn)
pulumi.export("kms_key_id", kms_key.key_id)
pulumi.export("kms_key_arn", kms_key.arn)
pulumi.export("kms_alias_name", kms_alias.name)
pulumi.export("iam_policy_arn", app_policy.arn)

# Environment variables for the application
pulumi.export("environment_variables", {
    "DYNAMODB_TABLE": dynamodb_table.name,
    "KMS_KEY_ID": kms_alias.name,
    "AWS_REGION": aws.get_region().name
})
