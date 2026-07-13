# Emulith

Emulith is a from-scratch, Docker-first local cloud emulator for development and CI. It requires no cloud account, activation key, license key, telemetry, or phone-home service. It is not intended for production.

Project documentation: [architecture](docs/architecture.md), [AWS compatibility](docs/compatibility/aws.md), [roadmap](docs/roadmap.md), [contributing](CONTRIBUTING.md), [governance](GOVERNANCE.md), and [security](SECURITY.md).

The project is an early AWS-compatible POC. Protocol services are not implemented yet.

## Build and run

```bash
make build
./emulith serve
curl http://localhost:4566/_emulith/health
```

Run the complete local SDK-backed POC demo with:

```bash
make demo
```

Configuration can be supplied with `--addr`/`EMULITH_ADDR` and `--data-dir`/`EMULITH_DATA_DIR`; flags take precedence.

## Compatibility

| Service | Status |
| --- | --- |
| Admin health | Implemented |
| STS `GetCallerIdentity` | Supported |
| All other STS operations | Unsupported |
| IAM enforcement | Not implemented |
| S3 subset: CreateBucket, ListBuckets, Put/Get/Head/DeleteObject, ListObjectsV2 | Supported (path-style only) |
| S3 multipart, ACLs, policies, versioning, ranges, encryption, website hosting | Unsupported |
| SQS standard queue subset: Create/Get/List, Send/Receive/Delete, Purge, attributes | Supported; AWS JSON primary, Query fallback |
| SQS FIFO, redrive, attributes, batch, delay, long polling, permissions/tags | Unsupported |

Emulith does not claim full AWS parity.

## Compatibility verification

Run `make compatibility` to exercise real AWS SDK for Go v2 clients against an in-process, loopback-only Emulith server. The suite does not contact or compare against real AWS.

Local CI parity commands are `gofmt -l .`, `go vet ./...`, `go test -race ./...`, `make build`, `make compatibility`, and `make docker-build`.

## Reset local state

The unauthenticated reset endpoint is intended only for trusted local development networks and destructively removes all Emulith-managed state:

```bash
emulith reset
emulith reset --endpoint http://localhost:4566
```

## Experimental manifest

Create supported resources through the same public SDK APIs:

```bash
emulith apply -f examples/manifests/aws-basic/emulith.yaml
```

Schema version 1 accepts only S3 buckets and standard SQS queues and may change before a stable release.

## State snapshots

The experimental local-only snapshot API supports consistent export and validated restore:

```bash
emulith export -o emulith-state.tar.gz
emulith import emulith-state.tar.gz
emulith import --replace emulith-state.tar.gz
```

See [the state format](docs/state-format.md). These unauthenticated admin operations must remain on trusted local networks.
