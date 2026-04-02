# AllieCode Contributor/Operator Contract

This contract defines how AllieCode (and human contributors following the same workflow) must operate in this repository.

## 1) Operating Principles

- Safety first: do not run destructive commands (`rm -rf`, `git reset --hard`, force-push, schema/data destructive ops) unless explicitly requested and clearly scoped.
- Least privilege: only touch files required for the task; avoid broad refactors unless requested.
- Explain intent: for non-trivial changes, include a short rationale in PR/commit context (why, risk, validation).
- Keep diffs reviewable: prefer small, focused commits and avoid mixing unrelated changes.

## 2) Permissions and Guardrails

- Scope discipline: modify only files requested by the task; if additional edits are needed, state why before making them.
- Secret handling: never commit credentials, tokens, `.env` secrets, or generated secret material.
- Network/system actions: avoid infra, billing, or production-impacting operations without explicit instruction.
- Git safety: do not rewrite shared history unless explicitly asked; never force-push main/master.

## 3) Coding and Parity Expectations

- Behavioral parity required: preserve existing user-visible behavior unless a behavior change is explicitly requested.
- API/CLI parity: keep flags, output shape, and error semantics compatible when modifying existing paths.
- Cross-path parity: if updating one execution path, check and align equivalent code paths/tests.
- Backward compatibility first: when uncertain, prefer additive changes over breaking changes.

## 4) Subagent and Ownership Rules

- Single-writer rule: one agent owns a file set at a time; avoid overlapping edits to the same file.
- Clear boundaries: split work by package/module boundaries when parallelizing.
- Reconcile intentionally: if overlap is unavoidable, re-read final merged files and resolve for consistency before commit.
- No hidden side work: do not slip in opportunistic changes outside assigned scope.

## 5) Go + mise Workflow (No Make)

- Tooling source of truth: use `mise` for runtime/tools; do not add or rely on `make` targets.
- Standard command entry:
  - `mise install` (bootstrap toolchain)
  - `mise exec -- go version` (verify Go toolchain)
  - `mise exec -- go test ./...` (full test pass)
- Build/check examples:
  - `mise exec -- go build ./...`
  - `mise exec -- go test ./... -race`
  - `mise exec -- go test ./path/to/pkg -run TestName`
- Prefer repository-documented scripts/tasks when present, but invoke through `mise`.

## 6) Testing and Validation Contract

- Run relevant tests for every code change; default to `mise exec -- go test ./...` unless a narrower scope is justified.
- For concurrency-sensitive changes, include `-race` where feasible.
- If tests are skipped (time/env constraints), state exactly what was skipped and why.
- Do not claim success without command-level validation evidence.

## 7) Commit and Release Expectations

- Commit style: Conventional Commits (`feat:`, `fix:`, `docs:`, `refactor:`, `test:`, `chore:`).
- Commit message quality: concise subject, imperative mood, scoped to the change.
- Pre-release quality gate: tests pass, docs updated for behavior/config changes, and changelog/release notes prepared when applicable.
- Release posture: prioritize reproducibility and rollback clarity; avoid bundling unrelated risk in release-bound commits.

## 8) Definition of Done

- Change is scoped, reviewed for parity impact, and validated by appropriate tests.
- Documentation is updated when operator or user-facing behavior changes.
- Commit history is clean, readable, and traceable to intent.
