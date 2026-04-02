# Parity Accounting Ledger (TS -> Go)

This ledger gives a plain-English snapshot of why parity is not line-for-line yet, what remains by subsystem, and the directory-level port backlog.

## Why TS file/LOC is larger than Go right now

- Current measured code volume is heavily asymmetric: TS/TSX is about 1,884 files and ~513k LOC under `src/`, while Go is about 258 files and ~63k LOC across `cmd/`, `internal/`, and `tests/`.
- The TS tree still contains full historical product surface (rich custom Ink renderer, large component/dialog inventory, mature services layer, and broad utility library), while Go is intentionally shipping a narrower baseline-first implementation.
- Go consolidates behavior into fewer modules (especially command handlers, tool registry/runtime, and Bubble Tea UI helpers), so equivalent behavior can exist with fewer files/LOC.
- A large TS long tail (`src/utils/*`, `src/components/*`, `src/services/*`, `src/commands/*`) includes UX polish, edge handling, and provider/platform variants that are only partially ported.
- Some TS features are intentionally not 1:1 targets for local OSS parity (Anthropic account/cloud-specific paths), which keeps current Go LOC lower by design.

## TS accounting by subsystem (current tree)

| Subsystem | TS roots | File count | Rough LOC bucket |
|---|---|---:|---|
| Core runtime and orchestration | `src/main.tsx`, `src/query*`, `src/context*`, `src/entrypoints/*`, `src/tasks*`, `src/setup.ts`, `src/state/*` | 96 | ~29k (medium) |
| CLI commands and screens | `src/commands/*`, `src/cli/*`, `src/screens/*` | 215 | ~46k (medium-large) |
| TUI and components | `src/components/*`, `src/ink/*`, `src/hooks/*`, `src/keybindings/*`, `src/vim/*`, `src/voice/*` | 614 | ~130k (very large) |
| Tools and integrations | `src/tools/*`, `src/services/*`, `src/skills/*`, `src/plugins/*`, `src/memdir/*` | 345 | ~111k (very large) |
| Remote/bridge/server | `src/bridge/*`, `src/remote/*`, `src/server/*`, `src/upstreamproxy/*` | 41 | ~15k (small-medium) |
| Utilities/shared helpers | `src/utils/*` and small shared roots | 573 | ~182k (largest) |

## Subsystem -> Go mapping and parity status

| Subsystem | Go equivalents | Parity status | Notes |
|---|---|---|---|
| Boot/runtime and agent loop | `cmd/ac/*`, `internal/agent/*`, `internal/hooks/*`, `internal/state/*` | partial | Core loop, streaming, hooks, and budgets exist; full stop taxonomy, context dependency graph, and resume/replay parity still open. |
| Providers and model routing | `internal/providers/*` | partial | Anthropic/OpenAI/Ollama/OpenAI-compat/Gemini baselines exist; retry/auth/metadata and multimodal parity remain. |
| Tools framework and built-ins | `internal/tools/*`, `internal/tasks/*`, `internal/mcp/*` | partial | Major built-ins are implemented; lifecycle telemetry, cancellation nuances, and full advanced tool parity are incomplete. |
| Permissions and risk policy | `internal/permissions/*` | partial | Mode engine, classifier, and persisted store are in place; TS rule DSL and richer denials/rule UX parity remain. |
| Slash commands and CLI UX | `internal/commands/*` | partial | Registry/parser and broad command surface exist; many commands are runtime-state baselines rather than full TS-integrated workflows. |
| TUI rendering and interactions | `internal/tui/*`, `internal/keybindings/*`, `internal/vim/*`, `internal/voice/*` | partial | Bubble Tea baseline and key interaction helpers exist; high-fidelity custom Ink parity and full dialog/component matrix remain. |
| State/config/history/session | `internal/config/*`, `internal/settings/*`, `internal/history/*`, `internal/session/*`, `internal/migrations/*` | partial | Baseline persistence works; layered config merge, strict schema, migration, and doctor parity are still pending. |
| Bridge/remote/server path | `internal/bridge/*`, `internal/remote/*`, `internal/mcp/*` | partial | Self-hosted-first baseline is present; full session lifecycle, permission bridge, and transport parity still open. |
| Skills/plugins/memory | `internal/skills/*`, `internal/plugins/*` plus future `internal/memdir/*` | partial | Local skills and plugin runtime baseline exist; bundled skills index and memdir scan/rank/aging parity remain. |

## Intentionally ignored or jettisoned features

- Anthropic account-locked cloud workflow paths are not a strict 1:1 target for local OSS parity (billing/account/hosted control plane semantics).
- Anthropic-managed remote broker assumptions are jettisoned in favor of self-hosted-first bridge/remote transport.
- Proprietary telemetry/account analytics surfaces are treated as optional and replaceable, not hard parity blockers for local operation.
- Vendor-specific install/onboarding wrappers (for Anthropic-hosted app ecosystems) are not required for core local CLI agent parity.

## Open-source replacements for Anthropic-locked features

| Anthropic-locked capability class | OSS replacement direction |
|---|---|
| Hosted model/account gatekeeping | Multi-provider local routing in `internal/providers/*` (OpenAI-compatible endpoints, Ollama, Gemini, Anthropic when available). |
| Hosted remote session broker | Self-hosted bridge/remote stack in `internal/bridge/*` + `internal/remote/*` (direct connect/session pointer path). |
| Closed platform tool ecosystem | Local built-in tools + MCP extensibility in `internal/tools/*` and `internal/mcp/*`. |
| Proprietary skill distribution | Markdown/local skill loading in `internal/skills/*` with bundled-skill backlog. |
| Vendor cloud memory/context services | Local session/history stores (`internal/session/*`, `internal/history/*`) and planned memdir port. |

## Actionable grouped file-port backlog (directory-level)

### Group A - Core runtime parity

- `src/query*`, `src/context*`, `src/assistant/*` -> `internal/agent/*`, `internal/types/*`, `internal/hooks/*`
- Focus: stop reason taxonomy, dependency graph teardown order, session replay/resume fidelity.

### Group B - Command and CLI integration parity

- `src/commands/*`, `src/cli/*`, `src/screens/*` -> `internal/commands/*`, `cmd/ac/*`, `internal/tui/*`
- Focus: replace runtime-state placeholders with git/history/config/mcp-backed behavior, expand doctor/history/init/mcp/config advanced flows.

### Group C - TUI/component fidelity parity

- `src/components/*`, `src/ink/*`, `src/hooks/*` -> `internal/tui/*` (+ targeted `internal/keybindings/*`)
- Focus: high-fidelity wrapping/selection/search/scroll, modal/dialog state machine, quick-open/help/history/task detail UIs.

### Group D - Bridge/remote/server parity

- `src/bridge/*`, `src/remote/*`, `src/server/*`, `src/upstreamproxy/*` -> `internal/bridge/*`, `internal/remote/*`, `internal/mcp/*`
- Focus: secure pairing/session lifecycle, permission callback parity, reconnect/backpressure, transport fallback matrix.

### Group E - Tooling and MCP parity

- `src/tools/*`, `src/tasks/*` -> `internal/tools/*`, `internal/tasks/*`, `internal/mcp/*`
- Focus: tool event lifecycle telemetry, cancellation semantics, MCP resource/auth flows, remaining advanced tool families.

### Group F - Config/state/migration parity

- `src/setup.ts`, `src/schemas/*`, `src/migrations/*`, `src/state/*` -> `internal/config/*`, `internal/settings/*`, `internal/migrations/*`, `internal/state/*`
- Focus: layered config merge precedence, strict schema errors, migration ledger/doctor checks, onboarding state parity.

### Group G - Platform/support parity

- `src/services/*`, `src/keybindings/*`, `src/vim/*`, `src/voice/*`, `src/skills/*`, `src/plugins/*`, `src/memdir/*` -> corresponding `internal/*` packages
- Focus: service-level parity (auth/update/analytics), full keybinding/vim/voice behavior, bundled skills/plugins, memdir scan/rank/aging.

## Current parity posture

- Overall status: broad baseline achieved, deep behavioral parity still partial across most subsystems.
- Practical target: close Group A/B/D first to unlock end-to-end workflow parity, then Group C/F/G for fidelity and long-tail compatibility.
