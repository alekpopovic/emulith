# Task 33 — Block Blob core CRUD

Target agent: Codex with GPT-5.6  
Prerequisite: Tasks 00–32.

Implement basic Block Blob storage using SQLite metadata and filesystem bodies.

## Supported operations

1. Put Blob for `BlockBlob`
2. Get Blob
3. Get Blob Properties
4. Delete Blob
5. Set Blob Metadata
6. Set Blob HTTP Headers

Page Blob, Append Blob, Copy Blob, snapshots, versions, leases, tiers, and encryption scopes remain unsupported.

## Persistence model

Persist at least:

```text
account
container
blob name
blob type
body path
size
ETag
last modified
content type
content encoding
content language
cache control
content disposition
content MD5 when supplied/validated
user metadata
created_at
```

Use a composite logical key. Store body paths as Emulith-managed relative paths, not untrusted user names.

## Blob names and paths

Blob names may contain:

- slashes;
- spaces;
- Unicode;
- dots and `..` as literal logical segments;
- percent characters;
- XML-sensitive characters.

Requirements:

- decode URL path exactly once;
- preserve the logical blob name;
- never use raw blob names as filesystem paths;
- prevent traversal and symlink escape;
- reserve internal service/admin routes safely.

## Put Blob

Recognize the request emitted by the official SDK for a simple block-blob upload.

Requirements:

- require an existing container;
- accept `x-ms-blob-type: BlockBlob`;
- reject other types;
- stream request body to a temporary file;
- enforce configurable maximum blob size;
- compute content length and an internal SHA-256;
- validate `Content-MD5` or `x-ms-blob-content-md5` when supplied;
- atomically replace an existing blob;
- update metadata transactionally;
- remove old/failing body files;
- support zero-byte and binary blobs;
- generate new ETag/last-modified;
- reject unsupported transactional/content CRC headers unless implemented correctly;
- handle the official SDK's request body and checksum behavior.

Do not load arbitrary blobs fully into memory.

## Get Blob

Return exact bytes with:

```text
Content-Length
Content-Type
Content-Encoding
Content-Language
Cache-Control
Content-Disposition
Content-MD5 when stored
ETag
Last-Modified
Accept-Ranges: bytes only when range support is implemented
x-ms-blob-type: BlockBlob
x-ms-meta-*
```

Range requests may be explicitly rejected until Task 35.

## Get Blob Properties

Use `HEAD`. Return the relevant headers and no body.

## Delete Blob

- require existing container;
- deleting a missing blob returns `BlobNotFound`;
- remove metadata/body safely;
- reject snapshot/version query options;
- no soft delete.

## Set Blob Metadata

Replace the metadata map atomically and update ETag/last modified.

## Set Blob HTTP Headers

Support replacement/clearing semantics for modeled content headers.

Do not accidentally erase user metadata or body.

## Errors

Return Azure-shaped:

```text
ContainerNotFound
BlobNotFound
InvalidBlobType
InvalidHeaderValue
Md5Mismatch
RequestBodyTooLarge
ConditionNotMet or explicit unsupported condition error
```

No internal path leakage.

## Tests

Cover:

- put/get/head/delete;
- zero-byte/binary;
- nested/Unicode/`..` logical name;
- overwrite and old-file cleanup;
- metadata and HTTP header replacement;
- invalid MD5;
- body-size limit;
- missing container/blob;
- unsupported blob type;
- no symlink/path escape;
- interrupted upload cleanup;
- concurrent put/delete;
- restart persistence;
- reset/export/import.

## Official Azure SDK compatibility

Use the current official Blob SDK:

- upload a small byte slice/stream using the simple path;
- download and compare exact bytes;
- get properties;
- set metadata;
- set HTTP headers;
- overwrite;
- delete;
- verify decoded service errors.

Force explicit loopback endpoint and reject all non-loopback transport.

## Compatibility catalog

Add stable IDs. Mark staged/high-level large upload as unsupported until Task 34.

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
