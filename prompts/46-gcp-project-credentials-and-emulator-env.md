# Task 46 — GCP project, credentials contract, and emulator environment

Target agent: Codex with GPT-5.6  
Prerequisite: Tasks 00–45.

Define Emulith's local GCP project identity, permissive credential behavior, resource-name parsing, and emulator environment CLI.

## Default local identity

Use:

```text
Project ID: emulith-local
Firestore database: (default)
```

Suggested configuration:

```text
EMULITH_GCP_PROJECT_ID=emulith-local
EMULITH_GCP_AUTH_MODE=permissive
```

Only `permissive` is required now.

Requirements:

- validate project IDs using a documented practical subset;
- reject empty/malformed IDs;
- support explicit project override;
- no project registration, billing, organization, or remote lookup;
- no fake strict mode unless token validation is implemented correctly;
- cross-project access is rejected for the first POC unless explicitly configured later.

## Resource-name parser

Implement shared, typed parsing/formatting for:

```text
projects/{project}
projects/{project}/topics/{topic}
projects/{project}/subscriptions/{subscription}
projects/{project}/snapshots/{snapshot}
projects/{project}/databases/{database}
projects/{project}/databases/{database}/documents/{document-path}
```

Requirements:

- parse exactly once;
- reject malformed slash structure;
- preserve UTF-8 logical names where allowed;
- enforce service-specific name limits;
- no path traversal or accidental URL decoding at multiple layers;
- canonical formatter round trips;
- distinguish resource names from arbitrary document paths;
- return safe validation errors.

## Authentication contract

Parse or identify:

```text
Authorization: Bearer ...
x-goog-api-key
signed URL query parameters
anonymous requests
```

POC behavior:

- authentication/authorization is not enforced;
- accepted requests are explicitly local/permissive;
- credentials are never required for emulator-mode official clients;
- supplied tokens/keys are ignored after redaction/classification;
- unknown auth modes fail startup;
- documentation clearly states that IAM and signed URL validation are not simulated.

Do not use Google auth libraries in a way that starts ADC or metadata probing.

## Emulator environment CLI

Implement:

```bash
emulith gcp env
```

Output deterministic `KEY=value` lines suitable for shell export tooling:

```text
PUBSUB_EMULATOR_HOST=127.0.0.1:8085
PUBSUB_PROJECT_ID=emulith-local
FIRESTORE_EMULATOR_HOST=127.0.0.1:8080
FIRESTORE_PROJECT_ID=emulith-local
STORAGE_EMULATOR_HOST=http://127.0.0.1:9023
GOOGLE_CLOUD_PROJECT=emulith-local
```

Options:

```text
--host
--pubsub-port
--firestore-port
--storage-port
--project
```

Requirements:

- stdout contains only machine-consumable environment lines;
- human diagnostics go to stderr;
- IPv4 and bracketed IPv6 formatting;
- reject non-loopback output unless explicitly provided by the user;
- no credentials written;
- dependency-injected writer for tests;
- no shell-specific `export` prefix by default;
- optional `--format=dotenv|shell` only if implemented cleanly and tested.

## Official client smoke configuration

Create test helpers that construct official Go clients using only:

- explicit emulator endpoints;
- insecure/plaintext local transport where required;
- no credentials;
- custom dialer/HTTP transport rejecting non-loopback hosts.

At this stage, it is sufficient to reach provider-shaped `Unimplemented` responses without ADC errors.

## Tests

Cover:

- default/override project ID;
- invalid project IDs;
- resource-name round trip;
- malformed names;
- cross-project rejection;
- document resource path parsing;
- bearer/API-key/signed-query redaction;
- permissive auth behavior;
- unknown auth mode;
- exact CLI env output;
- IPv4/IPv6 formatting;
- no ADC/metadata access;
- official client reaches local endpoint;
- logs contain no token/key values.

## Compatibility catalog and docs

Treat project/auth/env behavior as infrastructure, not a supported cloud operation. Document permissive auth as a deliberate deviation.

## Verification

```bash
gofmt -w <changed-go-files>
go test ./...
go test -race ./...
go vet ./...
make build
make compatibility-check
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
