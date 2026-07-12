# Task 04 — Implement STS `GetCallerIdentity`

Target agent: Codex with GPT-5.6  
Prerequisite: Tasks 00–03.

Implement the first real AWS-compatible operation: STS `GetCallerIdentity`.

## Supported operation

Support the AWS Query protocol operation:

```text
Action=GetCallerIdentity
Version=2011-06-15
```

Accept both:

- query-string `GET`;
- form-encoded `POST`.

Ignore local credential validation, but do not expose or log supplied credentials.

## Response

Return an XML response compatible with standard AWS clients using these deterministic values:

```text
UserId: EMULITHUSER
Account: 000000000000
Arn: arn:aws:iam::000000000000:user/emulith
```

Requirements:

- HTTP `200`;
- provider-compatible XML namespace and result/metadata envelope;
- generated request ID in response metadata and header where appropriate;
- `Content-Type` suitable for XML;
- stable UTF-8 encoding.

Unsupported STS actions must return a provider-shaped error such as `InvalidAction`, not a generic HTML or JSON error.

Malformed requests must return a deterministic client error.

## SDK compatibility test

Add an end-to-end test using the current AWS SDK for Go v2 STS client:

- start an in-process Emulith server on an `httptest` listener;
- configure an explicit loopback endpoint;
- use static fake credentials;
- set region `us-east-1`;
- disable EC2 metadata and any external credential/profile fallback;
- call `GetCallerIdentity`;
- assert the exact account, ARN, and user ID;
- assert that no real network host can be contacted.

Use the current non-deprecated endpoint configuration supported by the pinned SDK version. Do not weaken the test into a hand-written HTTP request only; keep both handler-level and SDK-level tests.

## Documentation

Update the compatibility matrix:

```text
STS GetCallerIdentity — supported
All other STS operations — unsupported
IAM enforcement — not implemented
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
