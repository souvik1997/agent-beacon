#!/bin/sh
set -eu

BEACON_UNINSTALL_SCRIPT="${BEACON_UNINSTALL_SCRIPT:-/opt/beacon/scripts/uninstall-endpoint.sh}"

BEACON_BIN="${BEACON_BIN:-/opt/beacon/bin/beacon}" \
BEACON_KEEP_LOGS="${BEACON_KEEP_LOGS:-${1:-0}}" \
BEACON_KEEP_CONFIG="${BEACON_KEEP_CONFIG:-${2:-0}}" \
  "$BEACON_UNINSTALL_SCRIPT"
