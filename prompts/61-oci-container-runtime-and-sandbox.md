# Task 61 — OCI/Docker runtime and local sandbox

Target agent: Codex with GPT-5.6  
Prerequisite: Tasks 00–60.

Implement an OCI container runtime backend for the execution-engine contract using the Docker Engine API. Keep a fully testable fake backend for environments without Docker.

## Goal

Execute user-provided local function images with explicit resource limits, lifecycle cleanup, cold starts, warm reuse, and bounded logs.

This is a trusted local development feature, not a strong hostile multi-tenant sandbox.

## Runtime configuration

Support configuration equivalent to:

```text
runtime endpoint / Docker host
maximum total instances
maximum instances per function
default memory
maximum memory
CPU quota
PIDs limit
ephemeral /tmp size
idle instance timeout
container startup timeout
invocation timeout ceiling
log byte limit
network mode
```

Suggested environment variables must use the existing `EMULITH_*` style.

Do not auto-discover arbitrary remote Docker daemons. A nonlocal daemon endpoint must require explicit user configuration.

## Docker Engine integration

Use the maintained Docker client/API modules already appropriate for the Go version.

Requirements:

- negotiate a compatible API version;
- bounded connection/request timeouts;
- context cancellation;
- clear error when daemon unavailable;
- no shelling out to `docker` for core runtime behavior;
- no registry pull without explicit configuration;
- image must exist locally by default;
- optional pull policy may be `never` initially;
- inspect and record image digest/ID;
- label every managed container for cleanup;
- never delete unrelated containers/images.

## Container security defaults

Create containers with safe local defaults:

```text
non-root user required or explicit development override
no privileged mode
no host PID/IPC
no host network by default
read-only root filesystem where image supports it
writable isolated /tmp
drop all Linux capabilities where possible
no-new-privileges
PIDs limit
memory limit
CPU quota/shares
bounded ulimits
no device passthrough
no arbitrary host mounts
no Docker socket mount
```

Requirements:

- reject requested privileged mode;
- reject host paths outside explicitly allowed build/runtime roots;
- reject bind mounts in function manifests for v0.5;
- do not mount Emulith's state DB;
- provider service access uses an isolated development network or explicit loopback gateway strategy;
- document platform differences on macOS/Windows Docker Desktop.

## Network modes

Implement explicit modes:

```text
none
emulith-services
```

`emulith-services` should allow reaching Emulith's local cloud endpoints through a managed Docker network and stable aliases.

Requirements:

- no unrestricted host network default;
- no arbitrary DNS search/domain injection;
- function cannot reach public internet in `none`;
- for `emulith-services`, document whether internet egress remains possible and avoid claiming isolation if Docker networking cannot block it;
- tests verify provider endpoints are reachable in the selected mode.

## Invocation transport

Implement a generic, bounded transport usable by provider adapters, for example:

- HTTP invocation sidecar endpoint inside the function container; or
- stdin/stdout framed protocol for a custom generic test runtime.

Do not force one provider envelope into another.

Requirements:

- exact request/response byte preservation;
- content type/metadata support;
- response-size limit;
- cancellation;
- timeout;
- malformed/oversized response handling;
- no invocation overlap for single-concurrency instances.

Provider-specific runtime channels are implemented in later tasks.

## Warm pool

Implement:

- cold instance creation;
- ready probe;
- warm reuse;
- idle expiration;
- maximum instances;
- per-function concurrency;
- draining when revision changes;
- remove failed containers;
- no reuse across different function revisions;
- fair acquisition under contention;
- bounded wait queue.

## Lifecycle and cleanup

Handle:

- normal stop;
- invocation timeout;
- context cancellation;
- process/container exit;
- daemon restart;
- Emulith crash/restart;
- orphan scan by labels at startup;
- stale container cleanup;
- shutdown drain with deadline;
- forced kill fallback.

Never kill a container not owned by this Emulith instance/project.

## Logs

Capture stdout/stderr with:

- per-invocation attribution where possible;
- total byte bound;
- truncation marker;
- no unbounded goroutine;
- no secret/environment dumping;
- binary-safe handling;
- retained reference for observability task.

## Tests

### Fake backend

Cover all runtime behavior without Docker:

- cold/warm;
- timeout;
- cancellation;
- crash;
- capacity;
- idle cleanup;
- draining;
- log truncation;
- daemon unavailable;
- orphan classification.

### Docker integration, when available

Use tiny locally built test images:

- successful invocation;
- non-root verification;
- read-only root;
- writable `/tmp`;
- memory/process constraints where portable;
- timeout/forced kill;
- crash;
- warm reuse;
- new revision gets a new container;
- service-network connectivity;
- no host mount/socket;
- cleanup after restart;
- concurrent invocations;
- race/leak checks.

Docker-unavailable environments must skip only Docker integration tests with an explicit reason; fake tests remain mandatory.

## Documentation

Add a security-boundary section:

- Docker daemon access is trusted;
- not suitable for untrusted tenants;
- platform-specific limitations;
- network modes;
- resource-limit caveats.

## Compatibility catalog

Keep OCI runtime `experimental`; no provider invocation support yet.

## Verification

```bash
gofmt -w <changed-go-files>
go test ./...
go test -race ./...
go vet ./...
make build
make compatibility-check
make docker-build
```


## Execution contract

You are the implementation agent for this task. Complete the work in the current Emulith repository; do not stop after writing a plan.

1. Read every applicable `AGENTS.md` before changing files.
2. Inspect the repository, provider registry, listener lifecycle, state schema, migration history, SDK compatibility suite, generated compatibility catalog, Docker setup, release tooling, and documentation.
3. Run the relevant baseline checks before editing when practical.
4. State a concise implementation plan, then execute it immediately.
5. Make reasonable non-blocking assumptions instead of asking for confirmation.
6. Keep the change scoped to this task. Do not implement later roadmap items.
7. Preserve all existing AWS, Azure, and GCP behavior and compatibility unless this task explicitly fixes a defect.
8. Keep provider-specific control planes, invocation envelopes, and event semantics separate. Do not create a false universal cloud-function API.
9. Never depend on LocalStack, Azurite, Google emulators, SAM Local, Azure Functions Core Tools, Functions Framework emulators, Moto, MinIO, ElasticMQ, or another emulator as an Emulith runtime dependency.
10. Never contact real AWS, Azure, GCP, registries, metadata services, or public cloud endpoints during tests. All compatibility and end-to-end tests must be hermetic and loopback-only.
11. Do not use ambient cloud credentials, Docker registry credentials, user profiles, ADC, managed identity, instance metadata, or default credential chains in tests.
12. Do not add accounts, entitlement checks, license keys, forced telemetry, analytics, or phone-home behavior.
13. Do not commit, push, tag, publish a release/image, or open a pull request.
14. Bound all request bodies, event payloads, logs, build contexts, archive inputs, queues, retries, streams, concurrency, and allocations derived from untrusted input.
15. Never log secrets, authorization headers, environment-variable values marked sensitive, event payloads, queue messages, object contents, function request bodies, or function responses by default.
16. Treat access to a Docker/OCI daemon as a trusted local security boundary. Do not claim strong multi-tenant isolation.
17. Keep request IDs, invocation IDs, event IDs, error mapping, and compatibility claims deterministic and test-backed.
18. Format changed files and run every verification command applicable to the repository.
19. Fix failures caused by your changes. If the environment blocks Docker or another command, report the exact limitation and run the closest safe verification using fakes/in-process tests.
20. Update compatibility documentation only for behavior backed by executable SDK or protocol compatibility tests.
21. Finish with:
    - implementation summary;
    - important design decisions;
    - changed files;
    - exact commands run and outcomes;
    - compatibility status changes;
    - security-boundary notes;
    - genuine remaining limitations.

Emulith remains a development/CI emulator, not a production service.
