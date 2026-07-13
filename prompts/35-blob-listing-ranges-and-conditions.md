# Task 35 — Blob listing, ranges, and conditional requests

Target agent: Codex with GPT-5.6  
Prerequisite: Tasks 00–34.

Complete the useful Blob POC by adding listing, byte ranges, and HTTP precondition behavior.

## List Blobs

Support the official SDK's list operation with:

```text
prefix
delimiter
marker
maxresults
include=metadata
include=uncommittedblobs only if accurate
```

Requirements:

- deterministic lexical ordering by logical blob name;
- marker-based pagination with no duplicate/omitted item for an unchanged container;
- validated continuation marker;
- correct XML namespace/escaping;
- blob properties included only when modeled;
- metadata included only when requested;
- delimiter produces virtual directory/prefix entries;
- bounded page size;
- deletion between pages follows documented non-snapshot behavior;
- unsupported include values are rejected.

Do not expose internal staged block files as blobs.

## Byte ranges

Support:

```text
Range: bytes=start-end
x-ms-range: bytes=start-end
```

Requirements:

- a request must not supply conflicting ranges;
- support closed, open-ended, and suffix ranges where the official SDK emits them;
- return `206 Partial Content`;
- return exact `Content-Range`;
- return correct `Content-Length`;
- preserve relevant blob headers;
- `416` with provider-compatible error for unsatisfiable/invalid ranges;
- zero-byte blob edge cases;
- HEAD with range behaves consistently;
- stream only the selected section;
- no full-blob buffering.

Add `Accept-Ranges: bytes`.

## Conditional headers

Implement for properties, get, put/overwrite, metadata/header update, and delete where appropriate:

```text
If-Match
If-None-Match
If-Modified-Since
If-Unmodified-Since
```

Requirements:

- normalize and compare ETags correctly;
- wildcard `*` behavior;
- second-level time precision according to HTTP/Azure expectations;
- evaluate all supplied conditions in the correct logical combination;
- failed precondition causes no mutation;
- return `304 Not Modified` or `412 ConditionNotMet` as appropriate;
- do not apply conditions to unsupported snapshot/version selectors;
- test stale/fresh ETags under concurrent overwrite.

Create a shared condition evaluator reusable across Blob operations without coupling it to S3 semantics.

## Errors

Implement/test:

```text
BlobNotFound
ContainerNotFound
InvalidRange
ConditionNotMet
InvalidMarker
InvalidQueryParameterValue
UnsupportedHeader
```

## Tests

Cover:

- list empty/nonempty;
- prefix;
- delimiter virtual directories;
- pagination;
- metadata include;
- XML-sensitive/Unicode names;
- full, prefix, suffix, and open-ended ranges;
- invalid/conflicting ranges;
- zero-byte blob;
- all conditional headers;
- wildcard ETag;
- mutation rollback on failed condition;
- concurrent overwrite/read;
- official SDK pager behavior;
- official SDK download stream and range APIs.

## Official Azure SDK compatibility

Use real SDK operations/pagers for:

- list with prefix and pages;
- list hierarchy/delimiter if supported by the SDK;
- download range;
- get properties with conditions;
- conditional overwrite;
- conditional delete;
- decoded `304`/`412`/`416` behavior where exposed.

## Compatibility catalog and docs

Update exact support. Keep leases, snapshots, versions, copy, append/page blobs, and access tiers unsupported.

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
