# Emulith

Emulith is a from-scratch, Docker-first local cloud emulator for development and CI. It requires no cloud account, activation key, license key, telemetry, or phone-home service. It is not intended for production.

The project is an early AWS-compatible POC. Protocol services are not implemented yet.

## Build and run

```bash
make build
./emulith serve
curl http://localhost:4566/_emulith/health
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
| AWS SQS | Not implemented yet |

Emulith does not claim full AWS parity.
