# Task 41 — Formal Azure SDK compatibility suite

Target agent: Codex with GPT-5.6  
Prerequisite: Tasks 00–40.

Create a consolidated, hermetic Azure compatibility suite using current official Azure SDK for Go clients.

## Goal

Prove the documented Blob, Queue, and Table subsets through SDK-decoded operations, errors, paging, persistence, reset, and snapshot round trips.

This task should expose and fix defects found by the suite. Do not add unrelated operations.

## Harness

Extend the existing compatibility harness to start:

```text
AWS listener
Azure Blob listener
Azure Queue listener
Azure Table listener
```

on OS-assigned loopback ports.

Requirements:

- explicit service URLs;
- explicit development shared-key credential;
- no `DefaultAzureCredential`;
- no Azure CLI credential;
- no environment credential chain;
- no managed identity probes;
- custom transport rejects any non-loopback host;
- temporary state directory;
- restart using the same directory;
- bounded readiness;
- deterministic cleanup;
- stable compatibility test IDs;
- no Docker requirement for the core suite.

## Blob scenarios

Test with official SDK calls:

- create/list/get properties/set metadata/delete container;
- small upload/download;
- zero-byte and binary upload;
- high-level multi-block upload;
- get/set blob properties and metadata;
- list flat;
- list hierarchy/delimiter where supported;
- pager continuation;
- range download;
- conditional get/put/delete;
- decoded duplicate/missing/range/condition errors;
- restart persistence;
- export/reset/import.

## Queue scenarios

- queue lifecycle and metadata;
- pager continuation;
- enqueue;
- delayed message;
- peek;
- dequeue;
- visibility timeout;
- update body and visibility;
- pop receipt rotation;
- stale receipt error;
- delete;
- clear;
- TTL expiration with controlled clock;
- competing consumers;
- restart/export/reset/import.

## Table scenarios

- table lifecycle;
- all supported EDM property types;
- CRUD/upsert;
- ETag conflict;
- OData filters;
- projection;
- pagination/continuation;
- entity group transaction success;
- transaction rollback;
- restart/export/reset/import.

## Cross-provider isolation

Create resources with identical names across:

```text
AWS S3 and Azure Blob
AWS SQS and Azure Queue
DynamoDB and Azure Table
```

Verify no state collision, routing collision, reset inconsistency, or compatibility report confusion.

## Error assertions

Do not assert status codes only.

Where the SDK exposes service errors, assert:

```text
error code
status
request ID presence
typed response behavior
```

Keep direct HTTP tests for wire details, but SDK tests are required for support claims.

## Report integration

Add Azure entries and results to generated compatibility artifacts.

Expected structure:

```text
AWS
Azure
  Blob
  Queue
  Table
```

Report:

- SDK module versions;
- API versions tested;
- operation status;
- deviations;
- pass/fail IDs.

CI must fail if a supported Azure operation lacks a passing SDK test.

## Defect policy

When a scenario fails:

- fix the product;
- add a focused regression test;
- retain the SDK scenario;
- do not skip or weaken the assertion;
- downgrade compatibility status only when behavior intentionally remains partial.

## Verification

```bash
gofmt -w <changed-go-files>
go test ./...
go test -race ./...
go vet ./...
make compatibility
make compatibility-report
make compatibility-check
make build
make demo
```


## Execution contract

You are the implementation agent for this task. Complete the work in the current Emulith repository; do not stop after writing a plan.

1. Read every applicable `AGENTS.md` before changing files.
2. Inspect the repository, current architecture, migrations, tests, dependency versions, compatibility catalog, and documentation.
3. Run the relevant baseline checks before editing when practical.
4. State a concise implementation plan, then execute it immediately.
5. Make reasonable non-blocking assumptions instead of asking for confirmation.
6. Keep the change scoped to this task. Do not implement later roadmap items.
7. Preserve all existing AWS behavior and compatibility unless this task explicitly fixes a defect.
8. Prefer explicit provider-specific protocol code over a false universal cloud abstraction.
9. Never use Azurite, LocalStack, Moto, MinIO, ElasticMQ, or another emulator as an Emulith runtime dependency.
10. Never contact real Azure, AWS, or GCP endpoints. All tests must be hermetic and loopback-only.
11. Do not use `DefaultAzureCredential`, managed identity probing, user Azure CLI credentials, or ambient cloud profiles in compatibility tests.
12. Do not add accounts, entitlement checks, license keys, forced telemetry, analytics, or phone-home behavior.
13. Do not commit, push, tag, publish a release, or open a pull request.
14. Bound all parsers, request bodies, archive inputs, page sizes, and allocations derived from untrusted input.
15. Never log account keys, authorization headers, SAS tokens, request bodies containing user data, queue messages, entities, or blob bodies.
16. Format changed files and run every verification command applicable to the repository.
17. Fix failures caused by your change. If the environment blocks a command, report the exact limitation and run the closest safe verification.
18. Update compatibility documentation only for behavior backed by executable SDK compatibility tests.
19. Finish with:
    - implementation summary;
    - important design decisions;
    - changed files;
    - exact commands run and outcomes;
    - compatibility status changes;
    - genuine remaining limitations.

Emulith remains a development/CI emulator, not a production service.
