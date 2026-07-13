# Task 15 — Release automation and verifiable artifacts

Target agent: Codex with GPT-5.6  
Prerequisite: Tasks 00–14 have produced a working `v0.1.0-poc` codebase.

Build a safe, reproducible release pipeline for Emulith without publishing anything during this task.

## Outcome

The repository can create versioned binaries, checksums, an SBOM, and multi-platform container artifacts from a tagged commit. Pull requests only validate the release configuration; they never publish.

## Versioning

Use a single source of release version derived from a Git tag such as:

```text
v0.1.0-poc
```

Inject at build time:

```text
version
commit SHA
build timestamp
```

Requirements:

- `emulith version` shows all three values;
- local development builds remain deterministic enough to identify themselves as development builds;
- do not derive the version from an untrusted file in a pull request during a privileged workflow;
- document the expected semantic version/tag convention.

## Binary artifacts

Create release builds for at least:

```text
linux/amd64
linux/arm64
darwin/amd64
darwin/arm64
windows/amd64
windows/arm64
```

Produce archive names that include project, version, OS, and architecture.

Include in every archive:

- binary;
- `LICENSE`;
- `NOTICE`;
- concise release README or link/reference to project docs.

Create a `SHA256SUMS` file covering all release archives.

Prefer a small, transparent Go-based or standard shell build path. A release tool such as GoReleaser is acceptable only if configuration is pinned, understandable, and validated locally. Do not add a large release dependency without justification.

## SBOM

Generate a machine-readable SBOM for:

- the source/binary release;
- the container image where supported.

Use a standard format such as SPDX JSON or CycloneDX JSON. Pin the action/tool version. Do not download and execute unverified scripts.

## Container release

Prepare a multi-architecture image build for:

```text
linux/amd64
linux/arm64
```

Expected tags for an actual release:

```text
<registry>/<owner>/emulith:v0.1.0-poc
<registry>/<owner>/emulith:0.1
<registry>/<owner>/emulith:0
```

Do not create or push `latest` for a POC prerelease unless repository policy explicitly requires it.

Add OCI labels:

- title;
- description;
- source;
- revision;
- version;
- licenses.

Do not hardcode a fictional organization or registry path. Use repository metadata or overridable workflow variables.

## Signing and provenance

Prepare artifact signing and build provenance using keyless CI identity where the hosting platform supports it.

Requirements:

- pull-request jobs do not request signing or package-write permissions;
- release publishing is restricted to protected tag/ref conditions;
- permissions are least privilege;
- no long-lived signing key or registry password is committed;
- document how a consumer verifies checksums, signatures, and provenance.

If signing cannot be tested locally, validate configuration syntax and keep the workflow disabled from publishing except on an explicit release tag.

## GitHub Actions

Add or update workflows with two distinct behaviors.

### Pull request / normal push

- build release configuration in snapshot mode;
- validate archives;
- validate checksums;
- validate Docker build;
- generate SBOM;
- never publish;
- read-only permissions.

### Release tag

- build once from the tagged commit;
- publish GitHub Release assets;
- optionally publish the container image if repository package settings support it;
- sign artifacts/provenance;
- require no cloud credentials.

Do not use `pull_request_target`.

## Local commands

Add Make targets or scripts equivalent to:

```bash
make release-snapshot
make release-check
```

They must not publish.

## Documentation

Add `docs/releases.md` covering:

- tag format;
- local snapshot build;
- release workflow;
- verification instructions;
- architecture support;
- prerelease policy;
- rollback/correction policy;
- the fact that release automation does not change the AGPL license.

## Tests and verification

Verify:

```bash
go test ./...
go test -race ./...
go vet ./...
make build
make compatibility
make release-snapshot
make release-check
make docker-build
```

Inspect generated archives, metadata, checksums, and SBOM contents. Do not create a Git tag or publish artifacts.


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
