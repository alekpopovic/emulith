# Task 63 — Durable invocation queue, retries, scheduling, and dead-letter delivery

Target agent: Codex with GPT-5.6  
Prerequisite: Tasks 00–62.

Implement the internal durable event-delivery engine used by provider-specific async invocation and triggers.

## Goal

Provide local, persistent, at-least-once delivery with retries, concurrency control, scheduling, and dead-letter destinations without exposing a universal public cloud API.

## Delivery state machine

Use explicit states:

```text
pending
scheduled
leased
running
retry_wait
succeeded
failed_terminal
dead_letter_pending
dead_lettered
cancelled
```

Persist state transitions/audit entries.

## Delivery record

Store:

```text
event ID
provider
source service
source resource
trigger ID
target function/revision policy
opaque payload reference and checksum
content type
correlation/trace context
attempt count
maximum attempts
next attempt time
lease owner
lease expiration
deadline
priority if supported
status
last error category
created/updated/completed time
dead-letter destination
```

Requirements:

- event ID stable across attempts;
- attempt/invocation ID new each attempt;
- bounded payload;
- payload bytes stored safely outside ordinary logs;
- no provider envelope reinterpretation;
- idempotent enqueue using explicit provider delivery key when supplied;
- no exactly-once claim.

## Worker leasing

Implement durable workers:

- transactionally lease due events;
- lease expiration enables recovery;
- heartbeat only where invocation duration needs it;
- multiple Emulith worker goroutines cannot execute the same lease concurrently;
- process restart recovers expired leases;
- graceful shutdown stops new leases and handles running work;
- bounded batch selection;
- fair scheduling across functions/triggers.

Use injectable clock and deterministic tests.

## Retry policy

Support:

```text
maximum attempts
initial delay
maximum delay
multiplier
jitter
retryable error categories
non-retryable error categories
```

Requirements:

- exponential backoff with bounded deterministic/random source injection;
- timeout/crash/unavailable usually retryable;
- malformed provider event/config usually non-retryable;
- function-declared error classification remains provider-adapter controlled;
- no retry beyond event deadline/retention;
- attempt audit retained;
- policy snapshotted onto delivery so later config changes are deliberate.

## Concurrency

Enforce:

```text
global runtime limit
per-function limit
per-revision limit where useful
per-trigger limit
reserved concurrency metadata
```

Requirements:

- no starvation;
- waiting events remain durable;
- capacity release on cancellation/crash;
- revision switch policy documented: queued events use active-at-delivery or pinned revision;
- choose one policy and persist enough metadata to make it deterministic.

## Dead-letter destinations

Define an internal adapter:

```go
type DeadLetterSink interface {
    Deliver(context.Context, DeadLetterMessage) error
}
```

Provider sinks are added later.

Dead-letter message includes:

```text
event ID
source
target
attempt count
final error category
original opaque event payload or safe reference
timestamps
```

Requirements:

- dead-letter delivery itself is durable/retryable;
- no infinite loop;
- DLQ failure remains visible;
- payload size/redaction rules;
- successful DLQ marks terminal state;
- no automatic deletion before retention policy.

Provide an in-memory/fake sink for tests.

## Scheduling

Allow enqueue with `not_before`.

Requirements:

- efficient indexed due-time lookup;
- UTC;
- clock jumps handled safely;
- no busy polling;
- bounded wake-up;
- restart preserves schedule.

Full cron parsing is not in this task.

## Cancellation and reset

Support:

- cancel pending/retry-wait delivery;
- running cancellation through context where possible;
- terminal audit state;
- reset coordinated with workers;
- export obtains consistent durable state;
- import does not resume workers before validation/activation completes.

## Observability hooks

Emit structured internal events/counters without payloads:

```text
queued
leased
attempt_started
attempt_finished
retry_scheduled
dead_lettered
```

Task 71 exposes them.

## Tests

Cover:

- immediate success;
- scheduled delivery;
- retry/backoff/jitter;
- non-retryable error;
- timeout/crash;
- maximum attempts;
- DLQ success/failure/retry;
- process restart during lease;
- lease expiration;
- concurrent workers;
- concurrency limits;
- fairness;
- cancellation;
- function revision policy;
- clock jump;
- payload checksum failure;
- reset/export/import;
- migration from v0.4/v0.5 draft schema;
- race detector and goroutine leak.

## Documentation

Document at-least-once behavior, revision policy, retry categories, and DLQ semantics.

## Compatibility catalog

Internal delivery engine remains experimental; no provider trigger claim yet.

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
