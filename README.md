# AllieCode

**Open-source, model-agnostic AI coding agent.**

AllieCode works with any LLM provider — Anthropic, OpenAI, Google, Mistral, Ollama, and any OpenAI-compatible API. Local-first: works with Ollama out of the box, no account needed.

```
ac                          # start interactive coding session
ac --provider openai        # use OpenAI
ac --provider ollama -m llama3  # use local Ollama
ac undo                     # undo last destructive action
ac models                   # list available models
```

## Why AllieCode?

| Feature | AllieCode | Vendor-locked alternatives |
|---------|-----------|---------------------------|
| Model-agnostic | Any provider, any model | Single vendor only |
| Local-first | Ollama default, no signup | Account required |
| Open source | Apache 2.0 | Proprietary |
| Multi-model routing | Cheap model for tools, powerful for reasoning | Single model |
| Git-native undo | `ac undo` rolls back any change | Manual git management |
| WASM plugins | Skills in any language, sandboxed | Limited extensibility |
| Pair programming | Two models debate before executing | Single perspective |

## Install

### Homebrew

```bash
brew install alliecatowo/tap/alliecode
```

### From binary

```bash
# macOS / Linux
curl -fsSL https://raw.githubusercontent.com/alliecatowo/alliecode/main/install.sh | sh

# Or with Go
go install github.com/alliecatowo/alliecode/cmd/ac@latest
```

### From source

```bash
git clone https://github.com/alliecatowo/alliecode.git
cd alliecode
mise run build
# Binary at ./bin/ac
```

## CI/CD

- Pull requests run CI checks (`go build`, `go test`, `go vet`) via `.github/workflows/ci.yml`.
- Pushes to `main` also run CI for branch health.
- Releases are published only for version tags (`v*`) via `.github/workflows/release.yml`.
- This keeps release automation off normal PRs while maintaining strong pre-merge validation.

## Quick Start

### Interactive onboarding (recommended)

```bash
cp .env.example .env
ac
```

On first run, AllieCode walks you through provider selection and saves your choices in `~/.alliecode/config.yaml`.

### Provider setup options

Use `.env` (from `.env.example`) or export variables in your shell:

```bash
# OpenAI
export OPENAI_API_KEY=your_openai_api_key_here

# Anthropic (API key or OAuth token)
export ANTHROPIC_API_KEY=your_anthropic_api_key_here
export ANTHROPIC_OAUTH_TOKEN=your_anthropic_oauth_token_here

# Gemini
export GEMINI_API_KEY=your_gemini_api_key_here

# Ollama (local)
export OLLAMA_BASE_URL=http://localhost:11434
```

Then start with your preferred provider:

```bash
ac --provider ollama --model llama3
ac --provider anthropic --model claude-sonnet-4-20250514
ac --provider openai --model gpt-4o
```

For OpenAI-compatible endpoints:

```bash
ac --provider openai-compat --model my-model
# Set base_url in config: ~/.alliecode/config.yaml
```

## Configuration

AllieCode uses YAML configuration at `~/.alliecode/config.yaml` (global) and `.alliecode/config.yaml` (project-level).

```yaml
default_provider: ollama
default_model: llama3

providers:
  anthropic:
    api_key: ${ANTHROPIC_API_KEY}
  openai:
    api_key: ${OPENAI_API_KEY}
  ollama:
    base_url: http://localhost:11434

# Multi-model routing (save 60-70% on API costs)
model_routes:
  - primary: claude-sonnet-4-20250514   # complex reasoning
    fast: claude-haiku-3-5-20241022     # simple tool calls
    provider: anthropic

permissions:
  mode: default  # plan | default | auto | bypass

remote:
  enabled: true
  mode: self_hosted   # disabled | local | self_hosted | p2p
  listen_addr: 0.0.0.0:7777
  connect_addr: host.example.net:7777
  token_source: env   # none | env | config | keychain
```

### Self-hosted remote workflow

Use `/session host` on the machine that should host the session, then `/session connect` from another terminal/device.

```bash
# Host side
ac --slash-command "/session host 0.0.0.0:7777"

# Client side
ac --slash-command "/session connect host.example.net:7777 <token>"
```

The `remote` config block can preconfigure the same flow so `ac` starts with your preferred self-hosted mode and token source.

### macOS build note (LC_UUID)

If you saw `missing LC_UUID load command`, use native macOS builds for macOS artifacts:

```bash
mise run build-darwin-native
```

Cross-compiled Linux->Darwin binaries can miss Mach-O metadata expected by some macOS tooling.

## Tools

AllieCode comes with these built-in tools:

| Tool | Description |
|------|-------------|
| `Bash` | Execute shell commands with security classification |
| `Read` | Read files with line numbers |
| `Write` | Create or overwrite files |
| `Edit` | Surgical string replacement in files |
| `Grep` | Regex search across files (uses ripgrep when available) |
| `Glob` | File pattern matching |
| `Agent` | Spawn subagents for parallel work |
| `WebFetch` | Fetch web content |
| `WebSearch` | Search the web |
| `TodoWrite` | Task checklist management |
| `NotebookEdit` | Jupyter notebook editing |

## Skills

Skills are reusable prompts with tool access. Define them as markdown files:

```markdown
---
name: commit
description: Analyze changes and create a git commit
tools: [Bash, Read]
---
Look at the git diff and status, then create a well-crafted commit message...
```

Place in `.alliecode/skills/` or install community skills:

```bash
ac install @community/rails-expert
ac install @community/react-reviewer
```

## MCP Support

AllieCode supports the [Model Context Protocol](https://modelcontextprotocol.io/) for external tool integration:

```yaml
# In .alliecode/config.yaml
mcp_servers:
  - name: github
    command: npx
    args: [-y, @modelcontextprotocol/server-github]
```

## Hooks

Run shell commands on events:

```yaml
hooks:
  - event: pre_tool
    command: echo "Running tool: $ALLIECODE_TOOL_NAME"
  - event: post_chat
    command: ./scripts/log-conversation.sh
```

## Git-Native Checkpointing

Every destructive action auto-snapshots your working directory. Undo instantly:

```bash
ac undo        # restore last checkpoint
```

## Permission System

AllieCode classifies tool calls by risk level and asks before executing dangerous operations:

- **Plan mode**: Read-only tools only
- **Default mode**: Ask for non-read-only operations
- **Auto mode**: Allow safe operations, ask for destructive ones
- **Bypass mode**: Allow everything (use with caution)

Bash commands are automatically classified:
- **Critical**: `rm -rf /`, `dd`, fork bombs
- **High**: `rm -rf`, `git push --force`, `DROP TABLE`
- **Medium**: `rm`, `pip install`, `curl | sh`
- **Low**: `git add`, `mv`, `cp`
- **None**: `ls`, `cat`, `pwd`, `git status`

## Development

```bash
# Prerequisites: Go 1.22+, mise
mise run build     # build binary
mise run test      # run full test suite
mise run test-unit # run unit tests only
mise run test-integration # run deterministic integration tests
mise run test-e2e  # run deterministic e2e smoke tests
mise run test-snapshots # run TUI render snapshot tests
mise run lint      # run linters
mise run fmt       # format code
mise run check     # all of the above
mise run dev       # build and run with debug
```

### Integration and e2e testing

```bash
# Deterministic integration harness (no network required)
mise run test-integration

# Ollama-gated integration checks (requires local Ollama + at least one model)
AC_IT_OLLAMA=1 mise run test-integration-ollama

# Live provider integration lane (network + provider credentials required)
mise run test-live-providers

# Live provider focused lanes
mise run test-live-providers-basic
mise run test-live-providers-tools
mise run test-live-providers-classification

# Full e2e + integration lane with live-provider tests enabled
mise run test-live-lane

# Deterministic CLI/session persistence smoke tests
mise run test-e2e

# Snapshot-style regression tests for core TUI rendering helpers
mise run test-snapshots
```

### Live provider lane configuration

Live tests are opt-in and skip automatically with explicit reasons unless env vars are present.

```bash
# Required lane toggle
export AC_IT_LIVE_PROVIDERS=1

# OpenAI live checks
export OPENAI_API_KEY=your_openai_api_key
export AC_IT_OPENAI_MODEL=gpt-4o-mini    # optional override

# Anthropic live checks (API key or OAuth token)
export ANTHROPIC_API_KEY=your_anthropic_api_key
# export ANTHROPIC_AUTH_TOKEN=...
# export ANTHROPIC_ACCESS_TOKEN=...
export AC_IT_ANTHROPIC_MODEL=claude-haiku-3-5-20241022  # optional override

# Gemini live checks
export GEMINI_API_KEY=your_gemini_api_key
export AC_IT_GEMINI_MODEL=gemini-2.5-flash  # optional override

# Ollama live checks (requires local server + model)
export AC_IT_OLLAMA=1
export OLLAMA_BASE_URL=http://localhost:11434
export AC_IT_OLLAMA_MODEL=llama3

# Run only live provider tests
go test ./tests/integration -run LiveProvider -v

# Run focused live-provider test subsets
go test ./tests/integration -run 'LiveProvider_.*ChatSyncAndStream' -v
go test ./tests/integration -run 'LiveProvider_.*ToolPromptFallback' -v
go test ./tests/integration -run 'LiveProvider_Classification' -v

# Requested mega lane command (skips gated tests when env is missing)
go test ./tests/e2e ./tests/integration -v
```

## Architecture

```
cmd/ac/                  CLI entry point (cobra)
internal/
  agent/                 Core agent loop, context management, checkpointing
  providers/             Provider interface + implementations
    anthropic/           Anthropic (Claude) provider
    openai/              OpenAI provider
    ollama/              Ollama (local) provider
    openaicompat/        Generic OpenAI-compatible provider
  tools/                 Tool interface + built-in tools
  permissions/           Permission engine + bash risk classifier
  tui/                   Bubbletea terminal UI
  config/                YAML settings management
  mcp/                   MCP client
  skills/                Skill/plugin system
  hooks/                 Event hooks
  types/                 Universal message format & interfaces
```

## Contributing

We welcome contributions! Key areas:

- **Providers**: Add support for Google Gemini, Mistral, Cohere, etc.
- **Tools**: Build new tools or improve existing ones
- **Skills**: Create and share community skills
- **Plugins**: Build WASM plugins (coming in v0.3)
- **Tests**: Improve coverage
- **Docs**: Better documentation and examples

See [ROADMAP.md](ROADMAP.md) for the full project vision.

## License

Apache 2.0 — see [LICENSE](LICENSE).
