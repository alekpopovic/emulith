# Task 25 — Formal DynamoDB SDK compatibility suite

Target agent: Codex with GPT-5.6  
Prerequisite: Tasks 00–24.

Consolidate and strengthen DynamoDB compatibility coverage into a formal service-level suite.

## Goal

Prove that the current AWS SDK for Go v2 can use Emulith's documented DynamoDB subset end to end, across restarts and concurrent requests, without any real-cloud fallback.

This task primarily improves tests, reports, and defects found by those tests. Do not add unrelated DynamoDB operations.

## Harness

Reuse the loopback-only compatibility harness.

Ensure:

- explicit endpoint;
- static fake credentials;
- no shared config/profile;
- EC2 metadata disabled;
- a custom transport rejects non-loopback hosts;
- unique temporary data directory;
- restart support using the same directory;
- injectable clock where visibility/time assertions are relevant;
- stable compatibility test IDs.

## Required scenarios

### Table lifecycle

- Create partition-only table.
- Create composite-key table.
- Describe exact schema.
- List tables through multiple pages.
- Duplicate create typed error.
- Delete table.
- Missing table typed error.

### Attribute fidelity

Round-trip:

- string;
- high-precision positive/negative number;
- binary;
- boolean;
- null;
- map;
- list;
- string/number/binary sets.

Verify no precision or binary corruption.

### CRUD

- Put/Get.
- Overwrite with old values.
- Delete with old values.
- Update SET/REMOVE.
- Numeric increment.
- Set add/delete.
- Nested path update.
- Correct return-value modes.
- Missing item behavior.

### Conditions

- create-if-absent;
- compare-and-update;
- conditional delete;
- one winner under concurrent conditional writes;
- typed `ConditionalCheckFailedException`.

### Query and Scan

- partition query;
- string, numeric, and binary sort-key ordering;
- every supported sort predicate;
- reverse scan direction;
- filter and projection;
- multi-page pagination;
- empty filtered page with `LastEvaluatedKey`;
- Scan counts.

### Batch

- multi-table BatchWrite;
- multi-table BatchGet;
- missing items;
- delete request;
- validation failure leaves state consistent.

### Persistence and reset

- create data;
- stop server cleanly;
- restart with same data directory;
- verify tables/items remain;
- export/import round trip when Task 16 is present;
- reset removes all DynamoDB state and leaves other service behavior healthy.

### Concurrency

- concurrent distinct-key writes;
- concurrent same-key updates;
- pagination while mutation occurs, checked against documented non-snapshot behavior;
- no races under `go test -race`.

## Negative and limit tests

Use SDK calls where possible for:

- malformed key;
- wrong key type;
- unsupported index;
- unsupported billing/index fields;
- oversized item/expression;
- expression depth/token limit;
- invalid update path;
- invalid batch size;
- unsupported operation target.

## Compatibility report

Ensure every implemented DynamoDB operation has:

- stable test IDs;
- catalog entry;
- generated JSON result;
- generated Markdown status and deviations;
- pinned AWS SDK module version in the report.

No operation becomes `supported` solely because a handler unit test passes.

## Defect policy

If the formal suite finds a product defect:

- fix the handler/state/parser;
- add a focused regression test;
- keep the compatibility scenario;
- do not weaken assertions or mark a real failure skipped.

## Performance guard

Add modest local thresholds or benchmarks only to catch accidental quadratic behavior. Do not create brittle wall-clock CI gates. At minimum, verify a few thousand small items can be inserted/scanned within bounded memory and without test timeouts.

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
make demo
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
