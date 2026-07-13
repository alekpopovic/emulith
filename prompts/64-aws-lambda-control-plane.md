# Task 64 — AWS Lambda image-based control plane

Target agent: Codex with GPT-5.6  
Prerequisite: Tasks 00–63.

Implement an AWS Lambda-compatible control-plane subset for local image-based custom-runtime functions.

## Supported operations

Implement through the existing AWS gateway/protocol used by current AWS SDK for Go v2:

1. CreateFunction
2. GetFunction
3. GetFunctionConfiguration
4. ListFunctions
5. UpdateFunctionCode
6. UpdateFunctionConfiguration
7. DeleteFunction

`Invoke` is implemented in Task 65.

## Scope

Support:

```text
PackageType=Image
local OCI image URI/tag
x86_64 and arm64 metadata
description
timeout
memory size
environment
reserved concurrency metadata if modeled
role ARN as non-enforced metadata
```

Do not claim support for:

```text
managed language runtimes
ZIP packages unless converted safely in a later task
layers
versions/aliases
VPC
EFS
KMS
code signing
dead-letter config execution
tracing integration beyond metadata
SnapStart
provisioned concurrency
Lambda@Edge
IAM enforcement
```

## Protocol

Implement the REST/JSON routes and response shapes used by the official Lambda client.

Requirements:

- AWS request IDs;
- Lambda-shaped JSON errors;
- bounded bodies;
- API path/version handling;
- accepted SigV4 requests without validation;
- no auth/header logging.

## CreateFunction

Validate:

```text
FunctionName
PackageType=Image
Code.ImageUri
Role shape
Architectures
Timeout
MemorySize
Environment variables
Description
```

Requirements:

- image must resolve to a local inspected image/revision unless explicit deferred mode exists;
- create immutable initial function revision;
- duplicate returns `ResourceConflictException`;
- unsupported fields return `InvalidParameterValueException`;
- state initially becomes `Active` only when image/runtime readiness checks pass;
- failure state is represented accurately;
- role ARN is stored only as metadata;
- sensitive environment values are redacted in logs/inspect output.

Return modeled fields:

```text
FunctionName
FunctionArn
Runtime omitted/appropriate for image
Role
Handler if accepted as metadata
CodeSize or local approximation
Description
Timeout
MemorySize
LastModified
CodeSha256/image digest mapping
Version=$LATEST
State
PackageType=Image
Architectures
RevisionId
Environment
```

Use local account/region conventions.

## Get/List

- exact configured metadata;
- deterministic pagination for ListFunctions;
- `Marker`/`MaxItems`;
- missing returns `ResourceNotFoundException`;
- environment error/result fields only when real.

## UpdateFunctionCode

- accept a new local `ImageUri`;
- resolve image digest;
- create a new immutable revision;
- preserve old active revision if validation fails;
- atomically activate successful revision;
- update code hash/last modified/revision ID;
- no implicit registry pull.

## UpdateFunctionConfiguration

Support a safe subset:

```text
Description
Timeout
MemorySize
Environment
Role metadata
Architectures if runtime-compatible
```

Create a new immutable revision or configuration revision according to execution-engine design. Do not mutate a running revision in place.

## DeleteFunction

- prevent new invocations;
- define handling of queued/in-flight work;
- remove control-plane definition after safe transition;
- preserve audit history according to retention;
- missing returns `ResourceNotFoundException`;
- reject unsupported qualifier/version parameters.

## State/migrations

Map AWS public identity to internal function/revision records without using ARN as sole primary key.

## Tests

Cover:

- create/get/config/list/update/delete;
- duplicate/missing;
- unsupported ZIP/runtime fields;
- image missing;
- image digest change under same tag;
- pagination;
- environment redaction;
- concurrent create/update/delete;
- queued/in-flight delete policy;
- restart;
- reset/export/import;
- authentic migration;
- AWS SDK decoded errors.

## Official AWS SDK compatibility

Use Lambda client with explicit loopback endpoint and static fake credentials:

- create;
- get;
- list;
- update code;
- update configuration;
- delete;
- typed service errors.

No real AWS fallback.

## Compatibility catalog and docs

Add stable Lambda control-plane IDs. Mark image-only/custom-runtime scope explicitly.

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
