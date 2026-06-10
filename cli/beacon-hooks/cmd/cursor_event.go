package cmd

import (
	"github.com/spf13/cobra"

	"github.com/asymptote-labs/agent-beacon/cli/beacon-hooks/internal/logging"
)

var cursorEventCmd = &cobra.Command{
	Use:   "cursor-event",
	Short: "Record Cursor-specific lifecycle telemetry",
	Run:   runCursorEvent,
}

func init() {
	rootCmd.AddCommand(cursorEventCmd)
}

func runCursorEvent(cmd *cobra.Command, args []string) {
	input, err := readStdinJSON()
	if err != nil {
		outputJSON(emptyResponse)
		return
	}
	sessionID := resolveSessionID(input, platformFlag)
	var logger *logging.Logger
	if sessionID != "" {
		logger = logging.NewSessionLogger("cursor-event", platformFlag, sessionID)
	} else {
		logger = logging.NewLoggerForPlatform("cursor-event", platformFlag)
	}
	switch getFirstStr(input, "hook_event_name", "hookEventName") {
	case "preCompact":
		emitHookEvent(logger, "session.compacting", "session", "info", "Context compaction observed", input, sessionFields(sessionID, input))
	default:
		emitHookEvent(logger, "session.event", "session", "info", "Cursor lifecycle event observed", input, sessionFields(sessionID, input))
	}
	maybeUploadCursorCloudTelemetry(logger)
	outputJSON(emptyResponse)
}
