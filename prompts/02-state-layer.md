# Task 02 — Implement the persistent state layer

Target agent: Codex with GPT-5.6  
Prerequisite: Tasks 00–01.

Implement Emulith's reusable persistence foundation using SQLite for metadata and the filesystem for object bodies.

## Outcome

The server can open and close a durable state store, migrations run automatically, reset safely clears service state, and filesystem helpers cannot escape the configured data directory.

## Technical choices

- Use a maintained pure-Go SQLite driver so standard builds do not require CGO.
- Use `database/sql`.
- Keep migration SQL embedded in the binary.
- Configure SQLite for local concurrent use with a sensible busy timeout and foreign keys.
- Do not expose `*sql.DB` broadly unless a narrowly documented internal reason exists. Prefer service-focused store methods or transactions.

## Data layout

```text
<data-dir>/
  emulith.db
  objects/
    aws/
      s3/
  tmp/
```

Create directories with restrictive, cross-platform-safe permissions where supported.

## Initial schema

Create versioned migrations and a schema-version mechanism.

Required metadata tables:

```text
s3_buckets
- name TEXT PRIMARY KEY
- region TEXT NOT NULL
- created_at TIMESTAMP NOT NULL

s3_objects
- bucket TEXT NOT NULL
- key TEXT NOT NULL
- etag TEXT NOT NULL
- size INTEGER NOT NULL
- content_type TEXT
- last_modified TIMESTAMP NOT NULL
- body_path TEXT NOT NULL
- PRIMARY KEY (bucket, key)
- FOREIGN KEY (bucket) REFERENCES s3_buckets(name) ON DELETE CASCADE

sqs_queues
- name TEXT PRIMARY KEY
- url_path TEXT NOT NULL
- visibility_timeout_seconds INTEGER NOT NULL
- created_at TIMESTAMP NOT NULL

sqs_messages
- id TEXT PRIMARY KEY
- queue_name TEXT NOT NULL
- body TEXT NOT NULL
- md5 TEXT NOT NULL
- receipt_handle TEXT
- visible_at TIMESTAMP NOT NULL
- created_at TIMESTAMP NOT NULL
- FOREIGN KEY (queue_name) REFERENCES sqs_queues(name) ON DELETE CASCADE
```

Add indexes needed for receiving visible SQS messages efficiently.

## Store API

Create an API with responsibilities equivalent to:

```go
Open(ctx context.Context, dataDir string) (*Store, error)
(*Store).Close() error
(*Store).Reset(ctx context.Context) error
(*Store).DataDir() string
(*Store).ObjectsRoot() string
(*Store).NewObjectBodyPath(provider, service, namespace, key string) (string, error)
```

Exact names may vary if the design is clearer.

`NewObjectBodyPath` must:

- never place raw untrusted object keys directly into filesystem path segments;
- use a deterministic or unique hashed path;
- remain under the canonical object root;
- work with keys containing slashes, `..`, Unicode, spaces, and percent-like strings.

## Atomic body workflow

Provide helpers or documented conventions for:

1. streaming an incoming body to a temp file;
2. computing size and hash while streaming;
3. `fsync`/close where practical;
4. atomically moving the file to its final location;
5. committing metadata;
6. removing temp/final files when metadata commit fails.

Do not load arbitrary object bodies fully into memory.

## Reset safety

`Reset` must:

- be concurrency-safe;
- delete all service metadata transactionally where practical;
- remove only Emulith-managed object/temp files;
- preserve the configured root directory;
- never follow a symlink out of the root;
- never accept `/`, the home directory, or an ambiguous empty path as a destructive target;
- leave the database usable after reset.

Wire store opening and closing into `emulith serve`. A state-open failure must prevent the server from reporting healthy.

## Tests

Cover at least:

- initial open creates directories and schema;
- reopening preserves metadata;
- migrations are idempotent;
- reset clears every table;
- reset removes managed object files;
- reset preserves the data root and database usability;
- malicious keys cannot escape the root;
- a symlink inside the object tree cannot cause deletion outside the data directory;
- concurrent open/use behavior fails clearly or works deterministically.

Tests must use temporary directories.

## Required verification

```bash
gofmt -w <changed-go-files>
go test ./...
go vet ./...
make build
```

## Execution contract

You are the implementation agent for this task. Complete the task in the current repository; do not stop after producing a plan.

1. Read every applicable `AGENTS.md` before changing files.
2. Inspect the repository, existing conventions, current tests, and dependency versions.
3. Briefly state the implementation plan, then execute it immediately.
4. Make reasonable assumptions when details are non-blocking. Do not ask for confirmation merely to choose between equivalent implementation details.
5. Keep the change tightly scoped to this task. Do not implement later roadmap items.
6. Preserve working behavior and public interfaces unless this task explicitly changes them.
7. Prefer simple, readable Go over speculative abstractions.
8. Do not use LocalStack, Moto, MinIO, ElasticMQ, Azurite, or another cloud emulator as a runtime dependency.
9. Never contact real AWS, Azure, or GCP endpoints. Tests must be hermetic.
10. Do not add accounts, license keys, forced telemetry, analytics, or phone-home behavior.
11. Do not commit, push, create a tag, or open a pull request.
12. Format all changed Go files and run the required verification commands.
13. Fix failures caused by your changes before finishing. If an environment limitation prevents a command, report the exact limitation and run the closest safe verification.
14. Finish with:
    - implementation summary;
    - important design decisions;
    - changed files;
    - commands run and their results;
    - remaining limitations that are genuinely outside this task.
