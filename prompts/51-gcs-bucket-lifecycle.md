# Task 51 — Google Cloud Storage bucket lifecycle

Target agent: Codex with GPT-5.6  
Prerequisite: Tasks 00–50.

Implement a useful subset of the Google Cloud Storage JSON API for bucket lifecycle.

## Supported operations

Implement JSON API behavior equivalent to:

1. Buckets.Insert
2. Buckets.Get
3. Buckets.List
4. Buckets.Patch
5. Buckets.Delete

## Bucket model

Persist:

```text
project
bucket name
location
storage class
labels
custom metadata where applicable
time created
updated time
metageneration
ETag
versioning enabled flag only if supported
```

Use provider-specific versioned migrations.

## Validation

Implement practical GCS bucket-name validation:

- length;
- lowercase/domain-like characters;
- no invalid separators;
- IP-like names handled according to documented subset;
- no internal route collision;
- global uniqueness only within the local Emulith project/instance.

Return Google JSON API errors with appropriate HTTP status/reason.

## Insert

Support request fields:

```text
name
location
storageClass
labels
```

Requirements:

- project query parameter must match local project;
- create atomically;
- duplicate returns `409`;
- default location/storage class documented;
- reject unsupported IAM, retention lock, autoclass, soft-delete, hierarchical namespace, lifecycle rules, encryption keys, billing/requester-pays, and website fields rather than silently applying them;
- return a complete modeled bucket resource.

## Get

- return modeled fields;
- missing returns `404`;
- support projection query only if the official client emits it and behavior is clear.

## List

Support:

```text
project
prefix
pageToken
maxResults
```

Requirements:

- deterministic lexical order;
- validated opaque token;
- bounded page size;
- correct `nextPageToken`;
- no duplicate/omitted bucket for unchanged state;
- reject cross-project listing.

## Patch

Support a narrow field-mask-like JSON merge subset:

```text
labels
storageClass only if local semantics are defined
```

Requirements:

- use `ifMetagenerationMatch` / `ifMetagenerationNotMatch` when supplied;
- increment metageneration;
- update ETag/time;
- reject unsupported fields;
- patch semantics distinguish absent vs null/delete for labels;
- failed precondition causes no mutation.

## Delete

- require empty bucket unless forced deletion is a separate explicit admin action;
- missing returns `404`;
- nonempty returns a conflict/precondition error;
- honor metageneration preconditions;
- immediate local deletion.

## Auth/IAM deviation

No IAM policy enforcement or signed URL validation. Document permissive local behavior.

## Tests

Cover:

- valid/invalid names;
- insert/get/list/patch/delete;
- duplicate;
- prefix/pagination;
- labels;
- metageneration conditions;
- nonempty delete;
- unsupported fields;
- cross-project rejection;
- JSON error shape/request ID;
- concurrency;
- restart;
- reset/export/import;
- no collision with S3/Azure Blob names.

## Official Storage Go client compatibility

Use an official Storage client configured only for the local endpoint:

- create bucket;
- attrs/get;
- list pager/iterator;
- update supported labels;
- delete;
- decoded duplicate/not-found/precondition errors.

No ADC or real GCP network.

## Compatibility catalog and docs

Add stable IDs and exact deviations. Object operations remain unsupported until Task 52.

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
