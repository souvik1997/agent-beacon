package ingest

import beaconauth "github.com/asymptote-labs/agent-beacon/cli/beacon/internal/auth"

func Status(settings Settings, store Store, creds *beaconauth.Credentials) State {
	state := store.Load()
	applyRuntimeState(&state, settings, creds)
	return state
}

func applyRuntimeState(state *State, settings Settings, creds *beaconauth.Credentials) {
	state.Enabled = settings.Enabled
	state.Managed = settings.Managed
	state.SourceID = settings.SourceID
	state.LoggedIn = creds != nil && creds.Token != "" && !creds.IsExpired()
	if creds != nil {
		state.UserEmail = creds.Email
		state.OrgName = creds.OrgName
	}
	if state.FileOffsets == nil {
		state.FileOffsets = map[string]int64{}
	}
	if state.FileIDs == nil {
		state.FileIDs = map[string]string{}
	}
}
