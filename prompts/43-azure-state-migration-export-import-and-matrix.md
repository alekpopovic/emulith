# Task 43 — Azure state migration, snapshot integration, and compatibility reporting

Target agent: Codex with GPT-5.6  
Prerequisite: Tasks 00–42.

Integrate all Azure state into Emulith migrations, reset, export/import, and generated compatibility reports.

## Schema migration

Create tested migration fixtures from the real `v0.2.0` state schema.

Requirements:

- upgrade adds Azure accounts/container/blob/block/queue/message/table/entity tables safely;
- existing AWS metadata and object files remain intact;
- migration is idempotent;
- failure leaves a recoverable pre-migration state;
- no automatic downgrade;
- schema version is recorded;
- migration does not require internet or a real cloud account.

Do not invent historical schema details; derive fixtures from repository migrations/releases.

## Snapshot format

Increment the state snapshot format only if required.

Include:

### Blob

```text
containers
blob metadata
committed body files
uncommitted block metadata/files
ETags/timestamps/content headers/metadata
```

### Queue

```text
queue metadata
message body
insertion/expiration/visibility times
dequeue count
message ID
current pop receipt state
```

### Table

```text
table metadata
typed entities
ETags/timestamps
```

Requirements:

- manifest lists all payload files with size/SHA-256;
- import validates every file;
- no links/special files;
- archive limits;
- staged block references cannot escape staging root;
- imported account/service names are validated;
- unsupported newer schema/snapshot is rejected;
- failed import leaves original state usable;
- `--replace` remains explicit;
- activation is atomic.

## Reset

Verify reset removes:

```text
all Azure metadata
all blob bodies
all staged blocks
all queue messages
all table entities
```

while keeping listeners/store healthy.

AWS and Azure reset behavior must be coordinated so a single reset cannot leave one provider partially reset without reporting failure.

## Required round-trip scenarios

### Azure-only

1. create containers/blobs/blocks;
2. create queue/messages with visibility/TTL state;
3. create table/entities;
4. export;
5. reset;
6. import;
7. verify through official SDK calls.

### Mixed AWS + Azure

1. create S3 and Blob resources with same names;
2. create SQS and Azure Queue resources with same names;
3. create DynamoDB and Azure Table data;
4. include SNS/Logs data;
5. export;
6. restart;
7. reset;
8. import;
9. verify every provider through public SDKs.

### Backward compatibility

- import a supported v0.2 snapshot;
- migrate it;
- verify AWS data;
- Azure state begins empty;
- export in the new format;
- reject unsupported future format/schema.

## Compatibility report

Extend generated artifacts to support provider sections without hand-maintained duplication.

Requirements:

- deterministic provider/service/operation order;
- AWS results remain unchanged;
- Azure SDK/API versions included;
- generated Markdown clearly separates AWS and Azure;
- stale docs fail CI;
- supported status requires passing SDK test;
- partial entries show deviations;
- summary counts by provider/service.

## State documentation

Update:

```text
docs/state-format.md
docs/upgrade-v0.2-to-v0.3.md
docs/compatibility/azure.md or generated equivalent
```

Document snapshot stability, migration behavior, and uncommitted block/message-time semantics.

## Tests

Add failure injection for:

- missing blob/block file;
- checksum mismatch;
- malformed typed Table entity;
- corrupt pop receipt state;
- migration interruption;
- mixed-provider rollback;
- concurrent export with active Azure writes;
- reset during staged upload/queue receive/Table transaction.

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
make demo-azure
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
