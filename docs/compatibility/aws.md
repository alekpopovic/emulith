# AWS compatibility

Generated from `compatibility/aws.yaml`. Statuses: supported (default SDK test passes), partial (documented subset), experimental (may change), unsupported (not implemented).

| Service | Operation | Status | Protocol | Test ID | Notes | Known deviations | Since |
| --- | --- | --- | --- | --- | --- | --- | --- |
| dynamodb | BatchGetItem | partial | AWS-JSON-1.0 | aws.dynamodb.BatchGetItem.multi-table | Validated multi-table reads with projection and exact key validation. | No throttling; successful requests return empty UnprocessedKeys. | v0.2.0-dev |
| dynamodb | BatchWriteItem | partial | AWS-JSON-1.0 | aws.dynamodb.BatchWriteItem.multi-table | Up to 25 validated multi-table put/delete actions. | Local writes are all-or-nothing in one transaction; no throttling. | v0.2.0-dev |
| dynamodb | CreateTable | supported | AWS-JSON-1.0 | aws.dynamodb.table-lifecycle.basic | PAY_PER_REQUEST tables with scalar HASH and optional RANGE keys. | Tables become ACTIVE immediately; secondary indexes and advanced options are rejected. | v0.2.0-dev |
| dynamodb | DeleteItem | partial | AWS-JSON-1.0 |  | Atomic idempotent deletion with NONE and ALL_OLD. | Conditions are not yet supported. | v0.2.0-dev |
| dynamodb | DeleteTable | partial | AWS-JSON-1.0 |  | Immediate atomic local deletion. | No asynchronous deletion period. | v0.2.0-dev |
| dynamodb | DescribeTable | partial | AWS-JSON-1.0 | aws.dynamodb.persistence.restart | Returns persisted local table metadata and survives clean local restart. | Capacity and index metrics are omitted. | v0.2.0-dev |
| dynamodb | GetItem | partial | AWS-JSON-1.0 |  | Strongly consistent full-item reads. | Projection is not yet supported. | v0.2.0-dev |
| dynamodb | ListTables | partial | AWS-JSON-1.0 |  | Lexical pagination with Limit and ExclusiveStartTableName. | Local tables only. | v0.2.0-dev |
| dynamodb | PutItem | supported | AWS-JSON-1.0 | aws.dynamodb.item-crud.basic | Atomic validated item replacement with NONE and ALL_OLD. | Conditions and legacy parameters are rejected. | v0.2.0-dev |
| dynamodb | Query | partial | AWS-JSON-1.0 | aws.dynamodb.Query.primary-index | Primary-index key conditions filtering projection ordering and key-map pagination. | Secondary indexes are unsupported; pagination is not snapshot isolated. | v0.2.0-dev |
| dynamodb | Scan | partial | AWS-JSON-1.0 | aws.dynamodb.Scan.pagination | Deterministic primary-key scan with pre-filter Limit semantics and projection. | Parallel scan and secondary indexes are unsupported. | v0.2.0-dev |
| dynamodb | UpdateItem | partial | AWS-JSON-1.0 | aws.dynamodb.UpdateItem.concurrent-condition | Bounded AST-based SET REMOVE ADD DELETE updates and atomic one-winner conditions. | Expression subset is documented; ReturnValuesOnConditionCheckFailure is rejected. | v0.2.0-dev |
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
