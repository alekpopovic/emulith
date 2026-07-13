# Task 40 — Azure Table entity group transactions

Target agent: Codex with GPT-5.6  
Prerequisite: Tasks 00–39.

Implement Azure Table batch/entity group transactions using bounded multipart parsing and atomic local commits.

## Supported transaction operations

Within one transaction support:

```text
Insert
Replace
Merge
Insert or Replace
Insert or Merge
Delete
```

Query/read operations inside the transaction are unsupported unless the official API and SDK require them and they are implemented correctly.

## Core constraints

Enforce:

- maximum 100 operations;
- all operations target one table;
- all entities have the same `PartitionKey`;
- no duplicate operation for the same entity key unless the service contract explicitly permits it;
- bounded total request size;
- bounded part/header size;
- atomic all-or-nothing mutation;
- one transaction cannot escape its account/table/partition.

## Multipart parser

Parse the `multipart/mixed` request emitted by the official SDK.

Requirements:

- support nested changeset boundary structure where emitted;
- validate boundary syntax and length;
- limit part count;
- limit headers per part and total header bytes;
- reject duplicate/conflicting content headers;
- parse `Content-ID`;
- parse embedded HTTP request line/headers/body;
- no path traversal or cross-service URL;
- no recursive unbounded MIME parsing;
- no temp-file leak;
- malformed input returns a provider-shaped batch error, not a panic.

Do not rely on a permissive parser without adding resource guards.

## Transaction planning

Before writing:

1. parse every operation;
2. resolve table/partition/key;
3. validate body/types/ETag/conditions;
4. detect duplicate/conflicting operations;
5. create a deterministic mutation plan;
6. execute in one SQLite transaction;
7. generate per-operation results;
8. commit only after every operation succeeds.

If one operation fails:

- roll back all mutations;
- return a multipart response identifying the failing operation in an SDK-compatible way;
- preserve all original entities/ETags;
- do not leak SQL details.

## Multipart response

Return the structure expected by the SDK:

- outer boundary;
- per-operation HTTP status;
- `Content-ID` correlation;
- ETag/location/body where required;
- correct CRLF handling;
- bounded deterministic output.

Do not assume response order can differ from request order.

## ETag and upsert semantics

Reuse the single-entity implementation rather than duplicating logic.

- wildcard/stale ETag behavior;
- Replace vs Merge;
- upsert behavior;
- keys immutable;
- Timestamp/ETag updated only on committed success.

## Tests

Cover:

- multiple inserts;
- insert + merge + delete;
- all supported operation types;
- same-partition success;
- mixed PartitionKey rejected;
- mixed table rejected;
- duplicate key operation;
- stale ETag;
- one invalid entity rolls back all;
- max 100 boundary and 101 rejection;
- oversized body;
- malformed boundary;
- missing Content-ID;
- embedded cross-account URL;
- multipart CRLF variants accepted/rejected deliberately;
- concurrent transactions on same partition;
- persistence/restart;
- reset/export/import.

## Official Azure SDK compatibility

Use the SDK transaction API:

- submit successful transaction;
- verify all entities;
- submit a transaction with one conflict;
- verify full rollback;
- verify result order/status/ETags;
- exercise upsert/merge/delete.

## Compatibility catalog and docs

Add stable test IDs and document the exact transaction subset and local atomicity.

## Verification

```bash
gofmt -w <changed-go-files>
go test ./...
go test -race ./...
go vet ./...
make compatibility
make compatibility-check
make build
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
