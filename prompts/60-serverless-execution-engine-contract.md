# Task 60 — Serverless execution engine contract

Target agent: Codex with GPT-5.6  
Prerequisite: Tasks 00–59 have produced the AWS + Azure + GCP `v0.4.0` codebase.

Design and implement the provider-neutral *internal* execution-domain contract for local functions. This task must not start Docker containers or expose full provider control-plane APIs yet.

## Goal

Create a durable, testable model for:

```text
Function
FunctionRevision
Invocation
InvocationAttempt
InvocationResult
ExecutionContext
Runtime
RuntimeInstance
Trigger
EventDelivery
```

The model is internal infrastructure only. AWS Lambda, Azure Functions, and GCP function adapters must retain separate public APIs and event envelopes.

## Required state machines

### Function revision

```text
draft
building
ready
failed
disabled
deleted
```

### Invocation

```text
queued
starting
running
succeeded
failed
timed_out
cancelled
dead_lettered
```

### Runtime instance

```text
creating
starting
ready
busy
draining
stopped
failed
```

Transitions must be explicit, validated, and covered by tests. Invalid transitions must fail without silently corrupting state.

## Core identifiers

Define stable local identifiers for:

```text
function ID
revision ID
invocation ID
attempt ID
event ID
trigger ID
runtime instance ID
correlation ID
```

Requirements:

- unique without external coordination;
- safe for logs, URLs, and persistence;
- generated through injectable interfaces for deterministic tests;
- event ID remains stable across retries;
- invocation/attempt ID changes per attempt;
- no provider resource ARN/URL/name is used as the sole internal primary key.

## Function and revision model

A function definition must include:

```text
logical name
provider
provider-specific public identity
runtime kind
handler/entrypoint
image reference
image digest when resolved
timeout
memory limit
CPU limit or quota metadata
ephemeral storage limit
environment metadata
concurrency limit
revision
created/updated time
```

Requirements:

- immutable revisions;
- mutable logical function points to an active revision;
- revision updates are atomic;
- environment values can be marked sensitive;
- sensitive values are not returned by default in inspect APIs or logs;
- deleting a function does not orphan in-flight durable invocations;
- a function may be disabled without deleting history.

## Invocation request/result contracts

Define an internal invocation request with:

```text
provider
function/revision
invocation type: sync or async
provider event envelope bytes
content type
headers/metadata
deadline
correlation/trace context
event ID when event-driven
attempt number
```

Define a result with:

```text
status
provider response envelope bytes
function error category
runtime error category
exit code
duration
cold start flag
log reference
started/finished time
```

Requirements:

- provider payload remains opaque to the generic engine;
- generic engine enforces byte limits but does not reinterpret provider payload;
- result distinguishes function-declared error from runtime/infrastructure failure;
- cancellation, timeout, crash, invalid response, and successful response are distinct;
- logs are referenced, not embedded without a bound.

## Runtime interfaces

Create interfaces equivalent in responsibility to:

```go
type Runtime interface {
    Start(context.Context, RuntimeSpec) (Instance, error)
    Capabilities() RuntimeCapabilities
}

type Instance interface {
    ID() string
    Invoke(context.Context, InvocationRequest) (InvocationResult, error)
    Stop(context.Context) error
    State() InstanceState
}
```

Exact signatures should follow repository conventions.

Requirements:

- no Docker-specific types in the generic contract;
- context cancellation required;
- instances may be single-concurrency or multi-concurrency;
- runtime reports capability limits;
- invocation cannot outlive its deadline;
- stop is idempotent;
- state inspection is race-safe;
- errors use typed categories suitable for retry classification.

## Scheduler/pool interfaces

Define boundaries for future warm pools:

```text
instance acquisition
cold start
warm reuse
maximum instances
per-function concurrency
draining
idle expiration
```

Do not implement the Docker pool in this task. A deterministic in-memory fake runtime/pool must prove the contract.

## Persistence

Add versioned schema for:

```text
functions
function revisions
invocations
invocation attempts
trigger definitions placeholder
```

Do not persist running process handles.

Requirements:

- state survives restart;
- an invocation left in `starting` or `running` at process crash is recovered into a documented retryable/failed state;
- timestamps use UTC/injectable clock;
- payload storage is bounded and may use managed filesystem blobs;
- checksums for stored payloads;
- no raw secret values in ordinary metadata tables where an existing secret abstraction is available;
- migrations preserve all v0.4 provider state.

## Recovery policy

On server startup:

- find nonterminal invocations;
- classify what can be retried;
- never report a crashed in-flight invocation as succeeded;
- retain the same event ID;
- create a new attempt ID;
- respect maximum-attempt metadata once Task 63 wires it;
- produce deterministic audit records.

## Public exposure

Add only internal/admin experimental inspection endpoints or CLI output if needed for tests. Do not expose AWS Lambda, Azure Functions, or GCP function control planes yet.

## Tests

Cover:

- every valid/invalid state transition;
- ID stability/uniqueness;
- immutable revisions;
- active revision switch;
- sensitive environment redaction;
- sync/async request contracts;
- timeout/cancellation distinction;
- fake cold and warm invocation;
- stop idempotency;
- concurrent acquire/invoke/state reads;
- persistence/reopen;
- crash recovery of queued/starting/running invocations;
- payload checksum and size limits;
- migration from authentic v0.4 fixture;
- reset/export/import behavior for definitions/history;
- race detector.

## Documentation

Add/update:

```text
docs/serverless-architecture.md
docs/state-format.md
docs/architecture.md
```

Document why the execution engine is internal and provider envelopes remain separate.

## Compatibility catalog

Add a serverless platform section as `experimental`. No provider function operation is supported in this task.

## Verification

```bash
gofmt -w <changed-go-files>
go test ./...
go test -race ./...
go vet ./...
make build
make compatibility-check
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
