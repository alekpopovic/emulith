# Task 68 — Azure Functions bindings and Event Grid subset

Target agent: Codex with GPT-5.6  
Prerequisite: Tasks 00–67.

Connect existing Azure Storage services to custom-handler functions and implement a local Event Grid-compatible subset.

## Part A — Azure Queue trigger

Manifest example:

```yaml
triggers:
  - type: azure.queue
    queue: invoice-events
    batch_size: 1
    visibility_timeout: 30s
    max_dequeue_count: 5
    poison_queue: invoice-events-poison
```

Requirements:

- consume using existing Azure Queue visibility/pop-receipt semantics;
- construct custom-handler binding data;
- successful invocation deletes message;
- failure/timeout leaves message for redelivery;
- increment dequeue count;
- after maximum dequeues, move to poison queue durably;
- no message loss if poison delivery fails;
- disabled/deleted trigger stops new leases;
- restart resumes safely;
- no simultaneous delivery of one pop receipt.

## Part B — Blob trigger

Support events:

```text
blob created
blob overwritten
blob metadata updated only if configured
blob deleted only if configured
```

Manifest:

```yaml
triggers:
  - type: azure.blob
    container: invoices
    prefix: incoming/
    suffix: .json
```

Requirements:

- emit after committed mutation;
- event contains account/container/blob name, ETag, length, content type, timestamps, and URI;
- body delivered as bounded bytes or reference according to configured binding mode;
- no trigger on failed/conditional upload;
- retry/DLQ;
- restart-safe event ID;
- no exactly-once claim;
- staged blocks do not trigger until committed.

## Part C — output bindings

Allow a custom-handler response to request bounded outputs:

```text
Azure Queue message
Azure Blob content/metadata
Azure Table entity
HTTP response
```

Requirements:

- validate all outputs before mutation where possible;
- define local atomicity across multiple services; if not atomic, record accurate partial failure and retry behavior;
- prevent recursive trigger storms with no guard; preserve event lineage and max-hop/depth limit;
- no arbitrary account/endpoint output;
- use public service-layer interfaces, not SQL coupling;
- enforce payload/resource limits.

## Part D — Event Grid control plane

Implement a local subset:

```text
CreateTopic
GetTopic
ListTopics
DeleteTopic
CreateEventSubscription
GetEventSubscription
ListEventSubscriptions
DeleteEventSubscription
PublishEvents
```

Use a provider-specific local HTTP/REST shape documented by the compatibility matrix. If emulating Azure Resource Manager control-plane paths is excessive, expose only the documented Event Grid data-plane/public subset and mark control-plane partial.

Supported subscription targets:

```text
Azure custom-handler function
Azure Queue
local HTTP endpoint only when explicitly loopback
```

## Event schemas

Support:

```text
Event Grid event schema
CloudEvents 1.0 JSON
```

Validate:

```text
id
eventType/type
subject
eventTime/time
dataVersion
data
source
specversion
```

Preserve payload as opaque bounded JSON.

## Filtering

Support:

```text
included event types
subject begins with
subject ends with
exact field match only if implemented safely
```

No full advanced filter language.

## Retry and dead letter

- use durable delivery engine;
- configurable attempts/backoff;
- dead-letter to Azure Blob container;
- stable event ID across retries;
- delivery attempt metadata;
- no production webhook validation handshake claim;
- loopback-only HTTP target by default.

## Tests

Cover:

- queue trigger success/failure/poison;
- stale receipt/concurrent workers;
- blob create/overwrite/delete/filter;
- output bindings;
- recursive event depth guard;
- Event Grid topic/subscription lifecycle;
- both event schemas;
- filters;
- function/queue/HTTP target;
- retry/DLQ;
- restart;
- reset/export/import;
- malformed/oversized events;
- race/leak tests.

## Compatibility catalog

Separate function bindings, Event Grid event schema, target types, and retry semantics. Mark Azure Functions compatibility as custom-handler subset.

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
