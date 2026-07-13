# Task 22 — DynamoDB expression parser and conditional writes

Target agent: Codex with GPT-5.6  
Prerequisite: Tasks 00–21.

Build a real, bounded DynamoDB expression lexer/parser/evaluator and use it for conditional item operations and richer updates.

## Architecture

Create clear layers:

```text
lexer -> parser -> validated AST -> evaluator/mutator
```

Do not implement expressions as a chain of regular expressions or unsafe string substitutions.

Requirements:

- token positions for useful errors;
- recursion and token-count limits;
- expression length limit;
- deterministic evaluation;
- no panics on malformed input;
- path support for map fields and list indices where implemented;
- placeholder resolution through `ExpressionAttributeNames` and `ExpressionAttributeValues`;
- reject unused/missing placeholders deliberately and consistently;
- primary key attributes remain immutable.

## ConditionExpression subset

Support:

```text
attribute_exists(path)
attribute_not_exists(path)
attribute_type(path, type)
begins_with(path, operand)
contains(path, operand)
size(path)

=  <>  <  <=  >  >=
BETWEEN
IN
AND
OR
NOT
parentheses
```

Implement DynamoDB-compatible type comparison rules for the supported types. Do not coerce strings and numbers.

If a function/operator is not implemented completely, reject it with a provider-shaped validation error and document it.

Apply conditions atomically to:

- `PutItem`;
- `DeleteItem`;
- `UpdateItem`.

A false condition returns:

```text
ConditionalCheckFailedException
```

No mutation occurs.

Support `ReturnValuesOnConditionCheckFailure` only if it can be made SDK-compatible; otherwise reject it explicitly.

## UpdateExpression

Extend the AST to support:

```text
SET
REMOVE
ADD
DELETE
```

Include useful `SET` forms:

```text
path = value
path = path
path = path + value
path = path - value
if_not_exists(path, value)
list_append(list, list)
```

Requirements:

- each section appears at most once;
- paths do not overlap in an ambiguous way;
- update actions are applied with DynamoDB-like semantics;
- `ADD` supports numbers and sets only;
- `DELETE` supports sets only;
- empty sets are rejected;
- out-of-range list indexes behave consistently and are documented;
- all validation completes before persisting the mutation;
- return-value modes remain correct.

## Shared use

Design the expression engine so Task 23 can reuse it for:

- `KeyConditionExpression`;
- `FilterExpression`;
- `ProjectionExpression`.

Do not couple the AST to HTTP handlers or SQL.

## Errors

Return DynamoDB-shaped:

```text
ValidationException
ConditionalCheckFailedException
```

Error messages should be useful but need not copy AWS wording verbatim. Do not leak internals.

## Tests

Add table-driven parser/evaluator tests for:

- operator precedence;
- parentheses;
- reserved-word aliases;
- missing/unused placeholders;
- nested paths;
- list indexes;
- every supported function/operator;
- type mismatch;
- short-circuit behavior;
- false condition causes no write;
- concurrent conditional write where only one succeeds;
- arithmetic precision;
- set add/delete;
- malformed and adversarial expressions;
- depth/token/size limits;
- fuzz tests for lexer/parser with bounded input.

## SDK compatibility tests

Use real SDK calls for:

- conditional PutItem success/failure;
- conditional DeleteItem;
- conditional UpdateItem;
- `attribute_not_exists`;
- numeric increment;
- `if_not_exists`;
- list append;
- set add/delete;
- typed `ConditionalCheckFailedException`.

## Compatibility catalog and docs

Mark `UpdateItem` and conditional support according to exact test coverage, with deviations listed.

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
