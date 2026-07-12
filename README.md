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
| AWS STS | Not implemented yet |
| AWS S3 | Not implemented yet |
| AWS SQS | Not implemented yet |

Emulith does not claim full AWS parity.
