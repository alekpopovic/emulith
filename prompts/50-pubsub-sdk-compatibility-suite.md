# Task 50 — Formal Pub/Sub SDK compatibility suite

Target agent: Codex with GPT-5.6  
Prerequisite: Tasks 00–49.

Consolidate Pub/Sub testing into a formal official-client compatibility suite and fix any defects it reveals.

## Harness

Use the existing loopback-only multicloud harness.

Requirements:

- OS-assigned Pub/Sub gRPC port;
- `PUBSUB_EMULATOR_HOST` or explicit endpoint;
- project `emulith-local`;
- no credentials;
- no ADC;
- no metadata server;
- custom dialer rejects non-loopback addresses;
- temporary state directory;
- clean restart using the same state;
- bounded contexts;
- stable compatibility test IDs;
- no Docker requirement for core tests.

## Required scenarios

### Topic lifecycle

- create;
- get;
- labels/allowed update;
- list pagination;
- duplicate create;
- delete;
- missing get.

### Subscription lifecycle

- create;
- get;
- allowed update;
- list;
- list by topic;
- delete;
- unsupported push/dead-letter/exactly-once fields.

### Unary messaging

- binary publish;
- attributes;
- batch publish;
- pull;
- acknowledge;
- modify deadline;
- deadline zero;
- redelivery;
- stale ack ID;
- multiple subscriptions.

### StreamingPull

- high-level receive callback;
- acknowledgement;
- redelivery of unacked message;
- cancellation;
- reconnect;
- multiple receiver goroutines;
- slow callback/backpressure;
- subscription deletion while receiving.

### Ordering

- one ordering key remains serial;
- different keys progress independently;
- restart/reconnect maintains order for pending messages;
- no acked-message redelivery.

### Persistence and lifecycle

- clean server restart with pending messages;
- in-flight deadline across restart;
- reset removes resources/messages;
- export/import round trip;
- mixed AWS/Azure/GCP state remains isolated.

### Concurrency

- concurrent publishers;
- competing pull callers;
- several StreamingPull clients;
- ack vs deadline expiry race;
- topic/subscription delete vs publish/receive;
- `go test -race` clean.

## Error assertions

Assert official-client/gRPC behavior, not only status codes:

```text
codes.AlreadyExists
codes.NotFound
codes.InvalidArgument
codes.FailedPrecondition
codes.ResourceExhausted
```

Keep direct protocol tests for wire details.

## Compatibility report

Ensure Pub/Sub entries include:

```text
official client module version
gRPC method
test ID
status
known deviations
```

No supported status without a passing official-client test.

## Performance sanity

Add non-brittle benchmarks or tests for:

- several thousand small messages;
- batch publish/pull;
- active streaming receiver;
- bounded memory/outstanding delivery counts.

Do not create fragile wall-clock CI gates.

## Defect policy

Fix product bugs and add focused regressions. Do not skip failing official-client scenarios or weaken assertions to obtain green CI.

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
