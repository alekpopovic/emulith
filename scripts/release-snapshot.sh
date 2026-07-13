#!/bin/sh
set -eu
ROOT=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$ROOT"
VERSION=${VERSION:-snapshot}
COMMIT=${COMMIT:-$(git rev-parse HEAD)}
BUILT=${BUILT:-1970-01-01T00:00:00Z}
OUT=${RELEASE_DIR:-dist}
rm -rf "$OUT"
mkdir -p "$OUT"
for target in linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64 windows/arm64; do
  os=${target%/*}; arch=${target#*/}; name="emulith-${VERSION}-${os}-${arch}"; stage="$OUT/$name"
  mkdir -p "$stage"
  binary=emulith; [ "$os" = windows ] && binary=emulith.exe
  CGO_ENABLED=0 GOOS=$os GOARCH=$arch go build -trimpath -ldflags "-s -w -X main.version=$VERSION -X main.commit=$COMMIT -X main.built=$BUILT" -o "$stage/$binary" ./cmd/emulith
  cp LICENSE NOTICE "$stage/"
  printf 'Emulith %s\nDocumentation: https://github.com/alekpopovic/emulith\n' "$VERSION" > "$stage/RELEASE-README.txt"
  if [ "$os" = windows ]; then (cd "$OUT" && zip -qr "$name.zip" "$name"); else tar -C "$OUT" -czf "$OUT/$name.tar.gz" "$name"; fi
  rm -rf "$stage"
done
(cd "$OUT" && sha256sum ./*.tar.gz ./*.zip > SHA256SUMS)
cat > "$OUT/emulith-${VERSION}.spdx.json" <<EOF
{"spdxVersion":"SPDX-2.3","dataLicense":"CC0-1.0","SPDXID":"SPDXRef-DOCUMENT","name":"emulith-${VERSION}","documentNamespace":"https://github.com/alekpopovic/emulith/sbom/${COMMIT}","creationInfo":{"created":"${BUILT}","creators":["Tool: emulith-release-snapshot"]},"packages":[{"name":"github.com/alekpopovic/emulith","SPDXID":"SPDXRef-Package-emulith","versionInfo":"${VERSION}","downloadLocation":"NOASSERTION","filesAnalyzed":false,"licenseConcluded":"AGPL-3.0-or-later","licenseDeclared":"AGPL-3.0-or-later","copyrightText":"NOASSERTION"}]}
EOF
