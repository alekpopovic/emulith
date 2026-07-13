# Task 36 — Azure Queue Storage lifecycle

Target agent: Codex with GPT-5.6  
Prerequisite: Tasks 00–35.

Implement Azure Queue Storage queue lifecycle and metadata.

## Supported operations

1. Create Queue
2. Delete Queue
3. List Queues
4. Get Queue Properties/Metadata
5. Set Queue Metadata
6. Get Queue Service Properties with a minimal truthful subset

Message operations are implemented in Task 37.

## Persistence

Add versioned queue metadata:

```text
account
queue name
metadata map
created_at
last modified
ETag if used by the protocol
```

Future message rows must reference account and queue with cascading cleanup.

Do not reuse AWS SQS tables directly; semantics differ.

## Naming

Implement Azure Queue name validation:

- documented length;
- lowercase/alphanumeric/hyphen rules;
- no invalid consecutive/trailing separators;
- exact percent-decoding behavior;
- no internal route collision.

Return:

```text
InvalidResourceName
QueueAlreadyExists
QueueNotFound
```

## Create Queue

- parse `x-ms-meta-*`;
- create atomically;
- repeated creation behavior must match SDK expectations;
- reject unsupported encryption/immutability-like headers;
- return correct success status/headers.

## Delete Queue

- delete queue and future messages atomically;
- missing queue behavior must be provider-compatible;
- immediate local deletion; document lack of Azure's potential recreation delay.

## List Queues

Support:

```text
prefix
marker
maxresults
include=metadata
```

Requirements:

- deterministic lexical order;
- real validated pagination;
- bounded maximum page;
- correct XML response;
- correct `NextMarker`;
- include approximate message count only where the service API expects it.

## Get / Set Metadata

Get returns:

```text
x-ms-meta-*
x-ms-approximate-messages-count
```

Set replaces metadata atomically and enforces limits.

Approximate count may be calculated exactly in the local emulator, but document that it is returned through an approximate field.

## Service properties

The official SDK may probe service properties.

Implement a minimal response for modeled properties only, or reject unsupported operations with a clear Azure error. Do not fabricate CORS/analytics/retention settings.

## Conditional headers

Queue lifecycle conditions are not required unless the official SDK emits them. Reject unsupported conditional headers explicitly.

## Tests

Cover:

- name validation;
- create/duplicate;
- metadata;
- list prefix/pagination/include;
- delete/missing;
- count when empty;
- XML escaping;
- malformed marker;
- concurrency;
- restart;
- reset/export/import;
- AWS SQS remains unaffected by an Azure queue with the same name.

## Official Azure SDK compatibility

Using the current official Queue SDK and explicit loopback URL/shared key:

- create queue;
- get/set properties/metadata;
- list queues with pager;
- delete queue;
- decoded duplicate/missing errors.

## Compatibility catalog and docs

Add stable Azure Queue lifecycle IDs. Message operations remain unsupported until Task 37.

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
