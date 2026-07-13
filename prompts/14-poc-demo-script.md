# Task 14 — Create and validate the end-to-end POC demo

Target agent: Codex with GPT-5.6  
Prerequisite: Tasks 00–13.

Create a deterministic demo that proves the first Emulith POC works end to end with real AWS SDK for Go v2 clients.

## Required files

Create a layout equivalent to:

```text
scripts/demo-aws-poc.sh
scripts/README.md
examples/aws-go-sdk-demo/main.go
```

Keep the demo client small and readable. It is demonstration code, not a second test framework.

## Shell orchestration

`scripts/demo-aws-poc.sh` must:

1. enable strict shell mode;
2. resolve the repository root safely;
3. create a temporary data directory and log file;
4. build the `emulith` binary;
5. start `emulith serve` in the background on a configurable loopback port;
6. wait for `/_emulith/health` with a bounded retry loop;
7. run the Go SDK demo client against the explicit endpoint;
8. call `emulith reset`;
9. verify the original resources no longer exist using the SDK demo or a dedicated verification mode;
10. terminate Emulith gracefully;
11. always clean temporary files and child processes through a trap;
12. print a clear success/failure summary.

Defaults:

```text
endpoint: http://127.0.0.1:4566
region: us-east-1
```

Allow overriding the port/endpoint without allowing an accidental real AWS endpoint. Reject non-loopback hosts in the demo.

Set defense-in-depth environment variables:

```text
AWS_ACCESS_KEY_ID=test
AWS_SECRET_ACCESS_KEY=test
AWS_REGION=us-east-1
AWS_EC2_METADATA_DISABLED=true
AWS_SHARED_CREDENTIALS_FILE=<nonexistent temporary path>
AWS_CONFIG_FILE=<nonexistent temporary path>
```

Do not require AWS CLI, Docker, a cloud account, or internet access at demo runtime.

## Go SDK demo flow

Using current AWS SDK for Go v2 clients with explicit local endpoints:

### Health

Verify the Emulith health response before SDK calls.

### STS

Call `GetCallerIdentity` and print the account/ARN.

### S3

- create a uniquely named deterministic demo bucket;
- put a text object;
- put a small binary object;
- get and verify both;
- list with prefix;
- head one object;
- delete one object.

### SQS

- create a standard queue;
- send a message;
- receive and verify it;
- inspect attributes;
- delete the message.

### Reset verification

After CLI reset:

- listing buckets no longer contains the demo bucket;
- getting the queue URL returns the expected nonexistent-queue SDK error.

All comparisons must fail loudly on mismatch.

## Makefile

Add:

```text
make demo
```

The target must invoke the script from any working directory reliably.

## Documentation

`scripts/README.md` must explain:

- what the demo proves;
- prerequisites;
- how to run it;
- environment overrides;
- why it cannot contact real AWS;
- current supported operations and limitations.

Update the main README with a prominent POC demo command:

```bash
make demo
```

## Final POC acceptance gate

The task is complete only when these pass, subject to explicit environment limitations:

```bash
gofmt -w <changed-go-files>
go test ./...
go test -race ./...
go vet ./...
make build
make compatibility
make demo
make docker-build
```

If the demo exposes a product bug, fix the product and add a regression test rather than weakening the demo.

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
