# Task 45 — GCP provider, gRPC foundation, and local endpoints

Target agent: Codex with GPT-5.6  
Prerequisite: Tasks 00–44 have produced the AWS + Azure `v0.3.0` codebase.

Add the GCP provider foundation with plaintext local gRPC listeners and an HTTP endpoint for Cloud Storage. Do not implement real Pub/Sub, Firestore, or Storage operations yet.

## Goal

Emulith must start the existing AWS/Azure endpoints plus these default GCP endpoints:

```text
GCP Pub/Sub gRPC:       :8085
GCP Firestore gRPC:     :8080
GCP Cloud Storage HTTP: :9023
```

Example local addresses:

```text
127.0.0.1:8085
127.0.0.1:8080
http://127.0.0.1:9023
```

## Configuration

Add settings with the repository's normal flag/environment/default precedence.

Suggested environment variables:

```text
EMULITH_GCP_PUBSUB_ADDR=:8085
EMULITH_GCP_FIRESTORE_ADDR=:8080
EMULITH_GCP_STORAGE_ADDR=:9023
```

Add matching `emulith serve` flags:

```text
--gcp-pubsub-addr
--gcp-firestore-addr
--gcp-storage-addr
```

Requirements:

- each listener binds independently;
- partial startup failure rolls back all listeners started during the attempt;
- shared state is opened and closed exactly once;
- graceful shutdown drains HTTP and gRPC requests with bounded deadlines;
- address conflicts produce actionable errors;
- tests use OS-assigned loopback ports;
- existing AWS/Azure listeners and health behavior remain compatible.

## Provider structure

Add a provider boundary such as:

```text
providers/gcp/
  provider.go
  common/
  pubsub/
  storage/
  firestore/
```

Follow existing repository conventions if they differ.

GCP provider responsibilities:

- register service implementations with the service registry;
- own project/resource-name validation shared by GCP services;
- own request/trace IDs, logging helpers, status/error mapping, and endpoint composition;
- expose provider/service health;
- wire gRPC services and HTTP routes without leaking service-specific logic into the generic server.

## gRPC server

Create a reusable local plaintext gRPC server foundation for Pub/Sub and Firestore.

Requirements:

- h2c/plaintext local use;
- unary and stream interceptors;
- generated request ID/trace context;
- bounded receive/send message sizes;
- panic recovery mapped to safe gRPC `Internal`;
- structured logging for service, method, status, duration, request ID, peer loopback/non-loopback classification, and stream lifecycle;
- never log protobuf payloads by default;
- graceful stop with bounded fallback to hard stop;
- reflection disabled by default unless explicitly enabled for development;
- health service may be registered only if it does not conflict with existing health design;
- no TLS/auth claim in the POC.

## Cloud Storage HTTP root

Add an HTTP listener capable of routing future JSON API and upload requests.

Recognize route families such as:

```text
/storage/v1/b
/storage/v1/b/{bucket}
/storage/v1/b/{bucket}/o
/storage/v1/b/{bucket}/o/{object...}
/upload/storage/v1/b/{bucket}/o
/upload/resumable/{session-id}
```

At this stage, return Google-style JSON errors for recognized but unsupported operations.

## Error mapping

Create shared helpers for:

- gRPC status codes and details where useful;
- Google JSON API errors with HTTP status, reason/code, message, and request ID;
- safe internal error wrapping.

Do not leak SQL, filesystem paths, panic text, or Go type names.

## Logging and redaction

Redact at least:

```text
Authorization
Proxy-Authorization
x-goog-api-key
key
access_token
oauth_token
X-Goog-Signature
X-Goog-Credential
X-Goog-Algorithm
X-Goog-Date
X-Goog-Expires
X-Goog-SignedHeaders
```

Do not log raw resource payloads.

## Health

Extend `/_emulith/health` with:

```text
gcp.pubsub
gcp.firestore
gcp.storage
```

Keep top-level schema backward compatible.

Health checks must verify listener/store readiness only and must not call remote services.

## Placeholder services

Register placeholder Pub/Sub and Firestore gRPC methods only where generated service definitions require registration. Unsupported methods return `Unimplemented` with a request ID.

Storage placeholder routes return Google-style JSON `501`/appropriate unsupported errors.

No GCP operation may be marked supported in this task.

## Tests

Cover:

- all three GCP listeners start/stop;
- AWS/Azure listeners remain healthy;
- one GCP bind failure rolls back other newly started listeners;
- unary and streaming interceptor request IDs;
- payloads/tokens are not logged;
- message-size limit;
- graceful shutdown during a stream;
- Storage route classification;
- Google JSON error shape;
- health aggregation;
- race-free repeated start/stop;
- no network request leaves loopback.

## Compatibility catalog

Add GCP provider/service entries as `experimental` or `unsupported`.

## Documentation

Update architecture, endpoint, and port documentation. State clearly that only transport/provider scaffolding is added in this task.

## Verification

```bash
gofmt -w <changed-go-files>
go test ./...
go test -race ./...
go vet ./...
make build
make compatibility
make compatibility-check
make demo
make demo-azure
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
