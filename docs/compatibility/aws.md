# AWS compatibility

Generated from `compatibility/aws.yaml`. Statuses: supported (default SDK test passes), partial (documented subset), experimental (may change), unsupported (not implemented).

| Service | Operation | Status | Protocol | Test ID | Notes | Known deviations | Since |
| --- | --- | --- | --- | --- | --- | --- | --- |
| dynamodb | CreateTable | experimental | AWS-JSON-1.0 |  | Protocol is recognized but operations are not implemented yet. | Returns UnknownOperationException. | v0.2.0-dev |
| s3 | CreateBucket | supported | REST-XML | aws.s3.lifecycle.basic | Path-style local bucket lifecycle. | No virtual-host addressing. | v0.1.0-poc |
| s3 | DeleteObject | partial | REST-XML |  | Idempotent local deletion. | No version markers. | v0.1.0-poc |
| s3 | GetObject | partial | REST-XML |  | Basic full-body reads. | No ranges or versioning. | v0.1.0-poc |
| s3 | HeadObject | partial | REST-XML |  | Basic object metadata. | Limited metadata. | v0.1.0-poc |
| s3 | ListBuckets | partial | REST-XML |  | Basic listing is covered by the S3 lifecycle test. | Owner details are local placeholders. | v0.1.0-poc |
| s3 | ListObjectsV2 | partial | REST-XML |  | Prefix listing. | Limited pagination semantics. | v0.1.0-poc |
| s3 | PutObject | partial | REST-XML |  | Basic streamed object writes. | No multipart or checksums. | v0.1.0-poc |
| sqs | CreateQueue | supported | AWS-JSON-1.0 | aws.sqs.lifecycle.basic | Standard queue lifecycle and messages. | FIFO unsupported. | v0.1.0-poc |
| sqs | ReceiveMessage | partial | AWS-JSON-1.0 |  | Visibility-based receive. | No long polling. | v0.1.0-poc |
| sqs | SendMessage | partial | AWS-JSON-1.0 |  | Standard message send. | Attributes are limited. | v0.1.0-poc |
| sts | GetCallerIdentity | supported | Query | aws.sts.GetCallerIdentity.basic | Deterministic local identity. | No IAM evaluation. | v0.1.0-poc |
