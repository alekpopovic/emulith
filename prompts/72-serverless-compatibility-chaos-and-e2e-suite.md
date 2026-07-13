# Task 72 — Serverless compatibility, chaos, and end-to-end suite

Target agent: Codex with GPT-5.6  
Prerequisite: Tasks 00–71.

Build the formal serverless compatibility and resilience suite. Fix defects found by the suite; do not add unrelated services.

## Harness

Extend the multicloud harness with:

```text
fake execution runtime for mandatory CI
Docker execution runtime for integration when available
test clock
fault injection
loopback-only clients/transports
temporary state
restart support
log/metric inspection
```

Requirements:

- official SDK clients where public provider control planes are implemented;
- real custom-runtime/custom-handler/Functions Framework fixture images;
- no cloud credentials;
- no registry/public image pull during tests;
- deterministic bounded timeouts;
- stable compatibility IDs;
- no fixed ports;
- clean container/process teardown.

## AWS scenarios

### Control plane

- Create/Get/List/Update/Delete image function;
- invalid/missing/conflict errors.

### Invoke/runtime

- RequestResponse;
- Event;
- DryRun;
- cold/warm;
- init error;
- handled/unhandled error;
- timeout;
- crash;
- LogType Tail;
- concurrent/reserved concurrency;
- revision update/drain.

### Triggers

- SQS success/failure/partial batch;
- SNS trigger;
- S3 create/delete/filter;
- EventBridge target;
- scheduler;
- retry/DLQ;
- restart during queued/retry state.

### Optional streams

DynamoDB Streams only if actually implemented and documented.

## Azure scenarios

- custom-handler HTTP request/response;
- route parameters;
- timer;
- Queue trigger;
- poison queue;
- Blob finalized/overwrite/delete;
- output Queue/Blob/Table binding;
- Event Grid EventGrid/CloudEvents schema;
- retry/DLQ;
- handler crash/timeout;
- warm reuse;
- restart.

## GCP scenarios

- HTTP function;
- CloudEvent function binary/structured;
- Pub/Sub trigger/ack/redelivery/order;
- Storage finalized/deleted/metadata;
- Firestore create/update/delete/write;
- Eventarc filter;
- Scheduler HTTP/PubSub target;
- retry/DLQ;
- concurrency;
- restart.

## Cross-provider scenario

Use a provider event to produce output consumed by another provider only through application/function behavior and Emulith endpoints, for example:

1. S3 event invokes Lambda fixture.
2. Fixture writes Azure Queue message.
3. Azure Queue trigger invokes custom handler.
4. Handler publishes Pub/Sub.
5. Pub/Sub trigger invokes GCP CloudEvent fixture.
6. Fixture writes Firestore and GCS.
7. Correlation trace remains visible.

Document that cross-cloud behavior comes from function code, not an invented universal event bus.

## Chaos/fault scenarios

Inject:

```text
Docker daemon unavailable
container create failure
startup timeout
readiness failure
container crash before response
container crash after partial response
function timeout
client cancellation
log flood
oversized response
network-disabled function
state DB busy/transient failure
filesystem write failure
disk-full simulation where safe
Emulith shutdown during invocation
Emulith crash/restart during retry wait
reset during queued event
export during queued/running work
missing image after import
duplicate event enqueue
clock jump
slow consumer/backpressure
```

Requirements:

- no lost terminal audit state;
- no event falsely marked succeeded;
- retries bounded;
- no orphan containers;
- no deadlock/goroutine leak;
- reset/export/import policy enforced;
- provider source acknowledgement only after target success where required.

## Compatibility catalog

Add distinct dimensions:

```text
provider
service/control plane
invocation protocol
runtime type
trigger source
event envelope
delivery guarantee
retry/DLQ
test ID
status
known deviations
```

A handler unit test cannot justify `supported`.

## Test fixtures

Keep fixtures:

- minimal;
- source included;
- dependencies pinned;
- locally buildable;
- no network required after dependency cache;
- deterministic behavior selected by event/request fields;
- no embedded secrets.

## Performance/load sanity

Non-brittle scenarios:

- concurrent HTTP invocations;
- queue/topic batches;
- warm-pool reuse;
- hundreds/thousands of queued events;
- bounded logs;
- multiple providers active.

Assert invariants and memory/queue bounds rather than strict wall-clock marketing numbers.

## Defect policy

For every discovered bug:

- fix product code;
- add focused regression test;
- retain the end-to-end scenario;
- do not weaken or skip a required compatibility check.

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
make demo-gcp
make demo-multicloud
make docker-build
```

When Docker is available, run the Docker-backed serverless suite. Otherwise run the complete fake-runtime suite and report Docker skip precisely.


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
