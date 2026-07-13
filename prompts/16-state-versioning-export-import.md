# Task 16 — State format versioning, export, and import

Target agent: Codex with GPT-5.6  
Prerequisite: Tasks 00–15.

Define a safe snapshot format and implement durable export/import for Emulith state.

## Outcome

Users can back up an Emulith instance and restore it into an empty local state directory without corrupting metadata or allowing archive path traversal.

Required CLI:

```bash
emulith export -o emulith-state.tar.gz
emulith import emulith-state.tar.gz
```

Both commands use `--endpoint` and `EMULITH_ENDPOINT` consistently with existing client commands.

## Snapshot format

Create a versioned archive with a stable top-level layout similar to:

```text
emulith-snapshot/
  manifest.json
  metadata/
    emulith.db
  objects/
    ...
```

`manifest.json` must include at least:

```text
format_version
emulith_version
created_at
database_schema_version
files with path, size, and SHA-256
```

Requirements:

- archive format is documented;
- timestamps are UTC;
- paths use forward-slash archive semantics;
- file order is deterministic where practical;
- no absolute paths;
- no symlinks, hard links, devices, or special files;
- checksums cover every imported payload file;
- the format can evolve through explicit version checks.

Use `tar.gz` unless the existing repository has a stronger documented standard.

## Consistent export

An export must represent a consistent point-in-time state.

Implement one of these safe designs:

- a store-level maintenance lock plus SQLite online backup/snapshot;
- an equivalent transactional snapshot mechanism.

Do not copy a live SQLite database and object tree without synchronization.

Export requirements:

- writes are blocked or coordinated for the shortest practical interval;
- object files referenced by exported metadata all exist and match expected size/hash where available;
- temporary files are excluded;
- partial output is removed on failure;
- final archive is atomically renamed into place;
- output file overwrite requires an explicit flag such as `--force`.

## Safe import

Import must:

1. stream and validate archive entries into a temporary staging directory;
2. reject absolute paths, `..`, duplicate entries, case-collision ambiguity where relevant, links, and special files;
3. enforce configurable safe limits for entry count, individual file size, and total uncompressed size;
4. parse and validate `manifest.json`;
5. verify all SHA-256 values and declared sizes;
6. reject unsupported future format versions;
7. validate the database schema before activation;
8. require an empty target state unless `--replace` is explicitly supplied;
9. create a rollback backup or preserve the old state until activation succeeds;
10. atomically activate the imported state;
11. leave the original state untouched when validation or activation fails.

Do not trust archive filenames or database paths.

## Server/admin API

Choose a clear implementation:

- streaming admin endpoints used by the CLI; or
- direct offline state commands that require the server to be stopped.

Prefer online admin endpoints if the current CLI architecture already uses the server. If using online endpoints:

```http
GET  /_emulith/state/export
POST /_emulith/state/import
```

- set correct binary content type;
- use bounded streaming;
- protect concurrent writes/reset;
- document that admin endpoints are unauthenticated and local-only.

Do not load the entire archive into memory.

## Schema migration behavior

- Record the current schema version.
- Export it in the manifest.
- Import may accept older supported schema versions only if migrations are tested.
- Import must not silently downgrade a newer state.
- Add a compatibility policy to `docs/state-format.md`.

## Tests

Cover:

- export/import round trip containing S3 objects, SQS messages, and all currently implemented service metadata;
- binary and zero-byte objects;
- persistence of visibility timestamps and receipt state where intended;
- checksum mismatch;
- truncated archive;
- unsupported format version;
- unsupported/newer database schema;
- path traversal;
- absolute path;
- duplicate entries;
- symlink/hard-link entries;
- zip/tar bomb style size limits;
- failed import leaves old state usable;
- `--replace` behavior;
- deterministic manifest fields excluding expected timestamp/version differences;
- concurrent export while writes occur;
- reset and export/import mutual exclusion.

## Documentation

Add:

```text
docs/state-format.md
```

Document format stability as experimental until a later stable release.

## Verification

```bash
gofmt -w <changed-go-files>
go test ./...
go test -race ./...
go vet ./...
make build
make compatibility
```

Add a temporary-state end-to-end CLI round-trip test or script and run it.


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
