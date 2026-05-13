# beacon

Public CLI for Beacon Endpoint Agent.

## Build

```bash
make build
```

## Common Commands

```bash
./beacon endpoint install --user
./beacon endpoint status --json
./beacon endpoint discover --json
./beacon endpoint repair --user
./beacon endpoint dashboard --user
./beacon endpoint uninstall --user --keep-logs
```

## Dashboard

```bash
./beacon endpoint dashboard --user
./beacon endpoint dashboard --user --addr 127.0.0.1:8765
./beacon endpoint dashboard --user --open
```

The dashboard reads the configured runtime JSONL log and serves a local,
read-only view on loopback. It has no external network dependency during normal
use.

Use the search bar to find events by action, command, file path, MCP tool,
approval decision, repository, session, or message. Quick filters surface
high-severity events, failures, approvals, MCP activity, file changes, and events
that may need review.

## Wazuh

```bash
./beacon endpoint wazuh print-config --user
./beacon endpoint wazuh install-pack --output ./beacon-wazuh --user
./beacon endpoint wazuh validate --user
```

## Optional Integrations

```bash
./beacon endpoint hooks install --harness cursor --user
./beacon endpoint hooks status --harness cursor --user

./beacon endpoint integrations claude-cowork setup --endpoint https://collector.example.com --user --open
./beacon endpoint integrations claude-cowork setup --ngrok --user --open
./beacon endpoint integrations claude-cowork validate --user --since 10m
```

Claude Cowork monitoring is configured in the Claude admin console at
`https://claude.ai/admin-settings/cowork`. The OTLP endpoint must be reachable
by Claude Cowork, so use a durable public HTTPS Collector endpoint for ongoing
monitoring. The `--ngrok` mode is for short-lived local testing and prints an
authenticated tunnel URL plus the matching `Authorization` header.

## Test

```bash
go test ./...
go test -race ./internal/endpoint/...
```
