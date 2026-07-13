# Task 00 — Create the repository-level `AGENTS.md`

Target agent: Codex with GPT-5.6  
Project: Emulith

Create a concise but complete repository-level `AGENTS.md` that Codex will automatically use for all later tasks.

## Project definition

Emulith is a from-scratch, Docker-first, open-source local cloud emulator. Its purpose is to let applications and CI use standard cloud SDKs against local endpoints without requiring a cloud account, license key, proprietary runtime, or closed-core component.

The first POC is AWS-compatible only. Azure and GCP are later roadmap items, but the architecture must not make them impossible.

## Required content for `AGENTS.md`

Include the following sections.

### Mission and non-negotiable principles

- Core implementation remains open source.
- No LocalStack dependency or copied LocalStack source.
- No account, activation token, license key, or online entitlement check.
- No forced telemetry or phone-home behavior.
- Docker-first local development and CI.
- Compatibility with standard provider SDKs is preferred over creating an Emulith-specific SDK.
- Emulith is for development and CI, never production.
- Emulate only documented public API behavior and independently written implementation logic.
- Do not claim full AWS parity.

### Current POC scope

The target for `v0.1.0-poc` is:

- one Go binary named `emulith`;
- HTTP server on `:4566` by default;
- health and reset admin endpoints;
- AWS STS `GetCallerIdentity`;
- AWS S3 basic bucket/object operations;
- AWS SQS standard queue operations;
- SQLite metadata and filesystem object bodies;
- Docker image and Compose example;
- AWS SDK for Go v2 compatibility tests;
- manifest-based creation of an S3 bucket and SQS queue;
- no Azure, GCP, Lambda, DynamoDB, SNS, dashboard, IAM enforcement, or production mode.

### Technical baseline

- Go.
- Cobra CLI.
- Standard `net/http` unless an already-adopted lightweight router clearly improves the code.
- `log/slog` for logging.
- Pure-Go SQLite driver so the default build does not require CGO.
- YAML configuration where needed.
- AWS SDK for Go v2 only in compatibility/client code.
- Metadata in SQLite; object bytes on the filesystem.
- Dependency versions must be pinned by `go.mod`/`go.sum`.

### Repository expectations

Document the intended layout:

```text
cmd/emulith/
internal/cli/
internal/config/
internal/server/
internal/state/
internal/manifest/
providers/aws/
providers/aws/s3/
providers/aws/sqs/
providers/aws/sts/
test/compatibility/aws/
examples/
scripts/
docs/
```

State that existing repository conventions take precedence if the layout evolves for a concrete reason.

### Engineering rules

- Keep packages cohesive and avoid circular dependencies.
- Avoid global mutable state.
- Pass dependencies explicitly.
- Use contexts for I/O and request lifetimes.
- Close response bodies, rows, files, and database resources.
- Use transactions for multi-step metadata mutations.
- Make state reset concurrency-safe.
- Protect all filesystem operations against traversal and symlink escape.
- Do not log credentials, authorization headers, message bodies, or object bodies.
- Return deterministic, provider-shaped errors where compatibility requires them.
- Generate request IDs locally.
- Public identifiers and persisted state formats need tests before being treated as stable.
- No placeholder implementation may be marked supported in the compatibility matrix.

### Testing and verification

Define the standard commands:

```bash
gofmt -w <changed-go-files>
go test ./...
go vet ./...
make build
make compatibility
make docker-build
make demo
```

Explain that a task should run only the commands applicable to the current repository state, but `go test ./...` is mandatory after Go changes.

Require:

- unit tests for parsers and state;
- handler tests using `httptest`;
- end-to-end compatibility tests using real AWS SDK for Go v2 clients pointed only at loopback;
- no user AWS profile, EC2 metadata, environment credential chain, or real cloud fallback;
- race tests during hardening/CI where practical.

### Completion reporting

Require every Codex task to finish with:

- concise summary;
- design decisions;
- changed files;
- exact verification commands and outcomes;
- known limitations.

### Licensing and contribution constraints

- Core license target: `AGPL-3.0-or-later`.
- Do not add a CLA that assigns copyright or grants special relicensing rights.
- Contributions use DCO sign-off once contribution docs exist.
- Do not introduce code with unclear or incompatible licensing.
- Do not copy provider documentation examples wholesale when a clean-room implementation is sufficient.

## Scope guard

This task creates only `AGENTS.md`. Do not bootstrap the Go project or add product code yet.

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
