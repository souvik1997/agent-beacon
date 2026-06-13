package ingest

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Store struct {
	Path string
}

func (s Store) Load() State {
	data, err := os.ReadFile(s.Path)
	if err != nil {
		return emptyState()
	}
	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		return emptyState()
	}
	if state.FileOffsets == nil {
		state.FileOffsets = map[string]int64{}
	}
	if state.FileIDs == nil {
		state.FileIDs = map[string]string{}
	}
	return state
}

func (s Store) Save(state State) error {
	if state.FileOffsets == nil {
		state.FileOffsets = map[string]int64{}
	}
	if state.FileIDs == nil {
		state.FileIDs = map[string]string{}
	}
	if err := os.MkdirAll(filepath.Dir(s.Path), 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.Path, data, 0600)
}

func emptyState() State {
	return State{FileOffsets: map[string]int64{}, FileIDs: map[string]string{}}
}
