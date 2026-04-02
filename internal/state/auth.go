package state

import "strings"

type AuthState struct {
	LoggedIn      bool   `json:"logged_in"`
	Provider      string `json:"provider,omitempty"`
	Account       string `json:"account,omitempty"`
	ProviderReady bool   `json:"provider_ready"`
	LoginCount    int    `json:"login_count,omitempty"`
	LogoutCount   int    `json:"logout_count,omitempty"`
}

func ReadAuthState(path string) (AuthState, error) {
	var state AuthState
	if err := readJSONState(path, "auth state", &state); err != nil {
		return AuthState{}, err
	}
	return normalizeAuthState(state), nil
}

func WriteAuthState(path string, state AuthState) error {
	state = normalizeAuthState(state)
	return writeJSONState(path, "auth state", state)
}

func normalizeAuthState(state AuthState) AuthState {
	state.Provider = strings.TrimSpace(state.Provider)
	state.Account = strings.TrimSpace(state.Account)
	if state.LoginCount < 0 {
		state.LoginCount = 0
	}
	if state.LogoutCount < 0 {
		state.LogoutCount = 0
	}
	return state
}
