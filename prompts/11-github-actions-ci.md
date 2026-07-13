# Task 11 — Add GitHub Actions CI

Target agent: Codex with GPT-5.6  
Prerequisite: Tasks 00–10.

Create a secure, dependency-free CI workflow for pull requests and the default branch.

## Workflow

Add:

```text
.github/workflows/ci.yml
```

Triggers:

- pull requests;
- pushes to `main`;
- manual dispatch.

Use least-privilege permissions:

```yaml
permissions:
  contents: read
```

Do not request cloud credentials, repository write permissions, package publication, OIDC, or secrets.

## Required checks

Organize jobs clearly without unnecessary duplication.

### Go quality

Run:

```bash
gofmt check
go vet ./...
go test -race ./...
make build
make compatibility
```

A gofmt check must fail with a readable diff/list and must not modify files in CI.

Use the Go version from `go.mod` or a single repository source of truth. Cache modules/build data using the supported setup action behavior.

### Docker build

Run:

```bash
docker build .
```

Do not push the image.

Where practical, run the container and validate:

```http
GET /_emulith/health
```

Use a bounded readiness loop and always clean up.

## Supply-chain basics

- Pin actions to stable major versions at minimum; prefer immutable commit SHAs if the repository policy already requires them and maintainability is acceptable.
- Do not execute untrusted pull-request code with elevated tokens.
- Do not use `pull_request_target`.
- Do not download arbitrary scripts with `curl | sh`.
- Do not make real AWS calls.
- Set `AWS_EC2_METADATA_DISABLED=true` for tests as defense in depth.

## Local parity

Add or adjust Make targets only when they improve local reproduction. README should show the local commands corresponding to CI.

## Verification

Validate YAML syntax using available tools, inspect the rendered workflow logically, and run all local equivalents:

```bash
gofmt -w <changed-go-files>
go test -race ./...
go vet ./...
make build
make compatibility
make docker-build
```

If Docker is unavailable, report that exact limitation; do not remove the Docker job.

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
