package commands

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/alliecatowo/alliecode/internal/agent"
	"github.com/alliecatowo/alliecode/internal/mcp"
	"github.com/alliecatowo/alliecode/internal/permissions"
	"github.com/alliecatowo/alliecode/internal/remote"
	"github.com/alliecatowo/alliecode/internal/types"
)

// RuntimeState is mutable slash-command state owned by the app/session.
type RuntimeState struct {
	Model                   string
	ModelRef                string
	LoggedIn                bool
	AuthProvider            string
	AuthAccount             string
	LoginCount              int
	LogoutCount             int
	ConfigValues            map[string]string
	CompactRequested        bool
	CompactMode             string
	CompactCount            int
	LastCompactTarget       string
	BuddyHatched            bool
	BuddyPetCount           int
	BuddyMuted              bool
	PermissionMode          permissions.Mode
	PermissionRules         []string
	PermissionDenials       []PermissionDenial
	ResumeRequested         bool
	Agent                   *agent.Agent
	MCPConnections          map[string]bool
	MCPDoctorCount          int
	MCPRepairCount          int
	MCPLastQuickFix         string
	VimEnabled              bool
	VoiceEnabled            bool
	PrintMode               bool
	OutputStyle             string
	OutputFormat            string
	TransportMode           string
	ConfigPath              string
	ProviderName            string
	ProviderReady           bool
	CostInputTokens         int64
	CostOutputTokens        int64
	CostCacheRead           int64
	CostCacheWrite          int64
	WorkspaceDirs           []string
	HistoryEntries          []HistoryEntry
	HistoryViews            int
	HistoryLastFilterType   string
	HistoryLastFilterValue  string
	HistoryLastLimit        int
	ClearCount              int
	ClearContextCount       int
	ClearDiffCount          int
	LastClearScope          string
	LastResumeTarget        string
	ResumeCount             int
	CopyCount               int
	LastCopiedText          string
	Branches                []string
	ActiveBranch            string
	BranchCount             int
	BranchSwitchCount       int
	DiffMode                string
	DiffCount               int
	DiffClearCount          int
	DiffEntries             []DiffEntry
	MemoryEntries           []string
	MemoryWrites            int
	MemoryRemoves           int
	MemoryClears            int
	LastMemory              string
	PrivacyTelemetry        bool
	PrivacyTraining         bool
	PrivacyUpdates          int
	HookPreEnabled          bool
	HookPostEnabled         bool
	SandboxMode             string
	SandboxWorkspaceLocked  bool
	SandboxExcludedCommands []string
	SandboxSettingsPath     string
	ConfigRepairCount       int
	ConfigDoctorCount       int
	ConfigLastRepairProfile string
	Tasks                   []string
	TasksCompleted          int
	ContextFiles            []string
	FilesAdds               int
	FilesRemoves            int
	FilesClears             int
	LastFileAction          string
	LastFilePath            string
	ThemeSetCount           int
	OutputStyleSetCount     int
	StatuslineCount         int
	LastStatusline          string
	StatuslineLastRenderMS  int
	StatuslineToolCalls     int
	StatuslinePermissions   int
	StatuslineTransitions   int
	StatusViewCount         int
	StatusDiagnosticsCount  int
	UpgradeRequested        bool
	UpgradeCount            int
	LastUpgradePlan         string
	TerminalConfigured      bool
	TerminalSetupCount      int
	TerminalProfile         string
	TerminalDetectedProfile string
	TerminalDetectSource    string
	TerminalSetupHints      []string
	IDEEditor               string
	IDEDetectedEditor       string
	IDEDetectSource         string
	IDEOpenCount            int
	IDEConfigCount          int
	ReleaseNotesSeen        int
	LastReleaseVersion      string
	GitHubAppInstalls       int
	LastGitHubRepo          string
	SlackAppInstalls        int
	FeedbackCount           int
	LastFeedback            string
	DoctorRunCount          int
	DoctorFixCount          int
	DoctorLastStatus        string
	DoctorLastWarnSections  int
	DoctorLastQuickFixes    int
	KeybindingsEnabled      bool
	KeybindingsPath         string
	KeybindingsExists       bool
	KeybindingsOpens        int
	ExitRequested           bool
	ExitCount               int
	LastExitCommand         string
	PlanModeEnabled         bool
	PlanEnableCount         int
	PlanOpenCount           int
	LastPlanDescription     string
	ReviewCount             int
	LastReviewTarget        string
	RemoteSessionURL        string
	SessionHosted           bool
	SessionHostAddr         string
	SessionConnected        bool
	SessionConnectedAddr    string
	SessionTokenPrefix      string
	SessionTokenSource      string
	SessionHostCount        int
	SessionConnectCount     int
	SessionDisconnectCount  int
	SessionMode             string
	SessionToken            string
	SessionTransport        remote.Transport
	SessionTCPTransport     *remote.TCPTransport
	SessionManager          *remote.SessionManager
	SessionID               string
	SessionPath             string
	SessionViewCount        int
	SessionDiagnosticsCount int
	SessionRepairCount      int
	SessionLastRepairAction string
	SessionLastErrorClass   string
	ProjectPaths            []string
	Skills                  []string
	SkillsSources           map[string]string
	SkillsOrigins           map[string]string
	SkillsEnabled           map[string]bool
	SkillsConflictCount     int
	SkillsViewCount         int
	SkillsSyncCount         int
	SkillsDoctorCount       int
	SkillsLastSyncSource    string
	RewindCount             int
	LastRewindTarget        string
	CurrentTag              string
	TagUpdates              int
	RemoteEnvironment       string
	RemoteEnvUpdates        int
	SecurityReviewCount     int
	LastSecurityTarget      string
	AdvisorModel            string
	AdvisorUpdates          int
	BtwUseCount             int
	LastBtwQuestion         string
	ChromeDefault           bool
	ChromeExtension         bool
	ChromeConnected         bool
	ChromeActions           int
	SessionColor            string
	ColorSetCount           int
	DesktopHandoffCount     int
	LastDesktopTarget       string
	MobilePlatform          string
	MobileQRCount           int
	FastMode                bool
	FastToggleCount         int
	EffortLevel             string
	EffortSetCount          int
	PluginsInstalled        []string
	PluginsEnabled          []string
	PluginMarketplaces      []string
	PluginReloadPending     bool
	PluginMutations         int
	PluginReloadCount       int
	PluginDoctorCount       int
	ExportsCount            int
	LastExportPath          string
	LastExportFormat        string
	ExtraUsageEnabled       bool
	ExtraUsageRequests      int
	RateLimitPrompts        int
	LastRateLimitAction     string
	PRCommentsFetches       int
	LastPRCommentsRef       string
	WebSetupConnected       bool
	WebSetupCount           int
	BridgeKickCount         int
	BridgeKickLastAction    string
	BridgeKickLastCode      int
	BriefOnly               bool
	BriefToggleCount        int
	CommitCount             int
	LastCommitMessage       string
	CommitPushPRCount       int
	LastPRURL               string
	InitVerifiersCount      int
	LastVerifierName        string
	InsightsCount           int
	LastInsightsScope       string
	PassesVisitCount        int
	PassesRemaining         int
	RenameCount             int
	SessionTitle            string
	StickersCount           int
	ThinkbackCount          int
	ThinkbackLastAction     string
	ThinkbackPlayCount      int
	TeleportCount           int
	TeleportLastTarget      string
	SummaryCount            int
	ResetLimitsCount        int
	EnvSetCount             int
	IssueProvider           string
	IssueNextID             int
	IssueLastAction         string
	Issues                  []IssueRecord
	WorkflowRuns            []WorkflowRun
	WorkflowLastAction      string
	ProactiveEnabled        bool
	ProactiveLastAction     string
	ProactiveRules          []string
	AssistantMode           string
	AssistantSessionID      string
	AssistantLastAction     string
	ShareLinks              []ShareLink
	ShareLastAction         string
	OAuthRefreshStates      map[string]OAuthRefreshState
	OAuthRefreshLastAction  string
	BridgeEnabled           bool
	BridgeTransitions       int
	AntTraceEnabled         bool
	AntTraceMarks           []string
	AntTraceCount           int
	AutofixPRCount          int
	AutofixPRLastAction     string
	AutofixPRLastRef        string
	BackfillRuns            int
	BackfillLastCount       int
	BackfillLastAction      string
	CacheBreakCount         int
	CacheBreakLastScope     string
	CacheBreakScopes        map[string]int
	BughunterCount          int
	BughunterLastScope      string
	ContextVizCount         int
	ContextVizLastAction    string
	DebugToolCallCount      int
	DebugToolCallLastTool   string
	GoodClaudeEnabled       bool
	GoodClaudeAskCount      int
	HeapdumpCount           int
	HeapdumpLastReason      string
	InstallCount            int
	InstallLastAction       string
	MockLimitsEnabled       bool
	MockLimitsValue         int
	MockLimitsChanges       int
	OnboardingCompleted     bool
	OnboardingRuns          int
	PerfIssueCount          int
	PerfIssueLastTitle      string
	UltraplanCount          int
	UltraplanLastAction     string
	UltraplanLastTarget     string
	Runtime                 types.AgentRuntimeSnapshot
}

// PermissionDenial represents a denied permission request record.
type PermissionDenial struct {
	ID      string
	Command string
	Reason  string
}

// HistoryEntry represents a local session history record.
type HistoryEntry struct {
	ID        string
	Path      string
	CreatedAt string
	Model     string
	Turns     int
	Title     string
	Summary   string
}

// DiffEntry represents a deterministic diff row for slash-command output.
type DiffEntry struct {
	Path      string
	Added     int
	Removed   int
	Modified  int
	Staged    bool
	Untracked bool
}

// IssueRecord stores provider-agnostic issue lifecycle state.
type IssueRecord struct {
	ID       string
	Title    string
	Status   string
	Provider string
	Assignee string
	Labels   []string
}

// WorkflowRun stores provider-agnostic workflow execution state.
type WorkflowRun struct {
	Name     string
	Status   string
	Provider string
	LastRun  string
}

// ShareLink stores published share references.
type ShareLink struct {
	ID         string
	Scope      string
	Visibility string
	URL        string
	Revoked    bool
}

// OAuthRefreshState stores provider refresh token lifecycle metadata.
type OAuthRefreshState struct {
	Provider    string
	Refreshed   bool
	TokenPrefix string
	ExpiresIn   int
	Error       string
}

// Context is passed to command handlers.
type Context struct {
	State      *RuntimeState
	MCPManager *mcp.Manager
}

// Result is the output returned from executing a slash command.
type Result struct {
	Message       string
	Handled       bool
	RenderIntents []types.RenderIntent
}

// Command is implemented by slash command handlers.
type Command interface {
	Name() string
	Aliases() []string
	Description() string
	Usage() string
	Execute(ctx context.Context, cmdCtx Context, inv Invocation) (Result, error)
}

// Registry stores and resolves slash commands.
type Registry struct {
	commands map[string]Command
	names    []string
}

// Suggestion is a command candidate for slash autocomplete UIs.
type Suggestion struct {
	Name         string
	Description  string
	Usage        string
	Aliases      []string
	Category     string
	Group        string
	PaletteGroup string
	ArgumentHint string
	HelpHint     string
	Shortcuts    []string
	Keywords     []string
	Examples     []string
	Contexts     []string
	Diagnostics  []string
	MatchReason  string
}

// NewRegistry creates an empty command registry.
func NewRegistry() *Registry {
	return &Registry{commands: make(map[string]Command)}
}

// Register adds command and aliases to the registry.
func (r *Registry) Register(cmd Command) error {
	if cmd == nil {
		return fmt.Errorf("command cannot be nil")
	}

	name := strings.ToLower(strings.TrimSpace(cmd.Name()))
	if name == "" {
		return fmt.Errorf("command name cannot be empty")
	}
	if _, exists := r.commands[name]; exists {
		return fmt.Errorf("command already registered: %s", name)
	}

	r.commands[name] = cmd
	r.names = append(r.names, name)

	for _, alias := range cmd.Aliases() {
		a := strings.ToLower(strings.TrimSpace(alias))
		if a == "" {
			continue
		}
		if _, exists := r.commands[a]; exists {
			return fmt.Errorf("alias already registered: %s", a)
		}
		r.commands[a] = cmd
	}

	sort.Strings(r.names)
	return nil
}

// Lookup resolves a command by name or alias.
func (r *Registry) Lookup(name string) (Command, bool) {
	cmd, ok := r.commands[strings.ToLower(strings.TrimSpace(name))]
	return cmd, ok
}

// CanonicalCommands returns unique commands ordered by canonical name.
func (r *Registry) CanonicalCommands() []Command {
	out := make([]Command, 0, len(r.names))
	for _, n := range r.names {
		if cmd, ok := r.commands[n]; ok {
			out = append(out, cmd)
		}
	}
	return out
}

// Suggestions returns slash-command suggestions ranked by query relevance.
// Query should be the command token without a leading slash.
func (r *Registry) Suggestions(query string) []Suggestion {
	q := strings.ToLower(strings.TrimSpace(query))
	type ranked struct {
		name         string
		description  string
		usage        string
		aliases      []string
		category     string
		argumentHint string
		keywords     []string
		examples     []string
		reason       string
		rank         int
		group        string
		helpHint     string
		shortcuts    []string
		paletteGroup string
		contexts     []string
		diagnostics  []string
	}

	rankedRows := make([]ranked, 0, len(r.names))
	for _, cmd := range r.CanonicalCommands() {
		name := strings.TrimSpace(cmd.Name())
		if name == "" {
			continue
		}
		description := strings.TrimSpace(cmd.Description())
		meta := commandMetadataForName(name)
		rank, reason, ok := commandSuggestionRank(cmd, q, meta)
		if !ok {
			continue
		}
		rankedRows = append(rankedRows, ranked{
			name:         name,
			description:  description,
			usage:        strings.TrimSpace(cmd.Usage()),
			aliases:      append([]string(nil), cmd.Aliases()...),
			category:     strings.TrimSpace(meta.Category),
			argumentHint: strings.TrimSpace(meta.ArgumentHint),
			keywords:     append([]string(nil), meta.Keywords...),
			examples:     append([]string(nil), meta.Examples...),
			reason:       reason,
			rank:         rank,
			group:        strings.TrimSpace(meta.Group),
			helpHint:     strings.TrimSpace(meta.HelpHint),
			shortcuts:    append([]string(nil), meta.Shortcuts...),
			paletteGroup: strings.TrimSpace(meta.PaletteGroup),
			contexts:     append([]string(nil), meta.Contexts...),
			diagnostics:  append([]string(nil), meta.Diagnostics...),
		})
	}

	sort.Slice(rankedRows, func(i, j int) bool {
		if rankedRows[i].rank != rankedRows[j].rank {
			return rankedRows[i].rank < rankedRows[j].rank
		}
		if len(rankedRows[i].name) != len(rankedRows[j].name) {
			return len(rankedRows[i].name) < len(rankedRows[j].name)
		}
		return rankedRows[i].name < rankedRows[j].name
	})

	out := make([]Suggestion, 0, len(rankedRows))
	for _, row := range rankedRows {
		out = append(out, Suggestion{
			Name:         row.name,
			Description:  row.description,
			Usage:        row.usage,
			Aliases:      row.aliases,
			Category:     row.category,
			Group:        row.group,
			PaletteGroup: row.paletteGroup,
			ArgumentHint: row.argumentHint,
			HelpHint:     row.helpHint,
			Shortcuts:    row.shortcuts,
			Keywords:     row.keywords,
			Examples:     row.examples,
			Contexts:     row.contexts,
			Diagnostics:  row.diagnostics,
			MatchReason:  row.reason,
		})
	}
	return out
}

func commandSuggestionRank(cmd Command, query string, meta CommandMetadata) (int, string, bool) {
	if strings.TrimSpace(query) == "" {
		return 0, "browse", true
	}
	name := strings.ToLower(strings.TrimSpace(cmd.Name()))
	desc := strings.ToLower(strings.TrimSpace(cmd.Description()))
	usage := strings.ToLower(strings.TrimSpace(cmd.Usage()))

	if strings.HasPrefix(name, query) {
		return 0, "name-prefix", true
	}
	for _, alias := range cmd.Aliases() {
		a := strings.ToLower(strings.TrimSpace(alias))
		if a == "" {
			continue
		}
		if strings.HasPrefix(a, query) {
			return 1, "alias-prefix", true
		}
	}
	if strings.Contains(name, query) {
		return 2, "name-contains", true
	}
	for _, alias := range cmd.Aliases() {
		a := strings.ToLower(strings.TrimSpace(alias))
		if a == "" {
			continue
		}
		if strings.Contains(a, query) {
			return 3, "alias-contains", true
		}
	}
	if strings.Contains(desc, query) {
		return 4, "description", true
	}
	if strings.Contains(usage, query) {
		return 5, "usage", true
	}
	for _, keyword := range meta.Keywords {
		k := strings.ToLower(strings.TrimSpace(keyword))
		if k == "" {
			continue
		}
		if strings.HasPrefix(k, query) {
			return 6, "keyword", true
		}
		if strings.Contains(k, query) {
			return 7, "keyword", true
		}
	}
	for _, example := range meta.Examples {
		ex := strings.ToLower(strings.TrimSpace(example))
		if ex == "" {
			continue
		}
		if strings.Contains(ex, query) {
			return 8, "example", true
		}
	}
	for _, context := range meta.Contexts {
		c := strings.ToLower(strings.TrimSpace(context))
		if c == "" {
			continue
		}
		if strings.Contains(c, query) {
			return 9, "context", true
		}
	}
	for _, diagnostic := range meta.Diagnostics {
		d := strings.ToLower(strings.TrimSpace(diagnostic))
		if d == "" {
			continue
		}
		if strings.Contains(d, query) {
			return 10, "diagnostic", true
		}
	}
	return 0, "", false
}

// Dispatch parses and executes a slash command.
func (r *Registry) Dispatch(ctx context.Context, cmdCtx Context, input string) (Result, error) {
	inv, err := Parse(input)
	if err != nil {
		return Result{}, err
	}

	cmd, ok := r.Lookup(inv.Name)
	if !ok {
		return Result{}, fmt.Errorf("unknown slash command: /%s", inv.Name)
	}

	return cmd.Execute(ctx, cmdCtx, inv)
}

// DefaultRegistry returns parity-focused built-in slash commands.
func DefaultRegistry() *Registry {
	r := NewRegistry()
	_ = r.Register(NewHelpCommand(r))
	_ = r.Register(NewModelCommand())
	_ = r.Register(NewProviderCommand())
	_ = r.Register(NewCompactCommand())
	_ = r.Register(NewBranchCommand())
	_ = r.Register(NewDiffCommand())
	_ = r.Register(NewCostCommand())
	_ = r.Register(NewDoctorCommand())
	_ = r.Register(NewConfigCommand())
	_ = r.Register(NewInitCommand())
	_ = r.Register(NewCopyCommand())
	_ = r.Register(NewVersionCommand())
	_ = r.Register(NewUsageCommand())
	_ = r.Register(NewContextCommand())
	_ = r.Register(NewExitCommand())
	_ = r.Register(NewPlanCommand())
	_ = r.Register(NewReviewCommand())
	_ = r.Register(NewSessionCommand())
	_ = r.Register(NewSkillsCommand())
	_ = r.Register(NewRewindCommand())
	_ = r.Register(NewTagCommand())
	_ = r.Register(NewRemoteEnvCommand())
	_ = r.Register(NewSecurityReviewCommand())
	_ = r.Register(NewPermissionsCommand())
	_ = r.Register(NewResumeCommand())
	_ = r.Register(NewAddDirCommand())
	_ = r.Register(NewAgentsCommand())
	_ = r.Register(NewClearCommand())
	_ = r.Register(NewHistoryCommand())
	_ = r.Register(NewMCPCommand())
	_ = r.Register(NewVimCommand())
	_ = r.Register(NewVoiceCommand())
	_ = r.Register(NewLoginCommand())
	_ = r.Register(NewLogoutCommand())
	_ = r.Register(NewBuddyCommand())
	_ = r.Register(NewThemeCommand())
	_ = r.Register(NewFilesCommand())
	_ = r.Register(NewOutputStyleCommand())
	_ = r.Register(NewStatuslineCommand())
	_ = r.Register(NewIDECommand())
	_ = r.Register(NewKeybindingsCommand())
	_ = r.Register(NewStatusCommand())
	_ = r.Register(NewStatsCommand())
	_ = r.Register(NewMemoryCommand())
	_ = r.Register(NewPrivacySettingsCommand())
	_ = r.Register(NewUpgradeCommand())
	_ = r.Register(NewTerminalSetupCommand())
	_ = r.Register(NewReleaseNotesCommand())
	_ = r.Register(NewAdvisorCommand())
	_ = r.Register(NewBtwCommand())
	_ = r.Register(NewChromeCommand())
	_ = r.Register(NewColorCommand())
	_ = r.Register(NewDesktopCommand())
	_ = r.Register(NewMobileCommand())
	_ = r.Register(NewFastCommand())
	_ = r.Register(NewEffortCommand())
	_ = r.Register(NewPluginCommand())
	_ = r.Register(NewReloadPluginsCommand())
	_ = r.Register(NewExportCommand())
	_ = r.Register(NewExtraUsageCommand())
	_ = r.Register(NewRateLimitOptionsCommand())
	_ = r.Register(NewPRCommentsCommand())
	_ = r.Register(NewWebSetupCommand())
	_ = r.Register(NewInstallGitHubAppCommand())
	_ = r.Register(NewInstallSlackAppCommand())
	_ = r.Register(NewFeedbackCommand())
	_ = r.Register(NewHooksCommand())
	_ = r.Register(NewSandboxCommand())
	_ = r.Register(NewTasksCommand())
	_ = r.Register(NewBridgeKickCommand())
	_ = r.Register(NewBriefCommand())
	_ = r.Register(NewCommitCommand())
	_ = r.Register(NewCommitPushPRCommand())
	_ = r.Register(NewInitVerifiersCommand())
	_ = r.Register(NewInsightsCommand())
	_ = r.Register(NewPassesCommand())
	_ = r.Register(NewRenameCommand())
	_ = r.Register(NewStickersCommand())
	_ = r.Register(NewThinkbackCommand())
	_ = r.Register(NewThinkbackPlayCommand())
	_ = r.Register(NewTeleportCommand())
	_ = r.Register(NewSummaryCommand())
	_ = r.Register(NewResetLimitsCommand())
	_ = r.Register(NewEnvCommand())
	_ = r.Register(NewIssueCommand())
	_ = r.Register(NewWorkflowsCommand())
	_ = r.Register(NewProactiveCommand())
	_ = r.Register(NewAssistantCommand())
	_ = r.Register(NewShareCommand())
	_ = r.Register(NewOAuthRefreshCommand())
	_ = r.Register(NewBridgeCommand())
	_ = r.Register(NewAntTraceCommand())
	_ = r.Register(NewAutofixPRCommand())
	_ = r.Register(NewBackfillSessionsCommand())
	_ = r.Register(NewBreakCacheCommand())
	_ = r.Register(NewBughunterCommand())
	_ = r.Register(NewCtxVizCommand())
	_ = r.Register(NewDebugToolCallCommand())
	_ = r.Register(NewGoodClaudeCommand())
	_ = r.Register(NewHeapdumpCommand())
	_ = r.Register(NewInstallCommand())
	_ = r.Register(NewMockLimitsCommand())
	_ = r.Register(NewOnboardingCommand())
	_ = r.Register(NewPerfIssueCommand())
	_ = r.Register(NewSandboxToggleCommand())
	_ = r.Register(NewRemoteSetupCommand())
	_ = r.Register(NewUltraplanCommand())
	return r
}
