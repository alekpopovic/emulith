# Task 39 — Azure Table OData query and pagination

Target agent: Codex with GPT-5.6  
Prerequisite: Tasks 00–38.

Implement a bounded OData filter parser/evaluator, projection, and continuation paging for Azure Table queries.

## Query surface

Support the query options emitted by the current official SDK:

```text
$filter
$select
$top
```

Return continuation headers:

```text
x-ms-continuation-NextPartitionKey
x-ms-continuation-NextRowKey
```

Also apply query support to listing tables where appropriate.

## Parser architecture

Build:

```text
lexer -> parser -> typed AST -> validator -> evaluator/planner
```

Do not use a collection of regular expressions.

Requirements:

- token positions;
- bounded expression length;
- token/depth limits;
- no panics;
- OData string escaping with doubled quotes;
- typed literals;
- deterministic comparison;
- useful Azure-shaped syntax/validation errors.

## Filter subset

Support:

```text
eq
ne
lt
le
gt
ge
and
or
not
parentheses
```

Operands:

- property references;
- string literals;
- Int32/Int64;
- double;
- boolean;
- datetime;
- guid;
- binary when the SDK emits a supported literal form.

Requirements:

- no cross-type coercion that loses meaning;
- missing property behavior is documented and tested;
- comparisons on PartitionKey/RowKey use correct lexical semantics;
- numeric comparisons are numeric, not string based;
- Timestamp is queryable;
- unsupported functions such as `startswith` are rejected unless fully implemented.

## Query execution

Entity query supports:

- entire table scan with bounds;
- efficient PartitionKey equality/range where indexes permit;
- RowKey conditions;
- combined conditions;
- post-filtering in Go only when resource limits preserve correctness;
- stable order by PartitionKey then RowKey;
- `$top` validation;
- `$select` projection;
- empty result shape;
- exact typed JSON/OData output.

Do not add arbitrary order-by; Azure Table query order is key based.

## Pagination

Requirements:

- continuation keys identify the last evaluated key;
- no duplicate/omitted entity across pages for unchanged data;
- `$top` and server page size interact correctly;
- empty filtered page may still return continuation;
- malformed/foreign continuation is rejected;
- continuation values are correctly encoded in headers;
- no snapshot isolation claim during mutation;
- bounded maximum page size.

Where the SDK supplies continuation values back as request parameters, parse them safely and require a complete valid pair.

## Projection

`$select` must:

- support aliases only if the API uses them;
- return selected user properties;
- preserve required system/key fields according to service behavior;
- reject invalid/duplicate property paths;
- not mutate stored entities.

## Tests

Cover:

- every operator;
- precedence/parentheses/not;
- every supported literal type;
- Int64 precision;
- escaped string;
- missing property;
- PartitionKey/RowKey ranges;
- multi-partition query;
- `$select`;
- `$top`;
- pagination;
- empty filtered page with continuation;
- malformed filter;
- depth/token limits;
- malformed continuation;
- mutation-between-pages documented behavior;
- fuzz parser seeds.

## Official Azure SDK compatibility

Use the official Table SDK pager/query API for:

- partition equality;
- RowKey range;
- numeric/date property filter;
- combined boolean filter;
- projection;
- multiple pages;
- continuation resume;
- invalid filter decoded error.

## Compatibility catalog and docs

Mark query partial and list the exact grammar. Do not claim full OData support.

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

You are the implementation agent for this task. Complete the work in the current Emulith repository; do not stop after writing a plan.

1. Read every applicable `AGENTS.md` before changing files.
2. Inspect the repository, current architecture, migrations, tests, dependency versions, compatibility catalog, and documentation.
3. Run the relevant baseline checks before editing when practical.
4. State a concise implementation plan, then execute it immediately.
5. Make reasonable non-blocking assumptions instead of asking for confirmation.
6. Keep the change scoped to this task. Do not implement later roadmap items.
7. Preserve all existing AWS behavior and compatibility unless this task explicitly fixes a defect.
8. Prefer explicit provider-specific protocol code over a false universal cloud abstraction.
9. Never use Azurite, LocalStack, Moto, MinIO, ElasticMQ, or another emulator as an Emulith runtime dependency.
10. Never contact real Azure, AWS, or GCP endpoints. All tests must be hermetic and loopback-only.
11. Do not use `DefaultAzureCredential`, managed identity probing, user Azure CLI credentials, or ambient cloud profiles in compatibility tests.
12. Do not add accounts, entitlement checks, license keys, forced telemetry, analytics, or phone-home behavior.
13. Do not commit, push, tag, publish a release, or open a pull request.
14. Bound all parsers, request bodies, archive inputs, page sizes, and allocations derived from untrusted input.
15. Never log account keys, authorization headers, SAS tokens, request bodies containing user data, queue messages, entities, or blob bodies.
16. Format changed files and run every verification command applicable to the repository.
17. Fix failures caused by your change. If the environment blocks a command, report the exact limitation and run the closest safe verification.
18. Update compatibility documentation only for behavior backed by executable SDK compatibility tests.
19. Finish with:
    - implementation summary;
    - important design decisions;
    - changed files;
    - exact commands run and outcomes;
    - compatibility status changes;
    - genuine remaining limitations.

Emulith remains a development/CI emulator, not a production service.
