package ingest

import (
	"encoding/json"
	"net/http"

	beaconauth "github.com/asymptote-labs/agent-beacon/cli/beacon/internal/auth"
)

const (
	DefaultBatchEvents = 500
	DefaultBatchBytes  = 1024 * 1024
)

type Settings struct {
	Enabled          bool
	Managed          bool
	IngestURL        string
	SourceID         string
	ContentRetention string
}

type State struct {
	Enabled       bool             `json:"enabled"`
	Managed       bool             `json:"managed"`
	LoggedIn      bool             `json:"logged_in"`
	UserEmail     string           `json:"user_email,omitempty"`
	OrgName       string           `json:"org_name,omitempty"`
	SourceID      string           `json:"source_id,omitempty"`
	LastUploadAt  string           `json:"last_upload_at,omitempty"`
	LastEventAt   string           `json:"last_event_at,omitempty"`
	LastCursor    Cursor           `json:"last_cursor,omitempty"`
	AcceptedCount int              `json:"accepted_count"`
	RejectedCount int              `json:"rejected_count"`
	LastError     string           `json:"last_error,omitempty"`
	FileOffsets   map[string]int64 `json:"file_offsets,omitempty"`
	UpdatedAt     string           `json:"updated_at,omitempty"`
}

type Cursor struct {
	LogPath string `json:"log_path,omitempty"`
	Offset  int64  `json:"offset,omitempty"`
	Line    int    `json:"line,omitempty"`
	Archive string `json:"archive,omitempty"`
}

type SourceMetadata struct {
	SourceID         string
	Hostname         string
	EndpointMode     string
	LogPath          string
	ContentRetention string
	ManagedMode      bool
}

type Batch struct {
	Cursor   Cursor
	Events   []json.RawMessage
	Rejected int
}

type Source interface {
	Metadata() SourceMetadata
	Batches(state State, maxEvents int, maxBytes int) ([]Batch, error)
}

type Options struct {
	Settings   Settings
	Creds      *beaconauth.Credentials
	Store      Store
	Client     Client
	Source     Source
	HTTPClient *http.Client
}

type Result struct {
	State    State
	Uploaded bool
}

type uploadRequest struct {
	UploadID string            `json:"upload_id"`
	Source   uploadSource      `json:"source"`
	Cursor   Cursor            `json:"cursor"`
	Events   []json.RawMessage `json:"events"`
}

type uploadSource struct {
	SourceID         string `json:"source_id,omitempty"`
	Hostname         string `json:"hostname,omitempty"`
	EndpointMode     string `json:"endpoint_mode"`
	LogPath          string `json:"log_path,omitempty"`
	ContentRetention string `json:"content_retention"`
	ManagedMode      bool   `json:"managed_mode"`
}

type uploadResponse struct {
	Accepted int `json:"accepted"`
	Rejected int `json:"rejected"`
}
