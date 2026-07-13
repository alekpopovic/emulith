# Task 13 — Harden Emulith before the first real POC

Target agent: Codex with GPT-5.6  
Prerequisite: Tasks 00–12.

Perform a focused audit and hardening pass. Fix defects; do not add large roadmap features.

## Required workflow

1. Read the repository and all tests.
2. Run the full baseline suite before edits.
3. Produce a concise prioritized findings list grouped as:
   - critical;
   - high;
   - medium;
   - low.
4. Immediately implement all critical/high fixes that fit the POC scope.
5. Implement medium fixes that materially improve correctness, security, or usability.
6. Do not spend the task on cosmetic refactors.
7. Re-run the complete suite, including race and compatibility tests.

Do not stop after the audit.

## Audit areas

### Filesystem and reset safety

Verify and fix:

- object keys cannot traverse;
- percent/Unicode edge cases cannot escape storage;
- symlink attacks cannot make writes/deletes/reset leave the data root;
- temp files are cleaned;
- overwrite/delete/reset are atomic enough for the documented POC;
- empty/root/home data directories are rejected safely;
- file permissions are reasonable;
- internal paths never appear in provider errors.

### SQLite and concurrency

Verify and fix:

- foreign keys and busy timeout are active on every relevant connection;
- transactions prevent duplicate receive of one SQS message;
- concurrent object overwrite/delete does not orphan state silently;
- reset cannot race into corrupt metadata;
- rows/files/bodies are closed;
- contexts and cancellation are honored;
- server shutdown closes the store exactly once;
- race detector passes.

### HTTP robustness

Verify and fix:

- server timeouts;
- maximum request/body/form/JSON sizes;
- malformed XML/JSON/form handling;
- content-type parsing;
- unsupported methods;
- request IDs on success and errors;
- no credential/body logging;
- admin routes cannot be confused with bucket names;
- forwarded headers are not blindly trusted;
- panic recovery returns a safe error and logs request ID without secrets.

### S3 behavior

Verify and fix:

- exact bytes and binary safety;
- ETag calculation for simple uploads;
- zero-length objects;
- overwrite cleanup;
- HEAD has no body;
- delete idempotency;
- XML namespaces/escaping;
- UTF-8/object-key edge cases;
- ListObjectsV2 ordering, prefix, max keys, and truncation claims;
- SDK default checksum/chunked behavior;
- typed SDK errors for missing bucket/key.

Do not add multipart upload, policies, ACLs, versioning, or range support.

### SQS behavior

Verify and fix:

- AWS JSON protocol works with default current SDK v2;
- query fallback remains correct;
- queue URL parsing cannot target/corrupt another queue;
- message MD5;
- visibility timeout with injectable clock;
- unique/stale receipt handles;
- no duplicate concurrent receive;
- approximate counts are internally consistent;
- JSON/XML error shapes;
- message size and valid body handling.

Do not add FIFO, DLQ, batch, long polling, attributes, or delay features.

### STS behavior

Verify exact response envelope, request metadata, unsupported-action errors, and SDK decoding.

### CLI and developer experience

Verify and fix:

- `serve`, `version`, `reset`, and `apply` help/output/exit codes;
- environment/flag precedence;
- endpoint safety;
- graceful shutdown;
- README quickstart;
- Docker non-root behavior and volume ownership;
- Make targets;
- compatibility claims match implementation.

### Test quality

Add targeted regression tests for every fixed bug.

Add fuzz tests where they provide real value, especially for:

- AWS request classification;
- S3 path/key extraction;
- SQS JSON/query parsing;
- manifest parsing.

Fuzz tests must have seed cases and remain bounded in normal test runs.

## Explicit scope exclusions

Do not add:

- Azure or GCP;
- DynamoDB, SNS, Lambda, IAM enforcement, CloudWatch;
- dashboard;
- TLS/auth system;
- real-cloud differential testing;
- plugin framework;
- Kubernetes;
- production deployment mode.

## Required verification

Run all applicable commands and fix failures:

```bash
gofmt -w <changed-go-files>
go test ./...
go test -race ./...
go vet ./...
make build
make compatibility
make docker-build
```

When Docker is available, also start the image, verify health, execute at least one SDK-backed S3/SQS/STS flow, and stop it cleanly.

The final report must include:

- findings fixed;
- regression tests added;
- verification output summary;
- remaining POC limitations;
- any issue that should block a public tag.

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
