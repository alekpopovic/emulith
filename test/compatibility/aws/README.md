# AWS SDK compatibility suite

`make compatibility` proves that pinned AWS SDK for Go v2 STS, path-style S3, and standard SQS clients can use the documented Emulith subset. It does not compare behavior with real AWS.

The shared harness starts a full in-process server with temporary state and installs a transport that rejects non-loopback destinations. Static fake credentials and explicit service endpoints prevent cloud fallback.

To add an operation, implement its handler and state behavior, then extend the relevant SDK flow with decoded-output and typed-error assertions.
