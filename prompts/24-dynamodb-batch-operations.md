# Task 24 — DynamoDB batch operations

Target agent: Codex with GPT-5.6  
Prerequisite: Tasks 00–23.

Implement DynamoDB `BatchGetItem` and `BatchWriteItem` for local development.

## Supported operations

1. `BatchGetItem`
2. `BatchWriteItem`

Do not implement transactional APIs or PartiQL in this task.

## BatchGetItem

Support:

- multiple tables in one request;
- up to the documented POC limit, with a preferred AWS-compatible maximum of 100 requested keys;
- `ConsistentRead`;
- `ProjectionExpression`;
- `ExpressionAttributeNames`;
- deterministic response grouping;
- duplicate-key validation;
- exact key-schema validation;
- missing items omitted from results;
- `UnprocessedKeys`.

Because Emulith does not simulate throttling, successful valid local requests should normally return empty `UnprocessedKeys`. Preserve the response shape and design internals so future fault injection can produce unprocessed keys without breaking the API.

Enforce aggregate request/response size limits. Do not allocate based on untrusted counts without bounds.

## BatchWriteItem

Support:

- multiple tables;
- up to 25 write requests per call;
- `PutRequest`;
- `DeleteRequest`;
- one action per entry;
- duplicate operation on the same table/key rejected consistently;
- complete validation before mutation;
- `UnprocessedItems`.

Use one local transaction when the state backend can safely cover all included tables. If atomic all-or-nothing behavior differs from AWS's partial processing model, document this deliberate local-development behavior. Never return a partial success without accurately representing unprocessed items.

`PutRequest` must reuse normal PutItem validation and storage semantics. `DeleteRequest` must reuse DeleteItem semantics.

## Unsupported parameters

Safely support or reject:

- `ReturnConsumedCapacity`;
- `ReturnItemCollectionMetrics`.

Do not fabricate capacity values unless the compatibility policy explicitly defines zero/local values.

## Errors

Return provider-shaped errors for:

- missing table;
- invalid key;
- duplicate key/action;
- too many items;
- oversized request;
- malformed AttributeValue;
- mixed/invalid request entry;
- internal transaction failure.

Validation must occur before any mutation when the implementation promises atomic local behavior.

## Tests

Cover:

- multi-table BatchGet;
- missing items;
- projection;
- duplicate keys;
- maximum count boundary;
- oversized request;
- multi-table BatchWrite;
- put and delete;
- duplicate action on same key;
- rollback on one invalid entry;
- concurrent batches;
- exact number/binary/nested value round trip;
- deterministic output suitable for tests without claiming AWS response order where it is not guaranteed.

## SDK compatibility tests

Use the real DynamoDB SDK client for:

- BatchWrite across two tables;
- BatchGet from both tables;
- empty unprocessed maps;
- missing item behavior;
- delete requests;
- typed validation and missing-table errors.

## Compatibility catalog and docs

Add stable test IDs and state the local no-throttling/transaction behavior explicitly.

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
