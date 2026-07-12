# Task 03 — Implement the AWS gateway and protocol router

Target agent: Codex with GPT-5.6  
Prerequisite: Tasks 00–02.

Add a single-endpoint AWS-compatible gateway that detects STS, S3, and SQS requests and dispatches them to service handlers.

## Outcome

Requests on Emulith's root endpoint are classified consistently, admin routes always win, unsupported operations return provider-shaped errors, and each service can be implemented independently in later tasks.

## Architecture

Create a package under `providers/aws` containing:

- a router/gateway;
- request classification;
- request ID generation;
- common AWS response/error helpers;
- narrow interfaces for STS, S3, and SQS handlers.

Inject the state store and logger. Avoid package globals.

Routing precedence:

1. `/_emulith/*` admin routes;
2. SQS AWS JSON protocol;
3. STS/SQS query protocol;
4. S3 REST-style requests;
5. generic AWS-style unsupported response.

## Detection rules

### SQS AWS JSON protocol

Detect requests such as:

```http
POST /
X-Amz-Target: AmazonSQS.CreateQueue
Content-Type: application/x-amz-json-1.0
```

Accept case-insensitive header matching and media-type parameters.

### Query protocol

Detect `Action` and `Version` from:

- URL query parameters;
- `application/x-www-form-urlencoded` POST bodies.

Classify known STS actions separately from SQS actions. Reading a form body must not make it unavailable to the selected handler; parse once and pass a structured request or safely restore the body.

### S3 REST protocol

Classify remaining plausible S3 requests by method/path/query/header shape, including:

- `GET /`;
- bucket paths;
- object paths;
- `list-type=2`;
- `x-amz-*` headers.

Do not rely on hostname-based bucket routing yet. The POC supports path-style access only.

## Common behavior

- Ignore SigV4 credential validation for the POC, but accept signed requests without breaking body handling.
- Never log `Authorization`, security tokens, object bodies, or message bodies.
- Add response request IDs:
  - `x-amz-request-id` for S3;
  - `x-amzn-RequestId` for JSON/query services where appropriate.
- Add structured access logging with method, sanitized path, detected provider/service/operation, status, duration, and request ID.
- Bound form/JSON parsing sizes.
- Return `405` or provider-shaped unsupported-operation errors consistently rather than panicking.

## Error helpers

Implement:

- S3 XML error envelope;
- AWS Query XML error envelope for STS and legacy SQS query requests;
- AWS JSON error envelope for SQS JSON requests.

Include request IDs and correct content types.

## Placeholder handlers

Create explicit placeholder handlers for services not yet implemented. They must return `NotImplemented`/`InvalidAction` style errors and must not be marked supported in documentation.

## Tests

Test classification and response shape for at least:

- `X-Amz-Target: AmazonSQS.CreateQueue` -> SQS JSON;
- form `Action=CreateQueue` -> SQS Query;
- `Action=GetCallerIdentity` -> STS;
- `PUT /my-bucket` -> S3 bucket request;
- `PUT /my-bucket/a.txt` -> S3 object request;
- `GET /my-bucket?list-type=2` -> S3 ListObjectsV2;
- admin health route is not swallowed by AWS routing;
- signed request headers are accepted but not logged;
- malformed oversized form/JSON request is rejected safely;
- unknown action returns the protocol-appropriate error format.

## Required verification

```bash
gofmt -w <changed-go-files>
go test ./...
go vet ./...
make build
```

## Execution contract

You are the implementation agent for this task. Complete the task in the current repository; do not stop after producing a plan.

1. Read every applicable `AGENTS.md` before changing files.
2. Inspect the repository, existing conventions, current tests, and dependency versions.
3. Briefly state the implementation plan, then execute it immediately.
4. Make reasonable assumptions when details are non-blocking. Do not ask for confirmation merely to choose between equivalent implementation details.
5. Keep the change tightly scoped to this task. Do not implement later roadmap items.
6. Preserve working behavior and public interfaces unless this task explicitly changes them.
7. Prefer simple, readable Go over speculative abstractions.
8. Do not use LocalStack, Moto, MinIO, ElasticMQ, Azurite, or another cloud emulator as a runtime dependency.
9. Never contact real AWS, Azure, or GCP endpoints. Tests must be hermetic.
10. Do not add accounts, license keys, forced telemetry, analytics, or phone-home behavior.
11. Do not commit, push, create a tag, or open a pull request.
12. Format all changed Go files and run the required verification commands.
13. Fix failures caused by your changes before finishing. If an environment limitation prevents a command, report the exact limitation and run the closest safe verification.
14. Finish with:
    - implementation summary;
    - important design decisions;
    - changed files;
    - commands run and their results;
    - remaining limitations that are genuinely outside this task.
