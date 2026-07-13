# Task 38 — Azure Table service and entity CRUD

Target agent: Codex with GPT-5.6  
Prerequisite: Tasks 00–37.

Implement Azure Table Storage table lifecycle and entity CRUD using the OData/JSON protocol emitted by the current official Azure SDK.

## Supported table operations

1. Create Table
2. Delete Table
3. List Tables

## Supported entity operations

1. Insert Entity
2. Get Entity
3. Replace Entity
4. Merge Entity
5. Insert or Replace
6. Insert or Merge
7. Delete Entity

Query and batch transactions are implemented in Tasks 39 and 40.

## Entity model

Every entity has:

```text
PartitionKey
RowKey
Timestamp
ETag
dynamic properties
```

Support typed values:

```text
Edm.String
Edm.Int32
Edm.Int64
Edm.Double
Edm.Boolean
Edm.DateTime
Edm.Guid
Edm.Binary
```

Requirements:

- preserve Int64 without JSON float loss;
- preserve DateTime in UTC with documented precision;
- validate GUID format;
- decode/encode binary safely;
- distinguish null, missing, and empty string;
- enforce property count/name/value/entity-size limits;
- reserve system property names;
- store a migration-friendly typed representation, not Go gob;
- composite primary key is `(account, table, PartitionKey, RowKey)`;
- deterministic ETag generation/versioning.

## Table lifecycle

### Create Table

- validate table name;
- create atomically;
- duplicate returns a provider-compatible conflict;
- return the entity/service shape expected by the SDK.

### List Tables

Support:

```text
$top
continuation headers
```

Basic `$filter` for table names may wait until Task 39 unless the SDK requires it.

### Delete Table

- delete all entities;
- missing table returns correct error;
- immediate local deletion.

## Entity addressing

Parse the key syntax emitted by the SDK safely, for example logical forms equivalent to:

```text
Table(PartitionKey='...',RowKey='...')
```

Requirements:

- correct OData string escaping;
- percent decode exactly once;
- reject malformed/injected key expressions;
- do not implement key parsing with fragile regex alone;
- separate route parsing from general query expression parsing.

## Insert Entity

- require PartitionKey and RowKey;
- reject duplicate key;
- assign Timestamp/ETag;
- support return-content/minimal preference used by the SDK;
- validate all types before writing.

## Get Entity

- return exact typed properties;
- include ETag/Timestamp;
- missing entity returns correct not-found response;
- `$select` may be rejected until Task 39.

## Replace / Merge

- Replace overwrites user properties while retaining keys/system fields;
- Merge updates supplied properties and preserves others;
- use `If-Match`;
- wildcard `*` accepted;
- stale ETag returns condition failure;
- keys cannot change;
- atomically update ETag/Timestamp.

## Upserts

Implement:

```text
Insert or Replace
Insert or Merge
```

according to the HTTP methods/headers emitted by the SDK.

## Delete Entity

Require key and condition semantics. Wildcard ETag is supported. Missing/stale behavior must be SDK-compatible.

## Errors

Implement OData/Azure-shaped:

```text
TableAlreadyExists
TableNotFound
EntityAlreadyExists
ResourceNotFound
UpdateConditionNotSatisfied
InvalidInput
PropertyValueTooLarge
```

## Tests

Cover:

- table names/lifecycle;
- every EDM type;
- Int64 precision;
- binary/date/GUID;
- insert/get;
- duplicate;
- replace;
- merge;
- upserts;
- delete;
- wildcard/stale ETag;
- key escaping;
- malicious/malformed entity path;
- size/property limits;
- concurrency;
- restart;
- reset/export/import.

## Official Azure SDK compatibility

Use the current official Table SDK for:

- table lifecycle;
- add/get/update/merge/upsert/delete entity;
- typed properties;
- ETag conflict;
- decoded errors.

Use explicit loopback endpoint/shared key only.

## Compatibility catalog and docs

Mark CRUD operations supported/partial only according to SDK tests. Query/batch remain unsupported.

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
