package ingest

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	beaconauth "github.com/asymptote-labs/agent-beacon/cli/beacon/internal/auth"
)

type fakeSource struct {
	metadata SourceMetadata
	batches  []Batch
	called   bool
}

func (s *fakeSource) Metadata() SourceMetadata {
	return s.metadata
}

func (s *fakeSource) Batches(state State, maxEvents int, maxBytes int) ([]Batch, error) {
	s.called = true
	return s.batches, nil
}

func TestStatusIsReadOnly(t *testing.T) {
	statePath := filepath.Join(t.TempDir(), "upload-state.json")
	state := Status(Settings{Enabled: true, Managed: true, SourceID: "source-1"}, Store{Path: statePath}, &beaconauth.Credentials{
		Token:  "token",
		UserID: "user",
		Email:  "user@example.com",
	})

	if !state.Enabled || !state.Managed || !state.LoggedIn || state.SourceID != "source-1" {
		t.Fatalf("unexpected status: %#v", state)
	}
	if _, err := os.Stat(statePath); !os.IsNotExist(err) {
		t.Fatalf("Status wrote state file, err=%v", err)
	}
}

func TestUploadIsInertUnlessManagedEnabled(t *testing.T) {
	source := &fakeSource{}
	res := Upload(context.Background(), Options{
		Settings: Settings{Enabled: false, Managed: false},
		Store:    Store{Path: filepath.Join(t.TempDir(), "upload-state.json")},
		Creds:    &beaconauth.Credentials{Token: "token", UserID: "user"},
		Source:   source,
	})

	if source.called {
		t.Fatal("disabled upload should not read source batches")
	}
	if res.Uploaded || res.State.AcceptedCount != 0 || res.State.LastUploadAt != "" {
		t.Fatalf("disabled upload should not send events: %#v", res.State)
	}
}

func TestUploadRequiresLogin(t *testing.T) {
	source := &fakeSource{}
	res := Upload(context.Background(), Options{
		Settings: Settings{Enabled: true, Managed: true, SourceID: "source-1"},
		Store:    Store{Path: filepath.Join(t.TempDir(), "upload-state.json")},
		Source:   source,
	})

	if source.called {
		t.Fatal("upload without login should not read source batches")
	}
	if res.State.LastError == "" {
		t.Fatalf("expected login error: %#v", res.State)
	}
}

func TestUploadPostsBatchAndPersistsCursor(t *testing.T) {
	var received uploadRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer token" {
			t.Fatalf("Authorization = %q, want bearer token", got)
		}
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Fatal(err)
		}
		_, _ = w.Write([]byte(`{"accepted":1,"rejected":0}`))
	}))
	defer server.Close()

	statePath := filepath.Join(t.TempDir(), "upload-state.json")
	source := &fakeSource{
		metadata: SourceMetadata{SourceID: "source-1", EndpointMode: "user", LogPath: "runtime.jsonl", ContentRetention: "redacted", ManagedMode: true},
		batches: []Batch{{
			Cursor: Cursor{LogPath: "runtime.jsonl", Offset: 128, Line: 1},
			Events: []json.RawMessage{json.RawMessage(`{"event":"ok"}`)},
		}},
	}
	res := Upload(context.Background(), Options{
		Settings: Settings{Enabled: true, Managed: true, SourceID: "source-1", IngestURL: server.URL},
		Store:    Store{Path: statePath},
		Creds:    &beaconauth.Credentials{Token: "token", UserID: "user"},
		Source:   source,
	})

	if res.State.LastError != "" {
		t.Fatalf("LastError = %q", res.State.LastError)
	}
	if !res.Uploaded || res.State.AcceptedCount != 1 || res.State.FileOffsets["runtime.jsonl"] != 128 {
		t.Fatalf("unexpected upload state: %#v", res.State)
	}
	if received.Source.SourceID != "source-1" || !received.Source.ManagedMode || len(received.Events) != 1 {
		t.Fatalf("unexpected request: %#v", received)
	}
	if _, err := os.Stat(statePath); err != nil {
		t.Fatalf("state file not written: %v", err)
	}
}

func TestUploadDoesNotAdvanceCursorOnFailedPost(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "nope", http.StatusInternalServerError)
	}))
	defer server.Close()

	source := &fakeSource{
		metadata: SourceMetadata{SourceID: "source-1", EndpointMode: "user", LogPath: "runtime.jsonl", ContentRetention: "redacted", ManagedMode: true},
		batches: []Batch{{
			Cursor: Cursor{LogPath: "runtime.jsonl", Offset: 128, Line: 1},
			Events: []json.RawMessage{json.RawMessage(`{"event":"ok"}`)},
		}},
	}
	res := Upload(context.Background(), Options{
		Settings: Settings{Enabled: true, Managed: true, SourceID: "source-1", IngestURL: server.URL},
		Store:    Store{Path: filepath.Join(t.TempDir(), "upload-state.json")},
		Creds:    &beaconauth.Credentials{Token: "token", UserID: "user"},
		Source:   source,
	})

	if res.State.LastError == "" {
		t.Fatal("expected upload error")
	}
	if res.State.FileOffsets["runtime.jsonl"] != 0 || res.State.LastCursor.Offset != 0 {
		t.Fatalf("cursor advanced on failed post: %#v", res.State)
	}
}
