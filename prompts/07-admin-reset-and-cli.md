# Task 07 — Implement safe admin reset and the `emulith reset` CLI

Target agent: Codex with GPT-5.6  
Prerequisite: Tasks 00–06.

Expose the state reset capability through an admin endpoint and CLI.

## Admin endpoint

Implement:

```http
POST /_emulith/reset
```

Success:

```json
{
  "status": "ok",
  "reset": true
}
```

Requirements:

- `Content-Type: application/json`;
- no cache;
- only `POST`; other methods return a clear method error;
- call the state store's concurrency-safe reset;
- do not route through AWS protocol detection;
- do not leak filesystem or SQL details;
- log only request ID, result, and duration.

For the POC, the endpoint may bind to the same listener, but document that it is unauthenticated and intended only for trusted local development networks. Do not add fake security.

## CLI command

Implement:

```text
emulith reset
```

Flags/environment:

```text
--endpoint
EMULITH_ENDPOINT
default: http://localhost:4566
```

Do not reuse `EMULITH_ADDR` as a client URL because `:4566` is a listen address, not a complete endpoint.

CLI requirements:

- construct the exact admin URL safely;
- use an HTTP client with a timeout;
- send `POST`;
- verify status and JSON response;
- print a concise success message;
- print actionable errors to stderr;
- return non-zero on connection error, non-2xx response, malformed response, or reset failure;
- support dependency injection for tests;
- never silently target a non-local host unless the user explicitly passed it via flag/environment.

## Tests

Cover:

- reset removes S3 buckets and objects;
- reset removes SQS queues and messages;
- state remains usable immediately after reset;
- managed files are removed but external/symlink targets are untouched;
- concurrent request behavior is safe;
- `GET /_emulith/reset` is rejected;
- CLI default endpoint behavior;
- CLI explicit endpoint;
- CLI handles server error and malformed JSON;
- output goes to the correct writer;
- no fixed ports.

## Documentation

Add a reset section with warning about destructive local state and examples:

```bash
emulith reset
emulith reset --endpoint http://localhost:4566
```

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
