# Task 71 — Serverless observability: logs, traces, metrics, and inspection

Target agent: Codex with GPT-5.6  
Prerequisite: Tasks 00–70.

Add bounded, local-first observability for functions, invocations, retries, triggers, and runtime instances.

## Goals

Provide enough visibility to debug local event-driven systems without remote telemetry or payload leakage.

## Invocation records

Persist/query:

```text
provider
function
revision
invocation ID
attempt ID
event ID
trigger ID
source resource
status
start/end time
duration
cold start
runtime instance ID
exit code
error category
retry scheduled time
log reference
trace ID
```

Do not store request/response/event bodies in ordinary invocation records.

## CLI

Implement:

```bash
emulith functions logs <function>
emulith functions logs <function> --follow
emulith functions invocations <function>
emulith functions invocation <invocation-id>
emulith functions instances <function>
```

Useful filters:

```text
--since
--until
--status
--event-id
--attempt
--limit
--output=json
```

Requirements:

- deterministic pagination;
- bounded follow buffer;
- cancellation/terminal handling;
- stdout/stderr separation;
- sensitive values redacted;
- no unbounded historical scan;
- missing invocation/function errors.

## Log store

Capture stdout/stderr with:

- timestamp;
- stream;
- invocation attribution where possible;
- sequence/order;
- byte limits per invocation and per function;
- truncation marker;
- retention/cleanup;
- safe UTF-8 display with binary fallback;
- no cross-function log mixing.

If a warm process logs outside an active invocation, classify as runtime/init/background log rather than assigning incorrectly.

## Provider integrations

### AWS

Write Lambda invocation logs to the existing CloudWatch Logs subset:

```text
/aws/lambda/{function-name}
```

Include local `START`, `END`, and `REPORT`-like lines only when clearly labeled as Emulith-local behavior. Do not claim exact billing metrics.

### Azure

Expose structured custom-handler invocation records with correlation IDs. Do not claim Application Insights API compatibility.

### GCP

Support local structured JSON log ingestion/display compatible with common Functions Framework stdout patterns. Do not claim complete Cloud Logging API compatibility.

## Tracing

Support W3C Trace Context:

```text
traceparent
tracestate
baggage with strict bounds/redaction
```

Propagate across:

```text
HTTP invocation
queue/topic event delivery
retry
output binding
cross-service trigger
```

Requirements:

- generate trace/span IDs when absent;
- preserve trace ID across retries, new span per attempt;
- no payload attributes by default;
- cap attribute count/length;
- redaction;
- correlation visible in CLI.

## OpenTelemetry

Optional OTLP export:

```text
disabled by default
explicit endpoint required
loopback-only by default
```

Requirements:

- no network call when disabled;
- no automatic discovery;
- bounded queue;
- backpressure/drop counter;
- shutdown flush with deadline;
- TLS/insecure explicit;
- test exporter;
- no secrets/payloads in attributes.

Provide a local in-memory/file exporter for tests/development if useful.

## Metrics

Expose local metrics through:

```text
GET /_emulith/metrics
```

Prometheus text format or existing metrics convention.

Metrics include:

```text
emulith_function_invocations_total
emulith_function_invocation_duration_seconds
emulith_function_cold_starts_total
emulith_function_failures_total
emulith_function_timeouts_total
emulith_function_retries_total
emulith_function_active_instances
emulith_function_queued_events
emulith_function_dlq_events_total
emulith_function_log_bytes_total
```

Requirements:

- bounded label cardinality;
- function/provider/revision labels only where safe;
- never event/invocation ID as metric label;
- no payload/user property labels;
- metrics endpoint documented local-only;
- race-safe counters;
- reset behavior documented.

## Retention

Add configurable retention for:

```text
invocation metadata
logs
attempt history
terminal event deliveries
```

Cleanup:

- bounded batches;
- does not remove active records;
- restart-safe;
- snapshot policy documented;
- no object/body state damage.

## Tests

Cover:

- invocation/log attribution;
- init/background logs;
- truncation;
- follow/cancel;
- filters/pagination;
- redaction;
- provider log integrations;
- trace propagation across retries/triggers;
- OTLP disabled means no network;
- test exporter;
- metric correctness/cardinality;
- concurrent writes/readers;
- retention cleanup;
- reset/export/import;
- race/leak tests.

## Documentation

Add a local observability guide and privacy/redaction policy.

## Compatibility catalog

Observability APIs are Emulith platform features. Provider log integrations must be labeled partial/local where applicable.

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
