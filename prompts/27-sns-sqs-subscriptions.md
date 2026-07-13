# Task 27 — SNS-to-SQS subscriptions and delivery

Target agent: Codex with GPT-5.6  
Prerequisite: Tasks 00–26.

Connect Emulith SNS topics to Emulith SQS standard queues and prove cross-service delivery with real AWS SDK clients.

## Supported operations

Implement or complete:

1. `Subscribe`
2. `Unsubscribe`
3. `ListSubscriptions`
4. `ListSubscriptionsByTopic`
5. `GetSubscriptionAttributes`
6. `SetSubscriptionAttributes` for `RawMessageDelivery`
7. SNS `Publish` delivery to SQS

Support only:

```text
Protocol=sqs
```

Reject email, HTTP/S, Lambda, SMS, application, and Firehose protocols.

## Subscription model

Persist:

```text
subscription ARN
topic ARN
protocol
endpoint queue ARN
raw message delivery flag
created_at
```

Requirements:

- endpoint must identify an existing local Emulith SQS queue;
- queue and topic region/account must match the local instance;
- duplicate subscription behavior is deterministic and documented;
- deleting a topic deletes its subscriptions;
- deleting a queue deletes or invalidates related subscriptions consistently;
- `Unsubscribe` is safe and idempotent where compatible;
- counts in `GetTopicAttributes` become accurate.

Do not enforce real IAM/SQS queue policies in the POC. Document that authorization is intentionally not simulated.

## Queue ARN support

Ensure SQS exposes a stable queue ARN through `GetQueueAttributes`:

```text
arn:aws:sqs:<region>:000000000000:<queue-name>
```

Create shared, typed queue lookup APIs so SNS does not parse SQS SQL tables directly.

Avoid circular package dependencies. Prefer narrow interfaces injected into SNS, such as:

```go
type QueuePublisher interface {
    ResolveQueueARN(ctx context.Context, arn string) (...)
    Enqueue(ctx context.Context, queue ..., body string) (...)
}
```

## Publish delivery

A successful SNS `Publish` must deliver one message to every active SQS subscription.

Default envelope body must be valid JSON with useful SNS fields such as:

```text
Type
MessageId
TopicArn
Subject when present
Message
Timestamp
SignatureVersion
SigningCertURL or omitted if not meaningfully emulated
UnsubscribeURL or a safe local value/omission
MessageAttributes when supported
```

Do not fabricate cryptographic signatures or certificates. Document omitted/local-only fields.

When `RawMessageDelivery=true`, the SQS body is exactly the original SNS message.

Requirements:

- one generated SNS message ID shared across deliveries;
- no duplicate delivery within one local publish transaction;
- delivery is coordinated so partial failure is either represented accurately or rolled back according to documented local behavior;
- avoid deadlocks between SNS and SQS state locks;
- SQS message MD5 matches the final delivered body;
- publish to a topic with no subscriptions still succeeds;
- deleted subscription receives nothing.

## Message attributes

Support a bounded useful subset:

```text
String
Number
Binary
```

Map SNS attributes into the SNS envelope and/or SQS message attributes only if the existing SQS implementation supports them correctly. Otherwise preserve them in the standard SNS JSON envelope and document the limitation. Do not silently drop them.

## Listing and pagination

Implement deterministic pagination for:

- all subscriptions;
- subscriptions by topic.

Use validated opaque tokens and stable ordering.

## Tests

Cover:

- subscribe;
- duplicate subscribe;
- list/get attributes;
- set raw delivery;
- unsubscribe;
- topic/queue deletion cleanup;
- standard JSON envelope;
- raw delivery exact bytes/text;
- subject and message attributes;
- multiple queues receive one message each;
- no subscribers;
- invalid/nonlocal queue ARN;
- unsupported protocol;
- concurrent publish/unsubscribe;
- no deadlock/race;
- persistence across restart.

## Cross-service SDK compatibility flow

Using real SNS and SQS SDK clients:

1. Create topic.
2. Create queue.
3. Read queue ARN.
4. Subscribe queue ARN to topic.
5. Publish.
6. Receive from queue.
7. Decode and verify SNS envelope.
8. Enable raw delivery.
9. Publish again.
10. Receive exact raw message.
11. Unsubscribe.
12. Publish and verify no new message.

Use bounded polling without long sleeps.

## Compatibility catalog and docs

Add stable IDs for SNS operations and cross-service delivery. Clearly state that IAM policies, retries, DLQs, filtering, and non-SQS protocols are unsupported.

## Verification

```bash
gofmt -w <changed-go-files>
go test ./...
go test -race ./...
go vet ./...
make compatibility
make compatibility-check
make build
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
