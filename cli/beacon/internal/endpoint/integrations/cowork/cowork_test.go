package cowork

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestPrintConfigIncludesEndpointAndPrivacyNote(t *testing.T) {
	out := PrintConfig(Config{
		Endpoint:           "https://collector.example.com",
		Protocol:           "HTTP/protobuf",
		Headers:            "Authorization=Bearer token",
		ResourceAttributes: "deployment.environment=prod,service.name=claude-cowork",
	})
	if !strings.Contains(out, "https://collector.example.com") {
		t.Fatalf("missing endpoint: %s", out)
	}
	if !strings.Contains(out, AdminURL) {
		t.Fatalf("missing admin URL: %s", out)
	}
	if !strings.Contains(out, "Authorization=Bearer token") {
		t.Fatalf("missing headers: %s", out)
	}
	if !strings.Contains(out, "deployment.environment=prod,service.name=claude-cowork") {
		t.Fatalf("missing resource attributes: %s", out)
	}
	if !strings.Contains(out, "prompt text") {
		t.Fatalf("missing privacy note: %s", out)
	}
}

func TestLastCoworkEventAndSince(t *testing.T) {
	path := filepath.Join(t.TempDir(), "runtime.jsonl")
	data := strings.Join([]string{
		`{"timestamp":"2026-05-13T16:18:40Z","harness":{"name":"claude_cowork"}}`,
		`{"timestamp":"2026-05-13T16:19:40Z","harness":{"name":"cursor"}}`,
		`{"timestamp":"2026-05-13T16:20:40Z","harness":{"name":"claude_cowork"}}`,
		"",
	}, "\n")
	if err := os.WriteFile(path, []byte(data), 0644); err != nil {
		t.Fatal(err)
	}
	if !HasRecentCoworkEvent(path) {
		t.Fatal("expected cowork event to be detected")
	}
	last, ok := LastCoworkEvent(path)
	if !ok {
		t.Fatal("expected last cowork event")
	}
	if got, want := last.UTC().Format(time.RFC3339), "2026-05-13T16:20:40Z"; got != want {
		t.Fatalf("LastCoworkEvent = %s, want %s", got, want)
	}
	if !HasCoworkEventSince(path, time.Date(2026, 5, 13, 16, 20, 0, 0, time.UTC)) {
		t.Fatal("expected event after since timestamp")
	}
	if HasCoworkEventSince(path, time.Date(2026, 5, 13, 16, 21, 0, 0, time.UTC)) {
		t.Fatal("did not expect event after later since timestamp")
	}
}

func TestGetStatusIncludesLastEventObservedAt(t *testing.T) {
	path := filepath.Join(t.TempDir(), "runtime.jsonl")
	data := `{"timestamp":"2026-05-13T16:20:40Z","harness":{"name":"claude_cowork"}}` + "\n"
	if err := os.WriteFile(path, []byte(data), 0644); err != nil {
		t.Fatal(err)
	}
	status := GetStatus(path)
	if !status.LastEventObserved {
		t.Fatal("expected event to be observed")
	}
	if got, want := status.LastEventObservedAt, "2026-05-13T16:20:40Z"; got != want {
		t.Fatalf("LastEventObservedAt = %q, want %q", got, want)
	}
}
