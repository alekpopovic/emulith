# AWS POC compatibility

Emulith is development/CI-only and does not claim full AWS parity.

| Service | Operation | Status | Notes |
| --- | --- | --- | --- |
| STS | GetCallerIdentity | Supported | Deterministic local identity; no IAM enforcement. |
| S3 | CreateBucket | Supported | Path-style only. |
| S3 | ListBuckets | Supported | Lexical bucket ordering. |
| S3 | PutObject | Supported | Single-part body and MD5 ETag. |
| S3 | GetObject | Supported | Full object only; ranges unsupported. |
| S3 | HeadObject | Supported | Stored metadata subset. |
| S3 | DeleteObject | Supported | Idempotent for missing keys. |
| S3 | ListObjectsV2 | Partial | Prefix and max-keys up to 1000; no continuation tokens. |
| SQS | CreateQueue | Supported | Standard queues only. |
| SQS | GetQueueUrl | Supported | Local account `000000000000`. |
| SQS | ListQueues | Supported | Prefix and lexical ordering. |
| SQS | SendMessage | Supported | Text body up to 256 KiB. |
| SQS | ReceiveMessage | Partial | Short polling only, 1–10 messages. |
| SQS | DeleteMessage | Supported | Current receipt handle required. |
| SQS | PurgeQueue | Supported | Repeated purge allowed locally. |
| SQS | GetQueueAttributes | Partial | Documented POC attributes only. |

SQS AWS JSON is primary and Query is a compatibility fallback. Metadata persists in SQLite and S3 bodies on the filesystem. Multipart, ACL/policy, versioning, FIFO, batch, redrive, long polling, IAM, and production mode are unsupported.
