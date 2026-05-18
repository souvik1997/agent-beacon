# opencode payload fixtures

This directory is reserved for captured opencode plugin payload fixtures.

Capture representative payloads for:

- `chat.message`
- `session.created`
- `session.idle`
- `session.error`
- `session.diff`
- `command.executed`
- `permission.asked`
- `permission.replied`
- `permission.updated`

Keep secrets and real prompt content out of fixtures. Use sanitized payloads that
preserve the field shape needed by `beacon-hooks --platform opencode
opencode-event`.
