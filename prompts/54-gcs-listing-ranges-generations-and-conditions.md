# Task 54 — Cloud Storage listing, ranges, generations, and conditions

Target agent: Codex with GPT-5.6  
Prerequisite: Tasks 00–53.

Complete the useful Cloud Storage POC with object listing, ranges, generation semantics, and preconditions.

## Object listing

Implement behavior equivalent to `Objects.List`.

Support:

```text
prefix
delimiter
pageToken
maxResults
startOffset
endOffset
includeTrailingDelimiter
versions=false
```

Requirements:

- deterministic lexical object-name order;
- validated opaque page token;
- no duplicate/omitted object for unchanged state;
- bounded page size;
- correct `nextPageToken`;
- prefixes/delimiter output;
- Unicode/object-name escaping;
- `startOffset` inclusive and `endOffset` exclusive according to documented behavior;
- unsupported `versions=true` rejected;
- active resumable sessions never appear.

## Range download

Support HTTP byte ranges:

```text
bytes=start-end
bytes=start-
bytes=-suffix
```

Requirements:

- `206 Partial Content`;
- exact `Content-Range`;
- correct `Content-Length`;
- `Accept-Ranges: bytes`;
- zero-byte handling;
- invalid/unsatisfiable range returns `416`;
- no whole-object buffering;
- official client's range reader works.

## Generation model

Maintain:

- monotonic generation on content replacement;
- metageneration increment on metadata-only mutation;
- current generation lookup;
- `generation` parameter accepted only for current retained generation;
- old generations return not found because versioning is unsupported;
- generation values persist across restart/export/import.

Do not claim object versioning.

## Preconditions

Implement query/header conditions where applicable:

```text
ifGenerationMatch
ifGenerationNotMatch
ifMetagenerationMatch
ifMetagenerationNotMatch
If-Match
If-None-Match
```

Apply to:

- upload/finalization;
- get metadata/media;
- patch;
- delete.

Requirements:

- evaluate all supplied conditions correctly;
- failed precondition causes no mutation;
- generation `0` create-only semantics if used by the client;
- ETag normalization;
- concurrent overwrite tests;
- return `412`/`304` as appropriate.

## Listing/condition errors

Return Google JSON API errors for:

```text
invalid page token
invalid range
precondition failed
unsupported versions
invalid query parameter
```

## Tests

Cover:

- empty/nonempty list;
- prefix/delimiter;
- trailing delimiter;
- start/end offset;
- multi-page iteration;
- malformed token;
- full/open/suffix ranges;
- invalid range;
- generation increment;
- metageneration increment;
- create-only condition;
- stale generation/metageneration;
- ETag conditions;
- concurrent conditional writers;
- restart/snapshot;
- official iterator and range reader.

## Official Storage client compatibility

Use:

- bucket object iterator;
- prefix/delimiter query;
- page iteration;
- `NewRangeReader`;
- generation/metageneration conditions;
- conditional writer/update/delete;
- decoded `404`/`412`/`416` behavior where exposed.

## Compatibility catalog and docs

List exact supported listing fields and state that object versioning, compose, rewrite, copy, ACL/IAM, and lifecycle remain unsupported.

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
