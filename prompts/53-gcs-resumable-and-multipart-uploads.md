# Task 53 — Google Cloud Storage multipart and resumable uploads

Target agent: Codex with GPT-5.6  
Prerequisite: Tasks 00–52.

Implement the upload protocols used by high-level Google Cloud Storage clients.

## Supported upload modes

1. `uploadType=media`
2. `uploadType=multipart`
3. `uploadType=resumable`

Media upload may reuse Task 52.

## Multipart upload

Parse `multipart/related` requests containing:

- JSON object metadata;
- media body.

Requirements:

- bounded boundary/header/part counts;
- exactly the expected parts;
- strict content types;
- stream media rather than buffering arbitrarily;
- validate metadata/name/bucket;
- validate checksums;
- atomically create the object;
- safe malformed request errors;
- no MIME parser abuse or temp-file leaks.

## Resumable session creation

Recognize initiation requests.

Persist session metadata:

```text
session ID
project
bucket
object name
object metadata
expected total size when known
current committed offset
temporary body path
checksum state or recomputation metadata
created_at
updated_at
expiration
generation preconditions
status
```

Return a local session URI under:

```text
/upload/resumable/{session-id}
```

Requirements:

- unguessable session IDs;
- loopback/current public host construction;
- no signed credential in session URI;
- bounded session count and lifetime;
- metadata validated before session creation;
- session survives restart.

## Chunk upload

Support `Content-Range` forms needed by the official client:

```text
bytes start-end/total
bytes start-end/*
bytes */total
bytes */*
```

Requirements:

- enforce exact next offset;
- accept retry of an already committed identical chunk only according to documented idempotency;
- reject overlaps, gaps, and inconsistent total size;
- stream chunk to temp/session file;
- respond `308 Resume Incomplete` with range/progress information while incomplete;
- finalize only when total bytes are complete;
- calculate/validate CRC32C and MD5;
- atomically publish the final object;
- generation/precondition checked again at finalization;
- failed finalization leaves previous object intact;
- clean session/temp data after success or expiration;
- no arbitrary memory growth.

## Status query and recovery

Support client status probes.

- report current committed range;
- allow resume after process restart;
- reject expired/unknown session;
- no cross-bucket/session confusion;
- repeated final request returns deterministic result or safe gone/not-found behavior.

## Concurrency

Define/test:

- two clients writing same session;
- two independent sessions for same object;
- finalization race;
- object overwritten between initiation and completion;
- reset/import while session active;
- cleanup worker shutdown.

Use locking/transactions without deadlocks.

## High-level official client requirement

Use the Storage Go client's normal `Writer` path configured so content is definitely uploaded in multiple chunks/resumable mode.

Test:

- several MiB payload;
- exact download;
- metadata;
- interrupted writer/resume via lower-level request if high-level API cannot expose session;
- retry same chunk;
- wrong offset;
- checksum mismatch;
- restart before completion;
- concurrent sessions.

Do not weaken the test to a hand-written single request.

## Snapshot behavior

Include active resumable sessions in export/import only if the snapshot contract can restore them safely. Otherwise explicitly exclude/cancel sessions during export and document the behavior. Do not silently produce unusable sessions.

## Compatibility catalog and docs

Add operation/protocol statuses and local size/session limits.

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
