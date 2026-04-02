# AllieCode Differentiators Roadmap

This roadmap defines the product differentiators that should remain visible in planning, implementation, and release decisions.

## Scope and Planning Model

- **Parity work**: Changes required to match existing expected behavior, compatibility, and operator workflows.
- **Intentional improvements**: Net-new or opinionated capabilities that make AllieCode distinct beyond baseline parity.

## Differentiators

### 1) Multi-model routing

- **Category**: Intentional improvement
- **User value**: Users get better quality/cost/speed by routing tasks to the best model profile automatically or by policy.
- **Implementation status**: In progress (core routing exists, policy depth and observability are not complete).
- **Acceptance criteria**:
  - Users can select a routing strategy (`auto`, `cost`, `speed`, `quality`, pinned model) per session.
  - Routing decisions are explainable in logs/UI (selected model, reason, fallback path).
  - Failover works across at least two providers without user intervention.
  - Regression tests cover routing policy selection and fallback behavior.
- **Milestones**:
  - **M1**: Stabilize routing API and per-session policy selection.
  - **M2**: Add decision telemetry and user-facing explanation surface.
  - **M3**: Add provider health signals + automatic fallback hardening.
  - **M4**: Publish routing playbook and SLA-style behavior guarantees.

### 2) Local-first execution

- **Category**: Intentional improvement
- **User value**: Faster iteration, lower trust friction, and better privacy by default through local toolchain and workspace operations.
- **Implementation status**: In progress (core local execution is stable; final parity blockers are bridge/remote replacement coverage, full command parity, and deterministic end-to-end parity CI).
- **Acceptance criteria**:
  - Core workflows (read/edit/search/test/git) run without cloud dependency.
  - Clear controls exist for network access and external data egress.
  - Performance baseline demonstrates low latency for local edit loops.
  - Docs explicitly define local vs remote data handling boundaries.
- **Milestones**:
  - **M1**: Complete local capability matrix and gap list. (done)
  - **M2**: Add explicit network permission toggles and policy defaults.
  - **M3**: Introduce offline-safe fallback behavior for critical commands.
  - **M4**: Ship operator docs for privacy posture and deployment modes.
  - **Current blockers to 100% local parity**:
    - Bridge/remote/session workflows still depend on unported TS paths.
    - Some slash commands are still partial compared with TS (`doctor`, `history`, `mcp`, advanced `config`).
    - Full parity harness/release matrix is not yet deterministic across provider/TUI/MCP paths.

### 3) Buddy collaboration mode

- **Category**: Intentional improvement
- **User value**: Teams can pair with the assistant while preserving authorship clarity, reviewability, and shared context.
- **Implementation status**: In progress (buddy core, command wiring, TUI integration, config persistence, and baseline tests are implemented).
- **Acceptance criteria**:
  - Buddy mode can be enabled per task/session with visible state.
  - Suggestions and applied edits are attributable (who/what/when) in activity history.
  - Handoff flow supports resumable context between collaborators.
  - Collaboration paths include conflict-safe file ownership cues.
- **Milestones**:
  - **M1**: Define collaboration primitives and attribution model. (done)
  - **M2**: Implement buddy session state + handoff metadata. (done - baseline)
  - **M3**: Add UI/CLI affordances for co-authoring and conflict warnings. (in progress)
  - **M4**: Validate with pilot teams and tune ergonomics.

### 4) Git checkpoints

- **Category**: Intentional improvement
- **User value**: Users can move quickly with confidence via lightweight, reversible checkpoints before risky edits.
- **Implementation status**: In progress (manual workflows supported; automated checkpoint UX not complete).
- **Acceptance criteria**:
  - Users can create named checkpoints before multi-file or high-risk operations.
  - Restore flow is one command with clear preview of changed files.
  - Checkpoint metadata captures rationale and timestamp.
  - Checkpoint behavior avoids destructive history rewrite by default.
- **Milestones**:
  - **M1**: Add checkpoint create/list/restore primitives.
  - **M2**: Integrate auto-prompted checkpoints for risky operations.
  - **M3**: Add dry-run/preview and policy controls.
  - **M4**: Document best practices and recovery runbooks.

### 5) Plugin ecosystem

- **Category**: Intentional improvement
- **User value**: Teams can extend AllieCode into domain-specific workflows without forking core behavior.
- **Implementation status**: Planned (plugin boundary concepts exist; SDK, lifecycle, and trust model are pending).
- **Acceptance criteria**:
  - Stable plugin API with versioning and compatibility contract.
  - Plugin install/enable/disable lifecycle is observable and reversible.
  - Permission model constrains file/system/network access per plugin.
  - Reference plugins demonstrate real-world integrations.
- **Milestones**:
  - **M1**: Freeze minimal plugin API surface and lifecycle hooks.
  - **M2**: Release alpha SDK and local plugin loader.
  - **M3**: Implement permissions, signing policy, and audit logs.
  - **M4**: Launch curated registry with governance guidelines.

### 6) Safety UX

- **Category**: Intentional improvement
- **User value**: Users avoid accidental destructive actions through clear guardrails, previews, and explicit risk communication.
- **Implementation status**: In progress (guardrails exist; risk tiering and guided remediation need unification).
- **Acceptance criteria**:
  - High-risk actions present plain-language warnings and require explicit user confirmation.
  - Non-destructive alternatives are suggested before destructive paths.
  - Safety events are logged with actionable remediation guidance.
  - UX copy is consistent across CLI and any companion surfaces.
- **Milestones**:
  - **M1**: Define unified risk taxonomy and action classes.
  - **M2**: Add consistent confirmation prompts and preview-first patterns.
  - **M3**: Implement remediation hints and policy-based blocks.
  - **M4**: Run usability pass focused on operator trust and clarity.

### 7) Signature spinner and branding

- **Category**: Intentional improvement
- **User value**: A recognizable interaction identity improves confidence, delight, and product distinctiveness without slowing workflows.
- **Implementation status**: In progress (branded spinner baseline is implemented; accessibility/perf guardrails and contributor style guidance remain).
- **Acceptance criteria**:
  - Spinner/brand elements are consistent, lightweight, and accessible.
  - Animations respect reduced-motion preferences.
  - Branding never obscures critical status/error information.
  - Style guidance is documented for contributors.
- **Milestones**:
  - **M1**: Finalize motion and identity guidelines. (done - baseline)
  - **M2**: Implement spinner states for thinking/executing/waiting. (done - baseline)
  - **M3**: Add accessibility conformance checks and perf budget.
  - **M4**: Publish branding usage rules in contributor docs.

### 8) Test quality bar

- **Category**: Parity work + intentional improvement
- **User value**: Reliable behavior changes with fewer regressions and clearer confidence signals before release.
- **Implementation status**: In progress (buddy and spinner harness updates landed; broader parity matrix and release-gate consistency remain uneven).
- **Acceptance criteria**:
  - Defined minimum quality gate for changed code paths (unit + integration where applicable).
  - Standard CI path includes deterministic test execution and failure triage output.
  - Concurrency-sensitive areas include race-focused validation where feasible.
  - Release readiness requires passing documented quality gates.
- **Milestones**:
  - **M1**: Publish test policy and required checks by change type.
  - **M2**: Normalize CI gates and flaky-test handling workflow. (in progress - harness updated)
  - **M3**: Add race/concurrency test coverage for critical paths.
  - **M4**: Enforce release gate with audit-ready reporting.

## Parity vs Intentional Improvements Snapshot

### Parity work (must-have baseline)

- Preserve existing CLI/API behavior compatibility.
- Maintain safe defaults and non-destructive operation semantics.
- Enforce consistent, reproducible test gates across core workflows.

### Intentional improvements (distinctive value)

- Adaptive multi-model routing with explainability and failover.
- Local-first execution posture with explicit data-boundary controls.
- Buddy collaboration, checkpoints, and plugin extensibility.
- Signature interaction identity and trust-first safety UX.
