# Task 05 — Implement the S3 POC subset

Target agent: Codex with GPT-5.6  
Prerequisite: Tasks 00–04.

Implement a useful path-style S3 subset backed by the Emulith state store.

## Supported operations

Implement:

1. `CreateBucket`
2. `ListBuckets`
3. `PutObject`
4. `GetObject`
5. `HeadObject`
6. `DeleteObject`
7. `ListObjectsV2`

The first POC supports path-style addressing only.

## Operation behavior

### CreateBucket

```http
PUT /{bucket}
```

- Validate a practical general-purpose S3 bucket-name subset.
- Persist bucket region and creation time.
- Accept an absent location constraint as `us-east-1`.
- Parse a simple create-bucket location constraint when present.
- Return a compatible success response and `Location` header.
- A conflicting existing bucket must return a provider-shaped error; deterministic local idempotency is acceptable only if documented and tested.

### ListBuckets

```http
GET /
```

Return XML containing owner metadata and buckets ordered deterministically by name or creation time. Document the chosen ordering.

### PutObject

```http
PUT /{bucket}/{key...}
```

- Require an existing bucket.
- Stream the request body to a temporary file.
- Compute size and the simple single-part ETag as lowercase MD5 hex wrapped as S3 expects.
- Preserve the logical key exactly in metadata.
- Persist selected metadata:
  - content type;
  - last modified;
  - ETag;
  - size.
- Atomically replace an existing key.
- Clean up files if any stage fails.
- Handle request bodies produced by the pinned AWS SDK for Go v2. If the SDK uses checksum/trailer or `aws-chunked` framing for this operation, support the minimal correct decoding/validation needed for the default SDK client rather than bypassing the SDK test.

### GetObject

- Return exact stored bytes.
- Set `ETag`, `Content-Length`, `Last-Modified`, `Content-Type`, request ID, and binary-safe body handling.
- Full range requests are out of scope; return a clear provider-shaped unsupported response if a range is supplied.

### HeadObject

Return the same relevant metadata as `GetObject` with no response body.

### DeleteObject

Match S3's practical idempotent behavior for deleting a missing key. Remove metadata and the managed body safely.

### ListObjectsV2

```http
GET /{bucket}?list-type=2
```

Support:

- `prefix`;
- `max-keys` within a safe range;
- deterministic lexical key ordering;
- `KeyCount`;
- `IsTruncated=false` for the initial POC when all matching rows fit.

Do not falsely implement continuation tokens. If result pagination would be required, either implement a deterministic opaque continuation token or explicitly constrain/test the POC so it never claims unsupported pagination behavior.

Correctly XML-escape keys and bucket names.

## Errors

Implement provider-shaped XML errors for at least:

- `NoSuchBucket`;
- `NoSuchKey`;
- `BucketAlreadyExists` or the documented local equivalent;
- `InvalidBucketName`;
- `InvalidArgument`;
- `NotImplemented`;
- internal storage failure without leaking filesystem paths or SQL.

## Explicitly unsupported

- virtual-hosted-style addressing;
- multipart upload;
- ACLs and policies;
- versioning;
- presigned URL policy validation beyond accepting signed HTTP requests;
- server-side encryption;
- website hosting;
- range requests unless implemented completely.

## Tests

### Handler/state tests

Cover:

- bucket validation;
- create/list bucket;
- put/get/head/delete;
- overwrite cleans the old body;
- missing bucket/key;
- zero-byte object;
- binary object;
- nested key;
- key containing spaces, Unicode, `..`, percent characters, and XML-sensitive characters;
- no filesystem escape;
- list prefix and max keys;
- deterministic XML and ordering;
- rollback/cleanup on simulated metadata failure if the architecture permits fault injection.

### AWS SDK for Go v2 compatibility

Using an in-process Emulith server and a real S3 client with explicit loopback endpoint, fake static credentials, region `us-east-1`, and forced path style:

- CreateBucket;
- ListBuckets;
- PutObject;
- HeadObject;
- GetObject and verify exact bytes;
- ListObjectsV2;
- DeleteObject;
- verify the deleted object is missing.

The test must not call real AWS and must not read a user profile or metadata service.

## Documentation

Update README and detailed compatibility docs with exact supported and unsupported behavior. Do not say “S3 compatible” without the word “subset” or an operation matrix.

## Required verification

```bash
gofmt -w <changed-go-files>
go test ./...
go vet ./...
make build
```

## Execution contract

You are the implementation agent for this task. Complete the task in the current repository; do not stop after producing a plan.

1. Read every applicable `AGENTS.md` before changing files.
2. Inspect the repository, existing conventions, current tests, and dependency versions.
3. Briefly state the implementation plan, then execute it immediately.
4. Make reasonable assumptions when details are non-blocking. Do not ask for confirmation merely to choose between equivalent implementation details.
5. Keep the change tightly scoped to this task. Do not implement later roadmap items.
6. Preserve working behavior and public interfaces unless this task explicitly changes them.
7. Prefer simple, readable Go over speculative abstractions.
8. Do not use LocalStack, Moto, MinIO, ElasticMQ, Azurite, or another cloud emulator as a runtime dependency.
9. Never contact real AWS, Azure, or GCP endpoints. Tests must be hermetic.
10. Do not add accounts, license keys, forced telemetry, analytics, or phone-home behavior.
11. Do not commit, push, create a tag, or open a pull request.
12. Format all changed Go files and run the required verification commands.
13. Fix failures caused by your changes before finishing. If an environment limitation prevents a command, report the exact limitation and run the closest safe verification.
14. Finish with:
    - implementation summary;
    - important design decisions;
    - changed files;
    - commands run and their results;
    - remaining limitations that are genuinely outside this task.
