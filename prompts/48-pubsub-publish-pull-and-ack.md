# Task 48 — Pub/Sub publish, pull, acknowledge, and ack deadlines

Target agent: Codex with GPT-5.6  
Prerequisite: Tasks 00–47.

Implement the core Pub/Sub message delivery state machine for unary publishing and pulling.

## Supported operations

1. Publish
2. Pull
3. Acknowledge
4. ModifyAckDeadline

StreamingPull is implemented in Task 49.

## Message model

Persist at least:

### Published message

```text
message ID
topic resource
binary data
attributes map
ordering key
publish time
created_at
```

### Per-subscription delivery

```text
subscription resource
message ID
delivery state
current ack ID
ack deadline
delivery attempt
available_at
acked_at
created_at
updated_at
```

Requirements:

- message data is stored as raw bytes, not base64 text internally;
- attributes preserve exact UTF-8 strings;
- message ID is stable across redelivery;
- ack ID is distinct from message ID;
- ack ID rotates for each new delivery attempt;
- one publication creates delivery records for every active subscription at commit time;
- subscriptions created later do not receive historical messages unless explicitly documented otherwise;
- cleanup/retention behavior is bounded and documented.

## Publish

Support batch publish.

Requirements:

- require existing topic;
- require each message to contain non-empty data or at least one attribute;
- enforce per-message, per-batch count, and aggregate byte limits;
- validate attribute keys/values and ordering key;
- generate deterministic-format unique message IDs;
- set publish timestamps;
- persist publication and all subscription deliveries atomically where practical;
- preserve request order in returned message IDs;
- reject unsupported schema/KMS semantics;
- no duplicate delivery records from one successful request.

## Pull

Support:

```text
subscription
max_messages
return_immediately when present in older clients
```

Requirements:

- pull subscription only;
- select currently available delivery rows transactionally;
- no two concurrent pulls receive the same delivery attempt;
- set ack deadline using subscription default;
- rotate ack ID;
- increment delivery attempt;
- return published message data/attributes/order key/time;
- deterministic selection by availability/publish/message ID;
- zero results is a valid response;
- enforce `max_messages` bounds;
- short polling only in this task.

Use an injectable clock. Avoid sleeps in tests.

## Acknowledge

- accept a batch of ack IDs;
- ack only current valid delivery attempts for the named subscription;
- define behavior for stale/unknown IDs consistently with Pub/Sub client expectations;
- idempotent repeated ack may be accepted;
- no cross-subscription ack;
- atomically update delivery state.

## ModifyAckDeadline

Support:

```text
ack_ids
ack_deadline_seconds
```

Requirements:

- valid range;
- `0` immediately makes the delivery available for redelivery and invalidates/rotates the current ack path as appropriate;
- positive value extends from current server time;
- stale/unknown IDs handled consistently;
- no cross-subscription modification;
- batch atomicity or accurately documented partial validation behavior.

## Retention and cleanup

Implement a safe local policy:

- acked deliveries may be removed immediately or retained according to configured subscription retention;
- source messages can be garbage-collected when no delivery references remain;
- expiration/cleanup uses indexed queries and bounded batches;
- restart preserves in-flight deadlines and causes correct redelivery.

## Errors

Use appropriate gRPC statuses:

```text
NotFound
InvalidArgument
FailedPrecondition
ResourceExhausted
Unimplemented
Internal
```

## Tests

Cover:

- binary data;
- attributes;
- batch publish;
- multiple subscriptions each receive one copy;
- subscription created after publish receives nothing;
- pull;
- ack;
- redelivery after deadline using fake clock;
- ModifyAckDeadline extension;
- deadline zero;
- ack ID rotation;
- stale ack ID;
- cross-subscription ack rejection;
- two competing pull callers;
- batch/size limits;
- missing topic/subscription;
- restart with in-flight delivery;
- reset/export/import;
- race detector.

## Official client compatibility

Using the official Pub/Sub client or generated publisher/subscriber client:

- publish;
- pull;
- ack;
- modify ack deadline;
- redelivery;
- multiple subscriptions;
- typed gRPC errors.

No real GCP connection or ADC.

## Compatibility catalog and docs

Add stable operation IDs and document short-poll/unary-only behavior until Task 49.

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
