# Task 69 — GCP Functions Framework and Cloud Run-style local functions

Target agent: Codex with GPT-5.6  
Prerequisite: Tasks 00–68.

Implement image-based GCP local functions compatible with common Functions Framework HTTP and CloudEvents handler conventions.

Do not claim complete Cloud Functions or Cloud Run Admin API compatibility.

## Function types

Support:

```text
HTTP function
CloudEvent function
```

Manifest examples:

```yaml
functions:
  gcp-http:
    provider: gcp
    runtime: functions-framework
    signature_type: http
    image: emulith/gcp-http:dev
    port: 8080
```

```yaml
functions:
  gcp-event:
    provider: gcp
    runtime: functions-framework
    signature_type: cloudevent
    image: emulith/gcp-event:dev
    port: 8080
```

Validate provider, signature type, port, timeout, concurrency, memory, and environment.

## Runtime contract

Start the container with:

```text
PORT=<configured>
K_SERVICE=<function name>
K_REVISION=<revision ID>
K_CONFIGURATION=<function name>
GOOGLE_CLOUD_PROJECT=emulith-local
```

Optional Functions Framework environment fields may be passed only when standard and safe.

Requirements:

- readiness/startup probe;
- cold/warm reuse;
- configurable concurrency per instance;
- maximum/minimum local instances metadata;
- timeout/cancellation;
- container crash recovery;
- bounded response;
- no public internet requirement;
- provider service network access;
- no Cloud Run production sandbox claim.

## HTTP functions

Expose a local invocation URL.

Forward:

```text
method
path
query
headers
body
client cancellation
trace context
```

Requirements:

- stream request/response only with bounded controls;
- preserve status/headers;
- remove hop-by-hop headers;
- no arbitrary proxy target;
- function route cannot shadow admin/cloud provider endpoints;
- request ID/execution ID;
- timeout maps to safe local HTTP error;
- disabled/missing function behavior.

## CloudEvents functions

Support CloudEvents 1.0 HTTP binding:

```text
binary mode
structured JSON mode
```

Required attributes:

```text
specversion
id
source
type
time when supplied
subject when supplied
datacontenttype
data
```

Requirements:

- validate attributes;
- preserve extensions with bounds;
- stable event ID across retries;
- convert provider trigger events in Task 70;
- no payload logging;
- handler success/failure determined from HTTP response;
- non-2xx classified for retry according to trigger policy.

## Local control plane

Implement CLI and/or a small documented local API:

```bash
emulith functions deploy <name>
emulith functions get <name>
emulith functions list
emulith functions delete <name>
emulith functions invoke <name>
```

Reuse generic definitions/revisions. Do not fabricate full Google Cloud Functions or Cloud Run REST APIs unless a precise official-client subset is intentionally implemented and tested.

## Test images

Provide tiny Functions Framework-compatible fixtures for at least:

- HTTP echo;
- CloudEvent echo/record;
- failure;
- timeout;
- concurrent requests;
- log output;
- warm instance counter.

Fixtures may use Go or another language only if dependencies are pinned and builds remain hermetic after download.

## Tests

Cover:

- HTTP forwarding;
- CloudEvents binary/structured;
- attribute validation;
- extensions;
- non-2xx retry classification;
- timeout/cancel;
- cold/warm;
- per-instance concurrency;
- scaling limit;
- revision update;
- crash/recovery;
- logs;
- loopback-only invocation;
- reset/export/import;
- Docker unavailable fake runtime;
- race/leak tests.

## Compatibility catalog and docs

State clearly:

- Functions Framework-compatible HTTP contract;
- CloudEvents 1.0 subset;
- no Cloud Functions/Cloud Run Admin API parity;
- no IAM/identity enforcement;
- image-based local execution only.

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
