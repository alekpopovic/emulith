# Task 58 — GCP manifest, Docker, state, snapshot, and compatibility integration

Target agent: Codex with GPT-5.6  
Prerequisite: Tasks 00–57.

Integrate GCP services into Emulith's user experience, Docker setup, manifest, state migrations, snapshots, reset, and generated compatibility reports.

## Manifest resources

Extend the strict experimental manifest schema with:

```yaml
version: 1
project: demo

resources:
  - type: gcp.pubsub.topic
    name: invoice-events

  - type: gcp.pubsub.subscription
    name: invoice-worker
    topic: invoice-events
    ack_deadline_seconds: 30

  - type: gcp.storage.bucket
    name: invoices-local
    location: EU
```

Firestore collections are not created explicitly. Add optional seed documents:

```yaml
seed:
  - type: gcp.firestore.document
    path: users/alice
    data:
      name: Alice
      active: true
      score: 42
```

Requirements:

- strict YAML decoding;
- validate all GCP resource names;
- resolve subscription topic references;
- validate entire manifest before API mutations;
- mixed AWS/Azure/GCP resources in one manifest;
- deterministic apply order;
- idempotent repeated apply;
- use official public clients/endpoints, not state internals;
- no real cloud fallback;
- useful output and non-zero failure behavior;
- typed seed conversion without Int64 precision loss.

## CLI

Complete:

```bash
emulith gcp env
emulith apply -f emulith.yaml
```

Optionally add a concise:

```bash
emulith gcp status
```

only if it reports local endpoints/readiness without remote calls.

## Docker

Expose/document:

```text
4566  AWS
10000 Azure Blob
10001 Azure Queue
10002 Azure Table
8085  GCP Pub/Sub
8080  GCP Firestore
9023  GCP Storage
```

Update the Dockerfile/OCI metadata only as needed.

Create/update:

```text
examples/docker-compose/multicloud-basic/
```

Requirements:

- all ports;
- persistent volume;
- non-root;
- graceful shutdown;
- no privileged mode;
- no Docker socket;
- no cloud secrets;
- health checks use tools actually present;
- example environment for all providers;
- restart persistence demonstration.

## GCP demo

Create:

```bash
make demo-gcp
```

Use official Go clients.

Flow:

1. start Emulith;
2. create Pub/Sub topic/subscription;
3. publish/receive/ack;
4. create Storage bucket;
5. resumable upload/download;
6. write/read/query Firestore document;
7. stop cleanly.

Guard against non-loopback endpoints and ADC.

## Multicloud demo

Create or complete:

```bash
make demo-multicloud
```

Use one Emulith process and verify a concise flow across AWS, Azure, and GCP. Avoid duplicating server startup scripts.

## State migrations

Create authentic v0.3 -> v0.4 migration coverage.

Include GCP tables/state for:

### Pub/Sub

```text
topics
subscriptions
messages
delivery/ack state
ordering state needed for restart
```

### Storage

```text
buckets
objects
body files
resumable sessions/temp files if snapshot-safe
```

### Firestore

```text
documents
typed fields
transaction metadata only if active transactions are snapshot-supported
```

Requirements:

- existing AWS/Azure data intact;
- migration idempotent;
- rollback/recovery on failure;
- no downgrade;
- version documented.

## Snapshot export/import

Round-trip mixed state.

Decide and document active ephemeral-state behavior:

- Pub/Sub in-flight deliveries;
- resumable uploads;
- active Firestore transactions;
- active StreamingPull streams.

Prefer:

- durable delivery deadlines preserved;
- streams not preserved;
- active transactions canceled/excluded;
- resumable sessions preserved only if fully validated.

Never create a snapshot that silently restores corrupt/unusable state.

Validate all file checksums, resource names, and schema versions.

## Reset

One reset must remove GCP durable state and files while leaving all listeners/store healthy. Coordinate with AWS/Azure atomically or report failure accurately.

## Compatibility report

Extend generated report sections:

```text
AWS
Azure
GCP
  Pub/Sub
  Cloud Storage
  Firestore
```

Include:

- official client versions;
- protocol/API method;
- status;
- deviations;
- test IDs;
- provider/service summaries.

No manual duplication.

## Tests

Cover:

- strict manifest;
- mixed-provider apply;
- GCP seed;
- demo guards;
- migration from v0.3 fixture;
- mixed snapshot round trip;
- active ephemeral-state policy;
- reset;
- Docker Compose config validation;
- generated report determinism;
- no credential/network fallback.

## Documentation

Update quickstart, ports, environment variables, state format, upgrade guide, architecture, and compatibility docs.

Add:

```text
docs/upgrade-v0.3-to-v0.4.md
```

## Verification

```bash
gofmt -w <changed-go-files>
go test ./...
go test -race ./...
go vet ./...
make build
make compatibility
make compatibility-report
make compatibility-check
make demo
make demo-azure
make demo-gcp
make demo-multicloud
make docker-build
```


## Execution contract

You are the implementation agent for this task. Complete the work in the current Emulith repository; do not stop after writing a plan.

1. Read every applicable `AGENTS.md` before changing files.
2. Inspect the repository, provider registry, listener lifecycle, migrations, tests, compatibility catalog, generated documentation, Docker setup, and dependency versions.
3. Run the relevant baseline checks before editing when practical.
4. State a concise implementation plan, then execute it immediately.
5. Make reasonable non-blocking assumptions instead of asking for confirmation.
6. Keep the change scoped to this task. Do not implement later roadmap items.
7. Preserve all existing AWS and Azure behavior and compatibility unless this task explicitly fixes a defect.
8. Prefer provider-specific protocol implementations over a false universal cloud abstraction.
9. Never use Google Cloud emulators, LocalStack, Azurite, Moto, MinIO, ElasticMQ, or another emulator as an Emulith runtime dependency.
10. Never contact real GCP, AWS, or Azure endpoints. All tests must be hermetic and loopback-only.
11. Do not use Application Default Credentials, `gcloud` user credentials, service-account files, metadata-server probing, workload identity, or ambient cloud profiles in compatibility tests.
12. Do not add accounts, entitlement checks, license keys, forced telemetry, analytics, or phone-home behavior.
13. Do not commit, push, tag, publish a release, or open a pull request.
14. Bound all parsers, protobuf/JSON bodies, gRPC messages, stream buffers, page sizes, upload sessions, query plans, and allocations derived from untrusted input.
15. Never log authorization headers, OAuth tokens, API keys, signed URLs, object bodies, Pub/Sub message data, Firestore document data, or other user payloads.
16. Keep request IDs, error mapping, and compatibility claims deterministic and test-backed.
17. Format changed files and run every verification command applicable to the repository.
18. Fix failures caused by your changes. If the environment blocks a command, report the exact limitation and run the closest safe verification.
19. Update compatibility documentation only for behavior backed by executable official-client compatibility tests.
20. Finish with:
    - implementation summary;
    - important design decisions;
    - changed files;
    - exact commands run and outcomes;
    - compatibility status changes;
    - genuine remaining limitations.

Emulith remains a development/CI emulator, not a production service.
