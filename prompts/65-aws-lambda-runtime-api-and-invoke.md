# Task 65 — AWS Lambda Runtime API and Invoke

Target agent: Codex with GPT-5.6  
Prerequisite: Tasks 00–64.

Implement image-based Lambda-compatible execution using the AWS Lambda Runtime API and the OCI runtime.

## Supported control-plane operation

Implement:

```text
Invoke
```

Invocation types:

```text
RequestResponse
Event
DryRun
```

Support `$LATEST` only unless qualifier/version support already exists and is complete.

## Runtime API

Expose a per-runtime-instance local endpoint implementing:

```text
GET  /2018-06-01/runtime/invocation/next
POST /2018-06-01/runtime/invocation/{request-id}/response
POST /2018-06-01/runtime/invocation/{request-id}/error
POST /2018-06-01/runtime/init/error
```

The container receives:

```text
AWS_LAMBDA_RUNTIME_API
AWS_LAMBDA_FUNCTION_NAME
AWS_LAMBDA_FUNCTION_MEMORY_SIZE
AWS_LAMBDA_FUNCTION_VERSION
AWS_LAMBDA_LOG_GROUP_NAME
AWS_LAMBDA_LOG_STREAM_NAME
AWS_REGION
AWS_DEFAULT_REGION
```

Requirements:

- Runtime API reachable only through the managed function network;
- invocation IDs cannot cross instances/functions;
- long-poll `next` is cancellable;
- only one outstanding invocation for a single-concurrency runtime;
- bounded request/response/error payload;
- no provider payload logging;
- malformed/duplicate response is rejected safely.

## Invocation headers

Return relevant Runtime API headers:

```text
Lambda-Runtime-Aws-Request-Id
Lambda-Runtime-Deadline-Ms
Lambda-Runtime-Invoked-Function-Arn
Lambda-Runtime-Trace-Id when present
Lambda-Runtime-Client-Context when supported
Lambda-Runtime-Cognito-Identity when supported
```

Do not fabricate identity/client context unless supplied.

## Initialization

Implement:

- container startup;
- runtime waits on `/next`;
- init timeout;
- `init/error`;
- state transition to ready/failed;
- cold-start flag;
- failed init container removal;
- retry policy for infrastructure failure;
- no endless restart loop.

## RequestResponse

- send exact invocation payload;
- wait for runtime response/error;
- enforce function timeout;
- cancellation on client disconnect where safe;
- return exact response bytes;
- set `X-Amz-Function-Error` for function errors;
- include `ExecutedVersion=$LATEST`;
- return provider-shaped errors for missing/inactive function/runtime failure;
- response-size limit;
- warm instance reuse.

## Event

- validate request;
- enqueue through durable delivery engine;
- return accepted status promptly;
- event ID/correlation retained;
- retries and DLQ policy use function configuration/trigger defaults;
- do not run synchronously before returning;
- restart preserves queued work.

## DryRun

- validate function/state/permissions metadata only;
- do not start a container;
- return compatible success or validation error.

## Logs

Capture stdout/stderr.

Support `LogType=Tail` for RequestResponse:

- last bounded bytes;
- base64 in response header;
- deterministic truncation;
- no unbounded memory;
- sensitive payload is not automatically echoed by Emulith, though user function logs remain user-controlled.

Forward Lambda logs to the existing CloudWatch Logs subset when enabled:

```text
/aws/lambda/{function-name}
```

Create deterministic local stream names.

## Failure classification

Distinguish:

```text
Handled function error
Unhandled runtime error
Init error
Container crash
Timeout
Cancellation
Invalid runtime response
Runtime unavailable
```

Map to Lambda invoke/control-plane behavior and retry categories.

## Test function image

Add a tiny custom runtime fixture image capable of:

- echo success;
- handled error;
- init error;
- timeout;
- crash;
- log output;
- state counter proving warm reuse.

## Tests

Cover:

- Runtime API next/response/error/init-error;
- sync echo;
- binary/JSON opaque payload;
- handled/unhandled error;
- timeout;
- crash;
- client cancellation;
- cold/warm;
- revision change drains old pool;
- LogType Tail;
- async accepted/retry/restart;
- DryRun no container;
- response/body limits;
- concurrent invocations and reserved concurrency;
- shutdown;
- Docker integration when available;
- fake runtime otherwise;
- race/leak tests.

## Official AWS SDK compatibility

Use the Lambda SDK client:

- RequestResponse;
- Event;
- DryRun;
- FunctionError decoding;
- LogResult;
- missing/inactive errors.

## Compatibility catalog and docs

Add stable Invoke/Runtime API IDs and exact custom-runtime limitations.

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
