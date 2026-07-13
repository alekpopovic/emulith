# Task 42 — Azure manifest resources, Docker Compose, and developer UX

Target agent: Codex with GPT-5.6  
Prerequisite: Tasks 00–41.

Make the Azure subset easy to run and consume without weakening the provider boundaries.

## Manifest resources

Extend the experimental manifest schema with:

```yaml
version: 1
project: demo

resources:
  - type: azure.storage.container
    name: invoices
    metadata:
      environment: local

  - type: azure.storage.queue
    name: invoice-events
    metadata:
      environment: local

  - type: azure.storage.table
    name: Users
```

Requirements:

- strict YAML decoding;
- validate names using the same service validators;
- reject unsupported fields/types;
- duplicate logical resource detection;
- validate entire manifest before mutations;
- apply in manifest order;
- idempotent repeated apply;
- use official Azure SDK clients against public local endpoints;
- do not call store internals;
- no real Azure fallback;
- actionable per-resource output.

If table/container/queue naming rules differ, preserve provider-specific validation.

## CLI

Complete:

```bash
emulith azure connection-string
emulith azure env
```

`azure env` should emit predictable `KEY=value` lines for local development, for example:

```text
AZURE_STORAGE_CONNECTION_STRING=...
EMULITH_AZURE_BLOB_ENDPOINT=...
EMULITH_AZURE_QUEUE_ENDPOINT=...
EMULITH_AZURE_TABLE_ENDPOINT=...
```

Requirements:

- stdout is machine-consumable;
- secrets are development-only but not duplicated in logs;
- optional shell formats only if cleanly implemented;
- IPv6-safe URL construction;
- explicit non-local endpoint requires user intent.

Update `emulith apply` endpoint/config handling so AWS and Azure resources can coexist in one manifest.

## Docker and Compose

Update the main image/docs to expose:

```text
4566
10000
10001
10002
```

Create:

```text
examples/docker-compose/multicloud-basic/
  docker-compose.yml
  emulith.yaml
  README.md
```

Requirements:

- non-root container;
- persistent named volume;
- all service ports;
- no privileged mode;
- no Docker socket;
- no cloud credentials;
- health/readiness checks that use tools actually available;
- environment and endpoints documented;
- restart demonstrates persistence.

## Azure demo

Create:

```bash
make demo-azure
```

Use a small Go program with official Azure SDKs, not only `curl`.

Demo flow:

1. start Emulith on loopback;
2. create Blob container;
3. upload/download a small blob;
4. create Queue;
5. enqueue/dequeue/delete a message;
6. create Table;
7. insert/query/delete an entity;
8. print a concise result;
9. cleanly stop.

Add:

```text
scripts/demo-azure.sh
examples/azure-go-sdk-demo/
```

Guard against non-loopback endpoints.

## Multicloud demo option

Add `make demo-multicloud` only if it cleanly composes the existing AWS and new Azure demos without duplicating server startup. Otherwise document running `make demo` and `make demo-azure`.

## Tests

Cover:

- strict manifest parse;
- Azure resource validation;
- no API calls on validation failure;
- mixed AWS/Azure apply;
- idempotent apply;
- CLI env/connection-string exact output;
- Docker Compose config validation;
- demo endpoint guard;
- no ambient credential use.

## Documentation

Update quickstart with:

- ports;
- connection string;
- official SDK configuration;
- Compose;
- mixed manifest;
- local-only/auth-deviation warning.

## Verification

```bash
gofmt -w <changed-go-files>
go test ./...
go test -race ./...
go vet ./...
make compatibility
make compatibility-check
make build
make demo
make demo-azure
make docker-build
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
