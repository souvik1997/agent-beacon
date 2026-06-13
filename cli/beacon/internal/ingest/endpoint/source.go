package endpoint

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	endpointconfig "github.com/asymptote-labs/agent-beacon/cli/beacon/internal/endpoint/config"
	"github.com/asymptote-labs/agent-beacon/cli/beacon/internal/ingest"
)

const stateFileName = "upload-state.json"

type Source struct {
	cfg      endpointconfig.Config
	logPath  string
	userMode bool
}

func NewSource(cfg endpointconfig.Config, logPath string, userMode bool) Source {
	return Source{
		cfg:      cfg,
		logPath:  logPath,
		userMode: userMode,
	}
}

func Settings(cfg endpointconfig.Config, logPath string, userMode bool) ingest.Settings {
	managed := cfg.ManagedUpload
	settings := ingest.Settings{
		SourceID:         sourceID(cfg, logPath, userMode),
		ContentRetention: contentRetention(managed),
	}
	if managed == nil {
		return settings
	}
	settings.Enabled = managed.Enabled
	settings.Managed = managed.Managed
	settings.IngestURL = strings.TrimSpace(managed.IngestURL)
	return settings
}

func Store(userMode bool) ingest.Store {
	return ingest.Store{Path: filepath.Join(endpointconfig.BaseDir(userMode), "managed", stateFileName)}
}

func (s Source) Metadata() ingest.SourceMetadata {
	settings := Settings(s.cfg, s.logPath, s.userMode)
	return ingest.SourceMetadata{
		SourceID:         settings.SourceID,
		Hostname:         hostname(),
		EndpointMode:     endpointMode(s.userMode),
		LogPath:          s.logPath,
		ContentRetention: settings.ContentRetention,
		ManagedMode:      settings.Managed,
	}
}

func sourceID(cfg endpointconfig.Config, logPath string, userMode bool) string {
	if cfg.ManagedUpload != nil && strings.TrimSpace(cfg.ManagedUpload.SourceID) != "" {
		return strings.TrimSpace(cfg.ManagedUpload.SourceID)
	}
	sum := sha256.Sum256([]byte(fmt.Sprintf("%t|%s", userMode, logPath)))
	return "beacon-local-" + hex.EncodeToString(sum[:])[:32]
}

func hostname() string {
	name, err := os.Hostname()
	if err != nil {
		return ""
	}
	return name
}

func endpointMode(userMode bool) string {
	if userMode {
		return "user"
	}
	return "system"
}

func contentRetention(cfg *endpointconfig.ManagedUpload) string {
	if cfg == nil || strings.TrimSpace(cfg.ContentRetention) == "" {
		return "redacted"
	}
	return strings.TrimSpace(cfg.ContentRetention)
}
