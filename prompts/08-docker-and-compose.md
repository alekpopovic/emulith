# Task 08 — Add production-quality POC container packaging and Compose example

Target agent: Codex with GPT-5.6  
Prerequisite: Tasks 00–07.

Package Emulith as a small, reproducible Docker image and provide a working Docker Compose example.

## Dockerfile

Create a multi-stage `Dockerfile`.

Builder:

- use a pinned Go major/minor image compatible with `go.mod`;
- download modules in a cache-friendly layer;
- build with `CGO_ENABLED=0` when supported by the chosen SQLite driver;
- inject version metadata through linker flags;
- produce `/out/emulith`.

Runtime:

- use a small, maintained base with CA certificates or a static/distroless base if all required behavior works;
- run as a non-root user;
- copy only the binary and necessary license/notice material;
- create or declare `/var/lib/emulith`;
- expose `4566`;
- default command equivalent to:

```text
emulith serve --addr :4566 --data-dir /var/lib/emulith
```

- do not include shell tools solely for convenience;
- add OCI labels for source, license, title, and description without inventing a repository URL that does not exist.

Add `.dockerignore`.

## Makefile

Add:

```text
docker-build
docker-run
```

Use overridable variables for image name and tag, for example:

```text
IMAGE ?= emulith/emulith
TAG ?= dev
```

`docker-run` must mount a named or local development volume and publish `4566`.

## Compose example

Create:

```text
examples/docker-compose/aws-basic/
  docker-compose.yml
  emulith.yaml   # only if already supported; otherwise omit until Task 10
  README.md
```

Compose requirements:

- one Emulith service;
- persistent named volume;
- port `4566:4566`;
- restart behavior appropriate for development;
- no privileged mode;
- no Docker socket mount;
- no cloud credentials;
- optional health check only if it can run using tools actually present in the image.

The example README must show:

```bash
docker compose up --build
curl http://localhost:4566/_emulith/health
```

Show fake local AWS environment:

```bash
AWS_ACCESS_KEY_ID=test
AWS_SECRET_ACCESS_KEY=test
AWS_REGION=us-east-1
AWS_EC2_METADATA_DISABLED=true
```

Show AWS CLI examples with explicit endpoint and note that the CLI version must support the implemented protocols:

```bash
aws --endpoint-url=http://localhost:4566 s3api create-bucket --bucket demo
aws --endpoint-url=http://localhost:4566 s3api put-object --bucket demo --key hello.txt --body README.md
aws --endpoint-url=http://localhost:4566 sqs create-queue --queue-name demo
aws --endpoint-url=http://localhost:4566 sts get-caller-identity
```

Do not claim the example is validated with every AWS CLI version.

## Verification

- Build the Go binary.
- Build the Docker image when Docker is available.
- Start the container, poll the health endpoint, and stop it when the environment permits.
- Confirm the process runs non-root.
- Confirm persistence survives one container restart when practical.
- Always run Go tests even if Docker is unavailable.

Required commands:

```bash
gofmt -w <changed-go-files>
go test ./...
go vet ./...
make build
make docker-build
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
