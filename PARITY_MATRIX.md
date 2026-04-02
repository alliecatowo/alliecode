# Parity Matrix (Current Repo State)

## Scope and Measurement

- Source of truth: current Go implementation under `cmd/ac` and `internal/*`, plus deterministic integration/snapshot tests in `tests/*`.
- Parity labels:
  - `full`: implemented and behaviorally close to TS target for current surface.
  - `partial`: implemented baseline/stubbed behavior, or parity gaps remain.
  - `missing`: not implemented in Go path for required TS behavior.
- Quant baseline from tracked parity ledger: `61 / 257` checklist items complete (`23.7%`) from `PARITY_TODO.md`.

## Quantified Subsystem Snapshot

| Subsystem | Status | Completion Signal | Notes |
|---|---|---:|---|
| Program boot/runtime wiring | partial | 33.3% (4/12) | CLI boot, onboarding, startup migration runner exist; parity flags/transports and diagnostics still incomplete. |
| Agent loop + orchestration | partial | 46.7% (7/15) | Streaming/tool loop/checkpoints/hooks/session append are present; full stop taxonomy/context overlay parity pending. |
| Providers + routing | partial | 42.9% (6/14) | `anthropic/openai/ollama/openai-compat/gemini` providers and policy routing exist; retry/auth/metadata parity gaps remain. |
| Tools framework + built-ins | partial | 38.5% (15/39 full) | 39 tools are now registered (including orchestration and remote trigger families); advanced TS lifecycle parity still open. |
| Permissions/policy engine | partial | 30.0% (3/10) | Mode engine, classifier, persistent store implemented; full TS rule DSL/worker UX parity pending. |
| Slash commands and UX | partial | 6.0% (4/67 full) | 67 commands are now registered; most remain deterministic runtime-state handlers vs TS-integrated workflows. |
| TUI shell/rendering | partial | 46.2% (6/13) | Bubble Tea baseline + permission/buddy/search helpers implemented; high-fidelity renderer/component parity pending. |
| State filesystem | partial | 20.0% (2/10, config/migrations bucket) | `internal/state/paths.go` creates state roots/files; TS migration/state coupling remains incomplete. |
| History/session/settings persistence | partial | mixed | `internal/history/store.go`, `internal/session/store.go`, `internal/settings/manager.go` are functional baselines; full replay/search/settings parity pending. |
| Bridge/remote/MCP | partial | bridge+remote baseline landed (self-hosted-first) | Bridge config/protocol/runner/trusted-device/work-secret and remote adapter/transport/session-manager baselines now exist; self-hosted `/session host` and `/session connect` are the default remote direction while TS lifecycle parity remains. |
| Vim/keybindings/voice | partial | 0.0% (0/12 in dedicated bucket) | Vim/keybindings/voice engines exist in Go, but TS behavior parity remains largely open. |
| Buddy | partial | baseline complete | Buddy core + TUI integration + command flow exist; full TS feature surface and persistence nuances still open. |

## Providers Inventory

| Provider | Status | Current Go State | Primary Files |
|---|---|---|---|
| anthropic | partial | Chat stream + model listing present; full error/auth fallback parity pending. | `internal/providers/anthropic/anthropic.go`, `internal/providers/providers.go` |
| openai | partial | Chat stream + model listing present; full retry/options parity pending. | `internal/providers/openai/openai.go`, `internal/providers/providers.go` |
| ollama | partial | Local default/base URL support + models present; broader parity pending. | `internal/providers/ollama/ollama.go`, `internal/providers/providers.go` |
| openai-compat | partial | Base URL + stream path present; edge compatibility parity pending. | `internal/providers/openaicompat/openaicompat.go`, `internal/providers/providers.go` |
| gemini | partial | Provider implementation exists; end-to-end parity matrix incomplete. | `internal/providers/gemini/gemini.go`, `internal/providers/providers.go` |
| routing policy | partial | Capability-aware selection engine exists; TS behavior matrix incomplete. | `internal/providers/routing_policy.go`, `internal/providers/capability_matrix.go` |

## Command-by-Command Parity Matrix (67)

| Command | Status | Current State |
|---|---|---|
| `/help` | full | Registry help output complete for current command set. |
| `/model` | partial | Model switching validates known metadata; provider/runtime coupling still limited. |
| `/compact` | partial | Stateful command behavior implemented; full compaction lifecycle parity pending. |
| `/branch` | partial | Deterministic branch state commands; not full git-backed parity. |
| `/diff` | partial | Deterministic diff-entry handling; not full git/worktree parity. |
| `/cost` | partial | Token/cost counters surfaced from runtime state; full service parity pending. |
| `/doctor` | partial | Human/json diagnostic output exists; full TS doctor checks pending. |
| `/config` | partial | Show/get/set/unset baseline in runtime state; full config CRUD parity pending. |
| `/init` | partial | Baseline init semantics; project bootstrap parity pending. |
| `/copy` | partial | Runtime copy state update exists; full clipboard/platform parity pending. |
| `/version` | full | Version command stable. |
| `/usage` | full | Usage command output stable. |
| `/context` | partial | Show/clear behavior implemented; full context graph parity pending. |
| `/exit` | full | Exit state signaling complete for current flow. |
| `/plan` | partial | Plan workflow status exists; full TS plan orchestration pending. |
| `/review` | partial | Review counters/state implemented; external review workflow parity pending. |
| `/session` | partial | Session status/url state handling baseline only. |
| `/skills` | partial | List/status/add/remove in runtime state; full skill loading lifecycle parity pending. |
| `/rewind` | partial | Rewind state/status implemented; actual rewind integration pending. |
| `/tag` | partial | Tag status/set baseline; TS parity incomplete. |
| `/remote-env` | partial | Remote env status/set baseline; remote integration pending. |
| `/security-review` | partial | Command flow exists; full security-review workflow parity pending. |
| `/permissions` | partial | Modes/rules/denials command surface exists; full DSL and UI parity pending. |
| `/resume` | partial | Resume request/status surface exists; full transcript resume UX parity pending. |
| `/add-dir` | partial | Workspace dir registration baseline only. |
| `/agents` | partial | Agent list/create/status baseline only. |
| `/clear` | partial | Status/all/context/display counters implemented; richer clear semantics pending. |
| `/history` | partial | List/show over runtime entries; full persistent history UX parity pending. |
| `/mcp` | partial | list/connect/disconnect/status command exists; full server lifecycle parity pending. |
| `/vim` | partial | enable/disable/status command surface exists; full editor parity pending. |
| `/voice` | partial | enable/disable/status command surface exists; full voice integration parity pending. |
| `/login` | partial | Login state/provider/account flow in runtime state; real auth parity pending. |
| `/logout` | partial | Logout state transitions baseline. |
| `/buddy` | partial | Hatch/pet/mute/status/help command flow implemented. |
| `/theme` | partial | get/list/set/cycle/preview command surface implemented. |
| `/files` | partial | list/add/remove/clear/status runtime state flow exists. |
| `/output-style` | partial | status/get/list/set command surface exists. |
| `/statusline` | partial | setup/status baseline exists; full terminal integration parity pending. |
| `/keybindings` | partial | status/path/open/enable/disable command surface exists. |
| `/status` | partial | Status summary exists; full TS status matrix parity pending. |
| `/stats` | partial | Stats command output exists; full telemetry parity pending. |
| `/memory` | partial | status/list/add/remove/clear baseline exists; full memdir parity pending. |
| `/privacy-settings` | partial | telemetry/training toggles implemented; broader privacy feature parity pending. |
| `/upgrade` | partial | status/plan variants available; updater/account parity pending. |
| `/terminal-setup` | partial | status/detect/apply baseline exists; full terminal profile parity pending. |
| `/release-notes` | partial | latest/list/status baseline exists. |
| `/advisor` | partial | status/get/set runtime flow exists; advisor routing/policy parity pending. |
| `/btw` | partial | status/ask baseline exists; conversational helper parity pending. |
| `/chrome` | partial | status/default/extension/connect baseline exists; full browser bridge parity pending. |
| `/color` | partial | status/get/set baseline exists; full session-color UX parity pending. |
| `/desktop` | partial | status/handoff baseline exists; OS integration parity pending. |
| `/mobile` | partial | status/qr baseline exists; full mobile handoff parity pending. |
| `/fast` | partial | status/on/off/toggle baseline exists; model-routing parity pending. |
| `/effort` | partial | status/get/set/list baseline exists; full reasoning-level parity pending. |
| `/plugin` | partial | list/install/uninstall/enable/disable/marketplaces baseline exists. |
| `/reload-plugins` | partial | plugin reload command baseline exists; runtime plugin lifecycle parity pending. |
| `/export` | partial | markdown/json/path baseline exists; full export format parity pending. |
| `/extra-usage` | partial | status/request baseline exists; account/quotas parity pending. |
| `/rate-limit-options` | partial | status/list/choose baseline exists; upstream policy parity pending. |
| `/pr-comments` | partial | status/fetch baseline exists; GitHub API parity pending. |
| `/web-setup` | partial | status/connect baseline exists; web onboarding parity pending. |
| `/install-github-app` | partial | status/repo/start command flow baseline exists. |
| `/install-slack-app` | partial | status/start command flow baseline exists. |
| `/feedback` | partial | status/submit flow baseline exists. |
| `/hooks` | partial | status/enable/disable flow exists; full hook runtime parity pending. |
| `/sandbox` | partial | status + mode set command exists; sandbox service parity pending. |
| `/tasks` | partial | list/add/done/clear runtime flow exists; full async task parity pending. |

Command parity totals: `full=4`, `partial=63`, `missing=0`.

## Tool-by-Tool Parity Matrix (Current Known Inventory, 39)

| Tool | Status | Current Go State | Primary Files |
|---|---|---|---|
| `Bash` | full | Command exec + risk guard + truncation metadata + tests. | `internal/tools/bash.go`, `internal/permissions/classifier.go` |
| `Read` | full | Line-numbered reads + guardrails + stale-read metadata support. | `internal/tools/fileread.go`, `internal/tools/read_guardrails.go` |
| `Write` | full | Create/overwrite + stale-read conflict detection. | `internal/tools/filewrite.go` |
| `Edit` | full | Targeted string replacement with validation. | `internal/tools/fileedit.go` |
| `Grep` | full | Regex search + truncation metadata behavior. | `internal/tools/grep.go` |
| `Glob` | full | Glob search + truncation metadata behavior. | `internal/tools/glob.go` |
| `Agent` | partial | Subagent task baseline exists; full persona/memory parity pending. | `internal/tools/agent.go`, `internal/tasks/manager.go` |
| `WebFetch` | partial | URL fetch + format transforms baseline; TS edge parity pending. | `internal/tools/webfetch.go` |
| `WebSearch` | partial | Search baseline exists; provider abstraction parity pending. | `internal/tools/websearch.go` |
| `TodoWrite` | partial | Checklist tool baseline exists; richer TS semantics pending. | `internal/tools/todowrite.go` |
| `NotebookEdit` | partial | Notebook edit baseline exists; cell-level parity pending. | `internal/tools/notebook.go` |
| `AskUserQuestion` | full | Structured question schema + validation contract tested. | `internal/tools/ask_user_question.go` |
| `Sleep` | full | Bounded wait + cancel semantics tested. | `internal/tools/sleep.go` |
| `task_list` | full | Task listing contract implemented. | `internal/tools/tasklist.go` |
| `task_get` | full | Task fetch contract implemented. | `internal/tools/task_get.go` |
| `task_create` | full | Task creation contract implemented. | `internal/tools/task_create.go` |
| `task_update` | full | Task update contract implemented. | `internal/tools/task_update.go` |
| `task_stop` | full | Task stop contract implemented. | `internal/tools/task_stop.go` |
| `task_output` | full | Task output polling/edge behaviors implemented. | `internal/tools/task_output.go` |
| `tool_search` | full | Tool discovery/search contract implemented. | `internal/tools/tool_search.go` |
| `skill` | partial | list/invoke/history baseline exists; bundled-skill parity still expanding. | `internal/tools/skill.go` |
| `config` | partial | get/set/list/unset baseline exists; full layered config parity pending. | `internal/tools/config.go` |
| `lsp` | partial | hover/definition baseline exists; full TS LSP workflow parity pending. | `internal/tools/lsp.go` |
| `worktree_enter` | partial | Worktree enter lifecycle baseline exists; git-integrated parity pending. | `internal/tools/worktree_tools.go` |
| `worktree_exit` | partial | Worktree exit/remove baseline exists; safety/policy parity pending. | `internal/tools/worktree_tools.go` |
| `send_message` | partial | Team message dispatch baseline exists; coordination parity pending. | `internal/tools/send_message.go` |
| `team_create` | partial | Team creation baseline exists; teammate orchestration parity pending. | `internal/tools/team_tools.go` |
| `team_delete` | partial | Team deletion baseline exists; lifecycle parity pending. | `internal/tools/team_tools.go` |
| `cron_create` | partial | Deterministic cron creation baseline exists; scheduler parity pending. | `internal/tools/cron_tools.go` |
| `cron_delete` | partial | Deterministic cron deletion baseline exists; scheduler parity pending. | `internal/tools/cron_tools.go` |
| `cron_list` | partial | Deterministic cron listing baseline exists; scheduler parity pending. | `internal/tools/cron_tools.go` |
| `remote_trigger` | partial | create/list/run baseline exists; remote service parity pending. | `internal/tools/remote_trigger.go` |
| `brief` | partial | Structured completion brief baseline exists; reporting parity pending. | `internal/tools/brief.go` |
| `ls` | full | Directory listing utility tool implemented. | `internal/tools/ls.go` |
| `memory` | partial | Memory command tool baseline only; full memdir parity pending. | `internal/tools/memory.go` |
| `mcp_proxy` (dynamic) | partial | MCP tool proxy exists via manager registration path. | `internal/tools/mcp_proxy.go`, `internal/tools/registry.go` |
| `mcp_resource_list` | partial | MCP resource listing tool exists; broader MCP workflow parity pending. | `internal/tools/mcp_resource_tools.go` |
| `mcp_resource_read` | partial | MCP resource read tool exists; broader MCP workflow parity pending. | `internal/tools/mcp_resource_tools.go` |
| `mcp_auth_local` | partial | Local MCP auth tool exists; full auth lifecycle parity pending. | `internal/tools/mcp_resource_tools.go` |

Tool parity totals: `full=15`, `partial=24`, `missing=0`.

## Blocker Tiers (Execution Priority)

| Tier | Blockers | Count | Gate Condition |
|---|---|---:|---|
| `P0` | Command/workflow non-integration, parity harness gap, permissions DSL mismatch, self-hosted-first bridge/remote production hardening | 12 | Blocks end-to-end behavioral parity and safe remote operation. |
| `P1` | Provider retry/auth/metadata parity, TUI renderer/components parity, state-store/on-change parity, MCP lifecycle parity | 24 | Blocks high-confidence migration and production-grade UX. |
| `P2` | Polish and ecosystem parity (plugins/memdir/services), performance tuning, docs/release gates | 19 | Important for completeness but not first blocking path. |

## Massive Implementation Backlog (Grouped, Actionable, File-Targeted)

### P0 - Bridge/Remote/MCP Critical Path (Self-Hosted First)

- [ ] Wire bridge enablement end-to-end through startup/runtime flow in `cmd/ac/main.go` using `internal/bridge/config.go` and `internal/bridge/runner.go`.
- [ ] Complete bridge inbound/outbound request orchestration and command execution integration in `internal/bridge/protocol.go` and `internal/bridge/remote.go`.
- [ ] Finish permission prompt callback parity for remote asks in `internal/bridge/permission.go` with `internal/permissions/permissions.go`.
- [ ] Harden trusted-device persistence/rotation semantics in `internal/bridge/trusted_devices.go` and `internal/bridge/state.go`.
- [ ] Finish self-hosted direct-connect/session pointer lifecycle parity in `internal/bridge/session_pointer.go` and server wiring.
- [ ] Extend `internal/remote/transport.go` fallback behavior (SSE/WebSocket/Hybrid) to match TS retry semantics.
- [ ] Complete remote adapter fidelity for tool/stream/cancel events in `internal/remote/adapter.go`.
- [ ] Harden reconnect/backpressure/heartbeat policies in `internal/remote/session_manager.go` and `internal/remote/retry_heartbeat.go`.
- [ ] Add MCP server lifecycle operations (`add/remove/test/reconnect`) in `internal/commands/handlers.go` and `internal/mcp/manager.go`.
- [ ] Add MCP permission bridge parity in `internal/mcp/manager.go` and `internal/tools/mcp_proxy.go`.

### P0 - Command Integration and Runtime Parity

- [ ] Replace runtime-state-only `/branch` and `/diff` behavior with git-backed execution in `internal/commands/handlers.go` and `internal/tools/bash.go` integration path.
- [ ] Back `/history` with `internal/history/store.go` retrieval flows and session linkage from `internal/session/store.go`.
- [ ] Back `/session` and `/resume` with real transcript store operations in `internal/commands/handlers.go` and `internal/agent/loop.go`.
- [ ] Wire `/config` CRUD to layered config persistence in `internal/config/config.go` and `internal/settings/manager.go`.
- [ ] Expand `/doctor` to real diagnostics (providers, filesystem, migrations) via `internal/config/diagnostics.go` and `internal/providers/*` probes.
- [ ] Integrate `/tasks` and task tools against `internal/tasks/manager.go` rather than command-local state only.

### P0 - Permissions and Safety Parity

- [ ] Implement TS-equivalent permission matcher DSL semantics in `internal/permissions/matcher.go` and `internal/permissions/rules_test.go`.
- [ ] Add workspace allow/deny rule persistence path in `internal/permissions/store.go` and `internal/config/config.go`.
- [ ] Add denial ledger with timestamps/origin metadata in `internal/permissions/store.go` and expose via `/permissions denials` in `internal/commands/handlers.go`.
- [ ] Add auto-mode nuanced safe-write classes in `internal/permissions/permissions.go` and `internal/permissions/classifier.go`.

### P0 - Parity Harness

- [ ] Build TS-vs-Go transcript parity harness under `tests/integration/parity_transcript_test.go` using fixture corpus in `tests/integration/testdata/parity_transcripts/*`.
- [ ] Add full command matrix golden tests for all 67 commands in `tests/integration/command_contract_golden_test.go` and new goldens under `tests/integration/testdata/command_contract/*`.
- [ ] Add permission matrix fixture tests in `tests/integration/permissions_matrix_test.go` with shared fixture loader.

### P1 - Providers, Routing, and Usage Accounting

- [ ] Implement provider option normalization and env/keychain/config auth precedence in `internal/providers/common/auth.go` and `internal/config/config.go`.
- [ ] Expand retry/backoff parity by status/network class in `internal/providers/common/retry.go`.
- [ ] Complete model metadata enrichment (context window/pricing/capabilities) in `internal/providers/model_registry.go` and `internal/providers/metadata_query.go`.
- [ ] Add streaming error translation parity in `internal/providers/common/errors.go`.
- [ ] Add vision/multimodal payload parity where supported in `internal/providers/openai/openai.go`, `internal/providers/anthropic/anthropic.go`, and `internal/providers/gemini/gemini.go`.
- [ ] Propagate cache-hit usage accounting into totals in `internal/agent/loop.go` and `internal/types/provider.go`.

### P1 - TUI and Interaction Fidelity

- [ ] Port high-fidelity wrapping/ansi/unicode behavior in `internal/tui/text_helpers.go` and `internal/tui/app.go`.
- [ ] Implement scroll-box anchor preservation parity in `internal/tui/scroll_helpers.go`.
- [ ] Implement selection/highlight parity in new `internal/tui/selection.go` and `internal/tui/search_models_helpers.go`.
- [ ] Implement terminal capability/focus/notification parity in `internal/tui/capabilities_helpers.go` and `internal/tui/app.go`.
- [ ] Expand app-level modal/dialog state machine in `internal/tui/app.go` and `internal/tui/permission_dialog_state.go`.
- [ ] Add component parity for help/quick-open/history search/task detail in `internal/tui/app.go` and new `internal/tui/dialog_*` modules.

### P1 - State Filesystem, History, Session, Settings

- [ ] Complete startup migration inventory and wiring in `internal/migrations/migrations.go` and `cmd/ac/main.go`.
- [ ] Add project/user layered merge edge parity in `internal/config/config.go` and tests in `internal/config/config_test.go`.
- [ ] Add strict schema validation parity for settings/keybindings/hooks in `internal/settings/validation.go` and `internal/config/config.go`.
- [ ] Align session replay semantics and compaction boundaries with TS in `internal/session/store.go` and `internal/agent/loop.go`.
- [ ] Extend history filtering/ranking semantics in `internal/history/store.go` and `/history` handlers in `internal/commands/handlers.go`.
- [ ] Implement onboarding state parity in `internal/state/onboarding.go` and boot flow in `cmd/ac/main.go`.

### P1 - MCP and Tooling Expansion

- [ ] Add `SkillTool` parity in `internal/tools/skill.go` and registry wiring in `internal/tools/registry.go`.
- [ ] Add schedule/cron execution tools parity in `internal/tools/schedule_cron.go` and `internal/tasks/manager.go`.
- [ ] Add remote trigger and REPL tools parity in `internal/tools/remote_trigger.go` and `internal/tools/repl.go`.
- [ ] Complete MCP auth challenge/result UX parity in `internal/tools/mcp_resource_tools.go` and `internal/mcp/manager.go`.

### P2 - Keybindings, Vim, Voice, Buddy

- [ ] Port keybinding parser/resolver/schema parity into `internal/keybindings/keybindings.go` and new `internal/keybindings/schema.go`.
- [ ] Add reserved shortcut enforcement and conflict diagnostics in `internal/keybindings/keybindings.go`.
- [ ] Extend Vim motion/operator/text-object parity in `internal/vim/engine.go`, `internal/vim/parser.go`, and `internal/vim/buffer.go`.
- [ ] Implement voice runtime integration parity (capability detection + active pipeline) in `internal/voice/local.go` and `internal/tui/app.go`.
- [ ] Expand buddy persistence/interaction parity in `internal/buddy/companion.go`, `internal/buddy/reaction.go`, and config bridge in `internal/config/config.go`.

### P2 - Skills, Plugins, Memdir, Services

- [ ] Add bundled skills registry + metadata parity in `internal/skills/skills.go` and new `internal/skills/bundled/*.go`.
- [ ] Add plugin discovery/runtime lifecycle parity in `internal/plugins/runtime.go`.
- [ ] Add memdir scanner/relevance/aging parity in new `internal/memdir/scanner.go`, `internal/memdir/ranking.go`, `internal/memdir/aging.go`.
- [ ] Port analytics/auth/update/session-memory services parity in new `internal/services/analytics/*.go`, `internal/services/auth/*.go`, `internal/services/update/*.go`, `internal/services/sessionmemory/*.go`.

## Control Metrics to Track Weekly

- Command parity: keep `full/partial/missing` counts and raise `full` from `4` to `>=30` before bridge GA.
- Tool parity: keep `full/partial/missing` counts and move MCP/agent/memory/orchestration tool rows to `full`.
- Subsystem completion: recompute from `PARITY_TODO.md` and target `>=60%` before TS deprecation window.
- P0 burn-down: reduce P0 blockers from `12` to `0` before enabling remote/bridge by default.
