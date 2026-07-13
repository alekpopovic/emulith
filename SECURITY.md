# Security policy

## Supported versions

Until the first public release, only the current `main` branch receives security fixes. After release, this table will identify supported versions.

## Reporting a vulnerability

Before public launch, the repository owner must replace this placeholder with a private reporting channel. Until then, do not publish exploit details; contact the repository owner privately through an established channel.

Maintainers will acknowledge and assess reports as availability permits, coordinate a fix when appropriate, and communicate realistic progress without guaranteed response deadlines.

## Threat model

Emulith is an unauthenticated local emulator for trusted development and CI networks. Do not expose it to untrusted networks or store production data, secrets, or regulated data in it. Production deployment, hostile multi-tenancy, IAM enforcement, and real-cloud security guarantees are out of scope.
