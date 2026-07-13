# Task 70 — GCP service triggers, Eventarc subset, and Cloud Scheduler

Target agent: Codex with GPT-5.6  
Prerequisite: Tasks 00–69.

Connect Pub/Sub, Cloud Storage, and Firestore to CloudEvent functions, then implement local Eventarc and Cloud Scheduler control-plane subsets.

## Part A — Pub/Sub trigger

Manifest:

```yaml
triggers:
  - type: gcp.pubsub
    topic: invoice-events
    retry:
      max_attempts: 5
```

CloudEvent must include a Pub/Sub message envelope with:

```text
message.data base64
message.attributes
message.messageId
message.publishTime
message.orderingKey
subscription
```

Requirements:

- create/use a managed local subscription;
- acknowledge only after successful function response;
- failure/timeout causes redelivery;
- stable event/message ID;
- delivery attempt metadata;
- ordering behavior inherited from Pub/Sub;
- trigger deletion cleans managed subscription safely;
- no push/exactly-once claim;
- DLQ adapter optional/configurable.

## Part B — Cloud Storage trigger

Support event types:

```text
google.cloud.storage.object.v1.finalized
google.cloud.storage.object.v1.deleted
google.cloud.storage.object.v1.metadataUpdated
```

Event data includes modeled object resource fields:

```text
bucket
name
generation
metageneration
size
contentType
timeCreated
updated
```

Requirements:

- emit only after committed mutation;
- resumable session emits only on finalization;
- no event on failed precondition;
- prefix/suffix filter;
- retry/DLQ;
- stable CloudEvent ID;
- object body not embedded by default.

## Part C — Firestore trigger

Support:

```text
document created
document updated
document deleted
document written
```

Manifest path pattern:

```text
users/{userId}/orders/{orderId}
```

Requirements:

- validate alternating collection/document pattern;
- bind wildcard values;
- old/new document snapshots;
- update mask/changed fields;
- event emitted after commit;
- multi-document commit emits one event per changed document with shared correlation/commit metadata;
- transaction rollback emits nothing;
- retry/DLQ;
- no recursive trigger storm beyond configured lineage depth.

## Part D — Eventarc subset

Implement local control-plane operations equivalent in responsibility to:

```text
CreateTrigger
GetTrigger
ListTriggers
DeleteTrigger
```

Trigger fields:

```text
name
location metadata
event type
matching criteria
destination function
service account metadata not enforced
retry policy
```

Supported filters:

```text
type
source/service name
subject/resource prefix
exact attribute match
```

Requirements:

- local project only;
- deterministic pagination;
- immutable fields validated;
- no channels, partners, audit-log triggers, regional routing, or IAM enforcement;
- provider service trigger definitions may map internally to Eventarc records but public semantics remain documented.

## Part E — Cloud Scheduler subset

Implement:

```text
CreateJob
GetJob
ListJobs
UpdateJob
DeleteJob
RunJob
PauseJob
ResumeJob
```

Targets:

```text
HTTP function
Pub/Sub topic
```

Schedule:

```text
five-field Unix cron subset
UTC timezone only
```

Optional rate helper may exist in manifest, but public job model remains clear.

Requirements:

- real bounded parser;
- next-run calculation;
- durable state;
- no duplicate normal run after restart;
- at-least-once after crash;
- manual `RunJob`;
- pause/resume;
- bounded HTTP payload/headers;
- no OAuth/OIDC tokens;
- no App Engine targets;
- retry/backoff through durable delivery engine;
- missed-run policy documented.

## Tests

Cover:

- Pub/Sub trigger success/failure/redelivery/order;
- Storage finalized/deleted/metadata event/filter;
- Firestore create/update/delete/write/wildcards/transaction rollback;
- Eventarc lifecycle/filtering/pagination;
- Scheduler lifecycle/cron/manual/pause/restart;
- HTTP and Pub/Sub targets;
- DLQ;
- trigger/function/resource deletion;
- recursive depth guard;
- mixed triggers;
- reset/export/import;
- race/leak tests.

## Official client/protocol testing

Use official GCP clients where practical for source service operations and Scheduler/Eventarc public APIs only if implemented. CloudEvent function delivery must use real runtime fixtures.

## Compatibility catalog

Separate source event type, envelope, trigger control plane, retry behavior, and scheduler subset. State all unsupported production semantics.

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
