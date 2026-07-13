# Task 17 — Service registry and provider boundaries

Target agent: Codex with GPT-5.6  
Prerequisite: Tasks 00–16.

Refactor the current service wiring into explicit provider/service boundaries before adding more AWS services.

## Goal

Make it possible to register STS, S3, SQS, and future services without growing a single monolithic router or introducing a premature dynamic plugin system.

This is an internal architecture task. Existing endpoints and compatibility tests must continue to pass unchanged.

## Design constraints

- No dynamic shared-object plugins.
- No RPC plugin framework.
- No reflection-based auto-registration.
- No service locator globals.
- No universal multi-cloud resource abstraction.
- Provider-specific protocol behavior remains provider-specific.
- Dependency injection remains explicit and testable.

## Registry model

Create a small registry with concepts equivalent to:

```go
type Service interface {
    Provider() string
    Name() string
    RegisterRoutes(RouteRegistrar) error
    Health(context.Context) HealthStatus
    Reset(context.Context) error
}

type Registry interface {
    Register(Service) error
    Services() []Service
    Find(provider, name string) (Service, bool)
}
```

Exact interfaces may differ if a clearer design fits existing code.

Requirements:

- duplicate provider/service names fail at startup;
- registration order does not create unstable routing behavior;
- service names and health IDs are deterministic;
- registry exposes immutable/copy views rather than mutable internals;
- lifecycle hooks receive contexts;
- reset ordering is documented;
- a service failure cannot be hidden behind a global “healthy” status.

Do not require every service to implement export/import separately if the shared state layer already handles it. Add hooks only where they provide real value.

## Provider boundary

Create a provider-level composition root for AWS:

```text
providers/aws/
  provider.go
  registry or router integration
```

AWS provider responsibilities:

- register service handlers;
- classify AWS protocols;
- delegate to registered services;
- expose provider health;
- own shared AWS request/error helpers.

Service responsibilities:

- own operation parsing and behavior;
- own compatibility status;
- use narrow state interfaces;
- avoid importing CLI/server packages.

The generic HTTP server must not know S3/SQS/STS operation details.

## Health model

Expand:

```http
GET /_emulith/health
```

Keep backward-compatible top-level fields and add a deterministic service section, for example:

```json
{
  "status": "ok",
  "name": "emulith",
  "version": "dev",
  "services": {
    "aws.s3": {"status": "ok"},
    "aws.sqs": {"status": "ok"},
    "aws.sts": {"status": "ok"}
  }
}
```

Requirements:

- overall status is degraded/unhealthy when a required service is unavailable;
- no internal filesystem/database path leakage;
- stable JSON schema documented as experimental;
- service checks are bounded and do not perform destructive operations.

## Reset integration

The admin reset path must use the registry or a coordinated store reset without double-deleting shared data.

Define and test:

- deterministic hook ordering;
- behavior when one hook fails;
- whether the shared store reset is one atomic operation;
- no partially reported success.

## Tests

Add tests for:

- duplicate registration;
- deterministic listing;
- route registration;
- provider/service lookup;
- health aggregation;
- one unhealthy service;
- reset ordering/error propagation;
- existing S3/SQS/STS compatibility tests unchanged;
- server can start with a minimal fake service registry;
- no import cycle.

## Documentation

Update architecture docs with the composition flow:

```text
CLI -> server -> provider registry -> AWS provider -> service handler -> state
```

Explain why this is not a plugin API and why Azure/GCP will use separate provider protocol layers.

## Verification

```bash
gofmt -w <changed-go-files>
go test ./...
go test -race ./...
go vet ./...
make build
make compatibility
make demo
```


## Execution contract

You are the implementation agent for this task. Complete the work in the current repository; do not stop after writing a plan.

1. Read every applicable `AGENTS.md` before changing files.
2. Inspect the repository, current architecture, tests, dependency versions, and documentation.
3. Run the relevant baseline tests before making changes when practical.
4. State a concise implementation plan, then execute it immediately.
5. Make reasonable non-blocking assumptions instead of asking for confirmation.
6. Keep the change scoped to this task. Do not implement later roadmap items.
7. Preserve existing working behavior and compatibility unless this task explicitly changes it.
8. Prefer simple, maintainable Go and explicit protocol behavior over speculative abstraction.
9. Never use LocalStack, Moto, MinIO, ElasticMQ, Azurite, or another cloud emulator as an Emulith runtime dependency.
10. Never contact real AWS, Azure, or GCP endpoints. Tests must be hermetic and loopback-only.
11. Do not add accounts, entitlement checks, license keys, forced telemetry, analytics, or phone-home behavior.
12. Do not commit, push, tag, publish a release, or open a pull request.
13. Format changed files and run all verification commands applicable to the repository.
14. Fix failures caused by the change. If the environment blocks a command, report the exact limitation and run the closest safe verification.
15. Update compatibility documentation only for behavior backed by executable tests.
16. Finish with:
    - implementation summary;
    - important design decisions;
    - changed files;
    - exact commands run and outcomes;
    - genuine remaining limitations.

Unless a task explicitly changes the release scope, Emulith remains a development/CI emulator, not a production service.
