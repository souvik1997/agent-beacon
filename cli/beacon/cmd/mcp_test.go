package cmd

import "testing"

func TestMCPCommandsRegistered(t *testing.T) {
	cmd, _, err := rootCmd.Find([]string{"mcp", "serve"})
	if err != nil {
		t.Fatalf("Find mcp serve returned error: %v", err)
	}
	if cmd == nil || cmd.Use != "serve" {
		t.Fatalf("mcp serve command not registered: %#v", cmd)
	}
	for _, name := range []string{"transport", "addr", "user", "system", "log-path"} {
		if cmd.Flags().Lookup(name) == nil {
			t.Fatalf("mcp serve missing --%s flag", name)
		}
	}

	doctor, _, err := rootCmd.Find([]string{"mcp", "doctor"})
	if err != nil {
		t.Fatalf("Find mcp doctor returned error: %v", err)
	}
	if doctor == nil || doctor.Use != "doctor" {
		t.Fatalf("mcp doctor command not registered: %#v", doctor)
	}
}

func TestMCPDoctorRejectsNonLoopbackHTTP(t *testing.T) {
	oldTransport := mcpOpts.transport
	oldAddr := mcpOpts.addr
	t.Cleanup(func() {
		mcpOpts.transport = oldTransport
		mcpOpts.addr = oldAddr
	})
	mcpOpts.transport = "http"
	mcpOpts.addr = "0.0.0.0:8766"
	if err := runMCPDoctor(mcpDoctorCmd, nil); err == nil {
		t.Fatal("runMCPDoctor accepted non-loopback HTTP address")
	}
}
