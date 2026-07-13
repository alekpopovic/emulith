# Task 47 — Pub/Sub topic and subscription lifecycle

Target agent: Codex with GPT-5.6  
Prerequisite: Tasks 00–46.

Implement the first Google Cloud Pub/Sub operations through the public gRPC API used by official Go clients.

## Supported topic operations

Implement:

1. CreateTopic
2. GetTopic
3. UpdateTopic with a small documented field-mask subset
4. DeleteTopic
5. ListTopics

## Supported subscription operations

Implement:

1. CreateSubscription
2. GetSubscription
3. UpdateSubscription with a small documented field-mask subset
4. DeleteSubscription
5. ListSubscriptions
6. ListTopicSubscriptions

Only pull subscriptions are supported.

## Resource model

Persist at least:

### Topic

```text
project
topic ID
full resource name
labels
message retention duration if supported
created_at
updated_at
```

### Subscription

```text
project
subscription ID
full resource name
topic resource name
ack deadline seconds
retain acked messages flag if supported
message retention duration
expiration policy metadata
filter string reserved for later
created_at
updated_at
```

Use provider-specific tables and versioned migrations.

## Validation

Validate:

- one configured local project;
- topic/subscription IDs;
- full resource names;
- ack deadline range;
- retention/expiration values;
- update masks;
- duplicate resources;
- referenced topic exists.

Reject:

```text
push_config with a real endpoint
BigQuery config
Cloud Storage config
dead-letter policy
retry policy unless implemented later
exactly-once delivery
detached subscriptions
cross-project topics/subscriptions
schemas
KMS
```

Do not silently ignore semantic fields.

## Topic behavior

### CreateTopic

- atomically create;
- duplicate returns gRPC `AlreadyExists`;
- return canonical resource.

### GetTopic

- missing returns `NotFound`.

### UpdateTopic

Support only selected fields such as labels using a valid field mask.

- reject immutable name;
- reject unsupported paths;
- update atomically;
- missing returns `NotFound`.

### DeleteTopic

- delete the topic;
- define behavior for subscriptions referencing the topic:
  - preferably preserve subscriptions in a detached/not-found-topic state only if modeled correctly; otherwise delete them transactionally and document the local deviation.
- no orphan delivery rows.

### ListTopics

Support:

```text
page_size
page_token
```

Requirements:

- deterministic lexical resource-name order;
- validated opaque token;
- no duplicate/omitted rows for unchanged state;
- bounded page size;
- correct empty token when complete.

## Subscription behavior

### CreateSubscription

- require existing topic;
- pull only;
- default ack deadline;
- duplicate returns `AlreadyExists`;
- subscription starts receiving only messages published after successful creation unless retained-topic history is explicitly modeled later.

### UpdateSubscription

Support a narrow field-mask subset, for example:

```text
ack_deadline_seconds
labels
message_retention_duration
```

Reject unsupported paths.

### Delete/Get/List

Provider-compatible `NotFound`, deterministic pagination, and topic filtering.

## gRPC errors

Map validation/storage failures to appropriate statuses:

```text
InvalidArgument
AlreadyExists
NotFound
FailedPrecondition
Unimplemented
Internal
```

Attach/request-log a local request ID without leaking internals.

## Tests

Cover:

- resource validation;
- create/get/update/delete topic;
- duplicate/missing topic;
- list pagination;
- create/get/update/delete subscription;
- referenced missing topic;
- unsupported push/BigQuery/Storage/dead-letter fields;
- list subscriptions;
- list topic subscriptions;
- malformed token;
- cross-project rejection;
- concurrency;
- restart;
- reset/export/import;
- no AWS SNS/SQS collision.

## Official Pub/Sub Go client compatibility

Use the official client against `PUBSUB_EMULATOR_HOST` or explicit local endpoint with no credentials:

- create topic;
- get/update/list/delete topic;
- create subscription;
- get/update/list/delete subscription;
- decoded gRPC errors.

Use a loopback-only dialer.

## Compatibility catalog

Add stable test IDs. Mark only operations proven through official client calls as supported; update methods may remain partial.

## Documentation

Document pull-only scope and every rejected subscription type/policy.

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
