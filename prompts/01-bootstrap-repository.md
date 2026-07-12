# Task 01 — Bootstrap the Emulith repository

Target agent: Codex with GPT-5.6  
Prerequisite: Task 00 has created `AGENTS.md`.

Create the initial, buildable Go repository for Emulith.

## Outcome

At the end of this task:

- `go test ./...` passes;
- `make build` produces an `emulith` binary;
- `emulith serve` exposes a health endpoint;
- `emulith version` prints deterministic version information;
- the repository clearly documents that protocol services are not implemented yet.

## Required implementation

### Go module and binary

- Module path: `github.com/emulith/emulith`.
- Binary: `emulith`.
- Use the stable Go toolchain already available in the environment and record the compatible Go version in `go.mod`.
- Entry point: `cmd/emulith/main.go`.
- Use Cobra for the CLI.

### Commands

Implement:

```text
emulith serve
emulith version
```

Reserve `reset` and `apply` for later tasks; do not add fake successful commands.

`serve` flags:

```text
--addr
--data-dir
```

Configuration precedence:

1. explicit CLI flag;
2. environment variable;
3. default.

Environment variables and defaults:

```text
EMULITH_ADDR=:4566
EMULITH_DATA_DIR=./data
```

`version` should print at least:

```text
emulith <version>
commit: <commit>
built: <timestamp>
```

Use link-time variables with useful development defaults such as `dev`, `unknown`, and `unknown`.

### Server

Create clear packages under `internal/config`, `internal/server`, and `internal/cli`.

Expose:

```http
GET /_emulith/health
```

Successful response:

```json
{
  "status": "ok",
  "name": "emulith",
  "version": "dev"
}
```

Requirements:

- status `200`;
- `Content-Type: application/json`;
- no cache;
- bounded server timeouts;
- graceful shutdown on `SIGINT` and `SIGTERM`;
- simple structured logging with `log/slog`;
- admin paths must be isolated so later AWS routing cannot swallow them.

Do not implement S3, SQS, STS, SQLite, reset, or manifest support yet.

### Repository files

Add:

- `README.md`;
- `LICENSE` containing the complete AGPL v3 license text and a clear `AGPL-3.0-or-later` project statement;
- `NOTICE`;
- `.gitignore`;
- `Makefile`.

Make targets:

```text
build
test
run
clean
```

The README must include:

- what Emulith is;
- POC status;
- development/CI-only warning;
- no-account/no-key/no-telemetry principles;
- local build and run commands;
- health check example;
- an explicit “not implemented yet” compatibility table.

### Tests

Add tests for:

- the health handler;
- JSON content type and schema;
- unsupported path behavior;
- `version` output through an injectable writer or command test;
- config precedence for at least environment vs default.

Avoid sleeps and fixed network ports in tests.

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
