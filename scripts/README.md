# AWS POC demo

`make demo` builds and starts Emulith, exercises STS, binary-safe path-style S3, and standard SQS through real AWS SDK for Go v2 clients, resets state, and verifies resources disappeared.

Prerequisites are Go, Make, a POSIX shell, and curl. Docker, AWS CLI, credentials, a cloud account, and runtime internet access are not required.

Override the loopback endpoint with `EMULITH_DEMO_ENDPOINT` or port with `EMULITH_DEMO_PORT`. Non-loopback endpoints are rejected; SDK endpoints, fake credentials, disabled metadata, and nonexistent shared config files prevent real AWS calls.

The demo covers only the documented POC operations. It does not add IAM, multipart S3, FIFO SQS, production security, or full AWS parity.
