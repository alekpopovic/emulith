# Task 73 — Serverless manifest, Docker runtime wiring, state, and developer UX

Target agent: Codex with GPT-5.6  
Prerequisite: Tasks 00–72.

Integrate functions, triggers, builds, runtime configuration, snapshots, Docker Compose, demos, and CLI into a coherent v0.5 user experience.

## Manifest

Support a strict mixed-provider example:

```yaml
version: 1
project: demo

resources:
  - type: aws.sqs.queue
    name: invoice-events

  - type: azure.storage.container
    name: invoices

  - type: gcp.pubsub.topic
    name: audit-events

functions:
  invoice-worker:
    provider: aws
    image: emulith/invoice-worker:dev
    runtime: custom
    timeout: 30s
    memory: 256MiB
    environment:
      APP_ENV: local
    triggers:
      - type: aws.sqs
        queue: invoice-events
        batch_size: 10
        report_batch_item_failures: true

  azure-blob-worker:
    provider: azure
    image: emulith/azure-blob-worker:dev
    runtime: custom-handler
    handler:
      port: 8080
    triggers:
      - type: azure.blob
        container: invoices
        prefix: incoming/

  gcp-event-worker:
    provider: gcp
    image: emulith/gcp-event-worker:dev
    runtime: functions-framework
    signature_type: cloudevent
    triggers:
      - type: gcp.pubsub
        topic: audit-events
```

Requirements:

- strict unknown-field rejection;
- full dependency validation before mutation;
- topological apply order: resources, functions/revisions, triggers;
- cycles/references rejected;
- provider-specific trigger fields;
- deterministic idempotent apply;
- failure rollback or accurate partial-state report;
- no generic trigger fields that erase provider semantics;
- source builds supported through Task 62;
- sensitive environment references redacted.

## CLI

Complete:

```bash
emulith functions build [name]
emulith functions deploy [name]
emulith functions list
emulith functions inspect <name>
emulith functions revisions <name>
emulith functions invoke <name>
emulith functions logs <name>
emulith functions invocations <name>
emulith functions delete <name>
emulith triggers list
emulith triggers inspect <id>
emulith triggers enable <id>
emulith triggers disable <id>
```

Requirements:

- provider-aware invoke options;
- `--payload`/`--payload-file` with size bounds;
- no payload echo in debug logs;
- JSON output;
- actionable Docker/runtime unavailable messages;
- endpoint guards;
- exit codes;
- completion/help consistency.

## Runtime endpoint and Docker Compose

Define a secure, explicit way for Emulith to access Docker:

1. local host process uses local Docker API; and/or
2. containerized Emulith uses an explicitly mounted Docker socket or a configured remote runtime endpoint.

Requirements:

- never auto-mount socket;
- Compose example containing socket mount has a prominent security warning;
- no privileged mode;
- socket read/write implications documented;
- optional Docker socket proxy/sidecar pattern may be documented but not falsely presented as complete isolation;
- external runtime endpoint requires explicit config;
- managed function network;
- build-cache volume;
- state volume;
- all existing service ports;
- non-root Emulith container where socket permissions permit;
- clear Linux/macOS/Windows notes.

Create/update:

```text
examples/docker-compose/serverless/
examples/functions/aws-custom-runtime/
examples/functions/azure-custom-handler/
examples/functions/gcp-functions-framework/
```

## Snapshot and migration

Create authentic v0.4 -> v0.5 migration coverage.

Snapshot includes:

```text
function definitions
immutable revisions
image references/digests
trigger definitions
queued/scheduled/retry deliveries
DLQ state
scheduler state
invocation/attempt metadata according to retention
logs according to documented policy
```

Snapshot excludes:

```text
running containers
warm pools
open HTTP/gRPC connections
Docker image layers
active stream process handles
in-memory locks
active Firestore transactions
```

On import:

- validate all metadata/checksums;
- mark missing image revisions unavailable;
- preserve queued deliveries;
- do not lease work until import activation completes;
- missing image events remain visible/retryable/terminal according to policy;
- no silent rebuild or registry pull;
- schedulers resume without duplicate normal fire;
- in-flight-at-export policy is explicit.

## Reset

Reset must:

- stop/drain workers;
- stop/remove managed containers;
- remove function/trigger/delivery state;
- preserve unrelated Docker resources;
- leave cloud listeners/state healthy;
- report partial failure accurately;
- not race with build/export/import.

## Demos

Add:

```bash
make demo-serverless
make demo-multicloud-serverless
```

### Serverless demo

- build local fixture images;
- deploy one function per provider;
- invoke directly;
- exercise one trigger per provider;
- show logs/invocations;
- prove warm reuse;
- clean up.

### Multicloud serverless demo

- one event chain crossing provider services through function code;
- trace/correlation visible;
- intentional retry and DLQ;
- restart and recovery;
- no real cloud/network.

## Documentation

Add/update:

```text
docs/serverless-quickstart.md
docs/functions-manifest.md
docs/serverless-security.md
docs/state-format.md
docs/upgrade-v0.4-to-v0.5.md
README.md
```

## Tests

Cover:

- strict mixed manifest;
- dependency validation;
- idempotent apply;
- CLI payload handling/redaction;
- Compose validation;
- socket-security docs check;
- migration/snapshot/reset;
- missing image import;
- scheduler/retry resume;
- demos' endpoint guards;
- no ambient credentials/public network.

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
make demo-serverless
make demo-multicloud-serverless
make docker-build
```


## Execution contract

You are the implementation agent for this task. Complete the work in the current Emulith repository; do not stop after writing a plan.

1. Read every applicable `AGENTS.md` before changing files.
2. Inspect the repository, provider registry, listener lifecycle, state schema, migration history, SDK compatibility suite, generated compatibility catalog, Docker setup, release tooling, and documentation.
3. Run the relevant baseline checks before editing when practical.
4. State a concise implementation plan, then execute it immediately.
5. Make reasonable non-blocking assumptions instead of asking for confirmation.
6. Keep the change scoped to this task. Do not implement later roadmap items.
7. Preserve all existing AWS, Azure, and GCP behavior and compatibility unless this task explicitly fixes a defect.
8. Keep provider-specific control planes, invocation envelopes, and event semantics separate. Do not create a false universal cloud-function API.
9. Never depend on LocalStack, Azurite, Google emulators, SAM Local, Azure Functions Core Tools, Functions Framework emulators, Moto, MinIO, ElasticMQ, or another emulator as an Emulith runtime dependency.
10. Never contact real AWS, Azure, GCP, registries, metadata services, or public cloud endpoints during tests. All compatibility and end-to-end tests must be hermetic and loopback-only.
11. Do not use ambient cloud credentials, Docker registry credentials, user profiles, ADC, managed identity, instance metadata, or default credential chains in tests.
12. Do not add accounts, entitlement checks, license keys, forced telemetry, analytics, or phone-home behavior.
13. Do not commit, push, tag, publish a release/image, or open a pull request.
14. Bound all request bodies, event payloads, logs, build contexts, archive inputs, queues, retries, streams, concurrency, and allocations derived from untrusted input.
15. Never log secrets, authorization headers, environment-variable values marked sensitive, event payloads, queue messages, object contents, function request bodies, or function responses by default.
16. Treat access to a Docker/OCI daemon as a trusted local security boundary. Do not claim strong multi-tenant isolation.
17. Keep request IDs, invocation IDs, event IDs, error mapping, and compatibility claims deterministic and test-backed.
18. Format changed files and run every verification command applicable to the repository.
19. Fix failures caused by your changes. If the environment blocks Docker or another command, report the exact limitation and run the closest safe verification using fakes/in-process tests.
20. Update compatibility documentation only for behavior backed by executable SDK or protocol compatibility tests.
21. Finish with:
    - implementation summary;
    - important design decisions;
    - changed files;
    - exact commands run and outcomes;
    - compatibility status changes;
    - security-boundary notes;
    - genuine remaining limitations.

Emulith remains a development/CI emulator, not a production service.
