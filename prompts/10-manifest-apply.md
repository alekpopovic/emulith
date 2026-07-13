# Task 10 — Implement `emulith apply` and a minimal manifest

Target agent: Codex with GPT-5.6  
Prerequisite: Tasks 00–09.

Add a small declarative manifest that creates supported local resources through Emulith's public AWS-compatible APIs.

## Manifest format

Support this exact initial schema:

```yaml
version: 1
project: demo

resources:
  - type: aws.s3.bucket
    name: invoices
    region: us-east-1

  - type: aws.sqs.queue
    name: invoice-events
```

Only `version`, `project`, and `resources` belong at the root.

Resource fields:

```text
type        required
name        required
region      optional only for aws.s3.bucket
```

Reject unknown fields using strict YAML decoding so typos are not silently ignored.

## CLI

Implement:

```text
emulith apply -f emulith.yaml
```

Flags/environment:

```text
-f, --file
--endpoint
EMULITH_ENDPOINT
default endpoint: http://localhost:4566
```

Behavior:

- parse and validate the entire manifest before making changes;
- require `version: 1`;
- require a non-empty project name using a documented safe character set;
- reject duplicate logical resources;
- support only:
  - `aws.s3.bucket`;
  - `aws.sqs.queue`;
- return a useful unsupported-type error;
- apply resources in manifest order;
- make repeated apply idempotent for the supported POC resources;
- print one concise result line per resource plus a summary;
- return non-zero if any resource fails;
- do not silently continue after a failure unless an explicit and documented best-effort mode is implemented, which is not required.

## Dogfooding requirement

Use the same public local AWS-compatible APIs that application SDKs use:

- AWS SDK for Go v2 S3 client for buckets;
- AWS SDK for Go v2 SQS client for queues.

Do not call state-store internals or add a private resource-creation shortcut.

Configure:

- explicit endpoint;
- static fake credentials;
- region;
- path style for S3;
- no metadata/profile fallback;
- local-host safety checks.

## Validation examples

Reject:

- missing/unsupported version;
- empty project;
- missing type/name;
- unknown YAML key;
- unsupported provider/resource;
- `.fifo` queue;
- invalid bucket name;
- duplicate resource identity.

Validation errors should identify the resource index and field.

## Files and examples

Add:

```text
internal/manifest/
examples/manifests/aws-basic/emulith.yaml
examples/manifests/aws-basic/README.md
```

Update the Docker Compose example to optionally mount/use the manifest only after the server is healthy. Do not make the server auto-apply manifests in this task.

## Tests

Cover:

- valid strict parse;
- every validation failure above;
- no API call on validation failure;
- successful apply against a full `httptest` Emulith server;
- second apply succeeds without duplicating resources;
- CLI output and exit behavior;
- endpoint guard rejects accidental public AWS host.

## Documentation

Add a manifest reference and clearly label the schema experimental for `v0.1.0-poc`.

## Required verification

```bash
gofmt -w <changed-go-files>
go test ./...
go vet ./...
make compatibility
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
