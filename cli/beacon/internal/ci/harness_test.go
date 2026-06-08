package ci

import (
	"strings"
	"testing"
)

func TestClaudeEnvIncludesPromptLogging(t *testing.T) {
	env := strings.Join(ClaudeEnv([]string{"PATH=/bin", "OTEL_LOG_USER_PROMPTS=0"}, "http://127.0.0.1:4317"), "\n")
	for _, want := range []string{
		"CLAUDE_CODE_ENABLE_TELEMETRY=1",
		"OTEL_EXPORTER_OTLP_ENDPOINT=http://127.0.0.1:4317",
		"OTEL_LOG_USER_PROMPTS=1",
	} {
		if !strings.Contains(env, want) {
			t.Fatalf("ClaudeEnv missing %q in:\n%s", want, env)
		}
	}
}
