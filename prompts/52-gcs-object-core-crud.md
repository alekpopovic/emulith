# Task 52 — Google Cloud Storage object core CRUD

Target agent: Codex with GPT-5.6  
Prerequisite: Tasks 00–51.

Implement simple/media object upload, metadata retrieval/update, download, and delete.

## Supported operations

Implement JSON API behavior equivalent to:

1. Objects.Insert using simple media upload
2. Objects.Get metadata
3. Objects.Get media
4. Objects.Patch
5. Objects.Delete

Multipart and resumable uploads are implemented in Task 53.

## Object model

Persist:

```text
project
bucket
object name
generation
metageneration
size
CRC32C
MD5
content type
content encoding
content language
cache control
content disposition
custom metadata
time created
updated
ETag
body path
```

Requirements:

- generation increments monotonically for each new object content in a bucket/name;
- metageneration starts at 1 and increments on metadata patch;
- object versioning remains unsupported; only current content is retained;
- do not reuse raw object names as filesystem paths;
- body paths are managed, safe, and snapshot-compatible.

## Object names

Support logical names containing:

- slashes;
- spaces;
- Unicode;
- dots/`..`;
- percent characters;
- JSON-sensitive characters.

Decode query/path values exactly once and preserve logical names.

## Simple/media upload

Recognize upload endpoints and parameters used by the official client for a direct media upload.

Requirements:

- require existing bucket;
- stream body to a temp file;
- enforce maximum object size;
- calculate size, CRC32C, and MD5 while streaming;
- validate supplied checksum headers/metadata when present;
- atomically replace current content;
- increment generation;
- reset metageneration appropriately;
- apply content headers/metadata supplied through supported mechanisms;
- clean temp/old files on failure/success;
- zero-byte and binary safe;
- no full-body buffering.

If the official client uses multipart by default for metadata-bearing small uploads, support the minimal multipart form in Task 53; keep direct media tests explicit here.

## Get metadata

Return modeled object resource with base64-encoded hashes and string/integer fields in the official JSON shape.

Support:

```text
generation for current generation only
ifGenerationMatch
ifGenerationNotMatch
ifMetagenerationMatch
ifMetagenerationNotMatch
```

Full condition coverage may be completed in Task 54; unsupported supplied conditions must not be ignored.

## Get media

- exact bytes;
- correct content headers;
- checksum/ETag/generation headers as emitted by the API;
- no range support until Task 54;
- stream from disk;
- missing bucket/object returns Google JSON API error.

## Patch metadata

Support:

```text
contentType
contentEncoding
contentLanguage
cacheControl
contentDisposition
metadata
```

Requirements:

- preserve body/generation;
- increment metageneration;
- update time/ETag;
- JSON patch semantics for custom metadata;
- preconditions;
- atomically update;
- reject unsupported ACL/KMS/retention/holds.

## Delete

- current generation only;
- honor generation/metageneration preconditions;
- remove metadata/body safely;
- missing returns `404`;
- no soft delete/version restore.

## Tests

Cover:

- upload/get metadata/download/delete;
- binary/zero-byte;
- checksums;
- Unicode/nested/`..` logical names;
- overwrite generation increment;
- metadata patch/metageneration;
- checksum mismatch;
- missing bucket/object;
- conditions accepted/rejected;
- body-size limit;
- interrupted upload cleanup;
- concurrent overwrite/delete;
- restart;
- reset/export/import;
- filesystem/symlink safety.

## Official client compatibility

Using the Storage Go client:

- writer configured for direct/simple upload when possible;
- read exact bytes;
- attrs;
- update metadata;
- overwrite;
- delete;
- decoded errors.

All HTTP requests must remain loopback-only.

## Compatibility catalog and docs

Mark simple upload and CRUD according to official-client tests. Resumable/high-level writer remains partial until Task 53.

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
2. Inspect the repository, provider registry, listener lifecycle, migrations, tests, compatibility catalog, generated documentation, Docker setup, and dependency versions.
3. Run the relevant baseline checks before editing when practical.
4. State a concise implementation plan, then execute it immediately.
5. Make reasonable non-blocking assumptions instead of asking for confirmation.
6. Keep the change scoped to this task. Do not implement later roadmap items.
7. Preserve all existing AWS and Azure behavior and compatibility unless this task explicitly fixes a defect.
8. Prefer provider-specific protocol implementations over a false universal cloud abstraction.
9. Never use Google Cloud emulators, LocalStack, Azurite, Moto, MinIO, ElasticMQ, or another emulator as an Emulith runtime dependency.
10. Never contact real GCP, AWS, or Azure endpoints. All tests must be hermetic and loopback-only.
11. Do not use Application Default Credentials, `gcloud` user credentials, service-account files, metadata-server probing, workload identity, or ambient cloud profiles in compatibility tests.
12. Do not add accounts, entitlement checks, license keys, forced telemetry, analytics, or phone-home behavior.
13. Do not commit, push, tag, publish a release, or open a pull request.
14. Bound all parsers, protobuf/JSON bodies, gRPC messages, stream buffers, page sizes, upload sessions, query plans, and allocations derived from untrusted input.
15. Never log authorization headers, OAuth tokens, API keys, signed URLs, object bodies, Pub/Sub message data, Firestore document data, or other user payloads.
16. Keep request IDs, error mapping, and compatibility claims deterministic and test-backed.
17. Format changed files and run every verification command applicable to the repository.
18. Fix failures caused by your changes. If the environment blocks a command, report the exact limitation and run the closest safe verification.
19. Update compatibility documentation only for behavior backed by executable official-client compatibility tests.
20. Finish with:
    - implementation summary;
    - important design decisions;
    - changed files;
    - exact commands run and outcomes;
    - compatibility status changes;
    - genuine remaining limitations.

Emulith remains a development/CI emulator, not a production service.
