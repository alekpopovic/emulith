# Upgrade from v0.2 to v0.3

The store migration is forward-only and idempotent. Existing AWS metadata and object bodies are preserved; Azure tables and GCP metadata are created empty when upgrading an older state. Snapshots with a newer schema are rejected. Always keep a backup before import and use `--replace` explicitly.

Azure and GCP authentication is local development/permissive mode only; no IAM or remote credential lookup is performed.
