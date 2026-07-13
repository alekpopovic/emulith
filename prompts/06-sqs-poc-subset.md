# Task 06 — Implement the SQS POC subset

Target agent: Codex with GPT-5.6  
Prerequisite: Tasks 00–05.

Implement a standard-queue SQS subset that works with the current AWS SDK for Go v2.

## Protocol requirement

Modern AWS SDK clients use the SQS AWS JSON protocol by default. Implement it as the primary protocol:

```http
POST /
X-Amz-Target: AmazonSQS.<Operation>
Content-Type: application/x-amz-json-1.0
```

Also retain a minimal legacy AWS Query protocol implementation for the same supported operations where the gateway already recognizes it:

- URL query `GET`;
- form-encoded `POST`;
- XML responses.

Do not force SDK compatibility tests to use the legacy query protocol.

## Supported operations

1. `CreateQueue`
2. `GetQueueUrl`
3. `ListQueues`
4. `SendMessage`
5. `ReceiveMessage`
6. `DeleteMessage`
7. `PurgeQueue`
8. `GetQueueAttributes`

## Queue model

- Standard queues only.
- Account ID: `000000000000`.
- Generate queue URLs from the inbound public endpoint plus a stable path such as:

```text
/000000000000/<queue-name>
```

Do not hardcode `localhost` when the request host is available. Validate forwarded-host handling conservatively; do not trust arbitrary forwarding headers by default.

Queue-name validation must be explicit and tested. Names ending in `.fifo` must be rejected as unsupported for this POC.

Default visibility timeout: `30` seconds. Allow a valid `VisibilityTimeout` queue attribute during creation if straightforward; otherwise document the fixed POC behavior.

## Operation behavior

### CreateQueue

- Require `QueueName`.
- Return the existing queue URL for a compatible repeated request.
- Reject conflicting unsupported attributes clearly.

### GetQueueUrl

- Require `QueueName`.
- Return a deterministic nonexistent-queue error when missing.

### ListQueues

Support optional `QueueNamePrefix`. Return deterministic ordering.

### SendMessage

- Resolve the target using `QueueUrl` from JSON/query input.
- Require `MessageBody`.
- Validate a safe maximum size and valid UTF-8/provider-acceptable text for the POC.
- Persist:
  - stable message ID;
  - body;
  - MD5 of body;
  - creation time;
  - initial visibility.
- Return `MessageId` and `MD5OfMessageBody`.

### ReceiveMessage

Support:

- `MaxNumberOfMessages`, default `1`, valid range `1..10`;
- `VisibilityTimeout`, default queue value;
- short polling only;
- deterministic transaction that selects visible messages and updates their visibility before returning;
- a new unique receipt handle for each successful receive.

Do not return one message to two concurrent receivers.

### DeleteMessage

Delete by current receipt handle and queue. A stale or invalid handle must return `ReceiptHandleIsInvalid`.

### PurgeQueue

Delete all queue messages. A repeated local purge may be allowed even though AWS has timing restrictions; document the deliberate POC deviation.

### GetQueueAttributes

Support at least:

- `ApproximateNumberOfMessages`;
- `ApproximateNumberOfMessagesNotVisible`;
- `QueueArn`;
- `CreatedTimestamp`;
- `VisibilityTimeout`;
- `All`.

Return attribute values as strings in the protocol-appropriate shape.

## Errors

Return protocol-appropriate AWS JSON or Query errors for:

- missing required parameter;
- invalid parameter;
- nonexistent queue;
- invalid receipt handle;
- unsupported FIFO queue;
- unsupported action;
- internal error without SQL details.

Include request IDs and correct media types.

## Explicitly unsupported

- FIFO semantics;
- dead-letter queues/redrive policy;
- message attributes;
- batch operations;
- delay queues/messages;
- long polling;
- permissions/tags;
- change-message-visibility unless added in a later task.

## Tests

### Unit/integration

Cover:

- JSON target parsing and response media type;
- legacy query parsing;
- create/get/list queue;
- send/receive/delete;
- visibility timeout;
- receipt-handle rotation;
- stale receipt rejection;
- purge;
- attributes;
- prefix filtering;
- two concurrent receivers cannot receive the same message;
- malformed/oversized JSON and body rejection.

Use an injectable clock where necessary; avoid slow sleeps.

### AWS SDK for Go v2 compatibility

With an explicit loopback endpoint, static fake credentials, disabled metadata/profile fallback, and a real SQS client using its default protocol:

- CreateQueue;
- GetQueueUrl;
- ListQueues;
- SendMessage;
- ReceiveMessage;
- GetQueueAttributes;
- DeleteMessage;
- PurgeQueue.

Assert that requests use only the in-process server.

## Documentation

Update the exact SQS operation matrix and state that JSON is primary, Query is compatibility fallback, and only standard queues are supported.

## Required verification

```bash
gofmt -w <changed-go-files>
go test ./...
go vet ./...
make build
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
