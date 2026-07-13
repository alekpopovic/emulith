# Changelog

## Unreleased

- Continued POC hardening and documentation.

## 0.2.0 (release preparation)

- Added DynamoDB expressions, Query/Scan, batch operations, and formal SDK coverage.
- Added SNS topics, Publish, and local SNS-to-SQS delivery.
- Added a bounded CloudWatch Logs subset.
- Added forward-only state migrations and local-only release documentation.

## 0.1.0-poc (draft)

- Go CLI and non-root Docker image.
- Persistent SQLite/filesystem state with local reset.
- STS `GetCallerIdentity`, path-style S3 subset, and standard SQS subset.
- Loopback-only AWS SDK for Go v2 compatibility tests and experimental manifests.

Limitations include development/CI-only use, no IAM enforcement, and no claim of full AWS parity.
