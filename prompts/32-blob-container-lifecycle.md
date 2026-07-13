# Task 32 — Azure Blob container lifecycle

Target agent: Codex with GPT-5.6  
Prerequisite: Tasks 00–31.

Implement the first Azure Blob Storage operations: container lifecycle and metadata.

## Supported operations

Implement the REST operations used by the current official Azure Blob SDK for:

1. Create Container
2. Delete Container
3. Get Container Properties
4. Get Container Metadata
5. Set Container Metadata
6. List Containers

Container ACL/public access, leases, stored access policies, immutability, and legal hold remain unsupported.

## Persistence

Add versioned migrations for container metadata:

```text
account name
container name
ETag
last modified
metadata map
created_at
```

Requirements:

- primary key includes account and container;
- metadata encoding is deterministic and migration-friendly;
- container deletion cascades to blob metadata and managed bodies once blobs exist;
- no generic provider-neutral bucket table.

## Naming

Implement a documented practical Azure container-name validation subset:

- lowercase where required;
- length limits;
- allowed characters;
- no consecutive invalid separators;
- reject malformed percent encoding;
- reserved internal admin paths cannot be used as container names.

Return Azure-shaped:

```text
InvalidResourceName
ContainerAlreadyExists
ContainerNotFound
```

## Create Container

Recognize the SDK request form, including `restype=container`.

Behavior:

- create atomically;
- capture `x-ms-meta-*` headers;
- generate ETag and `Last-Modified`;
- duplicate create returns `409 ContainerAlreadyExists`;
- unsupported public access/encryption/immutability headers are rejected rather than ignored;
- return provider-compatible status and headers.

## Delete Container

Behavior:

- require existing container;
- delete metadata and all future contained blobs safely;
- return expected success status;
- missing container returns `ContainerNotFound`;
- no asynchronous delete simulation.

## Get Properties / Metadata

Return:

```text
ETag
Last-Modified
x-ms-meta-*
x-ms-lease-status/state/duration only if accurately modeled; otherwise omit
```

Do not fabricate properties that are not modeled.

Use `HEAD` semantics correctly: no body.

## Set Metadata

- replace the complete user metadata set according to Azure semantics;
- normalize metadata header names without losing values;
- enforce count/key/value/aggregate limits;
- generate a new ETag and last-modified time;
- reject invalid metadata names;
- update atomically.

## List Containers

Support:

```text
prefix
marker
maxresults
include=metadata
```

Requirements:

- deterministic lexical ordering;
- real marker-based pagination;
- opaque, validated marker or provider-compatible continuation encoding;
- no duplicate/omitted container for an unchanged dataset;
- correct XML namespace and escaping;
- `NextMarker` empty when complete;
- bounded maximum page size;
- unsupported include values rejected.

## Conditional headers

For this task, either implement or explicitly reject:

```text
If-Match
If-None-Match
If-Modified-Since
If-Unmodified-Since
```

Do not silently ignore them. Full condition support may be completed in Task 35.

## Tests

Cover:

- valid/invalid names;
- create;
- duplicate create;
- metadata round trip;
- metadata replacement;
- get properties via HEAD;
- delete/missing delete;
- list prefix;
- pagination across pages;
- metadata include;
- XML escaping;
- invalid marker;
- unsupported headers;
- concurrent create/delete;
- persistence across restart;
- reset/export/import integration.

## Official Azure SDK compatibility

Using the official Blob SDK with explicit loopback service URL and shared-key credential:

- create container;
- get properties;
- set/get metadata;
- list containers with prefix/pagination;
- delete container;
- typed/decoded duplicate and missing errors.

No request may leave loopback.

## Compatibility catalog

Add stable test IDs. Mark only operations proven by SDK tests as supported; conditional behavior may remain partial.

## Documentation

Update Azure Blob compatibility docs with exact operations and deviations.

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
