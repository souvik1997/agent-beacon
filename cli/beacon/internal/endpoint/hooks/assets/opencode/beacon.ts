// __BEACON_MANAGED_MARKER__
// Beacon endpoint telemetry plugin for opencode.
// Managed by beacon endpoint hooks install --harness opencode.

const beaconCommand = "__BEACON_COMMAND__"
const debugEnabled = process.env.BEACON_OPENCODE_DEBUG === "1"
const forwardedEvents = new Set([
  "session.created",
  "session.idle",
  "session.error",
  "session.diff",
  "command.executed",
  "permission.asked",
  "permission.replied",
  "permission.updated",
])

async function debugLog(client, message, extra) {
  if (!debugEnabled) return
  try {
    await client?.app?.log?.({
      body: {
        service: "beacon-opencode-plugin",
        level: "debug",
        message,
        extra,
      },
    })
  } catch {
    // Debug logging must stay best-effort.
  }
}

async function sendToBeacon(client, payload) {
  try {
    const proc = Bun.spawn(["/bin/sh", "-lc", beaconCommand], {
      stdin: "pipe",
      stdout: "ignore",
      stderr: "ignore",
    })
    proc.stdin.write(JSON.stringify(payload))
    proc.stdin.end()
    const code = await proc.exited
    if (code !== 0) {
      await debugLog(client, "Beacon hook command exited non-zero", { code, type: payload?.type })
    }
  } catch (err) {
    await debugLog(client, "Beacon hook command failed", {
      error: err instanceof Error ? err.message : String(err),
      type: payload?.type,
    })
    // Beacon telemetry must never interrupt opencode execution.
  }
}

function sessionID(value) {
  return value?.sessionID || value?.session_id || value?.id || value?.session?.id || value?.info?.sessionID || ""
}

function modelName(value) {
  const model = value?.model || value?.modelInfo
  if (!model || typeof model === "string") return model || ""
  if (model.providerID && model.modelID) return model.providerID + "/" + model.modelID
  return model.modelID || model.id || model.name || ""
}

export const BeaconEndpointPlugin = async ({ project, directory, worktree, client }) => {
  const context = { project, directory, worktree }
  return {
    "chat.message": async (input, output) => {
      await sendToBeacon(client, {
        type: "chat.message",
        session_id: sessionID(input),
        model: modelName(input),
        input,
        output,
        ...context,
      })
    },
    event: async ({ event }) => {
      const type = event?.type || "event"
      if (!forwardedEvents.has(type)) return

      const properties = event?.properties || {}
      await sendToBeacon(client, {
        type,
        session_id: sessionID(properties),
        model: modelName(properties?.info || properties),
        properties,
        event,
        ...context,
      })
    },
  }
}
