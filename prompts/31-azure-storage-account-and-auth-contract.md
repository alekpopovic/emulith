# Task 31 — Azure Storage account and local authentication contract

Target agent: Codex with GPT-5.6  
Prerequisite: Tasks 00–30.

Define Emulith's local Azure Storage account model, connection strings, and credential parsing behavior.

## Goal

Official Azure SDK clients can be configured against Emulith with one deterministic development account, while Emulith never requires a real Azure identity.

## Default account

Use:

```text
AccountName=devstoreaccount1
```

Provide a deterministic development-only base64 account key suitable for SDK credential construction. Keep it in one documented source of truth.

Requirements:

- key is not a secret and must be labeled development-only;
- users may override account name/key through explicit local configuration;
- account name validation is explicit;
- duplicate account records are prevented;
- account metadata is persisted only if needed by the architecture;
- no cloud account registration or activation.

Suggested variables:

```text
EMULITH_AZURE_ACCOUNT_NAME
EMULITH_AZURE_ACCOUNT_KEY
```

Do not accept an empty key silently when a shared-key client is configured.

## Account path extraction

For all three services, extract and validate the account from the leading path segment:

```text
/{account}/...
```

Requirements:

- percent decoding occurs exactly once at the correct layer;
- malformed encoding is rejected;
- account names cannot alter routing;
- no path traversal;
- requests for unknown accounts return an Azure-shaped account/resource error;
- service root behavior is documented.

## Authentication parsing

Parse enough of these schemes to interoperate with official clients:

```text
SharedKey
SharedKeyLite
SAS query parameters
anonymous request
```

POC behavior:

- parse and classify credentials;
- accept validly shaped local SharedKey/SharedKeyLite requests;
- actual cryptographic signature enforcement may remain disabled;
- accept anonymous requests only according to an explicit local setting;
- SAS parameters may be parsed/redacted without cryptographic validation;
- never claim that authorization or SAS permissions are enforced.

Add a configuration mode such as:

```text
EMULITH_AZURE_AUTH_MODE=permissive
```

Only `permissive` is required now. Reject unknown modes. Do not add a fake “strict” mode that does not verify signatures correctly.

## Security and logging

Never log or expose:

```text
Authorization
AccountKey
sig
se
sp
spr
sv
st
skoid
sktid
skt
ske
sks
skv
```

Sanitize query strings before logs and errors.

Provider errors must not echo the supplied signature or key.

## Connection string CLI

Implement:

```bash
emulith azure connection-string
```

Options:

```text
--host
--blob-port
--queue-port
--table-port
--account-name
--account-key
```

Prefer an existing endpoint/config abstraction rather than duplicating parsing.

Output a connection string equivalent to:

```text
DefaultEndpointsProtocol=http;
AccountName=devstoreaccount1;
AccountKey=<development-key>;
BlobEndpoint=http://127.0.0.1:10000/devstoreaccount1;
QueueEndpoint=http://127.0.0.1:10001/devstoreaccount1;
TableEndpoint=http://127.0.0.1:10002/devstoreaccount1;
```

Requirements:

- one-line default output suitable for command substitution;
- optional human-readable or environment format only if consistent with existing CLI;
- no trailing debug output on stdout;
- support injected writers for tests;
- reject non-local endpoints unless explicitly provided;
- bracket IPv6 hosts correctly.

Optionally add:

```bash
emulith azure env
```

only if it cleanly outputs shell-neutral `KEY=value` lines. Otherwise reserve it for Task 42.

## Official SDK smoke tests

Use current official Azure SDK clients with:

- explicit Emulith service URLs;
- explicit shared-key credential;
- custom transport that rejects non-loopback hosts.

At this stage, prove request parsing reaches the intended service and returns a recognized Azure error rather than failing credential construction or protocol parsing.

Do not use `DefaultAzureCredential`.

## Tests

Cover:

- default account/key;
- override precedence;
- account path extraction;
- malformed percent encoding;
- unknown account;
- SharedKey parsing;
- SharedKeyLite parsing;
- SAS redaction;
- anonymous mode behavior;
- connection string exact shape;
- IPv4 and IPv6 formatting;
- CLI output/exit behavior;
- logs contain no credential material;
- restart preserves account behavior if persisted.

## Compatibility catalog

Account/auth parsing remains infrastructure, not a supported cloud operation. Document permissive auth as a deliberate deviation.

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
