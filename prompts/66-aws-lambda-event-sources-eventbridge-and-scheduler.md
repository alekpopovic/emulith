# Task 66 — AWS Lambda event sources, EventBridge, and scheduler

Target agent: Codex with GPT-5.6  
Prerequisite: Tasks 00–65.

Connect existing Emulith AWS services to Lambda functions through provider-compatible event envelopes and durable delivery.

## Part A — SQS event source mappings

Implement:

```text
CreateEventSourceMapping
GetEventSourceMapping
ListEventSourceMappings
UpdateEventSourceMapping
DeleteEventSourceMapping
```

Support source:

```text
standard SQS queue ARN
```

Fields:

```text
FunctionName
EventSourceArn
Enabled
BatchSize
MaximumBatchingWindowInSeconds with a small bounded subset
FunctionResponseTypes including ReportBatchItemFailures
```

Reject FIFO, streams not implemented, tumbling windows, destinations, filters unless explicitly added.

### Poller behavior

- leases SQS messages using existing visibility semantics;
- builds Lambda SQS event envelope;
- invokes pinned/active revision according to documented policy;
- deletes messages only after successful handling;
- on function failure, messages become visible after timeout;
- supports partial batch response:

```json
{
  "batchItemFailures": [
    {"itemIdentifier": "message-id"}
  ]
}
```

- validates duplicate/unknown item identifiers;
- successful items deleted, failed items retained;
- no duplicate simultaneous batch ownership;
- mapping disabled/deleted stops new polling;
- restart resumes mappings.

## Part B — SNS Lambda subscriptions

Allow SNS topic subscriptions with Lambda protocol/endpoint.

Requirements:

- function ARN validation;
- SNS notification event envelope;
- async durable invocation;
- retry/DLQ policy;
- topic deletion/function deletion cleanup;
- no IAM permission simulation;
- one event per publish per subscription.

Extend existing SNS APIs only as needed and test with official SNS/Lambda clients.

## Part C — S3 notifications

Implement a documented subset of bucket notification configuration, using the public S3 API shape if practical:

```text
s3:ObjectCreated:Put
s3:ObjectRemoved:Delete
prefix filter
suffix filter
Lambda target
SQS target where existing support permits
SNS target where existing support permits
```

Requirements:

- event generated only after object mutation commits;
- no event on failed/conditional write;
- S3 event JSON compatible with Lambda consumption;
- URL-encoded object key behavior documented/tested;
- event ID stable through retries;
- configuration persistence/restart;
- deletion cleanup.

## Part D — DynamoDB Streams experimental subset

Implement only if feasible within this task without destabilizing DynamoDB.

Potential operations:

```text
DescribeStream
GetShardIterator
GetRecords
ListStreams
```

And Lambda event source mapping from a table stream.

Requirements if implemented:

- per-table append-only change records;
- INSERT/MODIFY/REMOVE;
- sequence ordering;
- NEW/OLD image according to configured view type;
- checkpoint per mapping;
- batch retry;
- trim/retention;
- official SDK/event tests.

If not fully implemented, leave it explicitly unsupported. Do not add placeholder “supported” claims.

## Part E — EventBridge subset

Implement:

```text
PutRule
DeleteRule
ListRules
DescribeRule
PutTargets
RemoveTargets
ListTargetsByRule
PutEvents
EnableRule
DisableRule
```

Support targets:

```text
Lambda
SQS
SNS
```

Event pattern subset:

```text
source
detail-type
resources
detail field exact match
prefix match
anything-but only if complete
exists only if complete
```

Build a bounded structured matcher, not regex substitution.

Requirements:

- enabled rules only;
- event envelope preserves ID/source/time/region/account/detail;
- durable target delivery;
- target-level retry/DLQ where modeled;
- pagination;
- no archive/replay/pipes/API destinations/global endpoints.

## Part F — local scheduler

Add scheduler rules through EventBridge-style schedule expressions or a small provider API subset.

Support:

```text
rate(value minute|minutes|hour|hours|day|days)
limited cron(...)
one-time local schedule if provider API supports it
UTC only initially
```

Requirements:

- real parser with bounds;
- injectable clock;
- durable next-run state;
- no duplicate fire after restart;
- at-least-once documented;
- pause/disable;
- missed-run policy documented;
- no full timezone/DST claim.

## Tests

Cover every implemented control-plane operation and event path:

- SQS success/failure/partial batch;
- mapping update/disable/delete/restart;
- SNS trigger;
- S3 create/delete/filter;
- EventBridge rule/pattern/targets;
- scheduler rate/cron/restart;
- DLQ;
- target/function deletion;
- concurrent pollers;
- reset/export/import;
- official SDK decoded errors;
- race/leak tests.

## Compatibility catalog

Separate:

```text
Lambda event source mapping control plane
event envelope
delivery semantics
retry/partial failure
EventBridge control plane
scheduler expression subset
```

Do not mark DynamoDB Streams supported without complete tests.

## Verification

```bash
gofmt -w <changed-go-files>
go test ./...
go test -race ./...
go vet ./...
make compatibility
make compatibility-check
make build
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
