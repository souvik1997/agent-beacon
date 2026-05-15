# macOS Deployment

This directory contains macOS deployment assets for the Beacon Endpoint Agent,
including Jamf- and Fleet-ready package scripts, policy helpers, inventory
queries, and Jamf Extension Attributes.

Beacon's MDM support is deployment-native: Jamf or Fleet installs and
inventories a local-only endpoint agent, while Beacon continues to write local
JSONL telemetry without requiring a hosted account, remote policy fetch, or MDM
API credentials.

## Package Layout

The package builder assembles this payload:

```text
/opt/beacon/bin/beacon
/opt/beacon/bin/beacon-otelcol
/opt/beacon/scripts/install-endpoint.sh
/opt/beacon/scripts/uninstall-endpoint.sh
/opt/beacon/jamf/extension-attributes/*.sh
/opt/beacon/jamf/scripts/*.sh
/opt/beacon/fleet/queries/*.sql
/opt/beacon/fleet/scripts/*.sh
```

The endpoint install creates system configuration and runtime state:

```text
/Library/Application Support/Beacon/Endpoint/config.json
/Library/Application Support/Beacon/Endpoint/otelcol.yaml
/Library/LaunchDaemons/com.beacon.endpoint.collector.plist
/var/log/beacon-agent/runtime.jsonl
```

## Build A Test Package

Build Beacon and `beacon-otelcol`, then assemble the macOS package:

```bash
cd cli/beacon
make build
cd ../..

cd collector-builder
ocb --config builder.yaml
cd ..

sh packaging/macos/build-pkg.sh
```

Set `PKG_SIGN_IDENTITY` to sign with `pkgbuild`, and set
`NOTARYTOOL_PROFILE` to submit and staple the package with Apple's notary
service.

## Manual Install

```bash
sudo beacon endpoint install --system
beacon endpoint status
beacon endpoint wazuh print-config
```

The macOS package installs Beacon's custom `beacon-otelcol` collector at
`/opt/beacon/bin/beacon-otelcol`, so the CLI discovers it automatically.

## Smoke Test

Run the non-root endpoint smoke test on a macOS host or VM:

```bash
sh packaging/macos/smoke-endpoint.sh
```

The smoke test builds a temporary Beacon binary, uses a temporary `HOME`, runs a
user-mode install with `--no-start`, validates status/Wazuh output, installs
Cursor hooks, and uninstalls while preserving the runtime log for assertions.

## MDM Deployment Model

Use the signed and notarized `.pkg` as the base deployment artifact. The package
installs Beacon under `/opt/beacon`, writes system endpoint configuration, and
loads the local collector LaunchDaemon. The shared endpoint wrapper installs
system-level configuration and writes logs to
`/var/log/beacon-agent/runtime.jsonl`.

Environment variables take precedence, followed by MDM script parameters:

```text
BEACON_ENDPOINT_HARNESSES: default claude,codex
BEACON_CONTENT_RETENTION: default full
BEACON_OTLP_GRPC_PORT: default 4317
BEACON_OTLP_HTTP_PORT: default 4318
BEACON_COLLECTOR: default /opt/beacon/bin/beacon-otelcol when present
BEACON_NO_START: accepts 1/true/yes
```

Recommended rollout:

1. Upload the signed/notarized package to a pilot group.
2. Confirm the LaunchDaemon is running and `beacon endpoint wazuh validate`
   writes a validation event.
3. Add inventory signals for version, service health, log freshness, retention,
   harnesses, and log writability.
4. Scope repair/remediation to unhealthy devices.
5. Broaden deployment in stages after inventory and validation stay healthy.

Cursor hook installation is separate from the base system package because Cursor
configuration is per user. Run the Cursor hook helper only when an interactive
console user is present.

## Jamf Pro

Upload the generated `.pkg` to Jamf Pro and create a Policy scoped to a pilot
Smart Group. The package postinstall performs the default system install, so no
script is required for the common path.

Use `/opt/beacon/jamf/scripts/install.sh` when a policy needs explicit
parameters or a reinstall action:

```bash
/opt/beacon/jamf/scripts/install.sh "$@"
```

Jamf script parameters:

```text
Parameter 4: harnesses, default claude,codex
Parameter 5: content retention, default full
Parameter 6: OTLP gRPC port, default 4317
Parameter 7: OTLP HTTP port, default 4318
Parameter 8: collector path, default /opt/beacon/bin/beacon-otelcol
Parameter 9: no-start flag for install.sh only
```

Use `/opt/beacon/jamf/scripts/repair.sh` as a remediation policy for Macs where
Extension Attributes report a stale or unhealthy install. Use
`/opt/beacon/jamf/scripts/install-cursor-hooks.sh` as a separate user-context
policy for Cursor telemetry.

### Jamf Extension Attributes

Upload scripts from `packaging/macos/jamf/extension-attributes` to inventory:

- Beacon version
- Collector service health
- Last runtime event age in seconds
- Content retention mode
- Configured harnesses
- Runtime log writability

Suggested Smart Groups:

- Beacon version is `not_installed`
- Collector service health is not `running`
- Last runtime event age is greater than `86400`
- Content retention is not `full`
- Runtime log writability is not `writable` or `creatable`

### Jamf Validation

After deploying the package, run:

```bash
sudo /opt/beacon/bin/beacon endpoint status --json
sudo /opt/beacon/bin/beacon endpoint wazuh validate
sudo launchctl print system/com.beacon.endpoint.collector
```

## Fleet

Upload the signed/notarized `.pkg` as Fleet software and scope it to a pilot
team or label. The package postinstall performs the default system install.

Fleet scripts are installed under `/opt/beacon/fleet/scripts`:

```text
install.sh: reinstall or install with optional arguments
validate.sh: status JSON, Wazuh validation, and LaunchDaemon check
repair.sh: preserve logs/config while repairing collector and harness config
uninstall.sh: remove endpoint service files
```

Fleet install script positional arguments:

```text
install.sh argument 1: harnesses, default claude,codex
install.sh argument 2: content retention, default full
install.sh argument 3: OTLP gRPC port, default 4317
install.sh argument 4: OTLP HTTP port, default 4318
install.sh argument 5: collector path, default /opt/beacon/bin/beacon-otelcol
install.sh argument 6: no-start flag, accepts 1/true/yes
```

Fleet repair script positional arguments:

```text
repair.sh argument 1: harnesses, default claude,codex
repair.sh argument 2: content retention, default full
repair.sh argument 3: OTLP gRPC port, default 4317
repair.sh argument 4: OTLP HTTP port, default 4318
```

Add queries from `packaging/macos/fleet/queries` as Fleet policies or labels.
They cover package/service/log/config presence and freshness; run
`/opt/beacon/fleet/scripts/validate.sh` for full CLI-level validation of status,
content retention, harness configuration, Wazuh validation, and launchd health.

- `beacon-version.sql`
- `collector-service-health.sql`
- `last-event-age-seconds.sql`
- `content-retention.sql`
- `configured-harnesses.sql`
- `runtime-log-writable.sql`

Recommended Fleet policies:

- Beacon install state is not `not_installed`
- Collector service health is `running`
- Last runtime event age is less than `86400`
- Endpoint config state is `present`
- Runtime log state is `present`

## Uninstall And Rollback

Use the vendor uninstall helper to remove endpoint service files. Set
`BEACON_KEEP_LOGS=1` or the first uninstall argument to preserve runtime logs
during removal. Set `BEACON_KEEP_CONFIG=1` or the second uninstall argument to
preserve harness telemetry configuration.

```bash
/opt/beacon/jamf/scripts/uninstall.sh "$@"
/opt/beacon/fleet/scripts/uninstall.sh "$@"
```

The endpoint uninstall removes service/configuration state. Package payload
removal remains under the MDM/package receipt lifecycle.

## Troubleshooting

- `launchctl print system/com.beacon.endpoint.collector` fails: run the repair
  helper and check `/Library/LaunchDaemons/com.beacon.endpoint.collector.plist`.
- `endpoint wazuh validate` fails: confirm `/var/log/beacon-agent` exists and is
  writable by the collector.
- No recent runtime events: confirm supported harnesses are configured and the
  local OTLP ports are not in use by another process.
- Cursor hooks are missing: run the Cursor hook helper while a non-root console
  user is logged in.

