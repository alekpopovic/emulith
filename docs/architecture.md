# Architecture

```text
CLI (serve/reset/apply/export/import)
          |
net/http admin mux ---- health/reset/state snapshot
          |
provider registry -> AWS provider/gateway ---- request IDs / safe access logs
       /      |      \
     STS     S3      SQS
              \      /
          state.Store
       SQLite + filesystem
```

`serve` resolves configuration, opens the store, explicitly registers provider services, starts a bounded-timeout HTTP server, and closes it after graceful shutdown. Duplicate deterministic service IDs fail startup. Health aggregates immutable registry views; reset hooks run in sorted order before one coordinated shared-store reset. Admin routes are registered before the AWS root handler. The gateway distinguishes SQS JSON, Query, and path-style S3 without validating local SigV4 credentials.

SQLite stores metadata with foreign keys, WAL, busy timeout, and transactions; S3 body streams use temporary files and atomic rename. Reset is serialized and removes managed state without following symlink targets. Request IDs are generated locally, and logs exclude credentials and bodies.

The SDK harness starts a temporary full server and rejects non-loopback traffic. Emulith is unauthenticated and must stay on trusted development networks.

The registry is a small internal composition mechanism, not a dynamic plugin API: there is no reflection, shared-object loading, RPC, or global locator. Future Azure/GCP providers should have separate gateways and protocol models behind explicit routing rather than a fictional universal cloud API.
