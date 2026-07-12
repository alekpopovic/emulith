# Emulith Repository Instructions

## Mission and non-negotiable principles

Emulith is a from-scratch, Docker-first, open-source local cloud emulator for development and CI, never production. Applications should use standard provider SDKs against local endpoints; prefer that compatibility over an Emulith-specific SDK. The first POC is AWS-compatible, while the architecture must leave room for later Azure and GCP support.

- Keep the core implementation open source. Do not depend on LocalStack or copy LocalStack source.
- Require no account, activation token, license key, proprietary runtime, closed-core component, or online entitlement check.
- Add no forced telemetry, analytics, or phone-home behavior.
- Keep local development and CI Docker-first.
- Emulate only documented public API behavior with independently written implementation logic.
- Do not claim full AWS parity.

## Current POC scope

The `v0.1.0-poc` target is one Go binary named `emulith`, serving HTTP on `:4566` by default, with:

- health and reset admin endpoints;
- AWS STS `GetCallerIdentity`;
- basic AWS S3 bucket and object operations;
- standard AWS SQS queue operations;
- SQLite metadata and filesystem-backed object bodies;
- a Docker image and Compose example;
- compatibility tests using AWS SDK for Go v2; and
- manifest-based creation of an S3 bucket and SQS queue.

Azure, GCP, Lambda, DynamoDB, SNS, a dashboard, IAM enforcement, and production mode are outside this POC.

## Technical baseline

- Use Go, Cobra for the CLI, and `log/slog` for logging.
- Prefer standard `net/http`; use an already-adopted lightweight router only when it clearly improves the code.
- Use a pure-Go SQLite driver so default builds do not require CGO.
- Use YAML configuration where needed.
- Use AWS SDK for Go v2 only in compatibility tests and other client code, not in provider implementation code.
- Store metadata in SQLite and object bytes on the filesystem.
- Pin dependency versions through `go.mod` and `go.sum`.

## Repository expectations

The intended layout is:

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

If this layout evolves for a concrete reason, established repository conventions take precedence.

## Prompt execution workflow

- Treat the numbered Markdown files in `prompts/` as the ordered implementation plan. Execute them in numeric filename order unless the user explicitly selects another prompt or changes the order.
- Before starting, read the selected prompt completely and check `prompts/LAST_EXECUTED.md` to avoid accidentally repeating completed work.
- After a prompt is implemented and its applicable verification passes, update `prompts/LAST_EXECUTED.md` with the prompt filename, completion date, commit hash if available, and a one-line result. Keep a short history so progress remains auditable.
- After each successfully completed prompt, stage only files belonging to that prompt (including the tracking file), create one focused Git commit, and push the current branch to its configured upstream. Never include unrelated user changes.
- If commit or push is impossible, do not hide the failure: preserve the completed local work and report the exact command and reason.

## Engineering rules

- Keep packages cohesive, avoid circular dependencies and global mutable state, and pass dependencies explicitly.
- Use contexts for I/O and request lifetimes.
- Close response bodies, database rows and resources, and files.
- Use transactions for multi-step metadata mutations.
- Make state reset concurrency-safe.
- Protect every filesystem operation against path traversal and symlink escape.
- Never log credentials, authorization headers, message bodies, or object bodies.
- Return deterministic, provider-shaped errors where compatibility requires them, and generate request IDs locally.
- Test public identifiers and persisted state formats before treating them as stable.
- Never mark a placeholder implementation as supported in the compatibility matrix.

## Testing and verification

Standard verification commands are:

```bash
gofmt -w <changed-go-files>
go test ./...
go vet ./...
make build
make compatibility
make docker-build
make demo
```

Run only commands applicable to the current repository state and task. After any Go change, `go test ./...` is mandatory. Fix failures caused by the task; if the environment prevents a command, report the exact limitation and run the closest safe verification.

- Add unit tests for parsers and state code.
- Test handlers with `httptest`.
- Use real AWS SDK for Go v2 clients for end-to-end compatibility tests, configured to contact loopback endpoints only.
- Tests must not use a user AWS profile, EC2 instance metadata, the environment credential chain, or any fallback to real cloud services.
- Run race tests during hardening and in CI where practical.

## Completion reporting

Every Codex task must finish with:

- a concise implementation summary;
- important design decisions;
- changed files;
- exact verification commands and outcomes; and
- known limitations genuinely outside the task.

## Licensing and contribution constraints

- The core license target is `AGPL-3.0-or-later`.
- Do not add a CLA that assigns copyright or grants special relicensing rights.
- Use DCO sign-off for contributions once contribution documentation exists.
- Do not introduce code with unclear or incompatible licensing.
- Do not copy provider documentation examples wholesale when a clean-room implementation is sufficient.
