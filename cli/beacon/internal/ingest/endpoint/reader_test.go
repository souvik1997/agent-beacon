package endpoint

import (
	"os"
	"path/filepath"
	"testing"

	endpointconfig "github.com/asymptote-labs/agent-beacon/cli/beacon/internal/endpoint/config"
	"github.com/asymptote-labs/agent-beacon/cli/beacon/internal/ingest"
)

func TestEndpointSourceBatchesValidEventsAndMalformedRejects(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "runtime.jsonl")
	content := []byte("{\"event\":\"one\"}\nnot-json\n{\"event\":\"two\"}\n")
	if err := os.WriteFile(logPath, content, 0600); err != nil {
		t.Fatal(err)
	}

	source := NewSource(endpointconfig.Config{UserMode: true, LogPath: logPath}, logPath, true)
	batches, err := source.Batches(ingest.State{FileOffsets: map[string]int64{}}, 500, 1024)
	if err != nil {
		t.Fatal(err)
	}

	if len(batches) != 1 {
		t.Fatalf("len(batches) = %d, want 1", len(batches))
	}
	if got := len(batches[0].Events); got != 2 {
		t.Fatalf("events = %d, want 2", got)
	}
	if batches[0].Rejected != 1 {
		t.Fatalf("Rejected = %d, want 1", batches[0].Rejected)
	}
	if batches[0].Cursor.Offset != int64(len(content)) {
		t.Fatalf("Offset = %d, want %d", batches[0].Cursor.Offset, len(content))
	}
}

func TestEndpointSourceHonorsSavedOffset(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "runtime.jsonl")
	first := "{\"event\":\"one\"}\n"
	second := "{\"event\":\"two\"}\n"
	if err := os.WriteFile(logPath, []byte(first+second), 0600); err != nil {
		t.Fatal(err)
	}

	source := NewSource(endpointconfig.Config{UserMode: true, LogPath: logPath}, logPath, true)
	batches, err := source.Batches(ingest.State{FileOffsets: map[string]int64{logPath: int64(len(first))}}, 500, 1024)
	if err != nil {
		t.Fatal(err)
	}

	if len(batches) != 1 || len(batches[0].Events) != 1 {
		t.Fatalf("unexpected batches: %#v", batches)
	}
	if string(batches[0].Events[0]) != "{\"event\":\"two\"}" {
		t.Fatalf("event = %s", batches[0].Events[0])
	}
}
