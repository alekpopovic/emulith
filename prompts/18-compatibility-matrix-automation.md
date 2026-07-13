# Task 18 — Automate the compatibility matrix

Target agent: Codex with GPT-5.6  
Prerequisite: Tasks 00–17.

Make Emulith's public compatibility claims machine-readable and enforce them in CI.

## Goal

A service operation cannot be documented as supported unless an executable compatibility test proves that a real AWS SDK client can use it against Emulith.

## Source of truth

Create a machine-readable catalog, for example:

```text
compatibility/aws.yaml
```

Each entry must include:

```text
provider
service
operation
status
protocol
test_id
notes
known_deviations
since
```

Allowed statuses:

```text
supported
partial
experimental
unsupported
```

Define them precisely:

- `supported`: intended POC behavior is covered by a passing SDK compatibility test;
- `partial`: the operation works but documented parameters/semantics are incomplete;
- `experimental`: available but public behavior may change;
- `unsupported`: recognized or planned but not implemented.

A missing test ID cannot be `supported`.

## Test linkage

Give compatibility tests stable IDs, for example:

```text
aws.sts.GetCallerIdentity.basic
aws.s3.PutObject.basic
aws.sqs.ReceiveMessage.visibility
```

Implement a registration/reporting mechanism that:

- records executed IDs;
- rejects duplicate IDs;
- reports pass/fail/skip;
- emits deterministic JSON;
- does not turn ordinary unit tests into compatibility claims;
- detects catalog entries whose referenced test does not exist;
- detects executed tests missing from the catalog when they claim public compatibility.

Do not parse human-readable `go test` output with fragile regular expressions if a direct Go reporter can produce structured output.

## Generated artifacts

Generate:

```text
build/compatibility/aws.json
docs/compatibility/aws.md
```

The JSON report should include:

```text
Emulith version
commit
Go version
SDK module versions
generation time
catalog entries
test results
summary counts
```

The Markdown document must be generated from the catalog/report, not manually duplicated.

Use deterministic ordering. Keep generated timestamps out of committed Markdown if they create pointless diffs.

## Commands

Add:

```bash
make compatibility
make compatibility-report
make compatibility-check
```

Suggested behavior:

- `compatibility`: run the SDK compatibility suite;
- `compatibility-report`: create JSON and Markdown outputs;
- `compatibility-check`: regenerate into a temp location and fail when committed docs/catalog linkage are stale or invalid.

## CI

Require `compatibility-check` on pull requests.

The check must fail when:

- a supported entry lacks a passing test;
- a test ID is duplicated;
- a catalog operation is duplicated;
- generated docs are stale;
- an implemented operation is newly claimed without notes/status;
- a “supported” operation test is skipped by default.

## Initial catalog

Include all existing operations for:

- STS;
- S3;
- SQS.

Represent unsupported major operations explicitly only where useful; do not create hundreds of speculative rows.

Ensure every status reflects actual code and tests.

## Tests

Test the compatibility tooling itself:

- valid catalog;
- invalid status;
- duplicate entry;
- missing test ID;
- missing test;
- duplicate test ID;
- deterministic generation;
- Markdown escaping;
- JSON schema stability;
- non-zero command exit on mismatch.

## Documentation

Update contributor docs with the workflow for adding an operation:

1. implement handler;
2. add unit tests;
3. add real SDK compatibility test with stable ID;
4. update catalog;
5. regenerate/check report.

## Verification

```bash
gofmt -w <changed-go-files>
go test ./...
go test -race ./...
go vet ./...
make compatibility
make compatibility-report
make compatibility-check
make build
```


## Execution contract

You are the implementation agent for this task. Complete the work in the current repository; do not stop after writing a plan.

1. Read every applicable `AGENTS.md` before changing files.
2. Inspect the repository, current architecture, tests, dependency versions, and documentation.
3. Run the relevant baseline tests before making changes when practical.
4. State a concise implementation plan, then execute it immediately.
5. Make reasonable non-blocking assumptions instead of asking for confirmation.
6. Keep the change scoped to this task. Do not implement later roadmap items.
7. Preserve existing working behavior and compatibility unless this task explicitly changes it.
8. Prefer simple, maintainable Go and explicit protocol behavior over speculative abstraction.
9. Never use LocalStack, Moto, MinIO, ElasticMQ, Azurite, or another cloud emulator as an Emulith runtime dependency.
10. Never contact real AWS, Azure, or GCP endpoints. Tests must be hermetic and loopback-only.
11. Do not add accounts, entitlement checks, license keys, forced telemetry, analytics, or phone-home behavior.
12. Do not commit, push, tag, publish a release, or open a pull request.
13. Format changed files and run all verification commands applicable to the repository.
14. Fix failures caused by the change. If the environment blocks a command, report the exact limitation and run the closest safe verification.
15. Update compatibility documentation only for behavior backed by executable tests.
16. Finish with:
    - implementation summary;
    - important design decisions;
    - changed files;
    - exact commands run and outcomes;
    - genuine remaining limitations.

Unless a task explicitly changes the release scope, Emulith remains a development/CI emulator, not a production service.
