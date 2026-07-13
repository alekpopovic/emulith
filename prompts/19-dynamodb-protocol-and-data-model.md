# Task 19 — DynamoDB protocol foundation and data model

Target agent: Codex with GPT-5.6  
Prerequisite: Tasks 00–18.

Add the protocol and persistence foundation for a DynamoDB-compatible subset. Do not implement table or item operations yet beyond explicit unsupported responses.

## Protocol

Recognize AWS JSON 1.0 requests:

```http
POST /
X-Amz-Target: DynamoDB_20120810.<Operation>
Content-Type: application/x-amz-json-1.0
```

Requirements:

- case-insensitive header parsing;
- media-type parameter support;
- bounded JSON body;
- duplicate/unknown critical field behavior is deliberate and tested;
- request ID headers;
- DynamoDB-shaped JSON error responses;
- no SigV4 validation for the local POC;
- signed requests remain readable;
- authorization/security-token headers are never logged.

Register DynamoDB through the service registry created earlier.

## AttributeValue model

Implement a provider-specific internal representation for:

```text
S
N
B
BOOL
NULL
M
L
SS
NS
BS
```

Requirements:

- exactly one AttributeValue variant per value;
- recursively validate maps/lists;
- preserve number strings exactly; do not convert to `float64`;
- validate DynamoDB-style decimal syntax and canonical comparison without losing precision;
- decode/encode binary values safely using base64 at the protocol boundary;
- reject invalid set elements and duplicate set members according to the documented POC behavior;
- impose recursion depth, item-size, map/list-count, and aggregate allocation limits;
- deterministic canonical encoding for equality, keys, hashing, and persistence;
- distinguish absent values from `NULL`.

Do not create a generic cloud value type.

## Key schema model

Define types for:

```text
HASH partition key
RANGE sort key
S, N, or B key scalar
```

Implement safe canonical key encoding that:

- cannot collide across type/value boundaries;
- preserves numeric comparison semantics;
- supports composite keys;
- is deterministic across processes and architectures;
- has round-trip tests.

## Persistence schema

Add migrations for DynamoDB metadata and items, with tables conceptually equivalent to:

```text
dynamodb_tables
dynamodb_attributes
dynamodb_items
```

Store enough information for:

- table name;
- table ID/ARN;
- status;
- creation time;
- billing mode;
- partition/sort key definitions;
- canonical primary key;
- complete item payload;
- item size.

Use a format that supports future migrations. Do not serialize Go-specific gob data as the durable format.

Add indexes for table/key lookup and future Query/Scan behavior.

## Public package boundaries

Keep:

- JSON wire structs;
- validated domain values;
- persistence records;
- expression ASTs planned for later;

as separate concepts where useful. Avoid exposing raw SQL or unvalidated wire maps to handlers.

## Unsupported operation behavior

For DynamoDB operations not yet implemented:

- route them to the DynamoDB service;
- return a provider-shaped unsupported/unknown-operation error;
- do not mark them supported.

## Tests

Cover:

- AWS target parsing;
- correct JSON content type;
- error envelope and request ID;
- every AttributeValue variant;
- invalid multiple variants;
- number precision and ordering;
- binary round trip;
- deep/nested input limits;
- duplicate/invalid sets;
- deterministic canonical encoding;
- composite key collision resistance;
- migrations and reopen behavior;
- fuzz tests for AttributeValue JSON and key decoding with bounded resources.

## Compatibility catalog

Add DynamoDB service entries as `experimental` or `unsupported`; no operation may be `supported` in this task.

## Verification

```bash
gofmt -w <changed-go-files>
go test ./...
go test -race ./...
go vet ./...
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
