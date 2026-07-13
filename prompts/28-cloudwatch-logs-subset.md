# Task 28 — CloudWatch Logs POC subset

Target agent: Codex with GPT-5.6  
Prerequisite: Tasks 00–27.

Implement a useful AWS CloudWatch Logs-compatible subset for local application logging.

## Protocol

Recognize the CloudWatch Logs AWS JSON protocol target namespace used by the current AWS SDK for Go v2, such as:

```text
Logs_20140328.<Operation>
```

Use the exact media type and response/error shapes required by the pinned SDK.

Requirements:

- bounded JSON parsing;
- request IDs;
- provider-shaped errors;
- no SigV4 validation;
- no credential/body logging;
- service registration through the AWS registry.

## Supported operations

1. `CreateLogGroup`
2. `DescribeLogGroups`
3. `DeleteLogGroup`
4. `CreateLogStream`
5. `DescribeLogStreams`
6. `PutLogEvents`
7. `GetLogEvents`
8. `FilterLogEvents` with a clearly bounded subset

## Data model

Persist:

```text
log groups
log streams
log events
```

Each event includes:

```text
timestamp
message
ingestion timestamp
stable event ID or ordering key
```

Requirements:

- group and stream names are validated;
- one stream belongs to one group;
- events are ordered deterministically by event timestamp and ingestion/order tie-breaker;
- persistence survives restart;
- group deletion cascades safely;
- indexes support time-range and stream queries.

## Create/Delete/Describe

### CreateLogGroup

- idempotency/error behavior compatible with the SDK;
- duplicate creation returns the appropriate already-exists error;
- support no tags/KMS/retention settings unless explicitly implemented.

### CreateLogStream

- require group;
- duplicate stream returns already-exists;
- immediate availability.

### DescribeLogGroups / DescribeLogStreams

Support:

- prefix filtering;
- bounded `limit`;
- real `nextToken`;
- deterministic order;
- malformed token rejection.

For streams, support a useful subset of ordering parameters only when semantics are correct; reject unsupported combinations.

### DeleteLogGroup

- remove streams/events;
- missing group returns the documented local/provider error.

## PutLogEvents

- require existing group/stream;
- validate event count, message size, aggregate request size, and timestamp range using documented POC limits;
- require non-decreasing event timestamps or sort only if the provider-compatible behavior is documented;
- accept optional sequence-token fields used by SDK models but do not fabricate concurrency guarantees;
- return the response fields expected by the current SDK;
- write all accepted events atomically;
- report rejected event information only when correctly implemented.

Do not claim exact CloudWatch ingestion delay or throttling behavior.

## GetLogEvents

Support:

- start/end time;
- start from head;
- bounded limit;
- forward/backward pagination tokens;
- deterministic empty-page behavior;
- no duplicate event across pages for an unchanged stream.

## FilterLogEvents

Support:

- group;
- optional stream names/prefix;
- start/end time;
- limit;
- pagination;
- a deliberately small filter-pattern subset.

A safe initial filter subset may include:

```text
empty pattern -> all
quoted literal phrase
space-separated literal terms with AND semantics
leading `-term` exclusion
```

Do not pretend to support full CloudWatch filter syntax. Reject unsupported JSON/metric/filter constructs with a validation error.

## Tests

Cover:

- lifecycle;
- duplicate/missing resources;
- put/get ordering;
- equal timestamps;
- binary-like UTF-8 text and newline content;
- request limits;
- pagination in both directions;
- prefix filtering;
- FilterLogEvents subset;
- malformed token;
- persistence/restart;
- reset/export/import;
- concurrent writers;
- race detector.

## AWS SDK compatibility tests

With a real CloudWatch Logs client:

- CreateLogGroup;
- CreateLogStream;
- PutLogEvents;
- Describe groups/streams;
- GetLogEvents;
- FilterLogEvents;
- DeleteLogGroup;
- typed errors.

Use the request fields emitted by the current pinned SDK rather than hand-crafted-only HTTP tests.

## Compatibility catalog and docs

Mark `FilterLogEvents` partial and list the exact filter syntax. Document no throttling, IAM, subscription filters, metric filters, Insights, or real-time tail.

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
