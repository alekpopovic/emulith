# Task 30 — Azure provider gateway and local service endpoints

Target agent: Codex with GPT-5.6  
Prerequisite: Tasks 00–29 have produced the AWS-first `v0.2.0` codebase.

Add the Azure provider foundation without implementing real Blob, Queue, or Table operations yet.

## Goal

Emulith must support an Azure provider with separate local listeners while preserving the existing AWS endpoint and all AWS tests.

Default listeners:

```text
AWS gateway:         :4566
Azure Blob:          :10000
Azure Queue:         :10001
Azure Table:         :10002
```

Default local account path:

```text
/devstoreaccount1
```

Example service endpoints:

```text
http://127.0.0.1:10000/devstoreaccount1
http://127.0.0.1:10001/devstoreaccount1
http://127.0.0.1:10002/devstoreaccount1
```

## Configuration

Add explicit settings with flag, environment, and default precedence consistent with the existing config package.

Suggested variables:

```text
EMULITH_AZURE_BLOB_ADDR=:10000
EMULITH_AZURE_QUEUE_ADDR=:10001
EMULITH_AZURE_TABLE_ADDR=:10002
```

Add matching `serve` flags:

```text
--azure-blob-addr
--azure-queue-addr
--azure-table-addr
```

Requirements:

- each listener binds independently;
- startup fails clearly if a required listener cannot bind;
- partial startup is rolled back cleanly;
- graceful shutdown stops all listeners and closes shared state exactly once;
- address conflicts produce actionable errors;
- tests use OS-assigned ports rather than fixed ports.

Do not replace the existing AWS listener or multiplex Azure protocols onto `:4566`.

## Provider structure

Create a provider-level composition boundary such as:

```text
providers/azure/
  provider.go
  common/
  blob/
  queue/
  table/
```

The exact structure may follow repository conventions.

Azure provider responsibilities:

- register the three Azure services;
- construct listeners/handlers;
- provide shared request IDs, version headers, errors, logging helpers, and account-path extraction;
- expose provider/service health;
- coordinate provider reset through existing shared mechanisms.

Generic server code must not know Blob, Queue, or Table operation details.

## Request metadata

For Azure service responses, add support for common headers where appropriate:

```text
x-ms-request-id
x-ms-version
Date
Server
```

Requirements:

- generate a unique local `x-ms-request-id`;
- echo or negotiate a supported local `x-ms-version` according to a documented policy;
- reject clearly unsupported future API versions when behavior would be unsafe or ambiguous;
- preserve current supported versions in one provider-level policy, not scattered constants;
- do not claim support for every Azure Storage API version.

## Routing

### Blob endpoint

Recognize paths shaped like:

```text
/{account}
/{account}/{container}
/{account}/{container}/{blob...}
```

### Queue endpoint

Recognize:

```text
/{account}
/{account}/{queue}
/{account}/{queue}/messages
/{account}/{queue}/messages/{message-id}
```

### Table endpoint

Recognize service/table/entity paths used by the current official Azure SDK.

At this stage, route recognized requests to explicit provider-shaped `UnsupportedOperation` or `NotImplemented` responses. Do not return generic HTML.

## Error model

Implement shared Azure Storage error helpers supporting XML and JSON/OData response forms as required by each service.

Include:

```text
HTTP status
service error code
human-readable message
request ID
time
```

Never expose SQL, filesystem paths, Go panic text, or secrets.

## Logging

Structured access logs must include:

```text
provider=azure
service
operation
method
sanitized path
status
duration
request ID
```

Redact:

- `Authorization`;
- `x-ms-copy-source` query credentials;
- all SAS query parameters;
- account keys;
- request bodies;
- message/entity/blob content.

## Health

Extend the existing health response with:

```text
azure.blob
azure.queue
azure.table
```

Keep backward compatibility for the top-level health schema.

Health must report listener/store readiness without performing destructive or remote operations.

## Tests

Cover:

- all three listeners start and stop;
- AWS listener remains functional;
- one listener bind failure rolls back the others;
- admin health reflects Azure services;
- account/service route classification;
- unsupported operations return Azure-shaped errors;
- request ID and version headers;
- unsupported API version;
- authorization/SAS values are not logged;
- graceful shutdown with in-flight requests;
- race detector has no listener lifecycle race.

## Compatibility catalog

Add Azure service entries as `experimental` or `unsupported`. No Azure operation may be marked `supported` in this task.

## Documentation

Update architecture and port documentation. State clearly that Azure endpoints exist but operations are not implemented yet.

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
```


## Execution contract

You are the implementation agent for this task. Complete the work in the current Emulith repository; do not stop after writing a plan.

1. Read every applicable `AGENTS.md` before changing files.
2. Inspect the repository, current architecture, migrations, tests, dependency versions, compatibility catalog, and documentation.
3. Run the relevant baseline checks before editing when practical.
4. State a concise implementation plan, then execute it immediately.
5. Make reasonable non-blocking assumptions instead of asking for confirmation.
6. Keep the change scoped to this task. Do not implement later roadmap items.
7. Preserve all existing AWS behavior and compatibility unless this task explicitly fixes a defect.
8. Prefer explicit provider-specific protocol code over a false universal cloud abstraction.
9. Never use Azurite, LocalStack, Moto, MinIO, ElasticMQ, or another emulator as an Emulith runtime dependency.
10. Never contact real Azure, AWS, or GCP endpoints. All tests must be hermetic and loopback-only.
11. Do not use `DefaultAzureCredential`, managed identity probing, user Azure CLI credentials, or ambient cloud profiles in compatibility tests.
12. Do not add accounts, entitlement checks, license keys, forced telemetry, analytics, or phone-home behavior.
13. Do not commit, push, tag, publish a release, or open a pull request.
14. Bound all parsers, request bodies, archive inputs, page sizes, and allocations derived from untrusted input.
15. Never log account keys, authorization headers, SAS tokens, request bodies containing user data, queue messages, entities, or blob bodies.
16. Format changed files and run every verification command applicable to the repository.
17. Fix failures caused by your change. If the environment blocks a command, report the exact limitation and run the closest safe verification.
18. Update compatibility documentation only for behavior backed by executable SDK compatibility tests.
19. Finish with:
    - implementation summary;
    - important design decisions;
    - changed files;
    - exact commands run and outcomes;
    - compatibility status changes;
    - genuine remaining limitations.

Emulith remains a development/CI emulator, not a production service.
