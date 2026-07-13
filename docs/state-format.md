# Experimental state snapshot format

Format version 1 is a gzip-compressed tar rooted at `emulith-snapshot/`, containing `manifest.json`, `metadata/emulith.db`, and managed `objects/`. The manifest records UTC creation time, Emulith version, schema version, sizes, and SHA-256 checksums.

Imports reject unknown versions, unsafe or duplicate paths, links, special files, checksum mismatches, truncation, and bounded size/count violations. Import requires empty state unless `--replace` is explicit. Export, import, writes, and reset coordinate through the store maintenance lock. The unauthenticated endpoints are local/trusted-network only. Format stability is experimental.
