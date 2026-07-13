# Task 26 — SNS topics and Publish

Target agent: Codex with GPT-5.6  
Prerequisite: Tasks 00–25.

Implement a core AWS SNS-compatible subset for local topics and publishing.

## Protocol

SNS uses the AWS Query protocol.

Support:

- query-string requests where already supported by the gateway;
- form-encoded POST requests as the primary path;
- protocol-compatible XML response/error envelopes;
- request IDs;
- bounded form parsing;
- no signature validation for the local POC.

Register SNS through the AWS service registry.

## Supported operations

1. `CreateTopic`
2. `ListTopics`
3. `GetTopicAttributes`
4. `DeleteTopic`
5. `Publish`

No subscription delivery is implemented until Task 27.

## Topic model

- standard topics only;
- account ID `000000000000`;
- region from Emulith configuration, default `us-east-1`;
- topic ARN:

```text
arn:aws:sns:<region>:000000000000:<name>
```

- deterministic metadata persistence;
- lexical topic listing;
- immediate creation/deletion.

Reject FIFO topic names/attributes clearly.

## CreateTopic

- validate a practical SNS topic-name subset;
- repeated creation returns the same ARN;
- reject unsupported attributes that alter semantics;
- persist atomically.

## ListTopics

Implement real pagination using a local opaque `NextToken`:

- deterministic order;
- bounded page size;
- token integrity/versioning;
- reject malformed tokens;
- no duplicate or omitted topic for an unchanged dataset.

Document that mutation during pagination is not snapshot-isolated.

## GetTopicAttributes

Return at least:

```text
TopicArn
DisplayName
SubscriptionsConfirmed
SubscriptionsPending
SubscriptionsDeleted
```

Counts are zero until Task 27.

Return attribute values as strings.

## DeleteTopic

- idempotent local behavior consistent with SDK expectations;
- remove associated future subscription metadata safely;
- no effect on SQS queues.

## Publish

Support:

```text
TopicArn
Message
Subject
MessageStructure
MessageAttributes
```

For this task:

- require a valid existing topic;
- generate and return `MessageId`;
- persist no message when there are no subscribers unless a debug/event history feature already exists;
- support plain text `Message`;
- either implement `MessageStructure=json` validation correctly or reject it explicitly;
- either implement message attributes correctly for future delivery or reject unsupported types explicitly;
- enforce safe message size and valid UTF-8/provider-compatible content.

Do not implement SMS, phone number, target ARN, mobile push, email, or direct endpoint delivery.

## Errors

Return SNS-shaped errors for:

- invalid parameter;
- not-found topic;
- unsupported FIFO;
- oversized message;
- malformed JSON message structure;
- unsupported action;
- internal error.

## State migrations

Add topic metadata with future subscription support in mind, but do not create speculative generic messaging tables if provider-specific tables are clearer.

## Tests

Cover:

- create/idempotent create;
- name validation;
- list pagination;
- get attributes;
- delete;
- publish;
- missing topic;
- message size/UTF-8;
- malformed token;
- request/response XML escaping;
- concurrent create/delete/publish;
- persistence across restart.

## AWS SDK compatibility tests

With a real SNS SDK client:

- CreateTopic;
- ListTopics;
- GetTopicAttributes;
- Publish;
- DeleteTopic;
- typed missing/invalid errors where the SDK exposes them.

## Compatibility catalog and docs

Add exact statuses and state that subscriptions/delivery are not supported until Task 27.

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
