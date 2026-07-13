# Upgrade guide: v0.1 to v0.2

Emulith v0.2 opens existing v0.1 SQLite state and applies migrations automatically. Existing S3 buckets/objects and SQS queues/messages remain readable; DynamoDB, SNS, and CloudWatch Logs tables are added without changing those records. Migrations are forward-only and idempotent. Back up the data directory or export a snapshot before upgrading; downgrade is unsupported.

The v0.2 subset includes STS, S3, SQS, DynamoDB, SNS/SQS delivery, and CloudWatch Logs. It remains local development/CI software only: IAM, throttling, production security, and full AWS parity are not provided.
