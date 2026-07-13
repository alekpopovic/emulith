# Task 67 — Azure Functions custom-handler HTTP and timer runtime

Target agent: Codex with GPT-5.6  
Prerequisite: Tasks 00–66.

Implement a local Azure Functions-compatible subset for custom-handler containers, focusing on HTTP and timer triggers.

Do not attempt to emulate the full Azure Functions Host or language worker protocols.

## Function declaration

Support manifest configuration:

```yaml
functions:
  azure-api:
    provider: azure
    runtime: custom-handler
    image: emulith/azure-api:dev
    timeout: 30s
    memory: 256MiB
    handler:
      port: 8080
    triggers:
      - type: azure.http
        name: req
        route: api/invoices/{id}
        methods: [GET, POST]
        auth_level: anonymous
```

Timer:

```yaml
      - type: azure.timer
        name: timer
        schedule: "0 */5 * * * *"
```

Strictly validate routes, methods, bindings, timeout, and schedule.

## Custom-handler invocation protocol

Implement the HTTP payload/response contract used by Azure Functions custom handlers closely enough for documented examples/SDK-neutral handlers.

Invocation request should include:

```text
Data
Metadata
request ID/invocation ID
function name
binding data
HTTP request method/url/headers/query/params/body
```

Response should support:

```text
Outputs
Logs
ReturnValue
HTTP status
headers
body
cookies only if modeled
```

Requirements:

- provider-specific Azure envelope kept out of generic runtime;
- bounded JSON/body;
- binary body encoding documented;
- exact route/query/header behavior;
- cancellation/timeout;
- malformed response handled as function error;
- warm container reuse;
- custom handler startup/readiness probe;
- no language-worker protocol claim.

## HTTP routing

Implement local function HTTP routes on a dedicated configurable endpoint or existing Azure function gateway.

Requirements:

- route templates with literals and `{parameter}`;
- no ambiguous duplicate routes;
- method filtering;
- URL decoding exactly once;
- query and headers;
- body streaming/bound;
- response status/headers/body;
- client cancellation;
- function timeout;
- 404 vs 405 behavior;
- route update through new revision;
- function disabled/deleted behavior.

Auth levels:

```text
anonymous
function
admin
```

Only `anonymous` behavior is required. Other levels may be accepted as metadata but must not be described as enforced. Prefer rejecting them unless a clear permissive-development deviation is documented.

## Timer trigger

Parse Azure/NCRONTAB-like six-field schedules for a bounded subset.

Requirements:

- UTC only for v0.5;
- next occurrence calculation;
- durable schedule state;
- `IsPastDue`;
- no duplicate execution after clean restart;
- at-least-once after crash;
- disable/update/delete;
- manual run test hook/CLI where appropriate;
- missed-run policy documented;
- no DST/timezone claim.

Use the internal durable delivery engine.

## Runtime/environment

Provide development environment metadata similar to Azure custom handlers:

```text
FUNCTIONS_WORKER_RUNTIME=custom
FUNCTIONS_CUSTOMHANDLER_PORT
WEBSITE_HOSTNAME or local equivalent only if useful
```

Do not expose secrets automatically.

## Logging

Capture handler logs through generic runtime. Include invocation/function correlation without payloads.

## Tests

Cover:

- route registration/conflict;
- GET/POST/path/query/header/body;
- binary/large body bounds;
- response status/headers;
- malformed handler response;
- timeout/cancel/crash;
- cold/warm;
- timer parsing/next run/past due;
- restart/update/disable;
- function deletion;
- reset/export/import;
- fake runtime and Docker integration;
- no AWS/GCP route collision;
- race/leak tests.

## Compatibility claims

Because this is a custom-handler subset rather than full Azure Functions Host compatibility, document exactly:

```text
supported custom-handler envelope
HTTP trigger subset
timer trigger subset
unsupported language workers and host features
```

No official Azure Functions Core Tools dependency/test is allowed as runtime. Protocol fixture tests and user-handler examples are required.

## Documentation

Add an Azure custom-handler example project/image and clear quickstart.

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
