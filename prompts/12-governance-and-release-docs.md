# Task 12 — Add open-source governance, security, and release documentation

Target agent: Codex with GPT-5.6  
Prerequisite: Tasks 00–11.

Prepare Emulith for a responsible public `v0.1.0-poc` release without changing product functionality.

## Required files

Create or complete:

```text
CHANGELOG.md
SECURITY.md
CONTRIBUTING.md
GOVERNANCE.md
OPEN_SOURCE_FOREVER.md
TRADEMARKS.md
CODE_OF_CONDUCT.md
DCO.md
docs/architecture.md
docs/compatibility/aws.md
docs/roadmap.md
```

Review `README.md`, `LICENSE`, and `NOTICE` for consistency.

## Licensing

- Project/core identifier: `AGPL-3.0-or-later`.
- Ensure the repository has the complete AGPL v3 license text.
- Use SPDX identifiers in documentation and newly touched source headers only if the project adopts that convention consistently; do not create noisy mass edits.
- Do not add a CLA.
- Contribution policy uses Developer Certificate of Origin sign-off.
- Do not claim that a policy document alone makes relicensing legally impossible.
- Explain clearly that open source allows commercial use and redistribution while the AGPL imposes source-availability obligations under its terms.
- Do not provide individualized legal advice.

## `OPEN_SOURCE_FOREVER.md`

State the project commitments plainly:

- no account or license key to run the core locally or in CI;
- no forced telemetry;
- no closed “correctness” tier for supported protocol handlers;
- compatibility tests remain public;
- core releases use an OSI-approved open-source license;
- governance changes are public;
- future hosted support/services may be paid without making the local core proprietary.

Frame these as governance commitments, not guarantees beyond the license and contributor rights.

## Governance

Document:

- roles: maintainers and contributors;
- transparent decision process;
- issue/PR workflow;
- DCO requirement;
- security exception process;
- release process;
- how maintainers are added/removed;
- no single-company entitlement to private core features;
- how compatibility status is approved.

Keep governance suitable for a small project; do not invent a foundation or committee that does not exist.

## Security policy

Include:

- supported versions;
- private vulnerability reporting placeholder clearly marked for the repository owner to replace before public launch;
- response expectations without unrealistic guaranteed times;
- local emulator threat model;
- warning not to expose Emulith to untrusted networks;
- no production-data recommendation;
- what is out of scope.

Do not invent a working email address.

## Compatibility document

Create an operation-level table for:

### STS
- `GetCallerIdentity`.

### S3
- all operations implemented in Task 05.

### SQS
- all operations implemented in Task 06.

For every operation use an explicit status such as:

```text
Supported
Partial
Unsupported
Experimental
```

Document deviations, addressing/protocol limits, persistence behavior, and development-only scope. Ensure the document matches tests and code exactly.

## Architecture document

Cover:

- CLI/server lifecycle;
- admin routing;
- AWS gateway classification;
- protocol handlers;
- state store;
- SQLite/filesystem atomicity;
- request IDs and logging;
- SDK compatibility harness;
- security boundaries;
- how future providers can be added without pretending one universal cloud API exists.

Include a Mermaid diagram only if repository rendering supports it; otherwise use an ASCII diagram.

## Roadmap

Use non-binding milestones:

```text
v0.1 — S3/SQS/STS POC hardening
v0.2 — DynamoDB, SNS, CloudWatch Logs candidates
v0.3 — Azure Blob/Queue/Table candidates
v0.4 — GCP Pub/Sub/Storage/Firestore candidates
```

Mark all future scope as tentative and driven by compatibility tests.

## Changelog

Add:

- `Unreleased`;
- `0.1.0-poc` draft section;
- notable capabilities and limitations;
- no fabricated release date unless the current release date is explicitly part of repository context.

## Verification

- Check all internal links.
- Check commands against actual Make targets and CLI help.
- Check every support claim against tests and handlers.
- Run:

```bash
go test ./...
go vet ./...
make build
make compatibility
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
