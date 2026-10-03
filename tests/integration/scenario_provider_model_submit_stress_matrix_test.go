package integration_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/alliecatowo/alliecode/internal/commands"
	"github.com/alliecatowo/alliecode/internal/permissions"
)

type providerModelSubmitStressCase struct {
	Name                string                    `json:"name"`
	InitialProviderName string                    `json:"initial_provider_name"`
	InitialModel        string                    `json:"initial_model"`
	InitialModelRef     string                    `json:"initial_model_ref"`
	InitialLoggedIn     bool                      `json:"initial_logged_in"`
	InitialAuthProvider string                    `json:"initial_auth_provider"`
	Steps               []string                  `json:"steps"`
	Checkpoints         []providerModelCheckpoint `json:"checkpoints"`
	WantProviderName    string                    `json:"want_provider_name"`
	WantModel           string                    `json:"want_model"`
	WantModelRef        string                    `json:"want_model_ref"`
	WantAuthProvider    string                    `json:"want_auth_provider"`
	WantProviderReady   *bool                     `json:"want_provider_ready"`
	WantLoggedIn        *bool                     `json:"want_logged_in"`
	WantStatus          []string                  `json:"want_status"`
	WantStatusline      []string                  `json:"want_statusline"`
}

type providerModelCheckpoint struct {
	AfterStepIndex    int    `json:"after_step_index"`
	WantProviderName  string `json:"want_provider_name"`
	WantModel         string `json:"want_model"`
	WantModelRef      string `json:"want_model_ref"`
	WantAuthProvider  string `json:"want_auth_provider"`
	WantProviderReady *bool  `json:"want_provider_ready"`
	WantLoggedIn      *bool  `json:"want_logged_in"`
}

func TestScenarioMatrix_ProviderModelSubmitStress(t *testing.T) {
	paths := loadScenarioMatrixPaths(t, "provider_model_submit_stress", "*.json")
	for _, path := range paths {
		path := path
		t.Run(filepath.Base(path), func(t *testing.T) {
			var tc providerModelSubmitStressCase
			decodeScenarioCase(t, path, &tc)

			state := &commands.RuntimeState{
				ProviderName:   tc.InitialProviderName,
				Model:          tc.InitialModel,
				ModelRef:       tc.InitialModelRef,
				LoggedIn:       tc.InitialLoggedIn,
				AuthProvider:   tc.InitialAuthProvider,
				PermissionMode: permissions.ModeDefault,
			}
			r := commands.DefaultRegistry()
			ctx := commands.Context{State: state}

			for i, step := range tc.Steps {
				if _, err := r.Dispatch(context.Background(), ctx, step); err != nil {
					t.Fatalf("dispatch %q failed: %v", step, err)
				}
				for _, checkpoint := range tc.Checkpoints {
					if checkpoint.AfterStepIndex != i {
						continue
					}
					assertProviderModelChainState(t, state, checkpoint.WantProviderName, checkpoint.WantModel, checkpoint.WantModelRef, checkpoint.WantAuthProvider, checkpoint.WantProviderReady, checkpoint.WantLoggedIn)
				}
			}

			assertProviderModelChainState(t, state, tc.WantProviderName, tc.WantModel, tc.WantModelRef, tc.WantAuthProvider, tc.WantProviderReady, tc.WantLoggedIn)

			statusRes, err := r.Dispatch(context.Background(), ctx, "/status")
			if err != nil {
				t.Fatalf("/status failed: %v", err)
			}
			statuslineRes, err := r.Dispatch(context.Background(), ctx, "/statusline status")
			if err != nil {
				t.Fatalf("/statusline status failed: %v", err)
			}

			for _, want := range tc.WantStatus {
				if !containsLineToken(statusRes.Message, want) {
					t.Fatalf("status missing %q: %q", want, statusRes.Message)
				}
			}
			for _, want := range tc.WantStatusline {
				if !containsLineToken(statuslineRes.Message, want) {
					t.Fatalf("statusline missing %q: %q", want, statuslineRes.Message)
				}
			}
		})
	}
}

func assertProviderModelChainState(t *testing.T, state *commands.RuntimeState, wantProviderName, wantModel, wantModelRef, wantAuthProvider string, wantProviderReady, wantLoggedIn *bool) {
	t.Helper()
	if wantProviderName != "" && state.ProviderName != wantProviderName {
		t.Fatalf("state.ProviderName = %q, want %q", state.ProviderName, wantProviderName)
	}
	if wantModel != "" && state.Model != wantModel {
		t.Fatalf("state.Model = %q, want %q", state.Model, wantModel)
	}
	if wantModelRef != "" && state.ModelRef != wantModelRef {
		t.Fatalf("state.ModelRef = %q, want %q", state.ModelRef, wantModelRef)
	}
	if wantAuthProvider != "" && state.AuthProvider != wantAuthProvider {
		t.Fatalf("state.AuthProvider = %q, want %q", state.AuthProvider, wantAuthProvider)
	}
	if wantProviderReady != nil && state.ProviderReady != *wantProviderReady {
		t.Fatalf("state.ProviderReady = %t, want %t", state.ProviderReady, *wantProviderReady)
	}
	if wantLoggedIn != nil && state.LoggedIn != *wantLoggedIn {
		t.Fatalf("state.LoggedIn = %t, want %t", state.LoggedIn, *wantLoggedIn)
	}
}

func containsLineToken(message, token string) bool {
	for _, line := range splitLines(message) {
		if line == token {
			return true
		}
	}
	return false
}

func splitLines(message string) []string {
	if message == "" {
		return nil
	}
	lines := make([]string, 0, 32)
	start := 0
	for i := 0; i < len(message); i++ {
		if message[i] != '\n' {
			continue
		}
		if i > start {
			lines = append(lines, message[start:i])
		}
		start = i + 1
	}
	if start < len(message) {
		lines = append(lines, message[start:])
	}
	return lines
}
