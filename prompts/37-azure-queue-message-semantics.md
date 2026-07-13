# Task 37 — Azure Queue message semantics

Target agent: Codex with GPT-5.6  
Prerequisite: Tasks 00–36.

Implement Azure Queue Storage messages, including visibility, pop receipts, updates, TTL, and competing consumers.

## Supported operations

1. Put Message
2. Get Messages
3. Peek Messages
4. Update Message
5. Delete Message
6. Clear Messages

## Message persistence

Persist at least:

```text
account
queue
message ID
body in logical decoded form or exact protocol form
insertion time
expiration time
next visible time
dequeue count
current pop receipt
pop receipt generation/version
last update time
```

Requirements:

- stable message ID;
- new unique pop receipt on every successful receive or update;
- stale receipts are rejected;
- expired messages are not returned and are cleaned safely;
- visibility comparisons use an injectable clock;
- concurrent receivers cannot receive the same visible message;
- no AWS SQS table reuse.

## Encoding contract

Determine how the current official Azure Queue SDK encodes message text and implement the protocol it actually emits.

Requirements:

- XML request/response;
- exact UTF-8 and XML escaping;
- base64 behavior only where the SDK/client option requires it;
- binary-like payloads are supported only through the documented SDK encoding mode;
- no double encoding/decoding;
- message size limit measured at the correct protocol stage;
- invalid XML or encoding returns a provider-shaped error.

Document the supported encoding modes.

## Put Message

Support:

```text
visibilitytimeout
messagettl
```

Requirements:

- defaults follow documented local/Azure-compatible values;
- validate ranges;
- allow never-expire only if the current API version supports it and it is implemented safely;
- return message ID, insertion/expiration time, pop receipt, and next visible time where expected;
- initial pop receipt behavior follows the operation response contract;
- enqueue atomically.

## Get Messages

Support:

```text
numofmessages, 1..32
visibilitytimeout
```

Requirements:

- short polling only;
- select visible messages transactionally;
- increment dequeue count;
- set next-visible time;
- rotate pop receipt;
- deterministic selection by insertion order/message ID;
- return at most requested count;
- no duplicate concurrent delivery.

## Peek Messages

- return up to 32 visible messages;
- do not hide messages;
- do not increment dequeue count;
- do not issue a usable pop receipt;
- preserve order.

## Update Message

Address by message ID and require current pop receipt.

Support:

```text
visibilitytimeout
new message body when supplied
```

Requirements:

- rotate pop receipt;
- update next-visible time;
- stale/missing receipt returns provider-compatible error;
- update is atomic;
- message ID remains stable.

## Delete Message

Require message ID and current pop receipt.

- valid receipt deletes;
- stale/invalid receipt fails;
- deleting expired/missing message returns the correct error;
- no accidental deletion of another queue's message.

## Clear Messages

Atomically delete all queue messages. Repeated clear succeeds according to documented local behavior.

## Errors

Implement/test:

```text
QueueNotFound
MessageNotFound
PopReceiptMismatch
InvalidQueryParameterValue
OutOfRangeInput
MessageTooLarge
InvalidXmlDocument
```

Use the actual service error code expected by the SDK where it differs.

## Tests

Cover:

- send;
- delayed visibility;
- TTL expiration;
- peek semantics;
- receive/hide/dequeue count;
- visibility timeout without sleeps using fake clock;
- update body/visibility;
- pop receipt rotation;
- stale receipt delete/update;
- clear;
- 32-message boundary;
- oversized/invalid body;
- two concurrent receivers;
- persistence/restart with remaining TTL/visibility;
- reset/export/import;
- race detector.

## Official Azure SDK compatibility

Use real Queue SDK calls for the full flow:

1. create queue;
2. enqueue;
3. peek;
4. dequeue;
5. update;
6. dequeue after visibility advance;
7. delete;
8. enqueue multiple;
9. clear;
10. verify empty.

Use bounded polling and an injectable server clock/test hook rather than long sleeps where possible.

## Compatibility catalog and docs

Add exact operation support and deviations: no long polling, no server-side poison queue, no auth enforcement.

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
