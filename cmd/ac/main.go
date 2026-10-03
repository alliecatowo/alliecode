package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/alliecatowo/alliecode/internal/agent"
	"github.com/alliecatowo/alliecode/internal/commands"
	"github.com/alliecatowo/alliecode/internal/config"
	"github.com/alliecatowo/alliecode/internal/history"
	"github.com/alliecatowo/alliecode/internal/migrations"
	"github.com/alliecatowo/alliecode/internal/permissions"
	"github.com/alliecatowo/alliecode/internal/providers"
	"github.com/alliecatowo/alliecode/internal/session"
	"github.com/alliecatowo/alliecode/internal/state"
	"github.com/alliecatowo/alliecode/internal/tools"
	"github.com/alliecatowo/alliecode/internal/tui"
	"github.com/alliecatowo/alliecode/internal/types"
)

const version = "dev"

var runStartupMigrationsHook = runStartupMigrations
var runTUIHook = runTUI
var sessionIDGeneratorHook = generateSessionID
var providerModelAvailabilityHook = fetchProviderModelIDs

var startupMigrations = []migrations.Migration{}

func main() {
	if err := newRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func newRootCmd() *cobra.Command {
	var (
		configPath     string
		provider       string
		model          string
		debug          bool
		printMode      bool
		skipOnboarding bool
		slashInput     string
		outStyle       string
		outFormat      string
		transport      string
	)

	cmd := &cobra.Command{
		Use:     "ac",
		Aliases: []string{"alliecode"},
		Short:   "AllieCode CLI",
		Long:    "AllieCode is an open-source, model-agnostic AI coding agent.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := loadCLIConfig(configPath)
			if err != nil {
				return renderConfigLoadRepairGuidance(cmd, configPath, err)
			}

			resolvedStyle, _, err := normalizeOutputStyle(outStyle)
			if err != nil {
				return err
			}
			outStyle = resolvedStyle
			resolvedFormat, _, err := normalizeOutputFormat(outFormat)
			if err != nil {
				return err
			}
			outFormat = resolvedFormat
			resolvedTransport, _, err := normalizeTransportMode(transport)
			if err != nil {
				return err
			}
			transport = resolvedTransport

			if err := runStartupMigrationsHook(); err != nil {
				return fmt.Errorf("startup migrations failed: %w", err)
			}

			activeProvider, activeModel := resolveRuntimeSelection(cfg, provider, model)

			if strings.TrimSpace(slashInput) != "" {
				runtimeState, err := startupRuntimeState(activeProvider, activeModel, printMode, outStyle, outFormat, transport)
				if err != nil {
					return err
				}
				return runSlashCommand(cmd.Context(), cmd, slashInput, runtimeState)
			}

			if printMode {
				_, diagnostics, err := startupRuntimeStateWithDiagnostics(activeProvider, activeModel, printMode, outStyle, outFormat, transport)
				if err != nil {
					return err
				}
				return renderStartupInfo(cmd, startupInfo{
					ConfigPath:   configPath,
					Provider:     activeProvider,
					Model:        activeModel,
					Debug:        debug,
					PrintMode:    printMode,
					OutputStyle:  outStyle,
					OutputFormat: outFormat,
					Transport:    transport,
					Diagnostics:  diagnostics,
				})
			}

			progress, progressErr := readOnboardingProgress()
			if progressErr != nil {
				return fmt.Errorf("load onboarding progress: %w", progressErr)
			}

			skipInteractiveSetup := skipOnboarding || envFlagEnabled("ALLIECODE_SKIP_ONBOARDING")
			if repair := evaluateStartupRepair(cfg, activeProvider, activeModel); repair.Required && !skipInteractiveSetup {
				if err := runStartupRepair(cmd, cfg, repair); err != nil {
					return err
				}
				activeProvider, activeModel = resolveRuntimeSelection(cfg, provider, model)
			}

			if shouldRunOnboarding(cfg, skipInteractiveSetup) {
				if err := runOnboarding(cmd, cfg, progress); err != nil {
					_ = writeOnboardingProgress(onboardingProgress{HasCompleted: false, SeenCount: progress.SeenCount + 1})
					return err
				}
				_ = writeOnboardingProgress(onboardingProgress{HasCompleted: true, SeenCount: progress.SeenCount + 1})
				activeProvider, activeModel = resolveRuntimeSelection(cfg, provider, model)
			}

			runtimeState, err := startupRuntimeState(activeProvider, activeModel, printMode, outStyle, outFormat, transport)
			if err != nil {
				return err
			}

			return runTUIHook(runtimeState, debug, cfg)
		},
	}

	cmd.Flags().StringVar(&configPath, "config", "", "Path to config file")
	cmd.Flags().StringVar(&provider, "provider", "", "LLM provider")
	cmd.Flags().StringVarP(&model, "model", "m", "", "Model name")
	cmd.Flags().BoolVar(&debug, "debug", false, "Enable debug mode")
	cmd.Flags().BoolVar(&printMode, "print", false, "Print resolved startup settings and exit")
	cmd.Flags().BoolVar(&skipOnboarding, "skip-onboarding", false, "Skip first-run onboarding (useful for CI)")
	cmd.Flags().StringVar(&slashInput, "slash-command", "", "Dispatch one slash command non-interactively and exit")
	cmd.Flags().StringVar(&outStyle, "output-style", "human", "Output style placeholder: human|compact")
	cmd.Flags().StringVar(&outFormat, "output-format", "text", "Output format placeholder: text|json")
	cmd.Flags().StringVar(&transport, "transport", "local", "Transport mode placeholder: local|remote")

	cmd.AddCommand(newModelsCmd(&provider, &configPath))

	return cmd
}

func resolveRuntimeSelection(cfg *config.Config, provider, model string) (string, string) {
	activeProvider := provider
	if activeProvider == "" {
		activeProvider = cfg.DefaultProvider
	}
	if activeProvider == "" {
		activeProvider = "ollama"
	}

	activeModel := model
	if activeModel == "" {
		activeModel = cfg.DefaultModel
	}
	if explicitProvider, explicitModel, ok := parseExplicitProviderModel(activeModel); ok {
		activeProvider = explicitProvider
		activeModel = explicitModel
	}
	activeModel = preferredModelForProvider(activeProvider, activeModel)

	return activeProvider, activeModel
}

func parseExplicitProviderModel(raw string) (string, string, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" || !strings.Contains(raw, "/") {
		return "", "", false
	}
	parts := strings.SplitN(raw, "/", 2)
	providerName := strings.ToLower(strings.TrimSpace(parts[0]))
	modelName := strings.TrimSpace(parts[1])
	if providerName == "" || modelName == "" {
		return "", "", false
	}
	if len(providers.ListModelsByProvider(providerName)) == 0 {
		return "", "", false
	}
	return providerName, modelName, true
}

func shouldRunOnboarding(cfg *config.Config, bypass bool) bool {
	if bypass || cfg == nil {
		return false
	}
	if cfg.HasCompletedOnboarding {
		return false
	}
	return !hasAnyProviderCredentials(cfg)
}

func hasAnyProviderCredentials(cfg *config.Config) bool {
	if cfg == nil {
		return false
	}
	for _, provider := range cfg.Providers {
		if provider == nil {
			continue
		}
		if strings.TrimSpace(provider.APIKey) != "" ||
			strings.TrimSpace(provider.AuthToken) != "" ||
			strings.TrimSpace(provider.AccessToken) != "" {
			return true
		}
	}
	return false
}

func envFlagEnabled(name string) bool {
	raw, ok := os.LookupEnv(name)
	if !ok {
		return false
	}
	v := strings.ToLower(strings.TrimSpace(raw))
	switch v {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

type startupRepair struct {
	Required bool
	Issues   []string
	Check    startupSelectionValidation
}

type onboardingProgress struct {
	HasCompleted bool
	SeenCount    int
}

type startupSelectionValidation struct {
	Provider           string
	Model              string
	CredentialPresence config.ProviderCredentialPresence
	ModelRecognized    bool
	ModelChecked       bool
	ModelAvailable     bool
	ModelCheckError    string
	AvailableModels    []string
	Issues             []string
}

func evaluateStartupRepair(cfg *config.Config, providerName, modelName string) startupRepair {
	if cfg == nil {
		return startupRepair{Required: true, Issues: []string{"configuration is unavailable"}}
	}
	issues := make([]string, 0, 6)
	if err := cfg.Validate(); err != nil {
		issues = append(issues, "configuration contains invalid values")
	}
	check := validateStartupSelectionWithAvailability(cfg, providerName, modelName)
	issues = append(issues, check.Issues...)
	return startupRepair{Required: len(issues) > 0, Issues: issues, Check: check}
}

func runStartupRepair(cmd *cobra.Command, cfg *config.Config, repair startupRepair) error {
	out := cmd.OutOrStdout()
	in := cmd.InOrStdin()
	reader := bufio.NewReader(in)

	fmt.Fprintln(out, "Your current startup config needs repair:")
	for _, issue := range repair.Issues {
		fmt.Fprintf(out, "- %s\n", issue)
	}
	fmt.Fprintln(out, "Guided steps:")
	fmt.Fprintln(out, "  1) validate provider credentials")
	fmt.Fprintln(out, "  2) validate model availability")
	fmt.Fprintln(out, "  3) confirm and save")
	if quick := startupRepairQuickFix(repair.Check); quick != "" {
		fmt.Fprintf(out, "Quick fix: %s\n", quick)
	}
	for i, line := range startupWorkflowGuidanceLines() {
		fmt.Fprintf(out, "  cmd.%d %s\n", i+1, line)
	}
	for i, loop := range startupCorrectiveLoops(repair.Check) {
		fmt.Fprintf(out, "  loop.%d [%s] state=%s action=%s next=%s\n", i+1, loop.Area, loop.State, loop.Action, loop.Next)
	}
	fmt.Fprint(out, "Run guided repair now? [Y/n]: ")
	raw, err := readLine(reader)
	if err != nil {
		return err
	}
	if strings.EqualFold(raw, "n") || strings.EqualFold(raw, "no") {
		return fmt.Errorf("startup config remains unready; rerun and accept guided repair, or use /provider set <name> and /model <provider/model>")
	}
	progress, err := readOnboardingProgress()
	if err != nil {
		return err
	}
	return runOnboarding(cmd, cfg, progress)
}

func startupWorkflowGuidanceLines() []string {
	return []string{
		"/doctor fix",
		"/config doctor",
		"/permissions summary",
		"/history status",
		"/session diagnostics",
		"/mcp diagnostics",
	}
}

func startupRepairQuickFix(check startupSelectionValidation) string {
	providerName := strings.TrimSpace(check.Provider)
	modelName := strings.TrimSpace(check.Model)
	if providerName == "" {
		return "set a provider quickly with /provider set ollama"
	}
	if !check.CredentialPresence.HasCredentials && providerName != "ollama" {
		envVars := strings.Join(check.CredentialPresence.EnvVars, "|")
		if strings.TrimSpace(envVars) == "" {
			envVars = "provider-specific env vars"
		}
		return fmt.Sprintf("add credentials for %s (config or %s), then run /login provider %s and /provider status", providerName, envVars, providerName)
	}
	if modelName == "" || !check.ModelRecognized || (check.ModelChecked && !check.ModelAvailable) {
		if providerName == "" {
			providerName = "<provider>"
		}
		return fmt.Sprintf("pick a valid model with /model list %s then /model %s/<model>", providerName, providerName)
	}
	if check.ModelCheckError != "" {
		return fmt.Sprintf("recheck provider auth/connectivity, then run /login provider %s, /provider status, and /model %s/%s", providerName, providerName, modelName)
	}
	return "run /provider status, then /model list <provider> and /model <provider>/<model> to correct settings quickly"
}

type startupCorrectiveLoop struct {
	Area   string
	State  string
	Action string
	Next   string
}

func startupCorrectiveLoops(check startupSelectionValidation) []startupCorrectiveLoop {
	providerState := "ready"
	providerAction := "/provider status"
	providerNext := "/provider doctor"
	if strings.TrimSpace(check.Provider) == "" {
		providerState = "missing"
		providerAction = "/provider set ollama"
		providerNext = "/provider status"
	} else if !check.CredentialPresence.HasCredentials && strings.TrimSpace(check.Provider) != "ollama" {
		providerState = "degraded"
		providerAction = "/login provider " + strings.TrimSpace(check.Provider)
		providerNext = "/provider status"
	}

	modelState := "ready"
	modelAction := "/model doctor"
	modelNext := "/status"
	if strings.TrimSpace(check.Model) == "" || !check.ModelRecognized {
		modelState = "missing"
		provider := strings.TrimSpace(check.Provider)
		if provider == "" {
			provider = "<provider>"
		}
		modelAction = "/model list " + provider
		modelNext = "/model " + provider + "/<model>"
	} else if check.ModelChecked && !check.ModelAvailable {
		modelState = "degraded"
		provider := strings.TrimSpace(check.Provider)
		if provider == "" {
			provider = "<provider>"
		}
		modelAction = "/model list " + provider
		modelNext = "/model repair " + provider + "/<model>"
	}

	permissionsState := "ready"
	permissionsAction := "/permissions summary"
	permissionsNext := "/permissions set auto"

	settingsState := "ready"
	settingsAction := "/config doctor"
	settingsNext := "/config repair"
	if len(check.Issues) > 0 {
		settingsState = "degraded"
	}

	historyState := "ready"
	historyAction := "/history status"
	historyNext := "/history latest"

	sessionState := "ready"
	sessionAction := "/session status"
	sessionNext := "/session diagnostics"

	mcpState := "ready"
	mcpAction := "/mcp diagnostics"
	mcpNext := "/mcp repair auto"

	doctorState := "ready"
	doctorAction := "/doctor"
	doctorNext := "/doctor fix"
	if len(check.Issues) > 0 {
		doctorState = "degraded"
		doctorAction = "/doctor fix"
		doctorNext = "/status diagnostics"
	}

	return []startupCorrectiveLoop{
		{Area: "provider", State: providerState, Action: providerAction, Next: providerNext},
		{Area: "model", State: modelState, Action: modelAction, Next: modelNext},
		{Area: "permissions", State: permissionsState, Action: permissionsAction, Next: permissionsNext},
		{Area: "settings", State: settingsState, Action: settingsAction, Next: settingsNext},
		{Area: "history", State: historyState, Action: historyAction, Next: historyNext},
		{Area: "session", State: sessionState, Action: sessionAction, Next: sessionNext},
		{Area: "mcp", State: mcpState, Action: mcpAction, Next: mcpNext},
		{Area: "doctor", State: doctorState, Action: doctorAction, Next: doctorNext},
	}
}

func validateStartupSelectionWithAvailability(cfg *config.Config, providerName, modelName string) startupSelectionValidation {
	providerName = strings.ToLower(strings.TrimSpace(providerName))
	modelName = strings.TrimSpace(modelName)
	check := startupSelectionValidation{
		Provider: providerName,
		Model:    modelName,
	}
	issues := make([]string, 0, 6)
	if providerName == "" {
		issues = append(issues, "default provider is not set")
		check.Issues = issues
		return check
	}
	if len(onboardingModelChoices(providerName)) == 0 {
		issues = append(issues, fmt.Sprintf("provider %q is not recognized by this build", providerName))
		check.Issues = issues
		return check
	}

	check.CredentialPresence = cfg.ProviderCredentialPresence(providerName)
	if providerName != "ollama" && !check.CredentialPresence.HasCredentials {
		sourceHint := "credentials"
		if len(check.CredentialPresence.EnvVars) > 0 {
			sourceHint = fmt.Sprintf("credentials (config or %s)", strings.Join(check.CredentialPresence.EnvVars, "|"))
		}
		issues = append(issues, fmt.Sprintf("provider %q has no usable %s (run /login provider %s after adding credentials)", providerName, sourceHint, providerName))
	}

	if modelName == "" {
		issues = append(issues, "default model is not set")
		check.Issues = issues
		return check
	}
	if _, ok := providers.LookupModelMetadata(providerName, modelName); !ok {
		issues = append(issues, fmt.Sprintf("model %q is not recognized for provider %q (run /model list %s, then /model %s/<model>)", modelName, providerName, providerName, providerName))
		check.Issues = issues
		return check
	}
	check.ModelRecognized = true

	if providerName != "ollama" && !check.CredentialPresence.HasCredentials {
		check.Issues = issues
		return check
	}

	availableModels, err := providerModelAvailabilityHook(cfg, providerName)
	check.ModelChecked = true
	if err != nil {
		check.ModelCheckError = err.Error()
		check.Issues = issues
		return check
	}
	check.AvailableModels = availableModels
	for _, available := range availableModels {
		if strings.EqualFold(strings.TrimSpace(available), modelName) {
			check.ModelAvailable = true
			break
		}
	}
	if !check.ModelAvailable {
		issues = append(issues, fmt.Sprintf("model %q is not currently available from provider %q (run /model list %s and pick an available id)", modelName, providerName, providerName))
	}
	check.Issues = issues
	return check
}

func fetchProviderModelIDs(cfg *config.Config, providerName string) ([]string, error) {
	providerName = strings.ToLower(strings.TrimSpace(providerName))
	if providerName == "" {
		return nil, fmt.Errorf("provider is required")
	}
	prov, err := providers.New(providerName, cfg.GetProviderSettings(providerName))
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	models, err := prov.ListModels(ctx)
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(models))
	seen := make(map[string]struct{}, len(models))
	for _, m := range models {
		id := strings.TrimSpace(m.ID)
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids, nil
}

func renderConfigLoadRepairGuidance(cmd *cobra.Command, configPath string, loadErr error) error {
	out := cmd.OutOrStdout()
	fmt.Fprintln(out, "Unable to load configuration.")
	fmt.Fprintf(out, "Reason: %v\n", loadErr)

	if strings.TrimSpace(configPath) != "" {
		fmt.Fprintln(out, "Guided repair steps:")
		fmt.Fprintf(out, "  1) open %s and fix YAML/validation errors\n", configPath)
		fmt.Fprintln(out, "  2) rerun ac --print to verify startup settings")
		fmt.Fprintln(out, "  3) rerun ac to launch guided onboarding if needed")
		for i, line := range startupWorkflowGuidanceLines() {
			fmt.Fprintf(out, "  cmd.%d %s\n", i+1, line)
		}
		return fmt.Errorf("load config: %w", loadErr)
	}

	wd, err := os.Getwd()
	if err == nil {
		globalPath, projectPath, pathErr := config.ConventionalPaths(wd)
		if pathErr == nil {
			diags := append(config.PathDiagnostics(globalPath, projectPath), config.Diagnostic{
				Severity:    config.DiagnosticError,
				Code:        "config_load_failed",
				Message:     loadErr.Error(),
				Remediation: "fix the reported file, or run onboarding to regenerate defaults",
			})
			rendered := config.FormatDiagnostics(diags)
			if strings.TrimSpace(rendered) != "" {
				fmt.Fprintln(out, rendered)
			}
			fmt.Fprintln(out, "Guided repair steps:")
			fmt.Fprintf(out, "  1) inspect project config: %s\n", projectPath)
			fmt.Fprintf(out, "  2) inspect global config: %s\n", globalPath)
			fmt.Fprintln(out, "  3) rerun ac after fixing invalid fields")
			for i, line := range startupWorkflowGuidanceLines() {
				fmt.Fprintf(out, "  cmd.%d %s\n", i+1, line)
			}
		}
	}

	return fmt.Errorf("load config: %w", loadErr)
}

func runOnboarding(cmd *cobra.Command, cfg *config.Config, progress onboardingProgress) error {
	out := cmd.OutOrStdout()
	in := cmd.InOrStdin()
	reader := bufio.NewReader(in)

	fmt.Fprintln(out, "Welcome to AllieCode!")
	if progress.SeenCount > 0 && !progress.HasCompleted {
		fmt.Fprintf(out, "Resuming onboarding attempt #%d. Prior choices can be adjusted before save.\n", progress.SeenCount+1)
	}
	fmt.Fprintln(out, "Let's configure your startup provider + model.")

	providerChoices := onboardingProviderChoices()
	selectedProvider := preferredProvider(cfg.DefaultProvider)
	selectedModel := preferredModelForProvider(selectedProvider, cfg.DefaultModel)

	for {
		providerChoice, err := promptProviderChoice(out, reader, providerChoices, selectedProvider)
		if err != nil {
			return fmt.Errorf("onboarding provider selection: %w", err)
		}
		selectedProvider = providerChoice

		switch err := ensureOnboardingProviderCredentials(out, reader, cfg, selectedProvider); {
		case err == nil:
		case errors.Is(err, errOnboardingChangeProvider):
			continue
		default:
			return fmt.Errorf("onboarding credential setup: %w", err)
		}

		modelChoices, modelCheckMessage := onboardingModelChoicesWithAvailability(cfg, selectedProvider)
		if strings.TrimSpace(modelCheckMessage) != "" {
			fmt.Fprintln(out, modelCheckMessage)
		}
		selectedModel, err = promptModelChoice(out, reader, selectedProvider, modelChoices, preferredModelForProvider(selectedProvider, selectedModel))
		if err != nil {
			return fmt.Errorf("onboarding model selection: %w", err)
		}

		check := validateStartupSelectionWithAvailability(cfg, selectedProvider, selectedModel)
		if len(check.Issues) > 0 {
			fmt.Fprintln(out, "Startup validation checks:")
			for _, issue := range check.Issues {
				fmt.Fprintf(out, "- %s\n", issue)
			}
			if quick := startupRepairQuickFix(check); quick != "" {
				fmt.Fprintf(out, "Quick fix: %s\n", quick)
			}
			action, actionErr := promptOnboardingValidationAction(out, reader, selectedProvider)
			if actionErr != nil {
				return fmt.Errorf("onboarding validation action: %w", actionErr)
			}
			switch action {
			case "provider":
				continue
			case "credentials":
				switch err := ensureOnboardingProviderCredentials(out, reader, cfg, selectedProvider); {
				case err == nil:
					continue
				case errors.Is(err, errOnboardingChangeProvider):
					continue
				default:
					return fmt.Errorf("onboarding credential setup: %w", err)
				}
			case "model":
				continue
			default:
				return fmt.Errorf("onboarding cancelled")
			}
		}

		action, err := promptOnboardingConfirmation(out, reader, selectedProvider, selectedModel)
		if err != nil {
			return fmt.Errorf("onboarding confirmation: %w", err)
		}
		switch action {
		case "save":
			if err := applyOnboardingSelection(cfg, selectedProvider, selectedModel); err != nil {
				return err
			}
			if err := persistOnboardingConfig(cfg); err != nil {
				return fmt.Errorf("save onboarding config: %w", err)
			}
			fmt.Fprintf(out, "Saved provider %q and model %q. Starting AllieCode...\n", selectedProvider, selectedModel)
			return nil
		case "provider":
			continue
		case "model":
			continue
		default:
			return fmt.Errorf("onboarding cancelled")
		}
	}
}

var errOnboardingChangeProvider = errors.New("change onboarding provider")

func ensureOnboardingProviderCredentials(out io.Writer, reader *bufio.Reader, cfg *config.Config, providerName string) error {
	providerName = strings.ToLower(strings.TrimSpace(providerName))
	if providerName == "" || providerName == "ollama" {
		return nil
	}
	presence := cfg.ProviderCredentialPresence(providerName)
	if presence.HasCredentials {
		fmt.Fprintf(out, "Credential check: %s credentials found (%s).\n", providerName, credentialSourceLabel(presence))
		return nil
	}

	envHint := ""
	if len(presence.EnvVars) > 0 {
		envHint = strings.Join(presence.EnvVars, "|")
	}
	fmt.Fprintf(out, "Credential check: %s credentials missing.\n", providerName)
	if envHint != "" {
		fmt.Fprintf(out, "Provide credentials in config or env (%s).\n", envHint)
	}
	fmt.Fprintln(out, "Choose next step:")
	fmt.Fprintf(out, "Hint: /login provider %s verifies auth for the active provider.\n", providerName)
	fmt.Fprintln(out, "  1) Enter credentials now")
	fmt.Fprintln(out, "  2) Choose different provider")
	fmt.Fprintln(out, "  3) Continue without credentials (startup remains unready for remote providers)")
	fmt.Fprint(out, "Select action [1]: ")
	action, err := readLine(reader)
	if err != nil {
		return err
	}
	if action == "" || action == "1" {
		return promptAndStoreProviderCredential(out, reader, cfg, providerName)
	}
	switch strings.ToLower(strings.TrimSpace(action)) {
	case "2", "provider", "change provider":
		return errOnboardingChangeProvider
	case "3", "continue", "skip":
		return nil
	default:
		return fmt.Errorf("unknown credential action %q", action)
	}
}

func credentialSourceLabel(presence config.ProviderCredentialPresence) string {
	switch {
	case presence.FromConfig && presence.FromEnv:
		return "config+env"
	case presence.FromConfig:
		return "config"
	case presence.FromEnv:
		return "env"
	default:
		return "none"
	}
}

func promptAndStoreProviderCredential(out io.Writer, reader *bufio.Reader, cfg *config.Config, providerName string) error {
	if cfg.Providers == nil {
		cfg.Providers = make(map[string]*config.ProviderSettings)
	}
	settings := cfg.GetProviderSettings(providerName)
	if settings == nil {
		settings = &config.ProviderSettings{}
	}
	providerName = strings.ToLower(strings.TrimSpace(providerName))

	switch providerName {
	case "openai", "gemini":
		fmt.Fprintf(out, "Enter %s API key: ", providerName)
		value, err := readLine(reader)
		if err != nil {
			return err
		}
		value = strings.TrimSpace(value)
		if value == "" {
			return fmt.Errorf("credential cannot be empty")
		}
		settings.APIKey = value
	case "anthropic":
		fmt.Fprint(out, "Enter Anthropic auth token (preferred, blank to use API key): ")
		token, err := readLine(reader)
		if err != nil {
			return err
		}
		token = strings.TrimSpace(token)
		if token != "" {
			settings.AuthToken = token
			settings.APIKey = ""
			settings.AccessToken = ""
		} else {
			fmt.Fprint(out, "Enter Anthropic API key: ")
			apiKey, err := readLine(reader)
			if err != nil {
				return err
			}
			apiKey = strings.TrimSpace(apiKey)
			if apiKey == "" {
				return fmt.Errorf("credential cannot be empty")
			}
			settings.APIKey = apiKey
		}
	default:
		return nil
	}

	cfg.Providers[providerName] = settings
	fmt.Fprintf(out, "Credential saved for %s in current onboarding session.\n", providerName)
	return nil
}

func onboardingModelChoicesWithAvailability(cfg *config.Config, providerName string) ([]string, string) {
	providerName = strings.ToLower(strings.TrimSpace(providerName))
	staticChoices := onboardingModelChoices(providerName)
	if cfg == nil || providerName == "" {
		return staticChoices, ""
	}
	if providerName != "ollama" && !cfg.ProviderCredentialPresence(providerName).HasCredentials {
		return staticChoices, fmt.Sprintf("Model availability check: skipped live provider check for %s (credentials missing). Recovery: add credentials, run /login provider %s, then rerun model selection.", providerName, providerName)
	}
	available, err := providerModelAvailabilityHook(cfg, providerName)
	if err != nil {
		return staticChoices, fmt.Sprintf("Model availability check: unable to verify live models for %s (%v). Recovery: run /provider status, then /model list %s.", providerName, err, providerName)
	}
	if len(available) == 0 {
		return staticChoices, fmt.Sprintf("Model availability check: provider %s returned no models; using known defaults. Recovery: run /model list %s and choose a listed id.", providerName, providerName)
	}
	return available, fmt.Sprintf("Model availability check: %d live model(s) available for %s.", len(available), providerName)
}

func promptOnboardingValidationAction(out io.Writer, reader *bufio.Reader, providerName string) (string, error) {
	fmt.Fprintln(out, "Choose a correction step:")
	fmt.Fprintln(out, "Recommended order: provider auth -> model selection -> save")
	fmt.Fprintln(out, "  1) Change provider")
	fmt.Fprintln(out, "  2) Re-enter credentials")
	fmt.Fprintln(out, "  3) Change model")
	fmt.Fprintln(out, "  4) Cancel")
	fmt.Fprint(out, "Select action [3]: ")
	raw, err := readLine(reader)
	if err != nil {
		return "", err
	}
	if raw == "" || raw == "3" {
		return "model", nil
	}
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "1", "provider", "change provider":
		return "provider", nil
	case "2", "credentials", "credential", "login":
		if strings.EqualFold(strings.TrimSpace(providerName), "ollama") {
			return "provider", nil
		}
		return "credentials", nil
	case "4", "cancel", "quit", "exit":
		return "cancel", nil
	default:
		return "", fmt.Errorf("unknown validation action %q", raw)
	}
}

func applyOnboardingSelection(cfg *config.Config, providerName, modelName string) error {
	if cfg.Providers == nil {
		cfg.Providers = make(map[string]*config.ProviderSettings)
	}
	providerName = strings.TrimSpace(providerName)
	modelName = strings.TrimSpace(modelName)
	if providerName == "" {
		return fmt.Errorf("onboarding provider cannot be empty")
	}
	if modelName == "" {
		return fmt.Errorf("onboarding model cannot be empty")
	}

	cfg.DefaultProvider = providerName
	cfg.DefaultModel = modelName
	cfg.OnboardingProviderChoice = providerName
	cfg.HasCompletedOnboarding = true

	if providerName == "ollama" {
		settings := cfg.GetProviderSettings("ollama")
		if settings == nil {
			settings = &config.ProviderSettings{}
		}
		if strings.TrimSpace(settings.BaseURL) == "" {
			settings.BaseURL = "http://localhost:11434"
		}
		cfg.Providers["ollama"] = settings
	}

	return nil
}

func onboardingProviderChoices() []string {
	return []string{"ollama", "openai", "anthropic", "gemini"}
}

func preferredProvider(raw string) string {
	return commands.PreferredProvider(raw)
}

func preferredModelForProvider(providerName, raw string) string {
	return commands.PreferredModelForProvider(providerName, raw)
}

func onboardingModelChoices(providerName string) []string {
	metas := providers.ListModelsByProvider(providerName)
	if len(metas) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(metas))
	out := make([]string, 0, len(metas))
	for _, meta := range metas {
		name := strings.TrimSpace(meta.Model)
		if name == "" {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}

func promptProviderChoice(out io.Writer, reader *bufio.Reader, choices []string, fallback string) (string, error) {
	fmt.Fprintln(out, "Choose a provider:")
	defaultIndex := 1
	for i, name := range choices {
		marker := ""
		if name == fallback {
			marker = " (default)"
			defaultIndex = i + 1
		}
		fmt.Fprintf(out, "  %d) %s%s\n", i+1, name, marker)
	}
	fmt.Fprintf(out, "Select provider [%d]: ", defaultIndex)
	raw, err := readLine(reader)
	if err != nil {
		return "", err
	}
	if raw == "" {
		return fallback, nil
	}
	for i, option := range choices {
		if raw == strconv.Itoa(i+1) || strings.EqualFold(raw, option) {
			return option, nil
		}
	}
	return "", fmt.Errorf("unknown provider choice %q", raw)
}

func promptModelChoice(out io.Writer, reader *bufio.Reader, providerName string, choices []string, fallback string) (string, error) {
	if len(choices) == 0 {
		return "", fmt.Errorf("no known models available for provider %q", providerName)
	}
	fmt.Fprintf(out, "Choose a model for %s:\n", providerName)
	defaultIndex := 1
	for i, modelName := range choices {
		marker := ""
		if modelName == fallback {
			marker = " (default)"
			defaultIndex = i + 1
		}
		fmt.Fprintf(out, "  %d) %s%s\n", i+1, modelName, marker)
	}
	fmt.Fprintf(out, "Select model [%d]: ", defaultIndex)
	raw, err := readLine(reader)
	if err != nil {
		return "", err
	}
	if raw == "" {
		return fallback, nil
	}
	for i, modelName := range choices {
		if raw == strconv.Itoa(i+1) || strings.EqualFold(raw, modelName) {
			return modelName, nil
		}
	}
	return "", fmt.Errorf("unknown model choice %q", raw)
}

func promptOnboardingConfirmation(out io.Writer, reader *bufio.Reader, providerName, modelName string) (string, error) {
	fmt.Fprintln(out, "Summary:")
	fmt.Fprintf(out, "  provider: %s\n", providerName)
	fmt.Fprintf(out, "  model:    %s\n", modelName)
	fmt.Fprintln(out, "Confirm setup:")
	fmt.Fprintln(out, "  1) Save and continue")
	fmt.Fprintln(out, "  2) Change provider")
	fmt.Fprintln(out, "  3) Change model")
	fmt.Fprintln(out, "  4) Cancel")
	fmt.Fprint(out, "Select action [1]: ")
	raw, err := readLine(reader)
	if err != nil {
		return "", err
	}
	if raw == "" || raw == "1" {
		return "save", nil
	}
	switch strings.ToLower(raw) {
	case "2", "provider", "change provider":
		return "provider", nil
	case "3", "model", "change model":
		return "model", nil
	case "4", "cancel", "exit", "quit":
		return "cancel", nil
	default:
		return "", fmt.Errorf("unknown confirmation choice %q", raw)
	}
}

func readLine(reader *bufio.Reader) (string, error) {
	raw, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", err
	}
	return strings.TrimSpace(raw), nil
}

func readOnboardingProgress() (onboardingProgress, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return onboardingProgress{}, err
	}
	projectDir, err := os.Getwd()
	if err != nil {
		return onboardingProgress{}, err
	}
	paths, err := state.ResolvePaths(homeDir, projectDir)
	if err != nil {
		return onboardingProgress{}, err
	}
	if err := state.EnsureProjectState(paths); err != nil {
		return onboardingProgress{}, err
	}
	projectState, err := state.ReadProjectOnboardingState(paths.ProjectOnboardingStateFile)
	if err != nil {
		return onboardingProgress{}, err
	}
	return onboardingProgress{HasCompleted: projectState.HasCompleted, SeenCount: projectState.SeenCount}, nil
}

func writeOnboardingProgress(progress onboardingProgress) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	projectDir, err := os.Getwd()
	if err != nil {
		return err
	}
	paths, err := state.ResolvePaths(homeDir, projectDir)
	if err != nil {
		return err
	}
	if err := state.EnsureProjectState(paths); err != nil {
		return err
	}
	return state.WriteProjectOnboardingState(paths.ProjectOnboardingStateFile, state.ProjectOnboardingState{
		HasCompleted: progress.HasCompleted,
		SeenCount:    progress.SeenCount,
	})
}

func persistOnboardingConfig(cfg *config.Config) error {
	globalCfg, err := config.Load("")
	if err != nil {
		return err
	}

	globalCfg.DefaultProvider = cfg.DefaultProvider
	globalCfg.DefaultModel = cfg.DefaultModel
	globalCfg.HasCompletedOnboarding = cfg.HasCompletedOnboarding
	globalCfg.OnboardingProviderChoice = cfg.OnboardingProviderChoice

	if globalCfg.Providers == nil {
		globalCfg.Providers = make(map[string]*config.ProviderSettings)
	}
	for name, provider := range cfg.Providers {
		if provider == nil {
			continue
		}
		clone := *provider
		if provider.Options != nil {
			clone.Options = make(map[string]string, len(provider.Options))
			for k, v := range provider.Options {
				clone.Options[k] = v
			}
		}
		globalCfg.Providers[name] = &clone
	}

	return globalCfg.Save()
}

func runTUI(runtimeState commands.RuntimeState, debug bool, cfg *config.Config) error {
	providerName := strings.TrimSpace(runtimeState.ProviderName)
	modelName := strings.TrimSpace(runtimeState.Model)
	prov, err := providers.New(providerName, cfg.GetProviderSettings(providerName))
	if err != nil {
		return err
	}

	toolRegistry := tools.DefaultRegistry()
	resolveProvider := func(name string) (types.Provider, error) {
		return providers.New(name, cfg.GetProviderSettings(name))
	}
	ag := agent.New(agent.Config{
		Provider:          prov,
		ProviderName:      providerName,
		ResolveProvider:   resolveProvider,
		Tools:             toolRegistry.All(),
		Model:             modelName,
		WorkingDir:        ".",
		Debug:             debug,
		EnableCheckpoints: true,
	})
	runtimeState.Agent = ag
	commands.HydrateRuntimeSelection(&runtimeState)

	app := tui.New(tui.Config{
		Agent:           ag,
		Version:         version,
		Debug:           debug,
		InitialModel:    modelName,
		InitialState:    runtimeState,
		ActiveSessionID: runtimeState.SessionID,
	})
	return app.Run()
}

func newModelsCmd(provider *string, configPath *string) *cobra.Command {
	return &cobra.Command{
		Use:   "models",
		Short: "List available models for current provider",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := loadCLIConfig(*configPath)
			if err != nil {
				return err
			}

			activeProvider := *provider
			if activeProvider == "" {
				activeProvider = cfg.DefaultProvider
			}
			if activeProvider == "" {
				activeProvider = "ollama"
			}

			prov, err := providers.New(activeProvider, cfg.GetProviderSettings(activeProvider))
			if err != nil {
				return err
			}

			models, err := prov.ListModels(cmd.Context())
			if err != nil {
				return err
			}

			for _, m := range models {
				fmt.Fprintln(cmd.OutOrStdout(), m.ID)
			}
			return nil
		},
	}
}

type startupInfo struct {
	ConfigPath   string             `json:"config_path"`
	Provider     string             `json:"provider"`
	Model        string             `json:"model"`
	Debug        bool               `json:"debug"`
	PrintMode    bool               `json:"print"`
	OutputStyle  string             `json:"output_style"`
	OutputFormat string             `json:"output_format"`
	Transport    string             `json:"transport"`
	Diagnostics  startupDiagnostics `json:"diagnostics"`
}

type startupDiagnostics struct {
	StateInitialization   string `json:"state_initialization"`
	SessionInitialization string `json:"session_initialization"`
	HistoryInitialization string `json:"history_initialization"`
	MetadataHydration     string `json:"metadata_hydration"`
	ProjectRegistry       string `json:"project_registry"`
	SettingsCache         string `json:"settings_cache"`
}

func loadCLIConfig(configPath string) (*config.Config, error) {
	if strings.TrimSpace(configPath) != "" {
		return config.Load(configPath)
	}
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	return config.LoadLayered(wd)
}

func validateOutputStyle(v string) error {
	_, _, err := normalizeOutputStyle(v)
	return err
}

func normalizeOutputStyle(v string) (string, bool, error) {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "human", "compact":
		return strings.ToLower(strings.TrimSpace(v)), false, nil
	case "human-readable", "readable", "default":
		return "human", true, nil
	case "short", "brief":
		return "compact", true, nil
	default:
		return "", false, fmt.Errorf("invalid --output-style %q (expected: human|compact)", v)
	}
}

func validateOutputFormat(v string) error {
	_, _, err := normalizeOutputFormat(v)
	return err
}

func normalizeOutputFormat(v string) (string, bool, error) {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "text", "json":
		return strings.ToLower(strings.TrimSpace(v)), false, nil
	case "plain", "human":
		return "text", true, nil
	case "machine":
		return "json", true, nil
	default:
		return "", false, fmt.Errorf("invalid --output-format %q (expected: text|json)", v)
	}
}

func validateTransportMode(v string) error {
	_, _, err := normalizeTransportMode(v)
	return err
}

func normalizeTransportMode(v string) (string, bool, error) {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "local", "remote":
		return strings.ToLower(strings.TrimSpace(v)), false, nil
	case "offline":
		return "local", true, nil
	case "tcp", "network":
		return "remote", true, nil
	default:
		return "", false, fmt.Errorf("invalid --transport %q (expected: local|remote)", v)
	}
}

func startupRuntimeState(providerName, modelName string, printMode bool, outStyle, outFormat, transport string) (commands.RuntimeState, error) {
	runtimeState, _, err := startupRuntimeStateWithDiagnostics(providerName, modelName, printMode, outStyle, outFormat, transport)
	if err != nil {
		return commands.RuntimeState{}, err
	}
	return runtimeState, nil
}

func startupRuntimeStateWithDiagnostics(providerName, modelName string, printMode bool, outStyle, outFormat, transport string) (commands.RuntimeState, startupDiagnostics, error) {
	diagnostics := startupDiagnostics{}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return commands.RuntimeState{}, diagnostics, fmt.Errorf("resolve home directory for runtime state: %w", err)
	}
	projectDir, err := os.Getwd()
	if err != nil {
		return commands.RuntimeState{}, diagnostics, fmt.Errorf("resolve project directory for runtime state: %w", err)
	}

	paths, err := state.ResolvePaths(homeDir, projectDir)
	if err != nil {
		diagnostics.StateInitialization = "error"
		return commands.RuntimeState{}, diagnostics, fmt.Errorf("resolve state paths: %w", err)
	}
	if err := state.EnsureState(paths); err != nil {
		diagnostics.StateInitialization = "error"
		return commands.RuntimeState{}, diagnostics, fmt.Errorf("initialize state directories: %w", err)
	}
	diagnostics.StateInitialization = "ok"

	sessionID := strings.TrimSpace(sessionIDGeneratorHook())
	if sessionID == "" {
		diagnostics.SessionInitialization = "error"
		return commands.RuntimeState{}, diagnostics, fmt.Errorf("create startup session: empty session id")
	}
	sessionStore, err := session.Create(homeDir, sessionID)
	if err != nil {
		diagnostics.SessionInitialization = "error"
		return commands.RuntimeState{}, diagnostics, fmt.Errorf("create startup session %q: %w", sessionID, err)
	}
	diagnostics.SessionInitialization = "ok"

	historyStore, err := history.NewJSONLStore(homeDir)
	if err != nil {
		diagnostics.HistoryInitialization = "error"
		return commands.RuntimeState{}, diagnostics, fmt.Errorf("create history store: %w", err)
	}
	if err := historyStore.AppendSessionStartForProject(sessionID, projectDir); err != nil {
		diagnostics.HistoryInitialization = "error"
		return commands.RuntimeState{}, diagnostics, fmt.Errorf("append startup history event: %w", err)
	}
	diagnostics.HistoryInitialization = "ok"

	authState, err := state.ReadAuthState(paths.AuthStateFile)
	if err != nil {
		diagnostics.MetadataHydration = "error"
		return commands.RuntimeState{}, diagnostics, fmt.Errorf("read auth state: %w", err)
	}

	if err := persistStartupSessionMetadata(paths.SessionMetadataFile, state.SessionMetadata{
		SessionID:   sessionID,
		SessionPath: sessionStore.Path(),
		ProjectPath: projectDir,
	}); err != nil {
		diagnostics.MetadataHydration = "error"
		return commands.RuntimeState{}, diagnostics, fmt.Errorf("persist startup session metadata: %w", err)
	}

	projectRegistry, err := state.RecordProjectOpen(paths.ProjectsStateFile, state.ProjectOpenRecord{
		ProjectPath: projectDir,
		SessionID:   sessionID,
		SessionPath: sessionStore.Path(),
		OpenedAt:    time.Now().UTC(),
	})
	if err != nil {
		diagnostics.ProjectRegistry = "error"
		return commands.RuntimeState{}, diagnostics, fmt.Errorf("record project startup metadata: %w", err)
	}
	diagnostics.ProjectRegistry = "ok"

	settingsCacheMeta, err := state.ReadSettingsCacheMetadata(paths.SettingsCacheFile)
	if err != nil {
		diagnostics.SettingsCache = "error"
		return commands.RuntimeState{}, diagnostics, fmt.Errorf("read settings cache metadata: %w", err)
	}
	settingsCacheMeta, err = applyStartupSettingsCacheFreshness(paths.SettingsCacheFile, projectDir, settingsCacheMeta)
	if err != nil {
		diagnostics.SettingsCache = "error"
		return commands.RuntimeState{}, diagnostics, fmt.Errorf("refresh settings cache metadata: %w", err)
	}
	diagnostics.SettingsCache = "ok"

	recent, err := historyStore.RecentSessions(100)
	if err != nil {
		diagnostics.MetadataHydration = "error"
		return commands.RuntimeState{}, diagnostics, fmt.Errorf("load recent history summaries: %w", err)
	}

	timestamped, err := historyStore.GetTimestampedHistory(history.HistoryQuery{Project: projectDir, CurrentSessionID: sessionID, MaxItems: 100})
	if err != nil {
		diagnostics.MetadataHydration = "error"
		return commands.RuntimeState{}, diagnostics, fmt.Errorf("load project timestamped history: %w", err)
	}

	recent, err = hydrateSessionSummaries(paths.SessionMetadataFile, projectDir, recent, timestamped, projectRegistry)
	if err != nil {
		diagnostics.MetadataHydration = "error"
		return commands.RuntimeState{}, diagnostics, fmt.Errorf("hydrate startup summaries: %w", err)
	}
	diagnostics.MetadataHydration = "ok"

	configValues := map[string]string{
		"settings.output-style":                    strings.TrimSpace(outStyle),
		"settings.output-format":                   strings.TrimSpace(outFormat),
		"settings.transport":                       strings.TrimSpace(transport),
		"status.project_registry.count":            strconv.Itoa(len(projectRegistry.Projects)),
		"status.project_registry.last_opened":      strings.TrimSpace(projectRegistry.LastOpenedProject),
		"status.project_registry.last_session_id":  strings.TrimSpace(projectRegistry.LastSessionID),
		"status.settings_cache.fresh":              strconv.FormatBool(settingsCacheMeta.Fresh),
		"status.settings_cache.scope":              strings.TrimSpace(settingsCacheMeta.LastScope),
		"status.settings_cache.hits":               strconv.Itoa(settingsCacheMeta.CacheHitCount),
		"status.settings_cache.misses":             strconv.Itoa(settingsCacheMeta.CacheMissCount),
		"status.settings_cache.last_invalidate":    strings.TrimSpace(settingsCacheMeta.LastInvalidateCause),
		"status.settings_cache.last_load_at":       unixNanoToRFC3339(settingsCacheMeta.LastLoadUnixNano),
		"status.settings_cache.last_hit_at":        unixNanoToRFC3339(settingsCacheMeta.LastHitUnixNano),
		"status.settings_cache.last_invalidate_at": unixNanoToRFC3339(settingsCacheMeta.LastInvalidateAt),
	}
	if meta, ok := projectRegistry.Projects[strings.TrimSpace(projectDir)]; ok {
		configValues["status.project_registry.last_opened_at"] = unixToRFC3339(meta.LastOpenedAt)
		configValues["status.project_registry.last_open_count"] = strconv.Itoa(meta.OpenCount)
		configValues["status.project_registry.last_session_title"] = strings.TrimSpace(meta.LastSessionTitle)
		configValues["status.project_registry.last_session_path"] = strings.TrimSpace(meta.LastSessionPath)
		configValues["status.project_registry.last_session_summary"] = strings.TrimSpace(meta.LastSessionSummary)
	}

	runtimeState := commands.RuntimeState{
		LoggedIn:       authState.LoggedIn,
		AuthProvider:   strings.ToLower(strings.TrimSpace(authState.Provider)),
		AuthAccount:    authState.Account,
		ProviderReady:  authState.ProviderReady,
		LoginCount:     authState.LoginCount,
		LogoutCount:    authState.LogoutCount,
		PermissionMode: permissions.ModeDefault,
		PrintMode:      printMode,
		OutputStyle:    outStyle,
		OutputFormat:   outFormat,
		TransportMode:  transport,
		SessionID:      sessionID,
		SessionPath:    sessionStore.Path(),
		ProjectPaths:   projectsFromRegistry(projectRegistry),
		HistoryEntries: historySummariesToRuntimeEntries(recent),
		ConfigValues:   configValues,
	}
	if err := commands.ApplyProviderModelSelection(&runtimeState, commands.ProviderModelSelection{ProviderName: providerName, ModelName: modelName}); err != nil {
		diagnostics.MetadataHydration = "error"
		return commands.RuntimeState{}, diagnostics, fmt.Errorf("hydrate startup provider/model state: %w", err)
	}
	if strings.EqualFold(strings.TrimSpace(authState.Provider), strings.TrimSpace(runtimeState.ProviderName)) {
		runtimeState.ProviderReady = authState.ProviderReady
	}
	commands.HydrateRuntimeSelection(&runtimeState)
	return runtimeState, diagnostics, nil
}

func historySummariesToRuntimeEntries(summaries []history.SessionSummary) []commands.HistoryEntry {
	if len(summaries) == 0 {
		return nil
	}
	out := make([]commands.HistoryEntry, 0, len(summaries))
	for _, summary := range summaries {
		createdAt := ""
		if !summary.FirstEvent.IsZero() {
			createdAt = summary.FirstEvent.UTC().Format(time.RFC3339)
		}
		title := strings.TrimSpace(summary.Title)
		if title == "" {
			title = strings.TrimSpace(summary.SessionID)
		}
		if title == "" {
			title = "session"
		}
		summaryText := strings.TrimSpace(summary.Summary)
		if summaryText == "" {
			summaryText = fmt.Sprintf("events=%d messages=%d", summary.EventCount, summary.MessageCount)
		}
		out = append(out, commands.HistoryEntry{
			ID:        strings.TrimSpace(summary.SessionID),
			Path:      strings.TrimSpace(summary.Path),
			CreatedAt: createdAt,
			Turns:     summary.MessageCount,
			Title:     title,
			Summary:   summaryText,
		})
	}
	return out
}

func hydrateSessionSummaries(metadataPath, projectDir string, summaries []history.SessionSummary, timestamped []history.TimestampedHistoryEntry, projectRegistry state.ProjectsRegistry) ([]history.SessionSummary, error) {
	sessionRegistry, err := state.ReadSessionMetadataRegistry(metadataPath)
	if err != nil {
		return nil, err
	}
	if len(summaries) == 0 {
		return summaries, nil
	}

	timestampBySession := make(map[string]history.TimestampedHistoryEntry)
	for _, entry := range timestamped {
		sessionID := strings.TrimSpace(entry.Event.SessionID)
		if sessionID == "" {
			continue
		}
		if _, exists := timestampBySession[sessionID]; !exists {
			timestampBySession[sessionID] = entry
		}
	}

	projectMeta := projectRegistry.Projects[strings.TrimSpace(projectDir)]

	out := make([]history.SessionSummary, len(summaries))
	copy(out, summaries)
	for i := range out {
		summary := out[i]
		sessionID := strings.TrimSpace(summary.SessionID)
		meta, ok := sessionRegistry.Sessions[sessionID]
		if !ok {
			meta = state.SessionMetadata{}
		}
		if strings.TrimSpace(meta.ProjectPath) != "" && strings.TrimSpace(projectDir) != "" && strings.TrimSpace(meta.ProjectPath) != strings.TrimSpace(projectDir) {
			continue
		}
		if strings.TrimSpace(meta.SessionPath) != "" {
			summary.Path = strings.TrimSpace(meta.SessionPath)
		} else if projectMeta.LastSessionID == sessionID && strings.TrimSpace(projectMeta.LastSessionPath) != "" {
			summary.Path = strings.TrimSpace(projectMeta.LastSessionPath)
		}
		if strings.TrimSpace(meta.Title) != "" {
			summary.Title = strings.TrimSpace(meta.Title)
		} else if projectMeta.LastSessionID == sessionID && strings.TrimSpace(projectMeta.LastSessionTitle) != "" {
			summary.Title = strings.TrimSpace(projectMeta.LastSessionTitle)
		} else if ts, ok := timestampBySession[sessionID]; ok {
			summary.Title = strings.TrimSpace(ts.Display)
		}
		if strings.TrimSpace(meta.Summary) != "" {
			summary.Summary = strings.TrimSpace(meta.Summary)
		} else if projectMeta.LastSessionID == sessionID && strings.TrimSpace(projectMeta.LastSessionSummary) != "" {
			summary.Summary = strings.TrimSpace(projectMeta.LastSessionSummary)
		} else if ts, ok := timestampBySession[sessionID]; ok {
			summary.Summary = strings.TrimSpace(ts.Display)
		}
		out[i] = summary
	}
	return out, nil
}

func persistStartupSessionMetadata(metadataPath string, metadata state.SessionMetadata) error {
	registry, err := state.ReadSessionMetadataRegistry(metadataPath)
	if err != nil {
		return err
	}
	if registry.Sessions == nil {
		registry.Sessions = make(map[string]state.SessionMetadata)
	}
	sessionID := strings.TrimSpace(metadata.SessionID)
	if sessionID == "" {
		return nil
	}
	current := registry.Sessions[sessionID]
	if strings.TrimSpace(current.SessionID) == "" {
		current.SessionID = sessionID
	}
	if strings.TrimSpace(metadata.SessionPath) != "" {
		current.SessionPath = strings.TrimSpace(metadata.SessionPath)
	}
	if strings.TrimSpace(metadata.ProjectPath) != "" {
		current.ProjectPath = strings.TrimSpace(metadata.ProjectPath)
	}
	if strings.TrimSpace(metadata.Title) != "" {
		current.Title = strings.TrimSpace(metadata.Title)
	}
	if strings.TrimSpace(metadata.Summary) != "" {
		current.Summary = strings.TrimSpace(metadata.Summary)
	}
	current.UpdatedAt = time.Now().UTC().Unix()
	registry.Sessions[sessionID] = current
	return state.WriteSessionMetadataRegistry(metadataPath, registry)
}

func generateSessionID() string {
	return fmt.Sprintf("session-%d", time.Now().UTC().UnixNano())
}

func renderStartupInfo(cmd *cobra.Command, info startupInfo) error {
	if strings.EqualFold(info.OutputFormat, "json") {
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(info)
	}

	fields := map[string]string{
		"config_path":      info.ConfigPath,
		"provider":         info.Provider,
		"model":            info.Model,
		"debug":            fmt.Sprintf("%t", info.Debug),
		"print":            fmt.Sprintf("%t", info.PrintMode),
		"output_style":     info.OutputStyle,
		"output_format":    info.OutputFormat,
		"transport":        info.Transport,
		"startup_state":    info.Diagnostics.StateInitialization,
		"startup_session":  info.Diagnostics.SessionInitialization,
		"startup_history":  info.Diagnostics.HistoryInitialization,
		"startup_metadata": info.Diagnostics.MetadataHydration,
	}
	keys := make([]string, 0, len(fields))
	for k := range fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Fprintf(cmd.OutOrStdout(), "%s=%s\n", k, fields[k])
	}
	return nil
}

func runStartupMigrations() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("resolve migration ledger path: %w", err)
	}

	ledgerPath := filepath.Join(home, ".config", "alliecode", "migrations.jsonl")
	runner := migrations.NewRunner(migrations.NewLedger(ledgerPath))
	if _, err := runner.RunAll(startupMigrations); err != nil {
		return fmt.Errorf("run startup migrations with ledger %q: %w", ledgerPath, err)
	}

	return nil
}

func runSlashCommand(ctx context.Context, cmd *cobra.Command, input string, state commands.RuntimeState) error {
	normalized := strings.TrimSpace(input)
	if normalized == "" {
		return commands.ErrEmptyCommand
	}
	if !strings.HasPrefix(normalized, "/") {
		normalized = "/" + normalized
	}

	registry := commands.DefaultRegistry()
	res, err := registry.Dispatch(ctx, commands.Context{State: &state}, normalized)
	if err != nil {
		return err
	}
	if !res.Handled {
		return nil
	}
	commands.HydrateRuntimeSelection(&state)
	if strings.HasPrefix(strings.ToLower(normalized), "/rename") {
		if err := persistRenameMetadata(state); err != nil {
			return err
		}
	}
	if strings.TrimSpace(state.OutputFormat) == "json" {
		payload := struct {
			Handled bool           `json:"handled"`
			Message string         `json:"message"`
			State   map[string]any `json:"state"`
		}{
			Handled: true,
			Message: res.Message,
			State: map[string]any{
				"model":             state.Model,
				"model_ref":         state.ModelRef,
				"provider":          state.ProviderName,
				"provider_ready":    state.ProviderReady,
				"logged_in":         state.LoggedIn,
				"permission_mode":   state.PermissionMode,
				"compact_requested": state.CompactRequested,
				"resume_requested":  state.ResumeRequested,
				"active_branch":     state.ActiveBranch,
				"diff_entries":      len(state.DiffEntries),
			},
		}
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(payload)
	}
	if strings.TrimSpace(res.Message) != "" {
		message := res.Message
		if isStatusCommand(normalized) {
			message = augmentStatusReport(message, state)
		}
		fmt.Fprintln(cmd.OutOrStdout(), message)
	}
	return nil
}

func isStatusCommand(input string) bool {
	clean := strings.TrimSpace(strings.ToLower(input))
	if clean == "/status" {
		return true
	}
	return strings.HasPrefix(clean, "/status ")
}

func augmentStatusReport(message string, runtimeState commands.RuntimeState) string {
	if !strings.HasPrefix(message, "STATUS_REPORT") {
		return message
	}
	if len(runtimeState.ConfigValues) == 0 {
		return message
	}
	keys := make([]string, 0, len(runtimeState.ConfigValues))
	for key := range runtimeState.ConfigValues {
		if strings.HasPrefix(key, "status.") {
			keys = append(keys, key)
		}
	}
	if len(keys) == 0 {
		return message
	}
	sort.Strings(keys)
	lines := []string{message}
	for _, key := range keys {
		value := normalizeStatusValue(runtimeState.ConfigValues[key])
		lines = append(lines, fmt.Sprintf("%s=%s", strings.TrimPrefix(key, "status."), value))
	}
	return strings.Join(lines, "\n")
}

func normalizeStatusValue(v string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return "-"
	}
	v = strings.ReplaceAll(v, "\n", " ")
	v = strings.ReplaceAll(v, "\r", " ")
	v = strings.Join(strings.Fields(v), " ")
	if v == "" {
		return "-"
	}
	return v
}

func projectsFromRegistry(registry state.ProjectsRegistry) []string {
	out := make([]string, 0, len(registry.Projects))
	for projectPath := range registry.Projects {
		projectPath = strings.TrimSpace(projectPath)
		if projectPath == "" {
			continue
		}
		out = append(out, projectPath)
	}
	sort.Strings(out)
	return out
}

func applyStartupSettingsCacheFreshness(cachePath, projectDir string, meta state.SettingsCacheMetadata) (state.SettingsCacheMetadata, error) {
	projectDir = strings.TrimSpace(projectDir)
	lastProject := strings.TrimSpace(meta.LastProjectDir)
	now := time.Now().UTC().UnixNano()
	if lastProject != "" && projectDir != "" && lastProject != projectDir {
		meta.Fresh = false
		meta.LastInvalidateCause = "project_switch"
		meta.LastInvalidateAt = now
	}
	if meta.LastLoadUnixNano > 0 {
		age := now - meta.LastLoadUnixNano
		if age > int64(5*time.Second) {
			meta.Fresh = false
			if strings.TrimSpace(meta.LastInvalidateCause) == "" {
				meta.LastInvalidateCause = "freshness_window"
			}
			if meta.LastInvalidateAt == 0 {
				meta.LastInvalidateAt = now
			}
		}
	}
	meta.LastProjectDir = projectDir
	meta.LastScope = "layered"
	meta.CacheFilePath = strings.TrimSpace(cachePath)
	if err := state.WriteSettingsCacheMetadata(cachePath, meta); err != nil {
		return state.SettingsCacheMetadata{}, err
	}
	return state.ReadSettingsCacheMetadata(cachePath)
}

func unixToRFC3339(v int64) string {
	if v <= 0 {
		return ""
	}
	return time.Unix(v, 0).UTC().Format(time.RFC3339)
}

func unixNanoToRFC3339(v int64) string {
	if v <= 0 {
		return ""
	}
	return time.Unix(0, v).UTC().Format(time.RFC3339)
}

func persistRenameMetadata(runtimeState commands.RuntimeState) error {
	sessionID := strings.TrimSpace(runtimeState.SessionID)
	if sessionID == "" {
		return nil
	}
	title := strings.TrimSpace(runtimeState.SessionTitle)
	if title == "" {
		return nil
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("resolve home directory for rename persistence: %w", err)
	}
	projectDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("resolve project directory for rename persistence: %w", err)
	}
	paths, err := state.ResolvePaths(homeDir, projectDir)
	if err != nil {
		return fmt.Errorf("resolve state paths for rename persistence: %w", err)
	}
	if err := state.EnsureState(paths); err != nil {
		return fmt.Errorf("ensure state for rename persistence: %w", err)
	}
	return persistStartupSessionMetadata(paths.SessionMetadataFile, state.SessionMetadata{
		SessionID:   sessionID,
		SessionPath: strings.TrimSpace(runtimeState.SessionPath),
		ProjectPath: projectDir,
		Title:       title,
	})
}
