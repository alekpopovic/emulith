# Task 55 — Firestore protocol, value model, resource paths, and persistence

Target agent: Codex with GPT-5.6  
Prerequisite: Tasks 00–54.

Add the Firestore gRPC service foundation, typed value model, document path rules, and persistence schema. CRUD methods may remain `Unimplemented` until Task 56.

## Service and database scope

Support only:

```text
projects/{project}/databases/(default)
```

Document mode: Firestore Native-like document API subset.

Reject:

- named databases;
- Datastore mode;
- multi-project access;
- aggregation/listen/admin/index APIs unless implemented later.

Register the public Firestore gRPC service methods required by official clients.

## Resource paths

Implement typed parsing/formatting for:

```text
projects/{project}/databases/(default)/documents
projects/{project}/databases/(default)/documents/{collection}/{document}/...
```

Requirements:

- collection/document segment alternation;
- document path ends on a document segment;
- collection path ends on a collection segment;
- no empty segments;
- percent/UTF-8 handling at the correct gRPC/protobuf layer;
- canonical full name;
- limits for segment/path length;
- reserved names handled deliberately;
- no path traversal concept leakage into filesystem.

## Firestore Value model

Support:

```text
null_value
boolean_value
integer_value
double_value
timestamp_value
string_value
bytes_value
reference_value
geo_point_value
array_value
map_value
```

Requirements:

- preserve signed 64-bit integers exactly;
- distinguish integer and double;
- handle NaN, positive/negative infinity, and negative zero deliberately;
- timestamp UTC normalization and documented precision;
- bytes exact;
- reference validation/canonicalization;
- geo coordinate validation;
- arrays cannot directly contain arrays if matching Firestore constraints, unless represented through maps as allowed;
- recursive map/array depth, field count, and total size limits;
- deterministic canonical encoding;
- deterministic equality and ordering helpers for Task 57;
- distinguish missing field from null;
- reject malformed protobuf combinations.

Do not create a provider-neutral value abstraction.

## Field paths

Implement a parser/formatter for:

- simple identifiers;
- quoted identifiers;
- escaped backticks/backslashes;
- nested fields.

Requirements:

- bounded tokens/depth;
- no regex-only parser;
- round-trip tests;
- reserved/system field handling;
- reusable for masks, transforms, and queries.

## Document model

Persist:

```text
project
database
full document name
parent collection path
document ID
typed fields
create time
update time
document version
```

Use a migration-friendly serialization such as canonical protobuf/JSON with explicit versioning; do not use Go gob.

Add indexes for:

- exact document lookup;
- direct child listing by collection parent;
- collection-group lookup if planned;
- update/version checks.

## gRPC status mapping

Create Firestore-specific helpers for:

```text
InvalidArgument
NotFound
AlreadyExists
FailedPrecondition
Aborted
ResourceExhausted
Unimplemented
Internal
```

No storage/internal leakage.

## Placeholder methods

Register method handlers and return `Unimplemented` where Task 56/57 has not implemented behavior. No method may be marked supported here.

## Tests

Cover:

- database/document/collection path parsing;
- malformed alternation;
- every Value type;
- Int64/double special values;
- timestamp;
- bytes/reference/geo;
- nested map/array limits;
- deterministic canonical encoding/order;
- field path parser;
- persistence migration/reopen;
- fuzz protobuf/value/path inputs with bounds;
- request-size enforcement;
- no payload logging.

## Compatibility catalog and docs

Add Firestore entries as experimental/unsupported. Document default-database-only scope.

## Verification

```bash
gofmt -w <changed-go-files>
go test ./...
go test -race ./...
go vet ./...
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
