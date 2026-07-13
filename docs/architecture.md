# Architecture

```text
CLI (serve/reset/apply)
          |
net/http admin mux ---- /_emulith/health, /_emulith/reset
          |
AWS gateway classifier ---- request IDs / safe access logs
       /      |      \
     STS     S3      SQS
              \      /
          state.Store
       SQLite + filesystem
```

`serve` resolves configuration, opens the store, injects handlers, starts a bounded-timeout HTTP server, and closes it after graceful shutdown. Admin routes are registered before the AWS root handler. The gateway distinguishes SQS JSON, Query, and path-style S3 without validating local SigV4 credentials.

SQLite stores metadata with foreign keys, WAL, busy timeout, and transactions; S3 body streams use temporary files and atomic rename. Reset is serialized and removes managed state without following symlink targets. Request IDs are generated locally, and logs exclude credentials and bodies.

The SDK harness starts a temporary full server and rejects non-loopback traffic. Emulith is unauthenticated and must stay on trusted development networks.

Future providers should have their own gateways and protocol models behind explicit routing. They should not be forced into a fictional universal cloud API.
