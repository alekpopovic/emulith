# Task 20 — DynamoDB table lifecycle

Target agent: Codex with GPT-5.6  
Prerequisite: Tasks 00–19.

Implement the first DynamoDB operations using the protocol and state model from Task 19.

## Supported operations

Implement:

1. `CreateTable`
2. `DescribeTable`
3. `ListTables`
4. `DeleteTable`

Use AWS JSON 1.0 and a real AWS SDK for Go v2 compatibility suite.

## POC table model

Support:

- one partition key (`HASH`);
- optional sort key (`RANGE`);
- key attribute types `S`, `N`, or `B`;
- `BillingMode=PAY_PER_REQUEST`;
- table becomes `ACTIVE` immediately;
- deterministic local account ID `000000000000`;
- region from server/request configuration, default `us-east-1`;
- stable table ARN and generated table ID.

Reject clearly:

- provisioned throughput unless the implementation intentionally accepts and documents it as ignored;
- GSI and LSI;
- streams;
- SSE/KMS;
- table classes;
- replicas/global tables;
- tags unless implemented completely;
- duplicate or unused attribute definitions;
- invalid key schemas.

Do not silently ignore unsupported fields that change semantics.

## CreateTable

Validate:

- table name;
- attribute definitions;
- exactly one HASH key;
- at most one RANGE key;
- key attributes exist in definitions;
- no extra attribute definitions in the POC;
- supported scalar key types;
- billing mode.

Persist creation atomically.

Return a compatible `TableDescription` including at least:

```text
TableName
TableStatus=ACTIVE
CreationDateTime
KeySchema
AttributeDefinitions
BillingModeSummary
TableArn
TableId
ItemCount=0
TableSizeBytes=0
```

A duplicate name returns `ResourceInUseException`.

## DescribeTable

Return current metadata and exact key schema.

A missing table returns `ResourceNotFoundException`.

## ListTables

Support:

```text
Limit
ExclusiveStartTableName
```

Requirements:

- lexical deterministic order;
- valid range checking;
- correct `LastEvaluatedTableName`;
- no duplicate/omitted row across pages;
- empty response shape compatible with the SDK.

## DeleteTable

- atomically delete table metadata and all item rows;
- return the deleted table description in the expected transitional shape, documenting that deletion is immediate locally;
- missing table returns `ResourceNotFoundException`;
- no orphaned item rows.

## Concurrency

Test simultaneous create/delete/describe behavior. Duplicate creates must not create two tables. A describe must see either a valid table or a valid not-found result, never partial metadata.

## SDK compatibility tests

With a real DynamoDB client configured only for the loopback Emulith endpoint:

- CreateTable;
- waiter or direct DescribeTable sees `ACTIVE`;
- ListTables pagination;
- duplicate CreateTable typed error;
- DeleteTable;
- DescribeTable typed not-found error.

Do not let SDK waiters introduce long sleeps; use direct polling with bounded delay or configure waiter timings safely.

## Compatibility catalog

Add stable test IDs and statuses. Operations may be marked `supported` only when default SDK tests pass and documented unsupported fields are rejected.

## Documentation

Update the DynamoDB operation/deviation matrix.

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
