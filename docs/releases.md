# Releases

Release tags use semantic versions such as `v0.2.0`; prereleases do not receive a `latest` container tag. `make release-snapshot release-check` creates and validates six OS/architecture archives, SHA-256 checksums, embedded version metadata, and an SPDX JSON SBOM without publishing.

Tagged GitHub workflows must build from the tagged commit, use short-lived CI identity for signing/provenance, and publish only under tag-scoped permissions. Consumers verify archives with `sha256sum -c SHA256SUMS` and verify platform-provided attestations against the repository identity. Supported binaries are Linux and macOS on amd64/arm64 and Windows on amd64/arm64; container builds target Linux amd64/arm64.

Corrections use a new version rather than replacing immutable assets. A broken release may be marked withdrawn while its audit history remains. Release automation does not change the `AGPL-3.0-or-later` license.
