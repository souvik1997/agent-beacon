package cmd

import "testing"

func TestCICommandsRegistered(t *testing.T) {
	for _, path := range [][]string{
		{"ci"},
		{"ci", "exec"},
		{"ci", "validate"},
	} {
		cmd, _, err := rootCmd.Find(path)
		if err != nil {
			t.Fatalf("Find %v returned error: %v", path, err)
		}
		if cmd == nil {
			t.Fatalf("command %v not registered", path)
		}
	}
	for _, name := range []string{"harness", "log-path", "min-events", "require-harness", "json"} {
		if ciExecCmd.Flags().Lookup(name) == nil {
			t.Fatalf("ci exec missing --%s", name)
		}
		if ciValidateCmd.Flags().Lookup(name) == nil {
			t.Fatalf("ci validate missing --%s", name)
		}
	}
	for _, name := range []string{"base-dir", "work-dir", "collector", "otlp-grpc-port", "otlp-http-port", "content-retention", "keep-artifacts"} {
		if ciExecCmd.Flags().Lookup(name) == nil {
			t.Fatalf("ci exec missing --%s", name)
		}
	}
}

func TestCIHarnessDefaultIsClaudeOnly(t *testing.T) {
	flag := ciExecCmd.Flags().Lookup("harness")
	if flag == nil {
		t.Fatal("ci exec missing --harness")
	}
	if flag.DefValue != "claude" {
		t.Fatalf("ci exec --harness default = %q, want claude", flag.DefValue)
	}
}
