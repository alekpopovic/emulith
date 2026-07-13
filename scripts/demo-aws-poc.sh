#!/bin/sh
set -eu

ROOT=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
ENDPOINT=${EMULITH_DEMO_ENDPOINT:-http://127.0.0.1:${EMULITH_DEMO_PORT:-4566}}
case "$ENDPOINT" in
  http://127.0.0.1:*|http://localhost:*|http://\[::1\]:*) ;;
  *) echo "refusing non-loopback endpoint: $ENDPOINT" >&2; exit 1 ;;
esac
TMP=$(mktemp -d)
PID=
cleanup() {
  if [ -n "$PID" ] && kill -0 "$PID" 2>/dev/null; then kill -TERM "$PID"; wait "$PID" || true; fi
  rm -rf "$TMP"
}
trap cleanup EXIT INT TERM

export AWS_ACCESS_KEY_ID=test AWS_SECRET_ACCESS_KEY=test AWS_REGION=${AWS_REGION:-us-east-1} AWS_EC2_METADATA_DISABLED=true
export AWS_SHARED_CREDENTIALS_FILE="$TMP/nonexistent-credentials" AWS_CONFIG_FILE="$TMP/nonexistent-config"

cd "$ROOT"
make build
ADDR=${ENDPOINT#http://}
./emulith serve --addr "$ADDR" --data-dir "$TMP/data" >"$TMP/emulith.log" 2>&1 &
PID=$!
ready=false
for _ in $(seq 1 50); do
  if curl --fail --silent "$ENDPOINT/_emulith/health" >/dev/null 2>&1; then ready=true; break; fi
  sleep 0.1
done
if [ "$ready" != true ]; then cat "$TMP/emulith.log" >&2; echo "Emulith did not become healthy" >&2; exit 1; fi

go run ./examples/aws-go-sdk-demo --endpoint "$ENDPOINT"
./emulith reset --endpoint "$ENDPOINT"
go run ./examples/aws-go-sdk-demo --endpoint "$ENDPOINT" --verify-reset
echo "Emulith AWS POC demo passed"
