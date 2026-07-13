# Task 09 — Build the formal AWS SDK compatibility suite

Target agent: Codex with GPT-5.6  
Prerequisite: Tasks 00–08.

Create a dedicated, hermetic compatibility suite proving that real AWS SDK for Go v2 clients work against Emulith.

## Test layout

Create a clear layout equivalent to:

```text
test/compatibility/aws/
  harness/
  s3/
  sqs/
  sts/
```

A different Go-package layout is acceptable if `go test ./...` remains straightforward.

## Harness requirements

The harness must:

- start a full in-process Emulith server using a temporary state directory;
- listen on an OS-assigned loopback port;
- return an explicit base endpoint;
- use deterministic cleanup;
- expose no real network listener;
- fail fast if the resolved endpoint is not loopback;
- use static fake credentials;
- set region `us-east-1`;
- disable EC2 metadata;
- avoid shared AWS config/profile loading;
- never use default provider endpoints as a fallback;
- optionally install a restrictive custom HTTP transport that rejects non-loopback destinations;
- provide unique resource names per test without timestamps that make failures hard to reproduce.

Use the current, non-deprecated endpoint configuration APIs of the pinned AWS SDK v2 packages.

## S3 flow

Using a real S3 client with forced path-style addressing:

1. Create bucket.
2. List buckets and find it.
3. Put a binary-safe object.
4. Head object and verify size/ETag.
5. Get object and compare exact bytes.
6. Put a second nested-key object.
7. List with prefix.
8. Delete first object.
9. Verify `NoSuchKey` through SDK error handling.

## SQS flow

Using the real SQS client and its default protocol:

1. Create queue.
2. Get queue URL.
3. List queues.
4. Send message.
5. Receive message.
6. Verify MD5/body.
7. Get attributes.
8. Delete with receipt handle.
9. Send another message.
10. Purge queue.
11. Verify no visible messages remain.

## STS flow

Using the real STS client:

1. Call `GetCallerIdentity`.
2. Verify account, ARN, and user ID.

## Assertions

Do not assert only status codes. Assert decoded SDK outputs and typed/provider error behavior.

Do not hide compatibility failures with broad retries. A small readiness wait is allowed only for server startup and must be bounded.

## Makefile and docs

Add:

```text
make compatibility
```

It must run the dedicated suite and fail on incompatibility.

Document:

- what the suite proves;
- that it does not compare against real AWS yet;
- how the loopback guard prevents accidental cloud calls;
- how to add a new supported operation.

## CI suitability

Keep runtime low enough for pull requests. Do not require Docker, secrets, internet access after dependencies are available, or a real cloud account.

## Required verification

```bash
gofmt -w <changed-go-files>
go test ./...
go vet ./...
make compatibility
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
