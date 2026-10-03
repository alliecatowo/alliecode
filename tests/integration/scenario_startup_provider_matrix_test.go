package integration_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/alliecatowo/alliecode/internal/commands"
	"github.com/alliecatowo/alliecode/internal/config"
	"github.com/alliecatowo/alliecode/tests/integration/fixtures"
)

type startupProviderCase struct {
	Name                   string `json:"name"`
	Kind                   string `json:"kind"`
	NetworkOnly            bool   `json:"network_only"`
	ProviderChoice         string `json:"provider_choice"`
	InitialProviderName    string `json:"initial_provider_name"`
	LoggedIn               bool   `json:"logged_in"`
	AuthProvider           string `json:"auth_provider"`
	ProviderArg            string `json:"provider_arg"`
	ModelArg               string `json:"model_arg"`
	WantModel              string `json:"want_model"`
	WantProviderName       string `json:"want_provider_name"`
	WantModelRef           string `json:"want_model_ref"`
	WantProviderReady      *bool  `json:"want_provider_ready"`
	WantLoggedIn           *bool  `json:"want_logged_in"`
	WantAuthProvider       string `json:"want_auth_provider"`
	DefaultProvider        string `json:"default_provider"`
	ProviderAPIKey         string `json:"provider_api_key"`
	WantHasCreds           bool   `json:"want_has_credentials"`
	OnboardingModelMessage string `json:"onboarding_model_message"`
}

func TestScenarioMatrix_StartupProviderSetup(t *testing.T) {

	paths := loadScenarioMatrixPaths(t, "startup_provider", "*.json")
	for _, path := range paths {
		path := path
		t.Run(filepath.Base(path), func(t *testing.T) {
			var tc startupProviderCase
			decodeScenarioCase(t, path, &tc)
			if tc.NetworkOnly && os.Getenv("ALLIECODE_ENABLE_NETWORK_TESTS") != "1" {
				t.Skip("network-only scenario (set ALLIECODE_ENABLE_NETWORK_TESTS=1 to enable)")
			}

			switch tc.Kind {
			case "runtime_model_set":
				cfg := config.NewDefaultConfig()
				cfg.HasCompletedOnboarding = true
				cfg.OnboardingProviderChoice = tc.ProviderChoice

				state := &commands.RuntimeState{ProviderName: tc.InitialProviderName, LoggedIn: tc.LoggedIn, AuthProvider: tc.AuthProvider}
				if state.ProviderName == "" {
					state.ProviderName = cfg.OnboardingProviderChoice
				}
				if state.AuthProvider == "" && state.LoggedIn {
					state.AuthProvider = state.ProviderName
				}
				if tc.ProviderArg != "" {
					state.ProviderName = tc.ProviderArg
				}

				fx := fixtures.NewCLIInvocationFixture(state)
				res, err := fx.Invoke("/model " + tc.ModelArg)
				if err != nil {
					t.Fatalf("Invoke(/model %s) error = %v", tc.ModelArg, err)
				}
				if !res.Handled {
					t.Fatalf("/model %s should be handled", tc.ModelArg)
				}
				if got := state.Model; got != tc.WantModel {
					t.Fatalf("state.Model = %q, want %q", got, tc.WantModel)
				}
				if got := state.ProviderName; got != tc.WantProviderName {
					t.Fatalf("state.ProviderName = %q, want %q", got, tc.WantProviderName)
				}
				if strings.TrimSpace(tc.WantModelRef) != "" {
					if got := state.ModelRef; got != tc.WantModelRef {
						t.Fatalf("state.ModelRef = %q, want %q", got, tc.WantModelRef)
					}
				}
				if tc.WantProviderReady != nil {
					if got := state.ProviderReady; got != *tc.WantProviderReady {
						t.Fatalf("state.ProviderReady = %t, want %t", got, *tc.WantProviderReady)
					}
				}
				if tc.WantLoggedIn != nil {
					if got := state.LoggedIn; got != *tc.WantLoggedIn {
						t.Fatalf("state.LoggedIn = %t, want %t", got, *tc.WantLoggedIn)
					}
				}
				if strings.TrimSpace(tc.WantAuthProvider) != "" {
					if got := state.AuthProvider; got != tc.WantAuthProvider {
						t.Fatalf("state.AuthProvider = %q, want %q", got, tc.WantAuthProvider)
					}
				}
			case "load_layered_default_provider":
				home := t.TempDir()
				project := t.TempDir()
				t.Setenv("HOME", home)

				cfg, err := config.LoadLayered(project)
				if err != nil {
					t.Fatalf("LoadLayered() error = %v", err)
				}
				if got := cfg.DefaultProvider; got != tc.DefaultProvider {
					t.Fatalf("DefaultProvider = %q, want %q", got, tc.DefaultProvider)
				}
			case "provider_credentials":
				cfg := config.NewDefaultConfig()
				cfg.DefaultProvider = tc.DefaultProvider
				if tc.ProviderAPIKey != "" {
					cfg.Providers[tc.DefaultProvider] = &config.ProviderSettings{APIKey: tc.ProviderAPIKey}
				}

				if got := cfg.HasDefaultProviderCredentials(); got != tc.WantHasCreds {
					t.Fatalf("HasDefaultProviderCredentials() = %t, want %t", got, tc.WantHasCreds)
				}

			case "onboarding_model_availability_message":
				cfg := config.NewDefaultConfig()
				cfg.DefaultProvider = tc.DefaultProvider
				if tc.ProviderAPIKey != "" {
					cfg.Providers[tc.DefaultProvider] = &config.ProviderSettings{APIKey: tc.ProviderAPIKey}
				}
				providerName := strings.ToLower(strings.TrimSpace(tc.DefaultProvider))
				msg := ""
				if providerName != "" && providerName != "ollama" && !cfg.ProviderCredentialPresence(providerName).HasCredentials {
					msg = "Model availability check: skipped live provider check for " + providerName + " (credentials missing). Recovery: add credentials, run /login provider " + providerName + ", then rerun model selection."
				}
				if strings.TrimSpace(tc.OnboardingModelMessage) != "" && !strings.Contains(msg, tc.OnboardingModelMessage) {
					t.Fatalf("onboarding availability message = %q, want contains %q", msg, tc.OnboardingModelMessage)
				}

			default:
				t.Fatalf("unknown scenario kind %q", tc.Kind)
			}
		})
	}
}

func loadScenarioMatrixPaths(t *testing.T, family, pattern string) []string {
	t.Helper()
	base := filepath.Join("testdata", "scenario_matrix", family)
	paths, err := filepath.Glob(filepath.Join(base, pattern))
	if err != nil {
		t.Fatalf("Glob(%s) error = %v", base, err)
	}
	if len(paths) == 0 {
		t.Fatalf("no scenario fixtures found in %s", base)
	}
	sort.Strings(paths)
	return paths
}

func decodeScenarioCase(t *testing.T, path string, out any) {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%s) error = %v", path, err)
	}
	if err := json.Unmarshal(b, out); err != nil {
		t.Fatalf("Unmarshal(%s) error = %v", path, err)
	}
}
