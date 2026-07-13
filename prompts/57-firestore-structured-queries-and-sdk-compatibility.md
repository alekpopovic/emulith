# Task 57 — Firestore structured queries and formal SDK compatibility

Target agent: Codex with GPT-5.6  
Prerequisite: Tasks 00–56.

Implement `RunQuery` for a documented Firestore StructuredQuery subset and build the formal official-client compatibility suite.

## Supported RPC

Implement:

```text
RunQuery
```

AggregateQuery, Listen/watch, PartitionQuery, indexes admin, and collection import/export remain unsupported.

## Query scope

Support:

```text
parent
from collection
all_descendants=false initially
select
where
order_by
start_at
end_at
offset
limit
```

Collection-group queries (`all_descendants=true`) may be implemented only if fully tested; otherwise reject explicitly.

## Filter subset

Support field filters:

```text
EQUAL
NOT_EQUAL
LESS_THAN
LESS_THAN_OR_EQUAL
GREATER_THAN
GREATER_THAN_OR_EQUAL
ARRAY_CONTAINS
IN
NOT_IN
ARRAY_CONTAINS_ANY
```

Support unary filters:

```text
IS_NULL
IS_NOT_NULL
IS_NAN
IS_NOT_NAN
```

Support composite filters:

```text
AND
OR with documented limits
```

Requirements:

- validate operand cardinality for `IN`/`NOT_IN`/`ARRAY_CONTAINS_ANY`;
- no invalid null/NaN combinations;
- field existence semantics;
- exact Firestore value comparison/order;
- document name (`__name__`) support;
- deterministic evaluation;
- bounded filter tree depth/count.

## Ordering

Implement Firestore-like value ordering for supported types.

Requirements:

- explicit `order_by`;
- implicit inequality field ordering where required;
- implicit document-name tie-breaker;
- ascending/descending;
- documents missing an ordered field behave according to documented subset;
- NaN/null ordering tested;
- no locale-dependent string sorting;
- reference/geo/array/map ordering only if required and correctly modeled; otherwise reject such order clauses explicitly.

## Cursors

Support:

```text
startAt
startAfter
endAt
endBefore
```

through the StructuredQuery cursor fields.

Requirements:

- value count matches normalized order clauses;
- before/after semantics correct;
- document snapshot cursors from official client work;
- deterministic pagination-like behavior;
- malformed cursor returns `InvalidArgument`.

## Projection, offset, and limit

- selected fields plus required document metadata;
- nested field masks;
- bounded offset;
- limit validation;
- stream responses progressively rather than buffering an unbounded result;
- include transaction/read time consistently.

## Query planner/execution

Use safe indexed lookup where practical:

- collection parent index;
- document name;
- equality/range filtering.

Post-filtering in Go is acceptable for the POC when:

- resource scan limits are enforced;
- semantics remain correct;
- no false composite-index error is fabricated.

Document that production composite-index enforcement is not simulated.

## RunQuery streaming

Return one or more `RunQueryResponse` messages with:

```text
document
read_time
skipped_results when relevant
transaction when started through new_transaction
```

Handle empty result appropriately.

## Formal official-client suite

Use the official Firestore Go client and loopback-only emulator endpoint.

Required scenarios:

### CRUD/transactions regression

Re-run all Task 56 high-level flows.

### Queries

- collection documents;
- equality;
- range;
- chained AND;
- OR where supported;
- array contains;
- IN/not-in;
- null/NaN;
- order asc/desc;
- limit;
- offset;
- start/end cursors;
- document snapshot cursor;
- projection/select through supported API;
- multiple pages/streamed results conceptually;
- invalid query errors.

### Fidelity

- Int64 precision;
- timestamp;
- bytes;
- reference;
- geo point;
- arrays/maps;
- restart;
- reset;
- export/import.

### Concurrency

- transaction retry under concurrent writer;
- query during mutation, checked against documented non-snapshot semantics;
- race detector.

## Compatibility report

Every public method/flow needs stable IDs, official client module version, status, and deviations.

Mark `RunQuery` partial unless the documented StructuredQuery surface is fully covered.

## Performance sanity

Test a few thousand documents and several filtered/ordered queries for bounded memory and no obvious quadratic parser/evaluator behavior. Avoid brittle time gates.

## Documentation

Document:

- exact filter/operator support;
- cursor/order rules;
- no Listen/watch;
- no composite-index enforcement;
- no aggregation;
- default database only.

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

You are the implementation agent for this task. Complete the work in the current Emulith repository; do not stop after writing a plan.

1. Read every applicable `AGENTS.md` before changing files.
2. Inspect the repository, provider registry, listener lifecycle, migrations, tests, compatibility catalog, generated documentation, Docker setup, and dependency versions.
3. Run the relevant baseline checks before editing when practical.
4. State a concise implementation plan, then execute it immediately.
5. Make reasonable non-blocking assumptions instead of asking for confirmation.
6. Keep the change scoped to this task. Do not implement later roadmap items.
7. Preserve all existing AWS and Azure behavior and compatibility unless this task explicitly fixes a defect.
8. Prefer provider-specific protocol implementations over a false universal cloud abstraction.
9. Never use Google Cloud emulators, LocalStack, Azurite, Moto, MinIO, ElasticMQ, or another emulator as an Emulith runtime dependency.
10. Never contact real GCP, AWS, or Azure endpoints. All tests must be hermetic and loopback-only.
11. Do not use Application Default Credentials, `gcloud` user credentials, service-account files, metadata-server probing, workload identity, or ambient cloud profiles in compatibility tests.
12. Do not add accounts, entitlement checks, license keys, forced telemetry, analytics, or phone-home behavior.
13. Do not commit, push, tag, publish a release, or open a pull request.
14. Bound all parsers, protobuf/JSON bodies, gRPC messages, stream buffers, page sizes, upload sessions, query plans, and allocations derived from untrusted input.
15. Never log authorization headers, OAuth tokens, API keys, signed URLs, object bodies, Pub/Sub message data, Firestore document data, or other user payloads.
16. Keep request IDs, error mapping, and compatibility claims deterministic and test-backed.
17. Format changed files and run every verification command applicable to the repository.
18. Fix failures caused by your changes. If the environment blocks a command, report the exact limitation and run the closest safe verification.
19. Update compatibility documentation only for behavior backed by executable official-client compatibility tests.
20. Finish with:
    - implementation summary;
    - important design decisions;
    - changed files;
    - exact commands run and outcomes;
    - compatibility status changes;
    - genuine remaining limitations.

Emulith remains a development/CI emulator, not a production service.
