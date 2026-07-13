# Task 23 — DynamoDB Query, Scan, projection, and pagination

Target agent: Codex with GPT-5.6  
Prerequisite: Tasks 00–22.

Implement deterministic primary-index `Query` and table `Scan` using the shared expression engine.

## Supported operations

1. `Query`
2. `Scan`

Secondary indexes remain unsupported.

## Query

Require a valid `KeyConditionExpression` for the primary index.

Support:

- partition key equality;
- optional sort-key condition:
  - equality;
  - `<`, `<=`, `>`, `>=`;
  - `BETWEEN`;
  - `begins_with` for string/binary where appropriate;
- `ExpressionAttributeNames`;
- `ExpressionAttributeValues`;
- `ScanIndexForward`;
- `Limit`;
- `ExclusiveStartKey`;
- `FilterExpression`;
- `ProjectionExpression`;
- `Select` values that can be implemented correctly;
- `ConsistentRead`, documented as local strongly consistent behavior.

Reject:

- `IndexName`;
- invalid partition-key operators;
- conditions on non-key attributes inside `KeyConditionExpression`;
- unsupported legacy `KeyConditions` unless implemented fully.

Ordering must use DynamoDB-like type ordering for the declared sort-key type, not lexical JSON order.

## Scan

Support:

- deterministic primary-key order;
- `Limit`;
- `ExclusiveStartKey`;
- `FilterExpression`;
- `ProjectionExpression`;
- compatible `Count` and `ScannedCount`;
- `Select` where correct.

Reject or explicitly document as unsupported:

- parallel scan (`Segment`, `TotalSegments`);
- `IndexName`;
- legacy `ScanFilter`;
- consumed capacity simulation.

## Pagination semantics

Implement real pagination.

Requirements:

- `Limit` applies to evaluated items before filter semantics where DynamoDB requires it;
- `LastEvaluatedKey` identifies the last evaluated item, not merely the last returned item;
- `ExclusiveStartKey` must be a complete valid primary key for the same table;
- no duplicate or omitted item while paginating an unchanged table;
- stable behavior for empty pages caused by filters;
- deterministic ordering;
- bound maximum page/request resource use;
- mutation during pagination is documented as not snapshot-isolated unless a snapshot design is intentionally implemented.

Do not invent opaque continuation tokens; DynamoDB uses key maps.

## ProjectionExpression

Support:

- top-level and nested paths covered by the expression engine;
- aliases for reserved/special names;
- deterministic output;
- primary keys only when selected, matching expected projection semantics;
- validation for overlapping/invalid paths.

## SQL/state behavior

Use safe parameterized queries. Avoid loading the entire table when a bounded query can be performed. For the POC, post-filtering in Go is acceptable when documented, but the implementation must enforce resource limits and preserve correct `Limit`/`LastEvaluatedKey` semantics.

## SDK compatibility tests

Create tables with partition-only and composite keys, then test:

- Query partition equality;
- all supported sort-key predicates;
- forward/reverse order;
- pagination through multiple pages;
- filter producing an empty page with a continuation key;
- projection;
- Scan pagination;
- `Count` vs `ScannedCount`;
- invalid `ExclusiveStartKey`;
- rejected `IndexName`;
- typed validation/not-found errors.

Include numeric sort-key values that expose lexical-order bugs.

## Compatibility catalog and docs

List exact supported Query/Scan parameters and deviations. Mark both `partial` unless the matrix truly covers all documented parameters.

## Verification

```bash
gofmt -w <changed-go-files>
go test ./...
go test -race ./...
go vet ./...
make compatibility
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
