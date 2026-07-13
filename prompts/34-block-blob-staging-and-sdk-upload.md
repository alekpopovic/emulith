# Task 34 — Staged block upload and high-level SDK compatibility

Target agent: Codex with GPT-5.6  
Prerequisite: Tasks 00–33.

Implement staged Block Blob upload so the official Azure SDK's high-level upload helpers work for multi-block content.

## Supported operations

1. Put Block
2. Put Block List
3. Get Block List

Support only Block Blobs.

## Data model

Persist uncommitted and committed block metadata:

```text
account
container
blob
block ID raw bytes/canonical base64
body path
size
content hash
state: uncommitted or committed
created_at
commit order
```

Requirements:

- block ID canonicalization is deterministic;
- all block IDs in one block list have compatible decoded length where Azure requires it;
- duplicate block IDs are handled according to documented semantics;
- uncommitted blocks survive clean restart;
- committed block list can reconstruct the current blob;
- old/unreferenced block files are garbage-collected safely;
- no raw block IDs become filesystem paths.

## Put Block

Recognize:

```text
comp=block
blockid=<base64>
```

Requirements:

- require existing container;
- validate base64 and decoded ID limits;
- stream to a temp file;
- enforce per-block and aggregate limits;
- validate supplied MD5 where present;
- atomically publish the staged block metadata/body;
- replacing the same uncommitted block ID is deterministic;
- no change to the committed blob until Put Block List;
- cleanup on failure.

## Put Block List

Parse the XML block list with bounded XML decoding.

Support entries:

```text
Latest
Committed
Uncommitted
```

Requirements:

- validate list count and total committed size;
- resolve each requested block exactly;
- fail atomically when any block is missing/invalid;
- assemble the final blob by streaming blocks in requested order;
- do not concatenate the entire blob in memory;
- apply request content headers and metadata;
- generate new ETag/last-modified;
- replace previous committed content atomically;
- preserve/cleanup blocks according to documented Azure-like semantics;
- reject unsupported tags/tier/encryption options.

## Get Block List

Support:

```text
blocklisttype=committed
blocklisttype=uncommitted
blocklisttype=all
```

Return provider-compatible XML with deterministic order and sizes.

## High-level SDK requirement

Use the official Azure Blob SDK's high-level upload method that automatically splits content into blocks.

The compatibility test must not force a single-request path. Configure thresholds/concurrency so it definitely emits multiple Put Block calls followed by Put Block List.

Test:

- content several MiB in size;
- multiple blocks;
- exact download verification;
- metadata/content type;
- overwrite using a different block layout.

## Failure and recovery tests

Cover:

- reordered block list;
- duplicate block ID;
- missing block;
- invalid base64;
- inconsistent decoded block ID length;
- too many blocks;
- oversized aggregate;
- interrupted Put Block;
- failed commit leaves old committed blob intact;
- restart between staging and commit;
- replacement of staged block;
- cleanup of abandoned blocks;
- concurrent commits to the same blob produce one complete valid result;
- no orphan leakage after reset/import.

Use fault injection where necessary to prove rollback behavior.

## State/export/import

Include staged and committed block state in snapshots. On import, validate every block body checksum/path before activation.

## Compatibility catalog and docs

Add operation IDs and document local limits. Mark high-level upload supported only when the real SDK helper passes.

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
