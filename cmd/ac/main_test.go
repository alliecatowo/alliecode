package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"

	"github.com/alliecatowo/alliecode/internal/commands"
	"github.com/alliecatowo/alliecode/internal/config"
	"github.com/alliecatowo/alliecode/internal/history"
	"github.com/alliecatowo/alliecode/internal/session"
	"github.com/alliecatowo/alliecode/internal/state"
)

func TestRootCmdRunsStartupMigrations(t *testing.T) {
	originalHook := runStartupMigrationsHook
	originalTUIHook := runTUIHook
	t.Cleanup(func() {
		runStartupMigrationsHook = originalHook
		runTUIHook = originalTUIHook
	})

	called := 0
	runStartupMigrationsHook = func() error {
		called++
		return nil
	}
	runTUIHook = func(commands.RuntimeState, bool, *config.Config) error { return nil }

	cmd := newRootCmd()
	cmd.SetArgs([]string{"--print"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected execute success: %v", err)
	}
	if called != 1 {
		t.Fatalf("expected startup migrations to run once, got %d", called)
	}
}

func TestShouldRunOnboardingDecisionLogic(t *testing.T) {
	tests := []struct {
		name   string
		cfg    *config.Config
		bypass bool
		want   bool
	}{
		{
			name:   "nil config",
			cfg:    nil,
			bypass: false,
			want:   false,
		},
		{
			name: "bypass enabled",
			cfg: &config.Config{
				Providers: map[string]*config.ProviderSettings{},
			},
			bypass: true,
			want:   false,
		},
		{
			name: "already onboarded",
			cfg: &config.Config{
				HasCompletedOnboarding: true,
				Providers:              map[string]*config.ProviderSettings{},
			},
			bypass: false,
			want:   false,
		},
		{
			name: "has provider credentials",
			cfg: &config.Config{
				Providers: map[string]*config.ProviderSettings{
					"openai": {APIKey: "sk-test"},
				},
			},
			bypass: false,
			want:   false,
		},
		{
			name: "no credentials and not onboarded",
			cfg: &config.Config{
				Providers: map[string]*config.ProviderSettings{
					"ollama": {},
				},
			},
			bypass: false,
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shouldRunOnboarding(tt.cfg, tt.bypass)
			if got != tt.want {
				t.Fatalf("shouldRunOnboarding() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRootCmdRunsOnboardingAndLaunchesTUI(t *testing.T) {
	originalHook := runStartupMigrationsHook
	originalTUIHook := runTUIHook
	originalAvailabilityHook := providerModelAvailabilityHook
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("expected current working directory: %v", err)
	}
	t.Cleanup(func() {
		runStartupMigrationsHook = originalHook
		runTUIHook = originalTUIHook
		providerModelAvailabilityHook = originalAvailabilityHook
		_ = os.Chdir(originalWd)
	})

	runStartupMigrationsHook = func() error { return nil }
	providerModelAvailabilityHook = func(_ *config.Config, providerName string) ([]string, error) {
		if providerName == "ollama" {
			return []string{"llama3"}, nil
		}
		return nil, nil
	}

	home := t.TempDir()
	t.Setenv("HOME", home)
	workspace := t.TempDir()
	if err := os.Chdir(workspace); err != nil {
		t.Fatalf("expected chdir to isolated workspace: %v", err)
	}

	var gotProvider string
	var gotModel string
	runTUIHook = func(runtimeState commands.RuntimeState, _ bool, _ *config.Config) error {
		gotProvider = runtimeState.ProviderName
		gotModel = runtimeState.Model
		return nil
	}

	cmd := newRootCmd()
	out := &strings.Builder{}
	cmd.SetOut(out)
	cmd.SetIn(strings.NewReader("\n"))

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected execute success: %v", err)
	}

	if gotProvider != "ollama" {
		t.Fatalf("expected provider ollama after onboarding, got %q", gotProvider)
	}
	if gotModel != "llama3" {
		t.Fatalf("expected model llama3 after onboarding, got %q", gotModel)
	}

	cfgPath := filepath.Join(home, ".config", "alliecode", "config.yaml")
	cfg, err := config.Load(cfgPath)
	if err != nil {
		t.Fatalf("expected config load success: %v", err)
	}
	if !cfg.HasCompletedOnboarding {
		t.Fatalf("expected onboarding completion to be persisted")
	}
	if cfg.DefaultProvider != "ollama" {
		t.Fatalf("expected default provider to be ollama, got %q", cfg.DefaultProvider)
	}
	if strings.TrimSpace(cfg.GetProviderSettings("ollama").BaseURL) != "http://localhost:11434" {
		t.Fatalf("expected ollama base url to be persisted")
	}

	if !strings.Contains(out.String(), "Let's configure your startup provider + model") {
		t.Fatalf("expected onboarding output, got %q", out.String())
	}
}

func TestRootCmdSkipsOnboardingInCIEnvAndStillLaunchesTUI(t *testing.T) {
	originalHook := runStartupMigrationsHook
	originalTUIHook := runTUIHook
	t.Cleanup(func() {
		runStartupMigrationsHook = originalHook
		runTUIHook = originalTUIHook
	})

	runStartupMigrationsHook = func() error { return nil }
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("ALLIECODE_SKIP_ONBOARDING", "true")

	called := 0
	runTUIHook = func(commands.RuntimeState, bool, *config.Config) error {
		called++
		return nil
	}

	cmd := newRootCmd()
	out := &strings.Builder{}
	cmd.SetOut(out)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected execute success: %v", err)
	}
	if called != 1 {
		t.Fatalf("expected TUI to launch once, got %d", called)
	}
	if strings.Contains(out.String(), "Let's configure your startup provider + model") {
		t.Fatalf("expected onboarding to be skipped, got output %q", out.String())
	}
}

func TestOnboardingCorrectionFlowAllowsProviderAndModelChangeBeforeSave(t *testing.T) {
	originalAvailabilityHook := providerModelAvailabilityHook
	t.Cleanup(func() {
		providerModelAvailabilityHook = originalAvailabilityHook
	})
	providerModelAvailabilityHook = func(_ *config.Config, providerName string) ([]string, error) {
		switch providerName {
		case "ollama":
			return []string{"llama3"}, nil
		case "openai":
			return []string{"gpt-4o-mini"}, nil
		default:
			return nil, nil
		}
	}

	home := t.TempDir()
	t.Setenv("HOME", home)
	cfg := config.NewDefaultConfig()

	cmd := newRootCmd()
	input := strings.Join([]string{
		"2", // openai provider
		"2", // credentials prompt: choose different provider
		"1", // ollama provider
		"",  // default ollama model (llama3)
		"1", // save
	}, "\n") + "\n"
	out := &strings.Builder{}
	cmd.SetOut(out)
	cmd.SetIn(strings.NewReader(input))

	if err := runOnboarding(cmd, cfg, onboardingProgress{}); err != nil {
		t.Fatalf("expected onboarding success: %v", err)
	}
	if cfg.DefaultProvider != "ollama" {
		t.Fatalf("expected corrected provider ollama, got %q", cfg.DefaultProvider)
	}
	if cfg.DefaultModel != "llama3" {
		t.Fatalf("expected corrected model llama3, got %q", cfg.DefaultModel)
	}
	if !cfg.HasCompletedOnboarding {
		t.Fatalf("expected onboarding complete")
	}
	if !strings.Contains(out.String(), "Summary:") {
		t.Fatalf("expected summary output, got %q", out.String())
	}
}

func TestOnboardingShowsResumableAttemptMessageWhenPreviouslySeen(t *testing.T) {
	cfg := config.NewDefaultConfig()
	cmd := newRootCmd()
	out := &strings.Builder{}
	cmd.SetOut(out)
	cmd.SetIn(strings.NewReader("\n\n1\n"))

	if err := runOnboarding(cmd, cfg, onboardingProgress{SeenCount: 2, HasCompleted: false}); err != nil {
		t.Fatalf("expected onboarding success: %v", err)
	}
	if !strings.Contains(out.String(), "Resuming onboarding attempt #3") {
		t.Fatalf("expected resumable onboarding message, got %q", out.String())
	}
}

func TestOnboardingProgressPersistenceRoundTrip(t *testing.T) {
	home := t.TempDir()
	workspace := t.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(originalWd)
	})
	t.Setenv("HOME", home)
	if err := os.Chdir(workspace); err != nil {
		t.Fatalf("Chdir() error = %v", err)
	}

	if err := writeOnboardingProgress(onboardingProgress{HasCompleted: false, SeenCount: 4}); err != nil {
		t.Fatalf("writeOnboardingProgress() error = %v", err)
	}
	progress, err := readOnboardingProgress()
	if err != nil {
		t.Fatalf("readOnboardingProgress() error = %v", err)
	}
	if progress.SeenCount != 4 || progress.HasCompleted {
		t.Fatalf("unexpected progress: %+v", progress)
	}

	if err := writeOnboardingProgress(onboardingProgress{HasCompleted: true, SeenCount: 5}); err != nil {
		t.Fatalf("writeOnboardingProgress() second write error = %v", err)
	}
	progress, err = readOnboardingProgress()
	if err != nil {
		t.Fatalf("readOnboardingProgress() second read error = %v", err)
	}
	if progress.SeenCount != 5 || !progress.HasCompleted {
		t.Fatalf("unexpected progress after completion: %+v", progress)
	}
}

func TestRootCmdGuidedRepairRepairsInvalidRuntimeSelection(t *testing.T) {
	originalHook := runStartupMigrationsHook
	originalTUIHook := runTUIHook
	originalAvailabilityHook := providerModelAvailabilityHook
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("expected current working directory: %v", err)
	}
	t.Cleanup(func() {
		runStartupMigrationsHook = originalHook
		runTUIHook = originalTUIHook
		providerModelAvailabilityHook = originalAvailabilityHook
		_ = os.Chdir(originalWd)
	})
	runStartupMigrationsHook = func() error { return nil }
	providerModelAvailabilityHook = func(_ *config.Config, providerName string) ([]string, error) {
		if providerName == "ollama" {
			return []string{"llama3"}, nil
		}
		return []string{"gpt-4o-mini"}, nil
	}

	home := t.TempDir()
	t.Setenv("HOME", home)
	workspace := t.TempDir()
	if err := os.Chdir(workspace); err != nil {
		t.Fatalf("expected chdir to isolated workspace: %v", err)
	}

	if err := os.MkdirAll(filepath.Join(workspace, ".alliecode"), 0o755); err != nil {
		t.Fatalf("mkdir project config dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(workspace, ".alliecode", "config.yaml"), []byte("default_provider: broken\ndefault_model: broken-model\nhas_completed_onboarding: true\n"), 0o644); err != nil {
		t.Fatalf("write invalid project config: %v", err)
	}

	var gotProvider string
	var gotModel string
	runTUIHook = func(runtimeState commands.RuntimeState, _ bool, _ *config.Config) error {
		gotProvider = runtimeState.ProviderName
		gotModel = runtimeState.Model
		return nil
	}

	cmd := newRootCmd()
	out := &strings.Builder{}
	cmd.SetOut(out)
	cmd.SetIn(strings.NewReader("\n3\n\n\n"))

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected execute success with guided repair: %v", err)
	}
	if gotProvider != "ollama" || gotModel != "llama3" {
		t.Fatalf("expected repaired startup selection ollama/llama3, got %s/%s", gotProvider, gotModel)
	}
	if !strings.Contains(out.String(), "needs repair") {
		t.Fatalf("expected guided repair output, got %q", out.String())
	}
}

func TestEvaluateStartupRepairFlagsMissingEnvAndUnavailableModel(t *testing.T) {
	originalAvailabilityHook := providerModelAvailabilityHook
	t.Cleanup(func() {
		providerModelAvailabilityHook = originalAvailabilityHook
	})
	providerModelAvailabilityHook = func(_ *config.Config, providerName string) ([]string, error) {
		if providerName == "openai" {
			return []string{"gpt-4o"}, nil
		}
		return nil, nil
	}

	cfg := config.NewDefaultConfig()
	cfg.DefaultProvider = "openai"
	cfg.DefaultModel = "gpt-4o-mini"
	t.Setenv("OPENAI_API_KEY", "env-openai")

	repair := evaluateStartupRepair(cfg, cfg.DefaultProvider, cfg.DefaultModel)
	if !repair.Required {
		t.Fatalf("expected repair to be required")
	}
	message := strings.Join(repair.Issues, "\n")
	if !strings.Contains(message, "model \"gpt-4o-mini\" is not currently available") {
		t.Fatalf("expected unavailable model issue, got %q", message)
	}
}

func TestEvaluateStartupRepairModelAvailabilityErrorIsGuidanceOnly(t *testing.T) {
	originalAvailabilityHook := providerModelAvailabilityHook
	t.Cleanup(func() {
		providerModelAvailabilityHook = originalAvailabilityHook
	})
	providerModelAvailabilityHook = func(_ *config.Config, providerName string) ([]string, error) {
		if providerName == "openai" {
			return nil, fmt.Errorf("provider timeout")
		}
		return nil, nil
	}

	cfg := config.NewDefaultConfig()
	cfg.DefaultProvider = "openai"
	cfg.DefaultModel = "gpt-4o-mini"
	cfg.Providers["openai"] = &config.ProviderSettings{APIKey: "cfg-key"}

	repair := evaluateStartupRepair(cfg, cfg.DefaultProvider, cfg.DefaultModel)
	if repair.Required {
		t.Fatalf("expected availability lookup error to avoid hard repair, got %+v", repair)
	}
	if strings.TrimSpace(repair.Check.ModelCheckError) == "" {
		t.Fatalf("expected model check error details to be recorded")
	}
}

func TestStartupCorrectiveLoopsCoverProviderModelPermissionsSettings(t *testing.T) {
	check := startupSelectionValidation{Provider: "openai", Model: "", CredentialPresence: config.ProviderCredentialPresence{HasCredentials: false}}
	loops := startupCorrectiveLoops(check)
	if len(loops) != 8 {
		t.Fatalf("expected 8 loops, got %d", len(loops))
	}
	if loops[0].Area != "provider" || loops[1].Area != "model" || loops[2].Area != "permissions" || loops[3].Area != "settings" {
		t.Fatalf("unexpected loop ordering: %#v", loops)
	}
}

func TestRunStartupRepairRendersCorrectiveLoops(t *testing.T) {
	cmd := &cobra.Command{}
	out := &strings.Builder{}
	cmd.SetOut(out)
	cmd.SetIn(strings.NewReader("n\n"))
	err := runStartupRepair(cmd, nil, startupRepair{Required: true, Issues: []string{"provider missing"}, Check: startupSelectionValidation{}})
	if err == nil {
		t.Fatalf("expected startup repair decline to return error")
	}
	if !strings.Contains(out.String(), "loop.1") {
		t.Fatalf("expected startup repair output to include loop guidance, got %q", out.String())
	}
}

func TestRunStartupRepairIncludesQuickFixGuidance(t *testing.T) {
	cmd := newRootCmd()
	buf := &strings.Builder{}
	cmd.SetOut(buf)
	cmd.SetIn(strings.NewReader("n\n"))

	err := runStartupRepair(cmd, config.NewDefaultConfig(), startupRepair{
		Required: true,
		Issues:   []string{"provider has no credentials"},
		Check: startupSelectionValidation{
			Provider: "openai",
			Model:    "gpt-4o-mini",
			CredentialPresence: config.ProviderCredentialPresence{
				Provider: "openai",
				EnvVars:  []string{"OPENAI_API_KEY"},
			},
		},
	})
	if err == nil {
		t.Fatalf("expected repair decline error")
	}
	out := buf.String()
	if !strings.Contains(out, "Quick fix:") {
		t.Fatalf("expected quick fix output, got %q", out)
	}
	if !strings.Contains(out, "/provider status") {
		t.Fatalf("expected command guidance, got %q", out)
	}
}

func TestStartupWorkflowGuidanceLinesIncludeWave5Families(t *testing.T) {
	lines := startupWorkflowGuidanceLines()
	if len(lines) < 6 {
		t.Fatalf("expected workflow guidance lines, got %#v", lines)
	}
	want := map[string]bool{
		"/doctor fix":          false,
		"/config doctor":       false,
		"/permissions summary": false,
		"/history status":      false,
		"/session diagnostics": false,
		"/mcp diagnostics":     false,
	}
	for _, line := range lines {
		if _, ok := want[line]; ok {
			want[line] = true
		}
	}
	for line, seen := range want {
		if !seen {
			t.Fatalf("missing guidance line %q in %#v", line, lines)
		}
	}
}

func TestRenderConfigLoadRepairGuidanceIncludesWorkflowCommands(t *testing.T) {
	cmd := &cobra.Command{}
	b := &strings.Builder{}
	cmd.SetOut(b)
	err := renderConfigLoadRepairGuidance(cmd, "/tmp/test-config.yaml", fmt.Errorf("invalid config"))
	if err == nil {
		t.Fatalf("expected wrapped guidance error")
	}
	out := b.String()
	if !strings.Contains(out, "cmd.1") || !strings.Contains(out, "/mcp diagnostics") {
		t.Fatalf("expected workflow commands in guidance, got %q", out)
	}
}

func TestRunStartupRepairRendersWorkflowCommands(t *testing.T) {
	cmd := &cobra.Command{}
	b := &strings.Builder{}
	cmd.SetOut(b)
	cmd.SetIn(strings.NewReader("n\n"))
	err := runStartupRepair(cmd, nil, startupRepair{Required: true, Issues: []string{"broken"}, Check: startupSelectionValidation{}})
	if err == nil {
		t.Fatalf("expected decline error")
	}
	out := b.String()
	if !strings.Contains(out, "cmd.1") || !strings.Contains(out, "/history status") {
		t.Fatalf("expected startup workflow commands, got %q", out)
	}
}

func TestOnboardingCredentialPromptWritesProviderKey(t *testing.T) {
	cfg := config.NewDefaultConfig()
	reader := bufio.NewReader(strings.NewReader("sk-openai\n"))
	b := &strings.Builder{}
	err := promptAndStoreProviderCredential(b, reader, cfg, "openai")
	if err != nil {
		t.Fatalf("promptAndStoreProviderCredential() error = %v", err)
	}
	if got := strings.TrimSpace(cfg.GetProviderSettings("openai").APIKey); got != "sk-openai" {
		t.Fatalf("expected stored openai API key, got %q", got)
	}
}

func TestOnboardingModelChoicesWithAvailabilityFallbackMessage(t *testing.T) {
	originalAvailabilityHook := providerModelAvailabilityHook
	t.Cleanup(func() {
		providerModelAvailabilityHook = originalAvailabilityHook
	})
	providerModelAvailabilityHook = func(_ *config.Config, _ string) ([]string, error) {
		return nil, fmt.Errorf("network unreachable")
	}

	cfg := config.NewDefaultConfig()
	cfg.Providers["openai"] = &config.ProviderSettings{APIKey: "sk-openai"}
	choices, message := onboardingModelChoicesWithAvailability(cfg, "openai")
	if len(choices) == 0 {
		t.Fatalf("expected static choices fallback")
	}
	if !strings.Contains(message, "unable to verify live models") {
		t.Fatalf("expected fallback message, got %q", message)
	}
}

func TestPromptOnboardingValidationActionDefaultModel(t *testing.T) {
	action, err := promptOnboardingValidationAction(&strings.Builder{}, bufio.NewReader(strings.NewReader("\n")), "openai")
	if err != nil {
		t.Fatalf("promptOnboardingValidationAction() error = %v", err)
	}
	if action != "model" {
		t.Fatalf("expected default action model, got %q", action)
	}
}

func TestRootCmdSurfacesStartupMigrationFailure(t *testing.T) {
	originalHook := runStartupMigrationsHook
	t.Cleanup(func() {
		runStartupMigrationsHook = originalHook
	})

	runStartupMigrationsHook = func() error {
		return fmt.Errorf("boom")
	}

	cmd := newRootCmd()
	cmd.SetArgs([]string{"--print"})

	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected startup migration error")
	}
	if !strings.Contains(err.Error(), "startup migrations failed") {
		t.Fatalf("expected clear startup migration failure, got %q", err.Error())
	}
	if !strings.Contains(err.Error(), "boom") {
		t.Fatalf("expected wrapped migration error details, got %q", err.Error())
	}
}

func TestValidateOutputStyle(t *testing.T) {
	if err := validateOutputStyle("human"); err != nil {
		t.Fatalf("expected human style valid: %v", err)
	}
	if err := validateOutputStyle("compact"); err != nil {
		t.Fatalf("expected compact style valid: %v", err)
	}
	if err := validateOutputStyle("xml"); err == nil {
		t.Fatalf("expected invalid style error")
	}
}

func TestValidateOutputFormat(t *testing.T) {
	if err := validateOutputFormat("text"); err != nil {
		t.Fatalf("expected text format valid: %v", err)
	}
	if err := validateOutputFormat("json"); err != nil {
		t.Fatalf("expected json format valid: %v", err)
	}
	if err := validateOutputFormat("yaml"); err == nil {
		t.Fatalf("expected invalid format error")
	}
}

func TestValidateTransportMode(t *testing.T) {
	if err := validateTransportMode("local"); err != nil {
		t.Fatalf("expected local transport valid: %v", err)
	}
	if err := validateTransportMode("remote"); err != nil {
		t.Fatalf("expected remote transport valid: %v", err)
	}
	if err := validateTransportMode("grpc"); err == nil {
		t.Fatalf("expected invalid transport error")
	}
}

func TestLoadCLIConfigUsesLayeredDefaults(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd failed: %v", err)
	}
	t.Setenv("HOME", t.TempDir())

	tmp := t.TempDir()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir failed: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(wd)
	})

	if err := os.MkdirAll(filepath.Join(tmp, ".alliecode"), 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	projectConfig := "default_provider: ollama\ndefault_model: llama3\n"
	if err := os.WriteFile(filepath.Join(tmp, ".alliecode", "config.yaml"), []byte(projectConfig), 0o644); err != nil {
		t.Fatalf("write config failed: %v", err)
	}

	cfg, err := loadCLIConfig("")
	if err != nil {
		t.Fatalf("loadCLIConfig failed: %v", err)
	}
	if cfg.DefaultProvider != "ollama" {
		t.Fatalf("expected provider from layered config, got %q", cfg.DefaultProvider)
	}
	if cfg.DefaultModel != "llama3" {
		t.Fatalf("expected model from layered config, got %q", cfg.DefaultModel)
	}
}

func TestResolveRuntimeSelectionPrefersExplicitModelProviderTuple(t *testing.T) {
	cfg := config.NewDefaultConfig()
	cfg.DefaultProvider = "openai"
	cfg.DefaultModel = "gpt-4o-mini"

	provider, model := resolveRuntimeSelection(cfg, "openai", "anthropic/claude-opus-4-20250514")
	if provider != "anthropic" {
		t.Fatalf("provider = %q, want anthropic", provider)
	}
	if model != "claude-opus-4-20250514" {
		t.Fatalf("model = %q, want claude-opus-4-20250514", model)
	}
}

func TestResolveRuntimeSelectionKeepsProviderWhenModelIsAlias(t *testing.T) {
	cfg := config.NewDefaultConfig()
	cfg.DefaultProvider = "anthropic"
	cfg.DefaultModel = "claude-sonnet-4-20250514"

	provider, model := resolveRuntimeSelection(cfg, "anthropic", "sonnet")
	if provider != "anthropic" {
		t.Fatalf("provider = %q, want anthropic", provider)
	}
	if model != "claude-sonnet-4-20250514" {
		t.Fatalf("model = %q, want canonical sonnet", model)
	}
}

func TestParseExplicitProviderModel(t *testing.T) {
	tests := []struct {
		name string
		in   string
		ok   bool
		p    string
		m    string
	}{
		{name: "valid tuple", in: "openai/gpt-4o-mini", ok: true, p: "openai", m: "gpt-4o-mini"},
		{name: "whitespace tuple", in: " anthropic / claude-opus-4-20250514 ", ok: true, p: "anthropic", m: "claude-opus-4-20250514"},
		{name: "unknown provider", in: "unknown/model", ok: false},
		{name: "no slash", in: "gpt-4o-mini", ok: false},
		{name: "missing model", in: "openai/", ok: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, m, ok := parseExplicitProviderModel(tt.in)
			if ok != tt.ok {
				t.Fatalf("ok = %t, want %t", ok, tt.ok)
			}
			if ok {
				if p != tt.p || m != tt.m {
					t.Fatalf("got %s/%s, want %s/%s", p, m, tt.p, tt.m)
				}
			}
		})
	}
}

func TestResolveRuntimeSelectionIgnoresInvalidModelTupleProvider(t *testing.T) {
	cfg := config.NewDefaultConfig()
	cfg.DefaultProvider = "openai"
	cfg.DefaultModel = "gpt-4o-mini"

	provider, model := resolveRuntimeSelection(cfg, "openai", "unknown/gpt-4o-mini")
	if provider != "openai" {
		t.Fatalf("provider = %q, want openai", provider)
	}
	if model != "gpt-4o-mini" {
		t.Fatalf("model = %q, want gpt-4o-mini", model)
	}
}

func TestRootCmdPrintUsesModelTupleProviderAtStartup(t *testing.T) {
	originalHook := runStartupMigrationsHook
	t.Cleanup(func() {
		runStartupMigrationsHook = originalHook
	})
	runStartupMigrationsHook = func() error { return nil }

	cmd := newRootCmd()
	b := &strings.Builder{}
	cmd.SetOut(b)
	cmd.SetArgs([]string{"--print", "--provider", "openai", "--model", "anthropic/claude-opus-4-20250514"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	out := b.String()
	if !strings.Contains(out, "provider=anthropic") {
		t.Fatalf("expected provider from model tuple in output, got %q", out)
	}
	if !strings.Contains(out, "model=claude-opus-4-20250514") {
		t.Fatalf("expected model from model tuple in output, got %q", out)
	}
}

func TestRenderStartupInfoTextDeterministic(t *testing.T) {
	cmd := newRootCmd()
	b := &strings.Builder{}
	cmd.SetOut(b)

	err := renderStartupInfo(cmd, startupInfo{
		ConfigPath:   "/tmp/cfg.yaml",
		Provider:     "openai",
		Model:        "gpt-4o-mini",
		Debug:        false,
		PrintMode:    true,
		OutputStyle:  "human",
		OutputFormat: "text",
		Transport:    "local",
	})
	if err != nil {
		t.Fatalf("renderStartupInfo failed: %v", err)
	}
	output := b.String()
	if !strings.Contains(output, "provider=openai") {
		t.Fatalf("expected provider field in output: %q", output)
	}
	if !strings.Contains(output, "output_format=text") {
		t.Fatalf("expected output format field in output: %q", output)
	}
}

func TestRootCmdDispatchesSlashCommandInPrintPath(t *testing.T) {
	originalHook := runStartupMigrationsHook
	t.Cleanup(func() {
		runStartupMigrationsHook = originalHook
	})
	runStartupMigrationsHook = func() error { return nil }

	cmd := newRootCmd()
	b := &strings.Builder{}
	cmd.SetOut(b)
	cmd.SetArgs([]string{"--slash-command", "/compact status", "--print"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected execute success: %v", err)
	}
	if strings.TrimSpace(b.String()) != "COMPACT_STATUS\nmode=auto\nrequested=false\ncount=0\nlast_target=-" {
		t.Fatalf("unexpected slash output: %q", b.String())
	}
}

func TestRootCmdDispatchesSlashCommandJSONOutput(t *testing.T) {
	originalHook := runStartupMigrationsHook
	t.Cleanup(func() {
		runStartupMigrationsHook = originalHook
	})
	runStartupMigrationsHook = func() error { return nil }

	cmd := newRootCmd()
	b := &strings.Builder{}
	cmd.SetOut(b)
	cmd.SetArgs([]string{"--slash-command", "/resume abc-123", "--output-format", "json", "--print"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected execute success: %v", err)
	}
	var payload struct {
		Handled bool   `json:"handled"`
		Message string `json:"message"`
		State   struct {
			ResumeRequested bool   `json:"resume_requested"`
			Provider        string `json:"provider"`
		} `json:"state"`
	}
	if err := json.Unmarshal([]byte(b.String()), &payload); err != nil {
		t.Fatalf("expected json output, got %v", err)
	}
	if !payload.Handled || !payload.State.ResumeRequested {
		t.Fatalf("unexpected payload: %s", b.String())
	}
	if !strings.Contains(payload.Message, "RESUME_REQUEST") {
		t.Fatalf("expected resume message, got %q", payload.Message)
	}
}

func TestRootCmdDispatchesSlashCommandWithoutLeadingSlash(t *testing.T) {
	originalHook := runStartupMigrationsHook
	t.Cleanup(func() {
		runStartupMigrationsHook = originalHook
	})
	runStartupMigrationsHook = func() error { return nil }

	cmd := newRootCmd()
	b := &strings.Builder{}
	cmd.SetOut(b)
	cmd.SetArgs([]string{"--slash-command", "compact status", "--print"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected execute success: %v", err)
	}
	if strings.TrimSpace(b.String()) != "COMPACT_STATUS\nmode=auto\nrequested=false\ncount=0\nlast_target=-" {
		t.Fatalf("unexpected slash output: %q", b.String())
	}
}

func TestRunSlashCommandRejectsEmptyInput(t *testing.T) {
	cmd := newRootCmd()
	b := &strings.Builder{}
	cmd.SetOut(b)

	err := runSlashCommand(cmd.Context(), cmd, "   ", commands.RuntimeState{})
	if err == nil {
		t.Fatalf("expected empty slash command error")
	}
	if err.Error() != "slash command is empty" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestStartupRuntimeStateInitializesPersistenceAndSessionStores(t *testing.T) {
	home := t.TempDir()
	workspace := t.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}
	originalSessionIDHook := sessionIDGeneratorHook
	t.Cleanup(func() {
		_ = os.Chdir(originalWd)
		sessionIDGeneratorHook = originalSessionIDHook
	})

	if err := os.Chdir(workspace); err != nil {
		t.Fatalf("Chdir() error = %v", err)
	}
	t.Setenv("HOME", home)
	sessionIDGeneratorHook = func() string { return "startup-session-fixed" }

	runtimeState, err := startupRuntimeState("ollama", "llama3", false, "human", "text", "local")
	if err != nil {
		t.Fatalf("startupRuntimeState() error = %v", err)
	}

	if runtimeState.SessionID != "startup-session-fixed" {
		t.Fatalf("SessionID = %q, want %q", runtimeState.SessionID, "startup-session-fixed")
	}
	if runtimeState.SessionPath != session.SessionPath(home, "startup-session-fixed") {
		t.Fatalf("SessionPath = %q, want %q", runtimeState.SessionPath, session.SessionPath(home, "startup-session-fixed"))
	}

	paths, err := state.ResolvePaths(home, workspace)
	if err != nil {
		t.Fatalf("ResolvePaths() error = %v", err)
	}
	for _, p := range []string{paths.HomeDir, paths.ProjectDir, paths.SessionsDir, paths.GlobalHistoryFile, paths.ProjectsStateFile, paths.SettingsCacheFile, paths.AuthStateFile, paths.ProjectOnboardingStateFile} {
		if _, err := os.Stat(p); err != nil {
			t.Fatalf("expected %q to exist: %v", p, err)
		}
	}

	historyStore, err := history.NewJSONLStore(home)
	if err != nil {
		t.Fatalf("NewJSONLStore() error = %v", err)
	}
	historyEvents, err := historyStore.SessionEvents("startup-session-fixed")
	if err != nil {
		t.Fatalf("SessionEvents() error = %v", err)
	}
	if len(historyEvents) != 1 || historyEvents[0].Type != history.EventSessionStart {
		t.Fatalf("startup history events = %+v, want one session_start event", historyEvents)
	}
	if historyEvents[0].Project != workspace {
		t.Fatalf("startup history project = %q, want %q", historyEvents[0].Project, workspace)
	}

	transcript, err := session.LoadByID(home, "startup-session-fixed")
	if err != nil {
		t.Fatalf("LoadByID() error = %v", err)
	}
	if transcript.SessionID != "startup-session-fixed" {
		t.Fatalf("transcript.SessionID = %q, want %q", transcript.SessionID, "startup-session-fixed")
	}
}

func TestStartupRuntimeStateHydratesSelectedProviderAuthAndSessionMetadata(t *testing.T) {
	home := t.TempDir()
	workspace := t.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}
	originalSessionIDHook := sessionIDGeneratorHook
	t.Cleanup(func() {
		_ = os.Chdir(originalWd)
		sessionIDGeneratorHook = originalSessionIDHook
	})

	if err := os.Chdir(workspace); err != nil {
		t.Fatalf("Chdir() error = %v", err)
	}
	t.Setenv("HOME", home)
	sessionIDGeneratorHook = func() string { return "hydrated-session" }

	paths, err := state.ResolvePaths(home, workspace)
	if err != nil {
		t.Fatalf("ResolvePaths() error = %v", err)
	}
	if err := state.EnsureState(paths); err != nil {
		t.Fatalf("EnsureState() error = %v", err)
	}
	if err := state.WriteAuthState(paths.AuthStateFile, state.AuthState{
		LoggedIn:      true,
		Provider:      "anthropic",
		Account:       "dev@acme",
		ProviderReady: true,
	}); err != nil {
		t.Fatalf("WriteAuthState() error = %v", err)
	}

	runtimeState, err := startupRuntimeState("openai", "gpt-4o-mini", false, "human", "text", "local")
	if err != nil {
		t.Fatalf("startupRuntimeState() error = %v", err)
	}
	if runtimeState.ProviderName != "openai" {
		t.Fatalf("ProviderName = %q, want openai", runtimeState.ProviderName)
	}
	if runtimeState.Model != "gpt-4o-mini" {
		t.Fatalf("Model = %q, want gpt-4o-mini", runtimeState.Model)
	}
	if runtimeState.ModelRef != "openai/gpt-4o-mini" {
		t.Fatalf("ModelRef = %q, want openai/gpt-4o-mini", runtimeState.ModelRef)
	}
	if runtimeState.LoggedIn {
		t.Fatalf("expected startup auth hydration to clear mismatched provider login")
	}
	if runtimeState.AuthAccount != "" {
		t.Fatalf("AuthAccount = %q, want empty after provider mismatch", runtimeState.AuthAccount)
	}
	if runtimeState.AuthProvider != "" {
		t.Fatalf("AuthProvider = %q, want empty after provider mismatch", runtimeState.AuthProvider)
	}
	if runtimeState.ProviderReady {
		t.Fatalf("expected ProviderReady false after provider mismatch")
	}
	if runtimeState.SessionID != "hydrated-session" || runtimeState.SessionPath == "" {
		t.Fatalf("unexpected session hydration: %+v", runtimeState)
	}
}

func TestStartupRuntimeStateHydratesAnthropicSelectionAndAuthCoherently(t *testing.T) {
	home := t.TempDir()
	workspace := t.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}
	originalSessionIDHook := sessionIDGeneratorHook
	t.Cleanup(func() {
		_ = os.Chdir(originalWd)
		sessionIDGeneratorHook = originalSessionIDHook
	})

	if err := os.Chdir(workspace); err != nil {
		t.Fatalf("Chdir() error = %v", err)
	}
	t.Setenv("HOME", home)
	sessionIDGeneratorHook = func() string { return "hydrated-anthropic-session" }

	paths, err := state.ResolvePaths(home, workspace)
	if err != nil {
		t.Fatalf("ResolvePaths() error = %v", err)
	}
	if err := state.EnsureState(paths); err != nil {
		t.Fatalf("EnsureState() error = %v", err)
	}
	if err := state.WriteAuthState(paths.AuthStateFile, state.AuthState{
		LoggedIn:      true,
		Provider:      "anthropic",
		Account:       "dev@acme",
		ProviderReady: true,
	}); err != nil {
		t.Fatalf("WriteAuthState() error = %v", err)
	}

	runtimeState, err := startupRuntimeState("anthropic", "claude-opus-4-20250514", false, "human", "text", "local")
	if err != nil {
		t.Fatalf("startupRuntimeState() error = %v", err)
	}
	if runtimeState.ProviderName != "anthropic" {
		t.Fatalf("ProviderName = %q, want anthropic", runtimeState.ProviderName)
	}
	if runtimeState.Model != "claude-opus-4-20250514" {
		t.Fatalf("Model = %q, want claude-opus-4-20250514", runtimeState.Model)
	}
	if runtimeState.ModelRef != "anthropic/claude-opus-4-20250514" {
		t.Fatalf("ModelRef = %q, want anthropic/claude-opus-4-20250514", runtimeState.ModelRef)
	}
	if !runtimeState.LoggedIn || runtimeState.AuthProvider != "anthropic" || runtimeState.AuthAccount != "dev@acme" || !runtimeState.ProviderReady {
		t.Fatalf("unexpected auth hydration for matching provider: %+v", runtimeState)
	}
	if runtimeState.Runtime.ProviderName != "anthropic" || runtimeState.Runtime.Model != "claude-opus-4-20250514" || runtimeState.Runtime.ModelRef != "anthropic/claude-opus-4-20250514" {
		t.Fatalf("expected runtime snapshot hydrated from startup selection, got %+v", runtimeState.Runtime)
	}
	if !runtimeState.Runtime.LoggedIn || !runtimeState.Runtime.ProviderReady {
		t.Fatalf("expected runtime auth/readiness hydration, got %+v", runtimeState.Runtime)
	}
}

func TestRootCmdHistoryListUsesStartupLoadedRecentSummaries(t *testing.T) {
	originalHook := runStartupMigrationsHook
	originalSessionIDHook := sessionIDGeneratorHook
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}
	t.Cleanup(func() {
		runStartupMigrationsHook = originalHook
		sessionIDGeneratorHook = originalSessionIDHook
		_ = os.Chdir(originalWd)
	})

	runStartupMigrationsHook = func() error { return nil }
	sessionIDGeneratorHook = func() string { return "new-startup-session" }

	home := t.TempDir()
	workspace := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("ALLIECODE_SKIP_ONBOARDING", "true")
	if err := os.Chdir(workspace); err != nil {
		t.Fatalf("Chdir() error = %v", err)
	}

	store, err := history.NewJSONLStore(home)
	if err != nil {
		t.Fatalf("NewJSONLStore() error = %v", err)
	}
	if err := store.AppendSessionStartForProject("existing-session", workspace); err != nil {
		t.Fatalf("AppendSessionStartForProject() error = %v", err)
	}

	cmd := newRootCmd()
	b := &strings.Builder{}
	cmd.SetOut(b)
	cmd.SetArgs([]string{"--slash-command", "/history list", "--print"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := b.String()
	if !strings.Contains(output, "entry.1.id=existing-session") {
		t.Fatalf("expected existing-session in history list, got %q", output)
	}
	if !strings.Contains(output, "entry.2.id=new-startup-session") {
		t.Fatalf("expected startup-created session in history list, got %q", output)
	}
	if !strings.Contains(output, "entry.1.path="+history.SessionPath(home, "existing-session")) {
		t.Fatalf("expected existing session path in history output, got %q", output)
	}
}

func TestStartupRuntimeStateHydratesRecentSummaryMetadata(t *testing.T) {
	home := t.TempDir()
	workspace := t.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}
	originalSessionIDHook := sessionIDGeneratorHook
	t.Cleanup(func() {
		_ = os.Chdir(originalWd)
		sessionIDGeneratorHook = originalSessionIDHook
	})

	if err := os.Chdir(workspace); err != nil {
		t.Fatalf("Chdir() error = %v", err)
	}
	t.Setenv("HOME", home)

	store, err := history.NewJSONLStore(home)
	if err != nil {
		t.Fatalf("NewJSONLStore() error = %v", err)
	}
	if err := store.AppendSessionStartForProject("existing-session", workspace); err != nil {
		t.Fatalf("AppendSessionStartForProject() error = %v", err)
	}

	paths, err := state.ResolvePaths(home, workspace)
	if err != nil {
		t.Fatalf("ResolvePaths() error = %v", err)
	}
	if err := state.EnsureState(paths); err != nil {
		t.Fatalf("EnsureState() error = %v", err)
	}
	if err := state.WriteSessionMetadataRegistry(paths.SessionMetadataFile, state.SessionMetadataRegistry{
		Sessions: map[string]state.SessionMetadata{
			"existing-session": {
				SessionID:   "existing-session",
				SessionPath: history.SessionPath(home, "existing-session"),
				ProjectPath: workspace,
				Title:       "Hydrated Session Title",
				Summary:     "restored from metadata",
			},
		},
	}); err != nil {
		t.Fatalf("WriteSessionMetadataRegistry() error = %v", err)
	}

	sessionIDGeneratorHook = func() string { return "new-startup-session-for-hydration" }
	runtimeState, err := startupRuntimeState("ollama", "llama3", false, "human", "text", "local")
	if err != nil {
		t.Fatalf("startupRuntimeState() error = %v", err)
	}

	found := false
	for _, entry := range runtimeState.HistoryEntries {
		if entry.ID != "existing-session" {
			continue
		}
		found = true
		if entry.Title != "Hydrated Session Title" {
			t.Fatalf("entry.Title = %q, want %q", entry.Title, "Hydrated Session Title")
		}
		if entry.Summary != "restored from metadata" {
			t.Fatalf("entry.Summary = %q, want %q", entry.Summary, "restored from metadata")
		}
	}
	if !found {
		t.Fatalf("expected existing-session entry in startup history")
	}
}

func TestRenamePersistsTitleVisibleOnNextStartup(t *testing.T) {
	originalHook := runStartupMigrationsHook
	originalSessionIDHook := sessionIDGeneratorHook
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}
	t.Cleanup(func() {
		runStartupMigrationsHook = originalHook
		sessionIDGeneratorHook = originalSessionIDHook
		_ = os.Chdir(originalWd)
	})

	runStartupMigrationsHook = func() error { return nil }
	home := t.TempDir()
	workspace := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("ALLIECODE_SKIP_ONBOARDING", "true")
	if err := os.Chdir(workspace); err != nil {
		t.Fatalf("Chdir() error = %v", err)
	}

	sessionIDGeneratorHook = func() string { return "rename-session" }
	cmd := newRootCmd()
	cmd.SetOut(&strings.Builder{})
	cmd.SetArgs([]string{"--slash-command", "/rename Persisted Title", "--print"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() rename error = %v", err)
	}

	sessionIDGeneratorHook = func() string { return "post-rename-startup" }
	runtimeState, err := startupRuntimeState("ollama", "llama3", false, "human", "text", "local")
	if err != nil {
		t.Fatalf("startupRuntimeState() error = %v", err)
	}

	found := false
	for _, entry := range runtimeState.HistoryEntries {
		if entry.ID != "rename-session" {
			continue
		}
		found = true
		if entry.Title != "Persisted Title" {
			t.Fatalf("entry.Title = %q, want %q", entry.Title, "Persisted Title")
		}
	}
	if !found {
		t.Fatalf("expected rename-session entry in startup history")
	}
}

func TestPrintModeIncludesStartupDiagnostics(t *testing.T) {
	originalHook := runStartupMigrationsHook
	originalSessionIDHook := sessionIDGeneratorHook
	originalAvailabilityHook := providerModelAvailabilityHook
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}
	t.Cleanup(func() {
		runStartupMigrationsHook = originalHook
		sessionIDGeneratorHook = originalSessionIDHook
		providerModelAvailabilityHook = originalAvailabilityHook
		_ = os.Chdir(originalWd)
	})

	runStartupMigrationsHook = func() error { return nil }
	providerModelAvailabilityHook = func(_ *config.Config, providerName string) ([]string, error) {
		if providerName == "ollama" {
			return []string{"llama3"}, nil
		}
		return nil, nil
	}
	home := t.TempDir()
	workspace := t.TempDir()
	t.Setenv("HOME", home)
	if err := os.Chdir(workspace); err != nil {
		t.Fatalf("Chdir() error = %v", err)
	}
	sessionIDGeneratorHook = func() string { return "print-diag-session" }

	cmd := newRootCmd()
	b := &strings.Builder{}
	cmd.SetOut(b)
	cmd.SetArgs([]string{"--print"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	out := b.String()
	if !strings.Contains(out, "startup_state=ok") || !strings.Contains(out, "startup_session=ok") || !strings.Contains(out, "startup_history=ok") || !strings.Contains(out, "startup_metadata=ok") {
		t.Fatalf("expected startup diagnostics in output, got %q", out)
	}
}

func TestPrintModeJSONIncludesStartupDiagnostics(t *testing.T) {
	originalHook := runStartupMigrationsHook
	originalSessionIDHook := sessionIDGeneratorHook
	originalAvailabilityHook := providerModelAvailabilityHook
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}
	t.Cleanup(func() {
		runStartupMigrationsHook = originalHook
		sessionIDGeneratorHook = originalSessionIDHook
		providerModelAvailabilityHook = originalAvailabilityHook
		_ = os.Chdir(originalWd)
	})

	runStartupMigrationsHook = func() error { return nil }
	providerModelAvailabilityHook = func(_ *config.Config, providerName string) ([]string, error) {
		if providerName == "ollama" {
			return []string{"llama3"}, nil
		}
		return nil, nil
	}
	t.Setenv("HOME", t.TempDir())
	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatalf("Chdir() error = %v", err)
	}
	sessionIDGeneratorHook = func() string { return "print-json-diag-session" }

	cmd := newRootCmd()
	b := &strings.Builder{}
	cmd.SetOut(b)
	cmd.SetArgs([]string{"--print", "--output-format", "json"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	var payload struct {
		Diagnostics struct {
			StateInitialization   string `json:"state_initialization"`
			SessionInitialization string `json:"session_initialization"`
			HistoryInitialization string `json:"history_initialization"`
			MetadataHydration     string `json:"metadata_hydration"`
		} `json:"diagnostics"`
	}
	if err := json.Unmarshal([]byte(b.String()), &payload); err != nil {
		t.Fatalf("expected json output, got %v", err)
	}
	if payload.Diagnostics.StateInitialization != "ok" || payload.Diagnostics.SessionInitialization != "ok" || payload.Diagnostics.HistoryInitialization != "ok" || payload.Diagnostics.MetadataHydration != "ok" {
		t.Fatalf("expected diagnostics to be ok, got %+v", payload.Diagnostics)
	}
}

func TestStartupRuntimeStatePersistsProjectRegistryAndSettingsCacheStatus(t *testing.T) {
	home := t.TempDir()
	workspace := t.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}
	originalSessionIDHook := sessionIDGeneratorHook
	t.Cleanup(func() {
		_ = os.Chdir(originalWd)
		sessionIDGeneratorHook = originalSessionIDHook
	})

	if err := os.Chdir(workspace); err != nil {
		t.Fatalf("Chdir() error = %v", err)
	}
	t.Setenv("HOME", home)
	sessionIDGeneratorHook = func() string { return "status-augmented-session" }

	runtimeState, err := startupRuntimeState("ollama", "llama3", false, "human", "text", "local")
	if err != nil {
		t.Fatalf("startupRuntimeState() error = %v", err)
	}
	if runtimeState.ProviderName != "ollama" || runtimeState.Model != "llama3" || runtimeState.ModelRef != "ollama/llama3" {
		t.Fatalf("unexpected startup selection: %+v", runtimeState)
	}
	if runtimeState.Runtime.ProviderName != "ollama" || runtimeState.Runtime.Model != "llama3" || runtimeState.Runtime.ModelRef != "ollama/llama3" {
		t.Fatalf("expected hydrated runtime snapshot selection, got %+v", runtimeState.Runtime)
	}

	paths, err := state.ResolvePaths(home, workspace)
	if err != nil {
		t.Fatalf("ResolvePaths() error = %v", err)
	}
	registry, err := state.ReadProjectsRegistry(paths.ProjectsStateFile)
	if err != nil {
		t.Fatalf("ReadProjectsRegistry() error = %v", err)
	}
	if registry.LastOpenedProject != workspace || registry.LastSessionID != "status-augmented-session" {
		t.Fatalf("unexpected projects registry values: %+v", registry)
	}

	cmd := newRootCmd()
	b := &strings.Builder{}
	cmd.SetOut(b)
	if err := runSlashCommand(cmd.Context(), cmd, "/status", runtimeState); err != nil {
		t.Fatalf("runSlashCommand(/status) error = %v", err)
	}
	out := b.String()
	if !strings.Contains(out, "project_registry.count=") {
		t.Fatalf("expected project registry diagnostics in status output, got %q", out)
	}
	if !strings.Contains(out, "settings_cache.fresh=") {
		t.Fatalf("expected settings cache diagnostics in status output, got %q", out)
	}
}

func TestStartupRuntimeStateHydratesFromProjectRegistryWhenMetadataMissing(t *testing.T) {
	home := t.TempDir()
	workspace := t.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}
	originalSessionIDHook := sessionIDGeneratorHook
	t.Cleanup(func() {
		_ = os.Chdir(originalWd)
		sessionIDGeneratorHook = originalSessionIDHook
	})

	if err := os.Chdir(workspace); err != nil {
		t.Fatalf("Chdir() error = %v", err)
	}
	t.Setenv("HOME", home)

	store, err := history.NewJSONLStore(home)
	if err != nil {
		t.Fatalf("NewJSONLStore() error = %v", err)
	}
	if err := store.AppendSessionStartForProject("project-registry-session", workspace); err != nil {
		t.Fatalf("AppendSessionStartForProject() error = %v", err)
	}

	paths, err := state.ResolvePaths(home, workspace)
	if err != nil {
		t.Fatalf("ResolvePaths() error = %v", err)
	}
	if err := state.EnsureState(paths); err != nil {
		t.Fatalf("EnsureState() error = %v", err)
	}
	if err := state.WriteProjectsRegistry(paths.ProjectsStateFile, state.ProjectsRegistry{
		LastSessionID:     "project-registry-session",
		LastOpenedProject: workspace,
		Projects: map[string]state.ProjectMetadata{
			workspace: {
				LastSessionID:      "project-registry-session",
				LastSessionPath:    history.SessionPath(home, "project-registry-session"),
				LastSessionTitle:   "Registry Title",
				LastSessionSummary: "registry summary",
				LastOpenedAt:       time.Now().UTC().Unix(),
				OpenCount:          3,
			},
		},
	}); err != nil {
		t.Fatalf("WriteProjectsRegistry() error = %v", err)
	}

	sessionIDGeneratorHook = func() string { return "new-registry-hydration-session" }
	runtimeState, err := startupRuntimeState("ollama", "llama3", false, "human", "text", "local")
	if err != nil {
		t.Fatalf("startupRuntimeState() error = %v", err)
	}
	if runtimeState.Runtime.ProviderName != "ollama" || runtimeState.Runtime.ModelRef != "ollama/llama3" {
		t.Fatalf("expected startup hydration to seed runtime snapshot, got %+v", runtimeState.Runtime)
	}

	found := false
	newSessionPath := session.SessionPath(home, "new-registry-hydration-session")
	for _, entry := range runtimeState.HistoryEntries {
		if entry.ID != "project-registry-session" {
			continue
		}
		found = true
		if entry.Title != "project-registry-session" {
			t.Fatalf("entry.Title = %q, want project-registry-session", entry.Title)
		}
		if entry.Summary != "events=1 messages=0" {
			t.Fatalf("entry.Summary = %q, want events=1 messages=0", entry.Summary)
		}
		if entry.Path == newSessionPath {
			t.Fatalf("entry.Path unexpectedly overwritten by new startup session path: %q", entry.Path)
		}
	}
	if !found {
		t.Fatalf("expected project-registry-session entry in startup history")
	}
}
