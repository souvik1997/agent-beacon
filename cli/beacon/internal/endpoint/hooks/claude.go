package hooks

import (
	"fmt"
	"os"
	"path/filepath"
)

type ClaudeOptions struct {
	Level    Level
	LogPath  string
	UserMode bool
}

type ClaudeStatus struct {
	Installed    bool   `json:"installed"`
	BinaryPath   string `json:"binary_path,omitempty"`
	SettingsPath string `json:"settings_path,omitempty"`
	Message      string `json:"message,omitempty"`
}

var claudeRuntime = hookRuntime{
	displayName: "Claude Code",
	configPath:  claudeSettingsPath,
	install:     installClaudeSettings,
	uninstall:   removeClaudeEndpointHooks,
	isInstalled: isClaudeInstalledAt,
}

func InstallClaude(opts ClaudeOptions) (ClaudeStatus, error) {
	status, err := installRuntimeHooks(claudeRuntime, RuntimeOptions(opts))
	if err != nil {
		return ClaudeStatus{}, err
	}
	return claudeStatusFromRuntime(status), nil
}

func UninstallClaude(opts ClaudeOptions) (ClaudeStatus, error) {
	status, err := uninstallRuntimeHooks(claudeRuntime, RuntimeOptions(opts))
	if err != nil {
		return ClaudeStatus{}, err
	}
	return claudeStatusFromRuntime(status), nil
}

func ClaudeHookStatus(opts ClaudeOptions) ClaudeStatus {
	return claudeStatusFromRuntime(runtimeHookStatus(claudeRuntime, RuntimeOptions(opts)))
}

func IsClaudeInstalled(opts ClaudeOptions) bool {
	return isRuntimeInstalled(claudeRuntime, RuntimeOptions(opts))
}

func claudeStatusFromRuntime(status runtimeStatus) ClaudeStatus {
	return ClaudeStatus{
		Installed:    status.Installed,
		BinaryPath:   status.BinaryPath,
		SettingsPath: status.ConfigPath,
		Message:      status.Message,
	}
}

func installClaudeSettings(path, binaryPath, logPath, configPath string) error {
	prefix := endpointCommandPrefix("claude", binaryPath, logPath, configPath)
	endpointHooks := map[string]settingsHookGroup{
		"SessionStart":       {Hooks: []settingsHookRef{{Type: "command", Command: prefix + " session-start"}}},
		"UserPromptSubmit":   {Hooks: []settingsHookRef{{Type: "command", Command: prefix + " prompt-submit", Timeout: 30}}},
		"PreToolUse":         {Matcher: "Bash|Edit|Write|MultiEdit|Read|Glob|Grep|WebFetch|WebSearch|Agent|mcp__.*", Hooks: []settingsHookRef{{Type: "command", Command: prefix + " pre-tool"}}},
		"PostToolUse":        {Matcher: "*", Hooks: []settingsHookRef{{Type: "command", Command: prefix + " post-tool"}}},
		"PostToolUseFailure": {Matcher: "*", Hooks: []settingsHookRef{{Type: "command", Command: prefix + " post-tool"}}},
		"Stop":               {Hooks: []settingsHookRef{{Type: "command", Command: prefix + " stop", Timeout: 45}}},
		"SubagentStart":      {Hooks: []settingsHookRef{{Type: "command", Command: prefix + " subagent-start"}}},
		"SubagentStop":       {Hooks: []settingsHookRef{{Type: "command", Command: prefix + " subagent-stop"}}},
		"PermissionRequest":  {Matcher: "*", Hooks: []settingsHookRef{{Type: "command", Command: prefix + " permission-request"}}},
		"SessionEnd":         {Hooks: []settingsHookRef{{Type: "command", Command: prefix + " session-end"}}},
	}
	return installSettingsEndpointHooks(path, "claude", endpointHooks)
}

func removeClaudeEndpointHooks(path string) (bool, error) {
	return removeSettingsEndpointHooks(path, "claude")
}

func isClaudeInstalledAt(path string) bool {
	return isSettingsEndpointInstalledAt(path, "claude")
}

func claudeSettingsPath(level Level) (string, error) {
	switch level {
	case "", LevelUser:
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, ".claude", "settings.json"), nil
	case LevelProject:
		cwd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		return filepath.Join(cwd, ".claude", "settings.json"), nil
	default:
		return "", fmt.Errorf("unknown hook level %q", level)
	}
}
