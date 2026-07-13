# Contributing

Open an issue for significant behavior changes, then submit a focused pull request with tests and documentation. Run `gofmt`, `go test ./...`, `go vet ./...`, `make build`, and `make compatibility` as applicable.

Contributions must be original or compatibly licensed and include a Developer Certificate of Origin sign-off (`git commit -s`). No CLA is required. By contributing, you license the work under the repository's `AGPL-3.0-or-later` terms.

For a public AWS operation: implement the handler and unit tests, add a loopback-only real SDK compatibility test with a stable ID, update `compatibility/aws.yaml`, then run `make compatibility-report` and `make compatibility-check`. An operation is `supported` only when its default compatibility test passes.

Be respectful and follow [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md). Security reports follow [SECURITY.md](SECURITY.md), not public issues.
