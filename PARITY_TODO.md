# Parity Todo Ledger (TS -> Go)

This file tracks TypeScript-to-Go parity work across `src/`.
Mark items complete only when the acceptance criteria (AC) is met in Go.

## How parity is measured

- Functional parity: same user-facing behavior and side effects as TS for normal and error paths (AC: fixture-based golden scenarios pass against both implementations).
- UX parity: equivalent CLI/TUI affordances, prompts, state transitions, and key workflows (AC: scripted terminal snapshots and interaction recordings match expected deltas only).
- Test parity: each migrated subsystem has unit tests plus at least one integration/e2e path (AC: Go tests cover happy path, failure path, and cancellation/interrupt path).
- Operational parity: logs/metrics/hooks/session persistence are preserved (AC: no regression in observability fields and persisted transcript format).

## Swarm ownership lanes (no-overlap rules)

- Lane A - Core loop and providers: owns `src/query*`, `src/context*`, `src/services/api*`, `src/entrypoints/*` -> `internal/agent`, `internal/providers`, `internal/types`, `cmd/ac`.
- Lane B - Tools and permissions: owns `src/tools/*`, `src/Tool.ts`, `src/tools.ts`, `src/hooks/toolPermission/*` -> `internal/tools`, `internal/permissions`, `internal/mcp`.
- Lane C - Commands and CLI I/O: owns `src/commands*`, `src/cli/*`, `src/screens/*` -> `internal/commands`, `cmd/ac`, `internal/tui`.
- Lane D - UI/TUI parity: owns `src/components/*`, `src/ink/*`, `src/state/*`, `src/context/*` -> `internal/tui` (+ shared types only).
- Lane E - Remote/bridge/server: owns `src/bridge/*`, `src/remote/*`, `src/server/*`, `src/upstreamproxy/*` -> `internal/mcp`, `internal/session`, new `internal/remote` package.
- Lane F - Platform/support: owns `src/services/*`, `src/utils/*`, `src/keybindings/*`, `src/vim/*`, `src/voice/*`, `src/migrations/*`, `src/skills/*`, `src/memdir/*` -> corresponding new internal packages.
- No-overlap guidance: each lane edits only its package tree; cross-lane API changes must land first as interface-only PRs with approval from affected lane owners.

## Area map (major `src/` targets)

- [x] `src/query*`, `src/QueryEngine.ts` -> `internal/agent/*` (AC: loop, streaming, tool-use cycle, budget checks, compaction hooks compile and run).
- [x] `src/commands*` -> `internal/commands/*` baseline parser/registry/handlers (AC: `/help`, `/model`, `/permissions`, `/context` basic flows tested).
- [x] `src/tools/*` core built-ins -> `internal/tools/*` baseline (AC: Bash/Read/Write/Edit/Grep/Glob/Agent/WebFetch/WebSearch/TodoWrite/Notebook execute).
- [x] `src/services/api/*` provider bridges -> `internal/providers/*` baseline (AC: anthropic/openai/ollama/openai-compat chat + list models).
- [x] `src/hooks/*` lifecycle hooks -> `internal/hooks/hooks.go` baseline (AC: pre/post tool/chat + stop hook firing).
- [x] `src/state/session` equivalents -> `internal/session/store.go`, `internal/history/store.go` baseline (AC: transcript append/load works).
- [x] `src/tasks/*` async agent task baseline -> `internal/tasks/manager.go` (AC: start/get/wait lifecycle tested).
- [x] `src/components/*` + `src/ink/*` basic shell UI -> `internal/tui/*` baseline (AC: prompt, viewport, spinner, permission prompt compile/run).
- [ ] `src/bridge/*` -> new Go remote bridge package (AC: secure pairing/session lifecycle parity).
- [ ] `src/components/*` full interactive UX -> `internal/tui/*` expanded components (AC: dialog/menu parity matrix complete).
- [ ] `src/services/*` advanced platform features -> new Go services packages (AC: analytics/auth/update/onboarding parity).

## Backlog checklist

### 0) Program entry, boot, and runtime wiring

- [x] Port CLI entry bootstrap from `src/main.tsx` + `src/entrypoints/cli.tsx` into `cmd/ac/main.go` (AC: `ac` launches interactive session with provider/model selection).
- [x] Wire provider selection defaults from TS config conventions (AC: provider defaults to ollama when unset).
- [x] Wire model selection defaults from TS config conventions (AC: model defaults to `llama3` when unset).
- [x] Port `models` listing command baseline (AC: `ac models` returns provider model IDs).
- [ ] Add parity flags from TS runtime (`--print`, transport toggles, output styles) (AC: CLI flag surface documented and implemented).
- [ ] Add env override precedence parity (`src/setup.ts`, `src/context.ts`) (AC: env > project config > user config precedence tested).
- [ ] Port startup migration runner from `src/migrations/*` (AC: migrations execute once and persist migration version marker).
- [ ] Port startup diagnostics path from `src/screens/Doctor.tsx` and related utils (AC: doctor command returns machine-readable checks + human summary).
- [ ] Port onboarding state from `src/projectOnboardingState.ts` (AC: first-run prompts and persisted opt-outs match TS behavior).
- [ ] Port REPL launcher semantics from `src/replLauncher.tsx` and `src/screens/REPL.tsx` (AC: REPL mode command + reconnect behavior verified).
- [ ] Port resume conversation boot flow from `src/screens/ResumeConversation.tsx` (AC: resume list + selected transcript load path parity).
- [ ] Add runtime capability discovery parity for terminal features (`src/ink/terminal-querier.ts`) (AC: hyperlinks/truecolor/mouse capability detected and consumed).

### 1) Agent loop, query orchestration, and context management

- [x] Port core loop from `src/query.ts`, `src/QueryEngine.ts` to `internal/agent/loop.go` (AC: turn loop handles tool-use stop reasons and final response stop reasons).
- [x] Port streaming event pipeline baseline (AC: content/tool/thinking deltas forwarded through callback/event channels).
- [x] Port budget controls from `src/query/tokenBudget.ts` baseline (AC: token and USD budget stops emit typed stop reasons).
- [x] Port auto-compaction hooks baseline (AC: compaction is attempted near context limits without hard-failing loop).
- [x] Port permission ask/result event emission baseline (AC: ask->resolve->result events emitted for tool invocations requiring user decision).
- [x] Port hook integration points baseline (`pre_chat`, `post_chat`, `pre_tool`, `post_tool`, `stop`) (AC: hook failures surface with turn context).
- [x] Port checkpoint-before-destructive-tool baseline from TS behavior (AC: destructive tool call attempts checkpoint creation before execution).
- [ ] Port stop hook and explicit stop reason taxonomy parity from `src/query/stopHooks.ts` (AC: all TS stop reason labels mapped and persisted).
- [ ] Port full context dependency graph from `src/query/deps.ts` (AC: dependency injection structure reproduces TS lifecycle and teardown order).
- [ ] Port context trimming policy knobs from `src/query/config.ts` (AC: configurable thresholds and model-specific limits honored).
- [ ] Port queued message buffering model from `src/context/QueuedMessageContext.tsx` (AC: buffered assistant/user inserts preserve order under concurrency).
- [ ] Port prompt overlay contexts from `src/context/promptOverlayContext.tsx` (AC: overlay prompts are layered and restored correctly).
- [ ] Port mailbox/event bus model from `src/context/mailbox.tsx` (AC: subscribers receive ordered, lossless event stream with unsubscription safety).
- [ ] Port fps/stats context equivalents where relevant to TUI performance telemetry (`src/context/fpsMetrics.tsx`, `src/context/stats.tsx`) (AC: perf metrics exposed in debug output).
- [ ] Port session replay semantics from `src/assistant/sessionHistory.ts` (AC: replayed sessions reproduce original tool/result ordering).

### 2) Providers and model routing

- [x] Port Anthropic provider baseline from TS service/provider layer (AC: streamed chat + model listing supported).
- [x] Port OpenAI provider baseline (AC: streamed chat + model listing supported).
- [x] Port Ollama provider baseline (AC: local base URL default and model listing supported).
- [x] Port OpenAI-compatible provider baseline (AC: custom base URL support with streamed chat).
- [x] Port provider factory dispatch in `internal/providers/providers.go` (AC: unknown provider returns clear error with available values).
- [x] Port router baseline for primary/fast model selection (AC: simple-task switch uses `Fast` model when configured).
- [ ] Port provider option normalization from TS config stack (AC: headers, org IDs, and provider options parity-tested).
- [ ] Port provider-specific retry/backoff matrix from TS service layer (AC: retry decisions align by status code/network failure type).
- [ ] Port model metadata enrichment (context window, pricing, capability tags) from TS (`src/services/models/*`) (AC: metadata available to planner and budget logic).
- [ ] Port streaming error translation parity from TS API error utilities (`src/services/api/errorUtils.ts`) (AC: user-facing error classes/messages match expectations).
- [ ] Port auth fallback chain from TS (`env`, keychain, config) (AC: provider auth source precedence tested).
- [ ] Port image/vision payload support parity where TS supports it (AC: multimodal request path validated for enabled providers).
- [ ] Port cache hit accounting parity from TS usage metrics (AC: cache fields propagated to usage totals and emitted events).
- [ ] Port provider health checks from TS doctor routines (AC: provider connectivity checks included in diagnostics output).

### 3) Tooling framework and built-in tools

#### 3.1 Tool registry, schemas, and execution envelope

- [x] Port tool registry baseline (`src/tools.ts`, `src/Tool.ts`) to `internal/tools/registry.go` (AC: registry lookup, list, and tool defs generation works).
- [x] Port MCP-aware registry augmentation baseline (AC: MCP tools are appended when servers configured).
- [x] Port tool input schema serialization baseline (AC: each Go tool exports schema consumed by provider tool defs).
- [ ] Port tool result rich content parity (structured chunks vs plain text) from TS tool output contracts (AC: UI can render structured output blocks).
- [ ] Port tool lifecycle telemetry parity (`start`, `delta`, `done`, `error`) (AC: tool events include consistent IDs and timing fields).
- [ ] Port tool concurrency guard parity from TS tool metadata (AC: non-concurrency-safe tools run sequentially under mixed tool batches).
- [ ] Port tool-level cancellation semantics parity (AC: cancel signal interrupts long-running tool and returns deterministic cancellation message).

#### 3.2 File and search tools

- [x] Port `Bash` tool baseline from `src/tools/BashTool/*` (AC: command execution with cwd and timeout support).
- [x] Port `Read` tool baseline from `src/tools/FileReadTool/*` (AC: line-numbered reads with offset/limit behavior).
- [x] Port `Write` tool baseline from `src/tools/FileWriteTool/*` (AC: create/overwrite semantics implemented).
- [x] Port `Edit` tool baseline from `src/tools/FileEditTool/*` (AC: targeted replacement operations supported).
- [x] Port `Grep` tool baseline from `src/tools/GrepTool/*` (AC: regex search over include-filtered files).
- [x] Port `Glob` tool baseline from `src/tools/GlobTool/*` (AC: pattern file lookup returns sorted matches).
- [ ] Port Bash command security hardening parity from `src/tools/BashTool/UI.tsx` + shell classifiers (AC: risky command warnings include reason details).
- [ ] Port multi-command chaining policy parity (`&&`/`;`/parallel constraints) from TS tool policy docs (AC: policy violations blocked with actionable error).
- [ ] Port file read binary/media attachment behavior parity (AC: image/PDF read path returns attachment metadata equivalent to TS).
- [ ] Port write/edit conflict reporting parity (AC: returned error differentiates missing file, no-op edit, and multi-match ambiguity).
- [ ] Port grep result truncation and overflow-to-file parity (AC: large result sets return truncation metadata and artifact path).
- [ ] Port glob modification-time ordering parity exactly (AC: deterministic ordering test against fixture tree).

#### 3.3 Agent/task tools

- [x] Port `Agent` tool baseline from `src/tools/AgentTool/*` to `internal/tools/agent.go` (AC: start/wait/poll subagent tasks).
- [x] Port async task manager baseline from `src/tasks/*` into `internal/tasks/manager.go` (AC: running/completed/failed states with timestamps).
- [ ] Port built-in subagent persona set from `src/tools/AgentTool/built-in/*` (AC: each persona exposed with same prompt/constraints).
- [ ] Port agent memory snapshot logic from `src/tools/AgentTool/agentMemory*` (AC: subagent state snapshot/restoration parity).
- [ ] Port task stop/cancel API parity from `src/tasks/stopTask.ts` (AC: cancellation reflected in status and message protocol).
- [ ] Port local shell task wrappers from `src/tasks/LocalShellTask/*` (AC: shell task progress + detail rendering data available to TUI).
- [ ] Port remote agent task wrappers from `src/tasks/RemoteAgentTask/*` (AC: remote task status polling and terminal states match TS).

#### 3.4 Web and notebook tools

- [x] Port `WebFetch` baseline from `src/tools/WebFetchTool/*` (AC: URL fetch with markdown/text/html transforms).
- [x] Port `WebSearch` baseline from `src/tools/WebSearchTool/*` (AC: query execution and result serialization).
- [x] Port `NotebookEdit` baseline from `src/tools/NotebookEditTool/*` (AC: notebook update path available).
- [ ] Port web fetch normalization parity (automatic https upgrade, timeout handling, format defaults) (AC: behavior matches TS edge cases).
- [ ] Port web search provider abstraction parity from TS service layer (AC: configurable engines and fallback chain).
- [ ] Port notebook cell-targeted editing parity (AC: add/update/delete cell operations and metadata retention).
- [ ] Port notebook diff visualization hooks for TUI from TS components (AC: notebook edits produce renderable diff payloads).

#### 3.5 MCP and advanced tools

- [x] Port MCP client baseline from `src/tools/MCPTool/*` + `src/remote/*` subset into `internal/mcp/client.go` (AC: initialize session and execute MCP-backed tools).
- [x] Port MCP manager baseline with registry integration (AC: configured MCP servers register callable tools).
- [x] Port MCP proxy tool baseline (`internal/tools/mcp_proxy.go`) (AC: model can call MCP proxied tool through standard tool interface).
- [ ] Port list/read MCP resources tools from `src/tools/ListMcpResourcesTool/*` and `src/tools/ReadMcpResourceTool/*` (AC: resources discoverable and readable via dedicated tools).
- [ ] Port MCP auth flows from `src/tools/McpAuthTool/*` (AC: auth challenge/result workflow with persisted credentials).
- [ ] Port MCP server lifecycle controls from `src/cli/handlers/mcp.tsx` (AC: connect/reconnect/disconnect/status commands available).
- [ ] Port MCP tool permission bridge from `src/remote/remotePermissionBridge.ts` (AC: permission asks for remote tool calls route through unified resolver).
- [ ] Port `SkillTool` from `src/tools/SkillTool/*` (AC: skill discovery + loading with tool gating).
- [ ] Port `TaskCreate/TaskList/TaskGet/TaskUpdate/TaskStop/TaskOutput` family from TS tool suite (AC: CRUD operations exposed with stable schemas).
- [ ] Port `ToolSearchTool` from TS (AC: semantic tool search across built-ins and MCP with score ordering).
- [ ] Port `AskUserQuestionTool` from TS (AC: explicit user question handshake in non-interactive flows).
- [ ] Port `SleepTool` and scheduled execution (`ScheduleCronTool`) from TS (AC: delay/schedule semantics with cancellation).
- [ ] Port `RemoteTriggerTool`, `REPLTool`, and team tools from TS (AC: remote triggers and team task orchestration available).

### 4) Permissions, risk classification, and policy engine

- [x] Port permission engine baseline (`plan/default/auto/bypass`) into `internal/permissions/permissions.go` (AC: mode behavior matches documented baseline).
- [x] Port bash risk classifier baseline into `internal/permissions/classifier.go` (AC: critical/high/medium/low/none categories emitted with reasons).
- [x] Port persistent decision store baseline (AC: allow/deny remembers tool+input decisions across runs).
- [ ] Port TS permission rule DSL parity including exact regex semantics and file glob behavior (AC: same rule set yields same decision matrix).
- [ ] Port workspace-specific allowlist/denylist management from `src/components/permissions/rules/*` (AC: add/remove/list rules reflected in persisted config).
- [ ] Port recent denials ledger from TS UI state (AC: denials tracked and reviewable with timestamps/tool input summary).
- [ ] Port per-worker/per-agent permission badges and pending prompts (`src/components/permissions/Worker*`) (AC: UI indicates source of pending decision).
- [ ] Port shell permission helper messaging (`src/components/permissions/shellPermissionHelpers.tsx`) (AC: reason text and remediation hints parity).
- [ ] Port decision telemetry fields for audits from TS analytics hooks (AC: permission decisions include actor, scope, and rule origin).
- [ ] Port auto-mode nuanced allow path (safe write classes) from TS handler logic (AC: auto mode asks only for designated destructive classes).

### 5) Commands, parsing, and slash UX

- [x] Port slash parser baseline from TS command parser to `internal/commands/parser.go` (AC: quoted args, escapes, and slash detection tested).
- [x] Port command registry baseline to `internal/commands/registry.go` (AC: aliases, lookup, canonical ordering tested).
- [x] Port baseline handlers (`help/model/compact/permissions/context/resume/...`) into `internal/commands/handlers.go` (AC: each command parses and returns stable message).
- [x] Port `buddy` command baseline (AC: enable/disable/status flows are available through slash command UX).
- [ ] Port full command set from `src/commands/*` (AC: command inventory parity report reaches 100% coverage).
- [ ] Port `add-dir` command (`src/commands/add-dir/*`) (AC: workspace dir registration with validation and persistence).
- [ ] Port `agents` command suite (`src/commands/agents/*`) (AC: list/create/manage agents from CLI).
- [ ] Port `branch` command full behavior (`src/commands/branch/*`) (AC: branch status/actions integrated with git state).
- [ ] Port `bridge` commands (`src/commands/bridge*`) (AC: bridge setup/status/kick workflows available).
- [ ] Port `clear` command suite (`src/commands/clear/*`) (AC: selective clear operations with confirmations).
- [ ] Port `compact` advanced modes (`src/commands/compact/*`) (AC: manual and threshold-triggered compaction controls).
- [ ] Port `config` command full CRUD (`src/commands/config/*`) (AC: nested config get/set/unset/list with validation).
- [ ] Port `cost` command with session/model breakdown (`src/commands/cost/*`) (AC: cost report includes token and USD segments).
- [ ] Port `doctor` command detail suite (`src/commands/doctor/*`) (AC: machine readable + human report with actionable failures).
- [ ] Port `history` and resume commands (`src/commands/history*`) (AC: list/open/filter sessions parity).
- [ ] Port `init` command full project bootstrap (`src/commands/init/*`) (AC: config scaffolding and idempotence checks).
- [ ] Port `install` and plugin/skill install commands (`src/commands/install.tsx`) (AC: install path with signature/compat checks).
- [ ] Port `mcp` command suite from TS handlers (AC: server add/remove/list/test workflows).
- [ ] Port `model` command advanced features (provider-qualified models, fallback checks) (AC: rejects invalid model/provider combinations).
- [ ] Port `permissions` command subcommands parity (`get/list/set/rules`) (AC: command outputs + side effects match TS).
- [ ] Port `vim` and `voice` commands (`src/commands/vim/*`, `src/commands/voice/*`) (AC: enable/disable/toggle states persisted and reflected in UI).
- [ ] Port hidden/debug commands used by TS power workflows (AC: debug docs include all parity commands).

### 6) TUI shell, rendering, and interaction parity

- [x] Port baseline Bubble Tea app shell (`internal/tui/app.go`) equivalent to minimal TS app loop (AC: input/viewport/spinner/status render end-to-end).
- [x] Port baseline input model (`internal/tui/input.go`) (AC: submit, focus, disable behavior works across states).
- [x] Port baseline spinner and permission prompt models (`internal/tui/spinner.go`, `internal/tui/permissions.go`) (AC: thinking/executing/prompt transitions render correctly).
- [x] Port baseline diff rendering helper (`internal/tui/diff.go`) (AC: file diff output displays with add/remove markers).
- [x] Port buddy mode TUI integration baseline (AC: buddy state is visible in-session and updates during mode transitions).
- [x] Port spinner branding baseline (AC: branded spinner states render without obscuring status or error text).
- [ ] Port high-fidelity render engine parity with `src/ink/*` custom renderer (AC: wrapping, width measurement, unicode handling, and ansi behavior match fixtures).
- [ ] Port scroll box behavior from `src/ink/components/ScrollBox.tsx` (AC: virtual scroll and anchor preservation under streaming).
- [ ] Port selection and search highlight from `src/ink/selection.ts`, `src/ink/searchHighlight.ts` (AC: selection model and highlight output parity).
- [ ] Port terminal focus and notification hooks (`src/ink/terminal-focus-state.ts`, `src/ink/useTerminalNotification.ts`) (AC: focus changes update UI states and notifications).
- [ ] Port hyperlink support checks (`src/ink/supports-hyperlinks.ts`) (AC: links enable only on supported terminals).
- [ ] Port keyboard parsing parity from `src/ink/parse-keypress.ts` (AC: key combinations map identically including modifiers).
- [ ] Port render optimizer/node cache strategy from `src/ink/optimizer.ts`, `src/ink/node-cache.ts` (AC: performance benchmark meets baseline without visual regressions).

### 7) Components and dialogs (TS components -> Go TUI modules)

- [ ] Port app root orchestrator from `src/components/App.tsx` (AC: main state machine includes all modal/dialog transitions).
- [ ] Port global quick open and search dialogs (`src/components/QuickOpenDialog.tsx`, `src/components/GlobalSearchDialog.tsx`, `src/components/HistorySearchDialog.tsx`) (AC: fuzzy search and keyboard navigation parity).
- [ ] Port help screens (`src/components/HelpV2/*`) (AC: command/tool/help sections match TS content and keybind hints).
- [ ] Port configuration error dialogs (`InvalidConfigDialog`, `InvalidSettingsDialog`) (AC: validation errors displayed with corrective actions).
- [ ] Port permission dialogs (`BypassPermissionsModeDialog`, `AutoModeOptInDialog`) (AC: choice persistence and warning copy parity).
- [ ] Port onboarding dialogs (`IdeOnboardingDialog`, `ClaudeInChromeOnboarding`, `IdeAutoConnectDialog`) (AC: first-run onboarding progression parity).
- [ ] Port update dialogs (`AutoUpdater`, `AutoUpdaterWrapper`, `ChannelDowngradeDialog`) (AC: update status and channel transitions displayed).
- [ ] Port cost and usage dialogs (`CostThresholdDialog`, `CompactSummary`) (AC: threshold triggers and summary rendering parity).
- [ ] Port task dialogs (`BackgroundTasksDialog`, `ShellDetailDialog`, `RemoteSessionDetailDialog`, `DreamDetailDialog`, `AsyncAgentDetailDialog`) (AC: task detail panes show same fields/status mapping).
- [ ] Port context/visualization components (`ContextVisualization`, `ContextSuggestions`) (AC: context sources and suggestions shown with same ranking semantics).
- [ ] Port file edit result components (`FileEditToolDiff`, updated/rejected messages) (AC: diff + rejection reason presentation parity).
- [ ] Port tree/ordered list UI primitives (`TreeSelect`, `OrderedList*`) (AC: selection, indentation, and keyboard controls parity).
- [ ] Port wizard framework (`src/components/wizard/*`) (AC: multi-step flows retain state and validation between steps).
- [ ] Port teammates/team status UI (`src/components/teams/*`, `CoordinatorAgentStatus`) (AC: teammate lifecycle and status colors/messages parity).
- [ ] Port shell output primitives (`src/components/shell/*`) (AC: progressive shell output and elapsed time render parity).

### 8) State stores, hooks, and local app state

- [ ] Port app state store from `src/state/AppStateStore.ts`, `src/state/store.ts` (AC: central state mutations + selectors available in Go TUI runtime).
- [ ] Port state selectors from `src/state/selectors.ts` (AC: selectors return identical derived fields for fixture states).
- [ ] Port on-change side effect handlers from `src/state/onChangeAppState.ts` (AC: side effects trigger exactly once per relevant state change).
- [ ] Port teammate view helpers from `src/state/teammateViewHelpers.ts` (AC: teammate grouping/filtering parity).
- [ ] Port key notification hooks from `src/hooks/notifs/*` into Go event-notification layer (AC: each notification trigger condition mapped and tested).
- [ ] Port tool permission hooks from `src/hooks/toolPermission/*` (AC: interactive/coordinator/swarm worker permission flows match TS).
- [ ] Port generic hooks needed for UX (`useInterrupt`, `usePaste`, `useVirtualScroll`, `useVoiceIntegration`) (AC: corresponding interactions behave equivalently).
- [ ] Port unified suggestions + file suggestions hooks (`src/hooks/unifiedSuggestions.ts`, `src/hooks/fileSuggestions.ts`) (AC: suggestion ranking and insertion parity).
- [ ] Port async retry/debounce hook behaviors used by TS UI (AC: timing semantics preserved within tolerance).
- [ ] Port modal/overlay contexts from `src/context/modalContext.tsx`, `src/context/overlayContext.tsx` (AC: stacked modal behavior parity).

### 9) Bridge, remote sessions, server, and upstream proxy

- [ ] Port bridge config and enablement from `src/bridge/bridgeConfig.ts`, `bridgeEnabled.ts` (AC: bridge can be configured, toggled, and persisted).
- [ ] Port bridge main loop from `src/bridge/bridgeMain.ts` and `sessionRunner.ts` (AC: session lifecycle start/stop/reconnect parity).
- [ ] Port bridge messaging protocol from `src/bridge/bridgeMessaging.ts`, `inboundMessages.ts` (AC: message schema and routing parity).
- [ ] Port bridge permission callbacks from `src/bridge/bridgePermissionCallbacks.ts` (AC: remote permission asks integrate with local decision engine).
- [ ] Port secure work secret/JWT utilities from `src/bridge/workSecret.ts`, `jwtUtils.ts` (AC: token issue/verify and rotation behavior parity).
- [ ] Port trusted device model from `src/bridge/trustedDevice.ts` (AC: device trust registration/revocation flows implemented).
- [ ] Port session creation APIs from `src/bridge/createSession.ts`, `src/server/createDirectConnectSession.ts` (AC: direct connect sessions created with expected metadata).
- [ ] Port remote bridge core from `src/bridge/remoteBridgeCore.ts` (AC: remote execution path supports reconnect and backpressure).
- [ ] Port remote session manager from `src/remote/RemoteSessionManager.ts` (AC: remote session state machine parity).
- [ ] Port sessions websocket transport from `src/remote/SessionsWebSocket.ts` (AC: connect/auth/heartbeat/retry semantics parity).
- [ ] Port SDK message adapter from `src/remote/sdkMessageAdapter.ts` (AC: message conversion preserves tool and stream event fidelity).
- [ ] Port direct connect manager from `src/server/directConnectManager.ts` (AC: lifecycle and timeout handling parity).
- [ ] Port upstream proxy relay from `src/upstreamproxy/relay.ts` and `upstreamproxy.ts` (AC: request/response streaming passthrough parity).
- [ ] Port transport stack from `src/cli/transports/*` (`SSE`, `WebSocket`, `Hybrid`, batch uploaders) (AC: fallback and reconnection matrix matches TS).

### 10) Config, schemas, constants, and migrations

- [x] Port baseline YAML config load/save from `src/setup.ts` conventions to `internal/config/config.go` (AC: load missing file defaults and save path creation works).
- [x] Port buddy config persistence baseline (AC: buddy mode preferences persist across restarts and resume flows).
- [ ] Port project + user config merge semantics from TS (`.alliecode/config.yaml` + user config) (AC: layered merge precedence parity).
- [ ] Port strict config schema validation from TS (`src/schemas/hooks.ts`, keybinding schema, settings schema) (AC: invalid config errors include path and reason).
- [ ] Port constants parity from `src/constants/*` (AC: product/tool/prompt/system constants available in Go and tested against snapshots).
- [ ] Port output styles loader from `src/outputStyles/loadOutputStylesDir.ts` (AC: custom output styles discovered and applied).
- [ ] Port key migration scripts from `src/migrations/*` (AC: each migration has idempotent test and migration ledger update).
- [ ] Port legacy model migration helpers (`migrateFennecToOpus`, `migrateSonnet*`) (AC: old model IDs auto-updated once).
- [ ] Port auto mode/settings reset migrations (`resetAutoModeOptInForDefaultOffer`, `resetProToOpusDefault`) (AC: conditional resets match TS guard conditions).
- [ ] Port config doctor checks around migration state (AC: doctor flags partial migration or corrupted state).

### 11) Keybindings, Vim mode, and input behavior

- [ ] Port keybinding parser from `src/keybindings/parser.ts` (AC: modifier parsing and chord handling parity).
- [ ] Port keybinding resolver from `src/keybindings/resolver.ts` (AC: user bindings override defaults with conflict handling).
- [ ] Port shortcut matching and display from `src/keybindings/match.ts`, `useShortcutDisplay.ts` (AC: rendered key hints match resolved binding set).
- [ ] Port keybinding schema + validation from `src/keybindings/schema.ts`, `validate.ts` (AC: invalid bindings rejected with field-level diagnostics).
- [ ] Port default bindings and template generation from `src/keybindings/defaultBindings.ts`, `template.ts` (AC: generated template stable and complete).
- [ ] Port reserved shortcut rules from `src/keybindings/reservedShortcuts.ts` (AC: protected combinations cannot be rebound).
- [ ] Port provider/context wrappers (`KeybindingContext`, `KeybindingProviderSetup`) (AC: runtime rebind updates active handlers without restart).
- [ ] Port Vim motions from `src/vim/motions.ts` (AC: word/line motions parity).
- [ ] Port Vim operators from `src/vim/operators.ts` (AC: delete/change/yank operator state transitions parity).
- [ ] Port Vim text objects from `src/vim/textObjects.ts` (AC: object selection semantics parity).
- [ ] Port Vim transition and mode types from `src/vim/transitions.ts`, `types.ts` (AC: normal/insert/visual transitions parity).
- [ ] Port voice mode enablement from `src/voice/voiceModeEnabled.ts` and `src/services/voice*` basics (AC: voice mode toggle and capability detection parity).

### 12) Skills, plugins, memdir, and extensibility

- [x] Port baseline skills loading from `src/skills/loadSkillsDir.ts` into `internal/skills/skills.go` (AC: local markdown skills discoverable and loadable).
- [ ] Port bundled skills registry from `src/skills/bundled/*` (AC: bundled skills shipped with metadata and tool allowlists).
- [ ] Port bundled skills index from `src/skills/bundledSkills.ts` (AC: one-to-one skill inventory parity).
- [ ] Port MCP skill builders from `src/skills/mcpSkillBuilders.ts` (AC: MCP server capabilities produce generated skills).
- [ ] Port plugin discovery from `src/plugins/builtinPlugins.ts`, `src/plugins/bundled/index.ts` (AC: plugin manifest scan + registration parity).
- [ ] Port plugin lifecycle events from TS plugin runtime (AC: load/init/teardown hooks and failure isolation).
- [ ] Port memdir scanner from `src/memdir/memoryScan.ts` and `memdir.ts` (AC: project memory extraction + update cadence parity).
- [ ] Port memory relevance ranking from `src/memdir/findRelevantMemories.ts` (AC: top-N selection parity on fixture corpus).
- [ ] Port memory aging and path conventions from `src/memdir/memoryAge.ts`, `paths.ts`, `teamMemPaths.ts` (AC: TTL and file layout parity).
- [ ] Port team memory prompts from `src/memdir/teamMemPrompts.ts` (AC: generated prompt text and insertion points match TS).

### 13) Services layer parity (high-level)

- [ ] Port analytics services (`src/services/analytics/*`) with configurable sinks (AC: event buffering, flush, and kill-switch parity).
- [ ] Port auth/account services (`src/services/auth*`, oauth helpers) (AC: sign-in state and token refresh flows parity).
- [ ] Port update services (`src/services/update*`) (AC: channel check/download/install lifecycle parity).
- [ ] Port API bootstrap and client helpers (`src/services/api/bootstrap.ts`, `client.ts`) (AC: consistent request headers, retries, and error wrapping).
- [ ] Port token/cost tracking services (`src/cost-tracker.ts`, `src/costHook.ts`) (AC: per-turn and session cumulative cost parity).
- [ ] Port history services (`src/history.ts`) (AC: conversation history listing/filtering parity).
- [ ] Port sandbox services (`src/components/sandbox/*` + service dependencies) (AC: sandbox config doctor/deps/overrides flows parity).
- [ ] Port IDE integration services (status indicator/autoconnect) (AC: IDE detection and connect callouts parity).
- [ ] Port prompt suggestion and speculation services (`src/services/PromptSuggestion/*`) (AC: suggestion latency and ranking parity).
- [ ] Port session memory service (`src/services/SessionMemory/*`) (AC: summarize/store/retrieve flow parity).
- [ ] Port agent summary service (`src/services/AgentSummary/*`) (AC: background summary generation parity).
- [ ] Port magic docs service (`src/services/MagicDocs/*`) (AC: doc generation and prompt templates parity).

### 14) Tests, fixtures, and verification matrix

- [x] Maintain existing Go unit suites for core modules (`internal/agent`, `internal/commands`, `internal/tools`, `internal/mcp`, `internal/tasks`, `internal/session`) (AC: all current tests pass in CI).
- [x] Add buddy mode test coverage baseline (AC: buddy core, command wiring, and TUI integration paths are covered by unit/integration tests).
- [x] Update parity test harness baseline for buddy/spinner flows (AC: harness captures and validates branded spinner and buddy state transitions).
- [ ] Create TS-vs-Go parity fixture harness for agent loop transcripts (AC: same input fixture yields equivalent ordered events and terminal outputs).
- [ ] Add command parity tests for every migrated slash command (AC: parse + execute golden tests for happy and error paths).
- [ ] Add permission engine matrix tests from TS rule fixtures (AC: decisions match expected outputs across all modes).
- [ ] Add provider contract tests with mock servers for stream fragmentation and error retries (AC: deterministic pass across providers).
- [ ] Add TUI snapshot tests for core workflows (AC: snapshots stable across terminal width variants).
- [ ] Add end-to-end tests for tool-call loops (read/edit/write/grep/glob/webfetch/agent) (AC: scripted runs pass in isolated temp repos).
- [ ] Add MCP integration tests (AC: mock MCP server tool/resource/auth scenarios covered).
- [ ] Add checkpoint/undo workflow tests (AC: destructive tool path creates restorable checkpoint).
- [ ] Add session resume/compaction boundary tests (AC: resumed transcript behavior and compaction metadata intact).
- [ ] Add load/perf smoke tests for long streaming sessions (AC: no deadlocks; bounded memory growth).
- [ ] Add flaky-test quarantine and retry policy for networked provider tests (AC: deterministic CI status).

### 15) Documentation, developer workflow, and release gates

- [ ] Add `PARITY_STATUS.md` generated summary from this ledger (AC: machine-generated completion percentages by lane).
- [ ] Document TS-to-Go API mapping for every major subsystem (AC: each `src/*` top-level area linked to owning Go package).
- [ ] Add contribution guide for lane-based swarm workflow and interface-first PR process (AC: guide includes examples and do/dont overlap rules).
- [ ] Add migration playbooks for command/tool/component ports (AC: each playbook includes checklist + required tests).
- [ ] Add release parity gate checklist (AC: release blocked unless functional/UX/test parity thresholds pass).
- [ ] Add known-gaps section to README/ROADMAP with direct links to open ledger items (AC: docs stay synchronized via CI check).
- [ ] Add automated lint for unchecked AC format in this file (AC: CI fails if checklist item lacks `AC:` clause).

## Completion threshold targets

- [ ] Functional parity target: >= 95% of TS critical workflows covered in Go acceptance matrix (AC: matrix report attached to release PR).
- [ ] UX parity target: >= 90% of high-frequency TUI interactions match TS snapshots (AC: UX diff audit signed off).
- [ ] Test parity target: Go suite includes unit + integration coverage for every migrated lane (AC: minimum per-lane coverage budget met).
- [ ] Operational parity target: no P0/P1 regressions in hooks, permissions, session persistence, and provider failover (AC: burn-in run passes).

## Remaining blockers to 100% local parity

- [ ] Bridge/remote stack still required for equivalent local replacement workflows (`src/bridge/*`, `src/remote/*`, `src/upstreamproxy/*`) (AC: local operator can perform all current bridge-mediated workflows without TS fallback).
- [ ] Full command inventory parity is incomplete (notably `doctor`, `history`, `mcp`, `config` advanced CRUD, install/bootstrap paths) (AC: command parity report shows no missing user-facing commands).
- [ ] TUI high-fidelity rendering parity is incomplete (scroll/selection/search/key parsing/renderer optimizations) (AC: snapshot and interaction fixtures pass against TS-compatible outputs).
- [ ] Config/schema/migration parity remains incomplete (merge precedence, strict schema errors, migration ledger + doctor checks) (AC: legacy and mixed-config fixtures pass with identical diagnostics).
- [ ] Advanced tooling parity remains incomplete (MCP resource/auth flows, task CRUD tool family, SkillTool/tool search/user-question/scheduling tools) (AC: tool inventory and schema parity report reaches 100%).
- [ ] End-to-end parity harness is incomplete (full transcript, provider retry, TUI width variants, MCP integration, long-stream stability) (AC: CI parity matrix is deterministic and release-gating).
