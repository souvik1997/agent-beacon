#!/bin/sh
set -eu

BEACON_BIN="${BEACON_BIN:-/opt/beacon/bin/beacon}"

"$BEACON_BIN" endpoint status --json
"$BEACON_BIN" endpoint wazuh validate

if command -v launchctl >/dev/null 2>&1; then
  launchctl print system/com.beacon.endpoint.collector >/dev/null
fi
