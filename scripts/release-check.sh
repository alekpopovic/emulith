#!/bin/sh
set -eu
ROOT=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd); cd "$ROOT"
test -s dist/SHA256SUMS
(cd dist && sha256sum -c SHA256SUMS)
test "$(find dist -maxdepth 1 \( -name '*.tar.gz' -o -name '*.zip' \) | wc -l)" -eq 6
test -s dist/emulith-${VERSION:-snapshot}.spdx.json
grep -q 'SPDX-2.3' dist/emulith-${VERSION:-snapshot}.spdx.json
