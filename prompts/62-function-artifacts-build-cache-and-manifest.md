# Task 62 — Function artifacts, image build, cache, and manifest

Target agent: Codex with GPT-5.6  
Prerequisite: Tasks 00–61.

Add function declarations, OCI image references, safe local builds, immutable revisions, and a content-addressed build cache.

## Manifest schema

Extend the strict experimental manifest with a top-level `functions` section.

Support image-based functions:

```yaml
functions:
  invoice-handler:
    provider: aws
    image: emulith/invoice-handler:dev
    runtime: custom
    handler: bootstrap
    timeout: 30s
    memory: 256MiB
    cpu: "0.5"
    network: emulith-services
    environment:
      APP_ENV: local
```

Support source builds:

```yaml
functions:
  invoice-handler:
    provider: aws
    source: ./functions/invoice-handler
    dockerfile: Dockerfile
    runtime: custom
    handler: bootstrap
```

Requirements:

- exactly one of `image` or `source`;
- strict unknown-field rejection;
- provider/runtime validation;
- duration/size/CPU bounds;
- environment key validation;
- optional sensitive environment references use an existing local-secret mechanism if present;
- no inline plaintext “secret” feature is invented;
- no host mounts, privileged, devices, or Docker socket fields;
- full manifest validates before mutations/builds.

## CLI

Implement:

```bash
emulith functions build [name]
emulith functions list
emulith functions inspect <name>
emulith functions revisions <name>
emulith functions remove <name>
```

Behavior:

- machine-readable `--output=json` where consistent with existing CLI;
- human output to stdout, diagnostics to stderr;
- `inspect` redacts sensitive environment values;
- `remove` requires an explicit flag if active queued invocations/triggers would be affected;
- no automatic registry push.

## Build context safety

For `source`:

- resolve relative to manifest/repository root;
- reject path traversal and symlink escape;
- honor `.dockerignore`;
- enforce maximum file count and aggregate bytes;
- reject sockets/devices/FIFOs;
- define symlink policy explicitly;
- produce deterministic content digest from included files, modes, paths, Dockerfile, build args, and relevant config;
- never include `.git`, state, credentials, or arbitrary parent files unless explicitly part of context;
- no secret values in build context or logs.

## Docker/BuildKit build

Use the Docker API/BuildKit-capable path where available.

Requirements:

- bounded build timeout;
- cancellation;
- streamed/truncated logs;
- image tag scoped to Emulith local project;
- record image ID/digest;
- do not pull base images unless explicit pull policy allows it;
- no build-time host networking by default where controllable;
- no arbitrary build secrets initially;
- build args are allow-listed and redacted if sensitive;
- failure leaves previous ready revision active;
- successful build creates immutable revision and optionally activates it atomically.

If Docker is unavailable, return an actionable error; unit tests use a fake builder.

## Content-addressed cache

Cache key includes:

```text
source digest
Dockerfile content
build args
target platform
runtime metadata
builder version/config
```

Requirements:

- a cache hit skips rebuild only when referenced image still exists and digest matches;
- stale cache self-heals;
- cache metadata is persisted;
- bounded cleanup policy;
- concurrent identical builds coalesce or safely produce one revision;
- concurrent different builds do not corrupt active revision;
- no cross-function secret leakage.

## Image-based revisions

For `image`:

- inspect locally;
- record immutable image ID/digest;
- reject missing image unless an explicit deferred mode is supported;
- no implicit pull;
- activating a new digest creates a new revision even when tag text is unchanged;
- rollback to older revision is possible through a clear internal/CLI operation if straightforward.

## State and snapshots

Persist:

```text
function definitions
revisions
source digest
image reference/digest
build status
build log reference
active revision
```

Do not include Docker image layers in snapshots.

On import:

- preserve definitions/revisions;
- mark revision unavailable when its image digest is absent;
- do not silently rebuild without explicit user action;
- queued invocations wait/fail according to Task 63 policy.

## Tests

Cover:

- strict manifest validation;
- image/source exclusivity;
- context traversal/symlink;
- `.dockerignore`;
- digest determinism;
- cache hit/stale cache;
- build cancellation/failure;
- active revision preserved on failure;
- concurrent builds;
- image tag digest change;
- redaction;
- Docker unavailable;
- snapshot/import with missing image;
- reset/migration;
- Docker integration build when available.

## Documentation

Add:

```text
docs/functions-manifest.md
docs/function-builds.md
```

Document image-only runtime scope and no registry publishing.

## Compatibility catalog

Function artifact/build behavior remains platform-level experimental.

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
