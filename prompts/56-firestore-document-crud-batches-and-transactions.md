# Task 56 — Firestore document CRUD, batch writes, transforms, and transactions

Target agent: Codex with GPT-5.6  
Prerequisite: Tasks 00–55.

Implement Firestore document operations and atomic commits through the public gRPC API.

## Supported RPC methods

Implement:

1. GetDocument
2. BatchGetDocuments
3. CreateDocument
4. UpdateDocument
5. DeleteDocument
6. ListDocuments
7. Commit
8. BatchWrite
9. BeginTransaction
10. Rollback

RunQuery is implemented in Task 57.

## Common behavior

- default database only;
- validate all document names/paths;
- enforce document/value/request limits;
- use timestamps from an injectable clock;
- generate monotonically comparable document versions/update times;
- no real IAM/security rules;
- atomic SQLite transactions for commit operations;
- no payload logging.

## GetDocument

Support:

```text
mask
transaction
read_time only where modeled
```

Requirements:

- exact typed fields;
- create/update times;
- missing returns `NotFound`;
- field masks return selected nested fields correctly;
- transaction reads participate in the local transaction model.

## BatchGetDocuments

- preserve streaming response semantics;
- each requested name yields found/missing result;
- validate duplicates deliberately;
- support mask/transaction;
- deterministic request-order or documented server ordering;
- bounded batch size;
- one invalid name rejects safely.

## CreateDocument

- create under a collection parent;
- explicit document ID or generated ID if the official client uses auto-ID client-side;
- duplicate returns `AlreadyExists`;
- update mask semantics where applicable;
- validate before mutation.

## UpdateDocument

Support:

```text
update_mask
current_document.exists
current_document.update_time
```

Requirements:

- full replace when no mask;
- partial nested update with mask;
- field deletion through mask semantics;
- keys/name immutable;
- stale precondition fails without mutation;
- create-on-update only if protocol explicitly permits it.

## DeleteDocument

- support exists/update_time preconditions;
- missing behavior compatible with client;
- atomically delete;
- no recursive subcollection deletion unless explicitly documented; subcollection documents remain independently addressable if the model permits it.

## ListDocuments

Support:

```text
parent
collection_id
page_size
page_token
order_by __name__ only or default
mask
show_missing only if accurate
```

Requirements:

- direct children only;
- deterministic document-name order;
- validated token;
- no duplicates/omissions;
- bounded page size;
- unsupported ordering/options rejected.

## Commit and Write model

Support writes:

```text
update
delete
transform
update_mask
current_document
```

Commit requirements:

- validate all writes before mutation;
- atomic all-or-nothing;
- return commit time and write results;
- one failed precondition aborts everything;
- duplicate/conflicting writes to same document handled deliberately.

## Field transforms

Support a useful subset:

```text
set_to_server_value: REQUEST_TIME
increment
append_missing_elements
remove_all_from_array
maximum
minimum only if type/order semantics are complete
```

Requirements:

- exact integer/double behavior;
- no precision loss;
- array equality uses Firestore value equality;
- transforms applied in write order;
- transform results returned correctly;
- unsupported transforms rejected.

## BatchWrite

Firestore BatchWrite is not fully atomic across writes by service semantics.

Implement:

- bounded write count;
- per-write status;
- correct response shape;
- reuse write validation/mutation code;
- do not falsely claim atomicity if implemented independently.

If the official high-level client uses Commit for batches, keep BatchWrite supported only according to direct client tests.

## Transactions

Implement optimistic local transactions.

### BeginTransaction

Return opaque transaction bytes/token tied to:

```text
database
read version/snapshot marker
created_at
expiration
state
```

Support read-only or read-write options only where modeled.

### Transaction reads

Record read document versions.

### Commit

- validate transaction token;
- verify read versions/preconditions;
- atomically apply writes;
- conflict returns `Aborted`;
- consume transaction token after terminal commit;
- bounded lifetime and count.

### Rollback

Invalidate transaction; idempotent safe behavior.

No distributed/production isolation claim.

## Tests

Cover:

- all CRUD;
- nested masks;
- missing/null distinction;
- every Value type;
- create/update/delete preconditions;
- BatchGet found/missing;
- ListDocuments pagination;
- atomic Commit rollback;
- transforms;
- BatchWrite statuses;
- transaction success;
- concurrent conflict;
- retry path;
- expired/invalid token;
- restart behavior for committed data;
- active transaction snapshot/export policy;
- reset/export/import;
- race detector.

## Official Firestore Go client compatibility

Configure the official client for `FIRESTORE_EMULATOR_HOST` with project `emulith-local` and no credentials.

Test:

- Set/Create/Get/Update/Delete;
- merge/masks;
- `GetAll`;
- write batch;
- transaction with conflict/retry;
- transforms through high-level APIs;
- typed not-found/already-exists/precondition errors.

## Compatibility catalog and docs

Add stable method/flow IDs. Mark transactions/transforms according to exact coverage.

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
