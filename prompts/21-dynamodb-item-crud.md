# Task 21 — DynamoDB item CRUD

Target agent: Codex with GPT-5.6  
Prerequisite: Tasks 00–20.

Implement basic item operations on DynamoDB tables.

## Supported operations

1. `PutItem`
2. `GetItem`
3. `DeleteItem`
4. `UpdateItem` with a deliberately small, documented initial update-expression subset

All writes must be atomic and use validated AttributeValue/key representations.

## Common validation

For every operation:

- require an existing table;
- validate the supplied primary key against the table schema;
- reject missing, extra, or wrong-type key components;
- enforce item and request size/depth limits;
- never expose SQL or serialized storage internals;
- preserve exact number strings and binary bytes;
- calculate item size deterministically for local metadata;
- reject unsupported parameters rather than silently changing semantics.

`ConsistentRead` may be accepted because local reads are strongly consistent, but document the behavior.

## PutItem

- require all primary key attributes in the item;
- overwrite an existing item with the same complete key;
- persist the full item atomically;
- support `ReturnValues=NONE` and `ALL_OLD`;
- no conditions in this task; reject `ConditionExpression` and legacy conditional fields until Task 22;
- update table item count/size through queryable calculations or safely maintained metadata.

## GetItem

- return no `Item` member when the key is missing, matching SDK expectations;
- support `ProjectionExpression` only if a shared parser is already implemented; otherwise reject it explicitly until later;
- support or safely reject `AttributesToGet`;
- return exact nested values and binary content.

## DeleteItem

- deleting a missing key succeeds with no old attributes;
- support `ReturnValues=NONE` and `ALL_OLD`;
- remove exactly one item;
- no conditions yet.

## UpdateItem initial subset

Implement only enough syntax to be useful without building a fragile regex parser:

```text
SET #name = :value
REMOVE #name
```

Allow comma-separated actions within a single `SET` or `REMOVE` section if the parser is structured.

Requirements:

- use `ExpressionAttributeNames` and `ExpressionAttributeValues`;
- do not allow modification/removal of primary key attributes;
- create a new item only when the request includes a complete valid key and the update semantics permit it;
- support `ReturnValues=NONE`, `ALL_OLD`, `ALL_NEW`, `UPDATED_OLD`, and `UPDATED_NEW` only where the returned shape is correct;
- reject arithmetic, list append, `ADD`, `DELETE`, conditions, and unsupported functions until Task 22;
- implement a small lexer/parser rather than splitting blindly on commas or equals signs.

Task 22 may generalize/refactor the parser, so keep the AST extensible but not overengineered.

## Concurrency

Use transactions so concurrent updates to the same key do not lose unrelated changes silently. Add a regression test with concurrent updates or document/store-level serialization if the initial state engine deliberately serializes all writes.

## SDK compatibility tests

Using a real DynamoDB SDK client:

- PutItem;
- GetItem;
- overwrite with `ALL_OLD`;
- UpdateItem `SET`;
- UpdateItem `REMOVE`;
- DeleteItem with `ALL_OLD`;
- Get missing item;
- typed not-found table error;
- composite key table;
- nested map/list, binary, boolean, null, and large precise number round trips.

## Compatibility catalog

Add stable test IDs and exact partial/supported statuses. If `UpdateItem` is a subset, mark it `partial`.

## Documentation

Document exact `ReturnValues` and expression support.

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
