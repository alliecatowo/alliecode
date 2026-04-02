package integration_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/alliecatowo/alliecode/internal/commands"
	"github.com/alliecatowo/alliecode/internal/config"
	"github.com/alliecatowo/alliecode/tests/integration/fixtures"
)

type startupProviderCase struct {
	Name                string `json:"name"`
	Kind                string `json:"kind"`
	NetworkOnly         bool   `json:"network_only"`
	ProviderChoice      string `json:"provider_choice"`
	InitialProviderName string `json:"initial_provider_name"`
	ModelArg            string `json:"model_arg"`
	WantModel           string `json:"want_model"`
	WantProviderName    string `json:"want_provider_name"`
	DefaultProvider     string `json:"default_provider"`
	ProviderAPIKey      string `json:"provider_api_key"`
	WantHasCreds        bool   `json:"want_has_credentials"`
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

				state := &commands.RuntimeState{ProviderName: tc.InitialProviderName}
				if state.ProviderName == "" {
					state.ProviderName = cfg.OnboardingProviderChoice
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
