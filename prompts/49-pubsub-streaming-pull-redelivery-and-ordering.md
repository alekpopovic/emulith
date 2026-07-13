# Task 49 — Pub/Sub StreamingPull, redelivery, flow control, and ordering

Target agent: Codex with GPT-5.6  
Prerequisite: Tasks 00–48.

Implement the bidirectional gRPC `StreamingPull` method used by high-level Pub/Sub subscribers.

## Protocol requirements

The first client stream message must identify the subscription and may include stream-level settings. Later client messages carry:

```text
ack_ids
modify_deadline_ack_ids
modify_deadline_seconds
client flow-control settings where emitted
```

Requirements:

- reject missing/changed subscription on an established stream;
- parse bounded batches;
- validate matched modify-deadline arrays;
- return gRPC statuses without panics;
- never log message payloads or ack IDs in full.

## Server stream lifecycle

Implement:

- stream registration;
- delivery loop;
- client receive loop;
- coordinated cancellation;
- clean disconnect;
- reconnect;
- graceful server shutdown;
- bounded goroutine count;
- no send after stream close;
- no goroutine/channel leaks;
- request/stream ID in logs.

Use contexts and explicit ownership of goroutines/channels.

## Delivery behavior

- select available deliveries transactionally;
- assign/rotate ack IDs;
- set deadlines;
- increment delivery attempts;
- send bounded batches;
- avoid delivering one attempt to two streams;
- redeliver after deadline or disconnect according to documented behavior;
- preserve message ID/data/attributes/publish time/order key;
- maintain fairness among active streams without claiming production-grade global scheduling.

## Flow control and backpressure

Honor useful client settings or apply server defaults:

```text
max outstanding messages
max outstanding bytes
stream acknowledgement activity
```

Requirements:

- no unbounded buffering;
- a slow consumer cannot exhaust process memory;
- server pauses delivery when outstanding limits are reached;
- acknowledgements/deadline changes release/update capacity;
- stream cancellation releases leased deliveries for later redelivery;
- bounded send timeout or cancellation behavior.

## Ack/deadline processing

Reuse the unary state machine.

- acknowledgements on the stream are scoped to the subscription;
- modify deadline supports per-request common seconds as emitted by protocol;
- deadline zero immediately releases;
- stale ack IDs are ignored or rejected according to documented client-compatible behavior;
- one malformed client message must not corrupt the stream state.

## Ordering keys

Implement a useful local ordering subset.

When messages have an ordering key:

- only one unacked message for a given subscription + ordering key is outstanding at a time;
- later messages with that key wait;
- different ordering keys may be delivered concurrently;
- ack/release/deadline expiry advances the key;
- reconnect preserves ordering;
- messages with empty ordering key follow normal delivery;
- no exactly-once guarantee.

If the official publisher requires topic ordering configuration, model only the minimum needed and document deviations.

## Failure/retry semantics

Test:

- abrupt client disconnect;
- context cancellation;
- server shutdown;
- deadline expiry;
- reconnect and redelivery;
- publisher during active streams;
- subscription deletion during stream;
- topic deletion behavior;
- malformed stream requests;
- slow consumer;
- multiple streams on one subscription.

## Official high-level client compatibility

Use the official high-level subscription receive API that uses StreamingPull internally.

Test:

1. create topic/subscription;
2. start receiver;
3. publish several messages;
4. ack some;
5. intentionally not ack one;
6. observe redelivery;
7. verify ordering-key sequence;
8. cancel receiver;
9. restart receiver;
10. verify clean shutdown and no duplicate acked messages.

Use bounded contexts and no long sleeps; inject server clock where possible.

## Metrics/debug visibility

Do not add a dashboard. Optional internal counters may expose through existing health/debug endpoints only if they do not include payloads:

```text
active streams
outstanding deliveries
redeliveries
```

Keep them experimental.

## Tests

Add race/leak-focused tests and use available goroutine leak tooling only if lightweight and pinned.

## Compatibility catalog and docs

Mark `StreamingPull` supported/partial according to official client coverage. Explicitly list unsupported push, exactly-once, dead-letter, seek/snapshots, schemas, and production flow-control fidelity.

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
