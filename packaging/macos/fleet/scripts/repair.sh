#!/bin/sh
set -eu

BEACON_BIN="${BEACON_BIN:-/opt/beacon/bin/beacon}"
BEACON_COLLECTOR="${BEACON_COLLECTOR:-/opt/beacon/bin/beacon-otelcol}"
BEACON_ENDPOINT_HARNESSES="${BEACON_ENDPOINT_HARNESSES:-${1:-claude,codex}}"
BEACON_CONTENT_RETENTION="${BEACON_CONTENT_RETENTION:-${2:-full}}"
BEACON_OTLP_GRPC_PORT="${BEACON_OTLP_GRPC_PORT:-${3:-4317}}"
BEACON_OTLP_HTTP_PORT="${BEACON_OTLP_HTTP_PORT:-${4:-4318}}"

exec "$BEACON_BIN" endpoint repair \
  --collector "$BEACON_COLLECTOR" \
  --harness "$BEACON_ENDPOINT_HARNESSES" \
  --content-retention "$BEACON_CONTENT_RETENTION" \
  --otlp-grpc-port "$BEACON_OTLP_GRPC_PORT" \
  --otlp-http-port "$BEACON_OTLP_HTTP_PORT"
